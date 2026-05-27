// Pong — レトロゲーム #4
//
// 縦画面の Pong。上が CPU、下が自分のパドル。
//   - 自分のパドル: チルト（本体を左右に傾ける）で移動
//   - ボタンB: 音 ON/OFF
// 先に 5 点取ったら勝ち。ゲーム後は A でリスタート。
//
// 書き込み: make flash PROJ=pong
package pong

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
	scrW     = 135
	scrH     = 240
	barH        = 22 // 上のスコア表示の高さ
	courtTop    = barH
	courtBottom = scrH - 22 // 下に音表示用の余白を確保
	pw          = 34        // パドル幅
	ph          = 6         // パドル高さ
	cpuY        = courtTop + 8
	playerY     = courtBottom - 14
	bs          = 6 // ボールの一辺
	winScore = 5
	frameMS  = 25
	cpuSpeed = 3
)

var (
	colCourt  = color.RGBA{15, 20, 35, 255}
	colPaddle = color.RGBA{225, 225, 225, 255}
	colBall   = color.RGBA{245, 240, 120, 255}
	colBar    = color.RGBA{0, 0, 160, 255}
	colText   = color.RGBA{255, 255, 255, 255}
)

var display *m5stickc.Display

// Run はランチャー（または cmd/pong）から呼ばれるエントリ。IMU(チルト)必須。
func Run(con *m5stickc.Console, imu *m5stickc.IMU) {
	display = con.Display
	btnA, btnB, buzzer := con.BtnA, con.BtnB, con.Buzzer

	if imu == nil {
		display.FillScreen(colCourt)
		tinyfont.WriteLine(display, &freemono.Bold9pt7b, 8, 80, "IMU FAIL", color.RGBA{230, 40, 40, 255})
		time.Sleep(2 * time.Second)
		return
	}

	seed, exit := title(btnA, btnB, buzzer)
	if exit {
		return // メニューへ
	}
	rng := rand.New(rand.NewSource(seed))
	for {
		if !play(imu, btnA, btnB, buzzer, rng) {
			return // メニューへ
		}
	}
}

func title(btnA, btnB m5stickc.Button, bz *m5stickc.Buzzer) (int64, bool) {
	display.FillScreen(colCourt)
	tinyfont.WriteLine(display, &freemono.Bold18pt7b, 18, 100, "PONG", colText)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 6, 138, "Tilt: move", colText)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 6, 160, "A: start", colText)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 6, 182, "hold A:menu", colText)
	drawSound(bz.Muted())
	return m5stickc.WaitStart(btnA, btnB, bz, func() { drawSound(bz.Muted()) })
}

// play は1ゲーム実行し、リトライ(true)/メニュー復帰(false)を返す。
func play(imu *m5stickc.IMU, btnA, btnB m5stickc.Button, bz *m5stickc.Buzzer, rng *rand.Rand) bool {
	display.FillScreen(colCourt)
	youScore, cpuScore := 0, 0
	drawScore(youScore, cpuScore)
	drawSound(bz.Muted())

	cpuX := int16((scrW - pw) / 2)
	playerX := int16((scrW - pw) / 2)
	bx := int16((scrW - bs) / 2)
	by := int16(scrH / 2)
	vx := int16(2)
	vy := int16(3)
	if rng.Intn(2) == 0 {
		vx = -vx
	}

	// 初期描画。
	display.FillRectangle(cpuX, cpuY, pw, ph, colPaddle)
	display.FillRectangle(playerX, playerY, pw, ph, colPaddle)
	display.FillRectangle(bx, by, bs, bs, colBall)
	prevCpuX, prevPlayerX, prevBx, prevBy := cpuX, playerX, bx, by

	soundBtn := m5stickc.NewEdgeButton(btnB)
	for {
		// --- 入力 ---
		if soundBtn.Tapped() {
			bz.ToggleMuted()
			drawSound(bz.Muted())
		}

		ax, _, _ := imu.Acceleration()
		playerX += clamp(int16(ax/60000), -7, 7) // チルト量に比例（上限あり）
		playerX = clamp(playerX, 0, scrW-pw)

		// CPU はボールを追う（速度上限で加減）。
		target := bx - (pw-bs)/2
		cpuX += clamp(target-cpuX, -cpuSpeed, cpuSpeed)
		cpuX = clamp(cpuX, 0, scrW-pw)

		// --- ボール更新 ---
		bx += vx
		if bx <= 0 {
			bx = 0
			vx = -vx
		}
		if bx+bs >= scrW {
			bx = scrW - bs
			vx = -vx
		}
		by += vy

		scored := 0 // +1: 自分の得点, -1: CPU得点
		if vy < 0 && by <= cpuY+ph {
			if bx+bs > cpuX && bx < cpuX+pw {
				by = cpuY + ph
				vy = -vy
				vx += english(bx, cpuX)
				bz.Tone(m5stickc.NoteC6, 12)
			} else if by <= courtTop {
				scored = 1 // CPU の後ろに抜けた
			}
		}
		if vy > 0 && by+bs >= playerY {
			if bx+bs > playerX && bx < playerX+pw {
				by = playerY - bs
				vy = -vy
				vx += english(bx, playerX)
				bz.Tone(m5stickc.NoteC6, 12)
			} else if by+bs >= courtBottom {
				scored = -1
			}
		}
		vx = clamp(vx, -4, 4)

		// --- 描画（動いたものだけ） ---
		display.FillRectangle(prevBx, prevBy, bs, bs, colCourt) // 旧ボール消去
		if cpuX != prevCpuX {
			display.FillRectangle(prevCpuX, cpuY, pw, ph, colCourt)
		}
		display.FillRectangle(cpuX, cpuY, pw, ph, colPaddle)
		if playerX != prevPlayerX {
			display.FillRectangle(prevPlayerX, playerY, pw, ph, colCourt)
		}
		display.FillRectangle(playerX, playerY, pw, ph, colPaddle)
		display.FillRectangle(bx, by, bs, bs, colBall)
		prevCpuX, prevPlayerX, prevBx, prevBy = cpuX, playerX, bx, by

		// --- 得点処理 ---
		if scored != 0 {
			if scored > 0 {
				youScore++
				bz.Tone(m5stickc.NoteE5, 120)
			} else {
				cpuScore++
				bz.Tone(m5stickc.NoteG4, 120)
			}
			drawScore(youScore, cpuScore)
			if youScore >= winScore || cpuScore >= winScore {
				return gameOver(btnA, btnB, bz, youScore >= winScore)
			}
			// サーブし直し。
			display.FillRectangle(prevBx, prevBy, bs, bs, colCourt)
			bx = (scrW - bs) / 2
			by = scrH / 2
			vy = 3
			if scored > 0 {
				vy = -3
			}
			vx = 2
			if rng.Intn(2) == 0 {
				vx = -2
			}
			prevBx, prevBy = bx, by
			time.Sleep(400 * time.Millisecond)
		}

		time.Sleep(frameMS * time.Millisecond)
	}
}

// english はパドルのどこに当たったかで横方向の変化を付ける。
func english(ballX, padX int16) int16 {
	center := padX + pw/2
	d := (ballX + bs/2) - center
	if d < -8 {
		return -1
	}
	if d > 8 {
		return 1
	}
	return 0
}

func clamp(v, lo, hi int16) int16 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// --- 描画ヘルパ ---

func drawScore(you, cpu int) {
	display.FillRectangle(0, 0, scrW, barH, colBar)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 4, 16, "YOU "+strconv.Itoa(you), colText)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 74, 16, "CPU "+strconv.Itoa(cpu), colText)
}

func drawSound(muted bool) {
	display.FillRectangle(0, scrH-20, scrW, 20, colCourt)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 8, scrH-5, m5stickc.SoundLabel(muted), colText)
}

func gameOver(btnA, btnB m5stickc.Button, bz *m5stickc.Buzzer, win bool) bool {
	if win {
		bz.Tone(m5stickc.NoteC5, 100)
		bz.Tone(m5stickc.NoteE5, 100)
		bz.Tone(m5stickc.NoteG5, 200)
	} else {
		m5stickc.GameOverJingle(bz)
	}

	display.FillRectangle(0, 95, scrW, 90, colCourt)
	msg := "YOU WIN"
	if !win {
		msg = "YOU LOSE"
	}
	tinyfont.WriteLine(display, &freemono.Bold12pt7b, 8, 126, msg, colText)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 22, 152, "A: retry", colText)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 22, 172, "hold A:menu", colText)

	return m5stickc.WaitRetryOrExit(btnA, btnB, bz, func() { drawSound(bz.Muted()) })
}
