FROM golang:1.25-alpine AS backend-build
WORKDIR /src/backend
COPY backend/go.mod backend/go.sum ./
RUN go mod download
COPY backend ./
RUN CGO_ENABLED=0 GOOS=linux go build -trimpath -ldflags="-s -w" -o /out/cbmp-appliance ./cmd/appliance

FROM alpine:3.20
RUN addgroup -S cbmp && adduser -S cbmp -G cbmp
WORKDIR /app
COPY --from=backend-build /out/cbmp-appliance /app/cbmp-appliance
RUN mkdir -p /data && chown -R cbmp:cbmp /data /app
USER cbmp
ENV CBMP_ADDR=0.0.0.0:8088
ENV CBMP_DATA=/data/app.vault
EXPOSE 8088
VOLUME ["/data"]
ENTRYPOINT ["/app/cbmp-appliance"]
