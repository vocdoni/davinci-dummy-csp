# DaVinci dummy CSP

A tiny Go HTTP server that exposes two dummy endpoints to interact with an example of a **Credential Service Provider (CSP)**:

- `GET /root` – returns the current census root (JSON).
- `GET /proof?pid=<hex>&addr=<0xaddr>` – returns a census inclusion proof for a given **process ID** and **Ethereum address**.

This server uses [`github.com/vocdoni/davinci-node`](https://github.com/vocdoni/davinci-node) packages under the hood.

## Requirements

- Go **1.21+** (build-time) or Docker (to containerize)
- An appropriate **CSP seed** (see `CSP_SEED` below)

> If you’re just running the container, you only need Docker.

---

## Environment Variables

| Name       | Required | Default | Description |
|------------|----------|---------|-------------|
| `CSP_SEED` | **Yes**  | —       | Secret seed used to initialize the CSP. **Must be set in all environments.** |
| `CSP_PORT` | No       | `8080`  | Port the HTTP server listens on. |

> ⚠️ `CSP_SEED` is sensitive. Treat it as a secret. Don’t commit it, and don’t pass it in plain text in shared environments.

---

## Build & Run (Local)

### Run directly with Go

```bash
export CSP_SEED='your-super-secret-seed'
export CSP_PORT=8080

go mod download
go run .   # or: go build -o server . && ./server
```

### Test it

```bash
# Census root
curl -s http://localhost:8080/root | jq

# Proof (replace with real values)
curl -s "http://localhost:8080/proof?pid=<PROCESS_ID>&addr=<ETH_ADDRESS>" | jq
```

---

## Docker
A minimal, production-ready Dockerfile is included. It builds a static binary and runs on distroless (non-root).

### Build image 

```bash
docker build -t csp-server:latest .
```

### Run (detached)

```bash
docker run -d \
  --name csp-server \
  -p 8080:8080 \
  -e CSP_SEED='your-super-secret-seed' \
  -e CSP_PORT='8080' \
  csp-server:latest
```

---

## API

### GET `/root`

#### Response (200)

```json
{
  "root": "hex-encoded-root-bytes",
}
```

### GET `/proof?pid=<hex>&addr=<hex>`

#### Query params

* `pid` – Process ID as hex (with or without 0x).
* `addr` – Ethereum address (with or without checksum), e.g. 0xabc....

#### Response (200)

```json
{
  "censusOrigin": 2,
  "root": "hex-encoded-root-bytes",
  "address": "address-provided",
  "processId": "process-id-provided",
  "publicKey": "hex-encoded-csp-publickey-bytes",
  "signature": "hex-encoded-csp-signature-bytes"
}
```

#### Errors

* `400` – missing/invalid pid or addr.
* `500` – internal error generating or marshaling the proof.
