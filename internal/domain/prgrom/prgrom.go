package prgrom

type PRGROM interface {
	Read(addr uint16) byte
	InitPC() uint16
}
