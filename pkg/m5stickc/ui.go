package m5stickc

import "time"

// このファイルは各ゲームで重複しがちな「ライフサイクル/入力」の薄いヘルパを
// 提供する（合成して使う前提。ゲーム固有のループや描画は各ゲームが持つ）。

// Console は M5StickC Plus2 の基本デバイスをまとめて初期化したもの。
// IMU が要るゲームは別途 NewIMU を呼ぶ（初期化失敗があり得るため分離）。
type Console struct {
	Display *Display
	BtnA    Button
	BtnB    Button
	Buzzer  *Buzzer
}

// NewConsole は電源保持を行い、表示・ボタンA/B・ブザーを初期化して返す。
func NewConsole() *Console {
	HoldPower()
	return &Console{
		Display: NewDisplay(),
		BtnA:    NewButton(ButtonAPin),
		BtnB:    NewButton(ButtonBPin),
		Buzzer:  NewBuzzer(BuzzerPin),
	}
}

// EdgeButton はボタンの立ち上がり（押した瞬間）を検出する。
// `if pressed && !prev` の定型を置き換える。
type EdgeButton struct {
	btn  Button
	prev bool
}

// NewEdgeButton は b に対するエッジ検出器を返す。
func NewEdgeButton(b Button) *EdgeButton { return &EdgeButton{btn: b} }

// Tapped は押した瞬間だけ true を返す。
func (e *EdgeButton) Tapped() bool {
	now := e.btn.Pressed()
	tapped := now && !e.prev
	e.prev = now
	return tapped
}

// SoundLabel は音ON/OFF表示の標準文字列を返す（描画位置は各ゲームが決める）。
func SoundLabel(muted bool) string {
	if muted {
		return "B:Sound OFF"
	}
	return "B:Sound ON"
}

// WaitStart はタイトルで A が押されるまで待つ。待機中は B でミュート切替でき
// （切替後 onToggle を呼ぶ＝各ゲームが表示更新）、A 押下までの待ち時間を加えた
// 乱数シードを返す。
func WaitStart(a, b Button, bz *Buzzer, onToggle func()) int64 {
	be := NewEdgeButton(b)
	var n int64
	for !a.Pressed() {
		if be.Tapped() {
			bz.ToggleMuted()
			if onToggle != nil {
				onToggle()
			}
		}
		n++
		time.Sleep(5 * time.Millisecond)
	}
	for a.Pressed() { // 離されるまで待つ
		time.Sleep(10 * time.Millisecond)
	}
	return n + time.Now().UnixNano()
}

// GameOverJingle は共通の下降音（ゲームオーバー）を鳴らす。
func GameOverJingle(bz *Buzzer) {
	bz.Tone(NoteG4, 120)
	bz.Tone(NoteE4, 120)
	bz.Tone(NoteC4, 240)
}

// WaitRetry はゲームオーバー後に A が押されるのを待つ（一度離してから押下）。
// 待機中は B でミュート切替でき（切替後 onToggle）。
func WaitRetry(a, b Button, bz *Buzzer, onToggle func()) {
	for a.Pressed() { // まず離す
		time.Sleep(20 * time.Millisecond)
	}
	be := NewEdgeButton(b)
	for !a.Pressed() {
		if be.Tapped() {
			bz.ToggleMuted()
			if onToggle != nil {
				onToggle()
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
}
