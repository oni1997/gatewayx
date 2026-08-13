import { defineConfig } from 'vitepress'

export default defineConfig({
  title: 'GatewayX',
  description: 'Developer Infrastructure Platform',
  lang: 'en-US',
  cleanUrls: true,
  themeConfig: {
    logo: '/logo.png',
    nav: [
      { text: 'Guide', link: '/guide/getting-started' },
      { text: 'GitHub', link: 'https://github.com/oni1997/gatewayx' },
    ],
    sidebar: {
      '/guide/': [
        {
          text: 'Getting Started',
          items: [
            { text: 'Introduction', link: '/guide/introduction' },
            { text: 'Quick Start', link: '/guide/getting-started' },
            { text: 'Installation', link: '/guide/installation' },
          ],
        },
        {
          text: 'Configuration',
          items: [
            { text: 'Configuration', link: '/guide/configuration' },
            { text: 'Routing', link: '/guide/routing' },
            { text: 'Authentication', link: '/guide/authentication' },
            { text: 'Rate Limiting', link: '/guide/rate-limiting' },
            { text: 'Caching', link: '/guide/caching' },
          ],
        },
        {
          text: 'Operations',
          items: [
            { text: 'Dashboard', link: '/guide/dashboard' },
            { text: 'Metrics', link: '/guide/metrics' },
            { text: 'Security', link: '/guide/security' },
            { text: 'Plugins', link: '/guide/plugins' },
            { text: 'Deployment', link: '/guide/deployment' },
          ],
        },
      ],
    },
    socialLinks: [
      { icon: 'github', link: 'https://github.com/oni1997/gatewayx' },
    ],
    footer: {
      message: 'Released under the MIT License.',
      copyright: 'Copyright © 2026 Onesmus Maenzanise',
    },
  },
})
