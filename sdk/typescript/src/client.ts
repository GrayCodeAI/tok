/**
 * TokMan SDK Client
 * 
 * HTTP client for the TokMan compression server.
 */

import type {
  TokManConfig,
  CompressionResult,
  AnalysisResult,
  ServerStats,
  CompressionMode,
  ContentType,
  ErrorResponse,
} from './types';

/** Default configuration */
const DEFAULT_CONFIG: Required<Omit<TokManConfig, 'headers'>> & { headers: Record<string, string> } = {
  baseUrl: 'http://localhost:8080',
  timeout: 30000,
  defaultMode: 'balanced',
  debug: false,
  headers: {},
};

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
export class TokMan {
  private readonly config: Required<Omit<TokManConfig, 'headers'>> & { headers: Record<string, string> };

  constructor(config: TokManConfig = {}) {
    this.config = {
      baseUrl: config.baseUrl ?? DEFAULT_CONFIG.baseUrl,
      timeout: config.timeout ?? DEFAULT_CONFIG.timeout,
      defaultMode: config.defaultMode ?? DEFAULT_CONFIG.defaultMode,
      debug: config.debug ?? DEFAULT_CONFIG.debug,
      headers: config.headers ?? DEFAULT_CONFIG.headers,
    };
  }

  /**
   * Check server health
   */
  async health(): Promise<{ status: string; version: string }> {
    return this.request<{ status: string; version: string }>('GET', '/health');
  }

  /**
   * Compress text with optional mode
   */
  async compress(
    input: string,
    options: {
      mode?: CompressionMode;
      targetTokens?: number;
    } = {}
  ): Promise<CompressionResult> {
    const mode = options.mode ?? this.config.defaultMode;
    return this.request<CompressionResult>('POST', '/compress', {
      input,
      mode,
      target_tokens: options.targetTokens,
    });
  }

  /**
   * Compress text with adaptive content detection
   */
  async compressAdaptive(
    input: string,
    options: {
      contentType?: ContentType;
      targetTokens?: number;
    } = {}
  ): Promise<CompressionResult> {
    return this.request<CompressionResult>('POST', '/compress/adaptive', {
      input,
      content_type: options.contentType,
      target_tokens: options.targetTokens,
    });
  }

  /**
   * Analyze content without compression
   */
  async analyze(input: string): Promise<AnalysisResult> {
    return this.request<AnalysisResult>('POST', '/analyze', { input });
  }

  /**
   * Get server statistics
   */
  async stats(): Promise<ServerStats> {
    return this.request<ServerStats>('GET', '/stats');
  }

  /**
   * Make an HTTP request to the server
   */
  private async request<T>(
    method: string,
    path: string,
    body?: Record<string, unknown>
  ): Promise<T> {
    const url = `${this.config.baseUrl}${path}`;
    
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...this.config.headers,
    };

    if (this.config.debug) {
      console.log(`[TokMan] ${method} ${url}`, body);
    }

    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), this.config.timeout);

    try {
      const response = await fetch(url, {
        method,
        headers,
        body: body ? JSON.stringify(body) : undefined,
        signal: controller.signal,
      });

      clearTimeout(timeoutId);

      if (!response.ok) {
        const error = await response.json() as ErrorResponse;
        throw new TokManError(error.error || 'Request failed', error.code, response.status);
      }

      const data = await response.json();
      
      if (this.config.debug) {
        console.log(`[TokMan] Response:`, data);
      }

      // Transform snake_case to camelCase
      return this.transformResponse(data) as T;
    } catch (error) {
      clearTimeout(timeoutId);
      
      if (error instanceof TokManError) {
        throw error;
      }
      
      if (error instanceof Error) {
        if (error.name === 'AbortError') {
          throw new TokManError('Request timeout', 'TIMEOUT', 408);
        }
        throw new TokManError(error.message, 'NETWORK_ERROR', 0);
      }
      
      throw new TokManError('Unknown error', 'UNKNOWN', 0);
    }
  }

  /**
   * Transform snake_case keys to camelCase
   */
  private transformResponse(data: unknown): unknown {
    if (data === null || data === undefined) {
      return data;
    }

    if (Array.isArray(data)) {
      return data.map(item => this.transformResponse(item));
    }

    if (typeof data === 'object') {
      const result: Record<string, unknown> = {};
      for (const [key, value] of Object.entries(data as Record<string, unknown>)) {
        const camelKey = key.replace(/_([a-z])/g, (_, letter) => letter.toUpperCase());
        result[camelKey] = this.transformResponse(value);
      }
      return result;
    }

    return data;
  }
}

/**
 * TokMan SDK Error
 */
export class TokManError extends Error {
  constructor(
    message: string,
    public readonly code: string,
    public readonly statusCode: number
  ) {
    super(message);
    this.name = 'TokManError';
  }
}

// Re-export types
export * from './types';
