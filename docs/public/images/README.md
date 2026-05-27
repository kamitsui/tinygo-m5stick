# ドキュメント画像の置き場・規約

このディレクトリは VitePress が配信する静的アセット（`docs/public/`）。ここに置いた画像は
**VitePress からは `/images/...`**（base 自動付与）、**README からは相対パス
`docs/public/images/...`** で参照できる（単一の置き場として両方から使う）。

## 命名規約

```
docs/public/images/
├── menu.webp              # トップメニュー（VitePress home の hero 用）
└── games/
    ├── snake.webp         # 静止画（ゲーム名 = パッケージ名）
    ├── snake-demo.webp    # 操作デモ（アニメーション WebP）
    ├── 2048.webp / 2048-demo.webp
    ├── pong.webp / pong-demo.webp
    ├── dinosaur.webp / dinosaur-demo.webp
    ├── tetris.webp / tetris-demo.webp
    └── invaders.webp / invaders-demo.webp
```

- 形式は **WebP**（静止）/ **アニメーション WebP**（操作デモ）を主役にする。README・VitePress
  の両方でリポジトリ内ファイルとして軽量に表示できる唯一の形式。
- 1ファイルは数百KB以内を目安（合計が数MBを超えそうなら Git LFS を検討）。

## 変換（ffmpeg）

```bash
# 録画(mp4) → アニメーション WebP（3〜5秒・低fps・幅270px=実機2倍が目安）
ffmpeg -i in.mp4 -vf "fps=12,scale=270:-1:flags=lanczos" \
  -vcodec libwebp -lossless 0 -q:v 55 -loop 0 -an games/snake-demo.webp

# 静止画(png) → WebP
ffmpeg -i in.png -vcodec libwebp -lossless 0 -q:v 80 games/snake.webp
```

## 参照例

- VitePress home（hero）: `hero.image: { src: '/images/menu.webp' }`
- VitePress 本文: `![Snake](/images/games/snake-demo.webp)`
- README: `<img src="docs/public/images/games/snake-demo.webp" width="120">`
