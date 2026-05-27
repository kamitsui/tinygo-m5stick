// Invaders — レトロゲーム #6
//
// 縦画面のスペースインベーダー風。
//   - チルト左右: 自機を移動
//   - ボタンA:    弾を発射（タップで1発 / 長押しで連射モード ON/OFF）
//   - ボタンB:    音 ON/OFF（タイトル・ゲーム中・GAME OVER のどこでも切替可）
// 編隊を全滅させると次ウェーブ（加速）。編隊が自機の高さまで降下、または
// 爆弾が自機に当たるとゲームオーバー（A でリトライ）。音設定は画面遷移で保持。
//
// 描画は移動した物体だけを部分再描画してちらつきを抑える。
//
// 書き込み: make flash PROJ=invaders
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
	scrW    = 135
	barH    = 54 // スコア/サウンド/連射の3行
	icols   = 5
	irows   = 4
	invW    = 14
	invH    = 10
	gapX    = 4
	gapY    = 8
	formW   = icols*invW + (icols-1)*gapX // 86
	fxMin   = 4
	fxMax   = scrW - formW - 4 // 45
	stepX   = 4
	dropY   = 8
	shipY   = 224
	shipW   = 16
	shipH   = 8
	bulletW = 2
	bulletH = 6
	maxShot = 4
	bombW   = 3
	bombH   = 6
	maxBomb = 3
	fireCDr = 12 // 連射間隔（フレーム）≈360ms ≒ 約2.8発/秒
	holdThr = 16 // A長押し判定（フレーム）
	tiltTh  = 250000
	frameMS = 30
)

var (
	colBG     = color.RGBA{10, 12, 22, 255}
	colBar    = color.RGBA{0, 0, 80, 255}
	colText   = color.RGBA{255, 255, 255, 255}
	colInv    = color.RGBA{0, 220, 90, 255}
	colInvLeg = color.RGBA{0, 150, 60, 255}
	colShip   = color.RGBA{120, 220, 255, 255}
	colBullet = color.RGBA{245, 240, 120, 255}
	colBomb   = color.RGBA{255, 90, 70, 255}
)

var (
	display *m5stickc.Display
	aliveG  [irows][icols]bool
	fx, fy  int16
	dir     int16
)

type vec struct{ x, y int16 }

func main() {
	con := m5stickc.NewConsole()
	display = con.Display
	btnA, btnB, buzzer := con.BtnA, con.BtnB, con.Buzzer

	imu, err := m5stickc.NewIMU()
	if err != nil {
		display.FillScreen(colBG)
		tinyfont.WriteLine(display, &freemono.Bold9pt7b, 8, 80, "IMU FAIL", color.RGBA{230, 40, 40, 255})
		for {
			time.Sleep(time.Second)
		}
	}

	rng := rand.New(rand.NewSource(title(btnA, btnB, buzzer)))
	for {
		play(imu, btnA, btnB, buzzer, rng)
	}
}

func title(btnA, btnB m5stickc.Button, bz *m5stickc.Buzzer) int64 {
	display.FillScreen(colBG)
	tinyfont.WriteLine(display, &freemono.Bold18pt7b, 4, 86, "INVADE", colInv)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 6, 122, "Tilt: move", colText)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 6, 144, "A: fire", colText)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 6, 166, "A-hold: rapid", colText)
	drawTitleSound(bz.Muted())
	return m5stickc.WaitStart(btnA, btnB, bz, func() { drawTitleSound(bz.Muted()) })
}

func invX(c int) int16 { return fx + int16(c)*(invW+gapX) }
func invY(r int) int16 { return fy + int16(r)*(invH+gapY) }

func play(imu *m5stickc.IMU, btnA, btnB m5stickc.Button, bz *m5stickc.Buzzer, rng *rand.Rand) {
	score := 0
	wave := 0
	display.FillScreen(colBG)
	drawScore(score)
	drawSound(bz.Muted())
	rapid := false
	drawRapid(rapid)

	resetWave(wave)
	drawFormation(true)
	shipX := int16((scrW - shipW) / 2)
	drawShip(shipX, true)

	bullets := make([]vec, 0, maxShot)
	bombs := make([]vec, 0, maxBomb)

	fire := func() {
		if len(bullets) >= maxShot {
			return
		}
		bullets = append(bullets, vec{shipX + shipW/2 - 1, shipY - bulletH})
		bz.Tone(m5stickc.NoteG5, 10)
	}

	stepCD := stepFrames(wave)
	fireCD := 0
	aHeld := 0
	toggledHold := false
	var prevA bool
	soundBtn := m5stickc.NewEdgeButton(btnB)

	for {
		a := btnA.Pressed()
		if soundBtn.Tapped() {
			bz.ToggleMuted()
			drawSound(bz.Muted())
		}

		// A 長押しで連射モード切替。
		if a {
			aHeld++
			if aHeld >= holdThr && !toggledHold {
				rapid = !rapid
				toggledHold = true
				drawRapid(rapid)
				bz.Tone(m5stickc.NoteA5, 40)
			}
		} else {
			aHeld = 0
			toggledHold = false
		}

		// 発射: 連射モードは自動、通常はタップで1発。
		if rapid {
			fireCD--
			if fireCD <= 0 {
				fire()
				fireCD = fireCDr
			}
		} else if a && !prevA {
			fire()
		}
		prevA = a

		// 自機移動（チルト）。
		ax, _, _ := imu.Acceleration()
		nx := clamp(shipX+clamp(int16(ax/60000), -6, 6), 2, scrW-shipW-2)
		if nx != shipX {
			drawShip(shipX, false)
			shipX = nx
			drawShip(shipX, true)
		}

		// 自弾の移動。
		keptB := bullets[:0]
		for _, bl := range bullets {
			display.FillRectangle(bl.x, bl.y, bulletW, bulletH, colBG)
			bl.y -= 8
			if bl.y < barH {
				continue
			}
			if r, c, ok := hitInvader(bl.x, bl.y); ok {
				aliveG[r][c] = false
				drawInvader(invX(c), invY(r), false)
				score += 10
				drawScore(score)
				bz.Tone(m5stickc.NoteC6, 16)
				continue
			}
			display.FillRectangle(bl.x, bl.y, bulletW, bulletH, colBullet)
			keptB = append(keptB, bl)
		}
		bullets = keptB

		// 爆弾の移動。
		for _, bm := range bombs {
			display.FillRectangle(bm.x, bm.y, bombW, bombH, colBG)
		}
		keptM := bombs[:0]
		for _, bm := range bombs {
			bm.y += 3
			if bm.y >= shipY && bm.x+bombW > shipX && bm.x < shipX+shipW {
				gameOver(btnA, btnB, bz, score)
				return
			}
			if bm.y < shipY+shipH {
				display.FillRectangle(bm.x, bm.y, bombW, bombH, colBomb)
				keptM = append(keptM, bm)
			}
		}
		bombs = keptM
		if len(bombs) < maxBomb && rng.Intn(100) < 4 {
			c := rng.Intn(icols)
			for r := irows - 1; r >= 0; r-- {
				if aliveG[r][c] {
					bombs = append(bombs, vec{invX(c) + invW/2, invY(r) + invH})
					break
				}
			}
		}

		// 編隊のステップ移動。
		stepCD--
		if stepCD <= 0 {
			stepCD = stepFrames(wave)
			drawFormation(false)
			nf := fx + dir*stepX
			if nf < fxMin || nf > fxMax {
				dir = -dir
				fy += dropY
			} else {
				fx = nf
			}
			drawFormation(true)
			if lowestY()+invH >= shipY {
				gameOver(btnA, btnB, bz, score)
				return
			}
		}

		// 全滅 → 次ウェーブ。
		if remaining() == 0 {
			wave++
			bz.Tone(m5stickc.NoteE5, 60)
			bz.Tone(m5stickc.NoteG5, 60)
			bz.Tone(m5stickc.NoteC6, 120)
			display.FillRectangle(0, barH, scrW, shipY-barH, colBG)
			resetWave(wave)
			drawFormation(true)
			stepCD = stepFrames(wave)
			bullets = bullets[:0]
			bombs = bombs[:0]
		}

		time.Sleep(frameMS * time.Millisecond)
	}
}

// --- 状態ヘルパ ---

func resetWave(wave int) {
	for r := 0; r < irows; r++ {
		for c := 0; c < icols; c++ {
			aliveG[r][c] = true
		}
	}
	fx = fxMin
	fy = int16(barH + 8 + min(wave, 4)*dropY)
	dir = 1
}

func stepFrames(wave int) int {
	f := 14 - remaining()/3 - wave
	if f < 3 {
		f = 3
	}
	return f
}

func remaining() int {
	n := 0
	for r := 0; r < irows; r++ {
		for c := 0; c < icols; c++ {
			if aliveG[r][c] {
				n++
			}
		}
	}
	return n
}

func lowestY() int16 {
	var y int16 = -1
	for r := 0; r < irows; r++ {
		for c := 0; c < icols; c++ {
			if aliveG[r][c] && invY(r) > y {
				y = invY(r)
			}
		}
	}
	return y
}

func hitInvader(bx, by int16) (int, int, bool) {
	for r := 0; r < irows; r++ {
		for c := 0; c < icols; c++ {
			if !aliveG[r][c] {
				continue
			}
			x, y := invX(c), invY(r)
			if bx+bulletW > x && bx < x+invW && by < y+invH && by+bulletH > y {
				return r, c, true
			}
		}
	}
	return 0, 0, false
}

// --- 描画 ---

func drawFormation(on bool) {
	for r := 0; r < irows; r++ {
		for c := 0; c < icols; c++ {
			if aliveG[r][c] {
				drawInvader(invX(c), invY(r), on)
			}
		}
	}
}

func drawInvader(x, y int16, on bool) {
	if !on {
		display.FillRectangle(x, y, invW, invH, colBG)
		return
	}
	display.FillRectangle(x+1, y+1, invW-2, invH-4, colInv)
	display.FillRectangle(x, y+invH-3, 3, 3, colInvLeg)
	display.FillRectangle(x+invW-3, y+invH-3, 3, 3, colInvLeg)
	display.FillRectangle(x+3, y+3, 2, 2, colBG)
	display.FillRectangle(x+invW-5, y+3, 2, 2, colBG)
}

func drawShip(x int16, on bool) {
	if !on {
		display.FillRectangle(x, shipY, shipW, shipH, colBG)
		return
	}
	display.FillRectangle(x+2, shipY+3, shipW-4, shipH-3, colShip)
	display.FillRectangle(x+shipW/2-1, shipY, 2, 4, colShip)
}

func drawScore(score int) {
	display.FillRectangle(0, 0, scrW, 20, colBar)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 4, 15, "SCORE "+strconv.Itoa(score), colText)
}

func drawSound(muted bool) {
	display.FillRectangle(0, 20, scrW, 18, colBar)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 4, 34, m5stickc.SoundLabel(muted), colText)
}

func drawRapid(on bool) {
	display.FillRectangle(0, 38, scrW, 16, colBar)
	s := "A-hold: rapid"
	if on {
		s = "RAPID ON"
	}
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 4, 51, s, colText)
}

func drawTitleSound(muted bool) {
	display.FillRectangle(0, 200, scrW, 24, colBG)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 6, 218, m5stickc.SoundLabel(muted), colText)
}

func gameOver(btnA, btnB m5stickc.Button, bz *m5stickc.Buzzer, score int) {
	m5stickc.GameOverJingle(bz)

	display.FillRectangle(6, 100, scrW-12, 80, color.RGBA{0, 0, 0, 255})
	tinyfont.WriteLine(display, &freemono.Bold12pt7b, 22, 128, "GAME", colText)
	tinyfont.WriteLine(display, &freemono.Bold12pt7b, 22, 152, "OVER", colText)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 22, 176, "A: retry", colText)

	m5stickc.WaitRetry(btnA, btnB, bz, func() { drawSound(bz.Muted()) })
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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
