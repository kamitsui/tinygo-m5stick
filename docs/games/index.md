# レトロゲーム — 概要・設計思想

M5StickC Plus2（縦 135×240）上で動く 6 つのレトロゲーム（Snake / 2048 / Pong / Dinosaur / Tetris / Invaders）と、それらを束ねる**ランチャー**の設計をまとめる。

- ソース: [games/](https://github.com/kamitsui/tinygo-m5stick/tree/main/games) / [cmd/](https://github.com/kamitsui/tinygo-m5stick/tree/main/cmd) / 共通パッケージ [pkg/m5stickc](https://github.com/kamitsui/tinygo-m5stick/tree/main/pkg/m5stickc)
- 続き: [アーキテクチャ](./architecture)・[ハード/ソフトの仕組み](./hardware)・[各ゲームの実装](./implementations)

## 設計思想

### 1. 「薄いヘルパの合成」— エンジン化しない
Go は継承中心の OOP ではなく、**合成（composition）と小さなインターフェース**が基本。本作も `Update()/Render()` を強制する“ゲームエンジン”は作らない。各ゲームは**描画方式・入力・タイミングが大きく異なる**ため、無理に統一すると漏れのある抽象になりやすい。

代わりに、**安定して重複する部分だけ**を `pkg/m5stickc` の薄いヘルパとして共有し、各ゲームは自分のループ・状態・描画を持つ。

- 共有する（重複を排除）: デバイス初期化、タイトル待ち、ゲームオーバー、サウンド切替、ボタンのエッジ検出、ちらつき防止の描画基盤。
- 共有しない（各ゲームが持つ＝**許容する重複**）: ゲーム固有のルール・状態・描画（Canvas 一括転送 / 部分再描画の選択を含む）。

> _"A little copying is better than a little dependency."_（Go の格言）— 過度な DRY を避け、変更が独立する単位で部品化する。

### 2. 拡張・改造・メンテナンス性
- 新しいゲームは `Run(*Console, *IMU)` を実装し、ランチャーのメニュー配列に1行追加するだけ。
- 共通の操作・サウンド・ライフサイクルはヘルパに集約済みなので、各ゲームの本体はロジックに集中できる。
- 個別ビルド（`cmd/<name>`）と統合ビルド（`cmd/launcher`）の両方を維持し、単体テスト・デバッグがしやすい。

### 3. 組み込み（TinyGo）の制約を踏まえる
- 重い抽象より薄い部品。インターフェースはランチャーが各ゲームを呼ぶ最小限（関数値 `func(*Console, *IMU)`）に留める。
- メモリは Canvas（フレームバッファ）を使うゲームのみ実行時に確保。PWM 非対応・チルト（IMU）など[ハードの制約](./hardware)を前提に実装。

## 操作の統一

全ゲーム・全画面で操作体系をそろえた（例外: Snake は A/B とも操舵に使うためサウンド切替なし）。

| 入力 | 意味 |
|---|---|
| **A タップ** | その画面の主操作（メニュー=カーソル移動 / タイトル=開始 / ゲームオーバー=リトライ / ゲーム中=各ゲームの操作） |
| **A 長押し(~0.4s)** | 画面遷移（メニュー⇔ゲーム）。メニューでは決定、タイトル/ゲームオーバーではメニューへ戻る |
| **B** | サウンド ON/OFF（メニュー/タイトル/プレイ/ゲームオーバーで切替でき、状態は保持） |
| **チルト** | 自機・タイル・パドルの移動（2048/Pong/Tetris/Invaders）。IMU(MPU6886) を利用 |

## 構成（cmd レイアウト）

```
projects/   … チュートリアル（package main）
games/      … 各ゲームのロジック（ライブラリ package、Run(*Console,*IMU) を公開）
cmd/<name>/ … 各ゲームを単体で書き込む薄いラッパー（make flash PROJ=<name>）
cmd/launcher … 6ゲーム入りの単一バイナリ（メニューで選択）
pkg/m5stickc … 共通ハードウェア＋UIヘルパ
```

詳細は[アーキテクチャ](./architecture)へ。
