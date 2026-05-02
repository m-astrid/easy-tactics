# Easy Tactics

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

2. Start services:
```bash
docker-compose up --build
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

## Development

### Run Tests
```bash
cd api && go test ./...
```

### Run Migrations
```bash
goose -dir migrations sqlite3 /data/fighters.db up
```

## License

MIT