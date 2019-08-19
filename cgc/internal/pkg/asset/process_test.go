package asset

import (
	"log"
	"testing"
	"time"

	"github.com/google/uuid"

	"gotest.tools/assert"
)

type DummyAsset struct {
	pid     uuid.UUID
	status  DummyStatus
	control DummyControl
	device  DummyDevice
}

type DummyStatus struct {
	state string
}
type DummyControl struct {
	cmd string
}

type DummyConfig struct{}

func (d DummyAsset) PID() uuid.UUID {
	return d.pid
}

func (d *DummyAsset) Status() interface{} {
	return DummyStatus{}
}

func (d DummyAsset) Control(interface{}) {}

func (d DummyAsset) Config(interface{}) {}

func (d *DummyAsset) UpdateStatus() error {
	response, err := d.device.ReadDeviceStatus()
	d.status = response.(DummyStatus)
	return err
}

func (d DummyAsset) WriteControl() error {
	err := d.device.WriteDeviceControl(d.control)
	return err
}

type DummyDevice struct{}

func (d DummyDevice) ReadDeviceStatus() (interface{}, error) {
	log.Print("reading dummy device...")
	time.Sleep(time.Duration(100) * time.Millisecond)
	response := DummyStatus{state: "online"}
	return response, nil
}

func (d DummyDevice) WriteDeviceControl(interface{}) error {
	log.Print("writing dummy device...")
	time.Sleep(time.Duration(100) * time.Millisecond)
	return nil
}

func TestTarget(t *testing.T) {
	ch := make(chan interface{})
	msg := "testing..."
	rate := 1
	target := NewTarget(ch, msg, rate)
	assert.Assert(t, target.msg == msg)
	assert.Assert(t, target.rate == time.Duration(rate)*time.Millisecond)

	time.Sleep(time.Duration(rate*2) * time.Millisecond)

	go target.send()

	incoming := <-ch
	assert.Assert(t, incoming == msg)
}

func TestSchedulerAddTarget(t *testing.T) {
	ch := make(chan interface{})

	msgs := []string{"testing...1", "testing...2"}

	target1 := NewTarget(ch, msgs[0], 1)
	target2 := NewTarget(ch, msgs[1], 1)

	s := NewScheduler()
	s.addTarget(target1)
	s.addTarget(target2)

	for i, target := range s.targets {
		assert.Assert(t, target.msg == msgs[i])
	}
}

func TestSchedulerMsgTwoTargets(t *testing.T) {
	ch1 := make(chan interface{})
	ch2 := make(chan interface{})
	msgs := []string{"testing...1", "testing...2"}
	target1 := NewTarget(ch1, msgs[0], 100)
	target2 := NewTarget(ch2, msgs[1], 100)

	s := NewScheduler()
	s.addTarget(target1)
	s.addTarget(target2)

	s.run()
	defer s.stop()
	incoming1 := <-ch1
	incoming2 := <-ch2

	assert.Assert(t, incoming1 == msgs[0])
	assert.Assert(t, incoming2 == msgs[1])
}

func TestSchedulerMsgTargetRepeat(t *testing.T) {
	ch := make(chan interface{})
	target := NewTarget(ch, "testing...", 10)

	s := NewScheduler()
	s.addTarget(target)

	s.run()
	defer s.stop()

	max := 10
	incoming := make([]interface{}, max)
	for i := 0; i < max; i++ {
		incoming[i] = <-ch
		//fmt.Fprintf(os.Stdout, "recieved: %v\n", incoming[i])
	}

	assert.Assert(t, incoming[max-1] == "testing...")
}

func TestProcess(t *testing.T) {
	asset := DummyAsset{}
	inbox := InitializeProcess(&asset)

	inbox <- UpdateStatus{}
	inbox <- WriteControl{}
	inbox <- Stop{}
	close(inbox)
}

func TestProcessScheduled(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestProcessScheduled in short mode")
	}

	asset := DummyAsset{}
	inbox := InitializeProcess(&asset)

	time.Sleep(time.Duration(3) * time.Second)
	inbox <- Stop{}
	close(inbox)
}
func TestProcessStop(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestProcessScheduled in short mode")
	}

	asset := DummyAsset{}
	inbox := InitializeProcess(&asset)

	time.Sleep(time.Duration(2) * time.Second)
	inbox <- Stop{}
	time.Sleep(time.Duration(2) * time.Second)
	inbox <- Start{}
	time.Sleep(time.Duration(2) * time.Second)
	inbox <- Stop{}
	close(inbox)
}
