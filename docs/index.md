---
layout: home

hero:
  name: tinygo-m5stick
  text: TinyGo × M5StickC Plus2
  tagline: 小さなマイコンで Go を動かす。チュートリアルからレトロゲームまで。
  actions:
    - theme: brand
      text: はじめる
      link: /getting-started/
    - theme: alt
      text: TinyGoとは
      link: /tinygo/
    - theme: alt
      text: GitHub
      link: https://github.com/kamitsui/tinygo-m5stick

features:
  - title: 環境構築から実機まで
    details: TinyGo の導入と M5StickC Plus2 への書き込みを、つまずきポイント込みで解説。
  - title: ハードウェアを触る
    details: LED・表示器(ST7789V2)・ブザー・ボタンを TinyGo から動かすサンプル。
  - title: レトロゲーム
    details: Snake / 2048 / Pong / Dinosaur / Tetris / Invaders を作って遊ぶ（開発中）。
---

## このサイトについて

[TinyGo](https://tinygo.org/) を使って [M5StickC Plus2](https://docs.m5stack.com/en/core/M5StickC%20PLUS2)（ESP32-PICO-V3-02）でプログラムを動かすための解説とサンプル集です。

具体的なインストール手順やAPIは公式ドキュメントに譲り、**「初めて TinyGo を触る人が迷子にならない」導線**と、実機で得た**つまずきポイント**をまとめることを重視しています。

- ソースコード: [github.com/kamitsui/tinygo-m5stick](https://github.com/kamitsui/tinygo-m5stick)
- 進行管理: GitHub Issues（マイルストーン M1: 基盤整備 / M2: レトロゲーム）
