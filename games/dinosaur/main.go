// Dinosaur — レトロゲーム #3
//
// Chrome の恐竜ゲーム風エンドレスランナー（縦画面）。
//   - ボタンA: ジャンプ
//   - ボタンB: 音 ON/OFF
// サボテンを飛び越えて進み、ぶつかるとゲームオーバー（A でリトライ）。
//
// 書き込み: make flash PROJ=dinosaur
package dinosaur

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
	scrW    = 135
	bandTop = 110 // 走路バンドの画面上端
	bandH   = 100 // バンドの高さ（この範囲をフレームバッファでスクロール）

	groundY = 92 // バンド内ローカルの地面ライン（足元）
	dinoX   = 14
	dinoW   = 16
	dinoH   = 20
	dinoTop0 = groundY - dinoH // 接地時のローカル上端

	gravity = 2
	jumpV   = 14

	frameMS = 30
)

var (
	colSky    = color.RGBA{170, 210, 230, 255}
	colGround = color.RGBA{110, 95, 70, 255}
	colDino   = color.RGBA{70, 80, 70, 255}
	colCactus = color.RGBA{40, 130, 50, 255}
	colBar    = color.RGBA{0, 0, 160, 255}
	colBG     = color.RGBA{28, 26, 24, 255}
	colText   = color.RGBA{255, 255, 255, 255}
)

type obstacle struct {
	x, w, h int16
}

var (
	display *m5stickc.Display
	band    *m5stickc.Canvas
	bar     *m5stickc.Canvas
)

// Run はランチャー（または cmd/dinosaur）から呼ばれるエントリ。IMU は使わない。
func Run(con *m5stickc.Console, imu *m5stickc.IMU) {
	display = con.Display
	band = m5stickc.NewCanvas(scrW, bandH)
	bar = m5stickc.NewCanvas(scrW, 26)
	btnA, btnB, buzzer := con.BtnA, con.BtnB, con.Buzzer

	seed, exit := title(btnA, btnB, buzzer)
	if exit {
		return // メニューへ
	}
	rng := rand.New(rand.NewSource(seed))
	for {
		if !play(btnA, btnB, buzzer, rng) {
			return // メニューへ
		}
	}
}

func title(btnA, btnB m5stickc.Button, bz *m5stickc.Buzzer) (int64, bool) {
	display.FillScreen(colSky)
	display.FillRectangle(0, 0, scrW, 26, colBar)
	tinyfont.WriteLine(display, &freemono.Bold12pt7b, 6, 96, "DINO", color.RGBA{60, 60, 60, 255})
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 6, 134, "A: jump", color.RGBA{40, 40, 40, 255})
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 6, 156, "hold A:menu", color.RGBA{40, 40, 40, 255})
	drawSound(bz.Muted())
	return m5stickc.WaitStart(btnA, btnB, bz, func() { drawSound(bz.Muted()) })
}

// play は1ゲーム実行し、リトライ(true)/メニュー復帰(false)を返す。
func play(btnA, btnB m5stickc.Button, bz *m5stickc.Buzzer, rng *rand.Rand) bool {
	// 静的な背景（バンドの外）を一度だけ描く。
	display.FillScreen(colSky)
	display.FillRectangle(0, 0, scrW, 26, colBar)
	display.FillRectangle(0, bandTop+bandH, scrW, 240-(bandTop+bandH), colGround)
	drawSound(bz.Muted())

	dinoTop := int16(dinoTop0)
	vy := int16(0)
	grounded := true
	obstacles := make([]obstacle, 0, 4)
	speed := int16(3)
	spawnTimer := int16(30)
	score := 0
	var prevA bool
	soundBtn := m5stickc.NewEdgeButton(btnB)

	for {
		a := btnA.Pressed()
		if soundBtn.Tapped() {
			bz.ToggleMuted()
			drawSound(bz.Muted())
		}
		if a && !prevA && grounded { // ジャンプ
			vy = -jumpV
			grounded = false
			bz.Tone(m5stickc.NoteG5, 30)
		}
		prevA = a

		// 物理（ジャンプ）。
		if !grounded {
			dinoTop += vy
			vy += gravity
			if dinoTop >= dinoTop0 {
				dinoTop = dinoTop0
				vy = 0
				grounded = true
			}
		}

		// 障害物の移動・生成・削除。
		for i := range obstacles {
			obstacles[i].x -= speed
		}
		if len(obstacles) > 0 && obstacles[0].x+obstacles[0].w < 0 {
			obstacles = obstacles[1:]
		}
		spawnTimer -= 1
		if spawnTimer <= 0 {
			h := int16(18 + rng.Intn(3)*6) // 18 / 24 / 30
			obstacles = append(obstacles, obstacle{x: scrW, w: 12, h: h})
			spawnTimer = int16(28 + rng.Intn(40)) // ランダム間隔
		}

		// 当たり判定。
		if hit(dinoTop, obstacles) {
			return gameOver(btnA, btnB, bz, score)
		}

		// スコアと難易度。
		score++
		if score%400 == 0 && speed < 8 {
			speed++
		}

		render(dinoTop, obstacles)
		drawBar(score)
		time.Sleep(frameMS * time.Millisecond)
	}
}

func hit(dinoTop int16, obstacles []obstacle) bool {
	for _, o := range obstacles {
		oy := int16(groundY) - o.h
		if dinoX < o.x+o.w && dinoX+dinoW > o.x &&
			dinoTop < groundY && dinoTop+dinoH > oy {
			return true
		}
	}
	return false
}

// --- 描画 ---

func render(dinoTop int16, obstacles []obstacle) {
	band.Fill(colSky)
	band.FillRectangle(0, groundY, scrW, bandH-groundY, colGround) // 地面
	// 恐竜（体＋頭＋目）。
	band.FillRectangle(dinoX, dinoTop+6, dinoW-4, dinoH-6, colDino)
	band.FillRectangle(dinoX+dinoW-8, dinoTop, 8, 9, colDino)
	band.FillRectangle(dinoX+dinoW-3, dinoTop+2, 2, 2, colText) // 目
	// サボテン。
	for _, o := range obstacles {
		band.FillRectangle(o.x, groundY-o.h, o.w, o.h, colCactus)
	}
	display.DrawBuffer(0, bandTop, scrW, bandH, band.Buffer())
}

func drawBar(score int) {
	bar.Fill(colBar)
	tinyfont.WriteLine(bar, &freemono.Bold9pt7b, 6, 18, "SCORE "+strconv.Itoa(score), colText)
	display.DrawBuffer(0, 0, scrW, 26, bar.Buffer())
}

func drawSound(muted bool) {
	display.FillRectangle(0, 216, scrW, 24, colGround)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 8, 233, m5stickc.SoundLabel(muted), colText)
}

func gameOver(btnA, btnB m5stickc.Button, bz *m5stickc.Buzzer, score int) bool {
	m5stickc.GameOverJingle(bz)

	display.FillRectangle(0, 95, scrW, 95, colBG)
	tinyfont.WriteLine(display, &freemono.Bold12pt7b, 24, 120, "GAME", colText)
	tinyfont.WriteLine(display, &freemono.Bold12pt7b, 24, 144, "OVER", colText)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 22, 166, "A: retry", colText)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 22, 184, "hold A:menu", colText)

	return m5stickc.WaitRetryOrExit(btnA, btnB, bz, func() { drawSound(bz.Muted()) })
}
