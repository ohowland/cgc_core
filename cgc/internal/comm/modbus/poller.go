package comm

import (
	"encoding/binary"
	"errors"
	"log"
	"math"
	"os"
	"time"

	"github.com/goburrow/modbus"
	"github.com/ohowland/cgc/internal/comm"
)

// Poller continiously polls a target
type Poller struct {
	handler  *modbus.TCPClientHandler
	pollRate int
}

// PollerConfig is the configuration format for ModbusPoller
type PollerConfig struct {
	ipAddr       string
	port         string
	slaveID      byte
	timeout      int
	pollRate     int
	enableLogger bool
}

// NewPoller is a factory for the Poller struct
func newPoller(cfg PollerConfig) Poller {
	handler := modbus.NewTCPClientHandler(cfg.ipAddr + ":" + cfg.port)
	handler.Timeout = time.Millisecond * time.Duration(cfg.timeout)
	handler.SlaveId = cfg.slaveID

	if cfg.enableLogger {
		handler.Logger = log.New(os.Stdout, "modbus: ", log.LstdFlags)
	}

	return Poller{
		handler:  handler,
		pollRate: cfg.pollRate,
	}
}

func (m Poller) Read(registers []comm.Register) (map[string]float64, error) {
	err := m.handler.Connect()
	if err != nil {
		return nil, err
	}
	defer m.handler.Close()

	client := modbus.NewClient(m.handler)
	readValues := make(map[string]float64)
	for _, register := range registers {
		resp, readErr := client.ReadHoldingRegisters(register.address, sizeOf(register.datatype))
		if readErr != nil {
			readValues[register.name] = 0xBEEF
			err = readErr
		} else {
			readValues[register.name] = decode(resp, register)
		}
	}
	return readValues, err
}

func (m Poller) Write(registers []comm.Register, writeValues map[string]float64) error {
	err := m.handler.Connect()
	if err != nil {
		return err
	}
	defer m.handler.Close()

	client := modbus.NewClient(m.handler)
	for name, val := range writeValues {
		i, writeErr := findIndexByName(registers, name)
		if writeErr != nil {
			err = writeErr
		} else {
			valBytes := encode(val, registers[i])
			_, writeErr = client.WriteMultipleRegisters(registers[i].address, sizeOf(registers[i].datatype), valBytes)
			if writeErr != nil {
				err = writeErr
			}
		}
	}
	return err
}

// findIndexByName returns the index in the array of the register, if found. Returns -1 and error if not found.
func findIndexByName(registers []comm.Register, name string) (int, error) {
	for index, register := range registers {
		if register.name == name {
			return index, nil
		}
	}
	return -1, errors.New("register name not found in register array")
}

// encode convert a float64 into a byte array
func encode(val float64, register comm.Register) []byte {
	var bytes []byte
	endian := getByteOrder(register.endianness)
	switch register.datatype {
	case u16, i16:
		bytes = make([]byte, 2*sizeOf(u16))
		endian.PutUint16(bytes, uint16(val))
	case u32, i32:
		bytes = make([]byte, 2*sizeOf(u32))
		endian.PutUint32(bytes, uint32(val))
	case f32:
		bytes = make([]byte, 2*sizeOf(f32))
		endian.PutUint32(bytes, math.Float32bits(float32(val)))
	case u64, i64:
		bytes = make([]byte, 2*sizeOf(u64))
		endian.PutUint64(bytes, uint64(val))
	case f64:
		bytes = make([]byte, 2*sizeOf(f64))
		endian.PutUint64(bytes, math.Float64bits(val))
	}
	return bytes
}

// decode coverts byte arrays into float64s
func decode(bytes []byte, register comm.Register) float64 {
	var n float64
	endian := getByteOrder(register.endianness)
	switch register.datatype {
	case u16:
		n = float64(endian.Uint16(bytes))
	case i16:
		n = float64(int16(endian.Uint16(bytes)))
	case u32:
		n = float64(endian.Uint32(bytes))
	case i32:
		n = float64(int32(endian.Uint32(bytes)))
	case f32:
		bits := endian.Uint32(bytes)
		n = float64(math.Float32frombits(bits))
	case u64:
		n = float64(endian.Uint64(bytes))
	case i64:
		n = float64(int64(endian.Uint64(bytes)))
	case f64:
		bits := endian.Uint64(bytes)
		n = math.Float64frombits(bits)
	}
	return n
}

// getByteOrder returns the correct binary.endian object for the register type
func getByteOrder(e Endian) binary.ByteOrder {
	switch e {
	case bigEndian:
		return binary.BigEndian
	case littleEndian:
		return binary.LittleEndian
	}
	return binary.BigEndian
}

// sizeOf returns the number of u16 registers for the datatype
func sizeOf(t DataType) uint16 {
	switch t {
	case u16:
		return 1
	case i16:
		return 1
	case u32:
		return 2
	case i32:
		return 2
	case f32:
		return 2
	case u64:
		return 4
	case i64:
		return 4
	case f64:
		return 4
	}
	return 0
}
