/**
 * TokMan SDK Types
 *
 * Core types for the token reduction system.
 */
/** Compression mode presets */
type CompressionMode = 'conservative' | 'balanced' | 'aggressive';
/** Content type for adaptive compression */
type ContentType = 'code' | 'conversation' | 'logs' | 'documents' | 'mixed';
/** Layer configuration for fine-grained control */
interface LayerConfig {
    /** Enable/disable the layer */
    enabled: boolean;
    /** Layer-specific threshold (0.0 - 1.0) */
    threshold?: number;
    /** Additional layer parameters */
    params?: Record<string, unknown>;
}
/** Full pipeline configuration */
interface PipelineConfig {
    /** Compression mode (overrides individual layer settings) */
    mode?: CompressionMode;
    /** Content type for adaptive tuning */
    contentType?: ContentType;
    /** Custom layer configurations */
    layers?: Partial<Record<LayerName, LayerConfig>>;
    /** Target token count (0 = no target) */
    targetTokens?: number;
    /** Preserve critical sections (function names, numbers, etc.) */
    preserveCritical?: boolean;
}
/** Names of all 14 compression layers */
type LayerName = 'whitespace' | 'stopwords' | 'punctuation' | 'case_normalization' | 'number_abbr' | 'repetition' | 'template' | 'semantic' | 'structure' | 'context' | 'importance' | 'entropy' | 'huffman' | 'adaptive';
/** Statistics for a compression layer */
interface LayerStats {
    name: LayerName;
    inputTokens: number;
    outputTokens: number;
    reductionPercent: number;
    processingTimeMs: number;
}
/** Compression result */
interface CompressionResult {
    /** Compressed text */
    output: string;
    /** Original token count */
    originalTokens: number;
    /** Final token count */
    finalTokens: number;
    /** Overall reduction percentage */
    reductionPercent: number;
    /** Per-layer statistics */
    layerStats: LayerStats[];
    /** Detected content type */
    detectedContentType?: ContentType;
    /** Processing time in milliseconds */
    processingTimeMs: number;
}
/** Analysis result for content inspection */
interface AnalysisResult {
    /** Detected content type */
    contentType: ContentType;
    /** Confidence score (0.0 - 1.0) */
    confidence: number;
    /** Recommended compression mode */
    recommendedMode: CompressionMode;
    /** Token statistics */
    tokenStats: {
        total: number;
        unique: number;
        avgWordLength: number;
        repetitionRate: number;
    };
    /** Content characteristics */
    characteristics: {
        hasCode: boolean;
        hasConversation: boolean;
        hasLogs: boolean;
        hasNumbers: boolean;
        hasUrls: boolean;
    };
}
/** Server statistics */
interface ServerStats {
    version: string;
    layerCount: number;
    totalCompressions?: number;
    avgReductionPercent?: number;
}
/** Error response from the server */
interface ErrorResponse {
    error: string;
    code: string;
    details?: Record<string, unknown>;
}
/** Client configuration */
interface TokManConfig {
    /** Server URL (default: http://localhost:8080) */
    baseUrl?: string;
    /** Request timeout in milliseconds (default: 30000) */
    timeout?: number;
    /** Default compression mode */
    defaultMode?: CompressionMode;
    /** Enable request/response logging */
    debug?: boolean;
    /** Custom headers for all requests */
    headers?: Record<string, string>;
}
/** Streaming chunk for chunked compression */
interface StreamChunk {
    chunk: string;
    isFinal: boolean;
    tokensProcessed: number;
}

/**
 * TokMan SDK Client
 *
 * HTTP client for the TokMan compression server.
 */

/**
 * TokMan SDK Client
 *
 * @example
 * ```typescript
 * import { TokMan } from '@tokman/sdk';
 *
 * const client = new TokMan({ baseUrl: 'http://localhost:8080' });
 *
 * // Basic compression
 * const result = await client.compress('Your long text here...');
 * console.log(`Reduced from ${result.originalTokens} to ${result.finalTokens} tokens`);
 *
 * // Adaptive compression (auto-detects content type)
 * const adaptive = await client.compressAdaptive('function main() { return 42; }');
 * console.log(`Detected: ${adaptive.detectedContentType}`);
 * ```
 */
declare class TokMan {
    private readonly config;
    constructor(config?: TokManConfig);
    /**
     * Check server health
     */
    health(): Promise<{
        status: string;
        version: string;
    }>;
    /**
     * Compress text with optional mode
     */
    compress(input: string, options?: {
        mode?: CompressionMode;
        targetTokens?: number;
    }): Promise<CompressionResult>;
    /**
     * Compress text with adaptive content detection
     */
    compressAdaptive(input: string, options?: {
        contentType?: ContentType;
        targetTokens?: number;
    }): Promise<CompressionResult>;
    /**
     * Analyze content without compression
     */
    analyze(input: string): Promise<AnalysisResult>;
    /**
     * Get server statistics
     */
    stats(): Promise<ServerStats>;
    /**
     * Make an HTTP request to the server
     */
    private request;
    /**
     * Transform snake_case keys to camelCase
     */
    private transformResponse;
}
/**
 * TokMan SDK Error
 */
declare class TokManError extends Error {
    readonly code: string;
    readonly statusCode: number;
    constructor(message: string, code: string, statusCode: number);
}

export { type AnalysisResult, type CompressionMode, type CompressionResult, type ContentType, type ErrorResponse, type LayerConfig, type LayerName, type LayerStats, type PipelineConfig, type ServerStats, type StreamChunk, TokMan, type TokManConfig, TokManError };
