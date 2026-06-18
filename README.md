# Blok2ttrpg Character Sheet

A self-hostable web application for managing character sheets in the Blok2ttrpg tabletop RPG system. Built with Go, HTMX, and Tailwind CSS.

## Features

- **Character Attributes** — name, backstory, personality, custom fields, temporary attributes
- **General Traits** — 11 traits with proficiency ladder (Clumsy → Master), free re-speccing
- **Combative Traits** — Offense, Defense, and Vital stats sharing the same point pool
- **Ability Builder** — cascading wizard to design abilities with enactments, interactions, validations, and perks
- **Leveling** — level up/down with automatic point grants, snapshot-based undo
- **YAML Import/Export** — full character or individual abilities
- **Light/Dark Theme** — toggle with localStorage persistence
- **No Database** — YAML is the save format, session state lives in memory

## Quick Start

### Prerequisites

- [Go 1.21+](https://go.dev/dl/)
- [Node.js](https://nodejs.org/) (for Tailwind CSS CLI via npx)

### Run (Development)

```powershell
./start.ps1
```

This builds CSS, compiles the binary, and starts the server in dev mode (templates loaded from disk — edit and refresh).

Open http://localhost:8080

### Run (Manual)

```bash
# Build CSS
npx tailwindcss -i ./web/static/css/input.css -o ./web/static/css/output.css --minify

# Build and run
go build -o bin/charsheet.exe ./cmd/server
DEV=1 ./bin/charsheet.exe
```

### Production Build

Without `DEV` env var, templates and static assets are embedded in the binary:

```bash
npx tailwindcss -i ./web/static/css/input.css -o ./web/static/css/output.css --minify
go build -o charsheet ./cmd/server
./charsheet
```

Single binary, no external files needed.

### Docker (Linux deployment)

```bash
# Build and run
docker compose up -d

# Or build manually
docker build -t blok2ttrpg .
docker run -d -p 8080:8080 --name charsheet blok2ttrpg
```

The image is ~15MB (Alpine-based), runs on any Linux host with Docker.

### Deploy to a Linux server

```bash
# On your dev machine: build the image and push
docker build -t your-registry/blok2ttrpg:latest .
docker push your-registry/blok2ttrpg:latest

# On the server: pull and run
docker pull your-registry/blok2ttrpg:latest
docker run -d -p 8080:8080 --restart unless-stopped --name charsheet your-registry/blok2ttrpg:latest
```

Or copy `docker-compose.yml` to the server and run `docker compose up -d`.

## Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `PORT` | `8080` | HTTP port |
| `DEV` | (unset) | If set, loads templates from disk instead of embed |

## Project Structure

```
cmd/server/          — main entry point
internal/
  models/            — domain models (Character, Traits, Abilities, YAML)
  server/            — HTTP handlers, session store, view models
  server/tmplfuncs/  — template helper functions
  gamedata/          — TTRPG rule catalog (perks, enactments, interactions)
web/
  templates/         — Go HTML templates (layouts + partials)
  static/css/        — Tailwind input/output CSS
  embed.go           — embed directives for production builds
docs/                — TTRPG game design documents
```

## How It Works

- **Session**: cookie-based in-memory sessions (24h TTL). No accounts needed.
- **State**: character lives in server memory during session. Export to YAML to save, import to restore.
- **UI**: server-rendered HTML with HTMX for dynamic updates. No JavaScript framework.
- **Ability Builder**: cascading dropdowns — pick Type → add Enactments → set Interactions → configure Validations → apply Perks. Each perk dynamically updates the computed summary.

## Game System

Based on the Blok2ttrpg tabletop RPG. Key concepts:

- **Proficiency Ladder**: Clumsy (d4) → Untrained (d6) → Trained (d8) → Expert (d10) → Master (d12) → Legendary (d20)
- **Trait Points**: shared pool for General + Combative traits, formula-based on trait count
- **Ability Points**: separate pool for building abilities, gained per level
- **Max Level**: 10
- **Re-spec**: trait points freely reallocatable; ability points refunded on perk/ability removal

## License

MIT
