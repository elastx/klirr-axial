# Axial BBS - Architecture Documentation

## Table of Contents
1. [System Overview](#system-overview)
2. [Core Components](#core-components)
3. [Data Models](#data-models)
4. [Synchronization Architecture](#synchronization-architecture)
5. [Discovery and Networking](#discovery-and-networking)
6. [Cryptographic Security](#cryptographic-security)
7. [Frontend Architecture](#frontend-architecture)
8. [Development Environment](#development-environment)

---

## System Overview

Axial BBS is a decentralized Bulletin Board System built on a peer-to-peer architecture where each node maintains a full replica of the network's data and synchronizes with peers to achieve eventual consistency.

### Architecture Principles
- **Decentralized**: No central server; all nodes are equal peers
- **Content-Addressed Storage**: All data uses deterministic hash-based IDs
- **Eventual Consistency**: Nodes converge to the same state over time
- **Cryptographic Identity**: Users identified by PGP key fingerprints
- **Protocol Agnostic**: Designed to support multiple transport mechanisms

### Technology Stack
- **Backend**: Go 1.20+ with standard library HTTP server
- **Database**: PostgreSQL with GORM ORM
- **Frontend**: React 18 with TypeScript, Mantine UI, Vite
- **Cryptography**: OpenPGP (gopenpgp on backend, openpgp.js on frontend)
- **Deployment**: Docker + Kubernetes (via Tilt for development)

---

## Core Components

### 1. Main Application (`src/main.go`)

The entry point orchestrates all system components:

```go
func main() {
    cfg := config.LoadConfig()              // Load configuration
    models.InitDB(cfg.Database)             // Initialize PostgreSQL
    models.RefreshHashes(models.DB)         // Calculate data hashes
    
    // Network discovery
    connections := discovery.CreateMulticastSockets(cfg)
    for conn := range connections {
        go discovery.StartMulticastListener(cfg, conn)
        go discovery.StartBroadcast(cfg, conn)
    }
    
    api.RegisterRoutes()                    // HTTP API
    http.ListenAndServe(":8080", nil)       // Start server
}
```

**Startup Sequence**:
1. Load config from `config.yaml` or use defaults
2. Connect to PostgreSQL and run migrations
3. Calculate initial database hash
4. Start multicast/broadcast listeners on all network interfaces
5. Register HTTP routes (API + frontend SPA)
6. Start HTTP server on port 8080

### 2. Configuration (`src/config/`)

**Structure** (`types.go`):
```go
type Config struct {
    NodeID           string         // Unique node identifier
    MulticastAddress string         // Default: 255.255.255.255 (broadcast)
    MulticastPort    int            // Default: 45678
    APIPort          int            // Default: 8080
    LogLevel         string         // Default: info
    FileStoragePath  string         // Path for file storage
    MaxFileSize      int64          // Max file upload size
    Database         DatabaseConfig
}
```

**Loading** (`config.go`):
- Reads from `config.yaml` if present
- Falls back to sensible defaults (localhost PostgreSQL, broadcast discovery)
- Supports environment variables and command-line args via struct tags

### 3. Database Layer (`src/models/`)

#### Connection Management (`database.go`)
```go
var DB *gorm.DB  // Global database instance

func InitDB(cfg config.DatabaseConfig) error {
    // Connect to PostgreSQL
    // Run AutoMigrate for User, Message, Bulletin
    // Returns error if connection/migration fails
}
```

**Error Handling**:
- Detects PostgreSQL unique constraint violations (code 23505)
- Helper `IsDuplicateError()` for idempotent sync operations

---

## Data Models

All models inherit from `Base` and use content-based IDs for conflict-free replication.

### Base Model (`src/models/model_base.go`)

```go
type Base struct {
    ID        string    `gorm:"primaryKey"`
    CreatedAt time.Time `gorm:"column:created_at;not null"`
}
```

- **ID**: Deterministic hash of content (set in `BeforeCreate`)
- **CreatedAt**: Timestamp for ordering and range queries

### User Model (`src/models/model_user.go`)

```go
type User struct {
    Base
    PublicKey   string `gorm:"column:public_key;not null"`
    Fingerprint string `gorm:"uniqueIndex"`
}
```

**Hash Calculation**:
- User ID = PGP key fingerprint (extracted from public key)
- Fingerprint acts as both unique identifier and content hash

**Validation** (`BeforeCreate`):
1. Parse public key to extract fingerprint
2. Verify supplied fingerprint matches (anti-tampering)
3. Set `Base.ID` to fingerprint
4. Set `CreatedAt` if not provided

### Message Model (`src/models/model_message.go`)

```go
type Message struct {
    Base
    Sender     Fingerprint  `gorm:"column:sender;type:text"`
    Recipients Fingerprints `gorm:"column:recipients;type:jsonb"`
    Content    Crypto       `gorm:"column:content;not null"`
}
```

**Hash Calculation**:
```go
func (m *Message) Hash() string {
    idBytes := []byte(
        m.Sender + 
        join(m.Recipients) + 
        m.Content + 
        m.CreatedAt.RFC3339Nano()
    )
    return sha256(idBytes)
}
```

**Validation** (`BeforeCreate`):
1. Analyze PGP message: extract sender, recipients, verify encrypted + signed
2. Verify sender/recipients match supplied values (anti-tampering)
3. Calculate hash and set as ID
4. **Requirements**: Must be encrypted AND signed

### Bulletin Model (`src/models/model_bulletin.go`)

```go
type Bulletin struct {
    Base
    Sender   Fingerprint `gorm:"column:sender;not null"`
    Topic    string      `gorm:"column:topic;not null"`
    Content  Crypto      `gorm:"column:content;not null"`
    ParentID *string     `gorm:"column:parent_id;default:null"`
}
```

**Hash Calculation**:
```go
func (b *Bulletin) Hash() string {
    idBytes := []byte(
        b.Sender + 
        b.Topic + 
        b.Content + 
        b.CreatedAt.RFC3339Nano()
    )
    return sha256(idBytes)
}
```

**Validation** (`BeforeCreate`):
1. Analyze PGP content: extract sender, verify signed but NOT encrypted
2. Verify sender matches (anti-tampering)
3. Calculate hash and set as ID
4. **Requirements**: Must be signed, must NOT be encrypted, must have no recipients

**Threading**: `ParentID` enables reply chains (not yet fully implemented)

---

## Synchronization Architecture

### Overview

Axial uses a **hierarchical hash-based synchronization algorithm** to efficiently identify and transfer missing data between nodes.

### Hash System (`src/models/hashing_database.go`)

#### Global Hash Set
```go
type HashSet struct {
    Messages  string  // Hash of all message IDs
    Users     string  // Hash of all user fingerprints
    Bulletins string  // Hash of all bulletin IDs
    Full      string  // Combined hash of above
}
```

**Calculation** (`RefreshHashes`):
1. Sort messages by `created_at`, hash concatenated IDs → `Messages`
2. Sort bulletins by `created_at`, hash concatenated IDs → `Bulletins`
3. Sort users by fingerprint, hash concatenated fingerprints → `Users`
4. Hash "messages:{hash}\nbulletins:{hash}\nusers:{hash}" → `Full`

**Range Hashing**:
- `GetMessagesHashRanges(periods []Period)` → hashes for time windows
- `GetBulletinsHashRanges(periods []Period)` → hashes for bulletin time windows
- `GetUsersHashRanges(ranges []StringRange)` → hashes for fingerprint ranges

### Synchronization Process (`src/synchronization/sync_process.go`)

#### High-Level Flow

```
1. Node A broadcasts (NodeID, IP, Full Hash) via multicast
2. Node B receives broadcast, compares hashes
3. If hashes differ:
   a. Node B initiates sync with Node A
   b. Exchange coarse-grained hash ranges
   c. Drill down into mismatching ranges
   d. Transfer actual data when ranges are small enough
4. Both nodes update their databases
5. Recalculate hashes, repeat until convergence
```

#### Detailed Sync Algorithm

**Phase 1: Initiation** (`StartSync`)
```go
func StartSync(node remote.API, hash string) error {
    // Check if already syncing (mutex lock)
    if !models.StartSync() { return error }
    defer models.EndSync()
    
    // Generate initial hash ranges
    periods, stringRanges := startingSyncRanges()
    hashedMessagesPeriods := models.GetMessagesHashRanges(periods)
    hashedBulletinsPeriods := models.GetBulletinsHashRanges(periods)
    hashedUsers := models.GetUsersHashRanges(stringRanges)
    
    // Execute sync rounds
    messages, bulletins, users := Sync(node, hashedMessagesPeriods, ...)
    
    // Push local data missing on remote
    SyncUsers(node, users)
    SyncMessages(node, messages)
    SyncBulletins(node, bulletins)
}
```

**Phase 2: Hash Exchange** (`SyncWithRequester`)
```go
func SyncWithRequester(requester, node, hashedPeriods, ...) {
    // Build request with our hashes
    req := api.SyncRequest{
        MessageRanges:  hashedMessagesPeriods,
        BulletinRanges: hashedBulletinsPeriods,
        Users:          hashedUsers,
    }
    
    // Send to remote node
    resp := requester.RequestSync(node, req)
    
    // Process response...
}
```

**Phase 3: Server Response** (`api.ComputeSyncResponse`)
```go
func ComputeSyncResponse(db, req) SyncResponse {
    // Compare requested ranges with our hashes
    mismatches := findMismatchingRanges(req.MessageRanges)
    
    resp := SyncResponse{Hashes: GetDatabaseHashes(db)}
    
    for mismatch := range mismatches {
        count := CountMessagesByPeriod(mismatch)
        
        if count < maxBatchSize {
            // Return actual messages
            resp.Messages.append(GetMessagesByPeriod(mismatch))
        } else {
            // Split into smaller ranges, return hashes
            splits := SplitTimeRange(mismatch, numSplits)
            resp.MessageRanges.append(GetMessagesHashRanges(splits))
        }
    }
    
    // Same for bulletins and users...
    return resp
}
```

**Phase 4: Data Ingestion** (back in `SyncWithRequester`)
```go
// Insert messages from remote
for messagesPeriod := range syncResponse.Messages {
    ourMessages := GetMessagesByPeriod(period)
    
    for msg := range messagesPeriod.Messages {
        if !msg.In(ourMessages) {
            DB.Create(&msg)  // Insert into local DB
        }
    }
    
    // Track what we have that remote doesn't
    for ourMsg := range ourMessages {
        if !ourMsg.In(messagesPeriod.Messages) {
            messagesMissingInRemote.append(ourMsg)
        }
    }
}
```

**Phase 5: Recursive Drilling** (implicit loop)
```go
// If response contains more hash ranges (splits), recurse
if len(syncResponse.MessageRanges) > 0 {
    newHashedPeriods := GetMessagesHashRanges(syncResponse.MessageRanges)
    Sync(node, newHashedPeriods, ...)  // Recursive call
}
```

### Constants and Limits
- **maxBatchSize**: 1000 items per response (prevents overwhelming network/memory)
- **numRangeSplits**: Split large ranges into 10 parts
- **Dynamic Splitting**: Large ranges split into `count / (maxBatchSize * 10)` parts

### Sync State Management (`src/models/sync.go`)
```go
type SyncState struct {
    mu        sync.RWMutex
    isSyncing bool
    hashes    HashSet
}

func StartSync() bool    // Acquire lock, return false if already syncing
func EndSync()           // Release lock
func IsSyncing() bool    // Check status
```

**Concurrency Control**:
- Only one sync operation per node at a time
- Remote nodes return `IsBusy: true` if already syncing
- Prevents race conditions and database conflicts

---

## Discovery and Networking

### Multicast/Broadcast System (`src/discovery/multicast.go`)

#### Socket Creation (`CreateMulticastSockets`)

**Strategy**:
1. Attempt to bind to all interfaces (`0.0.0.0:45678`)
2. If fails, bind to each interface individually
3. Skip unusable interfaces (down, loopback, no IPv4)

```go
func CreateMulticastSockets(cfg) []MulticastConnection {
    // Try binding to all interfaces first
    if conn := setupMulticastConn(cfg, nil, addr); success {
        return []MulticastConnection{{Conn: conn, localIP: "0.0.0.0"}}
    }
    
    // Fall back to per-interface binding
    for iface := range interfaces {
        if isUsableInterface(iface) {
            conn := setupMulticastConn(cfg, &iface, addr)
            connections.append(MulticastConnection{
                Conn: conn, 
                iface: &iface, 
                localIP: getInterfaceIP(iface)
            })
        }
    }
}
```

**Socket Configuration**:
- **Broadcast Mode** (255.255.255.255):
  - Set `SO_BROADCAST` flag
  - Set `SO_REUSEADDR` for multiple listeners
  - Bind to `0.0.0.0`
- **Multicast Mode** (224.0.0.0/4):
  - Join multicast group on specific interface
  - Set `IP_MULTICAST_LOOP=1` (receive own messages for debugging)
  - Set `IP_MULTICAST_TTL=2` (local network only)
  - Set `SO_REUSEADDR`

#### Broadcasting (`StartBroadcast`)

```go
func StartBroadcast(cfg, conn) {
    for {
        hashes := models.GetHashes()
        packet := fmt.Sprintf("%s|%s|%s", 
            cfg.NodeID, 
            conn.localIP, 
            hashes.Full)
        
        conn.Conn.WriteTo([]byte(packet), multicastAddr)
        time.Sleep(5 * time.Second)
    }
}
```

**Broadcast Format**: `NodeID|IPAddress|FullHash`
- Example: `node-alpha|192.168.1.42|a3f2b1c9d4e5f6...`
- Sent every 5 seconds

#### Listening (`StartMulticastListener`)

```go
func StartMulticastListener(cfg, conn) {
    buffer := make([]byte, 1024)
    
    for {
        n, addr := conn.Conn.ReadFrom(buffer)
        packet := string(buffer[:n])
        
        parts := strings.Split(packet, "|")
        nodeID, ipAddress, hash := parts[0], parts[1], parts[2]
        
        if nodeID == cfg.NodeID { continue }  // Ignore self
        
        go handleDiscovery(nodeID, ipAddress, hash)
    }
}
```

#### Sync Triggering (`handleDiscovery`)

```go
func handleDiscovery(nodeID, ip, hash string) {
    ourHashes := models.GetHashes()
    if hash == ourHashes.Full { return }  // Already synced
    
    node := remote.API{NodeID: nodeID, Address: ip}
    synchronization.StartSync(node, hash)
}
```

---

## Cryptographic Security

### PGP Integration (`src/models/crypto.go`)

#### Content Analysis (`Crypto.Analyze`)

```go
func (c *Crypto) Analyze() (sender, []recipients, encrypted, signed, error) {
    // Try parsing as PGP message
    message := crypto.NewPGPMessageFromArmored(string(*c))
    
    if message != nil {
        // Standard encrypted/signed message
        recipients := message.GetHexEncryptionKeyIDs()
        sender := message.GetHexSignatureKeyIDs()[0]
        return sender, recipients, len(recipients)>0, len(sender)>0, nil
    }
    
    // Try parsing as clearsigned message (for bulletins)
    if strings.Contains(*c, "-----BEGIN PGP SIGNED MESSAGE-----") {
        sigArmored := extractSignatureBlock(*c)
        signature := Signature(sigArmored)
        signer := signature.GetSignerFingerprint()
        return signer, [], false, true, nil
    }
    
    return "", [], false, false, error("invalid PGP message")
}
```

**Returns**:
- **sender**: Fingerprint of signing key (16-char hex)
- **recipients**: List of encryption key fingerprints
- **encrypted**: True if message has encryption recipients
- **signed**: True if message has valid signature
- **error**: Validation failure

### Frontend Cryptography (`web/src/services/gpg.ts`)

#### Key Management
```typescript
class GPGService {
    private currentKeyPair: KeyPair | null
    
    async generateKey(name, email, passphrase) {
        const { privateKey, publicKey } = await openpgp.generateKey({
            userIDs: [{ name, email }],
            rsaBits: 4096,
            passphrase
        })
        
        const fingerprint = await this.computeFingerprintFromKey(publicKey)
        this.currentKeyPair = { privateKey, publicKey, fingerprint }
        localStorage.setItem(STORAGE_KEY, JSON.stringify(this.currentKeyPair))
    }
    
    async computeFingerprintFromKey(publicKey) {
        // Prefer signing subkey (matches backend behavior)
        const signKey = await publicKey.getSigningKey()
        return signKey.getKeyID().toHex().toLowerCase()
    }
}
```

**Fingerprint Compatibility**:
- Frontend uses 16-character hex key ID (8-byte)
- Matches backend's `GetHexEncryptionKeyIDs()` format
- Both prefer signing subkey when available

#### Message Operations
```typescript
async encryptAndSignMessage(content, recipientPubKeys) {
    const message = await openpgp.createMessage({ text: content })
    
    const encrypted = await openpgp.encrypt({
        message,
        encryptionKeys: recipientPubKeys,
        signingKeys: this.currentKeyPair.privateKey
    })
    
    return encrypted  // Armored PGP message
}

async clearSignMessage(content) {
    const message = await openpgp.createCleartextMessage({ text: content })
    
    const signed = await openpgp.sign({
        message,
        signingKeys: this.currentKeyPair.privateKey,
        detached: false
    })
    
    return signed  // Clearsigned PGP message
}
```

---

## Frontend Architecture

### Technology Stack
- **Framework**: React 18 with TypeScript
- **UI Library**: Mantine v7 (modern component library)
- **Routing**: React Router v6
- **Build Tool**: Vite (fast HMR, optimized production builds)
- **Crypto**: openpgp.js v5

### Application Structure (`web/src/`)

```
App.tsx                 # Root component, key check gate
├── Layout.tsx          # Main app shell with navigation
├── components/
│   ├── KeyGeneration.tsx    # Initial key setup wizard
│   ├── KeyManagement.tsx    # Settings, key export
│   ├── UserList.tsx         # Browse/search users
│   ├── Messages.tsx         # Private messaging UI
│   ├── BulletinBoard.tsx    # Public forum view
│   ├── UserPicker.tsx       # Recipient selector
│   └── avatar/              # Avatar generation system
├── services/
│   ├── api.ts               # Backend HTTP client
│   ├── gpg.ts               # Cryptography operations
│   └── fingerprint.ts       # Identicon utilities
└── types/
    └── index.ts             # TypeScript interfaces
```

### Key Components

#### App Root (`App.tsx`)
```tsx
function App() {
    const gpg = GPGService.getInstance()
    
    return (
        <BrowserRouter>
            <MantineProvider theme={...}>
                {!gpg.isKeyLoaded() ? <KeyGeneration /> : <Layout />}
            </MantineProvider>
        </BrowserRouter>
    )
}
```

**Flow**:
1. Check localStorage for saved key
2. If no key → show `KeyGeneration` wizard
3. If key exists → show main `Layout` with navigation

#### Navigation (`Layout.tsx`)
```tsx
function Layout() {
    const routes = [
        { path: "/users", component: UserList, icon: IconUsers },
        { path: "/messages", component: Messages, icon: IconMessages },
        { path: "/bulletin", component: BulletinBoard, icon: IconNotes },
        { path: "/settings", component: KeyManagement, icon: IconKey }
    ]
    
    return (
        <AppShell navbar={<NavBar routes={routes} />}>
            <Routes>
                {routes.map(r => <Route path={r.path} element={<r.component />} />)}
            </Routes>
        </AppShell>
    )
}
```

### API Service (`web/src/services/api.ts`)

**Singleton Pattern**:
```typescript
class APIService {
    private static instance: APIService
    private gpg: GPGService
    
    static getInstance() { ... }
    
    async getUsers(): Promise<User[]> {
        return axios.get('/v1/users')
            .then(res => res.data.map(hydrateUser))
    }
    
    async sendBulletinPost(topic, content, parentId?) {
        const armoredContent = await this.gpg.clearSignMessage(content)
        await axios.post('/v1/bulletin', {
            topic,
            content: armoredContent,
            parent_id: parentId
        })
    }
}
```

**Key Methods**:
- `getUsers()` → Fetch all users, hydrate with avatar data
- `getMessages()` → Fetch encrypted messages
- `sendPrivateMessage()` → Encrypt + sign, POST to `/v1/messages`
- `sendBulletinPost()` → Clearsign, POST to `/v1/bulletin`
- `registerUser()` → POST public key to `/v1/users`

### Avatar System (`web/src/components/avatar/`)

**Deterministic Avatar Generation**:
- Uses fingerprint as seed
- Generates unique geometric patterns (Jazzicon-inspired)
- Consistent across all nodes for same user
- No external image dependencies

---

## Development Environment

### Tilt Orchestration (`Tiltfile`)

```python
docker_build('node', 'src', dockerfile='src/Dockerfile')

k8s_yaml([
    'dev/node-0.yaml',
    'dev/node-1.yaml', 
    'dev/node-2.yaml',
    'dev/configmaps.yaml',
    'dev/multicast-relay.yaml'
])

k8s_resource('node-instance-0', port_forwards=[8080])
k8s_resource('node-instance-1', port_forwards=[8081])
k8s_resource('node-instance-2', port_forwards=[8082])
k8s_resource('multicast-relay', port_forwards=[9999])
```

**Setup**:
1. Builds Go application Docker image
2. Deploys 3 independent nodes (separate databases)
3. Deploys multicast relay (bridges UDP between containers)
4. Port-forwards each node to localhost (8080-8082)

### Testing Synchronization

**Scenario 1: Basic Sync**
```bash
# Terminal 1: Create user on node 0
curl -X POST http://localhost:8080/v1/users -d '{"public_key":"..."}'

# Terminal 2: Watch node 1 logs
kubectl logs -f node-instance-1

# Observe: User appears on node 1 within 5-10 seconds
curl http://localhost:8081/v1/users  # User is present
```

**Scenario 2: Network Partition Recovery**
```bash
# Partition network
kubectl scale deployment multicast-relay --replicas=0

# Create data on node 0
curl -X POST http://localhost:8080/v1/bulletin -d '{...}'

# Create different data on node 1
curl -X POST http://localhost:8081/v1/bulletin -d '{...}'

# Restore network
kubectl scale deployment multicast-relay --replicas=1

# Observe: Both nodes converge to union of data
```

### Database Schema

**PostgreSQL Tables** (via GORM AutoMigrate):

```sql
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    public_key TEXT NOT NULL,
    fingerprint TEXT UNIQUE NOT NULL
);

CREATE TABLE messages (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    sender TEXT NOT NULL,
    recipients JSONB,
    content TEXT NOT NULL
);

CREATE TABLE bulletin_board (
    id TEXT PRIMARY KEY,
    created_at TIMESTAMP NOT NULL,
    sender TEXT NOT NULL,
    topic TEXT NOT NULL,
    content TEXT NOT NULL,
    parent_id TEXT
);
```

**Indexes**:
- Primary key on `id` (hash) for all tables
- Unique index on `users.fingerprint`
- Implied indexes on `created_at` (for range queries)

---

## API Endpoints

### HTTP Routes (`src/api/router.go`)

#### Discovery
- `GET /v1/ping` → Node health check

#### Synchronization
- `POST /v1/sync` → Hierarchical sync exchange
- `POST /v1/sync/messages` → Batch message insert
- `POST /v1/sync/bulletins` → Batch bulletin insert
- `POST /v1/sync/users` → Batch user insert

#### Users
- `GET /v1/users` → List all users
- `GET /v1/users/{fingerprint}` → Get specific user
- `GET /v1/users/search?q={query}` → Search users
- `GET /v1/users/recent` → Recently active users
- `POST /v1/users` → Register new user

#### Messages
- `GET /v1/messages` → List messages (filtered by recipient)
- `POST /v1/messages` → Send private message

#### Bulletin Board
- `GET /v1/bulletin` → List bulletin posts
- `POST /v1/bulletin` → Create bulletin post

#### Frontend
- `GET /*` → Serve React SPA (fallback to `index.html`)

### CORS Middleware
```go
func corsMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "*")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
        
        if r.Method == "OPTIONS" {
            w.WriteHeader(http.StatusOK)
            return
        }
        next(w, r)
    }
}
```

---

## Security Model

### Threat Model

**Assumptions**:
- Network is hostile; messages can be intercepted
- Nodes may be malicious or compromised
- Database may be modified directly

**Protections**:
1. **Content Integrity**: All messages/bulletins use content-based IDs
   - Tampering changes hash → rejected as unknown content
   - Cannot modify existing data without detection

2. **Identity Verification**: PGP signatures enforce sender authenticity
   - Messages must be signed by claimed sender
   - Verification happens in `BeforeCreate` hooks

3. **Privacy**: Private messages encrypted to recipients
   - Only holders of private key can decrypt
   - Recipients list stored in plaintext for routing

4. **Anti-Spam**: All content requires valid PGP signature
   - Cannot forge sender identity
   - Prevents anonymous spam

### Attack Scenarios

**1. Replay Attack**
- **Attack**: Resend old message
- **Defense**: Duplicate key constraint (messages have deterministic IDs)
- **Outcome**: Rejected by database

**2. Impersonation**
- **Attack**: Claim to be another user
- **Defense**: `BeforeCreate` extracts sender from signature, rejects mismatch
- **Outcome**: Rejected before database insert

**3. Content Modification**
- **Attack**: Change message content during sync
- **Defense**: ID recalculated in `BeforeCreate`, won't match existing ID
- **Outcome**: Inserted as new message (original unchanged)

**4. Man-in-the-Middle**
- **Attack**: Intercept sync protocol, inject false data
- **Defense**: All data validated in `BeforeCreate` (signature checks)
- **Outcome**: Invalid data rejected

---

## Future Architecture Considerations

### Scalability
- **Current**: All data in memory hash calculations
- **Future**: Incremental hashing, bloom filters for large datasets
- **Limit**: ~100K messages per node before performance degrades

### Protocol Abstraction
- **Current**: HTTP + multicast hardcoded
- **Future**: Plugin system for transport layers
  - Meshtastic: Slow, packet-based sync
  - LoRa: Ultra-low bandwidth, critical data only
  - IPFS: Content-addressed storage integration

### Storage Efficiency
- **Current**: Full replication across all nodes
- **Future**: Archiving mechanism
  - Nodes "own" first-received content
  - Non-owners purge old data
  - Request restoration from owners on demand

### Advanced Features
- **Threading**: Bulletin replies (ParentID exists, not yet in UI)
- **File Attachments**: Binary data sync (infrastructure exists, disabled)
- **Versioning**: Node software updates via BBS itself
- **Conflict Resolution**: Handle clock skew edge cases

---

## Deployment

### Production Deployment

**Requirements**:
- PostgreSQL 12+
- Docker or native Go 1.20+
- Root privileges on macOS (for broadcast sockets)

**Single Node**:
```bash
go build -o axial src/main.go
sudo ./axial --config production.yaml
```

**Docker Compose**:
```bash
cd docker
docker-compose up -d
```

**Configuration**:
```yaml
node_id: prod-node-1
multicast_address: 239.255.0.1
multicast_port: 45678
api_port: 8080
database:
  host: postgres
  port: 5432
  user: axial
  password: secure_password
  name: axial_production
```

### Monitoring

**Health Checks**:
- `GET /v1/ping` → Returns 200 OK if alive
- Check PostgreSQL connection: `SELECT 1;`

**Metrics** (not implemented yet):
- Sync operations per minute
- Database size growth
- Active peers count
- Message propagation latency

---

## Appendix: Key Algorithms

### Time Range Splitting (`src/models/sync.go`)

```go
func SplitTimeRange(period Period, splits int) []Period {
    start := RealizeStart(period.Start)
    end := RealizeEnd(period.End)
    
    duration := end.Sub(start)
    chunkDuration := duration / time.Duration(splits)
    
    ranges := []Period{}
    for i := 0; i < splits; i++ {
        chunkStart := start.Add(chunkDuration * time.Duration(i))
        chunkEnd := start.Add(chunkDuration * time.Duration(i+1))
        ranges = append(ranges, Period{Start: &chunkStart, End: &chunkEnd})
    }
    
    return ranges
}
```

**Example**:
- Input: `[2025-01-01 to 2025-01-31]`, splits=10
- Output: 10 periods of ~3 days each

### Fingerprint Range Splitting (`src/models/sync.go`)

```go
func SplitFingerprintRange(start, end string, splits int) []StringRange {
    // Convert hex strings to big.Int
    startInt := hexToBigInt(start)
    endInt := hexToBigInt(end)
    
    rangeSize := (endInt - startInt) / splits
    
    ranges := []StringRange{}
    for i := 0; i < splits; i++ {
        chunkStart := startInt + (rangeSize * i)
        chunkEnd := startInt + (rangeSize * (i+1))
        ranges = append(ranges, StringRange{
            Start: bigIntToHex(chunkStart),
            End: bigIntToHex(chunkEnd)
        })
    }
    
    return ranges
}
```

**Example**:
- Input: `["0000" to "ffff"]`, splits=4
- Output: `[0000-3fff, 4000-7fff, 8000-bfff, c000-ffff]`

---

## Glossary

- **Fingerprint**: 16-character hex string identifying a PGP key (key ID)
- **Hash Range**: Time period or fingerprint range with associated hash
- **Sync Round**: One complete exchange of hash ranges and data
- **Multicast**: UDP broadcast to discover peers on local network
- **Content-Based ID**: Deterministic identifier derived from data (enables CRDT)
- **Clearsigned**: PGP signature format preserving plaintext readability
- **Armored**: ASCII-encoded PGP message (vs binary)
- **GORM**: Go Object-Relational Mapper (database abstraction)
- **SPA**: Single-Page Application (React frontend)

---

## Document Version

- **Version**: 1.0
- **Date**: 2026-02-04
- **Authors**: Architecture extracted from codebase analysis
- **Status**: Accurate as of commit at time of generation
