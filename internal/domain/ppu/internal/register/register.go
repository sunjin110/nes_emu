package register

type Register interface {
	SetCoarseX(target registerTarget, data byte)
	SetCoarseY(target registerTarget, data byte)
	SetNametableSelect(target registerTarget, data byte)
	SetFineY(target registerTarget, data byte)
	SetFindX(data byte)
	SetW(data wData) // 0 or 1

	// PPUADDR 反映用(紛らわしいけど、 PPUADDR としては使わない、PPUADDR への書き込みと PPUSCROLL への書き込みを混ぜて使ってるゲームのため)
	SetUpperPPUAddr(data byte)
	SetLowerPPUAddr(data byte)

	GetCoarseX(target registerTarget) byte
	GetCoarseY(target registerTarget) byte
	GetNametableSelect(target registerTarget) byte
	GetFineY(target registerTarget) byte
	GetFindX() byte
	GetW() wData

	// 描画中のインクリメント
	IncrementCoarseX()
	IncrementY()

	// 現在のタイルと attribute table のアドレス取得
	GetTileAddress() uint16
	GetAttributeAddress() uint16

	// t の変更を v に反映
	UpdateHorizontalV()
	UpdateVerticalV()
}

type registerTarget int

const (
	TargetV registerTarget = iota
	TargetT
)

type wData bool

const (
	WData0 wData = false
	WData1 wData = true
)

func NewRegister() Register {
	return &register{}
}

// register ppuのための内部register
type register struct {
	v uint16 // 現在参照するVRAMのアドレス 15bit
	t uint16 // temporary VRAM アドレス(v に書き込む値を構築するために使用) 15bit
	x uint8  // fine X scroll 3bit
	w bool   // PPUSCROLL, PPUADDR の書き込みが1回目なのか2回目なのかを判定するためのフラグ
}

// GetAttributeAddress implements Register.
func (r *register) GetAttributeAddress() uint16 {
	panic("unimplemented")
}

// GetCoarseX implements Register.
func (r *register) GetCoarseX(target registerTarget) byte {
	panic("unimplemented")
}

// GetCoarseY implements Register.
func (r *register) GetCoarseY(target registerTarget) byte {
	panic("unimplemented")
}

// GetFindX implements Register.
func (r *register) GetFindX() byte {
	panic("unimplemented")
}

// GetFineY implements Register.
func (r *register) GetFineY(target registerTarget) byte {
	panic("unimplemented")
}

// GetNametableSelect implements Register.
func (r *register) GetNametableSelect(target registerTarget) byte {
	panic("unimplemented")
}

// GetTileAddress implements Register.
func (r *register) GetTileAddress() uint16 {
	panic("unimplemented")
}

// GetW implements Register.
func (r *register) GetW() wData {
	panic("unimplemented")
}

// IncrementCoarseX implements Register.
func (r *register) IncrementCoarseX() {
	panic("unimplemented")
}

// IncrementY implements Register.
func (r *register) IncrementY() {
	panic("unimplemented")
}

// SetCoarseX implements Register.
func (r *register) SetCoarseX(target registerTarget, data byte) {
	panic("unimplemented")
}

// SetCoarseY implements Register.
func (r *register) SetCoarseY(target registerTarget, data byte) {
	panic("unimplemented")
}

// SetFindX implements Register.
func (r *register) SetFindX(data byte) {
	panic("unimplemented")
}

// SetFineY implements Register.
func (r *register) SetFineY(target registerTarget, data byte) {
	panic("unimplemented")
}

// SetLowerPPUAddr implements Register.
func (r *register) SetLowerPPUAddr(data byte) {
	panic("unimplemented")
}

// SetNametableSelect implements Register.
func (r *register) SetNametableSelect(target registerTarget, data byte) {
	panic("unimplemented")
}

// SetUpperPPUAddr implements Register.
func (r *register) SetUpperPPUAddr(data byte) {
	panic("unimplemented")
}

// SetW implements Register.
func (r *register) SetW(data wData) {
	panic("unimplemented")
}

// UpdateHorizontalV implements Register.
func (r *register) UpdateHorizontalV() {
	panic("unimplemented")
}

// UpdateVerticalV implements Register.
func (r *register) UpdateVerticalV() {
	panic("unimplemented")
}
