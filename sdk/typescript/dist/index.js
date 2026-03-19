"use strict";
var __defProp = Object.defineProperty;
var __getOwnPropDesc = Object.getOwnPropertyDescriptor;
var __getOwnPropNames = Object.getOwnPropertyNames;
var __hasOwnProp = Object.prototype.hasOwnProperty;
var __export = (target, all) => {
  for (var name in all)
    __defProp(target, name, { get: all[name], enumerable: true });
};
var __copyProps = (to, from, except, desc) => {
  if (from && typeof from === "object" || typeof from === "function") {
    for (let key of __getOwnPropNames(from))
      if (!__hasOwnProp.call(to, key) && key !== except)
        __defProp(to, key, { get: () => from[key], enumerable: !(desc = __getOwnPropDesc(from, key)) || desc.enumerable });
  }
  return to;
};
var __toCommonJS = (mod) => __copyProps(__defProp({}, "__esModule", { value: true }), mod);

// src/index.ts
var index_exports = {};
__export(index_exports, {
  TokMan: () => TokMan,
  TokManError: () => TokManError
});
module.exports = __toCommonJS(index_exports);

// src/client.ts
var DEFAULT_CONFIG = {
  baseUrl: "http://localhost:8080",
  timeout: 3e4,
  defaultMode: "balanced",
  debug: false,
  headers: {}
};
var TokMan = class {
  constructor(config = {}) {
    this.config = {
      baseUrl: config.baseUrl ?? DEFAULT_CONFIG.baseUrl,
      timeout: config.timeout ?? DEFAULT_CONFIG.timeout,
      defaultMode: config.defaultMode ?? DEFAULT_CONFIG.defaultMode,
      debug: config.debug ?? DEFAULT_CONFIG.debug,
      headers: config.headers ?? DEFAULT_CONFIG.headers
    };
  }
  /**
   * Check server health
   */
  async health() {
    return this.request("GET", "/health");
  }
  /**
   * Compress text with optional mode
   */
  async compress(input, options = {}) {
    const mode = options.mode ?? this.config.defaultMode;
    return this.request("POST", "/compress", {
      input,
      mode,
      target_tokens: options.targetTokens
    });
  }
  /**
   * Compress text with adaptive content detection
   */
  async compressAdaptive(input, options = {}) {
    return this.request("POST", "/compress/adaptive", {
      input,
      content_type: options.contentType,
      target_tokens: options.targetTokens
    });
  }
  /**
   * Analyze content without compression
   */
  async analyze(input) {
    return this.request("POST", "/analyze", { input });
  }
  /**
   * Get server statistics
   */
  async stats() {
    return this.request("GET", "/stats");
  }
  /**
   * Make an HTTP request to the server
   */
  async request(method, path, body) {
    const url = `${this.config.baseUrl}${path}`;
    const headers = {
      "Content-Type": "application/json",
      ...this.config.headers
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
        body: body ? JSON.stringify(body) : void 0,
        signal: controller.signal
      });
      clearTimeout(timeoutId);
      if (!response.ok) {
        const error = await response.json();
        throw new TokManError(error.error || "Request failed", error.code, response.status);
      }
      const data = await response.json();
      if (this.config.debug) {
        console.log(`[TokMan] Response:`, data);
      }
      return this.transformResponse(data);
    } catch (error) {
      clearTimeout(timeoutId);
      if (error instanceof TokManError) {
        throw error;
      }
      if (error instanceof Error) {
        if (error.name === "AbortError") {
          throw new TokManError("Request timeout", "TIMEOUT", 408);
        }
        throw new TokManError(error.message, "NETWORK_ERROR", 0);
      }
      throw new TokManError("Unknown error", "UNKNOWN", 0);
    }
  }
  /**
   * Transform snake_case keys to camelCase
   */
  transformResponse(data) {
    if (data === null || data === void 0) {
      return data;
    }
    if (Array.isArray(data)) {
      return data.map((item) => this.transformResponse(item));
    }
    if (typeof data === "object") {
      const result = {};
      for (const [key, value] of Object.entries(data)) {
        const camelKey = key.replace(/_([a-z])/g, (_, letter) => letter.toUpperCase());
        result[camelKey] = this.transformResponse(value);
      }
      return result;
    }
    return data;
  }
};
var TokManError = class extends Error {
  constructor(message, code, statusCode) {
    super(message);
    this.code = code;
    this.statusCode = statusCode;
    this.name = "TokManError";
  }
};
// Annotate the CommonJS export names for ESM import in node:
0 && (module.exports = {
  TokMan,
  TokManError
});
