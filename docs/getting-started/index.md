# はじめる（環境構築・書き込み）

このページのゴールは、**TinyGo を導入して M5StickC Plus2 に最初のプログラム（LED点滅）を書き込む**ところまで、迷わずたどり着くことです。インストールの細かい手順は公式に譲り、ここでは**全体の流れ**と、実機で実際にハマった**つまずきポイント**を重点的にまとめます。

対象環境: macOS（Apple Silicon）。コマンドは Homebrew 前提です。

## 必要なもの

- **M5StickC Plus2**（ESP32-PICO-V3-02 / Xtensa 系 ESP32）
- USB-C ケーブル（データ通信対応のもの）
- macOS + [Homebrew](https://brew.sh/)

## 1. ツールチェーンの導入

TinyGo 本体と、ESP32 への書き込みに使う `esptool` を入れます。

```bash
brew tap tinygo-org/tools
brew install tinygo esptool
```

導入できたら確認:

```bash
tinygo version   # => tinygo version 0.41.x ...
esptool version  # => 5.x
```

::: warning PATH に注意（Apple Silicon）
Homebrew は `/opt/homebrew/bin` に入ります。`tinygo: command not found` になる場合は、シェルの設定（`~/.zprofile` など）で PATH に `/opt/homebrew/bin` が通っているか確認してください。
:::

- 公式インストールガイド: [TinyGo / macOS](https://tinygo.org/getting-started/install/macos/)
- 概要: [Getting Started](https://tinygo.org/getting-started/overview/)

## 2. デバイスの接続とポート確認

M5StickC Plus2 を USB-C で接続し、シリアルポートを確認します。

```bash
ls /dev/cu.*
# 例: /dev/cu.wchusbserial58FA0591591
```

M5StickC Plus2 は **WCH 製の USB シリアル**チップを使っているため、`cu.wchusbserial…` という名前で現れます（macOS には標準でドライバが入っており、追加導入は通常不要）。

::: tip ポートが見えないとき
ケーブルがデータ通信対応か確認してください（充電専用ケーブルだと認識されません）。それでも出ない場合は WCH のドライバ導入を検討します。
:::

## 3. 最初の書き込み（Lチカ）

このリポジトリでは、ビルド・書き込みを **Makefile** に集約しています。`tinygo` を直接叩く必要はありません。

```bash
make list                 # projects/ の一覧を表示
make flash PROJ=01-blink  # ビルドして書き込み（ポートは自動検出）
```

うまくいくと本体の**赤色 LED が点滅**します。これでツールチェーンから実機までの経路が通ったことの確認になります。

ポートを明示したい場合:

```bash
make flash PROJ=01-blink PORT=/dev/cu.wchusbserial58FA0591591
```

::: details 書き込み時に出る警告・初回エラーについて
- `unable to read security info: ... timeout waiting for response to SLIP read` は、Xtensa 版 ESP32 で出る既知の**無害な警告**です。書き込みは続行されます。
- 初回に `open serial port ...: no such file or directory`（ポート消失）が出ることがあります。USB の再列挙による一過性のことが多く、**もう一度同じコマンドを実行**すれば通ります。
:::

## ビルドターゲットについて

TinyGo には **M5StickC Plus2 専用のターゲットが存在しません**。そこで汎用の ESP32 ターゲット `esp32-coreboard-v2` を使い、ピンの割り当てはコード側で指定しています（Makefile の既定値が `TARGET=esp32-coreboard-v2`）。

## ピン配置（M5StickC Plus2）

| 機能 | ピン | メモ |
|---|---|---|
| 赤色 LED | GPIO19 | IR エミッタと共用 |
| 電源保持 (HOLD) | GPIO4 | バッテリ駆動時に HIGH 保持で電源維持 |
| ボタン A（正面の大ボタン）| GPIO37 | 押下=LOW（入力専用ピン）|
| ボタン B（側面）| GPIO39 | 押下=LOW（入力専用ピン）|
| ボタン C（電源）| GPIO35 | |
| ブザー（パッシブ）| GPIO2 | |
| 表示 SCLK | GPIO13 | SPI2 |
| 表示 MOSI | GPIO15 | SPI2 |
| 表示 RST | GPIO12 | |
| 表示 DC | GPIO14 | |
| 表示 CS | GPIO5 | |
| 表示 バックライト (BL) | GPIO27 | |

これらは共通パッケージ `pkg/m5stickc` に定数として定義済みです。

## つまずきポイント（重要）

実機で動かして分かった、TinyGo + M5StickC Plus2 特有の注意点です。

### 表示器 (ST7789V2) の向きとオフセット
1.14" 135×240 パネルは GRAM の原点からずれて実装されています。ところが TinyGo の `st7789` ドライバ（v0.35.0）は **Rotation0 のとき RowOffset/ColumnOffset を 0 に潰す**実装になっており、そのままだと表示が左にずれます。

検証済みの正しい設定は **`Rotation180` + `ColumnOffset: 53` + `RowOffset: 40`** です（`ColumnOffset` は 52 だと左端に 1px 残るため 53）。この向きでは **USB-C を上にすると文字が正立**します。`pkg/m5stickc` の `NewDisplay()` がこの設定で初期化済みです。

### ブザーは PWM が使えない
TinyGo の Xtensa 版 ESP32 は **PWM(LEDC) 未対応**です。そのため PWM を前提とする `tinygo.org/x/drivers/tone` は使えません。本リポジトリでは **GPIO をソフトでトグルして矩形波を作る**方式（`pkg/m5stickc` の `Buzzer.Tone`）で音を鳴らしています。

### ボタンは active-low / 入力専用ピン
ボタンに使う GPIO34〜39 は**入力専用で内部プルアップがありません**。基板側に外部プルアップがあるため `PinInput` で読めます。押下時は **LOW** になるので、`pkg/m5stickc` の `Button.Pressed()` は反転して「押されていれば true」を返します。

### WiFi / BLE は使えない
Xtensa 版 ESP32 の無線は、現時点の TinyGo では非対応です（無線対応は ESP32-C3 / S3 のみ）。本リポジトリのチュートリアルとゲームでは無線を使わないため影響はありません。

## 共通パッケージを使う

新しいプログラムでは、検証済みの設定が入った `pkg/m5stickc` を再利用できます。

```go
package main

import (
	"image/color"
	"github.com/kamitsui/tinygo-m5stick/pkg/m5stickc"
)

func main() {
	m5stickc.HoldPower()                       // 電源保持
	display := m5stickc.NewDisplay()           // 表示器（向き・オフセット設定済み）
	display.FillScreen(color.RGBA{0, 0, 0, 255})

	buzzer := m5stickc.NewBuzzer(m5stickc.BuzzerPin)
	buzzer.Tone(m5stickc.NoteC5, 200)          // ド を 200ms

	btnA := m5stickc.NewButton(m5stickc.ButtonAPin)
	_ = btnA.Pressed()                         // 押されていれば true
}
```

## 次のステップ

- サンプルを一通り試す: `01-blink`（LED）→ `02-display`（表示器）→ `03-buzzer`（音）→ `04-button`（入力）→ `05-m5lib-demo`（共通パッケージのデモ）
- [TinyGo とは](/tinygo/) で TinyGo の背景を知る
- ソースと進行状況: [GitHub リポジトリ](https://github.com/kamitsui/tinygo-m5stick)

## 参考リンク

- [TinyGo macOS インストールガイド](https://tinygo.org/getting-started/install/macos/)
- [TinyGo Getting Started](https://tinygo.org/getting-started/overview/)
- [TinyGo machine パッケージ (m5stack)](https://tinygo.org/docs/reference/microcontrollers/machine/m5stack/)
- [M5StickC PLUS2 公式ドキュメント](https://docs.m5stack.com/en/core/M5StickC%20PLUS2)
