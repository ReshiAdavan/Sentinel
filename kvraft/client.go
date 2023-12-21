package raftkv

import (
	"crypto/rand"
	"math/big"
	"sync"

	"github.com/ReshiAdavan/Sentinel/rpc"
)

// Clerk is a client for a Raft-based key-value store.
type Clerk struct {
	servers   []*rpc.ClientEnd // List of RPC client endpoints for the Raft servers.
	mu        sync.Mutex       // Mutex to protect concurrent access to the next fields.
	clientId  int64            // Unique client identifier.
	requestId int64            // Incrementing request ID to distinguish different requests from the same client.
	leader    int              // Index of the server believed to be the leader.
}

// nrand generates a random 62-bit integer, used for generating unique client IDs.
func nrand() int64 {
	max := big.NewInt(int64(1) << 62)
	bigx, _ := rand.Int(rand.Reader, max)
	x := bigx.Int64()
	return x
}

// MakeClerk initializes a new Clerk instance with a list of server RPC endpoints.
func MakeClerk(servers []*rpc.ClientEnd) *Clerk {
	ck := new(Clerk)
	ck.servers = servers
	ck.clientId = nrand()
	ck.requestId = 0
	ck.leader = 0
	return ck
}

/*
 * Get fetches the current value for a key from the key-value store.
 * It returns an empty string if the key does not exist.
 * The function retries indefinitely in case of errors, trying to find the correct leader.
 */
func (ck *Clerk) Get(key string) string {
	args := GetArgs{}
	args.Key = key
	args.ClientId = ck.clientId

	// Locking to ensure that requestId is incremented atomically.
	ck.mu.Lock()
	args.RequestId = ck.requestId
	ck.requestId++
	ck.mu.Unlock()

	// Keep trying different servers until a valid response is received.
	for {
		server := ck.servers[ck.leader]
		reply := GetReply{}
		ok := server.Call("KVServer.Get", &args, &reply)
		if ok && !reply.WrongLeader {
			return reply.Value
		}
		ck.leader = (ck.leader + 1) % len(ck.servers)
	}
}

/*
 * PutAppend either puts a new value for a key or appends to an existing value, based on the operation type.
 * This is a helper function used by both Put and Append.
 */
func (ck *Clerk) PutAppend(key string, value string, op string) {
	args := PutAppendArgs{}
	args.Key = key
	args.Value = value
	args.Command = op
	args.ClientId = ck.clientId

	// Locking to ensure that requestId is incremented atomically.
	ck.mu.Lock()
	args.RequestId = ck.requestId
	ck.requestId++
	ck.mu.Unlock()

	// Keep trying different servers until a valid response is received.
	for {
		server := ck.servers[ck.leader]
		reply := PutAppendReply{}
		ok := server.Call("KVServer.PutAppend", &args, &reply)
		if ok && !reply.WrongLeader {
			return
		}
		ck.leader = (ck.leader + 1) % len(ck.servers)
	}
}

// Put inserts or updates the value for a given key in the key-value store.
func (ck *Clerk) Put(key string, value string) {
	ck.PutAppend(key, value, "put")
}

// Append appends the given value to the existing value for a given key in the key-value store.
func (ck *Clerk) Append(key string, value string) {
	ck.PutAppend(key, value, "append")
}
