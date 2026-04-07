# Session Continuity System Architecture

## Overview
Session Continuity provides PreCompact hooks and session management to maintain context across AI agent interactions, inspired by OMNI's semantic signal engine.

## Core Concepts

### Session
A session represents a continuous interaction context between an AI agent and the user. It persists across:
- Multiple commands
- Tool invocations
- Conversation turns
- Time periods

### Hooks
The system provides 4 hook points:
1. **SessionStart** - When a new session begins
2. **PreToolUse** - Before executing a tool
3. **PostToolUse** - After tool execution
4. **PreCompact** - Before conversation compression

### PreCompact Optimization
PreCompact is the critical hook that allows TokMan to:
- Extract key information before context pruning
- Create summaries of long conversations
- Preserve important state
- Inject context into compressed output

## Data Model

```
Session
├── ID (UUID)
├── Agent (claude, cursor, copilot, etc.)
├── ProjectPath
├── StartedAt
├── LastActivity
├── ContextBlocks []
│   ├── Type (user_query, tool_result, summary)
│   ├── Content
│   ├── Timestamp
│   └── Tokens
├── State
│   ├── Variables map[string]interface{}
│   ├── Focus string
│   └── NextAction string
└── Metadata
    ├── TotalTurns
    ├── TotalTokens
    └── CompressionRatio
```

## Hook Flow

```
SessionStart
    ↓
PreToolUse → [Execute Tool] → PostToolUse
    ↓
[Context Growing...]
    ↓
PreCompact → [Compress] → [Inject Summary]
    ↓
[Continue Session]
```

## Integration Points

### Pipeline Integration
Sessions integrate with the filter pipeline to:
- Inject session context into filtered output
- Apply session-aware compression
- Track token savings per session

### Archive Integration
Sessions work with RewindStore to:
- Archive session snapshots
- Restore previous session state
- Track session history

## Use Cases

### 1. Long Conversations
When conversation grows beyond context window:
- PreCompact extracts key learnings
- Creates executive summary
- Injects summary at start of compressed context

### 2. Multi-Tool Workflows
Track state across multiple tool calls:
- Remember file paths from earlier calls
- Maintain context of what was done
- Enable "continue where I left off"

### 3. Cross-Session Learning
- Hot file tracking across sessions
- Pattern discovery over time
- Project context persistence

## Implementation

### SessionManager
Central coordinator for all session operations:
- Create/destroy sessions
- Manage hooks
- Persist state
- Handle restoration

### HookRegistry
Manages hook implementations:
- Register custom hooks
- Execute hooks in order
- Handle hook errors

### ContextInjector
Injects session context into output:
- Format context blocks
- Calculate token budgets
- Optimize for relevance
