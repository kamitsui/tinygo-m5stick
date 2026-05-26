import { defineConfig } from 'vitepress'

// GitHub Pages (project site) で公開するため base を /<repo>/ に設定。
// https://kamitsui.github.io/tinygo-m5stick/
export default defineConfig({
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
})
