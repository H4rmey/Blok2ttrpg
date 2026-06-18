.PHONY: build run css css-watch dev clean test

# Build the CSS and Go binary
build: css
	go build -o bin/charsheet.exe ./cmd/server

# Run the server
run: css
	go run ./cmd/server

# Compile Tailwind CSS
css:
	npx tailwindcss -i ./web/static/css/input.css -o ./web/static/css/output.css --minify

# Watch Tailwind CSS for changes
css-watch:
	npx tailwindcss -i ./web/static/css/input.css -o ./web/static/css/output.css --watch

# Development: run server (assumes CSS is already built or being watched)
dev:
	go run ./cmd/server

# Run tests
test:
	go test ./...

# Clean build artifacts
clean:
	rm -rf bin/
