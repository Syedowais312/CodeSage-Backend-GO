# CodeSage

AI-powered pull request reviewer built with Go and Gin. CodeSage listens to GitHub PR webhooks, fetches the diff, analyzes changes with Gemini, and posts a friendly review comment back to the PR.

## Features

- Receives GitHub webhook events for pull requests
- Verifies webhook signatures (`X-Hub-Signature-256`)
- Fetches changed files via GitHub API
- Sends diffs to Gemini for analysis
- Posts a formatted review comment back to the PR
- Simple health endpoint (`GET /`)

## Requirements

- Go (as specified in `go.mod`)
- A GitHub webhook configured for your repository
- Environment variables for GitHub and AI providers

## Quick Start

1. Create a `.env` file in the project root with values relevant to your setup:

```
PORT=8080
GITHUB_TOKEN=ghp_...
GEMINI_API_KEY=...
HF_API_KEY=...               # optional, used by huggingface.go
GITHUB_APP_ID=...            # optional, used for installation tokens
GITHUB_APP_PRIVATE_KEY=...   # optional, PEM format
GITHUB_WEBHOOK_SECRET=...    # required to verify webhook signatures
GITHUB_OAUTH_CLIENT_ID=...   # optional, if you enable OAuth endpoints
GITHUB_OAUTH_CLIENT_SECRET=...
```

2. Build the project:

```
go build ./...
```

3. Run the server:

```
go run main.go
```

You should see logs indicating the server is listening on `:8080` and the registered routes.

## Endpoints

- `POST /github/webhook` — GitHub webhook receiver (expects PR events). Supports raw JSON and `application/x-www-form-urlencoded` payloads (`payload=` format) and verifies `X-Hub-Signature-256` using `GITHUB_WEBHOOK_SECRET`.
- `GET /` — Health check returning `{ "message": "CodeSage is running" }`.
- `GET /auth/github/login` — Placeholder endpoint for GitHub OAuth login.
- `GET /auth/github/callback` — Placeholder endpoint for GitHub OAuth callback.

## GitHub Webhook Setup

- In your GitHub repo settings, add a webhook:
  - Payload URL: `http://<your-host>/github/webhook`
  - Content type: `application/json`
  - Secret: set to the value of `GITHUB_WEBHOOK_SECRET`
  - Events: enable “Pull requests” (and “Ping” for testing)

When a PR is opened or synchronized, CodeSage:

1. Verifies the request signature
2. Reads and parses the payload
3. Fetches changed files from the PR
4. Builds a combined diff
5. Sends the diff and title to Gemini
6. Posts a formatted comment back to the PR

## Configuration Reference

Configuration is loaded from environment variables (with `.env` support) via `config/`:

- `PORT` — Server port, default `8080`
- `GITHUB_TOKEN` — Token used for GitHub API calls
- `GEMINI_API_KEY` — Required to analyze with Gemini
- `HF_API_KEY` — Optional; used by Hugging Face integration
- `GITHUB_APP_ID`, `GITHUB_APP_PRIVATE_KEY` — Optional; used to exchange installation tokens for GitHub App scenarios
- `GITHUB_WEBHOOK_SECRET` — Required to verify webhook signatures
- `GITHUB_OAUTH_CLIENT_ID`, `GITHUB_OAUTH_CLIENT_SECRET` — Optional; for OAuth endpoints

## Project Structure

- `main.go` — Entry point that loads config and starts the Gin server
- `server/router.go` — Router setup and route registration
- `config/config.go` — Environment configuration loader
- `github/webhook.go` — Webhook handler and PR flow logic
- `github/api.go` — GitHub API calls (PR files, comments)
- `github/app.go` — GitHub App helpers (signature verification, installation tokens)
- `ai/gemini.go` — Gemini integration
- `ai/huggingface.go` — Optional Hugging Face integration
- `utils/logger.go` — Minimal logger helpers

## Notes

- The AI analysis currently uses Gemini (`ai/gemini.go`). The Hugging Face helper is available but not wired in by default.
- Ensure your tokens have appropriate scopes to read PR files and post comments.
- For production, consider setting `GIN_MODE=release` and configuring trusted proxies for Gin.