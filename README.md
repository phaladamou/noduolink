# NoduoLink

**Your network. Everywhere.**

NoduoLink is an open-source networking project designed to make secure private connectivity and personal internet gateways accessible from anywhere in the world.

The goal is simple: connect your devices securely and let you route your internet traffic through a personal gateway you control, whether it is at home or on a remote network.

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

The initial architecture is planned around:

- **Go** — networking services and core components.
- **WireGuard** — encrypted network tunnels.
- **TypeScript** — management interface.
- **Linux** — initial gateway implementation.

These are initial design choices and may evolve as the project develops.

## Architecture

Two binaries plus shared packages:

- `cmd/noduolinkd` — coordinator: WebSocket signaling server (`/ws`), client registry, relays connection offers between nodes.
- `cmd/noduolink` — client: manages local X25519 identity, registers with the coordinator, handles incoming/outgoing connection requests, configures the local WireGuard interface.
- `internal/identity` — Curve25519 keypair generation and local storage (`PRIVATE_KEY` / `PUBLIC_KEY`, `0600`); `NodeID = sha256(public_key)`.
- `internal/coordination` — thread-safe registry of connected clients.
- `internal/tunnel` — WireGuard setup via `wgctrl` (interface config, peer, allowed IPs, endpoint).
- `internal/config` — strict YAML loader (unknown fields rejected, `wss://` or `https://` required).
- `api/proto` — signaling envelope: `register`, `ack`, `connect_request`, `connect_offer`, `connect_accept`, `connect_reject`.

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

## Project Status

NoduoLink is in its initial design and development phase. The architecture, security model, and first working prototype are being developed.

Planned capabilities are not yet production-ready features.

## Contributing

NoduoLink is intended to be an open-source project. Contributions, technical feedback, architecture discussions, and testing will help shape its development.

Contribution guidelines and security reporting instructions will be published as the project matures.

## License

NoduoLink is intended to be released under the Apache License 2.0. See `LICENSE` for the applicable terms.
