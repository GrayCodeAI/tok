/**
 * TokMan SDK for Node.js/TypeScript
 * 
 * @packageDocumentation
 * 
 * The world's best token reduction system with 14 research-backed layers.
 * 
 * @example
 * ```typescript
 * import { TokMan } from '@tokman/sdk';
 * 
 * const client = new TokMan();
 * 
 * // Compress text
 * const result = await client.compress('Your long text...');
 * console.log(`${result.reductionPercent}% reduction`);
 * 
 * // Adaptive compression
 * const adaptive = await client.compressAdaptive(codeString);
 * ```
 */

export { TokMan, TokManError } from './client';
export * from './types';
