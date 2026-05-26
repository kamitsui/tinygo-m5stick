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
type Buzzer struct {
	pin machine.Pin
}

// NewBuzzer は指定ピンを出力に設定して Buzzer を返す。
//
// 例: bz := m5stickc.NewBuzzer(m5stickc.BuzzerPin)
func NewBuzzer(pin machine.Pin) Buzzer {
	pin.Configure(machine.PinConfig{Mode: machine.PinOutput})
	pin.Low()
	return Buzzer{pin: pin}
}

// Tone は freqHz の矩形波を durMs ミリ秒だけ鳴らす。freqHz<=0 は無音の待機。
func (b Buzzer) Tone(freqHz, durMs int) {
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
func (b Buzzer) Off() {
	b.pin.Low()
}
