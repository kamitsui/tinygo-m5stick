import { defineConfig } from 'vitepress'
import { withMermaid } from 'vitepress-plugin-mermaid'

// GitHub Pages (project site) で公開するため base を /<repo>/ に設定。
// https://kamitsui.github.io/tinygo-m5stick/
// 図は Mermaid（vitepress-plugin-mermaid）で描画する。
export default withMermaid(defineConfig({
  lang: 'ja-JP',
  title: 'tinygo-m5stick',
  description: 'TinyGo を M5StickC Plus2 で動かす — チュートリアルとレトロゲーム',
  base: '/tinygo-m5stick/',
  lastUpdated: true,

  themeConfig: {
    nav: [
      { text: 'ホーム', link: '/' },
      { text: 'TinyGoとは', link: '/tinygo/' },
      { text: 'はじめる', link: '/getting-started/' },
      { text: 'ゲーム設計', link: '/games/' },
    ],

    sidebar: [
      {
        text: 'はじめに',
        items: [{ text: '概要', link: '/' }],
      },
      {
        text: 'TinyGo',
        items: [{ text: 'TinyGoとは', link: '/tinygo/' }],
      },
      {
        text: 'M5StickC Plus2',
        items: [{ text: 'はじめる（環境構築・書き込み）', link: '/getting-started/' }],
      },
      {
        text: 'レトロゲーム設計',
        items: [
          { text: '概要・設計思想', link: '/games/' },
          { text: 'アーキテクチャ（C4 / UML）', link: '/games/architecture' },
          { text: 'ハードウェア / ソフトウェアの仕組み', link: '/games/hardware' },
          { text: '各ゲームの実装', link: '/games/implementations' },
        ],
      },
      {
        text: '技術ノート',
        items: [
          { text: 'ブザー音と非ブロッキング再生', link: '/notes/nonblocking-audio' },
        ],
      },
    ],

    socialLinks: [
      { icon: 'github', link: 'https://github.com/kamitsui/tinygo-m5stick' },
    ],

    search: { provider: 'local' },

    editLink: {
      pattern: 'https://github.com/kamitsui/tinygo-m5stick/edit/main/docs/:path',
      text: 'このページを GitHub で編集',
    },
  },
}))
