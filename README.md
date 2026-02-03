# Axial BBS (Go + React/Vite)

## Overview
Axial BBS is a decentralized Bulletin Board System where nodes interact, synchronize data, and maintain eventual consistency across a network.

Each node:
- Hosts a local dataset, mainly a public bulletin board but also private mail, software updates, and more.
- Allows users to post and interact with messages using public key fingerprints as user IDs.
- Discovers and synchronizes with other nodes via multicast and direct API calls.
- Resolves data inconsistencies using a hierarchical hash-based synchronization algorithm.
- Supports different communication protocols, including high-speed (WiFi Mesh, Ethernet) and slow-speed (Meshtastic, JS8CALL).

### Key Features
1. **Decentralized Bulletin Board**:
   - Public message board where users can post messages identified by their public key fingerprints.
   - Messages propagate through the network, ensuring visibility across all nodes.

2. **Eventual Consistency**:
   - Nodes synchronize to achieve consistent bulletin board data sets across the network.
   - Designed for long delays, ensuring robust syncing even after months.

3. **Multicast Discovery**:
   - Nodes broadcast presence and data hash over multicast.
   - Other nodes use this information to initiate synchronization.

4. **Hierarchical Sync Algorithm**:
   - Resolves data discrepancies by drilling down into mismatched time chunks (week → day → hour).
   - Unique content-based IDs ensure conflict-free replication.

5. **Integration Interface**:
   - Modular architecture for supporting multiple protocols.
   - Examples include:
     - **Meshtastic**: Slow sync over digital radio.
     - **WiFi Mesh**: High-speed sync over local networks.

6. **Archiving Mechanism**:
   - Nodes can "own" data they first receive and archive it across the network.
   - Archived data is removed from non-owner nodes but can be restored upon request.

7. **User Interaction**:
   - Web-based UI for managing node settings and interacting with the bulletin board.
   - Mobile-friendly PWA hosted on each node for local user interaction.

## Project Goals
### Minimum Viable Product (MVP)
1. A functioning network synchronization protocol:
   - Nodes broadcast and listen on multicast addresses.
   - Synchronization resolves discrepancies and achieves eventual consistency.

2. A Tilt-based local testing environment:
   - Spin up multiple instances with different starting datasets.
   - Simulate node discovery and sync scenarios.

3. Basic HTTP API:
   - Allows nodes to exchange data during synchronization.
   - Supports hierarchical hash-based comparisons.

### Future Features
1. Enhanced Bulletin Board Functionality:
   - Support for threads and user interaction metrics.
   - Rich text formatting and attachments.

2. Protocol Integrations:
   - Meshtastic and JS8CALL for slow-sync scenarios.
   - TCP/IP over WiFi and Ethernet for high-speed operations.

3. Versioning and Updates:
   - Nodes self-update to compatible versions before syncing.

4. Archiving and Ownership:
   - Mechanisms for efficient data storage and recovery.

## Architecture
### Node Workflow
1. **Discovery**:
   - Nodes broadcast `node ID`, `hash`, and `IP` via multicast.
   - Other nodes listen for broadcasts and compare hashes.

2. **Synchronization**:
   - If hashes mismatch, nodes exchange metadata to identify discrepancies.
   - Missing data is transferred and hashes are recalculated iteratively.

3. **Data Representation**:
   - Bulletin board messages stored with unique IDs derived from content.
   - Metadata tracks ownership and archive status.

4. **Communication Protocols**:
   - Multicast for discovery.
   - HTTP API for data exchange.

### Integration Framework
Each integration implements a common interface:
- `initialize(config: map[string]interface{})`
- `broadcast(data: interface{})`
- `receive() interface{}`
- `sync(targetNode: string)`

Example Integrations:
- **Meshtastic**:
  - Mode: Slow.
  - Configuration: Device-specific settings.
  - Code: Microservice handling radio communication.

- **WiFi Mesh**:
  - Mode: Fast.
  - Configuration: Core integration with no admin reconfiguration.
  - Code: High-speed API sync for large datasets.

## Getting Started
### Prerequisites
1. [Go](https://go.dev/) (>= 1.25).
2. [Node.js + npm](https://nodejs.org/) for frontend builds.
3. [Docker](https://www.docker.com/) for Postgres + pgAdmin via Compose.

### Installation
1. Clone the repository:
   ```bash
   git clone https://github.com/skaramicke/axial-bbs.git
   cd axial-bbs
   ```

2. Build the application (and frontend):
   ```bash
   make src/axial
   ```

3. Run local stack (DB + app):
   ```bash
   make run
   ```

#### Running on macOS

On macOS, the application requires root privileges to send broadcast messages. You can run it with:

```bash
sudo ./axial --config /path/to/config.yaml
```



## Contributing
1. Fork the repository.
2. Create a feature branch:
   ```bash
   git checkout -b feature-name
   ```
3. Commit your changes and push:
   ```bash
   git commit -m "Add feature"
   git push origin feature-name
   ```
4. Open a pull request.

## License
This project is licensed under the MIT License. See the `LICENSE` file for details.

## Acknowledgments
- Inspired by decentralized protocols like CRDT, Meshtastic, and modern distributed systems.


# Notes

## Broadcast?
Perhaps we shouldn't use broadcast (255.255.255.255). Raspberry Pi should work fine with multicast.
