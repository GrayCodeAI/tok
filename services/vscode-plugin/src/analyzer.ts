import { TokManAPI } from './api';

export interface AnalysisResult {
  originalTokens: number;
  compressedTokens: number;
  tokensSaved: number;
  savingsPercent: number;
  compressionRatio: number;
  processingTimeMs: number;
  filterMode: string;
}

export class TokenAnalyzer {
  private cache: Map<string, AnalysisResult> = new Map();
  private cacheTimeout = 5 * 60 * 1000; // 5 minutes

  constructor(private api: TokManAPI, private model: string) {}

  async analyze(text: string): Promise<AnalysisResult> {
    if (!text) {
      return {
        originalTokens: 0,
        compressedTokens: 0,
        tokensSaved: 0,
        savingsPercent: 0,
        compressionRatio: 1,
        processingTimeMs: 0,
        filterMode: 'none',
      };
    }

    // Check cache
    const cached = this.cache.get(text);
    if (cached) {
      return cached;
    }

    try {
      const result = await this.api.analyze(text, this.model);
      this.cache.set(text, result);

      // Clear cache after timeout
      setTimeout(() => {
        this.cache.delete(text);
      }, this.cacheTimeout);

      return result;
    } catch (error) {
      console.error('Analysis error:', error);
      throw error;
    }
  }

  clearCache() {
    this.cache.clear();
  }

  getStats() {
    return {
      cachedAnalyses: this.cache.size,
      model: this.model,
    };
  }
}
