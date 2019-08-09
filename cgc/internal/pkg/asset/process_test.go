package asset

import (
	"log"
	"testing"
	"time"

	"gotest.tools/assert"
)

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

type DummyDevice struct{}

func (d DummyDevice) ReadDeviceStatus() error {
	log.Print("reading dummy device...")
	time.Sleep(time.Duration(100) * time.Millisecond)
	return nil
}

func (d DummyDevice) WriteDeviceControl() error {
	log.Print("writing dummy device...")
	time.Sleep(time.Duration(100) * time.Millisecond)
	return nil
}

func TestProcess(t *testing.T) {
	device := DummyDevice{}
	inbox := StartProcess(device)

	inbox <- UpdateStatus{}
	inbox <- WriteControl{}
	inbox <- Stop{}
}
