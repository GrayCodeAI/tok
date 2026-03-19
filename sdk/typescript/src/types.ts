/**
 * TokMan SDK Types
 * 
 * Core types for the token reduction system.
 */

/** Compression mode presets */
export type CompressionMode = 'conservative' | 'balanced' | 'aggressive';

/** Content type for adaptive compression */
export type ContentType = 'code' | 'conversation' | 'logs' | 'documents' | 'mixed';

/** Layer configuration for fine-grained control */
export interface LayerConfig {
  /** Enable/disable the layer */
  enabled: boolean;
  /** Layer-specific threshold (0.0 - 1.0) */
  threshold?: number;
  /** Additional layer parameters */
  params?: Record<string, unknown>;
}

/** Full pipeline configuration */
export interface PipelineConfig {
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
export type LayerName =
  | 'whitespace'
  | 'stopwords'
  | 'punctuation'
  | 'case_normalization'
  | 'number_abbr'
  | 'repetition'
  | 'template'
  | 'semantic'
  | 'structure'
  | 'context'
  | 'importance'
  | 'entropy'
  | 'huffman'
  | 'adaptive';

/** Statistics for a compression layer */
export interface LayerStats {
  name: LayerName;
  inputTokens: number;
  outputTokens: number;
  reductionPercent: number;
  processingTimeMs: number;
}

/** Compression result */
export interface CompressionResult {
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
export interface AnalysisResult {
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
export interface ServerStats {
  version: string;
  layerCount: number;
  totalCompressions?: number;
  avgReductionPercent?: number;
}

/** Error response from the server */
export interface ErrorResponse {
  error: string;
  code: string;
  details?: Record<string, unknown>;
}

/** Client configuration */
export interface TokManConfig {
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
export interface StreamChunk {
  chunk: string;
  isFinal: boolean;
  tokensProcessed: number;
}
