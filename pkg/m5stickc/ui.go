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

// WaitStart はタイトルで入力を待ち、(乱数シード, メニューへ戻るか) を返す。
//   - A タップ      → ゲーム開始（exit=false）
//   - A 長押し(~0.45s) → トップメニューへ戻る（exit=true。間違えて選んだ時用）
//   - B             → ミュート切替（onToggle!=nil のとき。切替後 onToggle 呼び出し）
// 最初に A の離しを待つので、メニューの選択操作（A長押し）が持ち越されても誤発火しない。
func WaitStart(a, b Button, bz *Buzzer, onToggle func()) (seed int64, exit bool) {
	for a.Pressed() { // 持ち越しの A を離すまで待つ
		time.Sleep(10 * time.Millisecond)
	}
	be := NewEdgeButton(b)
	var n int64
	held := 0
	for {
		if onToggle != nil && be.Tapped() {
			bz.ToggleMuted()
			onToggle()
		}
		if a.Pressed() {
			held++
			if held >= 90 { // 長押し → メニューへ
				for a.Pressed() {
					time.Sleep(10 * time.Millisecond)
				}
				return n + time.Now().UnixNano(), true
			}
		} else if held > 0 { // タップで離した → 開始
			return n + time.Now().UnixNano(), false
		}
		n++
		time.Sleep(5 * time.Millisecond)
	}
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

// WaitRetryOrExit はゲームオーバー後の入力を待ち、A タップ=リトライ(true) /
// A 長押し(約0.5秒)=メニューへ戻る(false) を返す。待機中は B でミュート切替
// でき（onToggle!=nil のときのみ）。ランチャーから呼ばれるゲームで使う。
func WaitRetryOrExit(a, b Button, bz *Buzzer, onToggle func()) bool {
	for a.Pressed() { // まず離す
		time.Sleep(20 * time.Millisecond)
	}
	be := NewEdgeButton(b)
	held := 0
	for {
		if onToggle != nil && be.Tapped() {
			bz.ToggleMuted()
			onToggle()
		}
		if a.Pressed() {
			held++
			if held >= 24 { // 長押し → メニューへ
				for a.Pressed() {
					time.Sleep(10 * time.Millisecond)
				}
				return false
			}
		} else if held > 0 { // 短押しで離した → リトライ
			return true
		}
		time.Sleep(20 * time.Millisecond)
	}
}
