package m5stickc

import (
	"machine"
	"time"
)

// Buzzer はパッシブブザーを表す。
//
// TinyGo の Xtensa ESP32 は PWM(LEDC) 未対応のため、tinygo.org/x/drivers/tone
// （PWM前提）は使えない。代わりに GPIO をソフトトグルして矩形波を生成する。
// time.Sleep のµsオーダー精度に依存するため音程は厳密ではないが実用上問題ない。
//
// Muted を立てると Tone は即座に無音で返る（ゲームの音ON/OFF切替に使う）。
type Buzzer struct {
	pin   machine.Pin
	muted bool
}

// NewBuzzer は指定ピンを出力に設定して *Buzzer を返す。
//
// 例: bz := m5stickc.NewBuzzer(m5stickc.BuzzerPin)
func NewBuzzer(pin machine.Pin) *Buzzer {
	pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pin.Low()
	return &Buzzer{pin: pin}
}

// SetMuted は消音の有効/無効を設定する。
func (b *Buzzer) SetMuted(m bool) { b.muted = m }

// ToggleMuted は消音状態を反転し、反転後の状態を返す。
func (b *Buzzer) ToggleMuted() bool {
	b.muted = !b.muted
	return b.muted
}

// Muted は現在消音中かどうかを返す。
func (b *Buzzer) Muted() bool { return b.muted }

// Tone は freqHz の矩形波を durMs ミリ秒だけ鳴らす。消音中は即座に無音で返る。
// freqHz<=0 は休符（durMs だけ待つ）。
func (b *Buzzer) Tone(freqHz, durMs int) {
	if b.muted {
		return
	}
	if freqHz <= 0 {
		time.Sleep(time.Duration(durMs) * time.Millisecond)
		return
	}
	halfPeriod := time.Duration(1_000_000/(2*freqHz)) * time.Microsecond
	cycles := freqHz * durMs / 1000
	for i := 0; i < cycles; i++ {
		b.pin.High()
		time.Sleep(halfPeriod)
		b.pin.Low()
		time.Sleep(halfPeriod)
	}
	b.pin.Low()
}

// Off はブザーを確実に停止（LOW）する。
func (b *Buzzer) Off() {
	b.pin.Low()
}
