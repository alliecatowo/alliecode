# AllieCode Roadmap

## Vision

AllieCode is an **open-source, model-agnostic AI coding agent** that works with any LLM provider. It's local-first, privacy-respecting, and built for the community.

Unlike vendor-locked alternatives, AllieCode lets you choose your model, your provider, and your workflow. Run it with Ollama on your laptop or Claude/GPT in the cloud — same tool, same experience.

## Core Principles

- **Model-agnostic**: Works with Anthropic, OpenAI, Google, Mistral, Ollama, and any OpenAI-compatible API
- **Local-first**: No account needed. Ollama is the default provider. Your code stays on your machine.
- **Open source**: Apache 2.0. Community-driven. No telemetry.
- **Safe by default**: Permission system with risk classification. Git-native checkpointing for undo.
- **Extensible**: WASM plugin sandbox, community skills marketplace, MCP support.

## Differentiators

### Multi-Model Routing
Use a cheap/fast model for simple operations (file reads, grep, git status) and a powerful model for reasoning and complex code generation. Configurable routing rules save 60-70% on API costs.

### Git-Native Checkpointing
Every destructive action auto-snapshots to a shadow branch. `ac undo` instantly rolls back any change. Zero-risk experimentation.

### WASM Plugin Sandbox
Skills and plugins can be written in any language (Rust, Go, Python, TypeScript) and compiled to WASM. Sandboxed execution for security. Community marketplace: `ac install @community/rails-expert`.

### Pair Programming Mode
Two models collaborate: Model A proposes a solution, Model B critiques it, consensus executes. Catches more bugs, produces better code.

## Milestones

### v0.1.0 — Foundation (Current)
- [x] Core agent loop with streaming
- [x] Provider abstraction (Anthropic, OpenAI, Ollama, OpenAI-compat)
- [x] Essential tools (bash, file ops, grep, glob, web fetch)
- [x] Permission system with bash command classification
- [x] Bubbletea TUI
- [x] MCP client support
- [x] Skills/slash-command system
- [x] Event hooks
- [x] YAML configuration
- [x] Git-native checkpointing & undo

### v0.2.0 — Multi-Model & Polish
- [ ] Multi-model routing (fast model for tools, powerful for reasoning)
- [ ] Context auto-compaction improvements
- [ ] Session persistence & resume
- [ ] Image/vision support across providers
- [ ] Improved diff rendering
- [ ] `ac init` project setup wizard

### v0.3.0 — Plugins & Community
- [ ] WASM plugin runtime (wasmtime-go)
- [ ] Plugin manifest format
- [ ] `ac install` / `ac publish` for community plugins
- [ ] Plugin sandboxing & permissions
- [ ] Community skill repository

### v0.4.0 — Pair Programming
- [ ] Pair programming mode (two-model debate)
- [ ] Configurable debate strategies
- [ ] Quality scoring for solutions
- [ ] Auto-selection of critic model

### v0.5.0 — Enterprise
- [ ] Team configuration sharing
- [ ] Audit logging
- [ ] SSO integration
- [ ] Custom model endpoints
- [ ] Usage dashboards

### v1.0.0 — Stable Release
- [ ] API stability guarantee
- [ ] Comprehensive documentation
- [ ] Performance benchmarks vs. alternatives
- [ ] Cross-platform installers (brew, apt, scoop, snap)

## Contributing

AllieCode is open source and welcomes contributions. See [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

Key areas where help is wanted:
- **Provider implementations**: Add support for more LLM providers
- **Tools**: Build new tools or improve existing ones
- **Skills**: Create and share community skills
- **Plugins**: Build WASM plugins
- **Documentation**: Improve docs and examples
- **Testing**: Write tests and improve coverage
