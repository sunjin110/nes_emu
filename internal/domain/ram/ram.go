package ram

// ファミコン（NES）のCPU（Ricoh 2A03/2A07）は、次のように設計されています：
// 0x0000〜0x07FF: 内蔵RAM（Work RAM）、サイズは 2KB（2048バイト）
// 0x0800〜0x1FFF: 内蔵RAMのミラーリング領域（ハードウェアでミラーリング）
// Sample RAM map: https://www.nesdev.org/wiki/Sample_RAM_map

const (
	ramSize = 2 * 1024 // 2KB
)

type RAM struct {
	data [ramSize]byte
}

func NewRAM() *RAM {
	return &RAM{
		data: [ramSize]byte{},
	}
}

func (ram *RAM) Read(offset int) byte {
	return ram.data[offset%ramSize]
}

func (ram *RAM) Write(offset int, data byte) {
	ram.data[offset%ramSize] = data
}
