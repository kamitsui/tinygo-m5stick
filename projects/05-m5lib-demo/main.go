// Project 05: pkg/m5stickc demo
//
// 共通パッケージ pkg/m5stickc を使ったデモ。低レベルな machine 呼び出しを
// 直接書かずに、表示・ボタン・ブザー・LED を扱えることを示す。
//
// 操作:
//   - ボタンA: カウントアップ + 上昇音 + LED点灯
//   - ボタンB: カウントリセット + 低い音
//
// 書き込み: make flash PROJ=05-m5lib-demo
package main

import (
	"image/color"
	"machine"
	"strconv"
	"time"

	"github.com/kamitsui/tinygo-m5stick/pkg/m5stickc"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freemono"
)

var (
	black = color.RGBA{R: 0, G: 0, B: 0, A: 255}
	white = color.RGBA{R: 255, G: 255, B: 255, A: 255}
	green = color.RGBA{R: 0, G: 200, B: 0, A: 255}
	blue  = color.RGBA{R: 0, G: 0, B: 200, A: 255}
)

func main() {
	m5stickc.HoldPower()

	led := m5stickc.LED
	led.Configure(machine.PinConfig{Mode: machine.PinOutput})

	btnA := m5stickc.NewButton(m5stickc.ButtonAPin)
	btnB := m5stickc.NewButton(m5stickc.ButtonBPin)
	buzzer := m5stickc.NewBuzzer(m5stickc.BuzzerPin)

	display := m5stickc.NewDisplay()
	display.FillScreen(black)
	display.FillRectangle(0, 0, m5stickc.DisplayWidth, 28, blue)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 8, 20, "m5stickc", white)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 8, 70, "A:count", white)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 8, 95, "B:reset", white)

	count := 0
	drawCount(display, count)

	var prevA, prevB bool
	for {
		a, b := btnA.Pressed(), btnB.Pressed()

		if a && !prevA { // 押した瞬間（立ち上がり）
			count++
			drawCount(display, count)
			led.High()
			buzzer.Tone(m5stickc.NoteC5, 60)
			buzzer.Tone(m5stickc.NoteE5, 60)
		}
		if !a && prevA {
			led.Low()
		}
		if b && !prevB {
			count = 0
			drawCount(display, count)
			buzzer.Tone(m5stickc.NoteG4, 120)
		}

		prevA, prevB = a, b
		time.Sleep(15 * time.Millisecond)
	}
}

func drawCount(d *m5stickc.Display, count int) {
	d.FillRectangle(0, 150, m5stickc.DisplayWidth, 40, black)
	tinyfont.WriteLine(d, &freemono.Bold12pt7b, 8, 180, "= "+strconv.Itoa(count), green)
}
