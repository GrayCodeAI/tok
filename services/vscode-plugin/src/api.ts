import axios from 'axios';
import { AnalysisResult } from './analyzer';

export class TokManAPI {
  private client = axios.create();

  constructor(private endpoint: string) {
    this.client = axios.create({
      baseURL: endpoint,
      timeout: 10000,
    });
  }

  async analyze(text: string, model: string): Promise<AnalysisResult> {
    try {
      const response = await this.client.post('/analyze', {
        content: text,
        model,
        mode: 'aggressive',
      });

      return {
        originalTokens: response.data.original_tokens || 0,
        compressedTokens: response.data.compressed_tokens || 0,
        tokensSaved: response.data.saved_tokens || 0,
        savingsPercent: response.data.savings_percent || 0,
        compressionRatio: response.data.compression_ratio || 1,
        processingTimeMs: response.data.processing_time_ms || 0,
        filterMode: response.data.filter_mode || 'aggressive',
      };
    } catch (error) {
      console.error('API Error:', error);
      // Fallback: rough local estimation
      return this.estimateTokens(text);
    }
  }

  async getStats(teamId: string) {
    try {
      const response = await this.client.get(`/stats/${teamId}`);
      return response.data;
    } catch (error) {
      console.error('Stats API Error:', error);
      return null;
    }
  }

  private estimateTokens(text: string): AnalysisResult {
    // Rough estimation: ~4 chars per token
    const originalTokens = Math.ceil(text.length / 4);
    const compressedTokens = Math.ceil(originalTokens * 0.7); // Assume 30% compression

    return {
      originalTokens,
      compressedTokens,
      tokensSaved: originalTokens - compressedTokens,
      savingsPercent: 30,
      compressionRatio: 0.7,
      processingTimeMs: 0,
      filterMode: 'local_estimate',
    };
  }

  setEndpoint(endpoint: string) {
    this.endpoint = endpoint;
    this.client = axios.create({
      baseURL: endpoint,
      timeout: 10000,
    });
  }
}
