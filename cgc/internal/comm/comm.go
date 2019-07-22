package comm

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

// Endian byte order of register
type Endian string

const (
	littleEndian Endian = "little"
	bigEndian    Endian = "big"
)

// ModbusComm interface
type ModbusComm interface {
	Read([]Register) (map[string]float64, error)
	Write([]Register, map[string]float64) error
}

// Register contains the data required to read and write a Modbus register
type Register struct {
	name         string
	address      uint16
	datatype     DataType
	functionCode int
	endianness   Endian
}
