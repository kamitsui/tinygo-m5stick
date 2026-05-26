// Project 03: Buzzer
//
// M5StickC Plus2 のパッシブブザー(GPIO2)で音を鳴らす検証。
//
// 注意: TinyGo の Xtensa ESP32 は PWM (LEDC) 未対応のため、tinygo.org/x/drivers/tone
// （PWM前提）は使えない。代わりに GPIO をソフトでトグルして矩形波を生成する。
// time.Sleep のµsオーダー精度に依存するため音程は厳密ではないが、鳴動確認には十分。
//
// 書き込み:
//
//	tinygo flash -target=esp32-coreboard-v2 \
//	  -port=/dev/cu.wchusbserial58FA0591591 \
//	  ./projects/03-buzzer
package main

import (
	"machine"
	"time"
)

const buzzerPin = machine.GPIO2

// 音名 → 周波数(Hz)
const (
	noteC5 = 523
	noteE5 = 659
	noteG5 = 784
	noteC6 = 1047
)

// tone は freqHz の矩形波を durMs ミリ秒だけ buzzerPin に出力する。
func tone(pin machine.Pin, freqHz, durMs int) {
	if freqHz <= 0 {
		time.Sleep(time.Duration(durMs) * time.Millisecond)
		return
	}
	halfPeriod := time.Duration(1_000_000/(2*freqHz)) * time.Microsecond
	cycles := freqHz * durMs / 1000
	for i := 0; i < cycles; i++ {
		pin.High()
		time.Sleep(halfPeriod)
		pin.Low()
		time.Sleep(halfPeriod)
	}
}

func main() {
	hold := machine.GPIO4
	hold.Configure(machine.PinConfig{Mode: machine.PinOutput})
	hold.High()

	buzzer := buzzerPin
	buzzer.Configure(machine.PinConfig{Mode: machine.PinOutput})
	buzzer.Low()

	led := machine.GPIO19
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	melody := []int{noteC5, noteE5, noteG5, noteC6}

	for {
		for _, n := range melody {
			led.High()
			tone(buzzer, n, 250)
			led.Low()
			time.Sleep(80 * time.Millisecond)
		}
		buzzer.Low()
		time.Sleep(1500 * time.Millisecond)
	}
}
