package m5stickc

import "machine"

// Button は M5StickC Plus2 のボタン（active-low）を表す。
type Button struct {
	pin machine.Pin
}

// NewButton は指定ピンを入力に設定して Button を返す。
// 内部プルアップは無いが基板側に外部プルアップがあるため PinInput でよい。
//
// 例: btnA := m5stickc.NewButton(m5stickc.ButtonAPin)
func NewButton(pin machine.Pin) Button {
	pin.Configure(machine.PinConfig{Mode: machine.PinInput})
	return Button{pin: pin}
}

// Pressed は押されていれば true を返す（押下で LOW のため反転）。
func (b Button) Pressed() bool {
	return !b.pin.Get()
}
