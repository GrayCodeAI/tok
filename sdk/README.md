# TokMan SDKs

Official SDKs for integrating TokMan into your applications.

## Supported Languages

- **Python** (`tokman-py`)
- **Node.js/TypeScript** (`tokman-js`)
- **Go** (`tokman-go`)
- **Rust** (`tokman-rs`)

## Installation

### Python

```bash
pip install tokman-py
```

```python
from tokman import TokmanClient

client = TokmanClient(
    api_key="your-api-key",
    team_id="your-team-id"
)

result = client.analyze("your code here")
print(f"Tokens saved: {result.tokens_saved}")
```

### Node.js/TypeScript

```bash
npm install tokman-js
```

```typescript
import { TokmanClient } from "tokman-js";

const client = new TokmanClient({
  apiKey: "your-api-key",
  teamId: "your-team-id",
});

const result = await client.analyze("your code here");
console.log(`Tokens saved: ${result.tokensSaved}`);
```

### Go

```bash
go get github.com/GrayCodeAI/tokman-go
```

```go
package main

import (
    "fmt"
    "github.com/GrayCodeAI/tokman-go"
)

func main() {
    client := tokman.NewClient(
        tokman.WithAPIKey("your-api-key"),
        tokman.WithTeamID("your-team-id"),
    )
    
    result, err := client.Analyze(context.Background(), "your code here")
    if err != nil {
        panic(err)
    }
    
    fmt.Printf("Tokens saved: %d\n", result.TokensSaved)
}
```

### Rust

```bash
cargo add tokman
```

```rust
use tokman::Client;

#[tokio::main]
async fn main() {
    let client = Client::new(
        "your-api-key",
        "your-team-id",
    ).await.unwrap();
    
    let result = client.analyze("your code here").await.unwrap();
    println!("Tokens saved: {}", result.tokens_saved);
}
```

## Features

All SDKs support:

- **Synchronous & Asynchronous** APIs
- **Streaming** for large files
- **Caching** to reduce API calls
- **Retry Logic** with exponential backoff
- **Rate Limiting** awareness
- **Comprehensive Error Handling**
- **Type Safety** (in typed languages)
- **Full API Coverage** (analytics, config, etc.)

## Usage Examples

### Analyzing Code

```python
# Python example
result = client.analyze(
    code="def hello():\n    print('Hello')",
    language="python",
    compression_level="aggressive"
)

print(f"Original: {result.original_tokens}")
print(f"Compressed: {result.compressed_tokens}")
print(f"Savings: {result.savings_percent}%")
```

### Batch Processing

```python
# Python example
files = ["file1.py", "file2.py", "file3.py"]
results = client.analyze_batch(files)

for file, result in results.items():
    print(f"{file}: {result.savings_percent}% saved")
```

### Analytics Queries

```python
# Python example
stats = client.get_stats(
    start_date="2024-01-01",
    end_date="2024-01-31",
    group_by="day"
)

for date, metrics in stats.items():
    print(f"{date}: {metrics['tokens_saved']} tokens saved")
```

### Team Management

```python
# Python example
team = client.get_team()
users = client.list_users()
analytics = client.get_analytics()

print(f"Team: {team.name}")
print(f"Users: {len(users)}")
print(f"Cost saved: ${analytics.cost_saved}")
```

## Configuration

### Environment Variables

All SDKs support configuration via environment variables:

```bash
export TOKMAN_API_KEY="your-api-key"
export TOKMAN_TEAM_ID="your-team-id"
export TOKMAN_API_ENDPOINT="https://api.tokman.dev"
```

### Configuration File

Create `~/.tokman/config.json`:

```json
{
  "api_key": "your-api-key",
  "team_id": "your-team-id",
  "api_endpoint": "https://api.tokman.dev",
  "compression_level": "aggressive",
  "cache_enabled": true,
  "cache_ttl": 3600
}
```

## Error Handling

All SDKs provide consistent error types:

```python
# Python example
from tokman import RateLimitError, AuthenticationError, ValidationError

try:
    result = client.analyze(code)
except AuthenticationError:
    print("Invalid API key")
except RateLimitError as e:
    print(f"Rate limited, retry after {e.retry_after}s")
except ValidationError as e:
    print(f"Invalid input: {e.message}")
```

## Performance

### Caching

SDKs automatically cache results:

```python
# Second call returns cached result
result1 = client.analyze("same code here")
result2 = client.analyze("same code here")  # Uses cache
```

### Streaming

For large files, use streaming:

```python
# Python example
with open("large_file.py") as f:
    result = client.analyze_stream(f)
```

## Testing

All SDKs include testing utilities:

```python
# Python example
from tokman.testing import MockClient

mock_client = MockClient()
mock_client.set_response(analyze={"tokens_saved": 1000})

result = mock_client.analyze("code")
assert result.tokens_saved == 1000
```

## Versioning

SDKs follow semantic versioning. Check compatibility:

```python
from tokman import __version__
print(f"SDK version: {__version__}")
```

## Support

- 📖 [Documentation](https://docs.tokman.dev)
- 💬 [Discord Community](https://discord.gg/tokman)
- 🐛 [GitHub Issues](https://github.com/GrayCodeAI/tokman/issues)

## License

MIT - See LICENSE for details
