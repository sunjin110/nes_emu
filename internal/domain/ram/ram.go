package ram

// ファミコン（NES）のCPU（Ricoh 2A03/2A07）は、次のように設計されています：
// 0x0000〜0x07FF: 内蔵RAM（Work RAM）、サイズは 2KB（2048バイト）
// 0x0800〜0x1FFF: 内蔵RAMのミラーリング領域（ハードウェアでミラーリング）
// Sample RAM map: https://www.nesdev.org/wiki/Sample_RAM_map

// 0x0100 ~ 0x01FFはスタックに割り当てられる
// 多くのゲームではSPの初期値として0x1FFを指定する
// pushした時に、追加してSPをdecrement
// popした時に、取り出してSPをincrement

const (
	ramSize = 2 * 1024 // 2KB
)

const (
	addrRAMStart       = 0x0000
	addrRAMMirrorStart = 0x0800
	addrRAMEnd         = 0x1FFF
)

type RAM struct {
	data [ramSize]byte
}

func NewRAM() *RAM {
	return &RAM{
		data: [ramSize]byte{},
	}
}

func (ram *RAM) Read(offset uint16) byte {
	return ram.data[offset%ramSize]
}

func (ram *RAM) Write(offset uint16, data byte) {
	ram.data[offset%ramSize] = data
}

func IsRAMRange(addr uint16) bool {
	return addr >= addrRAMStart && addr <= addrRAMEnd
}
