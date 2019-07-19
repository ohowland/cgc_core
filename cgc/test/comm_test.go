package cgc

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/goburrow/modbus"
)

func TestModbus(t *testing.T) {
	handler := modbus.NewTCPClientHandler("192.168.0.100:5020")
	handler.Timeout = 100 * time.Millisecond
	handler.Logger = log.New(os.Stdout, "test: ", log.LstdFlags)

	err := handler.Connect()
	defer handler.Close()

	if err != nil {
		t.Errorf("failed to connect to target")
		t.FailNow()
	}

	client := modbus.NewClient(handler)
	_, err = client.ReadHoldingRegisters(1, 4)

	if err != nil {
		t.Errorf("failed to read target registers")
		t.FailNow()
	}
}
