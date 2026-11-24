# NoduoLink

**Your network. Everywhere.**

NoduoLink is an open-source networking project designed to make secure private connectivity and personal internet gateways accessible from anywhere in the world.

The goal is simple: connect your devices securely and let you route your internet traffic through a personal gateway you control, whether it is at home or on a remote network.

## Contents

- [Vision](#vision)
- [Project Goals](#project-goals)
- [Technology](#technology)
- [Architecture](#architecture)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [Signaling Protocol](#signaling-protocol)
- [Project Structure](#project-structure)
- [Testing](#testing)
- [Security Notes](#security-notes)
- [Troubleshooting](#troubleshooting)
- [Project Status](#project-status)
- [Contributing](#contributing)
- [License](#license)

## Vision

Personal infrastructure should not be limited by physical location.

NoduoLink aims to provide the foundations for a secure, reliable, and easy-to-use private network across devices and locations.

## Project Goals

- Secure connectivity between trusted devices.
- Personal internet gateways and remote network access.
- Reliable reconnection after network interruptions.
- Direct peer connectivity where possible, with relay fallback.
- Transparent, open-source architecture.
- Simple deployment and clear operational diagnostics.

## Technology

What is actually implemented today:

- **Go** (see `go.mod` for the current toolchain version) — coordinator and client.
- **WireGuard** via `wgctrl` — encrypted network tunnels, configured programmatically.
- **WebSocket** via `gorilla/websocket` — signaling transport.
- **YAML** via `gopkg.in/yaml.v3` — strict configuration loading.
- **Linux** — initial gateway implementation (tunnel setup needs `CAP_NET_ADMIN` / `sudo` and a pre-created `wg0` interface).

> A TypeScript-based management interface is part of the longer-term vision, but it is not implemented yet: the repository currently contains Go code only.

## Architecture

Two binaries plus shared packages:

- `cmd/noduolinkd` — coordinator: WebSocket signaling server (`/ws`), client registry, relays connection offers between nodes. Also exposes `GET /health` (returns `OK`).
- `cmd/noduolink` — client: manages local X25519 identity, registers with the coordinator, handles incoming/outgoing connection requests, configures the local WireGuard interface.
- `internal/identity` — Curve25519 keypair generation and local storage (`PRIVATE_KEY` / `PUBLIC_KEY`, file mode `0600`); `NodeID = hex(sha256(base64(public_key)))`.
- `internal/coordination` — thread-safe registry of connected clients, lookup by full NodeID or prefix.
- `internal/tunnel` — WireGuard setup via `wgctrl` (interface config, peer, allowed IPs, endpoint).
- `internal/config` — strict YAML loader (unknown fields rejected, `wss://` or `https://` required for `coordinator.url`, `log_level` defaults to `info`).
- `api/proto` — signaling envelope: `register`, `ack`, `connect_request`, `connect_offer`, `connect_accept`, `connect_reject`.

### Coordinator endpoints

| Method | Path      | Description                          |
|--------|-----------|--------------------------------------|
| GET    | `/health` | Liveness probe, responds `OK`        |
| GET    | `/ws`     | WebSocket signaling endpoint         |

### How it works

1. Each client generates (or loads) its identity and computes its `NodeID`.
2. Clients connect to the coordinator over WebSocket and send `register` with their public key.
3. Client A sends `connect_request` with B's NodeID (full ID or prefix).
4. The coordinator forwards a `connect_offer` to B; B replies `connect_accept` or `connect_reject`, relayed back to A.
5. On accept, both sides configure their local WireGuard interface (`wg0`) with the peer's key and allowed IPs.

```text
Client A                  Coordinator                  Client B
   |--- register ----------->|                             |
   |<-- ack -----------------|--- register <---------------|
   |                         |-- ack --------------------->|
   |--- connect_request ---->|                             |
   |                         |--- connect_offer ---------->|
   |                         |<-- connect_accept ----------|
   |<-- connect_accept ------|                             |
   |================ WireGuard tunnel (wg0) ===============|
```

The coordinator only handles signaling; encrypted traffic flows directly between peers over WireGuard.

## Getting Started

### Prerequisites

- Go 1.27 or later (`go version`)
- Linux for WireGuard tunnel setup
- `sudo` or `CAP_NET_ADMIN` capability, and a pre-created `wg0` interface, when configuring tunnels
- Two terminals (or machines) to try a client-to-client pairing

### Build

```sh
go build -o bin/noduolinkd ./cmd/noduolinkd
go build -o bin/noduolink ./cmd/noduolink
```

### Configure

```sh
cp config.example.yaml config.yaml
# then edit coordinator.url and identity.private_key_path
```

See [Configuration](#configuration) for the full field reference.

### Run the coordinator

```sh
./bin/noduolinkd -config config.yaml
# listening on :8080 — verify with:
curl http://localhost:8080/health
# OK
```

### Run a client

```sh
./bin/noduolink -config config.yaml -id device-01
```

To request a connection to another node (full NodeID or prefix):

```sh
./bin/noduolink -config config.yaml -id device-02 -connect-to <node-id-or-prefix>
```

Client flags:

| Flag            | Default        | Description                                              |
|-----------------|----------------|----------------------------------------------------------|
| `-config`       | `config.yaml`  | Path to the configuration file                           |
| `-id`           | `device-test-01` | Display name of this device (logs only)                |
| `-connect-to`   | `""` (none)    | NodeID (or prefix) of the peer to connect to             |
| `-accept`       | `true`         | Automatically accept incoming connection requests        |
| `-wg-private`   | `""`           | WireGuard private key (base64) — required for tunnels    |
| `-wg-peer`      | `""`           | WireGuard public key of the peer (base64)                |
| `-wg-ip`        | `10.0.0.2/24`  | Local tunnel address assigned to `wg0`                   |
| `-wg-peer-ip`   | `10.0.0.1/32`  | Allowed IP of the peer in the tunnel                     |
| `-wg-endpoint`  | `""`           | Peer endpoint `IP:port` (optional)                       |

> Note: the client currently always dials `ws://localhost:8080/ws`, and the coordinator always listens on `:8080`. `coordinator.url` is loaded and validated from the config file but is not yet used for dialing — remote coordinators are on the roadmap.

## Configuration

Example (`config.example.yaml`):

```yaml
log_level: info

coordinator:
  # Must be wss:// or https://
  url: "wss://coordinator.noduolink.dev"

identity:
  private_key_path: "/etc/noduolink/keys/private.key"
```

| Field                        | Required | Description                                                        |
|------------------------------|----------|--------------------------------------------------------------------|
| `log_level`                  | No       | Log level, defaults to `info` if omitted                           |
| `coordinator.url`            | Yes      | Coordinator URL, must use the `wss://` or `https://` scheme        |
| `identity.private_key_path`  | Yes      | Path of the identity file holding `PRIVATE_KEY` / `PUBLIC_KEY`     |

Strictness rules (enforced by `internal/config`):

- Unknown YAML fields are rejected — a typo fails fast instead of being ignored.
- A missing or malformed `coordinator.url`, or a scheme other than `wss://` / `https://`, is a fatal error at startup.
- On first run the client generates a new X25519 identity and stores it at `identity.private_key_path` (directory created with `0700`, file with `0600`); on later runs the existing identity is loaded.

## Signaling Protocol

All signaling messages share one JSON envelope:

```json
{ "type": "<message-type>", "data": { } }
```

| Type              | Direction            | Payload                                                        |
|-------------------|----------------------|----------------------------------------------------------------|
| `register`        | client → coordinator | `{ "node_id": "...", "public_key": "<base64>" }`         |
| `ack`             | coordinator → client | Welcome string                                                 |
| `connect_request` | client → coordinator | `{ "target_node_id": "<full-id-or-prefix>" }`               |
| `connect_offer`   | coordinator → target | `{ "initiator_node_id": "...", "initiator_pub_key": "..." }` |
| `connect_accept`  | target → initiator (relayed) | `{ "target_node_id": "...", "accepted": true, "message": "..." }` |
| `connect_reject`  | either direction (relayed)   | `{ "target_node_id": "...", "accepted": false, "message": "..." }` |

Minimal pairing example:

```json
// A → coordinator
{ "type": "register", "data": { "node_id": "device-01", "public_key": "BASE64==" } }
// coordinator → A
{ "type": "ack", "data": "Bienvenue sur NoduoLink" }
// A → coordinator
{ "type": "connect_request", "data": { "target_node_id": "ab12cd34" } }
// coordinator → B
{ "type": "connect_offer", "data": { "initiator_node_id": "<A>", "initiator_pub_key": "BASE64==" } }
// B → coordinator → A
{ "type": "connect_accept", "data": { "target_node_id": "<A>", "accepted": true } }
```

## Project Structure

```text
.
├── api/proto/            # signaling envelope and message types
├── cmd/
│   ├── noduolink/        # client entrypoint
│   └── noduolinkd/       # coordinator entrypoint
├── internal/
│   ├── config/           # strict YAML loader + validation (+ tests)
│   ├── coordination/     # thread-safe client registry (+ tests)
│   ├── identity/         # X25519 keypairs, disk storage, NodeID (+ tests)
│   └── tunnel/           # WireGuard setup via wgctrl
├── config.example.yaml   # sample configuration
└── README.md
```

## Testing

```sh
go test ./...
```

Covered by unit tests: identity key handling (`internal/identity`), the coordination registry (`internal/coordination`), and the strict config loader (`internal/config`).

## Security Notes

- Identities are X25519 keypairs; the private key file is written with mode `0600` (directory `0700`).
- `NodeID` is deterministic: `hex(sha256(base64(public_key)))` — sharing a NodeID does not expose the private key.
- The coordinator only ever sees public keys and NodeIDs and relays signaling; encrypted traffic goes peer-to-peer over WireGuard.
- Known limitation (development stage): the coordinator's WebSocket upgrader currently accepts any origin, the client dials a hardcoded local URL, and there is no authentication or reconnection logic yet. Do not expose the coordinator to untrusted networks.

## Troubleshooting

| Symptom | Likely cause | Fix |
|---------|--------------|-----|
| `coordinator.url doit utiliser le protocole wss:// ou https://` | URL uses `ws://`, `http://`, or is empty | Set a `wss://` or `https://` URL in `config.yaml` |
| `erreur de parsing YAML (champ inconnu ou invalide)` | Typo or extra field in `config.yaml` | The loader is strict — remove or correct the unknown field |
| `Impossible de se connecter` (client) | Coordinator not running or not on `localhost:8080` | Start `noduolinkd` first; the client currently only dials `ws://localhost:8080/ws` |
| `impossible d'initialiser wgctrl` / tunnel errors | Missing privileges or missing `wg0` interface | Run with `sudo` / `CAP_NET_ADMIN` and pre-create `wg0` (`ip link add wg0 type wireguard`) |
| `Nœud cible introuvable ou hors ligne` | Target NodeID typo or peer offline | Check the full NodeID/prefix and make sure the peer is registered |

## Project Status

NoduoLink is in its initial design and development phase. The architecture, security model, and first working prototype are being developed.

Planned capabilities are not yet production-ready features.

## Contributing

NoduoLink is an open-source project. Contributions, technical feedback, architecture discussions, and testing help shape its development.

Contribution guidelines and security reporting instructions will be published as the project matures. In the meantime, please open an issue to discuss any substantial change before submitting a pull request.

## License

NoduoLink is released under the Apache License 2.0. See `LICENSE` for the applicable terms.
