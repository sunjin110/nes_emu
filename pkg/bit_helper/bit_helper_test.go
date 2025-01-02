package bit_helper_test

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/sunjin110/nes_emu/pkg/bit_helper"
)

// go test -v -count=1 -timeout 30s -run ^TestUint16ToBytes$ github.com/sunjin110/nes_emu/pkg/bit_helper
func TestUint16ToBytes(t *testing.T) {
	Convey("TestUint16ToBytes", t, func() {
		lower, upper := bit_helper.Uint16ToBytes(0x1234)
		So(upper, ShouldEqual, 0x12)
		So(lower, ShouldEqual, 0x34)
	})
}
