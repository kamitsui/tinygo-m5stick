# CLAUDE.md

TinyGo を M5StickC Plus2 (ESP32-PICO-V3-02 / Xtensa) で動かすリポジトリ。小さなプロジェクトを複数作り、最終的にレトロゲーム集をパッケージ化する。詳細な構想は `concept.md`、進行は GitHub Issues（マイルストーン M1: 基盤整備 / M2: レトロゲーム）で管理する。

## ビルド・書き込み（Makefile 経由が必須）

直接 `tinygo` を叩かず、必ず Makefile を使う。

```bash
make list                     # projects/ 一覧
make build PROJ=02-display    # ビルド
make flash PROJ=03-buzzer     # 実機書き込み（PORT は WCH USBシリアルを自動検出）
make flash PROJ=01-blink PORT=/dev/cu.xxx   # ポート指定
make help                     # ターゲット一覧と現在の変数
```

- ターゲットは `esp32-coreboard-v2`（M5StickC Plus2 専用ターゲットは TinyGo に無いため汎用ESP32で代用）。
- `tinygo` / `esptool` は Homebrew 導入。Apple Silicon では **PATH に `/opt/homebrew/bin`** が必要。
- ビルド生成物は `build/`（gitignore 済み）。

## リポジトリ構成

- `projects/NN-name/` — 単体で動く main（01-blink〜04-button はチュートリアル、05-m5lib-demo はパッケージ実演）。
- `pkg/m5stickc/` — 共通ハードウェアパッケージ。新規プロジェクトはこれを再利用する。
  - `HoldPower()` / `LED` / `NewButton()` / `NewBuzzer().Tone()` / `NewDisplay()` / 音名定数。
- モジュールパス: `github.com/kamitsui/tinygo-m5stick`。

## ハードウェアの要点（検証済み）

- ピン: LED=G19, HOLD=G4, BtnA=G37, BtnB=G39, BtnC=G35, Buzzer=G2。表示 SPI2: SCK=G13, MOSI=G15, RST=G12, DC=G14, CS=G5, BL=G27。
- **表示器 ST7789V2**: ドライバ(st7789 v0.35.0)は Rotation0 でオフセットを 0 に潰す。検証済みの `Rotation180 + ColumnOffset 53 + RowOffset 40` を使う。`pkg/m5stickc` の `NewDisplay()` はこれをラップした `*Display` を返し、**論理座標を反転して自然な向き（USB-C を下）**で使えるようにしている（ゲーム向け）。raw にドライバを使う 02-display 等は Rotation180（USB-C 上）のまま。
- **ブザー**: Xtensa ESP32 は PWM(LEDC) 非対応。`tone` ドライバは使えないので **GPIO ソフトトグル**で矩形波を出す（`Buzzer.Tone`）。
- **ボタン**: GPIO34-39 は入力専用・内部プルアップ無し。基板側に外部プルアップがあり `PinInput` で読める。押下=LOW（`Button.Pressed()` が反転して返す）。
- WiFi/BLE は Xtensa ESP32 では TinyGo 非対応（C3/S3 のみ）。本リポジトリの用途では不要。

## 開発フロー・規約

- **コミットは英語の Conventional Commits**（例: `feat(m5stickc): ...`, `build: ...`）。
- 作業は **feature ブランチ** で行い、**PR で main にマージ**（squash, ブランチ削除）。PR 本文に `Closes #N` を付ける。
- `.gitignore` は**ホワイトリスト方式**。新しい拡張子・拡張子なしファイルを追加したら、追跡されるか `git status` で確認（取りこぼしに注意）。
- 実機検証が必要な変更は、可能なら `make flash` で動作確認してからコミットする。
