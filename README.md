# Axial BBS (Go)

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
1. [Go](https://go.dev/) (>= 1.20).
2. [Docker](https://www.docker.com/) for containerization.
3. [Tilt](https://tilt.dev/) for local testing and orchestration.

### Installation
1. Clone the repository:
   ```bash
   git clone https://github.com/skaramicke/axial-bbs.git
   cd axial-bbs
   ```

2. Build the application:
   ```bash
   go build .
   ```

3. Run a single instance:
   ```bash
   ./axial --config /path/to/config.yaml
   ```

#### Running on macOS

On macOS, the application requires root privileges to send broadcast messages. You can run it with:

```bash
sudo ./axial --config /path/to/config.yaml
```


### Using Tilt for Local Testing
1. Install Tilt:
   ```bash
   curl -fsSL https://github.com/tilt-dev/tilt/releases/download/v0.32.0/tilt.0.32.0.linux.x86_64.tar.gz | tar xz
   mv tilt /usr/local/bin
   ```

2. Start the environment:
   ```bash
   tilt up
   ```

3. Access logs and UI:
   - Visit the Tilt dashboard at `http://localhost:10350`.

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

## Mitigating spam

### Messages
Messages need to be signed by the author. Messages received in synchronization that are not signed by the author are ignored.

### Users
During synchronization, users that are not signed by known users are ignored.
Only the first user on a node is allowed to post messages without being signed. This is, however, only a user interface
restriction since a malicious actor could simply add them to their node's models.
Other users can register and then be signed by other signed users, or the first user.
Unsigned users are never synchronized to other nodes, except for if the same user is found to be signed on another node.
If the same user is signed in one of two nodes but not the other, it would be considered a new user on the other node and
the that node would update the signature status when it discovers the users by fingerprint. Since the ID hash includes
the signature status, they would be lead to mismatching hashes until they are both signed with the same signature.
There is need of a separate sync step for users that are signed by different other users:
If the same user is signed by two different users, this is discovered when one of the nodes gets or sends that user to the remote.
The node that discovers the discrepancy compares the age of the signing user and chooses the oldest one. If it had the newest one,
it would sign the user with the oldest signature be done. If it had the oldest one, it would send its own version of the user to the remote and let it update its version.

### Nodes
For a node to be accepted during synchronization, its first user needs to be signed. If it's not, the node is added to a
list of pending nodes. The first user of the node can then be signed by a signed user, and the node will be accepted from then on.
The only exception is the new user signature reconciliation mentioned above. If the first user of a node is unsigned but the same
user is signed on a remote node, that signature will be used to sign the same first user on the new node, and then the node will be accepted and eventually have earlier users be the first user of that node after synchronization is complete.

### Problems?
1. What if a malicious person creates a node with a bunch of spam data and then copies the first user of another node?
Well, the spam messages are still not signed by any user of the signature chain of users, and would be ignored when received
during synchronization.
2. What if a malicious person creates a node with a bunch of spam data and then copies the first user of another node and signs it with a new user? Well the malicious person would have to have a signed account so that the new signed user is part of the signature chain. A malicious actor may set up hundreds of nodes but for those nodes to disrupt another network, they would have
to be socially engineered into the other network. If the malicious actor is part of the network, they could disrupt it in other ways as well.