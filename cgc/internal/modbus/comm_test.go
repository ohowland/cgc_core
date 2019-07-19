package modbus

import (
	"log"
	"math"
	"math/rand"
	"os"
	"testing"
	"time"

	"github.com/goburrow/modbus"
	"gotest.tools/assert"
)

// Encode-Decode U64
func TestEncodeU64Big(t *testing.T) {
	testReg := Register{"test", 0, u64, 3, bigEndian}
	var testVal float64 = 1234
	bytes := encode(testVal, testReg)
	t.Logf("float64: [%v] to U64 to big-endian []bytes: %v", testVal, bytes)

	assertBytes := [8]byte{0, 0, 0, 0, 0, 0, 4, 210}
	assert.Assert(t, bytes != nil)
	assert.Assert(t, len(bytes) == len(assertBytes[:]))
	for i := range bytes {
		assert.Assert(t, bytes[i] == assertBytes[i])
	}
}

func TestDecodeU64Big(t *testing.T) {
	testReg := Register{"test", 0, u64, 3, bigEndian}
	assertVal := rand.Float64() * 9223372036854775807
	testBytes := encode(assertVal, testReg)
	testVal := decode(testBytes[:], testReg)
	t.Logf("[]bytes: [%v] U64 big-endian to float64: [%v]", testBytes, testVal)

	assert.Assert(t, testVal == math.Floor(assertVal))
}

func TestEncodeU64Little(t *testing.T) {
	testReg := Register{"test", 0, u64, 3, littleEndian}
	var testVal float64 = 1234
	bytes := encode(testVal, testReg)
	t.Logf("float64: [%v] to U64 to little-endian []bytes: %v", testVal, bytes)

	assertBytes := [8]byte{210, 4, 0, 0, 0, 0, 0, 0}
	assert.Assert(t, bytes != nil)
	assert.Assert(t, len(bytes) == len(assertBytes[:]))
	for i := range bytes {
		assert.Assert(t, bytes[i] == assertBytes[i])
	}
}

func TestDecodeU64Little(t *testing.T) {
	testReg := Register{"test", 0, u64, 3, littleEndian}
	assertVal := rand.Float64() * 9223372036854775807
	testBytes := encode(assertVal, testReg)
	testVal := decode(testBytes[:], testReg)
	t.Logf("[]bytes: [%v] U64 little-endian to float64: [%v]", testBytes, testVal)

	assert.Assert(t, testVal == math.Floor(assertVal))
}

// Encode-Decode U32
func TestEncodeU32Big(t *testing.T) {
	testReg := Register{"test", 0, u32, 3, bigEndian}
	var testVal float64 = 1234
	bytes := encode(testVal, testReg)
	t.Logf("float64: [%v] to U32 to big-endian []bytes: %v", testVal, bytes)

	assertBytes := [4]byte{0, 0, 4, 210}
	assert.Assert(t, bytes != nil)
	assert.Assert(t, len(bytes) == len(assertBytes[:]))
	for i := range bytes {
		assert.Assert(t, bytes[i] == assertBytes[i])
	}
}

func TestDecodeU32Big(t *testing.T) {
	testReg := Register{"test", 0, u32, 3, bigEndian}
	assertVal := rand.Float64() * 4294967295
	testBytes := encode(assertVal, testReg)
	testVal := decode(testBytes[:], testReg)
	t.Logf("[]bytes: [%v] U64 big-endian to float64: [%v]", testBytes, testVal)

	assert.Assert(t, testVal == math.Floor(assertVal))
}

func TestEncodeU32Little(t *testing.T) {
	testReg := Register{"test", 0, u32, 3, littleEndian}
	var testVal float64 = 1234
	bytes := encode(testVal, testReg)
	t.Logf("float64: [%v] to U32 to little-endian []bytes: %v", testVal, bytes)

	assertBytes := [4]byte{210, 4, 0, 0}
	assert.Assert(t, bytes != nil)
	assert.Assert(t, len(bytes) == len(assertBytes[:]))
	for i := range bytes {
		assert.Assert(t, bytes[i] == assertBytes[i])
	}
}

func TestDecodeU32Little(t *testing.T) {
	testReg := Register{"test", 0, u32, 3, littleEndian}
	assertVal := rand.Float64() * 4294967295
	testBytes := encode(assertVal, testReg)
	testVal := decode(testBytes[:], testReg)
	t.Logf("[]bytes: [%v] U64 big-endian to float64: [%v]", testBytes, testVal)

	assert.Assert(t, testVal == math.Floor(assertVal))
}

// Encode-Decode U16
func TestEncodeU16Big(t *testing.T) {
	testReg := Register{"test", 0, u16, 3, bigEndian}
	var testVal float64 = 1234
	bytes := encode(testVal, testReg)
	t.Logf("float64: [%v] to U16 to big-endian []bytes: %v", testVal, bytes)

	assertBytes := [2]byte{4, 210}
	assert.Assert(t, bytes != nil)
	assert.Assert(t, len(bytes) == len(assertBytes[:]))
	for i := range bytes {
		assert.Assert(t, bytes[i] == assertBytes[i])
	}
}

func TestDecodeU16Big(t *testing.T) {
	testReg := Register{"test", 0, u16, 3, bigEndian}
	assertVal := rand.Float64() * 65535
	testBytes := encode(assertVal, testReg)
	testVal := decode(testBytes[:], testReg)
	t.Logf("[]bytes: [%v] U16 big-endian to float64: [%v]", testBytes, testVal)

	assert.Assert(t, testVal == math.Floor(assertVal))
}

func TestEncodeU16Little(t *testing.T) {
	testReg := Register{"test", 0, u16, 3, littleEndian}
	var testVal float64 = 1234
	bytes := encode(testVal, testReg)
	t.Logf("float64: [%v] to U16 to little-endian []bytes: %v", testVal, bytes)

	assertBytes := [2]byte{210, 4}
	assert.Assert(t, bytes != nil)
	assert.Assert(t, len(bytes) == len(assertBytes[:]))
	for i := range bytes {
		assert.Assert(t, bytes[i] == assertBytes[i])
	}
}

func TestDecodeU16Little(t *testing.T) {
	testReg := Register{"test", 0, u16, 3, littleEndian}
	assertVal := rand.Float64() * 65535
	testBytes := encode(assertVal, testReg)
	testVal := decode(testBytes[:], testReg)
	t.Logf("[]bytes: [%v] U16 big-endian to float64: [%v]", testBytes, testVal)

	assert.Assert(t, testVal == math.Floor(assertVal))
}

// Encode-Decode I64
func TestEncodeI64Big(t *testing.T) {
	testReg := Register{"test", 0, i64, 3, bigEndian}
	var testVal float64 = 1234
	bytes := encode(testVal, testReg)
	t.Logf("float64: [%v] to I64 to big-endian []bytes: %v", testVal, bytes)

	assertBytes := [8]byte{0, 0, 0, 0, 0, 0, 4, 210}
	assert.Assert(t, bytes != nil)
	assert.Assert(t, len(bytes) == len(assertBytes[:]))
	for i := range bytes {
		assert.Assert(t, bytes[i] == assertBytes[i])
	}
}

func TestDecodeI64Big(t *testing.T) {
	testReg := Register{"test", 0, i64, 3, bigEndian}
	assertVal := rand.Float64() * -9223372036854775807
	testBytes := encode(assertVal, testReg)
	testVal := decode(testBytes[:], testReg)
	t.Logf("[]bytes: [%v] I64 big-endian to float64: [%v]", testBytes, testVal)

	assert.Assert(t, testVal == math.Floor(assertVal))
}

func TestEncodeI64Little(t *testing.T) {
	testReg := Register{"test", 0, i64, 3, littleEndian}
	var testVal float64 = 1234
	bytes := encode(testVal, testReg)
	t.Logf("float64: [%v] to I64 to little-endian []bytes: %v", testVal, bytes)

	assertBytes := [8]byte{210, 4, 0, 0, 0, 0, 0, 0}
	assert.Assert(t, bytes != nil)
	assert.Assert(t, len(bytes) == len(assertBytes[:]))
	for i := range bytes {
		assert.Assert(t, bytes[i] == assertBytes[i])
	}
}

func TestDecodeI64Little(t *testing.T) {
	testReg := Register{"test", 0, i64, 3, littleEndian}
	assertVal := rand.Float64() * -9223372036854775807
	testBytes := encode(assertVal, testReg)
	testVal := decode(testBytes[:], testReg)
	t.Logf("[]bytes: [%v] I64 little-endian to float64: [%v]", testBytes, testVal)

	assert.Assert(t, testVal == math.Floor(assertVal))
}

// Encode-Decode I32
func TestEncodeI32Big(t *testing.T) {
	testReg := Register{"test", 0, i32, 3, bigEndian}
	var testVal float64 = -1234
	bytes := encode(testVal, testReg)
	t.Logf("float64: [%v] to I32 to big-endian []bytes: %v", testVal, bytes)

	assertBytes := [4]byte{255, 255, 251, 46}
	assert.Assert(t, bytes != nil)
	assert.Assert(t, len(bytes) == len(assertBytes[:]))
	for i := range bytes {
		assert.Assert(t, bytes[i] == assertBytes[i])
	}
}

func TestDecodeI32Big(t *testing.T) {
	testReg := Register{"test", 0, i32, 3, bigEndian}
	assertVal := rand.Float64() * -2147483647
	testBytes := encode(assertVal, testReg)
	testVal := decode(testBytes[:], testReg)
	t.Logf("[]bytes: [%v] I32 big-endian to float64: [%v]", testBytes, testVal)

	assert.Assert(t, testVal == math.Ceil(assertVal))
}

func TestEncodeI32Little(t *testing.T) {
	testReg := Register{"test", 0, i32, 3, littleEndian}
	var testVal float64 = -1234
	bytes := encode(testVal, testReg)
	t.Logf("float64: [%v] to I32 to little-endian []bytes: %v", testVal, bytes)

	assertBytes := [4]byte{46, 251, 255, 255}
	assert.Assert(t, bytes != nil)
	assert.Assert(t, len(bytes) == len(assertBytes[:]))
	for i := range bytes {
		assert.Assert(t, bytes[i] == assertBytes[i])
	}
}

func TestDecodeI32Little(t *testing.T) {
	testReg := Register{"test", 0, i32, 3, littleEndian}
	assertVal := rand.Float64() * -2147483647
	testBytes := encode(assertVal, testReg)
	testVal := decode(testBytes[:], testReg)
	t.Logf("[]bytes: [%v] I32 big-endian to float64: [%v]", testBytes, testVal)

	assert.Assert(t, testVal == math.Ceil(assertVal))
}

// Encode-Decode I16
func TestEncodeI16Big(t *testing.T) {
	testReg := Register{"test", 0, i16, 3, bigEndian}
	var testVal float64 = -1234
	bytes := encode(testVal, testReg)
	t.Logf("float64: [%v] to I16 to big-endian []bytes: %v", testVal, bytes)

	assertBytes := [2]byte{251, 46}
	assert.Assert(t, bytes != nil)
	assert.Assert(t, len(bytes) == len(assertBytes[:]))
	for i := range bytes {
		assert.Assert(t, bytes[i] == assertBytes[i])
	}
}

func TestDecodeI16Big(t *testing.T) {
	testReg := Register{"test", 0, i16, 3, bigEndian}
	assertVal := rand.Float64() * -32767
	testBytes := encode(assertVal, testReg)
	testVal := decode(testBytes[:], testReg)
	t.Logf("[]bytes: [%v] I16 big-endian to float64: [%v]", testBytes, testVal)

	assert.Assert(t, testVal == math.Ceil(assertVal))
}

func TestEncodeI16Little(t *testing.T) {
	testReg := Register{"test", 0, i16, 3, littleEndian}
	var testVal float64 = -1234
	bytes := encode(testVal, testReg)
	t.Logf("float64: [%v] to I16 to little-endian []bytes: %v", testVal, bytes)

	assertBytes := [2]byte{46, 251}
	assert.Assert(t, bytes != nil)
	assert.Assert(t, len(bytes) == len(assertBytes[:]))
	for i := range bytes {
		assert.Assert(t, bytes[i] == assertBytes[i])
	}
}

func TestDecodeI16Little(t *testing.T) {
	testReg := Register{"test", 0, i16, 3, littleEndian}
	assertVal := rand.Float64() * -32767
	testBytes := encode(assertVal, testReg)
	testVal := decode(testBytes[:], testReg)
	t.Logf("[]bytes: [%v] I16 big-endian to float64: [%v]", testBytes, testVal)

	assert.Assert(t, testVal == math.Ceil(assertVal))
}

func TestModbusPoller(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestModbusPoller in short mode")
	}
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
