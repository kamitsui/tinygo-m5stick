// Launcher — 6つのレトロゲームをメニューで選んで起動する単一バイナリ。
//
// 操作:
//   - A タップ      : カーソルを下へ移動（循環）
//   - A 長押し(~0.4s) : 選択中のゲームを起動
//   - B            : サウンド ON/OFF（全ゲーム共通の設定として保持）
// 各ゲームのタイトル/ゲームオーバーで A 長押し＝このメニューへ戻る。
//
// 書き込み: make flash PROJ=launcher
package main

import (
	"image/color"
	"time"

	g2048 "github.com/kamitsui/tinygo-m5stick/games/2048"
	"github.com/kamitsui/tinygo-m5stick/games/dinosaur"
	"github.com/kamitsui/tinygo-m5stick/games/invaders"
	"github.com/kamitsui/tinygo-m5stick/games/pong"
	"github.com/kamitsui/tinygo-m5stick/games/snake"
	"github.com/kamitsui/tinygo-m5stick/games/tetris"
	"github.com/kamitsui/tinygo-m5stick/pkg/m5stickc"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freemono"
)

type entry struct {
	name string
	run  func(*m5stickc.Console, *m5stickc.IMU)
}

var games = []entry{
	{"Snake", snake.Run},
	{"2048", g2048.Run},
	{"Pong", pong.Run},
	{"Dinosaur", dinosaur.Run},
	{"Tetris", tetris.Run},
	{"Invaders", invaders.Run},
}

const selectHold = 18 // A長押しで決定（20ms*18≈0.36s）

var (
	colBG   = color.RGBA{12, 14, 24, 255}
	colBar  = color.RGBA{0, 0, 90, 255}
	colSel  = color.RGBA{0, 200, 120, 255}
	colText = color.RGBA{255, 255, 255, 255}
	colDark = color.RGBA{10, 12, 20, 255}
)

func rowY(i int) int16 { return int16(58 + i*24) }

func main() {
	con := m5stickc.NewConsole()
	imu, _ := m5stickc.NewIMU() // 使わないゲームもあるが共通で渡す（nil でも各ゲームが対応）
	d := con.Display

	soundBtn := m5stickc.NewEdgeButton(con.BtnB)
	sel := 0
	drawMenuFull(d, sel, con.Buzzer.Muted())

	aHeld := 0
	for {
		if soundBtn.Tapped() {
			con.Buzzer.ToggleMuted()
			drawSound(d, con.Buzzer.Muted())
		}
		if con.BtnA.Pressed() {
			aHeld++
			if aHeld == selectHold { // 長押し → 決定
				games[sel].run(con, imu)
				drawMenuFull(d, sel, con.Buzzer.Muted()) // ゲームから戻った
				for con.BtnA.Pressed() {                 // 離されるまで待つ
					time.Sleep(10 * time.Millisecond)
				}
				aHeld = 0
			}
		} else {
			if aHeld > 0 && aHeld < selectHold { // タップ → 移動
				old := sel
				sel = (sel + 1) % len(games)
				drawItem(d, old, false)
				drawItem(d, sel, true)
			}
			aHeld = 0
		}
		time.Sleep(20 * time.Millisecond)
	}
}

// drawMenuFull は全体を描く（初期・ゲーム復帰時）。移動時は drawItem のみで
// 差分更新するためちらつかない。
func drawMenuFull(d *m5stickc.Display, sel int, muted bool) {
	d.FillScreen(colBG)
	d.FillRectangle(0, 0, 135, 30, colBar)
	tinyfont.WriteLine(d, &freemono.Bold12pt7b, 18, 22, "GAMES", colText)
	for i := range games {
		drawItem(d, i, i == sel)
	}
	tinyfont.WriteLine(d, &freemono.Bold9pt7b, 6, 200, "A: move", colText)
	tinyfont.WriteLine(d, &freemono.Bold9pt7b, 6, 218, "hold A: go", colText)
	drawSound(d, muted)
}

// drawItem は1行だけ描く（選択行はハイライト）。
func drawItem(d *m5stickc.Display, i int, selected bool) {
	y := rowY(i)
	bg, fg := colBG, colText
	if selected {
		bg, fg = colSel, colDark
	}
	d.FillRectangle(2, y-16, 131, 22, bg)
	tinyfont.WriteLine(d, &freemono.Bold9pt7b, 12, y, games[i].name, fg)
}

func drawSound(d *m5stickc.Display, muted bool) {
	d.FillRectangle(0, 222, 135, 18, colBG)
	tinyfont.WriteLine(d, &freemono.Bold9pt7b, 6, 236, m5stickc.SoundLabel(muted), colText)
}
