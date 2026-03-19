/**
 * TokMan WASM JavaScript Wrapper
 * Browser-based token compression with 14-layer pipeline
 */

let wasmModule = null;
let wasmReady = false;

/**
 * Initialize TokMan WASM module
 * @param {string} wasmPath - Path to tokman.wasm file
 * @returns {Promise<void>}
 */
async function init(wasmPath = './tokman.wasm') {
    const go = new Go();
    const result = await WebAssembly.instantiateStreaming(fetch(wasmPath), go.importObject);
    wasmModule = result.instance;
    go.run(wasmModule);
    wasmReady = true;
}

/**
 * Process content with token compression
 * @param {string} input - Text to compress
 * @param {Object} options - Compression options
 * @param {string} options.mode - 'minimal' or 'aggressive'
 * @param {number} options.budget - Target token budget
 * @returns {Object} Compression result
 */
function process(input, options = {}) {
    if (!wasmReady) {
        throw new Error('TokMan not initialized. Call init() first.');
    }
    return TokMan.process(input, options);
}

/**
 * Analyze content type
 * @param {string} input - Text to analyze
 * @returns {Object} Analysis result with contentType
 */
function analyze(input) {
    if (!wasmReady) {
        throw new Error('TokMan not initialized. Call init() first.');
    }
    return TokMan.analyze(input);
}

/**
 * Create streaming compressor
 * @param {Function} callback - Called with each compressed chunk
 * @param {Object} options - Compression options
 * @returns {Object} Stream controller with send() and close()
 */
function stream(callback, options = {}) {
    if (!wasmReady) {
        throw new Error('TokMan not initialized. Call init() first.');
    }
    return TokMan.stream(callback, options);
}

/**
 * Get TokMan version
 * @returns {string}
 */
function version() {
    return wasmReady ? TokMan.version : null;
}

// Browser module export
if (typeof window !== 'undefined') {
    window.TokManWASM = {
        init,
        process,
        analyze,
        stream,
        version
    };
}

// Node.js module export
if (typeof module !== 'undefined' && module.exports) {
    module.exports = {
        init,
        process,
        analyze,
        stream,
        version
    };
}
