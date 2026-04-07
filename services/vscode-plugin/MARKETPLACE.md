# VSCode Extension Marketplace Submission Guide

## Pre-Submission Checklist

### Code Quality
- [x] Extension follows VSCode guidelines
- [x] No console warnings or errors
- [x] Proper error handling throughout
- [x] TypeScript strict mode enabled
- [x] No hardcoded secrets or credentials
- [x] All dependencies pinned to specific versions

### Extension Manifest (package.json)
```json
{
  "name": "tokman",
  "displayName": "TokMan - Token Reduction for Claude",
  "description": "Reduce token usage in Claude AI prompts with intelligent code compression",
  "version": "1.0.0",
  "publisher": "GrayCodeAI",
  "license": "MIT",
  "engines": {
    "vscode": "^1.75.0"
  },
  "categories": [
    "AI",
    "Formatters",
    "Other"
  ],
  "keywords": [
    "ai",
    "claude",
    "compression",
    "tokens",
    "cost-reduction",
    "productivity"
  ],
  "icon": "media/icon.png",
  "repository": {
    "type": "git",
    "url": "https://github.com/GrayCodeAI/tokman"
  },
  "bugs": {
    "url": "https://github.com/GrayCodeAI/tokman/issues"
  },
  "homepage": "https://tokman.dev"
}
```

### Visual Assets Required

#### Icon (128x128 PNG)
- Simple, recognizable design
- Works at small sizes
- Clear in both light and dark themes
- Filename: `media/icon.png`

#### Banner Image (200x120 PNG, Optional but recommended)
- Shows extension in action
- Professional quality
- Clear typography
- Filename: `media/banner.png`

#### Screenshots (For Marketplace Page)
1. **Screenshot 1: Token Counter Hover**
   - Show hover preview with token count
   - File: `media/screenshot-1-hover.png`

2. **Screenshot 2: Inline Decorations**
   - Show inline decorations on code
   - File: `media/screenshot-2-decorations.png`

3. **Screenshot 3: Status Bar**
   - Show status bar with savings percentage
   - File: `media/screenshot-3-statusbar.png`

4. **Screenshot 4: Send to Claude Button**
   - Show integration with Claude
   - File: `media/screenshot-4-send.png`

### README.md for Marketplace

```markdown
# TokMan - Token Reduction for Claude

Reduce token usage and cut costs by up to 90% when using Claude AI as your coding assistant.

## Features

✨ **Real-Time Token Preview** - Hover over code to see token count and compression savings
🚀 **Inline Decorations** - See compression ratio directly in your editor
📊 **Dashboard Integration** - Track your savings across all projects
🤖 **Adaptive Learning** - Improves compression based on your codebase
💰 **Cost Tracking** - Monitor your API cost savings
🔌 **Claude Integration** - Send compressed code directly to Claude

## Quick Start

1. Install TokMan from the VSCode Marketplace
2. Get your API key from [tokman.dev/dashboard](https://tokman.dev/dashboard)
3. Set your API key in TokMan settings
4. Hover over code to see token savings
5. Use `Ctrl+Shift+T` to analyze any file

## Usage

### Analyze Code
- Select code → `Ctrl+Shift+T` (or right-click → "Analyze with TokMan")
- See immediate token compression results

### Toggle Preview
- `Ctrl+Shift+P` to toggle inline decorations
- View compression ratio on each line

### Send to Claude
- Click the "📤 Send to Claude" button in the status bar
- Compressed code automatically copied to clipboard

### View Dashboard
- Click the "📊 Dashboard" button
- See your savings across all files and projects

## Configuration

Open VSCode settings and search for "tokman":

- **API Endpoint** - Custom API endpoint (default: https://api.tokman.dev)
- **Compression Level** - "minimal", "moderate", "aggressive" (default: "aggressive")
- **Token Counter Model** - "claude-3-opus", "claude-3-sonnet", "claude-3-haiku"
- **Cache TTL** - Minutes to cache results (default: 5)
- **Show Inline Decorations** - Display savings directly in editor
- **Enable Status Bar Updates** - Show metrics in status bar

## Pricing

- **Free Tier**: 100 analyses/day, 1M tokens/month
- **Pro Tier**: 10,000 analyses/day, 50M tokens/month ($99/month)
- **Enterprise**: Unlimited usage, custom limits

[View full pricing](https://tokman.dev/pricing)

## Support

- 📧 [support@tokman.dev](mailto:support@tokman.dev)
- 💬 [Discord Community](https://discord.gg/tokman)
- 📖 [Documentation](https://tokman.dev/docs)
- 🐛 [GitHub Issues](https://github.com/GrayCodeAI/tokman/issues)

## Privacy & Security

- Your code is sent securely to TokMan API via HTTPS
- Enterprise deployments available for on-premise use
- GDPR and HIPAA compliance ready
- See [Privacy Policy](https://tokman.dev/privacy)

## Changelog

### Version 1.0.0
- Initial release
- Real-time token preview
- Inline decorations
- Claude integration
- Dashboard sync
- 3 compression levels

## License

MIT - See [LICENSE](LICENSE) for details

---

**Reduce tokens. Cut costs. Code faster.**
```

### Change Log

```markdown
# Changelog

## [1.0.0] - 2026-04-07

### Added
- Initial VSCode extension release
- Real-time token preview on hover
- Inline compression ratio decorations
- Integration with Claude for sending compressed code
- Dashboard sync for tracking savings
- Three compression levels (minimal, moderate, aggressive)
- Status bar with current metrics
- Cache for improved performance
- Support for 10+ programming languages
- Configuration settings for customization
- Error handling and recovery
- Telemetry for usage tracking

### Security
- HTTPS communication with API
- No sensitive data stored locally
- Secure token storage in VSCode keychain

## [0.9.0] - Pre-release
- Beta testing version
```

## Submission Steps

### 1. Create Microsoft Account
- Go to https://dev.azure.com
- Sign up for free Azure DevOps account
- Create Personal Access Token (PAT)
  - Scope: `Marketplace (manage)`
  - Expiration: 1 year

### 2. Install vsce (Visual Studio Code Extension CLI)
```bash
npm install -g @vscode/vsce
```

### 3. Package Extension
```bash
cd services/vscode-plugin
vsce package
# Creates: tokman-1.0.0.vsix
```

### 4. Verify Package
```bash
# List contents
unzip -l tokman-1.0.0.vsix | head -20

# Should include:
# - extension.js/bundle.js
# - package.json
# - README.md
# - LICENSE
# - icon.png
# - media/screenshots
```

### 5. Test Installation Locally
```bash
code --install-extension tokman-1.0.0.vsix
```

### 6. Create Publisher (One-time)
```bash
vsce create-publisher GrayCodeAI
# Email: dev@tokman.dev
# Name: Gray Code AI
```

### 7. Login to vsce
```bash
vsce login GrayCodeAI
# Paste PAT when prompted
```

### 8. Publish to Marketplace
```bash
vsce publish
# Or: vsce publish --packagePath ./tokman-1.0.0.vsix
```

### 9. Verify on Marketplace
- Visit: https://marketplace.visualstudio.com/items?itemName=GrayCodeAI.tokman
- Verify:
  - Title and description correct
  - Icon displays properly
  - Screenshots show
  - Repository link works
  - Documentation links work

## Post-Submission

### Monitor
- Check Marketplace metrics weekly
- Respond to reviews and feedback within 24 hours
- Track download statistics
- Monitor GitHub issues from extension users

### Updates
- For patches: `vsce publish patch`
- For minor versions: `vsce publish minor`
- For major versions: `vsce publish major`

### Promote
- Post in VS Code social channels
- Share with developer communities
- Highlight in TokMan blog
- Add showcase to tokman.dev

## Troubleshooting

### Extension not showing in Marketplace
- Verify publisher account
- Check extension.json validity
- Ensure all required files present
- Look for validation warnings in vsce output

### Download numbers low
- Check SEO keywords
- Improve extension description
- Add more screenshots
- Create demo video

### Users reporting issues
- Update changelog quickly
- Publish patches rapidly (within hours for critical issues)
- Respond to all reviews
- Link to documentation for common questions

## Success Metrics

Target for Month 1:
- 1,000+ downloads
- 4.5+ star rating
- 100+ active users
- < 1% uninstall rate

Target for Month 3:
- 10,000+ downloads
- 4.7+ star rating
- 1,000+ active users
- < 2% uninstall rate

---

**Ready to launch! Follow these steps to get TokMan into 500K+ VSCode users.**
