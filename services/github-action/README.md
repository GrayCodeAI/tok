# TokMan GitHub Action

Automatically analyze token usage and add cost analytics to your pull requests.

## Features

- 📊 **Analyze test logs, build logs, and CI output** for token usage
- 💰 **Cost analysis** with savings breakdown by model
- 📈 **PR comments** with token metrics and cost comparison
- ⚠️ **Budget alerts** when approaching monthly limits
- 🎯 **Compression preview** showing potential savings
- 🔒 **Privacy-first** - all processing is local by default

## Usage

### Basic Setup

```yaml
name: TokMan Analysis
on: [pull_request]

jobs:
  tokman:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Analyze with TokMan
        uses: GrayCodeAI/tokman@v1
        with:
          api_endpoint: 'http://api.tokman.dev'
          api_token: ${{ secrets.TOKMAN_API_TOKEN }}
          team_id: ${{ secrets.TOKMAN_TEAM_ID }}
          analyze_logs: 'true'
          comment_on_pr: 'true'
```

### With Budget Enforcement

```yaml
- name: Analyze and Check Budget
  uses: GrayCodeAI/tokman@v1
  with:
    api_endpoint: 'http://api.tokman.dev'
    api_token: ${{ secrets.TOKMAN_API_TOKEN }}
    team_id: ${{ secrets.TOKMAN_TEAM_ID }}
    token_budget: '10000000'  # 10M tokens/month
    fail_on_budget_exceeded: 'true'
    compression_level: 'aggressive'
```

### Advanced: Analyze Specific Files

```yaml
- name: Run Tests
  run: npm test > test-output.log 2>&1

- name: Analyze Test Output
  uses: GrayCodeAI/tokman@v1
  with:
    api_endpoint: 'http://api.tokman.dev'
    api_token: ${{ secrets.TOKMAN_API_TOKEN }}
    team_id: ${{ secrets.TOKMAN_TEAM_ID }}
    analyze_logs: 'true'
```

## Inputs

| Input | Description | Required | Default |
|-------|-------------|----------|---------|
| `api_endpoint` | TokMan API endpoint URL | No | `http://localhost:8083` |
| `api_token` | API authentication token | No | (uses TOKMAN_API_TOKEN env) |
| `team_id` | Team ID for analytics | No | (uses TOKMAN_TEAM_ID env) |
| `analyze_logs` | Analyze test/build logs | No | `true` |
| `token_budget` | Monthly token budget | No | `10000000` |
| `fail_on_budget_exceeded` | Fail workflow if budget exceeded | No | `false` |
| `comment_on_pr` | Add PR comment with results | No | `true` |
| `compression_level` | Compression level (minimal, aggressive) | No | `aggressive` |

## Outputs

| Output | Description |
|--------|-------------|
| `tokens_analyzed` | Total tokens analyzed |
| `tokens_saved` | Total tokens saved |
| `cost_before` | Estimated cost before compression |
| `cost_after` | Estimated cost after compression |
| `savings_percent` | Percentage of tokens saved |

## PR Comment Example

The action adds a comment like this to your PRs:

```
## 📊 TokMan Analytics

### Token Usage
- **Analyzed**: 45,123 tokens
- **Compressed**: 13,537 tokens
- **Saved**: 31,586 tokens (70%)

### Cost Analysis
- **Before**: $0.135
- **After**: $0.041
- **Savings**: $0.094 (70%)

### Top Logs
1. **test-output.log** - 25,000 tokens (12,500 saved)
2. **build-log.txt** - 15,000 tokens (10,500 saved)
3. **coverage-report.html** - 5,123 tokens (8,586 saved)

### Team Budget
- **Used**: 245,000 / 10,000,000 tokens (2.45%)
- **Remaining**: 9,755,000 tokens
```

## Environment Variables

You can also configure via environment variables:

```yaml
env:
  TOKMAN_API_ENDPOINT: 'http://api.tokman.dev'
  TOKMAN_API_TOKEN: ${{ secrets.TOKMAN_API_TOKEN }}
  TOKMAN_TEAM_ID: ${{ secrets.TOKMAN_TEAM_ID }}
```

## Budget Alerts

When approaching your monthly budget:

- **80% of budget**: ⚠️ Warning comment on PR
- **95% of budget**: 🔴 Critical alert on PR
- **100% of budget**: ❌ Fail workflow (if `fail_on_budget_exceeded: 'true'`)

## Permissions

The action needs the following permissions:

```yaml
permissions:
  pull-requests: write  # To comment on PRs
  checks: read          # To read workflow logs
```

## Examples

### Example 1: Basic Setup

```yaml
name: TokMan
on: [pull_request, push]

jobs:
  tokman:
    runs-on: ubuntu-latest
    permissions:
      pull-requests: write
    steps:
      - uses: actions/checkout@v4
      - uses: GrayCodeAI/tokman@v1
        with:
          api_token: ${{ secrets.TOKMAN_TOKEN }}
          team_id: ${{ secrets.TOKMAN_TEAM_ID }}
```

### Example 2: With Budget Check

```yaml
name: TokMan with Budget
on: [pull_request]

jobs:
  tokman:
    runs-on: ubuntu-latest
    permissions:
      pull-requests: write
    steps:
      - uses: actions/checkout@v4
      - run: npm test > test.log 2>&1
      - uses: GrayCodeAI/tokman@v1
        with:
          api_token: ${{ secrets.TOKMAN_TOKEN }}
          team_id: ${{ secrets.TOKMAN_TEAM_ID }}
          token_budget: '5000000'
          fail_on_budget_exceeded: 'true'
```

## Troubleshooting

### API Connection Issues

Check that:
1. `api_endpoint` is correct and accessible
2. `api_token` is valid
3. `team_id` exists in your TokMan account

### No PR Comment Appears

Ensure the workflow has `pull-requests: write` permission:

```yaml
permissions:
  pull-requests: write
```

### Budget Exceeded

Review your monthly token usage in the TokMan dashboard and adjust the budget or upgrade your plan.

## Support

- 📖 [Documentation](https://tokman.dev/docs)
- 💬 [Discord Community](https://discord.gg/tokman)
- 🐛 [Issue Tracker](https://github.com/GrayCodeAI/tokman/issues)

## License

MIT - See LICENSE for details
