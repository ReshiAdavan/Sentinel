package raftkv

import (
	"crypto/rand"
	"math/big"
	"sync"

	"github.com/ReshiAdavan/Sentinel/rpc"
)
type Clerk struct {
	servers []*rpc.ClientEnd
	mu        sync.Mutex
	clientId  int64
	requestId int64
	leader    int
}

func nrand() int64 {
	max := big.NewInt(int64(1) << 62)
	bigx, _ := rand.Int(rand.Reader, max)
	x := bigx.Int64()
	return x
}

func MakeClerk(servers []*rpc.ClientEnd) *Clerk {
	ck := new(Clerk)
	ck.servers = servers
	ck.clientId = nrand()
	ck.requestId = 0
	ck.leader = 0
	return ck
}

/*
 * Fetch the current value for a key. 
 * Returns "" if the key does not exist. 
 * Keeps trying forever in the face of all other errors
 * You can send an RPC with code like this:
   ** ok := ck.servers[i].Call("KVServer.Get", &args, &reply)
 * The types of args and reply (including whether they are pointers) 
 must match the declared types of the RPC handler function's 
 arguments and reply must be passed as a pointer.
 */

func (ck *Clerk) Get(key string) string {
	args := GetArgs{}
	args.Key = key
	args.ClientId = ck.clientId
	ck.mu.Lock()
	args.RequestId = ck.requestId
	ck.requestId++
	ck.mu.Unlock()

	for ; ; ck.leader = (ck.leader + 1) % len(ck.servers) {
		server := ck.servers[ck.leader]
		reply := GetReply{}
		ok := server.Call("KVServer.Get", &args, &reply)
		if ok && !reply.WrongLeader {
			return reply.Value
		}
	}
}

/*
 * Shared by Put and Append.
 * You can send an RPC with code like this:
   ** ok := ck.servers[i].Call("KVServer.PutAppend", &args, &reply)
 * The types of args and reply (including whether they are pointers) 
 must match the declared types of the RPC handler function's 
 arguments and reply must be passed as a pointer.
 */

func (ck *Clerk) PutAppend(key string, value string, op string) {
	args := PutAppendArgs{}
	args.Key = key
	args.Value = value
	args.Command = op
	args.ClientId = ck.clientId
	ck.mu.Lock()
	args.RequestId = ck.requestId
	ck.requestId++
	ck.mu.Unlock()

	for ; ; ck.leader = (ck.leader + 1) % len(ck.servers) {
		server := ck.servers[ck.leader]
		reply := PutAppendReply{}
		ok := server.Call("KVServer.PutAppend", &args, &reply)
		if ok && !reply.WrongLeader {
			return
		}
	}
}

func (ck *Clerk) Put(key string, value string) {
	ck.PutAppend(key, value, "put")
}
func (ck *Clerk) Append(key string, value string) {
	ck.PutAppend(key, value, "append")
}
