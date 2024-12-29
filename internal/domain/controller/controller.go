package controller

import "github.com/sunjin110/nes_emu/pkg/logger"

type Controller struct {
	state [2]byte // 0: 1P, 1: 2P
}

func NewController() *Controller {
	return &Controller{
		state: [2]byte{},
	}
}

const (
	addrController1P = 0x4016
	addrController2P = 0x4017
)

func (c *Controller) Read(addr uint16) byte {
	switch addr {
	case addrController1P:
		return c.state[0]
	case addrController2P:
		return c.state[1]
	}
	logger.Logger.Error("Controller: Read: invalid addr", "addr", addr)
	return 0xFF
}

func (c *Controller) Write(addr uint16, value byte) {
	switch addr {
	case addrController1P:
		c.state[0] = value
	case addrController2P:
		c.state[1] = value
	default:
		logger.Logger.Error("Controller: Write: invalid addr", "addr", addr)
	}
}

func IsControllerAddr(addr uint16) bool {
	return addr == addrController1P || addr == addrController2P
}
