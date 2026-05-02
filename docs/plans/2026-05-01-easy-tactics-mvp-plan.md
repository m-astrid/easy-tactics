# Easy Tactics MVP Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Создать MVP системы для анализа техники фехтовальщиков — API сервис + Telegram бот + Python AI Service с MCP

**Architecture:** Docker-compose с 5 сервисами: Telegram Bot (thin wrapper), API Service (Go gRPC), Python AI Service, MCP серверы (Hemagon — свой, VK — свой), SQLite

**Tech Stack:** Go (Fiber), Python (FastAPI), SQLite, Docker, Telegram Bot API, gRPC, MCP

---

## Компоненты системы

```
easy-tactics/
├── docker-compose.yml
├── api/                    # Go gRPC API Service
├── bot/                    # Telegram Bot (Go)
├── python/
│   ├── ai-service/         # Python AI Service
│   └── mcp/
│       ├── hemagon/        # Hemagon MCP Server
│       └── vk/             # VK Video MCP Server
└── data/                   # SQLite volume
```

---

## Phase 0: Auth System (Роли и пользователи)

### Task 0.1: Auth Service в gRPC

**Files:**
- Modify: `api/internal/handlers/auth.go` (new)
- Modify: `api/internal/db/auth.go` (new)

- [ ] **Step 1: Add users table migration**

```sql
CREATE TABLE IF NOT EXISTS users (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  telegram_id BIGINT UNIQUE NOT NULL,
  username TEXT,
  full_name TEXT,
  role TEXT DEFAULT 'fighter',
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
CREATE INDEX idx_users_telegram_id ON users(telegram_id);
CREATE INDEX idx_users_role ON users(role);
```

- [ ] **Step 1.5: Add owner creation from env**

```go
func initOwner(db *sql.DB) {
    ownerID := os.Getenv("OWNER_TELEGRAM_ID")
    if ownerID == "" {
        return
    }
    
    telegramID, err := strconv.ParseInt(ownerID, 10, 64)
    if err != nil {
        log.Printf("Invalid OWNER_TELEGRAM_ID: %v", err)
        return
    }
    
    // Check if owner already exists
    var count int
    err = db.QueryRow("SELECT COUNT(*) FROM users WHERE role = 'owner'").Scan(&count)
    if err != nil || count > 0 {
        return // Owner already exists
    }
    
    // Create owner
    _, err = db.Exec(`
        INSERT INTO users (telegram_id, username, full_name, role)
        VALUES (?, 'owner', 'Owner', 'owner')
    `, telegramID)
    if err != nil {
        log.Printf("Failed to create owner: %v", err)
    } else {
        log.Printf("Created owner with telegram_id: %d", telegramID)
    }
}
```

- [ ] **Step 2: Create auth handler**

```go
package handlers

type AuthHandler struct {
    db *sql.DB
}

func NewAuthHandler(db *sql.DB) *AuthHandler {
    return &AuthHandler{db: db}
}

func (h *AuthHandler) AddUser(ctx context.Context, req *fapiv1.AddUserRequest) (*fapiv1.User, error) {
    // Insert user into database
    // Return created user
}

func (h *AuthHandler) CheckAccess(ctx context.Context, req *fapiv1.CheckAccessRequest) (*fapiv1.AccessResponse, error) {
    // Check if telegram_id exists and role != blocked
    // Return allowed = true/false
}
```

- [ ] **Step 3: Commit**

```bash
git add api/internal/handlers/auth.go
git commit -m "feat(auth): add auth service"
```

---

### Task 0.2: Bot Middleware для проверки прав

**Files:**
- Modify: `bot/middleware/auth.go` (new)
- Modify: `bot/main.go`

- [ ] **Step 1: Create auth middleware**

```go
package middleware

type AuthMiddleware struct {
    client *grpc.Client
}

func NewAuthMiddleware(client *grpc.Client) *AuthMiddleware {
    return &AuthMiddleware{client: client}
}

func (m *AuthMiddleware) CheckAccess(telegramID int64, requiredRole string) (bool, string) {
    resp, err := m.client.CheckAccess(context.Background(), &fapiv1.CheckAccessRequest{
        TelegramId:    telegramID,
        RequiredRole:  requiredRole,
    })
    if err != nil {
        return false, "Ошибка проверки доступа"
    }
    return resp.Allowed, resp.Reason
}
```

- [ ] **Step 2: Add middleware to bot commands**

```go
func handleMessage(update tb.Update) {
    telegramID := update.Message.From.ID
    
    // Check access for all commands
    allowed, reason := authMiddleware.CheckAccess(telegramID, "fighter")
    if !allowed {
        msg := tb.NewMessage(update.Message.Chat.ID, reason)
        bot.Send(msg)
        return
    }
    
    // Handle commands...
}
```

- [ ] **Step 3: Add admin commands**

```go
func handleAdminCommand(update tb.Update, command string) {
    allowed, _ := authMiddleware.CheckAccess(telegramID, "admin")
    if !allowed {
        bot.Send(tb.NewMessage(chatID, "Нет доступа"))
        return
    }
    
    switch command {
    case "/add_user":
        // Parse: /add_user @username role
    case "/remove_user":
        // Parse: /remove_user @username
    case "/list_users":
        // List all users
    case "/set_role":
        // Parse: /set_role @username new_role
    }
}
```

- [ ] **Step 4: Commit**

```bash
git add bot/middleware/
git commit -m "feat(bot): add auth middleware"
```

---

## Phase 1: Infrastructure

### Task 1: Docker Compose Base

**Files:**
- Create: `docker-compose.yml`
- Create: `docker/api/Dockerfile`
- Create: `docker/bot/Dockerfile`
- Create: `docker/python/ai-service/Dockerfile`
- Create: `docker/python/mcp/hemagon/Dockerfile`

- [ ] **Step 1: Create docker-compose.yml with all services**

```yaml
version: '3.8'

services:
  api:
    build: ./docker/api
    ports:
      - "50051:50051"
    volumes:
      - ./data:/data
    environment:
      - OWNER_TELEGRAM_ID=${OWNER_TELEGRAM_ID}
      - DB_PATH=/data/fighters.db
      - AI_SERVICE_URL=http://python-ai:8000
      - HEMAGON_MCP_URL=http://hemagon-mcp:8080
    depends_on:
      - python-ai

  bot:
    build: ./docker/bot
    environment:
      - API_ADDR=api:50051
      - TELEGRAM_TOKEN=${TELEGRAM_TOKEN}
    depends_on:
      - api

  python-ai:
    build: ./docker/python/ai-service
    ports:
      - "8000:8000"
    environment:
      - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
      - OPENAI_API_KEY=${OPENAI_API_KEY}
      - HEMAGON_MCP_URL=http://hemagon-mcp:8080

  hemagon-mcp:
    build: ./docker/python/mcp/hemagon
    ports:
      - "8081:8080"

  # YouTube MCP (placeholder - add when ready)
  # youtube-mcp:
  #   image: mcp-youtube

volumes:
  data:

networks:
  default:
    name: easy-tactics-network
```

- [ ] **Step 2: Create minimal Dockerfiles**

`docker/api/Dockerfile`:
```dockerfile
FROM golang:1.21-alpine
WORKDIR /app
COPY api/ .
RUN go build -o server .
EXPOSE 50051
CMD ["./server"]
```

`docker/bot/Dockerfile`:
```dockerfile
FROM golang:1.21-alpine
WORKDIR /app
COPY bot/ .
RUN go build -o bot .
CMD ["./bot"]
```

`docker/python/ai-service/Dockerfile`:
```dockerfile
FROM python:3.11
WORKDIR /app
COPY python/ai-service/ .
RUN pip install -r requirements.txt
EXPOSE 8000
CMD ["uvicorn", "main:app", "--host", "0.0.0.0"]
```

`docker/python/mcp/hemagon/Dockerfile`:
```dockerfile
FROM python:3.11
WORKDIR /app
COPY python/mcp/hemagon/ .
RUN pip install -r requirements.txt
EXPOSE 8080
CMD ["uvicorn", "server:app", "--host", "0.0.0.0"]
```

- [ ] **Step 3: Commit**

```bash
git add docker-compose.yml docker/
git commit -m "feat: add docker infrastructure base"
```

---

## Phase 2: API Service (Go) — TDD

### Task 2: Go gRPC Server с тестами

**Files:**
- Create: `api/main.go`
- Create: `api/proto/fighter-agent.proto` (copy from docs)
- Create: `api/internal/config/config.go`
- Create: `api/internal/db/db.go`
- Create: `api/internal/handlers/fighter.go`
- Create: `api/internal/handlers/fighter_test.go` (TDD)
- Create: `api/internal/db/db_test.go` (TDD)

---

#### TDD Approach для API Service

Каждый handler тестируется ДО реализации:
1. Пишем тест → ожидаем FAIL
2. Запускаем тест → убеждаемся FAIL
3. Пишем минимальный код → тест PASS
4. Рефакторим

---

- [ ] **Step 1: Setup Go project + тестовые зависимости**

```bash
cd api
go mod init github.com/easy-tactics/api

# Production deps
go get google.golang.org/grpc
go get google.golang.org/protobuf
go get github.com/mattn/go-sqlite3

# Test deps
go get github.com/stretchr/testify
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

- [ ] **Step 2: Copy proto file and generate code**

Copy `fighter-agent.proto` to `api/proto/`

```bash
cd api/proto
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       fighter-agent.proto
```

- [ ] **Step 3: TDD - Fighter Handler Tests**

**Сначала тест (должен упасть):**

```go
// api/internal/handlers/fighter_test.go
package handlers

import (
    "testing"
    "github.com/stretchr/testify/assert"
    fapiv1 "github.com/easy-tactics/api/proto/gen/fighter/agent/v1"
)

func TestSearchFighter_EmptyDB_ReturnsNotFound(t *testing.T) {
    // Arrange
    db, _ := sql.Open("sqlite3", ":memory:")
    handler := NewFighterHandler(db)
    
    // Act
    resp, err := handler.SearchFighter(context.Background(), &fapiv1.SearchFighterRequest{
        Query: "Ivan",
    })
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, fapiv1.Source_SOURCE_NOT_FOUND, resp.Source)
    assert.Len(t, resp.Matches, 0)
}

func TestSearchFighter_WithData_ReturnsMatches(t *testing.T) {
    // Arrange - create test db with fighter
    db, _ := sql.Open("sqlite3", ":memory:")
    db.Exec("INSERT INTO fighters (uuid, slug, full_name) VALUES (?, ?, ?)", 
        "test-uuid", "ivan-petrov-msk", "Иван Петров")
    handler := NewFighterHandler(db)
    
    // Act
    resp, err := handler.SearchFighter(context.Background(), &fapiv1.SearchFighterRequest{
        Query: "Петров",
    })
    
    // Assert
    assert.NoError(t, err)
    assert.Equal(t, fapiv1.Source_SOURCE_LOCAL, resp.Source)
    assert.Len(t, resp.Matches, 1)
}
```

**Запустить тест (должен FAIL):**
```bash
cd api
go test ./internal/handlers/... -v
# Expected: FAIL - undefined: NewFighterHandler
```

- [ ] **Step 4: Minimal implementation (тест PASS)**

```go
// api/internal/handlers/fighter.go
package handlers

import (
    "context"
    "github.com/easy-tactics/api/internal/db"
    fapiv1 "github.com/easy-tactics/api/proto/gen/fighter/agent/v1"
)

type FighterHandler struct {
    db *sql.DB
}

func NewFighterHandler(db *sql.DB) *FighterHandler {
    return &FighterHandler{db: db}
}

func (h *FighterHandler) SearchFighter(ctx context.Context, req *fapiv1.SearchFighterRequest) (*fapiv1.SearchFighterResponse, error) {
    rows, err := h.db.Query(`
        SELECT uuid, slug, full_name, city, club, hemagon_url 
        FROM fighters 
        WHERE full_name LIKE ?
    `, "%"+req.Query+"%")
    if err != nil {
        return &fapiv1.SearchFighterResponse{
            Source: fapiv1.Source_SOURCE_NOT_FOUND,
        }, nil
    }
    defer rows.Close()
    
    var matches []*fapiv1.SearchFighterResponse_FighterMatch
    for rows.Next() {
        var m fapiv1.SearchFighterResponse_FighterMatch
        // Scan fields...
        matches = append(matches, &m)
    }
    
    if len(matches) == 0 {
        return &fapiv1.SearchFighterResponse{
            Source: fapiv1.Source_SOURCE_NOT_FOUND,
        }, nil
    }
    
    return &fapiv1.SearchFighterResponse{
        Source: fapiv1.Source_SOURCE_LOCAL,
        Matches: matches,
    }, nil
}
```

**Запустить тест:**
```bash
go test ./internal/handlers/... -v
# Expected: PASS
```

- [ ] **Step 5: Commit with TDD**

```bash
git add api/
git commit -m "feat(api): add FighterService with TDD

- Added SearchFighter handler
- Added tests: empty db, with data
- Test coverage for main query paths"
```

---

### Task 2.2: Auth Handler с TDD

**Сначала тест:**

```go
// api/internal/handlers/auth_test.go
package handlers

import (
    "testing"
    "github.com/stretchr/testify/assert"
    fapiv1 "github.com/easy-tactics/api/proto/gen/fighter/agent/v1"
)

func TestCheckAccess_BlockedUser_ReturnsFalse(t *testing.T) {
    db, _ := sql.Open("sqlite3", ":memory:")
    db.Exec("INSERT INTO users (telegram_id, role) VALUES (?, ?)", 123, "blocked")
    handler := NewAuthHandler(db)
    
    resp, err := handler.CheckAccess(context.Background(), &fapiv1.CheckAccessRequest{
        TelegramId: 123,
    })
    
    assert.NoError(t, err)
    assert.False(t, resp.Allowed)
    assert.Equal(t, "Пользователь заблокирован", resp.Reason)
}

func TestCheckAccess_NewUser_ReturnsTrue(t *testing.T) {
    db, _ := sql.Open("sqlite3", ":memory:")
    handler := NewAuthHandler(db)
    
    resp, err := handler.CheckAccess(context.Background(), &fapiv1.CheckAccessRequest{
        TelegramId: 456,
    })
    
    assert.NoError(t, err)
    assert.True(t, resp.Allowed) // New user gets fighter role by default
}
```

- [ ] **Step 6: Run auth tests (should FAIL)**

```bash
go test ./internal/handlers/... -run TestCheckAccess -v
# Expected: FAIL - undefined: NewAuthHandler
```

- [ ] **Step 7: Implement AuthHandler**

```go
// api/internal/handlers/auth.go
package handlers

func (h *AuthHandler) CheckAccess(ctx context.Context, req *fapiv1.CheckAccessRequest) (*fapiv1.AccessResponse, error) {
    var role string
    err := h.db.QueryRow(
        "SELECT COALESCE(role, 'fighter') FROM users WHERE telegram_id = ?",
        req.TelegramId,
    ).Scan(&role)
    
    if err == sql.ErrNoRows {
        // New user - allow with default role
        return &fapiv1.AccessResponse{Allowed: true}, nil
    }
    if err != nil {
        return &fapiv1.AccessResponse{Allowed: false, Reason: "Ошибка БД"}, err
    }
    
    if role == "blocked" {
        return &fapiv1.AccessResponse{Allowed: false, Reason: "Пользователь заблокирован"}, nil
    }
    
    return &fapiv1.AccessResponse{Allowed: true}, nil
}
```

- [ ] **Step 8: Run tests (should PASS)**

```bash
go test ./internal/handlers/... -v
```

- [ ] **Step 9: Commit**

```bash
git add api/internal/handlers/
git commit -m "feat(auth): add AuthService with TDD

- Added CheckAccess handler
- Tests: blocked user, new user"
```

- [ ] **Step 4: Create main.go**

```go
package main

import (
    "log"
    "net"

    "github.com/easy-tactics/api/internal/config"
    "github.com/easy-tactics/api/internal/db"
    "github.com/easy-tactics/api/internal/handlers"
    fapiv1 "github.com/easy-tactics/api/proto/gen/fighter/agent/v1"
    "google.golang.org/grpc"
)

func main() {
    cfg := config.Load()
    database, err := db.Connect(cfg.DBPath)
    if err != nil {
        log.Fatalf("failed to connect to db: %v", err)
    }

    lis, err := net.Listen("tcp", ":50051")
    if err != nil {
        log.Fatalf("failed to listen: %v", err)
    }

    grpcServer := grpc.NewServer()
    fighterHandler := handlers.NewFighterHandler(database)

    fapiv1.RegisterFighterServiceServer(grpcServer, fighterHandler)
    
    log.Println("gRPC server listening on :50051")
    if err := grpcServer.Serve(lis); err != nil {
        log.Fatalf("failed to serve: %v", err)
    }
}
```

- [ ] **Step 5: Create config.go**

```go
package config

import "os"

type Config struct {
    DBPath     string
    AIAddress string
}

func Load() Config {
    return Config{
        DBPath:     os.Getenv("DB_PATH"),
        AIAddress: os.Getenv("AI_ADDR"),
    }
}
```

- [ ] **Step 6: Create db.go**

```go
package db

import (
    "database/sql"
    "fmt"

    _ "github.com/mattn/go-sqlite3"
)

func Connect(path string) (*sql.DB, error) {
    db, err := sql.Open("sqlite3", path)
    if err != nil {
        return nil, fmt.Errorf("failed to open db: %w", err)
    }
    return db, nil
}
```

- [ ] **Step 7: Create basic handlers**

```go
package handlers

import (
    "context"

    fapiv1 "github.com/easy-tactics/api/proto/gen/fighter/agent/v1"
)

type FighterHandler struct {
    db *sql.DB
}

func NewFighterHandler(db *sql.DB) *FighterHandler {
    return &FighterHandler{db: db}
}

func (h *FighterHandler) SearchFighter(ctx context.Context, req *fapiv1.SearchFighterRequest) (*fapiv1.SearchFighterResponse, error) {
    // TODO: implement search
    return &fapiv1.SearchFighterResponse{
        Source: fapiv1.Source_SOURCE_NOT_FOUND,
    }, nil
}
```

- [ ] **Step 8: Run and verify**

```bash
cd api
go build -o server .
# Should compile without errors
```

- [ ] **Step 9: Commit**

```bash
git add api/
git commit -m "feat(api): add basic gRPC server structure"
```

---

## Phase 3: Python AI Service

### Task 3: Python AI Service with LLM

**Files:**
- Create: `python/ai-service/main.py`
- Create: `python/ai-service/requirements.txt`
- Create: `python/ai-service/services/llm.py`

- [ ] **Step 1: Create requirements.txt**

```txt
fastapi
httpx
anthropic
python-dotenv
```

- [ ] **Step 2: Create main.py**

```python
from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import httpx
import os

app = FastAPI()

LLM_API_KEY = os.getenv("OPENAI_API_KEY") or os.getenv("ANTHROPIC_API_KEY")
AI_SERVICE_URL = os.getenv("AI_SERVICE_URL", "http://python-ai:8000")

class GenerateSummaryRequest(BaseModel):
    fighter_uuid: str
    fight_uuids: list[str]

@app.post("/generate-summary")
async def generate_summary(req: GenerateSummaryRequest):
    # Call LLM to generate summary
    # TODO: implement
    return {"status": "processing", "task_id": "..."}

@app.get("/health")
async def health():
    return {"status": "ok"}
```

- [ ] **Step 3: Create LLM service**

```python
import anthropic
import os

class LLMService:
    def __init__(self):
        self.client = anthropic.Anthropic(
            api_key=os.getenv("ANTHROPIC_API_KEY")
        )
    
    def generate_summary(self, fighter_data: dict, fights: list[dict]) -> str:
        prompt = f"Проанализируй технику бойца {fighter_data['name']} на основе боев: {fights}"
        response = self.client.messages.create(
            model="claude-3-5-sonnet-20241022",
            max_tokens=4096,
            messages=[{"role": "user", "content": prompt}]
        )
        return response.content[0].text
```

- [ ] **Step 4: Test health endpoint**

```bash
cd python/ai-service
pip install -r requirements.txt
uvicorn main:app --host 0.0.0.0 --port 8000
curl http://localhost:8000/health
```

- [ ] **Step 5: Commit**

```bash
git add python/ai-service/
git commit -m "feat(ai-service): add basic Python AI service"
```

---

## Phase 4: MCP Servers

### Task 4: Hemagon MCP Server

**Files:**
- Create: `python/mcp/hemagon/main.py`
- Create: `python/mcp/hemagon/requirements.txt`
- Create: `python/mcp/hemagon/client.py`

- [ ] **Step 1: Create requirements.txt**

```txt
fastapi
httpx
playwright
beautifulsoup4
```

- [ ] **Step 2: Create MCP server**

```python
from fastapi import FastAPI
from pydantic import BaseModel

app = FastAPI()

class SearchFighterRequest(BaseModel):
    name: str

@app.post("/search_fighter")
async def search_fighter(req: SearchFighterRequest):
    # TODO: implement web scraping of hemagon.ru
    # Using Playwright or requests + BeautifulSoup
    return {"results": []}

@app.get("/health")
async def health():
    return {"status": "ok"}
```

- [ ] **Step 3: Test basic server**

```bash
cd python/mcp/hemagon
pip install -r requirements.txt
uvicorn main:app --host 0.0.0.0 --port 8080
curl http://localhost:8080/health
```

- [ ] **Step 4: Commit**

```bash
git add python/mcp/hemagon/
git commit -m "feat(mcp): add Hemagon MCP server placeholder"
```

---

## Phase 5: Telegram Bot

### Task 5: Go Telegram Bot

**Files:**
- Create: `bot/main.go`
- Create: `bot/config/config.go`
- Create: `bot/grpc/client.go`

- [ ] **Step 1: Initialize Go module**

```bash
cd bot
go mod init github.com/easy-tactics/bot
go get github.com/go-telegram-bot-api/telegram-bot-api/v5
go get google.golang.org/grpc
```

- [ ] **Step 2: Create main.go**

```go
package main

import (
    "log"
    "os"

    tb "github.com/go-telegram-bot-api/telegram-bot-api/v5"
    "github.com/easy-tactics/bot/grpc"
)

func main() {
    token := os.Getenv("TELEGRAM_TOKEN")
    if token == "" {
        log.Fatal("TELEGRAM_TOKEN not set")
    }

    bot, err := tb.NewBotAPI(token)
    if err != nil {
        log.Fatalf("failed to create bot: %v", err)
    }

    client, err := grpc.NewClient(os.Getenv("API_ADDR"))
    if err != nil {
        log.Fatalf("failed to connect to API: %v", err)
    }

    log.Printf("Bot started: %s", bot.Token)

    u := bot.GetUpdatesChan(tb.UpdateConfig{Timeout: 60})

    for update := range u {
        if update.Message == nil {
            continue
        }

        // Simple command handling
        switch update.Message.Command() {
        case "start":
            msg := tb.NewMessage(update.Message.Chat.ID, "Привет! Я помогу найти и проанализировать бойца.")
            bot.Send(msg)
        case "search":
            // TODO: call API
            msg := tb.NewMessage(update.Message.Chat.ID, "Ищу...")
            bot.Send(msg)
        }
    }
}
```

- [ ] **Step 3: Create gRPC client**

```go
package grpc

import (
    "context"
    "fmt"

    fapiv1 "github.com/easy-tactics/api/proto/gen/fighter/agent/v1"
    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"
)

type Client struct {
    conn   *grpc.ClientConn
    client fapiv1.FighterServiceClient
}

func NewClient(addr string) (*Client, error) {
    conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        return nil, fmt.Errorf("failed to connect: %w", err)
    }

    return &Client{
        conn:   conn,
        client: fapiv1.NewFighterServiceClient(conn),
    }, nil
}

func (c *Client) SearchFighter(ctx context.Context, query string) (*fapiv1.SearchFighterResponse, error) {
    return c.client.SearchFighter(ctx, &fapiv1.SearchFighterRequest{
        Query: query,
    })
}
```

- [ ] **Step 4: Build and verify**

```bash
cd bot
go build -o bot .
```

- [ ] **Step 5: Commit**

```bash
git add bot/
git commit -m "feat(bot): add basic Telegram bot"
```

---

## Phase 6: Integration

### Task 6: End-to-End Integration

- [ ] **Step 1: Update docker-compose with environment**

```yaml
# Add environment variables
environment:
  - TELEGRAM_TOKEN=${TELEGRAM_TOKEN}
  - ANTHROPIC_API_KEY=${ANTHROPIC_API_KEY}
  - DB_PATH=/data/fighters.db
  - API_ADDR=api:50051
```

- [ ] **Step 2: Add .env.example**

```bash
# Bot
TELEGRAM_TOKEN=your_bot_token

# Owner (опционально - создаст первого пользователя с ролью owner)
OWNER_TELEGRAM_ID=123456789

# AI
ANTHROPIC_API_KEY=your_api_key
OPENAI_API_KEY=your_openai_key  # optional

# URLs (для HTTP взаимодействия между сервисами)
AI_SERVICE_URL=http://python-ai:8000
HEMAGON_MCP_URL=http://hemagon-mcp:8080
```

- [ ] **Step 3: Test docker-compose**

```bash
docker-compose up --build
# Verify all services start
```

- [ ] **Step 4: Commit**

```bash
git add .
git commit -m "feat: add docker integration"
```

---

## Summary

| Phase | Component | Status |
|-------|-----------|--------|
| 0 | Auth System (roles) | Users table + middleware |
| 1 | Docker Compose | Placeholder |
| 2 | API Service (Go) | gRPC server skeleton |
| 3 | Python AI Service | FastAPI + LLM client |
| 4 | Hemagon MCP | Placeholder server |
| 5 | Telegram Bot | Basic bot |
| 6 | Integration | Docker compose |

---

**Plan complete and saved to `docs/plans/2026-05-01-easy-tactics-mvp-plan.md`. Two execution options:**

**1. Subagent-Driven (recommended)** - I dispatch a fresh subagent per task, review between tasks, fast iteration

**2. Inline Execution** - Execute tasks in this session using executing-plans, batch execution with checkpoints for review

**Which approach?**