package modbus

import (
	"encoding/binary"
	"errors"
	"math"
	"time"

	"github.com/goburrow/modbus"
)

// Endian byte order of register
type Endian string

const (
	littleEndian Endian = "little"
	bigEndian    Endian = "big"
)

// DataType is the datatypes supported by the Modbus protocl
type DataType string

const (
	u16 DataType = "u16"
	u32 DataType = "u32"
	u64 DataType = "u64"
	i16 DataType = "i16"
	i32 DataType = "i32"
	i64 DataType = "i64"
	f32 DataType = "f32"
	f64 DataType = "f64"
)

// Register contains the data required to read and write a Modbus register
type Register struct {
	name         string
	address      uint16
	datatype     DataType
	functionCode int
	endianness   Endian
}

// PollerConfig is the configuration format for ModbusPoller
type PollerConfig struct {
	ipAddr   string
	port     string
	timeout  int
	pollRate int
	offset   int
}

// Poller continiously polls a target
type Poller struct {
	handler  *modbus.TCPClientHandler
	pollRate int
	offset   int
}

// NewPoller is a factory for the Poller struct
func NewPoller(cfg PollerConfig) Poller {
	handler := modbus.NewTCPClientHandler(cfg.ipAddr + ":" + cfg.port)
	handler.Timeout = time.Millisecond * time.Duration(cfg.timeout)
	return Poller{
		handler:  handler,
		pollRate: cfg.pollRate,
	}
}

func (m Poller) Read(registers []Register) (map[string]float64, error) {
	err := m.handler.Connect()
	if err != nil {
		return nil, err
	}
	defer m.handler.Close()

	client := modbus.NewClient(m.handler)
	readValues := make(map[string]float64)
	for _, register := range registers {
		resp, readErr := client.ReadHoldingRegisters(register.address, sizeOf(register.datatype))
		if readErr == nil {
			readValues[register.name] = 0xBEEF
			err = readErr
		} else {
			readValues[register.name] = decode(resp, register)
		}
	}
	return readValues, err
}

func (m Poller) Write(registers []Register, writeValues map[string]float64) error {
	err := m.handler.Connect()
	if err != nil {
		return err
	}
	defer m.handler.Close()

	client := modbus.NewClient(m.handler)
	for name, val := range writeValues {
		i, writeErr := findRegisterByName(registers, name)
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

func findRegisterByName(registers []Register, name string) (int, error) {
	for index, register := range registers {
		if register.name == name {
			return index, nil
		}
	}
	return 0, errors.New("register name not found in register array")
}

// encode convert a float64 into a byte array
func encode(val float64, register Register) []byte {
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
func decode(bytes []byte, register Register) float64 {
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
