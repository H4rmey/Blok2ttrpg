FROM golang:1.23-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o ability-builder .

FROM alpine:3.20

RUN apk --no-cache add ca-certificates

WORKDIR /app

COPY --from=builder /app/ability-builder .
COPY --from=builder /app/config ./config
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/static ./static

RUN mkdir -p /app/data

EXPOSE 8080

ENV PORT=8080
ENV ABILITY_BUILDER_CONFIG=/app/config/ability-builder

CMD ["./ability-builder"]