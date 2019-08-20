package modbuscomm

import (
	"fmt"
	"math"
	"math/rand"
	"testing"

	"gotest.tools/assert"
)

// Encode-Decode U64
func TestEncodeU64Big(t *testing.T) {
	rand.Seed(10)
	testReg := Register{"test", 0, u64, 3, ro, bigEndian}
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
	testReg := Register{"test", 0, u64, 3, ro, bigEndian}
	assertVal := rand.Float64() * 9223372036854775807
	testBytes := encode(assertVal, testReg)
	testVal := decode(testBytes[:], testReg)
	t.Logf("[]bytes: [%v] U64 big-endian to float64: [%v]", testBytes, testVal)

	assert.Assert(t, testVal == math.Floor(assertVal))
}

func TestEncodeU64Little(t *testing.T) {
	testReg := Register{"test", 0, u64, 3, ro, littleEndian}
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
	testReg := Register{"test", 0, u64, 3, ro, littleEndian}
	assertVal := rand.Float64() * 9223372036854775807
	testBytes := encode(assertVal, testReg)
	testVal := decode(testBytes[:], testReg)
	t.Logf("[]bytes: [%v] U64 little-endian to float64: [%v]", testBytes, testVal)

	assert.Assert(t, testVal == math.Floor(assertVal))
}

// Encode-Decode U32
func TestEncodeU32Big(t *testing.T) {
	testReg := Register{"test", 0, u32, 3, ro, bigEndian}
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
	testReg := Register{"test", 0, u32, 3, ro, bigEndian}
	assertVal := rand.Float64() * 4294967295
	testBytes := encode(assertVal, testReg)
	testVal := decode(testBytes[:], testReg)
	t.Logf("[]bytes: [%v] U64 big-endian to float64: [%v]", testBytes, testVal)

	assert.Assert(t, testVal == math.Floor(assertVal))
}

func TestEncodeU32Little(t *testing.T) {
	testReg := Register{"test", 0, u32, 3, ro, littleEndian}
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
	testReg := Register{"test", 0, u32, 3, ro, littleEndian}
	assertVal := rand.Float64() * 4294967295
	testBytes := encode(assertVal, testReg)
	testVal := decode(testBytes[:], testReg)
	t.Logf("[]bytes: [%v] U64 big-endian to float64: [%v]", testBytes, testVal)

	assert.Assert(t, testVal == math.Floor(assertVal))
}

// Encode-Decode U16
func TestEncodeU16Big(t *testing.T) {
	testReg := Register{"test", 0, u16, 3, ro, bigEndian}
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
	testReg := Register{"test", 0, u16, 3, ro, bigEndian}
	assertVal := rand.Float64() * 65535
	testBytes := encode(assertVal, testReg)
	testVal := decode(testBytes[:], testReg)
	t.Logf("[]bytes: [%v] U16 big-endian to float64: [%v]", testBytes, testVal)

	assert.Assert(t, testVal == math.Floor(assertVal))
}

func TestEncodeU16Little(t *testing.T) {
	testReg := Register{"test", 0, u16, 3, ro, littleEndian}
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
	testReg := Register{"test", 0, u16, 3, ro, littleEndian}
	assertVal := rand.Float64() * 65535
	testBytes := encode(assertVal, testReg)
	testVal := decode(testBytes[:], testReg)
	t.Logf("[]bytes: [%v] U16 big-endian to float64: [%v]", testBytes, testVal)

	assert.Assert(t, testVal == math.Floor(assertVal))
}

// Encode-Decode I64
func TestEncodeI64Big(t *testing.T) {
	testReg := Register{"test", 0, i64, 3, ro, bigEndian}
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
	testReg := Register{"test", 0, i64, 3, ro, bigEndian}
	assertVal := rand.Float64() * -9223372036854775807
	testBytes := encode(assertVal, testReg)
	testVal := decode(testBytes[:], testReg)
	t.Logf("[]bytes: [%v] I64 big-endian to float64: [%v]", testBytes, testVal)

	assert.Assert(t, testVal == math.Floor(assertVal))
}

func TestEncodeI64Little(t *testing.T) {
	testReg := Register{"test", 0, i64, 3, ro, littleEndian}
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
	testReg := Register{"test", 0, i64, 3, ro, littleEndian}
	assertVal := rand.Float64() * -9223372036854775807
	testBytes := encode(assertVal, testReg)
	testVal := decode(testBytes[:], testReg)
	t.Logf("[]bytes: [%v] I64 little-endian to float64: [%v]", testBytes, testVal)

	assert.Assert(t, testVal == math.Floor(assertVal))
}

// Encode-Decode I32
func TestEncodeI32Big(t *testing.T) {
	testReg := Register{"test", 0, i32, 3, ro, bigEndian}
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
	testReg := Register{"test", 0, i32, 3, ro, bigEndian}
	assertVal := rand.Float64() * -2147483647
	testBytes := encode(assertVal, testReg)
	testVal := decode(testBytes[:], testReg)
	t.Logf("[]bytes: [%v] I32 big-endian to float64: [%v]", testBytes, testVal)

	assert.Assert(t, testVal == math.Ceil(assertVal))
}

func TestEncodeI32Little(t *testing.T) {
	testReg := Register{"test", 0, i32, 3, ro, littleEndian}
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
	testReg := Register{"test", 0, i32, 3, ro, littleEndian}
	assertVal := rand.Float64() * -2147483647
	testBytes := encode(assertVal, testReg)
	testVal := decode(testBytes[:], testReg)
	t.Logf("[]bytes: [%v] I32 big-endian to float64: [%v]", testBytes, testVal)

	assert.Assert(t, testVal == math.Ceil(assertVal))
}

// Encode-Decode I16
func TestEncodeI16Big(t *testing.T) {
	testReg := Register{"test", 0, i16, 3, ro, bigEndian}
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
	testReg := Register{"test", 0, i16, 3, ro, bigEndian}
	assertVal := rand.Float64() * -32767
	testBytes := encode(assertVal, testReg)
	testVal := decode(testBytes[:], testReg)
	t.Logf("[]bytes: [%v] I16 big-endian to float64: [%v]", testBytes, testVal)

	assert.Assert(t, testVal == math.Ceil(assertVal))
}

func TestEncodeI16Little(t *testing.T) {
	testReg := Register{"test", 0, i16, 3, ro, littleEndian}
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

// encode-decode Float32
func TestEncodeF32Big(t *testing.T) {
	testReg := Register{"test", 0, f32, 3, ro, bigEndian}
	var testVal float64 = -1234
	bytes := encode(testVal, testReg)
	t.Logf("float64: [%v] to F32 to big-endian []bytes: %v", testVal, bytes)

	assertBytes := [4]byte{196, 154, 64, 0}
	assert.Assert(t, bytes != nil)
	assert.Assert(t, len(bytes) == len(assertBytes[:]))
	for i := range bytes {
		assert.Assert(t, bytes[i] == assertBytes[i])
	}
}

func TestDecodeF32Big(t *testing.T) {
	testReg := Register{"test", 0, f32, 3, ro, bigEndian}
	assertVal := rand.Float64() * -32767
	testBytes := encode(assertVal, testReg)
	testVal := decode(testBytes[:], testReg)
	t.Logf("[]bytes: %v F32 big-endian to float64: [%v]", testBytes, testVal)

	assert.Assert(t, math.Floor(testVal) == math.Floor(assertVal))
}

// encode-decode Float64
func TestEncodeF64Big(t *testing.T) {
	testReg := Register{"test", 0, f64, 3, ro, bigEndian}
	var testVal float64 = -1234
	bytes := encode(testVal, testReg)
	t.Logf("float64: [%v] to F64 to big-endian []bytes: %v", testVal, bytes)

	assertBytes := [8]byte{192, 147, 72, 0, 0, 0, 0, 0}
	assert.Assert(t, bytes != nil)
	assert.Assert(t, len(bytes) == len(assertBytes[:]))
	for i := range bytes {
		assert.Assert(t, bytes[i] == assertBytes[i])
	}
}

func TestDecodeF64Big(t *testing.T) {
	testReg := Register{"test", 0, f64, 3, ro, bigEndian}
	assertVal := rand.Float64() * -32767
	testBytes := encode(assertVal, testReg)
	testVal := decode(testBytes[:], testReg)
	t.Logf("[]bytes: [%v] F64 big-endian to float64: [%v]", testBytes, testVal)

	assert.Assert(t, testVal == assertVal)
}

func TestDecodeI16Little(t *testing.T) {
	testReg := Register{"test", 0, i16, 3, ro, littleEndian}
	assertVal := rand.Float64() * -32767
	testBytes := encode(assertVal, testReg)
	testVal := decode(testBytes[:], testReg)
	t.Logf("[]bytes: [%v] I16 big-endian to float64: [%v]", testBytes, testVal)

	assert.Assert(t, testVal == math.Ceil(assertVal))
}

func TestFindRegisterByName(t *testing.T) {
	testReg1 := Register{"test1", 0, u16, 3, ro, bigEndian}
	testReg2 := Register{"test2", 1, u32, 3, ro, bigEndian}
	testReg3 := Register{"test3", 3, u64, 3, ro, bigEndian}
	testRegs := []Register{testReg1, testReg2, testReg3}

	i, err := findIndexByName(testRegs, "test2")

	assert.Assert(t, err == nil)
	assert.Assert(t, testRegs[i].Name == "test2")
	assert.Assert(t, testRegs[i].Address == 1)
	assert.Assert(t, testRegs[i].DataType == u32)
	assert.Assert(t, testRegs[i].FunctionCode == 3)
	assert.Assert(t, testRegs[i].AccessType == ro)
	assert.Assert(t, testRegs[i].Endianness == bigEndian)
}

func TestFindRegisterByNameFail(t *testing.T) {
	testReg1 := Register{"test1", 0, u16, 3, wo, bigEndian}
	testReg2 := Register{"test2", 1, u32, 3, wo, bigEndian}
	testReg3 := Register{"test3", 3, u64, 3, wo, bigEndian}
	testRegs := []Register{testReg1, testReg2, testReg3}

	i, err := findIndexByName(testRegs, "test42")
	assert.Assert(t, err.Error() == "register name not found in register array")
	assert.Assert(t, i == -1)
}

func TestPoller(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestModbusPoller in short mode")
	}

	pollerConfig := PollerConfig{"192.168.0.100", "5020", 0x01, 100, 500, true}

	reg1 := Register{"test1", 0, u16, 3, rw, bigEndian}
	reg2 := Register{"test2", 1, u16, 3, rw, bigEndian}
	reg3 := Register{"test3", 2, u16, 3, rw, bigEndian}
	regs := []Register{reg1, reg2, reg3}

	poller := NewPoller(pollerConfig)

	resp, err := poller.Read(regs)
	t.Logf("\nresponse: %v\n error: %v", resp, err)
	assert.Assert(t, err == nil)
}

func TestPollerFailOnTimeout(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping TestModbusPollerBadAddress in short mode")
	}

	testIP := "1.1.1.1"
	testPort := "123"

	pollerConfig := PollerConfig{testIP, testPort, 0x01, 100, 500, true}

	reg1 := Register{"test1", 0, u16, 3, ro, bigEndian}
	reg2 := Register{"test2", 1, u16, 3, ro, bigEndian}
	reg3 := Register{"test3", 2, u16, 3, ro, bigEndian}
	regs := []Register{reg1, reg2, reg3}

	poller := NewPoller(pollerConfig)

	_, err := poller.Read(regs)
	assert.Assert(t, err.Error() == fmt.Sprintf("dial tcp %v:%v: i/o timeout", testIP, testPort))
}
