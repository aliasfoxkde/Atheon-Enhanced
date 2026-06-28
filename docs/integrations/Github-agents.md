# GitHub AI Agent & Model Integration

GitHub offers several AI integration points relevant to an open-source Go project like Atheon. All of the options below are free for public repositories.

---

## GitHub Models API

**What it is:** A free inference API (`api.github.com/models`) backed by your GitHub PAT. No credit card. Rate-limited but sufficient for CI use.

**Available models (mid-2026):** GPT-4o, GPT-4o-mini, Llama 3.1 (8B, 70B, 405B), DeepSeek-R1, Mistral Large, Cohere Command R+, Phi-3.

**Authentication:** Standard GitHub PAT — no special scopes needed beyond `read:user`.

### Using GitHub Models in a workflow

```yaml
- name: AI-assisted pattern review
  env:
    GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
  run: |
    curl -s https://api.github.com/models/openai/gpt-4o/chat/completions \
      -H "Authorization: Bearer $GITHUB_TOKEN" \
      -H "Content-Type: application/json" \
      -d '{
        "messages": [
          {"role":"user","content":"Review this Atheon pattern YAML for false positive risk: '"$(cat community/secrets/new-pattern.yaml)"'"}
        ]
      }' | jq -r '.choices[0].message.content'
```

This is useful for automated pattern review on PRs — flag patterns with broad regexes before human review.

### Using GitHub Models from Go

```go
import "github.com/openai/openai-go"

client := openai.NewClient(
    option.WithBaseURL("https://api.github.com/models"),
    option.WithAPIKey(os.Getenv("GITHUB_TOKEN")),
)
```

The endpoint is OpenAI-SDK-compatible, so any OpenAI Go/Python/JS client works.

---

## GitHub Agentic Workflows (Technical Preview)

**What it is:** First-party AI agents (Claude, ChatGPT, Gemini, Cognition) that run as GitHub Actions steps. Available in technical preview as of February 2026. MIT-licensed, open-source.

**Cost:** Free during technical preview. Requires a paid GitHub Copilot subscription post-GA.

**How it works:** Agents run in sandboxed read-only mode by default. They can read files, run commands, and emit suggestions. Write access (creating commits) is opt-in.

### Example: Automated PR review agent

```yaml
name: AI Pattern Review
on:
  pull_request:
    paths: ['community/**/*.yaml']

jobs:
  ai-review:
    runs-on: ubuntu-latest
    permissions:
      pull-requests: write
    steps:
      - uses: actions/checkout@v4
      - name: Review changed patterns
        uses: github/agent-action@v1   # agentic workflow action (preview)
        with:
          model: claude-sonnet
          prompt: |
            Review the changed community YAML pattern files in this PR.
            Check: regex breadth (would this match common identifiers?),
            presence of test cases (true_positives / false_positives fields),
            and whether severity matches the actual risk.
            Post a concise review comment.
          tools: [read_file, post_pr_comment]
```

> **Note:** The exact action name and API are still in technical preview and may change. Monitor [github.blog/changelog](https://github.blog/changelog) for GA announcements.

### When to use agentic workflows vs. static CI

| Use case | Approach |
|----------|---------|
| Validate pattern YAML schema | Static CI (jq, yamllint) — deterministic |
| Flag overly broad regexes | AI review — judgment call |
| Summarize what a PR changes | AI — saves reviewer time |
| Enforce test coverage threshold | Static CI — Codecov |
| Suggest pattern improvements | AI — GitHub Models API in a workflow step |

---

## GitHub Copilot Extensions (MCP)

**What it is:** MCP servers registered with GitHub Copilot become available as tools in all Copilot surfaces (IDE chat, CLI, web chat).

**Status:** Available in VS Code, JetBrains, Xcode, Cursor, Windsurf. No central registry — users configure manually.

**For Atheon:** See `docs/integrations/mcp.md` for the full setup guide. The MCP server (`atheon-mcp`) works with any MCP-compatible client, not just GitHub Copilot.

---

## GitHub Copilot Workspace

**What it is:** A GitHub-hosted agentic environment for multi-file edits. You describe a task; Workspace generates a plan, edits files, and runs tests in a sandboxed environment.

**Relevance to Atheon:** Useful for reviewing community-contributed pattern PRs. Open the PR in Workspace, ask it to verify the pattern tests pass and check for false positives on the Atheon codebase.

**Cost:** Paid Copilot feature. No free tier for public repos.

---

## Recommendation: What to Set Up Now (Free)

In priority order:

1. **GitHub Models API in CI** — Add a workflow step that uses `GITHUB_TOKEN` + GitHub Models to auto-review new `community/` pattern PRs. Zero setup cost, immediate value.
2. **MCP server docs** — Make `atheon-mcp` easy to discover (see `docs/integrations/mcp.md`). Users configuring it in their AI client creates organic adoption.
3. **Agentic Workflows** — Opt into the technical preview when it goes GA. The pattern review use case is a natural fit.
4. **goreleaser** — Ship `atheon-mcp` as a release binary so users don't need Go installed to use the MCP server.
