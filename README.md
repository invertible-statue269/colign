# Colign

[![CI](https://github.com/gobenpark/colign/actions/workflows/ci.yml/badge.svg)](https://github.com/gobenpark/colign/actions/workflows/ci.yml)
[![License: AGPL-3.0](https://img.shields.io/badge/License-AGPL--3.0-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/Go-1.26-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Next.js](https://img.shields.io/badge/Next.js-16-000000?logo=next.js&logoColor=white)](https://nextjs.org)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)

A Spec-Driven Development (SDD) workflow platform where developers and non-developers collaboratively discuss and write specs with AI.

## Architecture

```
┌──────────────────┐         ┌──────────────────────┐
│    Next.js 16     │ Connect │     Go (net/http)     │
│    (Frontend)     │◄──────►│     (API Server)      │
│                   │ (.proto)│                       │
│  - React 19       │        │  - uptrace/bun (ORM)  │
│  - Tiptap + Y.js  │        │  - connectrpc/connect │
│  - shadcn/ui      │        │  - JWT + OAuth2       │
└──────────────────┘         └───────┬──────────────┘
                                     │
  ┌──────────────────┐              │
  │  Hocuspocus       │  Y.js       │
  │  (Node sidecar)   │◄────────────┘
  └──────────────────┘
                              ┌──────────┬──────────┐
                              │PostgreSQL│  Redis   │
                              └──────────┴──────────┘
```

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Frontend | Next.js 16, React 19, TypeScript, Tailwind CSS 4, shadcn/ui |
| Editor | Tiptap (ProseMirror) + Y.js (CRDT) |
| API | Connect RPC (buf.build) - gRPC-compatible with JSON support |
| Backend | Go, net/http, uptrace/bun |
| Auth | JWT + OAuth2 (GitHub, Google) |
| Realtime | Hocuspocus (Y.js server) |
| AI | Claude API (streaming), MCP Server |
| Database | PostgreSQL 16, Redis 7 |

## Prerequisites

- Go 1.26+
- Node.js 20+
- Docker & Docker Compose
- [buf](https://buf.build/docs/installation)

## Getting Started

```bash
# Start all services (API + DB + Redis)
docker-compose up --build

# Run frontend (separate terminal)
cd web && npm install && npm run dev
```

Open http://localhost:3000 to access the app.

## Development

```bash
# Generate proto (Go + TypeScript)
cd proto && buf generate

# Run API server locally
go run ./cmd/api

# Run tests
go test ./...
```

## Project Structure

```
.
├── cmd/
│   ├── api/            # API server entrypoint
│   └── mcp/            # MCP server entrypoint
├── internal/
│   ├── auth/           # Authentication (JWT, OAuth, Connect handlers)
│   ├── project/        # Project & Change management
│   ├── organization/   # Organization (workspace) management
│   ├── workflow/       # Workflow engine (state machine)
│   ├── models/         # Database models (bun)
│   ├── config/         # Configuration
│   ├── middleware/     # HTTP middleware (CORS, JWT)
│   ├── server/         # Server setup (net/http + Connect RPC)
│   └── database/       # Database connection
├── proto/              # Protobuf definitions (buf.build)
│   ├── auth/           # Auth service
│   ├── project/        # Project service
│   ├── organization/   # Organization service
│   └── workflow/       # Workflow service
├── gen/                # Generated Go proto code
├── migrations/         # SQL migrations (golang-migrate)
├── web/                # Next.js frontend
├── hocuspocus/         # Y.js collaboration server
├── docker-compose.yml
└── Dockerfile
```

## License

This project is licensed under the [GNU Affero General Public License v3.0](LICENSE) (AGPL-3.0).
