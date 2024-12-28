package cartridge_test

import (
	"testing"

	_ "embed"

	. "github.com/smartystreets/goconvey/convey"
	"github.com/sunjin110/nes_emu/internal/domain/cartridge"
)

var (
	//go:embed testdata/hello.nes
	helloNesRom []byte
)

// go test -v -count=1 -timeout 30s -run ^TestNewCartridge$ github.com/sunjin110/nes_emu/internal/domain/cartridge
func TestNewCartridge(t *testing.T) {
	Convey("TestNewCartridge", t, func() {
		type test struct {
			name    string
			data    []byte
			want    *cartridge.Cartridge
			wantErr error
		}

		tests := []test{
			{
				name: "nesのromが読み込めること",
				data: helloNesRom,
				want: &cartridge.Cartridge{
					PRG:          helloNesRom[16 : 16+2*16*1024],
					CHR:          helloNesRom[16+2*16*1024 : (16+2*16*1024)+8*1024],
					MapperNo:     0,
					PRGBankCount: 2,
					CHRBankCount: 1,
				},
			},
		}

		for _, tt := range tests {
			Convey(tt.name, func() {
				got, err := cartridge.NewCartridge(tt.data)
				if tt.wantErr != nil {
					So(err, ShouldBeError)
					So(err.Error(), ShouldEqual, tt.wantErr.Error())
					return
				}
				So(err, ShouldBeNil)
				So(got, ShouldResemble, tt.want)
			})
		}
	})
}
