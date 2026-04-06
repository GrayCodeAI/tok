# Release Guide

## Creating a New Release

### Standard Release Process

```bash
# 1. Update CHANGELOG.md
# Add new version entry with changes

# 2. Bump version in code
export VERSION="v0.2.0"

# 3. Commit changes
git add -A
git commit -m "chore: prepare release ${VERSION}"

# 4. Create and push tag
git tag -a $VERSION -m "Release $VERSION"
git push origin main --tags
```

### GitHub Actions

Tagging automatically triggers the release workflow:
- Builds for all platforms
- Creates release on GitHub
- Uploads binaries and checksums

Verify at: https://github.com/GrayCodeAI/tokman/actions

### Manual Release (if needed)

```bash
# Build binaries
make build-all

# Create archives
tar -czf tokman-darwin-amd64.tar.gz tokman-darwin-amd64
tar -czf tokman-darwin-arm64.tar.gz tokman-darwin-arm64
# ... other platforms

# Create checksums
sha256sum tokman-* > checksums.txt

# Create release on GitHub
gh release create $VERSION \
  --title "Tokman $VERSION" \
  --notes-file RELEASE_NOTES.md \
  tokman-* checksums.txt
```

### Homebrew Tap Update

After release, update Homebrew formula:

```bash
# Calculate new SHA256
curl -sL https://github.com/GrayCodeAI/tokman/archive/refs/tags/$VERSION.tar.gz | shasum -a 256

# Update tap repository
cd ~/homebrew-tokman
edit Formula/tokman.rb
  url "https://github.com/GrayCodeAI/tokman/archive/refs/tags/$VERSION.tar.gz"
  sha256 "calculated_sha"

# Test
brew install --build-from-source ./Formula/tokman.rb

# Commit and push
git add Formula/tokman.rb
git commit -m "tokman: update to $VERSION"
git push
```

### Post-Release Checklist

- [ ] Release created on GitHub
- [ ] Binaries downloadable
- [ ] Homebrew formula updated
- [ ] CHANGELOG.md updated
- [ ] Documentation updated
- [ ] Docker image pushed (if applicable)
- [ ] Announce in Discord (if applicable)