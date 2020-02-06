package manualdispatch

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"sync"

	"github.com/google/uuid"
	"github.com/ohowland/cgc/internal/pkg/asset"
	"github.com/ohowland/cgc/internal/pkg/dispatch"
	"github.com/ohowland/cgc/internal/lib/webservice"
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
	for pid, status := range c.GetStatus() {
		json, err := json.Marshal(status)
		if err != nil {
			log.Println(err)
			continue
		}
		bytes.NewBuffer(json)
		http.NewRequest("POST", c.handler.URL+c.handler.Port+"/asset/"+pid.String()+"/status", nil)
	}
}
