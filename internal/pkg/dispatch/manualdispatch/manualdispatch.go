package manualdispatch

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/lib/webservice"
	"github.com/ohowland/cgc/internal/pkg/asset"
	"github.com/ohowland/cgc/internal/pkg/asset/ess"
	"github.com/ohowland/cgc/internal/pkg/asset/feeder"
	"github.com/ohowland/cgc/internal/pkg/asset/grid"
	"github.com/ohowland/cgc/internal/pkg/asset/pv"
	"github.com/ohowland/cgc/internal/pkg/dispatch"
)

// ManualDispatch is the core datastructure
type ManualDispatch struct {
	mux           *sync.Mutex
	calcStatus    *dispatch.CalculatedStatus
	handler       webservice.Config
	memberStatus  map[uuid.UUID]interface{}
	memberControl map[uuid.UUID]interface{}
}

// New returns a configured ManualDispatch struct
func New(configPath string) (ManualDispatch, error) {
	jsonConfig, err := ioutil.ReadFile(configPath)
	if err != nil {
		return ManualDispatch{}, err
	}
	handler := webservice.Config{}
	if err := json.Unmarshal(jsonConfig, &handler); err != nil {
		return ManualDispatch{}, err
	}

	calcStatus, err := dispatch.NewCalculatedStatus()
	memberStatus := make(map[uuid.UUID]interface{})
	memberControl := make(map[uuid.UUID]interface{})
	return ManualDispatch{
			&sync.Mutex{},
			&calcStatus,
			handler,
			memberStatus,
			memberControl,
		},
		nil
}

// UpdateStatus ...
func (c *ManualDispatch) UpdateStatus(msg asset.Msg) {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.memberStatus[msg.PID()] = msg.Payload()
	c.calcStatus.AggregateMemberStatus(msg)
}

// DropAsset ...
func (c *ManualDispatch) DropAsset(pid uuid.UUID) error {
	c.mux.Lock()
	defer c.mux.Unlock()
	c.calcStatus.DropAsset(pid)
	delete(c.memberStatus, pid)
	delete(c.memberControl, pid)
	return nil
}

// GetControl ...
func (c ManualDispatch) GetControl() map[uuid.UUID]interface{} {
	return c.memberControl
}

// GetStatus ...
func (c ManualDispatch) GetStatus() map[uuid.UUID]interface{} {
	return c.memberStatus
}

// GetCalcStatus ...
func (c ManualDispatch) GetCalcStatus() map[uuid.UUID]dispatch.Status {
	return c.calcStatus.MemberStatus()
}

func (c ManualDispatch) updateHandler() {
	log.Println("STATUS: ", c.GetStatus())
	for pid, status := range c.GetStatus() {

		var b []byte
		var err error
		switch s := status.(type) {
		case ess.Status:
			b, err = json.Marshal(s)
		case grid.Status:
			b, err = json.Marshal(s)
		case feeder.Status:
			b, err = json.Marshal(s)
		case pv.Status:
			b, err = json.Marshal(s)
		default:
			b = json.RawMessage("{}")
			err = errors.New("manualDispatch.updateHandler: Could not cast status")
		}

		if err != nil {
			log.Println(err)
		}

		targetURL := c.handler.URL + "/assets/" + pid.String() + "/status"
		log.Println("TARGET:", targetURL)
		log.Println("JSON:", b)
		_, err = http.Post(targetURL, "Content-Type: application/json", bytes.NewBuffer(b))
	}
}
