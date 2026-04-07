# GitHub Action Marketplace Submission Guide

## Pre-Submission Checklist

### Action Metadata (action.yml)
```yaml
name: 'TokMan - Token Reduction for GitHub Actions'
description: 'Analyze token usage and reduce costs in your CI/CD workflows'
author: 'Gray Code AI'

inputs:
  api_endpoint:
    description: 'TokMan API endpoint'
    required: false
    default: 'https://api.tokman.dev'
  api_token:
    description: 'TokMan API token (use ${{ secrets.TOKMAN_API_TOKEN }})'
    required: true
  team_id:
    description: 'Your TokMan team ID'
    required: true
  analyze_logs:
    description: 'Analyze CI logs for token reduction'
    required: false
    default: 'true'
  token_budget:
    description: 'Monthly token budget for your team'
    required: false
    default: '0'
  fail_on_budget_exceeded:
    description: 'Fail workflow if budget exceeded'
    required: false
    default: 'false'
  comment_on_pr:
    description: 'Post analysis results as PR comment'
    required: false
    default: 'true'
  compression_level:
    description: 'Compression level: minimal, moderate, aggressive'
    required: false
    default: 'aggressive'

outputs:
  tokens_analyzed:
    description: 'Total tokens analyzed'
  tokens_saved:
    description: 'Tokens saved through compression'
  cost_before:
    description: 'Estimated cost before compression'
  cost_after:
    description: 'Estimated cost after compression'
  savings_percent:
    description: 'Percentage savings'

runs:
  using: 'docker'
  image: 'docker://graycodeai/tokman-action:latest'
```

### Branding
```yaml
branding:
  icon: 'zap'           # or 'activity', 'alert', 'arrow-right', etc.
  color: 'blue'         # Primary brand color
```

### Code Quality
- [x] Action follows GitHub Actions best practices
- [x] Proper error handling for all edge cases
- [x] Uses official GitHub actions (checkout, upload-artifact, comment-pr)
- [x] No hardcoded secrets or credentials
- [x] Logs are informative without being verbose
- [x] All inputs validated
- [x] Outputs clearly documented

## Repository Setup

### Branch Protection Rules
- Require reviews before merging to `main`
- Require status checks to pass
- Require branches to be up to date

### Topics
```yaml
topics:
  - github-actions
  - tokman
  - cost-reduction
  - ci-cd
  - optimization
```

### Description
```
TokMan GitHub Action: Analyze and reduce token usage in your CI/CD workflows.
Save up to 90% on AI API costs with intelligent compression.
```

## README.md for Marketplace

```markdown
# TokMan GitHub Action

Analyze token usage and reduce costs in your GitHub Actions workflows with intelligent compression.

## Features

✨ **CI/CD Integration** - Seamlessly integrates with your GitHub Actions workflows
📊 **Detailed Analysis** - See exactly where tokens are being used
💰 **Cost Tracking** - Monitor your API spend
🚫 **Budget Enforcement** - Fail builds when exceeding team budgets
🤖 **Smart Compression** - Up to 90% token reduction
📈 **PR Comments** - Share results directly in pull requests

## Usage

### Basic Setup

```yaml
name: Analyze with TokMan
on: [pull_request]

jobs:
  analyze:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      
      - name: Analyze with TokMan
        uses: GrayCodeAI/tokman-action@v1
        with:
          api_token: ${{ secrets.TOKMAN_API_TOKEN }}
          team_id: ${{ secrets.TOKMAN_TEAM_ID }}
          analyze_logs: true
          compression_level: aggressive
```

### With Budget Enforcement

```yaml
- name: TokMan Cost Analysis
  uses: GrayCodeAI/tokman-action@v1
  with:
    api_token: ${{ secrets.TOKMAN_API_TOKEN }}
    team_id: ${{ secrets.TOKMAN_TEAM_ID }}
    token_budget: 5000000  # 5M tokens/month
    fail_on_budget_exceeded: true
    comment_on_pr: true
```

### Advanced: Custom API Endpoint

```yaml
- name: TokMan Analysis (Self-Hosted)
  uses: GrayCodeAI/tokman-action@v1
  with:
    api_endpoint: https://tokman.internal.company.com
    api_token: ${{ secrets.TOKMAN_API_TOKEN }}
    team_id: ${{ secrets.TOKMAN_TEAM_ID }}
    compression_level: moderate
```

## Setup Instructions

### 1. Get Your API Token

1. Sign up at [tokman.dev](https://tokman.dev)
2. Get your API token from [tokman.dev/dashboard](https://tokman.dev/dashboard)
3. Get your Team ID from Dashboard → Settings

### 2. Add Secrets to GitHub

Go to your repository:
- Settings → Secrets and variables → Actions
- Add `TOKMAN_API_TOKEN` with your token
- Add `TOKMAN_TEAM_ID` with your team ID (optional, can be hardcoded)

### 3. Add to Your Workflow

See examples above.

## Outputs

The action provides these outputs for use in subsequent steps:

```yaml
- name: Use TokMan outputs
  run: |
    echo "Analyzed ${{ steps.tokman.outputs.tokens_analyzed }} tokens"
    echo "Saved ${{ steps.tokman.outputs.tokens_saved }} tokens"
    echo "Cost before: ${{ steps.tokman.outputs.cost_before }}"
    echo "Cost after: ${{ steps.tokman.outputs.cost_after }}"
    echo "Savings: ${{ steps.tokman.outputs.savings_percent }}%"
```

## PR Comments

By default, the action posts a comment on PRs with results:

```
## 📊 TokMan Analysis Results

**Tokens Analyzed**: 12,500
**Tokens Saved**: 10,250 (82%)

**Estimated Cost** (Claude API):
- Before: $0.625 (prompt) + $0.250 (completion) = **$0.875**
- After: $0.113 (prompt) + $0.045 (completion) = **$0.158**
- **Savings: $0.717 (82%)**

**Team Budget Usage**:
- Monthly Quota: 50,000,000 tokens
- Used This Month: 12,500,000 (25%)
- Remaining: 37,500,000

✅ Budget OK - Continue deployment
```

To disable, set `comment_on_pr: false`.

## Compression Levels

- **minimal** - Only remove comments and whitespace
- **moderate** - Remove comments, whitespace, and basic redundancy
- **aggressive** - Maximum compression, may reduce readability (default)

## Pricing

- **Free**: 100 analyses/month, 1M tokens/month
- **Pro**: Unlimited analyses, 50M tokens/month ($99/month)
- **Enterprise**: Custom limits and pricing

## Troubleshooting

### Error: "Invalid API token"
- Verify `TOKMAN_API_TOKEN` is set correctly
- Check token hasn't expired at [tokman.dev/dashboard](https://tokman.dev/dashboard)

### Error: "Team ID not found"
- Verify `TOKMAN_TEAM_ID` matches your dashboard
- Go to Dashboard → Settings to find your ID

### Budget exceeded but workflow didn't fail
- Ensure `fail_on_budget_exceeded: true` is set
- Check that `token_budget` is configured correctly

### No PR comment appearing
- Ensure `comment_on_pr: true` (default)
- Check that action has `pull-requests: write` permissions

## Permissions

The action requires these permissions:

```yaml
permissions:
  pull-requests: write  # For PR comments
  contents: read        # For reading repository contents
```

## Examples

### Full CI/CD Pipeline

```yaml
name: Build & Analyze

on:
  pull_request:
  push:
    branches: [main]

jobs:
  build:
    runs-on: ubuntu-latest
    permissions:
      pull-requests: write
      contents: read
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Build project
        run: npm run build
      
      - name: Run tests
        run: npm test
      
      - name: Analyze with TokMan
        id: tokman
        uses: GrayCodeAI/tokman-action@v1
        with:
          api_token: ${{ secrets.TOKMAN_API_TOKEN }}
          team_id: ${{ secrets.TOKMAN_TEAM_ID }}
          analyze_logs: true
          token_budget: 10000000
          fail_on_budget_exceeded: false
      
      - name: Report savings
        run: |
          echo "## 💰 Token Savings Summary" >> $GITHUB_STEP_SUMMARY
          echo "Tokens Saved: ${{ steps.tokman.outputs.tokens_saved }}" >> $GITHUB_STEP_SUMMARY
          echo "Savings: ${{ steps.tokman.outputs.savings_percent }}%" >> $GITHUB_STEP_SUMMARY
```

### Matrix Builds with TokMan

```yaml
jobs:
  test:
    runs-on: ubuntu-latest
    strategy:
      matrix:
        python-version: ['3.8', '3.9', '3.10', '3.11']
    
    steps:
      - uses: actions/checkout@v4
      
      - name: Set up Python
        uses: actions/setup-python@v4
        with:
          python-version: ${{ matrix.python-version }}
      
      - name: Analyze with TokMan
        uses: GrayCodeAI/tokman-action@v1
        with:
          api_token: ${{ secrets.TOKMAN_API_TOKEN }}
          team_id: ${{ secrets.TOKMAN_TEAM_ID }}
          compression_level: moderate
```

## API

For programmatic access, use the TokMan API:

```bash
curl -X POST https://api.tokman.dev/analyze \
  -H "Authorization: Bearer $TOKMAN_API_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"code":"...","language":"python"}'
```

See [API Documentation](https://tokman.dev/docs/api) for details.

## Support

- 📧 [support@tokman.dev](mailto:support@tokman.dev)
- 💬 [Discord Community](https://discord.gg/tokman)
- 📖 [Documentation](https://tokman.dev/docs)
- 🐛 [GitHub Issues](https://github.com/GrayCodeAI/tokman/issues)

## License

MIT - See [LICENSE](LICENSE) for details

---

**Reduce costs. Optimize workflows. Deploy faster.**
```

## Submission Steps

### 1. Verify Marketplace Eligibility

Your action must meet these requirements:
- [x] Published repository with MIT/Apache/GPL license
- [x] Minimum 5 GitHub stars (recommendation)
- [x] Complete README with examples
- [x] action.yml with proper schema
- [x] Branding (icon and color)
- [x] No malware or deceptive content

### 2. Create Release

```bash
# Tag version
git tag v1.0.0
git push origin v1.0.0

# Or release via GitHub UI:
# Releases → New Release → v1.0.0
```

### 3. Submit to Marketplace

Go to your repository page:
- Click "Packages" in sidebar
- Select "Publish this action to GitHub Marketplace"
- Verify metadata
- Complete submission

### 4. Verify Publication

- Visit: https://github.com/marketplace/actions/tokman-token-reduction-for-github-actions
- Verify:
  - Description correct
  - Icon displays
  - README renders properly
  - Code example works
  - Links functional

## Post-Submission

### Monitor Usage
- Check Marketplace metrics weekly
- Monitor GitHub discussions/issues
- Track stars and usage statistics
- Respond to issues within 24 hours

### Versioning Strategy

Use semantic versioning:
- v1.0.0 - Initial release
- v1.0.1 - Patch (bug fixes)
- v1.1.0 - Minor (new features, backward compatible)
- v2.0.0 - Major (breaking changes)

### Major/Minor Release Template

```yaml
# Version bump
npm version minor  # or patch, major

# Tag
git tag -a v1.1.0 -m "Add feature X"
git push origin v1.1.0

# Create release
gh release create v1.1.0 \
  --title "v1.1.0: Feature X" \
  --notes "## New Features\n- Feature X\n\n## Bug Fixes\n- Fix Y"
```

## Success Metrics

Target for Month 1:
- 100+ action runs
- 50+ GitHub stars
- 4.5+ rating
- 0 critical issues

Target for Month 3:
- 5,000+ action runs/month
- 500+ GitHub stars
- 4.7+ rating
- Active community on Discussions

---

**Ready to launch! Get TokMan Action into thousands of CI/CD pipelines.**
