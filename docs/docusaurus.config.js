// @ts-check
// Docusaurus configuration for the glixos documentation site.
// Deployed to https://powerreddude.github.io/glixos/ via GitHub Pages.

import { themes as prismThemes } from 'prism-react-renderer';

/** @type {import('@docusaurus/types').Config} */
const config = {
  title: 'glixos',
  tagline: 'Modular, flake-driven NixOS with a manifest you can read.',
  favicon: 'img/favicon.svg',

  url: 'https://powerreddude.github.io',
  baseUrl: '/glixos/',

  organizationName: 'powerreddude',
  projectName: 'glixos',
  deploymentBranch: 'gh-pages',
  trailingSlash: false,

  onBrokenLinks: 'warn',
  onBrokenMarkdownLinks: 'warn',

  i18n: {
    defaultLocale: 'en',
    locales: ['en'],
  },

  presets: [
    [
      'classic',
      /** @type {import('@docusaurus/preset-classic').Options} */
      ({
        docs: {
          path: 'content',
          routeBasePath: '/',
          sidebarPath: './sidebars.js',
          editUrl:
            'https://github.com/powerreddude/glixos/edit/main/docs/',
        },
        blog: false,
        theme: {
          customCss: './src/css/custom.css',
        },
      }),
    ],
  ],

  themeConfig:
    /** @type {import('@docusaurus/preset-classic').ThemeConfig} */
    ({
      colorMode: {
        defaultMode: 'dark',
        respectPrefersColorScheme: true,
      },
      navbar: {
        title: 'glixos',
        logo: {
          alt: 'glixos',
          src: 'img/logo.svg',
        },
        items: [
          {
            type: 'docSidebar',
            sidebarId: 'userSidebar',
            position: 'left',
            label: 'Users',
          },
          {
            type: 'docSidebar',
            sidebarId: 'contributorSidebar',
            position: 'left',
            label: 'Contributors',
          },
          {
            type: 'docSidebar',
            sidebarId: 'adrSidebar',
            position: 'left',
            label: 'ADRs',
          },
          {
            href: 'https://github.com/powerreddude/glixos',
            label: 'GitHub',
            position: 'right',
          },
        ],
      },
      footer: {
        style: 'dark',
        links: [
          {
            title: 'Docs',
            items: [
              { label: 'Users', to: '/user/getting-started' },
              { label: 'Contributors', to: '/contributor/overview' },
              { label: 'ADRs', to: '/adr/' },
            ],
          },
          {
            title: 'Project',
            items: [
              {
                label: 'GitHub',
                href: 'https://github.com/powerreddude/glixos',
              },
              {
                label: 'Issues',
                href: 'https://github.com/powerreddude/glixos/issues',
              },
            ],
          },
        ],
        copyright: `glixos is licensed under the GNU General Public License v3.0.`,
      },
      prism: {
        theme: prismThemes.github,
        darkTheme: prismThemes.dracula,
        additionalLanguages: ['nix', 'toml', 'bash', 'go', 'json'],
      },
    }),
};

export default config;
