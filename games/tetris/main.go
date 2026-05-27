// Tetris — レトロゲーム #5
//
// 縦画面の Tetris。操作:
//   - チルト左右: ピースを移動
//   - チルト下:   ソフトドロップ（速く落ちる）
//   - ボタンA:    回転（時計回り）
//   - ボタンB:    音 ON/OFF
// ライン消去でスコア、上まで積むとゲームオーバー（A でリトライ）。
//
// 画面: 左に枠付きの盤面、右に NEXT（次のピース）を表示。
//
// 書き込み: make flash PROJ=tetris
package tetris

import (
	"image/color"
	"math/rand"
	"strconv"
	"time"

	"github.com/kamitsui/tinygo-m5stick/pkg/m5stickc"
	"tinygo.org/x/tinyfont"
	"tinygo.org/x/tinyfont/freemono"
	"tinygo.org/x/tinyfont/proggy"
)

const (
	cols    = 10
	rows    = 20
	cell    = 9
	boardX0 = 2
	boardY0 = 44
	boardW  = cols * cell // 90
	boardH  = rows * cell // 180
	panelX  = 100         // 右パネル(NEXT)のX
	tiltTh  = 250000
	frameMS = 30
)

var (
	colBG     = color.RGBA{8, 8, 12, 255}     // 盤面の外側
	colBoard  = color.RGBA{22, 22, 30, 255}   // 盤面（空きマス）
	colBorder = color.RGBA{90, 90, 120, 255}  // 盤面の枠線
	colEdge   = color.RGBA{12, 12, 16, 255}   // 各ブロックの縁
	colBar    = color.RGBA{0, 0, 80, 255}
	colText   = color.RGBA{255, 255, 255, 255}
)

// pieceColors[0] は空き(盤面色)。1..7 が各テトリミノ。
var pieceColors = [8]color.RGBA{
	colBoard,
	{0, 220, 220, 255},  // I シアン
	{230, 220, 0, 255},  // O 黄
	{180, 60, 220, 255}, // T 紫
	{0, 210, 60, 255},   // S 緑
	{230, 50, 50, 255},  // Z 赤
	{50, 90, 230, 255},  // J 青
	{235, 140, 30, 255}, // L 橙
}

type pieceDef struct {
	box   int
	cells [4][2]int8
}

// 各ピースの回転0状態（box 内のセル）。
var pieces = [7]pieceDef{
	{4, [4][2]int8{{1, 0}, {1, 1}, {1, 2}, {1, 3}}}, // I
	{2, [4][2]int8{{0, 0}, {0, 1}, {1, 0}, {1, 1}}}, // O
	{3, [4][2]int8{{0, 1}, {1, 0}, {1, 1}, {1, 2}}}, // T
	{3, [4][2]int8{{0, 1}, {0, 2}, {1, 0}, {1, 1}}}, // S
	{3, [4][2]int8{{0, 0}, {0, 1}, {1, 1}, {1, 2}}}, // Z
	{3, [4][2]int8{{0, 0}, {1, 0}, {1, 1}, {1, 2}}}, // J
	{3, [4][2]int8{{0, 2}, {1, 0}, {1, 1}, {1, 2}}}, // L
}

// テトリスのテーマ（Korobeiniki）冒頭。{周波数, ミリ秒}。テンポはゆっくりめ。
var melody = []struct{ f, d int }{
	{m5stickc.NoteE5, 380}, {m5stickc.NoteB4, 190}, {m5stickc.NoteC5, 190}, {m5stickc.NoteD5, 380}, {m5stickc.NoteC5, 190}, {m5stickc.NoteB4, 190},
	{m5stickc.NoteA4, 380}, {m5stickc.NoteA4, 190}, {m5stickc.NoteC5, 190}, {m5stickc.NoteE5, 380}, {m5stickc.NoteD5, 190}, {m5stickc.NoteC5, 190},
	{m5stickc.NoteB4, 560}, {m5stickc.NoteC5, 190}, {m5stickc.NoteD5, 380}, {m5stickc.NoteE5, 380},
	{m5stickc.NoteC5, 380}, {m5stickc.NoteA4, 380}, {m5stickc.NoteA4, 560}, {0, 250},
}

// playMelody はメロディを1回再生する（ブロッキング）。再生中も btnB でミュート
// 切替でき、ミュートすると以降の音は即座に無音で進む（実質中断）。
func playMelody(bz *m5stickc.Buzzer, btnA, btnB m5stickc.Button) {
	soundBtn := m5stickc.NewEdgeButton(btnB)
	skipBtn := m5stickc.NewEdgeButton(btnA)
	for _, n := range melody {
		if skipBtn.Tapped() {
			return // A でスキップしてゲーム開始
		}
		if soundBtn.Tapped() {
			bz.ToggleMuted()
			drawSound(bz.Muted())
		}
		if n.f == 0 {
			bz.Tone(0, n.d)
			continue
		}
		bz.Tone(n.f, n.d-14)
		bz.Tone(0, 14) // 音の区切り
	}
}

type piece struct {
	kind       int
	rot        int
	posR, posC int
}

func (p piece) cells() [4][2]int {
	def := pieces[p.kind]
	n := def.box
	var out [4][2]int
	for i, c := range def.cells {
		r, col := int(c[0]), int(c[1])
		for k := 0; k < p.rot%4; k++ {
			r, col = col, n-1-r // CW 回転
		}
		out[i] = [2]int{p.posR + r, p.posC + col}
	}
	return out
}

var (
	display  *m5stickc.Display
	grid     [rows][cols]uint8
	cur      piece
	nextKind int
)

// Run はランチャー（または cmd/tetris）から呼ばれるエントリ。IMU(チルト)必須。
func Run(con *m5stickc.Console, imu *m5stickc.IMU) {
	display = con.Display
	btnA, btnB, buzzer := con.BtnA, con.BtnB, con.Buzzer

	if imu == nil {
		display.FillScreen(colBG)
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
	display.FillScreen(colBG)
	tinyfont.WriteLine(display, &freemono.Bold18pt7b, 6, 84, "TETRIS", colText)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 6, 120, "Tilt: move", colText)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 6, 140, "Tilt v: drop", colText)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 6, 160, "A: rotate", colText)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 6, 180, "hold A:menu", colText)
	drawTitleSound(bz.Muted())
	return m5stickc.WaitStart(btnA, btnB, bz, func() { drawTitleSound(bz.Muted()) })
}

// drawTitleSound はタイトル下部に音のON/OFF（Bで切替）を表示する。
func drawTitleSound(muted bool) {
	display.FillRectangle(0, 205, 135, 24, colBG)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 6, 222, m5stickc.SoundLabel(muted), colText)
}

// play は1ゲーム実行し、リトライ(true)/メニュー復帰(false)を返す。
func play(imu *m5stickc.IMU, btnA, btnB m5stickc.Button, bz *m5stickc.Buzzer, rng *rand.Rand) bool {
	grid = [rows][cols]uint8{}
	score := 0
	lines := 0

	display.FillScreen(colBG)
	drawScore(score)
	drawSound(bz.Muted())
	drawBorder()
	redrawBoard()

	nextKind = rng.Intn(7)
	spawn(rng)
	if collides(cur.cells()) {
		return gameOver(btnA, btnB, bz, score)
	}
	drawPiece(cur, true)

	playMelody(bz, btnA, btnB) // 開始メロディ。A でスキップ、B でミュート可。

	var prevA bool
	soundBtn := m5stickc.NewEdgeButton(btnB)
	moveCD := 0
	gravity := 0
	for {
		fallEvery := 18 - (lines/10)*2
		if fallEvery < 3 {
			fallEvery = 3
		}

		a := btnA.Pressed()
		if soundBtn.Tapped() {
			bz.ToggleMuted()
			drawSound(bz.Muted())
		}
		if a && !prevA {
			if tryRotate() {
				bz.Tone(m5stickc.NoteE5, 15) // 回転音
			}
		}
		prevA = a

		t := imu.Tilt(tiltTh)
		if t == m5stickc.DirLeft || t == m5stickc.DirRight {
			if moveCD <= 0 {
				dc := -1
				if t == m5stickc.DirRight {
					dc = 1
				}
				if tryMove(0, dc) {
					bz.Tone(m5stickc.NoteC4, 10) // 移動音
				}
				moveCD = 4
			} else {
				moveCD--
			}
		} else {
			moveCD = 0
		}

		interval := fallEvery
		if t == m5stickc.DirDown {
			interval = 2 // ソフトドロップ
		}
		gravity++
		if gravity >= interval {
			gravity = 0
			if !tryMove(1, 0) {
				lockPiece()
				if cleared := clearLines(); cleared > 0 {
					lines += cleared
					score += clearScore(cleared)
					drawScore(score)
					redrawBoard()
					bz.Tone(m5stickc.NoteG5, 30) // ライン消去音
					bz.Tone(m5stickc.NoteC6, 50)
				}
				spawn(rng)
				if collides(cur.cells()) {
					return gameOver(btnA, btnB, bz, score)
				}
				drawPiece(cur, true)
			}
		}

		time.Sleep(frameMS * time.Millisecond)
	}
}

// --- ピース操作 ---

func collides(cs [4][2]int) bool {
	for _, c := range cs {
		r, col := c[0], c[1]
		if col < 0 || col >= cols || r >= rows {
			return true
		}
		if r >= 0 && grid[r][col] != 0 {
			return true
		}
	}
	return false
}

func tryMove(dr, dc int) bool {
	np := cur
	np.posR += dr
	np.posC += dc
	if collides(np.cells()) {
		return false
	}
	drawPiece(cur, false)
	cur = np
	drawPiece(cur, true)
	return true
}

func tryRotate() bool {
	np := cur
	np.rot = (np.rot + 1) % 4
	for _, k := range []int{0, -1, 1, -2, 2} { // 簡易ウォールキック
		t := np
		t.posC += k
		if !collides(t.cells()) {
			drawPiece(cur, false)
			cur = t
			drawPiece(cur, true)
			return true
		}
	}
	return false
}

func spawn(rng *rand.Rand) {
	k := nextKind
	cur = piece{kind: k, rot: 0, posR: 0, posC: (cols - pieces[k].box) / 2}
	nextKind = rng.Intn(7)
	drawNext(nextKind)
}

func lockPiece() {
	for _, c := range cur.cells() {
		if c[0] >= 0 {
			grid[c[0]][c[1]] = uint8(cur.kind + 1)
		}
	}
}

func clearLines() int {
	n := 0
	for r := rows - 1; r >= 0; r-- {
		full := true
		for c := 0; c < cols; c++ {
			if grid[r][c] == 0 {
				full = false
				break
			}
		}
		if full {
			for rr := r; rr > 0; rr-- {
				grid[rr] = grid[rr-1]
			}
			grid[0] = [cols]uint8{}
			n++
			r++
		}
	}
	return n
}

func clearScore(n int) int {
	switch n {
	case 1:
		return 100
	case 2:
		return 300
	case 3:
		return 500
	default:
		return 800
	}
}

// --- 描画 ---

func cellXY(r, c int) (int16, int16) {
	return int16(boardX0 + c*cell), int16(boardY0 + r*cell)
}

// drawBlock は縁取り付きのブロックを描く。idx==0 は空きマス（縁なし）。
func drawBlock(x, y, size int16, idx uint8) {
	if idx == 0 {
		display.FillRectangle(x, y, size, size, colBoard)
		return
	}
	display.FillRectangle(x, y, size, size, colEdge)
	display.FillRectangle(x+1, y+1, size-2, size-2, pieceColors[idx])
}

func drawBoardCell(r, c int, idx uint8) {
	x, y := cellXY(r, c)
	drawBlock(x, y, cell, idx)
}

func drawPiece(p piece, on bool) {
	idx := uint8(0)
	if on {
		idx = uint8(p.kind + 1)
	}
	for _, c := range p.cells() {
		if c[0] >= 0 {
			drawBoardCell(c[0], c[1], idx)
		}
	}
}

func drawBorder() {
	display.FillRectangle(boardX0-2, boardY0-2, boardW+4, boardH+4, colBorder)
}

func redrawBoard() {
	display.FillRectangle(boardX0, boardY0, boardW, boardH, colBoard)
	for r := 0; r < rows; r++ {
		for c := 0; c < cols; c++ {
			if grid[r][c] != 0 {
				drawBoardCell(r, c, grid[r][c])
			}
		}
	}
}

// drawNext は右パネルに次のピースを小さく表示する。
func drawNext(kind int) {
	const pc = 6
	ox, oy := int16(panelX), int16(64)
	// 消去矩形は画面幅(135)を超えないようにする（超えると描画が無視される）。
	display.FillRectangle(panelX-4, 50, 135-(panelX-4), 40, colBG) // パネル消去
	tinyfont.WriteLine(display, &proggy.TinySZ8pt7b, panelX, 58, "NEXT", colText)
	for _, c := range pieces[kind].cells {
		x := ox + int16(c[1])*pc
		y := oy + int16(c[0])*pc
		drawBlock(x, y, pc, uint8(kind+1))
	}
}

func drawScore(score int) {
	display.FillRectangle(0, 0, 135, 20, colBar)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 4, 15, "SCORE "+strconv.Itoa(score), colText)
}

func drawSound(muted bool) {
	display.FillRectangle(0, 20, 135, 18, colBar)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 4, 34, m5stickc.SoundLabel(muted), colText)
}

func gameOver(btnA, btnB m5stickc.Button, bz *m5stickc.Buzzer, score int) bool {
	m5stickc.GameOverJingle(bz)

	display.FillRectangle(boardX0, 104, boardW, 96, color.RGBA{0, 0, 0, 255})
	tinyfont.WriteLine(display, &freemono.Bold12pt7b, 8, 130, "GAME", colText)
	tinyfont.WriteLine(display, &freemono.Bold12pt7b, 8, 154, "OVER", colText)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 6, 174, "A:retry", colText)
	tinyfont.WriteLine(display, &freemono.Bold9pt7b, 6, 192, "holdA:menu", colText)

	return m5stickc.WaitRetryOrExit(btnA, btnB, bz, func() { drawSound(bz.Muted()) })
}
