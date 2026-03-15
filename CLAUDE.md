# CLAUDE.md

This file provides guidance to Claude Code when working with
code in this repository.

OwnCord is a self-hosted Windows chat platform with two
components: a Go server (`chatserver.exe`) and a Tauri v2
desktop client. The server is implemented. The client is
being migrated from WPF/.NET 8 to Tauri v2
(Rust + TypeScript).

## Reference Files (read before implementing)

- **CHATSERVER.md** -- Master spec: phases, tasks, security
  priorities, Windows-specific details.
- **PROTOCOL.md** -- WebSocket message format. Every message
  type, payload shape, and rate limit. Server and client
  must agree on this exactly.
- **SCHEMA.md** -- SQLite table definitions, indexes, FTS5
  setup, permission bitfield definitions.
- **API.md** -- REST endpoints, request/response shapes,
  error codes. All paths start with `/api/` (NOT `/api/v1/`).
- **SETUP.md** -- Tooling requirements for both server and
  client development.
- **CLIENT-ARCHITECTURE.md** -- Tauri v2 client project
  structure, component map, store design, and conventions.
- **MIGRATION-PLAN.md** -- Detailed phase-by-phase TODO list
  for the WPF-to-Tauri migration.
- **TESTING-STRATEGY.md** -- Test infrastructure, coverage
  targets, and patterns for every test type.
- **AUDIT.md** -- Known issues found in project audit
  (2026-03-15). All Critical/High items must be fixed.
- **LANGUAGE-REVIEW.md** -- Framework assessment that led to
  the Tauri v2 decision.

## Project Structure

```text
OwnCord/
├── Server/                  # Go server (implemented)
│   ├── config/
│   ├── db/
│   ├── auth/
│   ├── api/
│   ├── ws/
│   ├── admin/static/
│   └── migrations/
├── Client/
│   ├── tauri-client/        # NEW: Tauri v2 client
│   │   ├── src-tauri/       #   Rust backend
│   │   │   └── src/
│   │   ├── src/             #   TypeScript frontend
│   │   │   ├── lib/         #     Core services
│   │   │   ├── stores/      #     Reactive state
│   │   │   ├── components/  #     UI components
│   │   │   ├── pages/       #     Page layouts
│   │   │   └── styles/      #     CSS (from mockups)
│   │   └── tests/
│   │       ├── unit/
│   │       ├── integration/
│   │       └── e2e/
│   ├── OwnCord.Client/     # LEGACY: WPF client (reference)
│   └── ui-mockup.html      # Design source of truth
└── docs/
```

## Build Commands

### Server (Go)

```bash
cd Server
go build -o chatserver.exe -ldflags "-s -w -X main.version=1.0.0" .
go test ./...                        # all tests
go test ./... -cover                 # with coverage
```

### Client (Tauri v2)

```bash
cd Client/tauri-client

# Development (hot reload)
npm run tauri dev

# Build release
npm run tauri build

# Run tests
npm test                             # all tests (vitest)
npm run test:unit                    # unit tests only
npm run test:integration             # integration tests
npm run test:e2e                     # Playwright E2E tests
npm run test:coverage                # with coverage report
```

### Legacy WPF Client (reference only)

```bash
cd Client
dotnet build OwnCord.Client/OwnCord.Client.csproj
dotnet test OwnCord.Client.Tests/
```

### Dev Tools

```bash
# Server
go install github.com/air-verse/air@latest
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# Client (installed via npm)
# vitest, playwright, typescript, vite — all in package.json
```

## Branch Strategy

- `main` -- stable releases
- `dev` -- current development (WPF client)
- `tauri-migration` -- Tauri v2 client migration

## Client Conventions (Tauri v2)

### TypeScript Frontend

- Strict TypeScript: `strict: true`,
  `noUncheckedIndexedAccess: true`
- Immutable state updates: never mutate, always create
  new objects
- Discriminated unions for all WS message types
- CSS custom properties from mockup `:root` block are the
  single source of truth for colors/spacing
- Component CSS files scoped per component
- No framework (vanilla TS + DOM) unless explicitly decided
  otherwise
- Path aliases: `@lib/`, `@stores/`, `@components/`

### Rust Backend (src-tauri)

- Minimal Rust: only for native APIs that the webview
  cannot access
- Tauri IPC commands for: credential storage, system tray,
  global hotkeys
- All FFI calls wrapped in `Result` with proper error
  handling
- Use `tauri-plugin-*` crates where available before
  writing custom code

### Critical Rules

- **API paths**: Always `/api/v1/*` (matches server router)
- **WS field names**: `threshold_mode` NOT `mode` in
  VoiceConfig and VoiceSpeakers payloads
- **Roles**: Always use role NAME strings ("admin",
  "member"), never numeric role\_id in UI-facing code
- **Rate limiting**: Client must respect PROTOCOL.md
  limits (typing 1/3s, presence 1/10s, voice 20/s)
- **Status values**: Only `online`, `idle`, `dnd`,
  `offline`. Never `invisible`.

## Server Conventions (Go)

- Use standard library where possible. Minimize
  dependencies.
- Router: `chi`. SQLite: `modernc.org/sqlite` (pure Go).
  WebSocket: `nhooyr.io/websocket`. WebRTC: `pion/webrtc`
  - `pion/turn`.
- Config via `config.yaml`; environment variable overrides.
- Structured logging via `log/slog`.
- Errors as JSON `{ "error": "CODE", "message": "detail" }`.
- All input sanitized with `bluemonday`.
- bcrypt cost 12+. Server-side session tokens in SQLite.
- Permission bitfield checks on every handler (SCHEMA.md).
- File uploads: validate magic bytes, reject executables,
  strip EXIF, UUID filenames.
- Target: `GOOS=windows GOARCH=amd64`.

## Security Rules

- Never trust client input -- all validation server-side.
- Never log passwords, tokens, or message content.
- Never expose upload directory directly.
- Never reveal whether a username exists on failed login.
- Rate limit everything: logins, messages, uploads, API.
- WebSocket connections must authenticate before any data.
- TLS on by default (self-signed generated on first run).
- Invite-only registration.
- Client stores tokens in Windows Credential Manager only.

## Testing Requirements

- **Coverage target**: 80%+ for all code
- **TDD workflow**: Write tests first (RED), implement
  (GREEN), refactor (IMPROVE)
- **Unit tests**: Every service, store, utility function
- **Integration tests**: Full WS message flows with mocked
  transport
- **E2E tests**: Login flow, chat send/receive, channel
  switching
- See TESTING-STRATEGY.md for full details.
