FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY backend/go.mod ./backend/go.mod
COPY backend/cmd ./backend/cmd
COPY backend/internal ./backend/internal
WORKDIR /src/backend
RUN go build -o /out/cbm-ops ./cmd/server

FROM alpine:3.20
WORKDIR /app
COPY --from=builder /out/cbm-ops /app/cbm-ops
COPY frontend /app/frontend
ENV CBM_OPS_ADDR=:8095
ENV CBM_OPS_DATA=/data/ops.json
ENV CBM_OPS_FRONTEND_DIR=/app/frontend
EXPOSE 8095
CMD ["/app/cbm-ops"]
