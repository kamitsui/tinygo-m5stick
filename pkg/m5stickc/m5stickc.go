// Package m5stickc は M5StickC Plus2 (ESP32-PICO-V3-02 / Xtensa) のハードウェアを
// TinyGo から扱うためのピン定義とヘルパをまとめる。
//
// 各値・設定は実機で検証済み（projects/01-blink〜04-button で確認）。
// ビルド/書き込みターゲットは esp32-coreboard-v2（M5StickC Plus2 専用ターゲットは
// TinyGo に無いため汎用ESP32ターゲットで代用）。
package m5stickc

import "machine"

// ピン定義（M5StickC Plus2）。
const (
	LED  = machine.GPIO19 // 赤色LED（IRエミッタと共用）
	hold = machine.GPIO4  // 電源保持（バッテリ駆動時に HIGH を保持）

	// ボタン（押下=LOW / active-low。GPIO34-39は入力専用・基板側に外部プルアップ有り）
	ButtonAPin = machine.GPIO37 // 正面の大きな M5 ボタン
	ButtonBPin = machine.GPIO39 // 側面ボタン
	ButtonCPin = machine.GPIO35 // 電源ボタン

	BuzzerPin = machine.GPIO2 // パッシブブザー

	// 表示器 ST7789V2 (SPI2)
	tftSCK = machine.GPIO13
	tftSDO = machine.GPIO15
	tftRST = machine.GPIO12
	tftDC  = machine.GPIO14
	tftCS  = machine.GPIO5
	tftBL  = machine.GPIO27

	// 内蔵 IMU (MPU6886) などが繋がる内部 I2C。
	imuSDA = machine.GPIO21
	imuSCL = machine.GPIO22
)

// HoldPower は電源保持ピンを HIGH にし、バッテリ駆動時に電源が落ちないようにする。
// USB 給電時は不要だが、呼んでも害は無い。各 main の冒頭で呼ぶことを推奨。
func HoldPower() {
	hold.Configure(machine.PinConfig{Mode: machine.PinOutput})
	hold.High()
}
