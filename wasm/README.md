# TokMan WASM

Browser-based token compression with 14-layer pipeline.

## Build

```bash
# Build WASM module
GOOS=js GOARCH=wasm go build -o tokman.wasm ./main.go

# Or use the build script
./build.sh
```

## Usage

### Browser

```html
<script src="wasm_exec.js"></script>
<script src="tokman.js"></script>
<script>
async function main() {
    await TokManWASM.init('./tokman.wasm');
    
    const result = TokManWASM.process('Long text to compress...', {
        mode: 'aggressive',
        budget: 1000
    });
    
    console.log(result);
    // {
    //   output: "Compressed text",
    //   originalTokens: 500,
    //   finalTokens: 50,
    //   tokensSaved: 450,
    //   reductionPercent: 90
    // }
}

main();
</script>
```

### Streaming

```javascript
const stream = TokManWASM.stream((chunk) => {
    console.log('Compressed:', chunk.content);
    console.log('Saved:', chunk.tokensSaved, 'tokens');
}, { mode: 'minimal' });

stream.send('First chunk of content...');
stream.send('Second chunk...');
stream.close();
```

### Content Analysis

```javascript
const analysis = TokManWASM.analyze('func main() { ... }');
// { contentType: 'code' }
```

## API

### `init(wasmPath)`

Initialize the WASM module. Must be called before other functions.

### `process(input, options)`

Compress text and return stats.

- `input`: Text to compress
- `options.mode`: 'minimal' or 'aggressive'
- `options.budget`: Target token budget

Returns: `{ output, originalTokens, finalTokens, tokensSaved, reductionPercent }`

### `analyze(input)`

Detect content type.

Returns: `{ contentType }` - 'code', 'logs', 'conversation', 'git', 'test', 'docker/infra', 'mixed', or 'unknown'

### `stream(callback, options)`

Create streaming compressor.

- `callback`: Called with each compressed chunk
- `options`: Same as `process()`

Returns: `{ send(text), close() }`

### `version()`

Get TokMan version string.

## Performance

- 95-99% token reduction
- No network requests (runs locally)
- ~2MB WASM binary size
- Works offline
