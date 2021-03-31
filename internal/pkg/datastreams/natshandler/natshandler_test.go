package natshandler

import (
	"testing"
	"time"

	nats "github.com/nats-io/nats.go"
	"gotest.tools/v3/assert"
)

func TestNatsConnector(t *testing.T) {
	nc, err := nats.Connect(nats.DefaultURL)

	assert.NilError(t, err)

	nc.Subscribe("foo", func(m *nats.Msg) {
		t.Logf("Recieved msg: %s\n", string(m.Data))
	})

	nc.Publish("foo", []byte("Hello World"))

	time.Sleep(2 * time.Second)

	nc.Close()
}
