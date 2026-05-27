// 2048 — レトロゲーム #2
//
// M5StickC Plus2 (135x240 縦) で動く 2048。操作は**チルト（傾け）**:
// 本体を傾けた方向にタイルがスライドする。いったん水平に戻してから次の傾け。
// ゲームオーバー後はボタンAでリスタート。
//
// 書き込み: make flash PROJ=2048
package g2048

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
	cell     = 30
	gap      = 3
	gridLeft = 3
	gridTop  = 46
	gridSpan = 4*cell + 3*gap // 129
	tiltTh   = 350000         // 0.35g
	frames   = 6              // スライドのコマ数
)

var (
	colBG    = color.RGBA{28, 26, 24, 255}
	colField = color.RGBA{50, 46, 42, 255} // 盤面（空きマス）の色
	colBar   = color.RGBA{0, 0, 160, 255}
	colText  = color.RGBA{255, 255, 255, 255}
	colDark  = color.RGBA{119, 110, 101, 255}
)

type movement struct {
	fromR, fromC, toR, toC int
	val                    int16
}

var (
	display *m5stickc.Display
	field   *m5stickc.Canvas // 盤面のオフスクリーン描画先
	board   [4][4]int16
)

// Run はランチャー（または cmd/2048）から呼ばれるエントリ。IMU(チルト)必須。
func Run(con *m5stickc.Console, imu *m5stickc.IMU) {
	display = con.Display
	field = m5stickc.NewCanvas(gridSpan, gridSpan)
	btnA, btnB, buzzer := con.BtnA, con.BtnB, con.Buzzer

	if imu == nil {
		display.FillScreen(colBG)
		tinyfont.WriteLine(display, &freemono.Bold9pt7b, 8, 80, "IMU FAIL", color.RGBA{230, 40, 40, 255})
		time.Sleep(2 * time.Second)
		return
	}

	seed, exit := title(imu, btnA, btnB, buzzer)
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

// title はタイトルを表示し、(乱数シード, メニューへ戻るか) を返す。
// 開始はチルト（A待ちではない）なので共通 WaitStart は使わず、待機中の
// B サウンド切替・A長押しでメニュー復帰を自前で行う。
func title(imu *m5stickc.IMU, btnA, btnB m5stickc.Button, bz *m5stickc.Buzzer) (int64, bool) {
	display.FillScreen(colBG)
	tinyfont.WriteLine(display, &freemono.Bold24pt7b, 14, 104, "2048", color.RGBA{237, 194, 46, 255})
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 10, 142, "Tilt to play", colText)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 10, 168, "hold A:menu", colText)
	drawSound(bz.Muted())

	for btnA.Pressed() { // 持ち越しの A を離す
		time.Sleep(10 * time.Millisecond)
	}
	soundBtn := m5stickc.NewEdgeButton(btnB)
	var n int64
	held := 0
	for {
		if soundBtn.Tapped() {
			bz.ToggleMuted()
			drawSound(bz.Muted())
		}
		if imu.Tilt(tiltTh) != m5stickc.DirNone {
			return n + time.Now().UnixNano(), false // 傾けて開始
		}
		if btnA.Pressed() {
			held++
			if held >= 90 { // A長押し → メニュー
				for btnA.Pressed() {
					time.Sleep(10 * time.Millisecond)
				}
				return n + time.Now().UnixNano(), true
			}
		} else {
			held = 0
		}
		n++
		time.Sleep(5 * time.Millisecond)
	}
}

// play は1ゲーム実行し、リトライ(true)/メニュー復帰(false)を返す。
func play(imu *m5stickc.IMU, btnA, btnB m5stickc.Button, bz *m5stickc.Buzzer, rng *rand.Rand) bool {
	board = [4][4]int16{}
	score := 0
	spawn(rng)
	spawn(rng)
	display.FillScreen(colBG)
	drawBar(score)
	drawBoard()
	drawSound(bz.Muted())

	ready := false
	soundBtn := m5stickc.NewEdgeButton(btnB)
	for {
		// ボタンB: 音の ON/OFF 切替。
		if soundBtn.Tapped() {
			bz.ToggleMuted()
			drawSound(bz.Muted())
		}

		d := imu.Tilt(tiltTh)
		if d == m5stickc.DirNone {
			ready = true
			time.Sleep(50 * time.Millisecond)
			continue
		}
		if !ready {
			time.Sleep(50 * time.Millisecond)
			continue
		}
		ready = false

		result, moves, gained, changed := computeMove(d)
		if !changed {
			continue
		}
		animate(moves)
		board = result
		score += gained
		bz.Tone(m5stickc.NoteC6, 25)
		spawn(rng)
		drawBar(score)
		drawBoard()

		if !movesPossible() {
			return gameOver(btnA, btnB, bz, score)
		}
	}
}

// --- 盤面ロジック ---

// cellAt は方向 d・ライン番号・ライン内インデックス(0が寄せ先)を盤面座標に変換する。
func cellAt(d m5stickc.Dir, lineNo, idx int) (r, c int) {
	switch d {
	case m5stickc.DirRight:
		return lineNo, 3 - idx
	case m5stickc.DirUp:
		return idx, lineNo
	case m5stickc.DirDown:
		return 3 - idx, lineNo
	default: // DirLeft
		return lineNo, idx
	}
}

// computeMove は方向 d のスライド結果・各タイルの移動・獲得点・変化有無を返す。
func computeMove(d m5stickc.Dir) (result [4][4]int16, moves []movement, gained int, changed bool) {
	moves = make([]movement, 0, 16)
	for line := 0; line < 4; line++ {
		var cv [4]int16
		var cs [4]int
		n := 0
		for idx := 0; idx < 4; idx++ {
			r, c := cellAt(d, line, idx)
			if board[r][c] != 0 {
				cv[n] = board[r][c]
				cs[n] = idx
				n++
			}
		}
		var out [4]int16
		oi := 0
		for i := 0; i < n; {
			fr, fc := cellAt(d, line, cs[i])
			if i+1 < n && cv[i] == cv[i+1] {
				out[oi] = cv[i] * 2
				gained += int(cv[i] * 2)
				tr, tc := cellAt(d, line, oi)
				moves = append(moves, movement{fr, fc, tr, tc, cv[i]})
				fr2, fc2 := cellAt(d, line, cs[i+1])
				moves = append(moves, movement{fr2, fc2, tr, tc, cv[i]})
				oi++
				i += 2
			} else {
				out[oi] = cv[i]
				tr, tc := cellAt(d, line, oi)
				moves = append(moves, movement{fr, fc, tr, tc, cv[i]})
				oi++
				i++
			}
		}
		for idx := 0; idx < 4; idx++ {
			r, c := cellAt(d, line, idx)
			result[r][c] = out[idx]
		}
	}
	return result, moves, gained, result != board
}

func spawn(rng *rand.Rand) {
	var empties [16][2]int
	n := 0
	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			if board[r][c] == 0 {
				empties[n] = [2]int{r, c}
				n++
			}
		}
	}
	if n == 0 {
		return
	}
	e := empties[rng.Intn(n)]
	val := int16(2)
	if rng.Intn(10) == 0 {
		val = 4
	}
	board[e[0]][e[1]] = val
}

func movesPossible() bool {
	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			if board[r][c] == 0 {
				return true
			}
			if c+1 < 4 && board[r][c] == board[r][c+1] {
				return true
			}
			if r+1 < 4 && board[r][c] == board[r+1][c] {
				return true
			}
		}
	}
	return false
}

// --- 描画 ---

// cellLocal は盤面キャンバス内のローカル座標（マス左上）を返す。
func cellLocal(r, c int) (x, y int16) {
	return int16(c * (cell + gap)), int16(r * (cell + gap))
}

func drawBar(score int) {
	display.FillRectangle(0, 0, 135, 30, colBar)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 6, 20, "SCORE "+strconv.Itoa(score), colText)
}

// blitField は盤面キャンバスを画面に一括転送する（ちらつき無し）。
func blitField() {
	display.DrawBuffer(gridLeft, gridTop, gridSpan, gridSpan, field.Buffer())
}

func drawBoard() {
	field.Fill(colField)
	for r := 0; r < 4; r++ {
		for c := 0; c < 4; c++ {
			if board[r][c] != 0 {
				x, y := cellLocal(r, c)
				renderTile(x, y, board[r][c])
			}
		}
	}
	blitField()
}

// animate は各タイルを移動元→移動先へ補間し、オフスクリーン描画→一括転送する。
func animate(moves []movement) {
	for f := int16(1); f <= frames; f++ {
		field.Fill(colField)
		for _, m := range moves {
			fx, fy := cellLocal(m.fromR, m.fromC)
			tx, ty := cellLocal(m.toR, m.toC)
			x := fx + (tx-fx)*f/frames
			y := fy + (ty-fy)*f/frames
			renderTile(x, y, m.val)
		}
		blitField()
		time.Sleep(16 * time.Millisecond)
	}
}

// label は表示文字列。4桁以上は K 表記（1024→1K, 2048→2K, 16384→16K）。
func label(v int16) string {
	if v >= 1024 {
		return strconv.Itoa(int(v)/1024) + "K"
	}
	return strconv.Itoa(int(v))
}

// renderTile は盤面キャンバスのローカル座標 (x,y) にタイルを描く。
func renderTile(x, y, val int16) {
	field.FillRectangle(x, y, cell, cell, tileColor(val))
	if val == 0 {
		return
	}
	s := label(val)
	if len(s) <= 2 {
		w := int16(len(s) * 14) // freemono Bold12pt7b の等幅
		tinyfont.WriteLine(field, &freemono.Bold12pt7b, x+(cell-w)/2, y+21, s, tileText(val))
	} else {
		w := int16(len(s) * 11) // freemono Bold9pt7b の等幅
		tinyfont.WriteLine(field, &freemono.Bold9pt7b, x+(cell-w)/2, y+20, s, tileText(val))
	}
}

func tileColor(v int16) color.RGBA {
	switch v {
	case 0:
		return colField
	case 2:
		return color.RGBA{238, 228, 218, 255}
	case 4:
		return color.RGBA{237, 224, 200, 255}
	case 8:
		return color.RGBA{242, 177, 121, 255}
	case 16:
		return color.RGBA{245, 149, 99, 255}
	case 32:
		return color.RGBA{246, 124, 95, 255}
	case 64:
		return color.RGBA{246, 94, 59, 255}
	case 128:
		return color.RGBA{237, 207, 114, 255}
	case 256:
		return color.RGBA{237, 204, 97, 255}
	case 512:
		return color.RGBA{237, 200, 80, 255}
	case 1024:
		return color.RGBA{237, 197, 63, 255}
	default:
		return color.RGBA{237, 194, 46, 255} // 2048+
	}
}

func tileText(v int16) color.RGBA {
	if v <= 4 {
		return colDark
	}
	return colText
}

// drawSound は画面下部に音のON/OFF状態（Bで切替）を表示する。
func drawSound(muted bool) {
	display.FillRectangle(0, 200, 135, 24, colBG)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 8, 216, m5stickc.SoundLabel(muted), colText)
}

func gameOver(btnA, btnB m5stickc.Button, bz *m5stickc.Buzzer, score int) bool {
	m5stickc.GameOverJingle(bz)

	display.FillRectangle(0, 95, 135, 95, colBG)
	tinyfont.WriteLine(display, &freemono.Bold12pt7b, 24, 120, "GAME", colText)
	tinyfont.WriteLine(display, &freemono.Bold12pt7b, 24, 144, "OVER", colText)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 22, 166, "A: retry", colText)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 22, 184, "hold A:menu", colText)

	return m5stickc.WaitRetryOrExit(btnA, btnB, bz, func() { drawSound(bz.Muted()) })
}
