# Using Factory CLI (Droid) and Amp CLI with CLIProxyAPI

## ⚠️ Important Update

**This fork has been merged upstream!** All Amp CLI integration features developed in this fork have been accepted and merged into the official [router-for-me/CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI) repository.

**Please use the upstream repository for the latest features, updates, and support:**

👉 **[github.com/router-for-me/CLIProxyAPI](https://github.com/router-for-me/CLIProxyAPI)**

This document is maintained solely for legacy link preservation from previous social media posts and shared documentation.

---

## Official Documentation

### Amp CLI Integration

For complete instructions on using Amp CLI with CLIProxyAPI, see the official documentation:

📖 **[Amp CLI Integration Guide](https://github.com/router-for-me/CLIProxyAPI/blob/main/docs/amp-cli-integration.md)**

This guide covers:
- OAuth setup for Gemini Pro/Ultra, ChatGPT Plus/Pro, and Claude Pro/Max subscriptions
- Configuration for Amp CLI and Amp IDE extensions
- Provider routing and management endpoints
- Troubleshooting and best practices

### Factory CLI (Droid) Integration

For instructions on using Factory AI's Droid CLI with CLIProxyAPI, see:

📖 **[Factory Droid Documentation](https://help.router-for.me/agent-client/droid.html)**

---

## Quick Reference: Factory CLI Custom Models

For quick reference, here's an example `~/.factory/config.json` configuration for using CLIProxyAPI with Factory CLI:

```json
{
  "custom_models": [
    {
      "model_display_name": "Claude Haiku 4.5 [Proxy]",
      "model": "claude-haiku-4-5-20251001",
      "base_url": "http://localhost:8317",
      "api_key": "dummy-not-used",
      "provider": "anthropic"
    },
    {
      "model_display_name": "Claude Sonnet 4.5 [Proxy]",
      "model": "claude-sonnet-4-5-20250929",
      "base_url": "http://localhost:8317",
      "api_key": "dummy-not-used",
      "provider": "anthropic"
    },
    {
      "model_display_name": "Claude Opus 4.1 [Proxy]",
      "model": "claude-opus-4-1-20250805",
      "base_url": "http://localhost:8317",
      "api_key": "dummy-not-used",
      "provider": "anthropic"
    },
    {
      "model_display_name": "GPT-5.1 Low [Proxy]",
      "model": "gpt-5.1-low",
      "base_url": "http://localhost:8317/v1",
      "api_key": "dummy-not-used",
      "provider": "openai"
    },
    {
      "model_display_name": "GPT-5.1 Medium [Proxy]",
      "model": "gpt-5.1-medium",
      "base_url": "http://localhost:8317/v1",
      "api_key": "dummy-not-used",
      "provider": "openai"
    },
    {
      "model_display_name": "GPT-5.1 High [Proxy]",
      "model": "gpt-5.1-high",
      "base_url": "http://localhost:8317/v1",
      "api_key": "dummy-not-used",
      "provider": "openai"
    },
    {
      "model_display_name": "GPT-5.1 Codex Low [Proxy]",
      "model": "gpt-5.1-codex-low",
      "base_url": "http://localhost:8317/v1",
      "api_key": "dummy-not-used",
      "provider": "openai"
    },
    {
      "model_display_name": "GPT-5.1 Codex Medium [Proxy]",
      "model": "gpt-5.1-codex-medium",
      "base_url": "http://localhost:8317/v1",
      "api_key": "dummy-not-used",
      "provider": "openai"
    },
    {
      "model_display_name": "GPT-5.1 Codex High [Proxy]",
      "model": "gpt-5.1-codex-high",
      "base_url": "http://localhost:8317/v1",
      "api_key": "dummy-not-used",
      "provider": "openai"
    },
    {
      "model_display_name": "GPT-5.1 Codex Mini Medium [Proxy]",
      "model": "gpt-5.1-codex-mini-medium",
      "base_url": "http://localhost:8317/v1",
      "api_key": "dummy-not-used",
      "provider": "openai"
    },
    {
      "model_display_name": "GPT-5.1 Codex Mini High [Proxy]",
      "model": "gpt-5.1-codex-mini-high",
      "base_url": "http://localhost:8317/v1",
      "api_key": "dummy-not-used",
      "provider": "openai"
    },
    {
      "model_display_name": "Gemini 3 Pro Preview [Proxy]",
      "model": "gemini-3-pro-preview",
      "base_url": "http://localhost:8317/v1",
      "api_key": "dummy-not-used",
      "provider": "openai"
    }
  ]
}
```

### Key Points

- **`base_url`**: Use `http://localhost:8317` for Anthropic models, `http://localhost:8317/v1` for OpenAI/generic models
- **`api_key`**: Use `"dummy-not-used"` when OAuth is configured via CLIProxyAPI
- **`provider`**: Set to `"anthropic"` for Claude models, `"openai"` for GPT/Gemini models

---

## Installation

Install the official CLIProxyAPI from the upstream repository:

```bash
git clone https://github.com/router-for-me/CLIProxyAPI.git
cd CLIProxyAPI
go build -o cli-proxy-api ./cmd/server
```

Or via Homebrew (macOS/Linux):

```bash
brew install cliproxyapi
brew services start cliproxyapi
```
