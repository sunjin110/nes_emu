package cpu_test

import (
	"strconv"
	"strings"
	"testing"

	_ "embed"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/sunjin110/nes_emu/internal/domain/cpu"
)

var (
	//go:embed testdata/opcode_byte.text
	opcodeByteStr string
)

// go test -v -count=1 -timeout 30s -run ^TestOpcodes$ github.com/sunjin110/nes_emu/internal/domain/cpu
func TestOpcodes(t *testing.T) {
	Convey("TestOpcodes", t, func() {
		opcodeBytes := strings.Split(opcodeByteStr, "\n")
		for _, opcodeByte := range opcodeBytes {
			trimmedStr := opcodeByte[2:]
			value, err := strconv.ParseUint(trimmedStr, 16, 8)
			So(err, ShouldBeNil)
			_, ok := cpu.Opcodes[byte(value)]
			So(ok, ShouldBeTrue)
		}
	})
}
