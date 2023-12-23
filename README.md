# Sentinel

&nbsp;&nbsp;&nbsp;&nbsp; A distributed key-value database, similar in architecture to Redis, DynamoDB, etc., built in golang and using RAFT consensus.

### Inspiration

&nbsp;&nbsp;&nbsp;&nbsp; I never understood the idea behind databases like Redis, and DynamoDB. Naive, post "data structure and algorithms me" thought hashmaps are simple data structures that do not need to be scaled to the level that Redis and DynamoDB have. However, after curious research and learning, I realized that I was very wrong.

&nbsp;&nbsp;&nbsp;&nbsp; So, I started by learning systems design principles, eventually distributed schemas, and their applications in networking and software. Eventually, I wanted to create a distributed systems application, and I gave into creating Sentinel.

### Topics

- **Languages**: Golang
- <ins>**Concepts**</ins>:
  - **Techniques**: Multi-threading & Concurrency
  - **Algorithms**: RAFT Consensus Algorithm
  - **Other**: CAP Theorem

### Use It Yourself

&nbsp;&nbsp;&nbsp;&nbsp; I am struggling as my new computer cannot handle this. No procedure in place at the moment (Work In Progress)

### Architecture

&nbsp;&nbsp;&nbsp;&nbsp; Sentinel was compartmentalized into separable components as described below.

**Relevant Folder Structure**:

```
.
├── gobWrapper
│   ├── ...
│   └── gobWrapper.go
├── kvraft
│   ├── ...
│   ├── client.go
│   ├── common.go
│   ├── config.go
│   └── server.go
├── linearizability
│   ├── ...
│   ├── bitset.go
│   ├── linearizability.go
│   ├── model.go
│   └── models.go
├── raft
│   ├── ...
│   ├── config.go
│   ├── persistor.go
│   ├── raft.go
│   └── util.go
├── references
│   ├── ...
│   └── raft-atc14.pdf
├── rpc
│   ├── ...
│   └── rpc.go
└── REAMDE.md
```

#### gobWrapper

&nbsp;&nbsp;&nbsp;&nbsp; Wrapper around Go's encoding/gob, that checks and warns about capitalization. Refer to the documentation for more info: https://pkg.go.dev/encoding/gob

##### `gobWrapper.go`

- Used for encoding (serializing) and decoding (deserializing) data structures.
- Capitalization Checks:
  - Ensures that the fields of structs are properly capitalized.
- Default Value Checks:
  - Checks for non-default values in structs being decoded.

#### kvraft

##### `client.go`

- Defines a client-side interface (`Clerk`) for interacting with a key-value store implemented over a Raft consensus cluster.
- The `Clerk` structure is equipped to handle key-value operations like getting, putting, and appending values.
  - It maintains a list of server endpoints and has mechanisms to keep track of the leader server for efficient request handling
  - The client generates unique identifiers for itself and its requests to ensure correct and idempotent operations.
  - In case of server failures or leadership changes, the `Clerk` is designed to retry operations, cycling through the list of servers to find the current leader.

##### `common.go`

- Defines data structures for client-server interactions in a distributed key-value store system.
- Establishes the formats for client requests and server responses for basic operations like retrieving, adding, or modifying data.
- Handles various scenarios, including success, errors, and requests to non-leader nodes in a Raft-based cluster.

##### `config.go`

- Designed for testing a distributed key-value store that operates on top of the Raft consensus protocol.
  It includes a comprehensive setup for creating, managing, and testing a network of Raft servers and key-value service clients.
- Key aspects include:
  - Generating random server handles,
  - Managing server connections
  - Handling network partitions
  - Tracking test metrics like log sizes and RPC counts.

##### `server.go`

&nbsp;&nbsp;&nbsp;&nbsp; Implementation of a key-value store server (`KVServer`) using the Raft consensus algorithm for distributed systems.

&nbsp;&nbsp;&nbsp;&nbsp; Key aspects of the `KVServer` include:

- **Operation Handling**: It defines structures (`Op` and `Result`) to represent client operations and their outcomes. Operations are identified by unique client and request IDs.
- **Concurrency and State Management**: The server uses mutex locks to manage concurrent access to its state, ensuring consistency across multiple operations.
- **Integration with Raft**: The server relies on a Raft instance for log replication and consensus. It appends client operations to the Raft log and applies committed entries.
- **Deduplication and Leader Check**: It includes mechanisms to avoid duplicating client requests and to handle operations correctly based on the server's role (leader or follower) in the Raft cluster.
- **Snapshotting**: The server implements logic for snapshotting its state when the Raft log grows beyond a certain size, helping in log compaction and efficient state recovery.
- **Main Loop**: The `Run` function contains the main loop where the server listens for committed Raft log entries and applies them to its key-value store.
- **Debugging and Error Handling**: The code includes a debug print function and structures for handling errors and operation results.

#### Linearizability

##### `bitset.go`

- Provides a bitset type for handling a collection of bits efficiently.
- It includes methods for:
  - Creating, cloning, setting, clearing, and retrieving bits
  - Popcount (counting the number of set bits),
  - Hash (computing a hash value)
  - Equals (checking for equality with another bitset).
- Optimized for performance and memory efficiency

##### `linearizability.go`

&nbsp;&nbsp;&nbsp;&nbsp; Linearizability is a correctness condition for concurrent systems, ensuring that operations appear to occur instantaneously at some point between their invocation and response.

- Data Structures: It defines structures for representing operations (entry, node), and it uses a doubly linked list to manage the sequence of operations.
- History Manipulation: Functions like makeEntries, renumber, and convertEntries process the input history (calls and returns) into a format suitable for analysis.
- Linearizability Check: The core functionality is in checkSingle, which attempts to linearize (order) a sequence of operations while adhering to the model's constraints. It uses a backtracking algorithm to explore different orderings.
- Caching: The caching mechanism (cacheEntry, cacheContains) is used to avoid re-evaluating the same state multiple times.
- Public API: The package exposes CheckOperations and CheckEvents (with their timeout variants) as the main functions to be used for checking the linearizability of operation and event histories, respectively. These functions handle partitioning the history and running the linearizability check on each partition concurrently.

##### `model.go`

- Provides structures and utilities for performing linearizability checks on a series of operations or events in concurrent or distributed systems.
- The Operation and Event structures represent individual operations and events, respectively.
- The Model struct encapsulates the behavior of the system under test, including how to initialize its state, how to transition between states (via the Step function), and how to partition histories for checking linearizability.
- Default implementations for partitioning and state comparison (NoPartition, NoPartitionEvent, ShallowEqual) are also provided.

##### `models.go`

&nbsp;&nbsp;&nbsp;&nbsp; Provides a specific model (KvModel) for use in linearizability checks of a key-value store. It Defines the structure for inputs (KvInput) and outputs (KvOutput) of key-value operations. The KvModel uses these structures to:

- Partition Operations: It partitions the history of operations by keys. Each partition (group of operations pertaining to the same key) is checked for linearizability independently.
- Initialize State: The initial state of each key in the key-value store is represented by a string.
- Define State Transitions: The Step function defines how the state of the model changes with each operation (get, put, append) and checks if the operation's output is consistent with the model's state.
- State Equality: The model uses ShallowEqual to check if two states are the same, suitable for simple data types like strings used in this model.

#### RAFT

&nbsp;&nbsp;&nbsp;&nbsp; RAFT Consensus Algorithm Implementation.

##### `raft.go`

&nbsp;&nbsp;&nbsp;&nbsp; Details an implementation of the Raft consensus algorithm, designed to manage a replicated log in a distributed system. Key aspects include:

- **Raft Server Structure (`Raft`)**: This represents a node in a Raft cluster, maintaining the state necessary for log replication and consensus, such as current term, vote count, log entries, and server state (follower, candidate, leader).
- **Log Management**: The `Raft` structure includes mechanisms to manage a log of commands (`LogEntry`), ensuring all nodes in the cluster agree on the sequence of commands.
- **Election Process**: The code handles leader election, with servers transitioning between follower, candidate, and leader states. It includes vote requesting (`RequestVote`) and handling mechanisms.
- **Log Replication**: Leaders send `AppendEntries` requests to followers to replicate log entries, ensuring consistency across the cluster. It also manages the commit index and applies committed log entries.
- **Snapshot Handling**: The server can create and recover from snapshots, allowing it to compact the log and handle large state sizes efficiently.
- **Server Operations**: Methods like `Start`, `Kill`, and `GetState` allow the server to start log entry consensus, stop operation, and report current state and term, respectively.
- **Persistence and Recovery**: The server can persist its state and recover from this persisted state, ensuring durability across restarts.
- **Main Loop (`Run`)**: This loop runs continuously, handling state transitions based on time-outs and received messages, ensuring the Raft protocol's correctness.

##### `config.go`

- Part of a test suite for Raft
- Includes a configuration structure (config) to set up and manage a network of Raft instances for testing.
- Functions are provided to create and manipulate this test environment, such as:
  - Starting and crashing Raft servers (start1, crash1)
  - Connecting and disconnecting them from the network (connect, disconnect)
  - Checking various properties of the Raft cluster (like leadership and term agreement).

&nbsp;&nbsp;&nbsp;&nbsp; The code also includes mechanisms for submitting commands to the Raft cluster and ensuring they are committed (one, wait, nCommitted). Utilities for tracking and reporting on test progress and results (begin, end) are also included, along with functions to adjust network properties like reliability and message reordering.

##### `persistor.go`

- Used for persisting the state of Raft-based servers, including both the Raft log and the state snapshots of the key-value store (kvraft).
- Provides methods to save and retrieve Raft state (SaveRaftState, ReadRaftState) and server snapshots (SaveStateAndSnapshot, ReadSnapshot).
- Copy method creates a deep copy of a Persister, useful for creating backups or new instances with the same state.
- Thread safety is ensured using mutexes (sync.Mutex) to prevent concurrent access issues.
- Plays a crucial role in maintaining the durability and consistency aspects of the Raft consensus algorithm and the kvraft server.

#### RPC

##### `rpc.go`

- Provides a simulated network environment for testing distributed systems.
- Includes structures to represent client and server endpoints (ClientEnd, Server) and to handle RPC requests and replies (reqMsg, replyMsg).
- Network struct manages the simulated network, handling connections, message reliability, delays, and reordering.
- Servers can host multiple services (Service), and each service can handle multiple methods.
  - The Call method in ClientEnd sends an RPC request and waits for a response, handling encoding and decoding of arguments and replies.
- Crucial for testing distributed algorithms like Raft in a controlled environment with various network conditions.

&nbsp;&nbsp;&nbsp;&nbsp; In more brief terms, it essentially replicates a subset of the functionality from package go rpc.

---

If you made it this far, congrats! That concludes Sentinel's README.
