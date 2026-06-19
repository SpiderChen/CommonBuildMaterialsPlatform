FROM golang:1.25-alpine AS build

WORKDIR /src
COPY industrial-control-gateway/go.mod ./
COPY industrial-control-gateway/cmd ./cmd
COPY industrial-control-gateway/internal ./internal
RUN go build -trimpath -ldflags="-s -w" -o /out/industrial-control-gateway ./cmd/industrial-control-gateway

FROM alpine:3.22
RUN adduser -D -H -u 10001 app
USER app
WORKDIR /app
COPY --from=build /out/industrial-control-gateway /usr/local/bin/industrial-control-gateway
EXPOSE 19101
ENTRYPOINT ["industrial-control-gateway"]
