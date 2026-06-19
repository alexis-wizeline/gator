# Gator

A CLI RSS feed aggregator written in Go. Register an account, follow RSS feeds, run the aggregator in the background, and browse the latest posts from the feeds you follow.

---

## Prerequisites

- [Go 1.21+](https://go.dev/dl/)
- [PostgreSQL](https://www.postgresql.org/)
- [goose](https://github.com/pressly/goose) — database migrations
- [sqlc](https://sqlc.dev/) — only needed if you modify SQL queries

```sh
go install github.com/pressly/goose/v3/cmd/goose@latest
go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest   # optional
```

---

## Installation

```sh
git clone https://github.com/alexis-wizeline/gator
cd gator
go build -o gator .
```

---

## Configuration

Gator reads its config from `~/.gatorconfig.json`. Create this file before running any command:

```json
{
  "db_url": "postgres://username:password@localhost:5432/gator?sslmode=disable",
  "current_user_name": ""
}
```

---

## Database Setup

Run all migrations from the project root:

```sh
goose -dir sql/migrations postgres "your-connection-string" up
```

This creates the following tables: `users`, `feeds`, `feed_follows`, `posts`.

---

## Commands

### Account

| Command | Arguments | Description |
|---|---|---|
| `register` | `<username>` | Create a new user and log in as them |
| `login` | `<username>` | Switch the active user |
| `users` | — | List all registered users (current user is marked) |
| `reset` | — | Delete all users and their data |

```sh
./gator register alice
./gator login alice
./gator users
```

### Feeds

| Command | Arguments | Description |
|---|---|---|
| `addfeed` | `<name> <url>` | Add a new RSS feed and auto-follow it |
| `feeds` | — | List all feeds and their owners |
| `follow` | `<url>` | Follow an existing feed by URL |
| `unfollow` | `<url>` | Unfollow a feed |
| `following` | — | List feeds the current user follows |

```sh
./gator addfeed "Wagslane" https://www.wagslane.dev/index.xml
./gator feeds
./gator follow https://www.wagslane.dev/index.xml
./gator unfollow https://www.wagslane.dev/index.xml
./gator following
```

### Aggregator

| Command | Arguments | Description |
|---|---|---|
| `agg` | `<interval>` | Continuously fetch all feeds at the given interval |

The interval accepts any Go duration string: `30s`, `1m`, `5m`, `1h`.

```sh
./gator agg 1m
```

> Run this in a separate terminal. It loops until interrupted (Ctrl+C).

### Browsing

| Command | Arguments | Description |
|---|---|---|
| `browse` | `<limit> [offset]` | Show the latest posts from your followed feeds |

```sh
./gator browse 10
./gator browse 10 20   # page 3
```

---

## Project Structure

```
gator/
├── main.go                        # Entry point, command registration
├── internal/
│   ├── commands/
│   │   └── commands.go            # All command handlers + middleware
│   ├── config/
│   │   └── config.go              # ~/.gatorconfig.json read/write
│   ├── gatordb/                   # sqlc-generated DB layer
│   │   ├── models.go
│   │   ├── users.sql.go
│   │   ├── feeds.sql.go
│   │   ├── feed_follows.sql.go
│   │   └── posts.sql.go
│   ├── rss/
│   │   └── rss.go                 # RSS feed fetching + parsing
│   └── state/
│       └── state.go               # Shared app state (DB + config)
└── sql/
    ├── migrations/                # goose migration files
    └── queries/                   # sqlc source queries
```

---

## Development

### Regenerate DB code after changing SQL queries

```sh
sqlc generate
```

### Add a new migration

```sh
goose -dir sql/migrations create <migration_name> sql
```

### Build

```sh
go build ./...
```
