# socket-hub

Realtime WebSocket hub for Gauas services.

## Endpoints

- `GET /health`
- `GET /connection/websocket?channel=<channel_id>`
- `GET /ws?channel=<channel_id>`
- `POST /api`

`POST /api` accepts Centrifugo-style newline-delimited commands, so existing
`github.com/centrifugal/gocent/v3` publishers can use it.

## Environment

```env
ENV=development
PORT=8085
SOCKET_API_KEY=
WRITE_TIMEOUT_SECONDS=10
```

Use the same value for notification-service `CENTRIFUGO_API_KEY` and socket-hub
`SOCKET_API_KEY`.
