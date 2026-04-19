import type { Plugin } from "@opencode-ai/plugin"

// Tok OpenCode plugin — rewrites commands to use tok for token savings.

export const TokOpenCodePlugin: Plugin = async ({ $ }) => {
  try {
    await $`which tok`.quiet()
  } catch {
    console.warn("[tok] tok binary not found in PATH — plugin disabled")
    return {}
  }

  return {
    "tool.execute.before": async (input, output) => {
      const tool = String(input?.tool ?? "").toLowerCase()
      if (tool !== "bash" && tool !== "shell") return
      const args = output?.args
      if (!args || typeof args !== "object") return

      const command = (args as Record<string, unknown>).command
      if (typeof command !== "string" || !command) return

      try {
        const result = await $`tok rewrite ${command}`.quiet().nothrow()
        const rewritten = String(result.stdout).trim()
        if (rewritten && rewritten !== command) {
          ;(args as Record<string, unknown>).command = rewritten
        }
      } catch {
        // tok rewrite failed — pass through unchanged
      }
    },
  }
}
