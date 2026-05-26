# リポジトリのコンセプト
* M5Stick C Plus 2のデバイスでプログラムを動かす。
* プログラミング言語はTiny Go
* このリポジトリで小さいプロジェクトをいくつか作っていく。
* プロジェクトで Tiny Go と M5 Stick C Plus 2 の開発環境を管理する。（将来的には、他のマイコンの開発環境も対応する予定 Raspberry pi pico, ESP）
* ビルドや書き込み関連はMakefileを活用する。（make help で各ルールの説明を出力する。）

## Git & GitHubの活用
* プロジェクト進行は issues で管理する。
* コミットメッセージは英語で conventional commits に沿ってください。
* `.gitignore` をホワイトリストで作成。`.gitattributes`の作成。 
* 後述のドキュメントをgithub pages で公開する。

## Claude Code を使った開発
* プロジェクトをClaude Code をより効率よく使うための設定ファイルなどを作る。（どんな方法があるかを知りたい）

## ドキュメント
* このリポジトリに関するドキュメントをVitePress を使ってまとめる。
* Tiny Go について調査内容：歴史、活用事例、パソコンでも動かせるか、マイコン（M5 Stick C Plus 2）で動かすために必要な情報（ハードウェアの仕様や公式URLや参考プロジェクトなど）
* 技術スタックや計画については、ユーザーとエージェントとの会話で決めていく。
* インストール方法やサンプルコードなどは、公式ドキュメントを参考にする。具体的な手順などは公式サイトに任せを参照するリンクなどや、初めてTiny Go を触ったりする人が迷子にならないようなドキュメントにしたい。

## 公式ドキュメント
* [macOS インストールガイド](https://tinygo.org/getting-started/install/macos/)
* [Getting Start](https://tinygo.org/getting-started/overview/)
* [m5stack](https://tinygo.org/docs/reference/microcontrollers/machine/m5stack/) : m5stick c plus 2 専用のパッケージではないかもしれないが、参考にしてください。

## プロジェクト１：環境構築とチュートリアル
* 初めに環境構築
* チュートリアル「Hello world」
* ハードウェア連携：ボタン
* ハードウェア連携：ブザーを鳴らす
* ハードウェア連携：表示器を使う
* 必要に応じて、M5Stick C Plus 2 用のオリジナルのパッケージを作る。（今後育てていく。他の機能も使えるようにするために）

## プロジェクト２：レトロゲーム
* [参考リポジトリ](https://github.com/YouMakeTech/PicoRetroGamingSystem) : MicroPython で作られたゲーム
* 作りたいゲームは、SnakeGame, 2048, Pong, Dinosaurus, Tetris, Invadors です。
* 最終的に６つほどのレトロゲームをパッケージ化して使えるようにする。
