# はじめる（環境構築・書き込み）

::: info このページは執筆中です
詳細は [Issue #7](https://github.com/kamitsui/tinygo-m5stick/issues/7) で執筆予定です。
:::

M5StickC Plus2 に TinyGo で書き込むまでの導線。執筆予定の項目:

- インストール（Homebrew: `tinygo` / `esptool`）と公式リンク
- `make flash` での書き込み手順、シリアルポートの確認
- ピン配置表（LED / ボタン / ブザー / 表示器）
- つまずきポイント（表示器のオフセット、PWM 非対応のブザー など）

## 最短の流れ（先取り）

```bash
# 1. ツールチェーン導入（macOS / Homebrew）
brew tap tinygo-org/tools && brew install tinygo esptool

# 2. ビルドして実機へ書き込み
make flash PROJ=01-blink
```

## 参考リンク

- [TinyGo macOS インストールガイド](https://tinygo.org/getting-started/install/macos/)
- [M5StickC PLUS2 ドキュメント](https://docs.m5stack.com/en/core/M5StickC%20PLUS2)
