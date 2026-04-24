# Packaging

Distribution artifacts for tok, organized by platform.

## Homebrew

Two copies of the formula exist:

- `Formula/tok.rb` — **canonical location**. Consumed by GitHub Releases tooling
  and by a dedicated Homebrew tap (see below). This is the file you edit.
- `packaging/brew/tok.rb` — older location kept as a pointer for anyone who
  finds this path first. Mirrors `Formula/tok.rb`; do not edit independently.

### Publish to a tap

1. Create a tap repo, e.g. `github.com/GrayCodeAI/homebrew-tok`.
2. Copy `Formula/tok.rb` into the tap's `Formula/` directory.
3. Users install with:

   ```sh
   brew tap GrayCodeAI/tok
   brew install tok
   ```

Automation: a future `.github/workflows/publish-tap.yml` will mirror
`Formula/tok.rb` on every release tag.

## Arch Linux (AUR)

`packaging/aur/PKGBUILD` — submit to AUR as `tok` (user-maintained). Run
`updpkgsums` after bumping `pkgver`.

## Docker

`Dockerfile` at repo root. Build + push to a registry of choice. The image is
ENTRYPOINT-scoped to `tok`, so:

```sh
docker run --rm -v "$PWD:/work" -w /work ghcr.io/GrayCodeAI/tok:latest --help
```

## Release signing

`packaging/sign.md` (to be written) will cover GPG key management and checksum
generation for GitHub Releases assets.
