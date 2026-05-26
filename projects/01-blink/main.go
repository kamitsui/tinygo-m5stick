// Project 01: Blink
//
// 動作確認用の最小プログラム。M5StickC Plus2 の赤色LED(GPIO19)を点滅させる。
// TinyGo のツールチェーン・書き込み経路がひと通り通ることを確認するのが目的。
//
// 書き込み:
//
//	tinygo flash -target=esp32-coreboard-v2 \
//	  -port=/dev/cu.wchusbserial58FA0591591 \
//	  ./projects/01-blink
//
// 注意: M5StickC Plus2 に専用ターゲットは無いため、汎用の esp32-coreboard-v2
// を使い、ピンはこのコード内で直接指定する。
package main

import (
	"machine"
	"time"
)

const (
	// 赤色LED（IRエミッタと共用）。
	ledPin = machine.GPIO19
	// 電源保持ピン。バッテリ駆動時に HIGH を保つと電源が落ちない。
	// USB給電時は無くても動くが、念のため立てておく。
	holdPin = machine.GPIO4
)

func main() {
	hold := holdPin
	hold.Configure(machine.PinConfig{Mode: machine.PinOutput})
	hold.High()

	led := ledPin
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	for {
		led.High()
		time.Sleep(500 * time.Millisecond)
		led.Low()
		time.Sleep(500 * time.Millisecond)
	}
}
