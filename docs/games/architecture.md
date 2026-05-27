# アーキテクチャ（C4 / UML）

[C4 モデル](https://c4model.com/)の階層（コンテキスト→コンテナ→コンポーネント）と UML（クラス図・シーケンス図）で全体像を示す。

## C4: システムコンテキスト

誰が・何を使うか。

```mermaid
flowchart LR
  player(["プレイヤー"]) -->|"A / B / チルト"| fw["M5StickC Plus2<br/>ファームウェア"]
  fw -->|"表示・音"| player
  dev(["開発者"]) -->|"make flash"| fw
  dev -->|"執筆"| docs["ドキュメント<br/>(VitePress → GitHub Pages)"]
```

## C4: コンテナ

単一バイナリ（`cmd/launcher`）の中身と依存。

```mermaid
flowchart TB
  subgraph FW["ファームウェア（単一バイナリ: cmd/launcher）"]
    launcher["ランチャー<br/>(メニュー)"]
    games["各ゲーム<br/>games/*: Run(Console, IMU)"]
    pkg["pkg/m5stickc<br/>共通ハード＋UI"]
    drivers["TinyGo drivers<br/>st7789 / mpu6886 / tinyfont"]
  end
  launcher --> games
  launcher --> pkg
  games --> pkg
  pkg --> drivers
  drivers --> hw[("M5StickC Plus2<br/>ESP32 ハードウェア")]
```

個別ビルド `cmd/<name>` も同じ `games/<name>.Run` を呼ぶ薄いラッパー。

## C4: コンポーネント（pkg/m5stickc）

共通パッケージの内部。

```mermaid
flowchart TB
  subgraph pkg["pkg/m5stickc"]
    Console["Console<br/>(デバイス集約)"]
    Display["Display<br/>(ST7789 ラッパ・自然な向き)"]
    Canvas["Canvas<br/>(オフスクリーン)"]
    Button["Button"]
    Buzzer["Buzzer<br/>(ミュート対応)"]
    IMU["IMU<br/>(MPU6886・チルト)"]
    UI["UIヘルパ<br/>WaitStart / WaitRetryOrExit<br/>EdgeButton / GameOverJingle / SoundLabel"]
  end
  Console --> Display
  Console --> Button
  Console --> Buzzer
  Display -. "DrawBuffer" .-> Canvas
  UI --> Button
  UI --> Buzzer
```

## UML: クラス図

主要な型と関係（Go なので「合成」が中心。継承は無い）。

```mermaid
classDiagram
  class Console {
    +Display Display
    +BtnA Button
    +BtnB Button
    +Buzzer Buzzer
  }
  class Display {
    +FillScreen(c)
    +FillRectangle(x,y,w,h,c)
    +DrawBuffer(x,y,w,h,buf)
    +SetPixel(x,y,c)
  }
  class Canvas {
    +Fill(c)
    +FillRectangle(x,y,w,h,c)
    +Buffer() buf
  }
  class Button {
    +Pressed() bool
  }
  class Buzzer {
    +Tone(freq,ms)
    +ToggleMuted() bool
    +Muted() bool
  }
  class IMU {
    +Tilt(th) Dir
    +Acceleration() xyz
  }
  class Game {
    <<func>>
    Run(Console, IMU)
  }
  Console o-- Display
  Console o-- Button
  Console o-- Buzzer
  Game ..> Console : 利用
  Game ..> IMU : 利用
  Display ..> Canvas : 一括転送
```

## UML: シーケンス（起動→メニュー→ゲーム→復帰）

```mermaid
sequenceDiagram
  actor U as プレイヤー
  participant L as ランチャー
  participant G as ゲーム (Run)
  participant D as Display
  U->>L: 起動 (cmd/launcher)
  L->>D: メニュー描画
  U->>L: A タップ
  L->>D: カーソル行を差分更新（ちらつき無し）
  U->>L: A 長押し（決定）
  L->>G: Run(con, imu)
  G->>D: タイトル描画
  U->>G: A タップ（開始）
  loop ゲームループ
    U->>G: 入力（チルト / ボタン）
    G->>D: 部分再描画 または Canvas 一括転送
  end
  G->>D: GAME OVER 表示
  U->>G: A 長押し（メニューへ）
  G-->>L: return
  L->>D: メニュー再描画
```

## パッケージ化の方針（まとめ）

- **薄いヘルパの合成**。各ゲームは自分のループ・描画を保持し、安定した重複だけを `pkg/m5stickc` に集約。
- **cmd レイアウト**で「個別ビルド」と「統合ランチャー」を両立。
- インターフェースは最小限（ランチャー→ゲームの `func(*Console, *IMU)`）。
- 詳しい背景は [概要・設計思想](./) を参照。
