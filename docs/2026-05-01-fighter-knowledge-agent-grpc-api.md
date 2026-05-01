# Fighter Knowledge Agent — gRPC API Архитектура

## 0. Система ролей и авторизации

### 0.1. Роли

| Роль | Описание | Права |
|------|----------|-------|
| `owner` | Владелец бота | Полный доступ, управление пользователями |
| `admin` | Администратор | Управление пользователями, доступ к анализу |
| `coach` | Тренер | Добавление наблюдений, просмотр профилей |
| `fighter` | Спортсмен | Просмотр своего профиля, добавление своих наблюдений |
| `blocked` | Заблокирован | Доступ запрещён |

### 0.2. Регистрация пользователей

- Новые пользователи **не могут сами зарегистрироваться**
- Только `owner` может добавлять пользователей
- При добавлении указывается: Telegram ID, имя, роль
- Бот отклоняет все запросы от неавторизованных пользователей

### 0.5. Создание первого owner

При первом запуске системы:

1. Проверяется переменная окружения `OWNER_TELEGRAM_ID`
2. Если указана и пользователь с таким ID не существует — создаётся user с ролью `owner`
3. Это позволяет назначить владельца без ручного редактирования БД

```bash
# docker-compose.yml
environment:
  - OWNER_TELEGRAM_ID=123456789
```

---

### 0.6. Архитектура взаимодействия сервисов

```
Bot (Go) ←────── gRPC ──────▶ API Service (Go)
                                  │
                                  ▼ HTTP
                            Python AI Service
                                  │
                                  ▼ HTTP
                            MCP Servers (Hemagon, YouTube)
```

### 0.3. Workflow авторизации

```
1. Пользователь пишет боту
2. Бот проверяет Telegram ID в базе
3. Если ID есть и роль ≠ blocked → обрабатывает запрос
4. Если ID нет или роль = blocked → "Доступ запрещён"
```

### 0.4. Команды управления (только owner/admin)

- `/add_user @username role` — добавить пользователя
- `/remove_user @username` — удалить пользователя
- `/list_users` — список пользователей
- `/set_role @username new_role` — изменить роль

---

## 1. Общая структура

### 1.1. Сервисы

```
AuthService        — авторизация, управление пользователями (НОВЫЙ)
FighterService     — профили бойцов, поиск
TournamentService  — турниры, парсинг
FightService       — бои, сетка, видео
AnalysisService    — саммари, наблюдения, редактирование
VideoService       — загрузка/хранение видео
TaskService        — асинхронные задачи, статус
```

---

## 2. Fighter Service

### 2.1. RPC: SearchFighter

**Запрос:**
```protobuf
message SearchFighterRequest {
  string query = 1;          // "Иван Петров Москва"
  bool search_hemagon = 2;   // искать на хемагоне, если не найдено в базе
}
```

**Ответ:**
```protobuf
message SearchFighterResponse {
  enum Source {
    SOURCE_UNKNOWN = 0;
    SOURCE_LOCAL = 1;      // найдено в локальной базе
    SOURCE_HEMAGON = 2;    // найдено на хемагоне
    SOURCE_NOT_FOUND = 3;  // не найдено
  }

  message FighterMatch {
    string uuid = 1;
    string slug = 2;
    string full_name = 3;
    string city = 4;
    string club = 5;
    string hemagon_url = 6;
    repeated string photos = 7;
    google.protobuf.Timestamp last_updated = 8;
    bool exists_in_db = 9;
  }

  Source source = 1;
  repeated FighterMatch matches = 2;
  FighterMatch selected = 3;  // если выбран
}
```

**Юзкейс:**
1. Поиск в SQLite `fighters` по slug LIKE `%query%` или full_name LIKE `%query%`
2. Если не найдено и `search_hemagon=true` → вызов Hemagon MCP
3. Возврат вариантов или "не найдено"

### 2.2. RPC: GetFighter

**Запрос:**
```protobuf
message GetFighterRequest {
  string uuid = 1;  // или slug
}
```

**Ответ:**
```protobuf
message Fighter {
  string uuid = 1;
  string slug = 2;
  string full_name = 3;
  string city = 4;
  string club = 5;
  string hemagon_url = 6;
  repeated FighterPhoto photos = 7;
  repeated string tags = 8;
  google.protobuf.Timestamp created_at = 9;
  google.protobuf.Timestamp updated_at = 10;
  Summary latest_summary = 11;
}

message FighterPhoto {
  string url = 1;
  string source = 2;       // hemagon, vk, user
  google.protobuf.Timestamp taken_at = 3;
}

message Summary {
  string content = 1;
  google.protobuf.Timestamp updated_at = 2;
}
```

**Юзкейс:**
1. Запрос к SQLite: `SELECT * FROM fighters WHERE uuid = ? OR slug = ?`
2. Загрузка latest_summary из `fighter_summaries` с сортировкой по updated_at

### 2.3. RPC: CreateFighter

**Запрос:**
```protobuf
message CreateFighterRequest {
  string full_name = 1;
  string city = 2;
  string club = 3;
  string hemagon_url = 4;
  repeated string photos = 5;
}
```

**Ответ:** `Fighter`

**Юзкейс:**
1. Генерация UUIDv7
2. Генерация slug: `full-name-city` → транслитерация, lower case (club НЕ включается, т.к. может меняться)
3. Проверка уникальности slug в SQLite
4. Создание записи в `fighters`
5. Создание папки `fighters/{slug}/`

### 2.4. RPC: UpdateFighter

**Запрос:**
```protobuf
message UpdateFighterRequest {
  string uuid = 1;
  string full_name = 2;
  string city = 3;
  string club = 4;
  string hemagon_url = 5;
  repeated string photos_to_add = 6;
  repeated string tags_to_add = 7;
}
```

**Ответ:** `Fighter`

---

## 3. Tournament Service

### 3.1. RPC: GetFighterTournaments

**Запрос:**
```protobuf
message GetFighterTournamentsRequest {
  string fighter_uuid = 1;
  int32 limit = 2;           // по умолчанию 10
  int32 offset = 3;
}
```

**Ответ:**
```protobuf
message TournamentList {
  repeated Tournament tournaments = 1;
  int32 total = 2;
}

message Tournament {
  string uuid = 1;
  string name = 2;
  string city = 3;
  string country = 4;
  google.protobuf.Timestamp start_date = 5;
  string hemagon_url = 6;
  FighterStats stats = 7;    // статистика бойца на этом турнире
}

message FighterStats {
  int32 wins = 1;
  int32 losses = 2;
  int32 place = 3;
  repeated string opponents = 4;  // uuid соперников
}
```

**Юзкейс:**
1. Запрос к SQLite: `SELECT * FROM tournaments WHERE fighter_uuid = ? ORDER BY start_date DESC`
2. Если данных нет → парсинг через Hemagon MCP

### 3.2. RPC: ParseTournaments

**Запрос:**
```protobuf
message ParseTournamentsRequest {
  string fighter_uuid = 1;
  google.protobuf.Timestamp date_from = 2;  // по умолчанию: год назад
}
```

**Ответ:** `TournamentList` + асинхронная задача

**Юзкейс:**
1. Вызов Hemagon MCP → получение списка турниров
2. Сохранение в `tournaments` таблицу
3. Возврат списка

---

## 4. Fight Service

### 4.1. RPC: GetFighterFights

**Запрос:**
```protobuf
message GetFighterFightsRequest {
  string fighter_uuid = 1;
  string tournament_uuid = 2;   // опционально
  int32 limit = 3;
  int32 offset = 4;
}
```

**Ответ:**
```protobuf
message FightList {
  repeated Fight fights = 1;
  int32 total = 2;
}

message Fight {
  string uuid = 1;
  string tournament_uuid = 2;
  string opponent_uuid = 3;
  string opponent_name = 4;
  int32 score_win = 5;
  int32 score_lose = 6;
  string round = 7;
  google.protobuf.Timestamp fight_date = 8;  // дата боя (для оценки актуальности наблюдений)
  string video_status = 9;
  string video_url = 10;
  int64 video_timestamp = 11;
  int32 video_duration = 12;
}
```

**Юзкейс:**
1. Запрос к SQLite: сетка боев
2. Если данных нет → парсинг через Hemagon MCP

### 4.2. RPC: SearchFightVideo

**Запрос:**
```protobuf
message SearchFightVideoRequest {
  repeated string fight_uuids = 1;
  bool search_all_sources = 2;  // VK + YouTube
}
```

**Ответ:**
```protobuf
message VideoSearchResult {
  message Video {
    string platform = 1;          // vk, youtube
    string video_url = 2;
    string title = 3;
    int32 duration_seconds = 4;  // длина всего видео
    
    message FightTimestamp {
      string fight_uuid = 1;
      int64 timestamp = 2;       // старт боя в видео (секунда)
      int32 duration_seconds = 3; // продолжительность боя
    }
    repeated FightTimestamp fights = 5;
  }

  repeated Video videos = 1;
  repeated string not_found = 2;  // fight_uuids без видео
}
```

**Юзкейс:**
1. Для каждого fight_uuid:
   - VK MCP: поиск по названию "Иванов vs Петров"
   - YouTube MCP: поиск по запросу
2. Агрегирование результатов
3. Сохранение в `fight_videos`

---

## 5. Analysis Service

### 5.1. RPC: GenerateSummary

**Запрос:**
```protobuf
message GenerateSummaryRequest {
  string fighter_uuid = 1;
  repeated string fight_uuids = 2;   // какие бои анализировать
  bool include_tournaments = 3;
  bool regenerate = 4;                // перегенерировать, если уже есть
}
```

**Ответ:** `Task` (асинхронный)

**Юзкейс:**
1. Создание задачи в `tasks` таблице
2. Запуск воркфлоу:
   - Загрузка данных о выбранных боях
   - Загрузка существующих наблюдений
   - Вызов LLM с промптом анализа
   - Сохранение summary в .md файл
   - Обновление статуса задачи
3. Возврат task_id для отслеживания

### 5.2. RPC: AddObservation

**Запрос:**
```protobuf
message AddObservationRequest {
  string fighter_uuid = 1;
  string content = 2;           // текст наблюдения
  string source = 3;            // user, coach, system
  repeated string tags = 4;
  string video_uuid = 5;        // опционально
  string tournament_uuid = 6;   // опционально
  string fight_uuid = 7;        // опционально
}
```

**Ответ:** `Observation`

**Юзкейс:**
1. Сохранение в `observations` таблицу
2. Генерация .md файла `fighters/{slug}/observations/{uuid}.md`
3. Опционально: вызов LLM для интеграции в summary

### 5.3. RPC: EditSummary

**Запрос:**
```protobuf
message EditSummaryRequest {
  string fighter_uuid = 1;
  string content = 2;           // новый контент (полная замена)
  string reason = 3;           // причина изменения
}
```

**Ответ:** `Summary`

**Юзкейс:**
1. Создание backup копии текущего summary в `.history/`
2. Сохранение нового контента
3. Логирование в историю

### 5.4. RPC: GetSummaryHistory

**Запрос:**
```protobuf
message GetSummaryHistoryRequest {
  string fighter_uuid = 1;
  int32 limit = 2;
}
```

**Ответ:**
```protobuf
message HistoryList {
  repeated HistoryEntry entries = 1;
}

message HistoryEntry {
  string version = 1;
  google.protobuf.Timestamp changed_at = 2;
  string changed_by = 3;       // user_id или "system"
  string reason = 4;
  string diff = 5;             // или ссылка на полный файл
}
```

### 5.5. RPC: GetObservation

**Запрос:**
```protobuf
message GetObservationsRequest {
  string fighter_uuid = 1;
  string source = 2;           // опционально
  int32 limit = 3;
}
```

**Ответ:**
```protobuf
message ObservationList {
  repeated Observation observations = 1;
}
```

---

## 6. Video Service

### 6.1. RPC: UploadUserVideo

**Запрос:**
```protobuf
message UploadUserVideoRequest {
  string fighter_uuid = 1;
  bytes video_data = 1;        // или URL
  string comment = 2;           // текстовый комментарий
  string video_format = 3;      // mp4, mov и т.д.
}
```

**Ответ:** `VideoMetadata`

**Юзкейс:**
1. Сохранение в S3: `videos/{fighter_uuid}/{uuid}.{format}`
2. Сохранение метаданных в `user_videos` таблице
3. Сохранение комментария как observation

---

## 7. Task Service (асинхронные операции)

### 7.1. RPC: CreateTask

**Запрос:**
```protobuf
message CreateTaskRequest {
  string type = 1;              // PARSE_TOURNAMENTS, GENERATE_SUMMARY, SEARCH_VIDEOS
  string fighter_uuid = 2;
  google.protobuf.Struct params = 3;
  string callback_url = 4;     // webhook для уведомления
}
```

**Ответ:** `Task`

### 7.2. RPC: GetTask

**Запрос:**
```protobuf
message GetTaskRequest {
  string task_uuid = 1;
}
```

**Ответ:**
```protobuf
message Task {
  string uuid = 1;
  string type = 2;
  string status = 3;           // PENDING, PROCESSING, COMPLETED, FAILED
  int32 progress = 4;          // 0-100
  string current_step = 5;     // "Парсинг турниров", "Поиск видео"
  google.protobuf.Timestamp created_at = 6;
  google.protobuf.Timestamp completed_at = 7;
  string result = 8;           // JSON результат
  string error = 9;
}
```

### 7.3. RPC: ListTasks

**Запрос:**
```protobuf
message ListTasksRequest {
  string fighter_uuid = 1;
  string status = 2;
  int32 limit = 3;
}
```

---

## 8. База данных (SQLite)

### 8.1. Таблицы

```sql
-- Пользователи бота
CREATE TABLE users (
  id INTEGER PRIMARY KEY,
  telegram_id BIGINT UNIQUE NOT NULL,
  username TEXT,
  full_name TEXT,
  role TEXT DEFAULT 'fighter',  -- owner, admin, coach, fighter, blocked
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Профили бойцов
CREATE TABLE fighters (
  id INTEGER PRIMARY KEY,
  uuid TEXT UNIQUE NOT NULL,
  slug TEXT UNIQUE NOT NULL,
  full_name TEXT NOT NULL,
  city TEXT,
  club TEXT,
  hemagon_url TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Фото бойцов
CREATE TABLE fighter_photos (
  id INTEGER PRIMARY KEY,
  fighter_uuid TEXT REFERENCES fighters(uuid),
  url TEXT NOT NULL,
  source TEXT,  -- hemagon, vk, user
  taken_at TIMESTAMP
);

-- Теги бойцов
CREATE TABLE fighter_tags (
  id INTEGER PRIMARY KEY,
  fighter_uuid TEXT REFERENCES fighters(uuid),
  tag TEXT NOT NULL,
  source TEXT,  -- auto, manual
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Турниры
CREATE TABLE tournaments (
  id INTEGER PRIMARY KEY,
  uuid TEXT UNIQUE NOT NULL,
  fighter_uuid TEXT REFERENCES fighters(uuid),
  name TEXT NOT NULL,
  city TEXT,
  country TEXT,
  start_date DATE,
  hemagon_url TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Бои (сетка)
CREATE TABLE fights (
  id INTEGER PRIMARY KEY,
  uuid TEXT UNIQUE NOT NULL,
  fighter_uuid TEXT REFERENCES fighters(uuid),
  tournament_uuid TEXT REFERENCES tournaments(uuid),
  opponent_uuid TEXT REFERENCES fighters(uuid),
  opponent_name TEXT,
  score_win INTEGER,
  score_lose INTEGER,
  round TEXT,
  fight_date DATE,  -- дата боя (может отличаться от даты турнира)
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Видео боев
CREATE TABLE fight_videos (
  id INTEGER PRIMARY KEY,
  fight_uuid TEXT REFERENCES fights(uuid),
  platform TEXT,  -- vk, youtube
  video_url TEXT NOT NULL,
  title TEXT,
  timestamp_seconds BIGINT,
  duration_seconds INTEGER,
  confidence REAL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Наблюдения
CREATE TABLE observations (
  id INTEGER PRIMARY KEY,
  uuid TEXT UNIQUE NOT NULL,
  fighter_uuid TEXT REFERENCES fighters(uuid),
  content TEXT NOT NULL,
  source TEXT,  -- user, coach, system
  video_uuid TEXT,
  tournament_uuid TEXT,   -- опционально
  fight_uuid TEXT,       -- опционально
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Наблюдения-теги
CREATE TABLE observation_tags (
  id INTEGER PRIMARY KEY,
  observation_uuid TEXT REFERENCES observations(uuid),
  tag TEXT NOT NULL
);

-- Саммари (история)
CREATE TABLE summaries (
  id INTEGER PRIMARY KEY,
  fighter_uuid TEXT REFERENCES fighters(uuid),
  content TEXT NOT NULL,
  version INTEGER NOT NULL,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  created_by TEXT
);

-- Пользовательские видео
CREATE TABLE user_videos (
  id INTEGER PRIMARY KEY,
  uuid TEXT UNIQUE NOT NULL,
  fighter_uuid TEXT REFERENCES fighters(uuid),
  s3_key TEXT NOT NULL,
  format TEXT,
  comment TEXT,
  analyzed BOOLEAN DEFAULT FALSE,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Задачи
CREATE TABLE tasks (
  id INTEGER PRIMARY KEY,
  uuid TEXT UNIQUE NOT NULL,
  type TEXT NOT NULL,
  fighter_uuid TEXT REFERENCES fighters(uuid),
  status TEXT DEFAULT 'PENDING',
  progress INTEGER DEFAULT 0,
  current_step TEXT,
  params TEXT,  -- JSON
  result TEXT,  -- JSON
  error TEXT,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  completed_at TIMESTAMP
);

-- Индексы
CREATE INDEX idx_fighters_slug ON fighters(slug);
CREATE INDEX idx_fighters_full_name ON fighters(full_name);
CREATE INDEX idx_tournaments_fighter ON tournaments(fighter_uuid);
CREATE INDEX idx_fights_fighter ON fights(fighter_uuid);
CREATE INDEX idx_fights_tournament ON fights(tournament_uuid);
CREATE INDEX idx_observations_fighter ON observations(fighter_uuid);
CREATE INDEX idx_tasks_fighter ON tasks(fighter_uuid);
CREATE INDEX idx_tasks_status ON tasks(status);
```

### 8.2. Запросы (типичные)

```sql
-- Поиск бойца по имени
SELECT * FROM fighters WHERE full_name LIKE '%Иван%' OR slug LIKE '%ivan%';

-- Получить профиль с фото и тегами
SELECT f.*, 
       json_group_array(p.url) as photos,
       json_group_array(t.tag) as tags
FROM fighters f
LEFT JOIN fighter_photos p ON f.uuid = p.fighter_uuid
LEFT JOIN fighter_tags t ON f.uuid = t.fighter_uuid
WHERE f.uuid = ?
GROUP BY f.uuid;

-- Последнее саммари
SELECT * FROM summaries 
WHERE fighter_uuid = ? 
ORDER BY version DESC 
LIMIT 1;
```

---

## 9. Внешние вызовы

### 9.1. Python AI Service (HTTP)

AI Service предоставляет REST API для генерации:

| Endpoint | Метод | Назначение |
|----------|-------|------------|
| `/generate-summary` | POST | Генерация саммари |
| `/integrate-observation` | POST | Интеграция наблюдения |
| `/recommend-fights` | POST | Рекомендация боев |
| `/generate-tags` | POST | Генерация тегов |

### 9.2. MCP серверы (HTTP)

| MCP | URL | Метод | Назначение |
|-----|-----|-------|------------|
| Hemagon MCP | http://hemagon-mcp:8080 | POST /search_fencer | Поиск профиля |
| Hemagon MCP | http://hemagon-mcp:8080 | POST /get_profile | Парсинг профиля |
| Hemagon MCP | http://hemagon-mcp:8080 | POST /get_tournaments | Список турниров |
| Hemagon MCP | http://hemagon-mcp:8080 | POST /get_tournament_grid | Сетка боев |
| VK MCP | http://vk-mcp:8080 | POST /search_group | Поиск группы |
| YouTube MCP | http://youtube-mcp:8080 | POST /search_videos | Поиск видео |

### 9.3. Конфигурация URL

```yaml
# docker-compose.yml
environment:
  - AI_SERVICE_URL=http://python-ai:8000
  - HEMAGON_MCP_URL=http://hemagon-mcp:8080
  - YOUTUBE_MCP_URL=http://youtube-mcp:8080
``` |

---

## 10. Workflow детализация

### 10.1. Поиск бойца

```
1. SearchFighter(query)
   ├── SQLite: SELECT * FROM fighters WHERE slug LIKE '%query%' OR full_name LIKE '%query%'
   └── Если не найдено && search_hemagon:
       └── Hemagon MCP: SearchFencer(query)
       └── Парсинг результатов
       └── Возврат списка вариантов
```

### 10.2. Выбор профиля и парсинг турниров

```
2. User выбирает профиль (через UI/бот)
3. GetFighterTournaments(fighter_uuid)
   ├── SQLite: SELECT * FROM tournaments WHERE fighter_uuid = ?
   └── Если пусто:
       └── CreateTask(PARSE_TOURNAMENTS, fighter_uuid)
       └── Hemagon MCP: GetTournaments(fencer_id)
       └── Сохранение в tournaments
       └── Возврат списка
```

### 10.3. Выбор боев и поиск видео

```
4. User выбирает турниры
5. GetFighterFights(fighter_uuid, tournament_uuids)
   ├── SQLite: SELECT * FROM fights WHERE fighter_uuid = ? AND tournament_uuid IN (...)
   └── Если пусто:
       └── Для каждого турнира: Hemagon MCP: GetTournamentGrid()
       └── Сохранение в fights
       └── Возврат списка
```

### 10.4. Генерация саммари

```
6. User выбирает бои для анализа
7. GenerateSummary(fighter_uuid, fight_uuids)
   ├── CreateTask(GENERATE_SUMMARY, fighter_uuid, {fight_uuids})
   └── Асинхронно:
       ├── Загрузка данных о боях из SQLite
       ├── Загрузка наблюдений из SQLite
       ├── LLM: GenerateSummary(fights, observations)
       ├── Сохранение summary.md
       ├── Обновление summaries таблицы (version++)
       └── UpdateTask(status=COMPLETED)
```

---

## 11. Примеры ошибок и обработка

| Ошибка | Обработка |
|--------|-----------|
| Hemagon MCP недоступен | Вернуть ошибку, предложить повторить позже |
| Боец не найден на хемагоне | Предложить создать профиль вручную |
| Видео не найдено | Пометить video_status=NOT_FOUND, не блокировать процесс |
| LLM вернул пустой результат | Повторить запрос, если не помог → ручное редактирование |
| Slug уже существует | Добавить суффикс: ivan-petrov-msk-2 |
| Задача зависла > 30 мин | Пометить FAILED, уведомить пользователя |

---

## 12. gRPC Reflection

Для удобства разработки включить gRPC reflection — можно смотреть методы через `grpcurl`.