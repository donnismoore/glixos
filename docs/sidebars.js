// @ts-check
// Two doc paths share the same content tree, separated by top-level
// directory: content/user/** and content/contributor/**. ADRs get their
// own sidebar so they can be browsed independently.

/** @type {import('@docusaurus/plugin-content-docs').SidebarsConfig} */
const sidebars = {
  userSidebar: [
    'intro',
    {
      type: 'category',
      label: 'Getting started',
      collapsed: false,
      items: [
        'user/getting-started',
        'user/installation',
        'user/quickstart',
      ],
    },
    {
      type: 'category',
      label: 'Daily use',
      items: [
        'user/commands',
        'user/manifest',
        'user/packages',
        'user/config',
        'user/registry',
      ],
    },
    {
      type: 'category',
      label: 'Beyond one machine',
      items: [
        'user/multi-host',
        'user/multi-user',
        'user/pinning',
      ],
    },
    {
      type: 'category',
      label: 'Operations',
      items: [
        'user/rollback',
        'user/gc',
        'user/troubleshooting',
      ],
    },
  ],

  contributorSidebar: [
    'contributor/overview',
    'contributor/design-principles',
    'contributor/architecture',
    {
      type: 'category',
      label: 'Codebase',
      items: [
        'contributor/codebase-tour',
        'contributor/cli-internals',
        'contributor/importManifest',
        'contributor/patcher',
        'contributor/resolver',
        'contributor/snapshots',
        'contributor/flake-contract',
      ],
    },
    {
      type: 'category',
      label: 'Project',
      items: [
        'contributor/build-test',
        'contributor/milestones',
        'contributor/release',
      ],
    },
  ],

  adrSidebar: [
    'adr/index',
    'adr/ADR-001-two-flakes',
    'adr/ADR-002-flake-contract',
    'adr/ADR-003-manifest-driven',
    'adr/ADR-004-resolver-chain',
    'adr/ADR-005-multi-host',
    'adr/ADR-006-secrets-out-of-scope',
    'adr/ADR-007-flakes-only',
    'adr/ADR-008-rollback-coupling',
    'adr/ADR-009-per-package-config',
    'adr/ADR-010-multi-user-routing',
  ],
};

export default sidebars;
