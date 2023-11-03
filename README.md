# Sentinel

A distributed key-value storage, similar in architecture to Redis, DynamoDB, etc., built in golang and using RAFT consensus.

### Inspiration

I never understood the idea behind databases like Redis, and DynamoDB. Naive, post "data structure and algorithms" me thought hashmaps are simple data structures that do not need to be scaled to the level that Redis and DynamoDB have.

However, after some curious research and learning, I realized how cool and important KV stores are in the SWE industry. Our key words are software cache, persistency and eventual consistency (CAP Theorem principles).

So, to light the path that I took to explore distributed systems, I started by learning systems design principles, and distributed schemes and their applications in networking and software. Eventually, I wanted to create a distributed systems application, and I gave into creating essentially a clone of Redis, a distributed KV store, aka Sentinel.

### Topics

- Languages: Golang
- Concepts

  - Techniques: Multi-threading & Concurrency
  - Algorithms: RAFT Consensus Algorithm
  - Other: CAP Theorem

### Use It Yourself

Not so fast hotshot. Its quite hard to run... even myself I am struggling as my new computer cannot handle this.

No procedure in place at the moment (Work In Progress)

### Architecture

Unlike NETINfra, the architecture is not split into services, since it is not a microservice project, however, we can treat the architecture almost as if it is, since Sentinel was compartmentalized into separable components as described below.

#### gobWrapper

Wrapper around Go's encoding/gob, that checks and warns about capitalization.

Note: It is not a large implementation, just a wrapper that simplifies and makes calling GOB easier for me.

##### Background

gob manages streams of gobs - binary values exchanged between an Encoder (transmitter) and a Decoder (receiver). A typical use is transporting arguments and results of remote procedure calls (RPCs) such as those provided by net/rpc.

The implementation compiles a custom codec for each data type in the stream and is most efficient when a single Encoder is used to transmit a stream of values, amortizing the cost of compilation.

Refer to the following for more info: https://pkg.go.dev/encoding/gob

#### kvraft

Main business logic around the key value store given RAFT procedure. Methods like retrieving and keys modifying as well as getting values for key value pairs.

There is also configuration for how servers coordinate together given the KV RAFT system and their respective business logic.This includes but does not limit to...

- How they record snapshots of the data, as it changes per node in the system.
- Modifying the logs of all transactions in the system, per node.
- Metrics on testing performance of the KV RAFT system.

...etc

#### Linearizability

Implementation Strategy behind the KV system.

Includes but does not limit to...

- Making nodes
- Inserting Before OR After
- Converting Entries, Linking them Together
- Cache and its associatives
- Hashing keys

#### RAFT

RAFT Consensus Algorithm Implementation.

Includes but does not limit to...

- Configuration of RAFT and its servers
- Shutdown polcies
- Leader election procedures
  - Re-election
  - Checking stale leaders
  - Agreement on terms
- Support for persistent Raft state and kv server snapsho
- Persisted States

...etc

#### RPC

Channel-based RPC that simulates a network that can lose requests, lose replies, delay messages, and entirely disconnect particular hosts. Adapted from Go net/rpc/server.go.

It essentially replicates a subset of the functionality from package go rpc.

Functionalities include...

- Sends gobWrapper-encoded values to ensure that RPCs don't include references to program objects.

##### Background

Package rpc provides access to the exported methods of an object across a network or other I/O connection. A server registers an object, making it visible as a service with the name of the type of the object. After registration, exported methods of the object will be accessible remotely. A server may register multiple objects (services) of different types but it is an error to register multiple objects of the same type.

Refer to the following for more info: https://pkg.go.dev/net/rpc
