# ERP deployment

This folder contains the customer-side Docker Compose deployment for the ERP appliance.

## Choose a mode

Run the initializer and choose one mode:

```bash
cd ERP/deploy
./init-deploy.sh local
# or
./init-deploy.sh server
```

The script writes `.env`, generates local secrets, and prints the initial `admin` password.

Use `local` when the customer needs the ERP to run on one machine without connecting to an external database or service-side stack. Local mode starts only:

- `gateway`
- `cbmp-appliance`
- encrypted local vault volume

Use `server` when the customer deployment should connect to the full service-side stack. Server mode starts:

- `gateway`
- `cbmp-appliance`
- Postgres
- Redis
- RabbitMQ
- MinIO
- ClickHouse

You can also copy a template manually:

```bash
cp .env.local.example .env  # local vault mode
cp .env.example .env        # server/Postgres mode
```

Edit `.env` before first start. For local mode, set strong values for:

- `CBMP_DATA_KEY`
- `CBMP_INITIAL_ADMIN_PASSWORD`
- `CBMP_PUBLIC_BASE_URL`
- `CBMP_LICENSE_TRUSTED_PUBLIC_KEYS` when importing OperationsPlatform-issued authorization packages

For server mode, also set:

- `POSTGRES_PASSWORD`
- `CBMP_POSTGRES_DSN`
- `RABBITMQ_DEFAULT_PASS`
- `CBMP_RABBITMQ_URL`
- `MINIO_ROOT_PASSWORD`
- `CBMP_PUBLIC_BASE_URL`
- `CBMP_LICENSE_TRUSTED_PUBLIC_KEYS`

`CBMP_DATA_KEY` encrypts the application snapshot stored either in the local vault file or in Postgres. Keep it stable after go-live. Changing it makes existing snapshots unreadable.

`CBMP_LICENSE_TRUSTED_PUBLIC_KEYS` contains Ed25519 public key(s) trusted for imported authorization packages. It must match the public key derived from OperationsPlatform `CBM_OPS_LICENSE_ISSUER_PRIVATE_KEY`; separate multiple keys with comma or semicolon during key rotation. ERP rejects imported authorization packages when the signing public key is missing or not trusted.

## Storage

The backend supports two storage modes:

- Local mode / no `CBMP_POSTGRES_DSN`: encrypted local vault at `CBMP_DATA`.
- With `CBMP_POSTGRES_DSN`: Postgres-backed encrypted snapshot plus domain-table projection support.

`docker-compose.local.yml` intentionally does not set `CBMP_POSTGRES_DSN`.
`docker-compose.enterprise.yml` uses Postgres, so `CBMP_POSTGRES_DSN` is required.

## Start

Local mode:

```bash
docker compose --env-file .env -f docker-compose.local.yml up -d --build
```

Server mode:

```bash
docker compose --env-file .env -f docker-compose.enterprise.yml up -d --build
```

The frontend gateway listens on port `8080` and proxies `/api/*` to `cbmp-appliance:8088`.

## Notes

- The built-in super administrator is `admin`. Set `CBMP_INITIAL_ADMIN_PASSWORD` in `.env` for production deployments.
- New data files start with runtime defaults only. Do not set `CBMP_SEED_DEMO=1` or `CBMP_ERP_SEED_DEMO=1` in customer deployments unless you intentionally want local demo business records.
- Imported authorization packages require a trusted issuer public key in `CBMP_LICENSE_TRUSTED_PUBLIC_KEYS`; do not put the OperationsPlatform private key on the customer ERP server.
- `CBMP_PUBLIC_BASE_URL` must match the customer-facing URL, otherwise generated signing links and callback URLs will point to the wrong host.
- Tax gateway and map provider variables can stay empty until real customer integrations are available. Tax submission will fail closed while no real tax endpoint is configured.
- Workflow webhook, tax gateway, collection SMS/WeCom, and renewal integrations reject simulator endpoints. Configure collection provider endpoints in ERP integration settings; use `CBMP_COLLECTION_PROVIDER_TOKEN` / `CBMP_COLLECTION_PROVIDER_SECRET` only when the provider requires outbound bearer or HMAC authentication.
- Redis, RabbitMQ, and ClickHouse are optional runtime services in code. Local mode leaves them disabled; server mode starts them for full delivery verification and operations visibility.
