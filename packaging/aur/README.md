# AUR package

`PKGBUILD` for [laucha-bin](https://aur.archlinux.org/packages/laucha-bin),
which installs the released binary rather than building from source.

The AUR keeps its own git repository, so this copy is the source of truth and
gets pushed there on every release:

```sh
git clone ssh://aur@aur.archlinux.org/laucha-bin.git
cp packaging/aur/PKGBUILD laucha-bin/
cd laucha-bin
updpkgsums          # refresh the checksums for the new release
makepkg --printsrcinfo > .SRCINFO
namcap PKGBUILD     # optional lint
git commit -am "laucha 1.2.3" && git push
```

Bump `pkgver` (and reset `pkgrel` to 1) for a new laucha release; bump only
`pkgrel` when the packaging changes but the version does not.
