# Stage 1: Build CSS
FROM node:20-alpine AS css
WORKDIR /app
COPY package.json package-lock.json tailwind.config.js ./
COPY web/static/css/input.css web/static/css/input.css
COPY web/templates/ web/templates/
RUN npm ci && npx tailwindcss -i ./web/static/css/input.css -o ./web/static/css/output.css --minify

# Stage 2: Build Go binary for Linux
FROM golang:1.24-alpine AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
COPY --from=css /app/web/static/css/output.css web/static/css/output.css
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o /charsheet ./cmd/server

# Stage 3: Minimal runtime image
FROM alpine:3.20
RUN apk --no-cache add ca-certificates
COPY --from=build /charsheet /usr/local/bin/charsheet
EXPOSE 8080
ENV PORT=8080
ENTRYPOINT ["charsheet"]
