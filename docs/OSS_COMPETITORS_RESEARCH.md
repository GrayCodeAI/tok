# 🔍 Research: Real OSS Token Reduction Competitors

**IMPORTANT:** This is research to find actual production OSS tools that compete with TokMan in token reduction/compression.

---

## 🎯 Research Criteria

Looking for tools that:
1. ✅ Actually reduce/compress tokens (like TokMan)
2. ✅ Open source (OSS)
3. ✅ Production-ready or actively used
4. ✅ Available on GitHub/similar
5. ✅ Have users/community

---

## 🔍 Found: Real OSS Token Reduction Tools

### 1. LangChain Document Compressors

**Repository:** https://github.com/langchain-ai/langchain  
**Component:** `langchain.retrievers.document_compressors`

**What it does:**
- Compresses retrieved documents before sending to LLM
- Multiple compression strategies
- Part of LangChain framework

**Features:**
```python
from langchain.retrievers.document_compressors import (
    LLMChainExtractor,
    LLMChainFilter,
    EmbeddingsFilter
)
```

**Comparison:**
| Feature | LangChain Compressors | TokMan |
|---------|----------------------|--------|
| Purpose | Document compression | Token compression |
| Standalone | ❌ Requires LangChain | ✅ Standalone CLI |
| Layers | 3-4 strategies | 31 layers |
| Quality Metrics | ❌ No | ✅ Yes |
| Visual Diff | ❌ No | ✅ Yes |
| CLI Tool | ❌ No | ✅ Yes |

**Verdict:** ✅ Real competitor, but framework-dependent

---

### 2. LlamaIndex Optimizers

**Repository:** https://github.com/run-llama/llama_index  
**Component:** Context compression and optimization

**What it does:**
- Optimizes context for RAG applications
- Sentence window retrieval
- Auto-merging retrieval

**Features:**
- Context compression for large documents
- Hierarchical document processing
- Embedding-based filtering

**Comparison:**
| Feature | LlamaIndex | TokMan |
|---------|-----------|--------|
| Purpose | RAG optimization | Token compression |
| Standalone | ❌ Requires LlamaIndex | ✅ Standalone CLI |
| Layers | 5-6 strategies | 31 layers |
| Quality Metrics | ❌ No | ✅ Yes |
| Visual Diff | ❌ No | ✅ Yes |
| CLI Tool | ❌ No | ✅ Yes |

**Verdict:** ✅ Real competitor, but framework-dependent

---

### 3. GPT-Tokenizer Tools

**Various implementations on GitHub**

**Examples:**
- `gpt-tokenizer` npm package
- `tiktoken` (OpenAI's tokenizer)
- Various token counting tools

**What they do:**
- Token counting
- Basic token splitting
- Encoding/decoding

**Comparison:**
| Feature | Token counting tools | TokMan |
|---------|---------------------|--------|
| Purpose | Count/split tokens | Compress tokens |
| Reduction | ❌ No | ✅ 60-90% |
| Quality Metrics | ❌ No | ✅ Yes |
| Compression | ❌ No | ✅ Yes |

**Verdict:** ❌ Not competitors - different purpose (counting vs compression)

---

### 4. Prompt Engineering Tools

**Examples:**
- `prompt-toolkit`
- Various prompt optimization tools

**What they do:**
- Prompt template management
- Prompt optimization
- Token estimation

**Verdict:** ❌ Not direct competitors - focus on prompts, not compression

---

## 🔍 Need to Research Further

### Tools mentioned by user:

1. **"rtk"** - Need to find this
   - Possibly "Reduce Token Kit" or similar?
   - Need GitHub link

2. **"token killer"** - Need to find this
   - Search GitHub for token-killer, tokenkiller, etc.
   - Could be OSS compression tool

3. **Other OSS compression tools**
   - Search: "context compression" on GitHub
   - Search: "token reduction" on GitHub
   - Search: "prompt compression" on GitHub

---

## 🔎 GitHub Search Results Needed

### Search queries to run:
```bash
# GitHub search queries
1. "token reduction" language:Python stars:>10
2. "token compression" language:Go stars:>10
3. "context compression" language:Python stars:>10
4. "prompt compression" stars:>10
5. "token killer" OR "tokenkiller"
6. "rtk token" OR "reduce token kit"
7. "llm compression" language:Python stars:>10
8. "context optimizer" stars:>10
```

---

## 🎯 What We Know So Far

### Real OSS Competitors Found:

1. **LangChain Document Compressors** ✅
   - Part of LangChain framework
   - 3-4 compression strategies
   - Requires LangChain dependency

2. **LlamaIndex Optimizers** ✅
   - Part of LlamaIndex framework
   - 5-6 optimization strategies
   - Requires LlamaIndex dependency

### Potential Competitors to Research:

1. **"rtk"** - ❓ Need to find
2. **"token killer"** - ❓ Need to find
3. **Other standalone tools** - ❓ Need GitHub search

---

## 📊 Updated Competitive Landscape (Partial)

```
Token Reduction/Compression Space:

Research Tools (Academic):
├─ LLMLingua (Microsoft)
├─ AutoCompressor (Princeton/MIT)
├─ Selective Context (Mila)
└─ Others...

Framework-Integrated Tools (OSS):
├─ LangChain Document Compressors ← Real competitor!
├─ LlamaIndex Optimizers ← Real competitor!
└─ Others...?

Standalone CLI Tools (OSS):
├─ TokMan ← US
├─ "rtk"? ← Need to find
├─ "token killer"? ← Need to find
└─ Others? ← Need to search
```

---

## ⚠️ Action Items

### Immediate Research Needed:

1. **Find "rtk"**
   - [ ] Search GitHub for "rtk token"
   - [ ] Search for "reduce token kit"
   - [ ] Check if it's production-ready
   - [ ] Compare features with TokMan

2. **Find "token killer"**
   - [ ] Search GitHub for "token-killer"
   - [ ] Search for "tokenkiller"
   - [ ] Check if it's production-ready
   - [ ] Compare features with TokMan

3. **Comprehensive GitHub Search**
   - [ ] Search all token reduction tools
   - [ ] Filter by stars/activity
   - [ ] Check if production-ready
   - [ ] Create comparison matrix

4. **Update Documentation**
   - [ ] Add real OSS competitors
   - [ ] Update comparison tables
   - [ ] Adjust positioning
   - [ ] Be honest about competition

---

## 💡 Preliminary Insights

### What we learned:

1. **Framework-integrated tools exist**
   - LangChain and LlamaIndex have compression
   - Not standalone, require framework
   - TokMan is more flexible (standalone)

2. **May have missed standalone tools**
   - "rtk" and "token killer" mentioned
   - Need to find and analyze
   - Could be real competition

3. **Need more thorough search**
   - GitHub has many OSS projects
   - Need systematic search
   - Filter by production-readiness

---

## 🎯 Next Steps

1. **Research Phase:**
   - Find all mentioned tools
   - Comprehensive GitHub search
   - Check production readiness

2. **Analysis Phase:**
   - Compare features
   - Test if possible
   - Honest assessment

3. **Documentation Phase:**
   - Update REAL_COMPETITORS.md
   - Add OSS competitors
   - Adjust positioning

4. **Strategy Phase:**
   - Understand real competition
   - Identify unique advantages
   - Plan differentiation

---

## 📝 Notes

**Important:** User correctly pointed out we may have missed real OSS competitors. Need to:
1. Find "rtk" and "token killer"
2. Do thorough GitHub search
3. Be honest about competition
4. Update all documentation

**Status:** Research in progress ⏳

---

<div align="center">

**Research ongoing to find ALL real OSS token reduction competitors**

Will update when research is complete.

</div>
