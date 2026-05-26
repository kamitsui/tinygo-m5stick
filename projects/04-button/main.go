// Project 04: Button
//
// M5StickC Plus2 のボタン A(GPIO37) / B(GPIO39) の入力を検証する。
// 押下状態を画面にリアルタイム表示し、A は赤色LED(GPIO19)にも連動させる。
//
// 注意: GPIO34-39 は入力専用で内部プルアップが無いが、基板側に外部プルアップが
// あるため PinInput で読める。ボタンは押下で LOW(false)、離すと HIGH(true)。
//
// 書き込み:
//
//	tinygo flash -target=esp32-coreboard-v2 \
//	  -port=/dev/cu.wchusbserial58FA0591591 \
//	  ./projects/04-button
package main

import (
	"image/color"
	"machine"
	"time"

	"tinygo.org/x/drivers"
	"tinygo.org/x/drivers/st7789"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freemono"
)

var (
	black = color.RGBA{R: 0, G: 0, B: 0, A: 255}
	white = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	green = color.RGBA{R: 0, G: 200, B: 0, A: 255}
	gray  = color.RGBA{R: 60, G: 60, B: 60, A: 255}
	blue  = color.RGBA{R: 0, G: 0, B: 200, A: 255}
)

func main() {
	hold := machine.GPIO4
	hold.Configure(machine.PinConfig{Mode: machine.PinOutput})
	hold.High()

	led := machine.GPIO19
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	btnA := machine.GPIO37
	btnB := machine.GPIO39
	btnA.Configure(machine.PinConfig{Mode: machine.PinInput})
	btnB.Configure(machine.PinConfig{Mode: machine.PinInput})

	machine.SPI2.Configure(machine.SPIConfig{
		Frequency: 40e6,
		SCK:       machine.GPIO13,
		SDO:       machine.GPIO15,
		SDI:       machine.NoPin,
	})
	display := st7789.New(machine.SPI2, machine.GPIO12, machine.GPIO14, machine.GPIO5, machine.GPIO27)
	display.Configure(st7789.Config{
		Width:        135,
		Height:       240,
		Rotation:     drivers.Rotation180,
		RowOffset:    40,
		ColumnOffset: 53,
	})

	display.FillScreen(black)
	display.FillRectangle(0, 0, 135, 28, blue)
	tinyfont.WriteLine(&display, &freemono.Bold9pt7b, 8, 20, "Buttons", white)

	// 前回状態を無効値で初期化し、初回に必ず描画させる。
	lastA, lastB := -1, -1
	for {
		// 押下=LOW なので反転して「押されている=1」にする。
		a := boolToInt(!btnA.Get())
		b := boolToInt(!btnB.Get())

		if a != lastA {
			drawState(&display, 60, "BtnA", a == 1)
			if a == 1 {
				led.High()
			} else {
				led.Low()
			}
			lastA = a
		}
		if b != lastB {
			drawState(&display, 110, "BtnB", b == 1)
			lastB = b
		}
		time.Sleep(20 * time.Millisecond)
	}
}

func drawState(d *st7789.Device, y int16, label string, pressed bool) {
	c := gray
	text := "----"
	if pressed {
		c = green
		text = "PUSH"
	}
	d.FillRectangle(0, y, 135, 26, black)
	tinyfont.WriteLine(d, &freemono.Bold9pt7b, 8, y+18, label, white)
	tinyfont.WriteLine(d, &freemono.Bold9pt7b, 80, y+18, text, c)
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
