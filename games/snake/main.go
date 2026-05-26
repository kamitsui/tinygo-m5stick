// Snake — レトロゲーム #1
//
// M5StickC Plus2 (135x240 縦) で動くスネークゲーム。
// 操作は2ボタン（相対回転）:
//   - ボタンA: 左に回転（反時計回り）
//   - ボタンB: 右に回転（時計回り）
//
// 書き込み: make flash PROJ=snake
package main

import (
	"image/color"
	"math/rand"
	"strconv"
	"time"

	"github.com/kamitsui/tinygo-m5stick/pkg/m5stickc"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freemono"
)

const (
	cell    = 9  // 1マスのピクセル数
	cols    = 15 // 135 / 9
	playTop = 22 // スコアバーの高さ
	rows    = 24 // (240-22) / 9
	tickMS  = 180
	pollMS  = 20
)

var (
	colBG    = color.RGBA{0, 0, 0, 255}
	colSnake = color.RGBA{0, 200, 0, 255}
	colHead  = color.RGBA{120, 255, 120, 255}
	colFood  = color.RGBA{230, 40, 40, 255}
	colText  = color.RGBA{255, 255, 255, 255}
	colBar   = color.RGBA{0, 0, 160, 255}
)

type pt struct{ x, y int16 }

var display *m5stickc.Display

func main() {
	m5stickc.HoldPower()
	display = m5stickc.NewDisplay()
	btnA := m5stickc.NewButton(m5stickc.ButtonAPin)
	btnB := m5stickc.NewButton(m5stickc.ButtonBPin)
	buzzer := m5stickc.NewBuzzer(m5stickc.BuzzerPin)

	rng := rand.New(rand.NewSource(titleAndSeed(btnA)))
	for {
		playGame(btnA, btnB, buzzer, rng)
	}
}

// titleAndSeed はタイトルを表示し、A が押されるまでの待ち時間を乱数シードに使う。
func titleAndSeed(btnA m5stickc.Button) int64 {
	display.FillScreen(colBG)
	tinyfont.WriteLine(display, &freemono.Bold12pt7b, 18, 100, "SNAKE", colSnake)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 16, 140, "Press A", colText)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 8, 185, "A: turn L", colText)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 8, 210, "B: turn R", colText)

	var n int64
	for !btnA.Pressed() {
		n++
		time.Sleep(5 * time.Millisecond)
	}
	for btnA.Pressed() { // 離されるまで待つ
		time.Sleep(10 * time.Millisecond)
	}
	return n + time.Now().UnixNano()
}

func playGame(btnA, btnB m5stickc.Button, bz *m5stickc.Buzzer, rng *rand.Rand) {
	display.FillScreen(colBG)
	drawBar(0)

	// 中央に縦3マスのヘビ、上向きで開始。
	body := make([]pt, 3, cols*rows)
	body[0] = pt{cols / 2, rows / 2}
	body[1] = pt{cols / 2, rows/2 + 1}
	body[2] = pt{cols / 2, rows/2 + 2}
	dir := pt{0, -1}
	for _, p := range body {
		drawCell(p, colSnake)
	}
	drawCell(body[0], colHead)

	food := placeFood(body, rng)
	drawCell(food, colFood)

	score := 0
	const movePolls = tickMS / pollMS
	poll := 0
	var prevA, prevB bool
	pendingTurn := 0 // -1: 左, +1: 右, 0: なし

	for {
		a, b := btnA.Pressed(), btnB.Pressed()
		if a && !prevA {
			pendingTurn = -1
		}
		if b && !prevB {
			pendingTurn = 1
		}
		prevA, prevB = a, b

		poll++
		if poll < movePolls {
			time.Sleep(pollMS * time.Millisecond)
			continue
		}
		poll = 0

		// 相対回転（画面は y 下向き）。
		switch pendingTurn {
		case -1: // 左（反時計回り）
			dir = pt{dir.y, -dir.x}
		case 1: // 右（時計回り）
			dir = pt{-dir.y, dir.x}
		}
		pendingTurn = 0

		head := pt{body[0].x + dir.x, body[0].y + dir.y}

		// 壁、または自分の体（末尾は移動で空くので除外）への衝突で終了。
		if head.x < 0 || head.x >= cols || head.y < 0 || head.y >= rows ||
			hits(body[:len(body)-1], head) {
			gameOver(btnA, bz, score)
			return
		}

		eating := head == food
		oldTail := body[len(body)-1]
		if eating {
			body = append(body, oldTail) // 1マス伸ばす（cap 確保済みで再確保なし）
		}
		copy(body[1:], body[:len(body)-1]) // 右シフト
		body[0] = head

		if eating {
			score++
			drawBar(score)
			bz.Tone(m5stickc.NoteC6, 40)
			food = placeFood(body, rng)
			drawCell(food, colFood)
		} else {
			drawCell(oldTail, colBG) // 末尾を消す
		}

		drawCell(body[0], colHead)
		drawCell(body[1], colSnake) // 直前の頭を体色に戻す
		time.Sleep(pollMS * time.Millisecond)
	}
}

func placeFood(body []pt, rng *rand.Rand) pt {
	for {
		f := pt{int16(rng.Intn(cols)), int16(rng.Intn(rows))}
		if !hits(body, f) {
			return f
		}
	}
}

func hits(body []pt, p pt) bool {
	for _, b := range body {
		if b == p {
			return true
		}
	}
	return false
}

func drawCell(p pt, c color.RGBA) {
	display.FillRectangle(p.x*cell, playTop+p.y*cell, cell-1, cell-1, c)
}

func drawBar(score int) {
	display.FillRectangle(0, 0, 135, playTop-1, colBar)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 6, 16, "SCORE "+strconv.Itoa(score), colText)
}

func gameOver(btnA m5stickc.Button, bz *m5stickc.Buzzer, score int) {
	bz.Tone(m5stickc.NoteG4, 120)
	bz.Tone(m5stickc.NoteE4, 120)
	bz.Tone(m5stickc.NoteC4, 240)

	display.FillRectangle(0, 88, 135, 80, colBG)
	tinyfont.WriteLine(display, &freemono.Bold12pt7b, 22, 116, "GAME", colText)
	tinyfont.WriteLine(display, &freemono.Bold12pt7b, 22, 140, "OVER", colText)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 18, 164, "A: retry", colText)

	for btnA.Pressed() { // まず離す
		time.Sleep(20 * time.Millisecond)
	}
	for !btnA.Pressed() { // 押されるまで待つ
		time.Sleep(20 * time.Millisecond)
	}
}
