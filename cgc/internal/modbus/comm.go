package cgc

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
	i16 DataType = "i16"
	i32 DataType = "i32"
	f32 DataType = "f32"
	f64 DataType = "f64"
)

// ModbusRegister contains the data required to read and write a Modbus register
type ModbusRegister struct {
	name         string
	address      uint16
	datatype     DataType
	functionCode int
	endianness   Endian
}

// ModbusPollerConfig is the configuration format for ModbusPoller
type ModbusPollerConfig struct {
	ipAddr   string
	port     string
	timeout  int
	pollRate int
	offset   int
}

// ModbusPoller continiously polls a target
type ModbusPoller struct {
	handler  *modbus.TCPClientHandler
	pollRate int
	offset   int
}

// NewModbusPoller is a factory for the ModbusPoller struct
func NewModbusPoller(cfg ModbusPollerConfig) ModbusPoller {
	handler := modbus.NewTCPClientHandler(cfg.ipAddr + ":" + cfg.port)
	handler.Timeout = time.Millisecond * time.Duration(cfg.timeout)
	return ModbusPoller{
		handler:  handler,
		pollRate: cfg.pollRate,
	}
}

func (m ModbusPoller) Read(registers []ModbusRegister) (map[string]float64, error) {
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

func (m ModbusPoller) Write(registers []ModbusRegister, writeValues map[string]float64) error {
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

func findRegisterByName(registers []ModbusRegister, name string) (int, error) {
	for index, register := range registers {
		if register.name == name {
			return index, nil
		}
	}
	return 0, errors.New("register name not found in register array")
}

// encode convert a float64 into a byte array
func encode(val float64, register ModbusRegister) []byte {
	var bytes []byte
	endian := getByteOrder(register.endianness)
	switch register.datatype {
	case u16, i16:
		coercedVal := uint16(val)
		bytes = make([]byte, 2*sizeOf(u16))
		endian.PutUint16(bytes, coercedVal)
	case u32, i32, f32:
		coercedVal := uint32(val)
		bytes = make([]byte, 2*sizeOf(u32))
		endian.PutUint32(bytes, coercedVal)
	case f64:
		coercedVal := uint64(val)
		bytes = make([]byte, 2*sizeOf(f64))
		endian.PutUint64(bytes, coercedVal)
	}
	return bytes
}

// decode coverts byte arrays into float64s
func decode(bytes []byte, register ModbusRegister) float64 {
	var bits uint64
	endian := getByteOrder(register.endianness)
	switch register.datatype {
	case u16, i16:
		bits = uint64(endian.Uint16(bytes))
	case u32, i32, f32:
		bits = uint64(endian.Uint32(bytes))
	case f64:
		bits = endian.Uint64(bytes)
	}
	return math.Float64frombits(bits)
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
	case f64:
		return 2
	}
	return 0
}
