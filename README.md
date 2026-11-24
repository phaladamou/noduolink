# NoduoLink

[![Go Version](https://img.shields.io/badge/go-1.27-blue)](go.mod)
[![License](https://img.shields.io/badge/license-Apache--2.0-green)](LICENSE)
[![Platform](https://img.shields.io/badge/platform-linux-lightgrey)](README.md)

**Your network. Everywhere.**

NoduoLink is an open-source networking project designed to make secure private connectivity and personal internet gateways accessible from anywhere in the world.

The goal is simple: connect your devices securely and let you route your internet traffic through a personal gateway you control, whether it is at home or on a remote network.

> **Project status:** NoduoLink is in its initial design and development phase. The architecture, security model, and first working prototype are being developed. Planned capabilities are not yet production-ready features.

## Contents

- [Vision](#vision)
- [Project Goals](#project-goals)
- [Technology](#technology)
- [Architecture](#architecture)
- [How It Works](#how-it-works)
- [Prerequisites](#prerequisites)
- [Getting Started](#getting-started)
- [Configuration](#configuration)
- [Signaling Protocol](#signaling-protocol)
- [Project Structure](#project-structure)
- [Testing](#testing)
- [Current Limitations](#current-limitations)
- [Troubleshooting](#troubleshooting)
- [Project Status](#project-status)
- [Contributing](#contributing)
- [Security](#security)
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

The initial implementation uses:

- **Go** (`go 1.27.1`, see `go.mod`) — networking services and core components.
- **WireGuard** (via `wgctrl`) — encrypted network tunnels, configured programmatically.
- **WebSocket** (via `gorilla/websocket`) — signaling transport.
- **YAML** (via `gopkg.in/yaml.v3`) — strict configuration loading.
- **Linux** — initial gateway implementation (the client configures `wg0` with `ip addr add` / `ip link set` and therefore requires Linux with WireGuard and `sudo` / `CAP_NET_ADMIN` access).

A TypeScript-based management interface is a future plan: there is currently no TypeScript code in this repository.

Key dependencies (see `go.mod` / `go.sum`):

- `github.com/gorilla/websocket` — WebSocket signaling.
- `golang.zx2c4.com/wireguard/wgctrl` — WireGuard device configuration.
- `gopkg.in/yaml.v3` — strict configuration parsing.

## Architecture

Two binaries plus shared packages:

- `cmd/noduolinkd` — coordinator: WebSocket signaling server (`/ws`), health check (`/health`), client registry, relays connection offers between nodes. Listens on `:8080`.
- `cmd/noduolink` — client: manages local X25519 identity, registers with the coordinator, handles incoming/outgoing connection requests, configures the local WireGuard interface.
- `internal/identity` — Curve25519 (X25519) keypair generation and local storage (`PRIVATE_KEY` / `PUBLIC_KEY`, file mode `0600`); `NodeID = hex(sha256(base64(public_key)))`.
- `internal/coordination` — thread-safe registry of connected clients, with lookup by full `NodeID` or by prefix.
- `internal/tunnel` — WireGuard setup via `wgctrl` (interface config, peer, allowed IPs, endpoint).
- `internal/config` — strict YAML loader (unknown fields rejected, `coordinator.url` must use `wss://` or `https://`, `identity.private_key_path` is required, `log_level` defaults to `info`).
- `api/proto` — signaling envelope: `register`, `ack`, `connect_request`, `connect_offer`, `connect_accept`, `connect_reject`.

### Coordinator endpoints

| Method | Path | Description |
| --- | --- | --- |
| GET | `/health` | Liveness probe, responds `OK` |
| GET | `/ws` | WebSocket signaling endpoint |

### How It Works

1. Each client generates (or loads) its identity and computes its `NodeID`.
2. Clients connect to the coordinator over WebSocket (`/ws`) and send `register` with their public key.
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

## Prerequisites

- Go `1.27.1` or later.
- Linux with a pre-created WireGuard interface named `wg0` if you want tunnel setup (the code calls `wgctrl` on `wg0`, then `sudo ip addr add <ip> dev wg0` and `sudo ip link set wg0 up`).
- `sudo` access for the interface address commands.
- Network access to the coordinator on port `8080` (default local endpoint: `ws://localhost:8080/ws`, health: `http://localhost:8080/health`).

## Getting Started

1. Clone and build:

```sh
git clone https://github.com/phaladamou/noduolink.git
cd noduolink
go build ./...
```

This produces the `noduolinkd` (coordinator) and `noduolink` (client) binaries, matching `cmd/noduolinkd` and `cmd/noduolink`.

2. Create a config file from the example:

```sh
cp config.example.yaml config.yaml
```

3. Start the coordinator (default config path: `config.yaml`):

```sh
./noduolinkd --config config.yaml
# listens on http://localhost:8080
#  GET /health -> OK
#  WS  /ws     -> signaling
```

4. Start a client (new identity is generated automatically when the key file does not exist):

```sh
./noduolink --config config.yaml --id device-test-01
```

Request a connection to another node (full `NodeID` or prefix):

```sh
./noduolink --config config.yaml --id device-test-02 --connect-to <target-NodeID-prefix>
```

To configure WireGuard on accept, pass the WireGuard keys and addresses explicitly:

```sh
./noduolink --config config.yaml --id gateway-01 \
  --wg-private "<base64-private-key>" \
  --wg-peer "<base64-peer-public-key>" \
  --wg-ip "10.0.0.2/24" \
  --wg-peer-ip "10.0.0.1/32" \
  --wg-endpoint "203.0.113.10:51820"
```

Full client flags (see `cmd/noduolink/main.go`):

| Flag | Default | Description |
| --- | --- | --- |
| `--config` | `config.yaml` | Path to the YAML config file. |
| `--id` | `device-test-01` | Display name used in log prefix. |
| `--connect-to` | `""` | Target `NodeID` (or prefix) to connect to on startup. |
| `--accept` | `true` | Automatically accept incoming connection requests. |
| `--wg-private` | `""` | WireGuard private key (base64). Tunnel setup is skipped if empty. |
| `--wg-peer` | `""` | WireGuard peer public key (base64). Tunnel setup is skipped if empty. |
| `--wg-ip` | `10.0.0.2/24` | Local tunnel address assigned to `wg0`. |
| `--wg-peer-ip` | `10.0.0.1/32` | Peer allowed IP (CIDR). |
| `--wg-endpoint` | `""` | Optional peer endpoint `IP:port`. |

Coordinator flags (see `cmd/noduolinkd/main.go`):

| Flag | Default | Description |
| --- | --- | --- |
| `--config` | `config.yaml` | Path to the YAML config file. |

## Configuration

Example (`config.example.yaml`):

```yaml
# NoduoLink - example configuration
log_level: info

coordinator:
  # Must be wss:// or https://
  url: "wss://coordinator.noduolink.dev"

identity:
  private_key_path: "/etc/noduolink/keys/private.key"
```

Validation rules (`internal/config/config.go`):

- Unknown YAML fields are rejected (strict decoding).
- `coordinator.url` is required and must use the `wss://` or `https://` scheme.
- `identity.private_key_path` is required.
- `log_level` defaults to `info` when empty.
- The identity file stores `PRIVATE_KEY=<base64>` and `PUBLIC_KEY=<base64>` with file mode `0600` (directory `0700`).

## Signaling Protocol

Envelope (`api/proto/messages.go`):

```json
{ "type": "register | ack | connect_request | connect_offer | connect_accept | connect_reject", "data": {} }
```

| Type | Sender -> Receiver | Payload |
| --- | --- | --- |
| `register` | client -> coordinator | `{ "node_id": "...", "public_key": "<base64>" }` |
| `ack` | coordinator -> client | `"Bienvenue sur NoduoLink"` |
| `connect_request` | client A -> coordinator | `{ "target_node_id": "<full ID or prefix>" }` |
| `connect_offer` | coordinator -> client B | `{ "initiator_node_id": "...", "initiator_pub_key": "..." }` |
| `connect_accept` / `connect_reject` | client B -> coordinator -> client A | `{ "target_node_id": "...", "accepted": true, "message": "..." }` |

The coordinator validates that `public_key` is 32 bytes (base64) and resolves the target with prefix matching (`GetByPrefix`). Unknown targets receive a `connect_reject` (`"Nœud cible introuvable ou hors ligne"`).

## Project Structure

```text
api/proto/               signaling message types
cmd/noduolink/           client entrypoint
cmd/noduolinkd/          coordinator entrypoint
config.example.yaml      example configuration
internal/config/         strict YAML loader + validation
internal/coordination/   thread-safe client registry
internal/identity/       X25519 keys + NodeID
internal/tunnel/         WireGuard configuration via wgctrl
```

## Testing

```sh
go test ./...
```

Existing tests cover identity key generation/storage, the coordination registry (including prefix lookup), and strict config loading/validation.

## Current Limitations

- The client currently dials the hardcoded URL `ws://localhost:8080/ws` and does not yet use `coordinator.url` for dialing; `coordinator.url` is still validated but only documents the intended remote coordinator.
- The coordinator always listens on `:8080`; there is no config option for the port yet.
- WireGuard setup assumes the `wg0` device already exists and requires Linux + `sudo ip ...` privileges.
- Reconnection, relay fallback, and gateway routing described in Project Goals are not implemented yet.

## Troubleshooting

- `Erreur critique de configuration`: check `--config` path, YAML field names (strict mode), and that `coordinator.url` starts with `wss://` or `https://`.
- `Impossible de se connecter`: ensure `noduolinkd` is running and reachable on port `8080`; check `GET /health` returns `OK`.
- `Nœud cible introuvable ou hors ligne`: the target is disconnected or the `NodeID` prefix is wrong/ambiguous.
- `impossible de configurer l'interface wg0`: create `wg0` first (`ip link add wg0 type wireguard`), run with sufficient privileges, and verify `--wg-private` / `--wg-peer` keys.
- `Note: impossible d'ajouter l'IP`: the address is usually already assigned to `wg0`; safe to ignore if the tunnel works.

## Project Status

NoduoLink is in its initial design and development phase. The architecture, security model, and first working prototype are being developed.

Planned capabilities are not yet production-ready features.

## Contributing

NoduoLink is an open-source project. Contributions, technical feedback, architecture discussions, and testing help shape its development.

Contribution guidelines and security reporting instructions will be published as the project matures. Until then, please open an issue to discuss any significant change before submitting a pull request.

## Security

This is experimental software under active development. Identities are X25519 keypairs stored as base64 in the identity file — protect the file (`0600`) and its host. `NodeID` is deterministic (`hex(sha256(base64(public_key)))`) and safe to share. The coordinator only ever sees public keys and NodeIDs and relays signaling; encrypted traffic goes peer-to-peer over WireGuard.

Known limitation: the coordinator's WebSocket upgrader currently accepts any origin, the client dials a hardcoded local URL, and there is no authentication or reconnection logic yet. Do not expose a coordinator to the public internet without authentication, TLS termination, and firewalling. If you find a vulnerability, please report it privately instead of opening a public exploit.

## License

NoduoLink is released under the Apache License 2.0. See `LICENSE` for the applicable terms.
