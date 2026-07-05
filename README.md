# Blok2 TTRPG — Ability Builder

A Go web application for building TTRPG abilities using cascading dropdowns, managing character sheets, and exporting abilities as YAML.

## Prerequisites

- [Go 1.23+](https://go.dev/dl/)
- [Docker](https://docs.docker.com/get-docker/) (optional)

## Quick Start

```bash
cd ability-builder

# Download dependencies
go mod tidy

# Run the server
go run main.go
```

Then open [http://localhost:8080](http://localhost:8080) in your browser.

## Docker

The application can be containerized using Docker.

### Build the Docker Image

```bash
docker build -t blok2ttrpg-ability-builder .
```

### Run the Container

```bash
docker run -d -p 8080:8080 --name ability-builder blok2ttrpg-ability-builder
```

### Docker Configuration

The container supports the following configuration via environment variables:

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `PORT` | `8080` | Server port |
| `ABILITY_BUILDER_CONFIG` | `/app/config/ability-builder.yaml` | Path to config file |

### Running with Custom Configuration

To use a custom configuration file, mount it as a volume:

```bash
docker run -d -p 8080:8080 \
  -v /path/to/your/config.yaml:/app/config/ability-builder.yaml \
  --name ability-builder blok2ttrpg-ability-builder
```

### Running with Persistent Data

To persist character data, mount a data directory:

```bash
docker run -d -p 8080:8080 \
  -v /path/to/your/data:/app/data \
  --name ability-builder blok2ttrpg-ability-builder
```

### Combined Example

```bash
docker run -d -p 8080:8080 \
  -v ./config/ability-builder.yaml:/app/config/ability-builder.yaml \
  -v ./data:/app/data \
  --name ability-builder \
  blok2ttrpg-ability-builder
```

Then open [http://localhost:8080](http://localhost:8080) in your browser.

## Features

### Character Sheet
- Create and manage characters with full trait proficiency configuration
- General Traits: Strength, Dexterity, Stealth, Perception, Nature, Crafting, People Skill, Performance, Thievery, Knowledge, Magic
- Combative Traits: Offense (Precision, Power, Mind, Magic) and Defense (Reflex, Constitution, Mind, Magic)
- Vital Stats: HP, Movement, Energy

### Ability Builder
- Cascading dropdown form powered by HTMX
- 4 Ability Types: Execution, Reaction, Phase, Minion
- 5 Enactment Types: Damage, Healing, Movement, Proficiency Shift, Persistent Effect
- 5 Interaction Types: Self, Direct, Ranged, Area, Area of Effect
- Validation configuration with Engagement and Counter rolls
- Perk system at every level with cost tracking

### Ability List
- Browse all abilities for a character
- View ability details with YAML output
- Export abilities as YAML files
- Delete abilities

## Project Structure

```
ability-builder/
├── main.go                          # Entry point, router
├── internal/
│   ├── handlers/                    # HTTP handlers
│   ├── models/                      # Data models + reference data
│   ├── storage/                     # JSON file persistence
│   ├── session/                     # In-memory builder session
│   └── export/                      # YAML export
├── templates/                       # HTML templates
│   └── partials/                    # HTMX partial templates
├── static/                          # CSS
└── data/                            # JSON data files
```

## Tech Stack

- **Go** stdlib `net/http` — no framework
- **html/template** — server-side rendering
- **HTMX** — dynamic UI without JavaScript framework
- **Tailwind CSS** (CDN) — styling
- **JSON files** — persistence

## Configuration

| Environment Variable | Default | Description |
|---------------------|---------|-------------|
| `PORT` | `8080` | Server port |
| `ABILITY_BUILDER_CONFIG` | `config/ability-builder.yaml` | Path to YAML configuration file |
