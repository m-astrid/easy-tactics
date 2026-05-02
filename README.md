# ⚡ Exploit/Knockout/Riposte/Outplay/Easy Tactics/? - use their moves against them!

AI-powered system for analyzing HEMA (Historical European Martial Arts) fencers.

## Overview

Easy Tactics helps coaches and fighters analyze technique by:
- Searching fighters on Hemagon (fencing database)
- Parsing tournaments and fight results
- Finding videos on VK/YouTube
- Analyzing fighter technique using LLM
- Storing summaries in markdown files

## Architecture

```
Bot (Go) ←────── gRPC ──────▶ API Service (Go)
                                  │
                                  ▼ HTTP
                            Python AI Service
                                  │
                                  ▼ HTTP
                            MCP Servers (Hemagon, YouTube)
```

## Tech Stack

- **API**: Go with gRPC/HTTP
- **AI Service**: Python (FastAPI)
- **Database**: SQLite
- **Bot**: Telegram Bot API
- **Container**: Docker Compose

## Services

| Service | Port | Description |
|---------|------|-------------|
| API | 50051 | gRPC API server |
| Python AI | 8000 | LLM analysis service |
| Hemagon MCP | 8080 | Hemagon data scraper |

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Telegram Bot Token
- Anthropic API Key (for LLM)

### Setup

1. Clone and configure:
```bash
cp .env.example .env
# Edit .env with your credentials
```

2. Build and start services:
```bash
make build
make up
```

3. Set owner (optional):
```bash
# In .env set:
OWNER_TELEGRAM_ID=your_telegram_id
```

## Project Structure

```
easy-tactics/
├── api/                     # Go API service
│   ├── handlers/            # gRPC handlers (TDD)
│   ├── domain/              # Domain entities
│   ├── storage/             # Database adapters
│   ├── proto/               # Protocol buffers
│   └── main.go
├── migrations/              # Database migrations (goose)
├── docker/                  # Dockerfiles
├── docs/                   # Architecture docs
└── docker-compose.yml
```

## User Roles

| Role | Description |
|------|-------------|
| `owner` | Full access, user management |
| `admin` | User management, analysis access |
| `coach` | Add observations, view profiles |
| `fighter` | View own profile |
| `blocked` | Access denied |

## API Services

- **AuthService**: User management, access control
- **FighterService**: Fighter CRUD, search
- **TournamentService**: Tournament operations
- **FightService**: Fight data, video search
- **AnalysisService**: Summaries, observations

## Commands

### Prerequisites (for local development)
```bash
# Install protobuf compiler
brew install protobuf

# Install Go protobuf plugins
GOPROXY=direct go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
GOPROXY=direct go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

### Code Generation
```bash
make generate     # Generate protobuf code
```

### Deployment
```bash
make build        # Build Docker images
make up           # Start all services
make down         # Stop all services
make restart      # Restart all services
make logs         # View logs (follow)
make clean        # Remove containers and volumes
```

### Testing
```bash
make test         # Run tests
make test-verbose # Run tests with verbose output
```

### Database
```bash
make migrate      # Run database migrations
make db-status    # Show migration status
make db-reset     # Reset database (deletes all data)
```

## License

MIT
