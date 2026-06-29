FROM golang:1.25.0-alpine AS builder

WORKDIR /app

COPY go.mod ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/socket-hub ./main.go

FROM alpine:3.22 AS runtime

ENV ENV=production

WORKDIR /app

RUN apk add --no-cache tini && \
    addgroup -S app && adduser -S -G app app

COPY --from=builder /out/socket-hub ./socket-hub
COPY entrypoint.sh ./entrypoint.sh

RUN chmod +x ./entrypoint.sh && \
    chown -R app:app /app

USER app

ENTRYPOINT ["/sbin/tini", "--", "./entrypoint.sh"]
