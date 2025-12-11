# CodeSage Architecture & Flow

This document explains how the codebase is organized and how a GitHub Pull Request event flows through CodeSage from webhook to AI review comment.

## High-Level Overview

- A Gin web server exposes endpoints, notably `POST /github/webhook`.
- The webhook handler verifies signatures, parses the payload, and reacts to PR events (`opened`, `synchronize`).
- It fetches changed files from the PR via GitHub API, constructs a combined diff, sends it to Gemini for analysis, and posts the result back as a PR comment.

## Module Map

- `main.go`
  - Loads configuration via `config.Load()`
  - Initializes the Gin router with `server.SetupRouter(cfg)`
  - Starts the HTTP server on `cfg.Port`

- `server/router.go`
  - Registers routes:
    - `POST /github/webhook` → `github.HandleWebhook`
    - `GET /` → health check
    - `GET /auth/github/login` and `/auth/github/callback` → placeholders

- `config/config.go`
  - Loads environment variables and `.env` via `godotenv`
  - Provides fields like `Port`, `GitHubToken`, `GeminiKey`, `GITHUB_WEBHOOK_SECRET`, etc.
  - Helper `getEnv(key, fallback)` for defaults

- `github/webhook.go`
  - Entry point for GitHub webhook flow: `HandleWebhook(c, cfg)`
  - Detects payload format (raw JSON or `payload=` form-encoded) and parses JSON
  - Verifies `X-Hub-Signature-256` using `github.VerifyWebhookSignature`
  - Validates action is `opened` or `synchronize`
  - Extracts repository, PR number, title, user
  - If installation context exists, exchanges a short-lived installation token via `GetInstallationToken` and temporarily replaces `cfg.GitHubToken`
  - Fetches PR files with `github.GetPRFiles`
  - Builds a combined diff and counts lines for logging
  - Calls `ai.AnalyzeWithGemini(diff, title)`
  - Formats the analysis into a Markdown comment and posts via `github.PostComment`

- `github/api.go`
  - `GetPRFiles(owner, repo, prNumber, cfg)` — GET `repos/{owner}/{repo}/pulls/{pr}/files`
  - `PostComment(owner, repo, prNumber, comment, cfg)` — POST `repos/{owner}/{repo}/issues/{pr}/comments`

- `github/app.go`
  - `VerifyWebhookSignature(secret, body, signatureHeader)` — HMAC SHA-256 check against `X-Hub-Signature-256`
  - `GetInstallationToken(cfg, installationID)` — Exchanges an App JWT for an installation access token
  - `generateAppJWT(cfg)` — Creates a short-lived JWT, signed with `GITHUB_APP_PRIVATE_KEY`

- `ai/gemini.go`
  - `AnalyzeWithGemini(diff, title)` — Calls Gemini `generateContent` API with a review prompt; requires `GEMINI_API_KEY`
  - Truncates large diffs to stay within API limits

- `ai/huggingface.go`
  - `AnalyzeWithHFCodeReview(diff, title)` — Optional integration; not currently used in the webhook path

- `utils/logger.go`
  - Minimal wrappers: `Infof`, `Errorf` (currently not widely used)

## Request Flow Details

1. Startup
   - `main.go` calls `config.Load()` and `server.SetupRouter(cfg)`
   - Gin registers endpoints and starts listening on `cfg.Port` (default `8080`)

2. Webhook Receipt (`POST /github/webhook`)
   - Logs receipt and inspects `X-GitHub-Event` to route handling
   - For `pull_request` events, `handlePullRequest` reads the raw body and verifies signature
   - Supports both raw JSON and `application/x-www-form-urlencoded` (`payload=`) formats

3. Validation & Extraction
   - Ensures `action` and `pull_request` fields exist and that action is `opened` or `synchronize`
   - Extracts PR number, title, user, repo/owner

4. GitHub Context & Tokens
   - If `installation.id` is present, exchanges an installation token and temporarily sets `cfg.GitHubToken` for subsequent API calls

5. Fetch Changes
   - `GetPRFiles` retrieves changed files and patches; builds a combined diff string

6. AI Analysis
   - Sends `diff` and `title` to `AnalyzeWithGemini`
   - Receives a concise review

7. Post Comment
   - Formats the analysis as Markdown and posts to the PR via `PostComment`

8. Response
   - Returns a JSON success response to the webhook sender

## Error Handling & Logging

- Signature verification failure → `401` with `invalid signature`
- JSON parsing or missing fields → `400` with details
- Upstream API failures (GitHub, Gemini) → `500` with message
- Non-actionable webhook events (e.g., `push`) → `200` with a message indicating they’re ignored

## Extending CodeSage

- Replace or augment `ai.AnalyzeWithGemini` with other providers
- Wire `ai.AnalyzeWithHFCodeReview` into the webhook flow if preferred
- Implement real OAuth at `/auth/github/*` endpoints
- Introduce structured logging or middleware using `utils/logger` or a logging library