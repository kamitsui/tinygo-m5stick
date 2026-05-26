// Project 02: Display (ST7789V2)
//
// M5StickC Plus2 の 1.14" 135x240 TFT (ST7789V2) に色塗りと "Hello World" を表示し、
// SPI 経路・パネル初期化・表示オフセットが正しいかを検証する。
//
// 配線（M5StickC Plus2 固定）:
//
//	SCLK=G13  MOSI=G15  DC=G14  RST=G12  CS=G5  BL(backlight)=G27
//
// 書き込み:
//
//	tinygo flash -target=esp32-coreboard-v2 \
//	  -port=/dev/cu.wchusbserial58FA0591591 \
//	  ./projects/02-display
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
	red   = color.RGBA{R: 255, G: 0, B: 0, A: 255}
	green = color.RGBA{R: 0, G: 255, B: 0, A: 255}
	blue  = color.RGBA{R: 0, G: 0, B: 255, A: 255}
)

func main() {
	// バッテリ駆動時の電源保持。USB給電でも害は無い。
	hold := machine.GPIO4
	hold.Configure(machine.PinConfig{Mode: machine.PinOutput})
	hold.High()

	machine.SPI2.Configure(machine.SPIConfig{
		Frequency: 40e6,
		SCK:       machine.GPIO13,
		SDO:       machine.GPIO15,
		SDI:       machine.NoPin,
	})

	display := st7789.New(machine.SPI2,
		machine.GPIO12, // RST
		machine.GPIO14, // DC
		machine.GPIO5,  // CS
		machine.GPIO27, // BL (backlight)
	)
	// 重要: このドライバ(v0.35.0)は Rotation0 のとき RowOffset/ColumnOffset を
	// 強制的に 0 にする。設定値が効くのは Rotation180/90/270 のみ。
	// 1.14"パネルは GRAM 原点からずれて実装されているため Rotation180 を採用。
	// ColumnOffset は 52 だと左端に 1px のゴミが残るため 53 が正値。
	// → デバイスは USB-C を上にして持つと正立する。
	display.Configure(st7789.Config{
		Width:        135,
		Height:       240,
		Rotation:     drivers.Rotation180,
		RowOffset:    40,
		ColumnOffset: 53,
	})

	// 1) 原色で全画面塗り → SPI/パネルが生きているかを目視確認。
	for _, c := range []color.RGBA{red, green, blue} {
		display.FillScreen(c)
		time.Sleep(500 * time.Millisecond)
	}

	// 2) 黒地にタイトル帯と "Hello World"。
	display.FillScreen(black)
	display.FillRectangle(0, 0, 135, 28, blue)
	tinyfont.WriteLine(&display, &freemono.Bold9pt7b, 8, 20, "M5Stick", white)
	tinyfont.WriteLine(&display, &freemono.Bold9pt7b, 8, 120, "Hello", green)
	tinyfont.WriteLine(&display, &freemono.Bold9pt7b, 8, 150, "World", green)

	for {
		time.Sleep(time.Second)
	}
}
