# Security policy

## Supported versions

Fixes land in the latest release. Please reproduce an issue on the newest
version before reporting it.

## Reporting a vulnerability

Report privately through GitHub: open the
[Security tab](https://github.com/lucas77x/laucha/security/advisories/new) and
file a draft advisory. Please do not open a public issue for a vulnerability.

Useful details: what an attacker can reach, the steps to reproduce it, the
laucha version, and your distribution and desktop session.

## What laucha touches

- It indexes the file names under the folders you configure and stores them,
  along with how often you open things, under `~/.local/share/laucha` with
  owner-only permissions.
- It launches applications and files as your user: desktop entries are
  executed as argv, never through a shell, and files are opened with
  `xdg-open`.
- It only talks to the network when you press "Check for updates", which asks
  the GitHub releases API for the latest tag.
