package modbuscomm

// ModbusComm interface
type ModbusComm interface {
	Read([]Register) (map[string]float64, error)
	Write([]Register, map[string]float64) error
}

// DataType defines the type of Modbus register for encoding/decoding
type DataType string

// Constants of DataType
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

// Access devices the register read/write type
type Access string

const (
	ro = "read-only"
	wo = "write-only"
	rw = "read-write"
)

// Endian byte order of Modbus register for encoding/decoding
type Endian string

// Constants of Endian
const (
	littleEndian Endian = "little"
	bigEndian    Endian = "big"
)

// Register contains the data required to read and write a Modbus register
type Register struct {
	Name         string   `json:"Name"`
	Address      uint16   `json:"Address"`
	DataType     DataType `json:"DataType"`
	FunctionCode int      `json:"FunctionCode"`
	AccessType   Access   `json:"Access"`
	Endianness   Endian   `json:"Endianness"`
}

// FilterRegisters returns registers from array with matching access type
func FilterRegisters(r []Register, a Access) []Register {
	filtered := make([]Register, 0)
	for _, reg := range r {
		if reg.AccessType == a || reg.AccessType == "read-write" {
			filtered = append(filtered, reg)
		}
	}
	return filtered
}
