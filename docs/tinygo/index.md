# TinyGo とは

[TinyGo](https://tinygo.org/) は、**マイコン・WebAssembly・小さな CLI ツール**といった「小さな場所」で Go を動かすための Go コンパイラです。標準の Go コンパイラとは別物で、[LLVM](https://llvm.org/) を backend に使って、限られたメモリ・容量でも動く非常にコンパクトなバイナリを生成します。

> Go compiler for small places. Microcontrollers, WebAssembly (WASM/WASI), and command-line tools. Based on LLVM.

## 歴史

TinyGo はオランダの開発者 **Ayke van Laëthem** 氏によって始められました。氏はもともと MicroPython など組み込み系プロジェクトに関わっており、「Go をマイコンで動かす」試みとして TinyGo を立ち上げました。その後 Ron Evans 氏（gobot プロジェクト等）らコミュニティを巻き込み、FOSDEM などで発信されながら成長してきました。

LLVM という成熟した最適化基盤の上に作られていることが、多様なアーキテクチャ対応と小さな出力を可能にしています。現在も活発に開発されており、本稿執筆時点（2026年5月）の最新は **0.41 系**です（0.41 では ESP32-C3 / S3 向けの無線サポートが追加されました）。

## 特徴

- **対応ターゲットが広い**: 150 以上のマイコンボード（BBC micro:bit、Arduino、Nordic / ST のチップなど）に加え、WebAssembly を出力できます。
- **バイナリが小さい**: 標準 Go のランタイムをそのまま積まず、必要最小限に切り詰めるため、数十 KB 級のマイコンにも載ります。
- **Go の言語仕様の「サブセット」**: goroutine やチャネルなど多くの機能は使えますが、リフレクションの一部制限や、軽量なガベージコレクタ、協調的スケジューラなど、標準 Go とは挙動が異なる部分があります。組み込みでは「使える機能・使えない機能」を意識する必要があります。
- **ペリフェラル抽象 `machine` パッケージ**: GPIO / SPI / I2C / UART / ADC などをボード非依存に近い形で扱えます（ただし PWM など一部はチップによって未対応 → [はじめる](/getting-started/#つまずきポイント-重要) のブザーの項を参照）。

## 活用事例

- **組み込み / IoT**: センサー、LED、ディスプレイ、ロボティクス（gobot など）。本リポジトリの M5StickC Plus2 もこの領域です。
- **WebAssembly / エッジ**: 非常に小さい WASM を生成できるため、ブラウザだけでなく **WASI** 対応のサーバ／エッジ環境でも動きます。Fastly Compute、Fermyon Spin、wazero などの WebAssembly ランタイムで TinyGo 製モジュールが利用されています。
- **小さな CLI ツール**: 起動が速く小さい実行ファイルを作れます。

## PC でも動かせる？

はい。大きく2つの観点があります。

1. **ロジックは標準 Go で開発・テストできる**: ゲームのルールやデータ構造など、ハードウェアに依存しない部分は普通の `go test` で書いて検証し、表示・入力などハード依存部分だけ TinyGo + 実機で動かす、という分担が可能です。
2. **TinyGo 自身もホスト向けにビルドできる**: `tinygo build` / `tinygo run` はホスト OS をターゲットにもできます。ただし通常の PC アプリ開発であれば標準の Go を使うほうが自然で、TinyGo の主目的はあくまで「小さな場所」です。

::: tip 本リポジトリでの方針
ハード依存の設定（ピン・表示器・ブザー）は `pkg/m5stickc` に集約しています。レトロゲームではゲームロジックとハード I/O を分けておくと、ロジックを PC 側でテストしやすくなります。
:::

## M5StickC Plus2 で動かすには

M5StickC Plus2 は **ESP32-PICO-V3-02（Xtensa 系 ESP32）**を搭載しています。TinyGo で動かすために押さえるべき点:

- **専用ターゲットは無い** → 汎用の `esp32-coreboard-v2` を使い、ピンはコードで指定する。
- 書き込みには `esptool` が必要（Homebrew で導入）。
- **PWM 非対応・無線非対応**（Xtensa ESP32）など、チップ固有の制約がある。

具体的な導入・書き込み手順と、実機で遭遇したつまずきポイントは **[はじめる（環境構築・書き込み）](/getting-started/)** にまとめてあります。

## 参考リンク

- [TinyGo 公式サイト](https://tinygo.org/)
- [Getting Started（概要）](https://tinygo.org/getting-started/overview/)
- [対応マイコン一覧](https://tinygo.org/docs/reference/microcontrollers/)
- [TinyGo GitHub](https://github.com/tinygo-org/tinygo)
- [TinyGo 0.41 — The Big Release（無線対応など）](https://tinygo.org/blog/2026/tinygo-0-41-the-big-release/)
- [machine パッケージ（m5stack 参考）](https://tinygo.org/docs/reference/microcontrollers/machine/m5stack/)
