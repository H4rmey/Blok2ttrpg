# syntax=docker/dockerfile:1

# ---- Build stage -----------------------------------------------------------
FROM golang:1.23-alpine AS build

WORKDIR /src

# Cache module downloads.
COPY go.mod go.sum ./
RUN go mod download

# Build the static binary.
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /out/blok2ttrpg .

# ---- Runtime stage ---------------------------------------------------------
FROM alpine:3.20

WORKDIR /app

# The server reads these directories at runtime; copy them alongside the binary.
COPY --from=build /out/blok2ttrpg /app/blok2ttrpg
COPY --from=build /src/config /app/config
COPY --from=build /src/templates /app/templates
COPY --from=build /src/static /app/static
COPY --from=build /src/docs /app/docs

# Character data is persisted here; mount a volume to keep it across restarts.
RUN mkdir -p /app/data

ENV PORT=8080
EXPOSE 8080

ENTRYPOINT ["/app/blok2ttrpg"]
