import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { TokMan, TokManError } from './client';

describe('TokMan Client', () => {
  const originalFetch = global.fetch;
  
  beforeEach(() => {
    vi.resetAllMocks();
  });
  
  afterEach(() => {
    global.fetch = originalFetch;
    vi.useRealTimers();
  });

  describe('constructor', () => {
    it('should use default config when none provided', () => {
      const client = new TokMan();
      // Config is private, but we can test behavior
      expect(client).toBeInstanceOf(TokMan);
    });

    it('should accept custom config', () => {
      const client = new TokMan({
        baseUrl: 'http://custom:9090',
        timeout: 5000,
        defaultMode: 'aggressive',
        debug: true,
      });
      expect(client).toBeInstanceOf(TokMan);
    });
  });

  describe('health', () => {
    it('should fetch health endpoint', async () => {
      const mockResponse = { status: 'ok', version: '1.2.0' };
      
      global.fetch = vi.fn().mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      });

      const client = new TokMan({ baseUrl: 'http://test:8080' });
      const result = await client.health();

      expect(result).toEqual(mockResponse);
      expect(global.fetch).toHaveBeenCalledWith(
        'http://test:8080/health',
        expect.objectContaining({ method: 'GET' })
      );
    });
  });

  describe('compress', () => {
    it('should compress text with default mode', async () => {
      const mockResponse = {
        output: 'compressed',
        original_tokens: 100,
        final_tokens: 50,
        reduction_percent: 50,
        layer_stats: [],
        processing_time_ms: 10,
      };

      global.fetch = vi.fn().mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      });

      const client = new TokMan();
      const result = await client.compress('test input');

      expect(result.originalTokens).toBe(100);
      expect(result.finalTokens).toBe(50);
      expect(result.reductionPercent).toBe(50);
    });

    it('should compress with custom mode', async () => {
      const mockResponse = {
        output: 'compressed',
        original_tokens: 100,
        final_tokens: 30,
        reduction_percent: 70,
        layer_stats: [],
        processing_time_ms: 15,
      };

      global.fetch = vi.fn().mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      });

      const client = new TokMan({ defaultMode: 'conservative' });
      const result = await client.compress('test', { mode: 'aggressive' });

      expect(result.reductionPercent).toBe(70);
      
      const callBody = JSON.parse((global.fetch as any).mock.calls[0][1].body);
      expect(callBody.mode).toBe('aggressive');
    });
  });

  describe('compressAdaptive', () => {
    it('should call adaptive endpoint', async () => {
      const mockResponse = {
        output: 'compressed',
        original_tokens: 100,
        final_tokens: 40,
        reduction_percent: 60,
        layer_stats: [],
        detected_content_type: 'code',
        processing_time_ms: 20,
      };

      global.fetch = vi.fn().mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      });

      const client = new TokMan();
      const result = await client.compressAdaptive('function test() {}');

      expect(result.detectedContentType).toBe('code');
      expect(global.fetch).toHaveBeenCalledWith(
        expect.stringContaining('/compress/adaptive'),
        expect.any(Object)
      );
    });
  });

  describe('analyze', () => {
    it('should analyze content', async () => {
      const mockResponse = {
        content_type: 'code',
        confidence: 0.95,
        recommended_mode: 'conservative',
        token_stats: { total: 100, unique: 60, avg_word_length: 4.5, repetition_rate: 0.2 },
        characteristics: { has_code: true, has_conversation: false, has_logs: false, has_numbers: false, has_urls: false },
      };

      global.fetch = vi.fn().mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      });

      const client = new TokMan();
      const result = await client.analyze('function main() { return 42; }');

      expect(result.contentType).toBe('code');
      expect(result.confidence).toBe(0.95);
    });
  });

  describe('stats', () => {
    it('should fetch server stats', async () => {
      const mockResponse = {
        version: '1.2.0',
        layer_count: 14,
      };

      global.fetch = vi.fn().mockResolvedValue({
        ok: true,
        json: () => Promise.resolve(mockResponse),
      });

      const client = new TokMan();
      const result = await client.stats();

      expect(result.version).toBe('1.2.0');
      expect(result.layerCount).toBe(14);
    });
  });

  describe('error handling', () => {
    it('should throw TokManError on HTTP error', async () => {
      global.fetch = vi.fn().mockResolvedValue({
        ok: false,
        status: 400,
        json: () => Promise.resolve({ error: 'Invalid input', code: 'INVALID_INPUT' }),
      });

      const client = new TokMan();
      
      await expect(client.compress('test')).rejects.toThrow(TokManError);
    });

    it('should have timeout configuration', () => {
      // Timeout is handled via AbortController
      // This test verifies the config is accepted
      const client = new TokMan({ timeout: 5000 });
      expect(client).toBeInstanceOf(TokMan);
    });
  });
});
