package memory

type Memory interface {
	Read(addr uint16) (byte, error)
	Write(addr uint16, value byte) error
}

// TODO PPUのメモリ構成を考える
type memory struct {
}
