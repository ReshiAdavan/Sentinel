package raftkv

import (
	"bytes"
	"log"
	"sync"
	"time"

	"github.com/ReshiAdavan/Sentinel/gobWrapper"
	"github.com/ReshiAdavan/Sentinel/raft"
	"github.com/ReshiAdavan/Sentinel/rpc"
)

const Debug = 1

func DPrintf(format string, a ...interface{}) (n int, err error) {
	if Debug > 0 {
		log.Printf(format, a...)
	}
	return
}

type Op struct {
	Command   string // "get" | "put" | "append"
	ClientId  int64
	RequestId int64
	Key       string
	Value     string
}

type Result struct {
	Command     string
	OK          bool
	ClientId    int64
	RequestId   int64
	WrongLeader bool
	Err         Err
	Value       string
}

type KVServer struct {
	mu      sync.Mutex
	me      int
	rf      *raft.Raft
	applyCh chan raft.ApplyMsg

	maxraftstate int // snapshot if log grows this big

	data     map[string]string   // key-value data
	ack      map[int64]int64     // client's latest request id (for deduplication)
	resultCh map[int]chan Result // log index to result of applying that entry
}

/*
* Try to append the entry to raft servers' log and return result.
* Result is valid if raft servers apply this entry before timeout.
*/

func (kv *KVServer) appendEntryToLog(entry Op) Result {
	index, _, isLeader := kv.rf.Start(entry)
	if !isLeader {
		return Result{OK: false}
	}

	kv.mu.Lock()
	if _, ok := kv.resultCh[index]; !ok {
		kv.resultCh[index] = make(chan Result, 1)
	}
	kv.mu.Unlock()

	select {
	case result := <-kv.resultCh[index]:
		if isMatch(entry, result) {
			return result
		}
		return Result{OK: false}
	case <-time.After(240 * time.Millisecond):
		return Result{OK: false}
	}
}

/*
 * Check if the result corresponds to the log entry.
 */

func isMatch(entry Op, result Result) bool {
	return entry.ClientId == result.ClientId && entry.RequestId == result.RequestId
}

func (kv *KVServer) Get(args *GetArgs, reply *GetReply) {
	entry := Op{}
	entry.Command = "get"
	entry.ClientId = args.ClientId
	entry.RequestId = args.RequestId
	entry.Key = args.Key

	result := kv.appendEntryToLog(entry)
	if !result.OK {
		reply.WrongLeader = true
		return
	}
	reply.WrongLeader = false
	reply.Err = result.Err
	reply.Value = result.Value
}

func (kv *KVServer) PutAppend(args *PutAppendArgs, reply *PutAppendReply) {
	entry := Op{}
	entry.Command = args.Command
	entry.ClientId = args.ClientId
	entry.RequestId = args.RequestId
	entry.Key = args.Key
	entry.Value = args.Value

	result := kv.appendEntryToLog(entry)
	if !result.OK {
		reply.WrongLeader = true
		return
	}
	reply.WrongLeader = false
	reply.Err = result.Err
}

/*
 * Apply operation on database and return result.
 */

func (kv *KVServer) applyOp(op Op) Result {
	result := Result{}
	result.Command = op.Command
	result.OK = true
	result.WrongLeader = false
	result.ClientId = op.ClientId
	result.RequestId = op.RequestId

	switch op.Command {
	case "put":
		if !kv.isDuplicated(op) {
			kv.data[op.Key] = op.Value
		}
		result.Err = OK
	case "append":
		if !kv.isDuplicated(op) {
			kv.data[op.Key] += op.Value
		}
		result.Err = OK
	case "get":
		if value, ok := kv.data[op.Key]; ok {
			result.Err = OK
			result.Value = value
		} else {
			result.Err = ErrNoKey
		}
	}
	kv.ack[op.ClientId] = op.RequestId
	return result
}

/*
 * Check if the request is duplicated with request id.
 */
 
func (kv *KVServer) isDuplicated(op Op) bool {
	lastRequestId, ok := kv.ack[op.ClientId]
	if ok {
		return lastRequestId >= op.RequestId
	}
	return false
}

func (kv *KVServer) Kill() {
	kv.rf.Kill()
}

func (kv *KVServer) Run() {
	for {
		msg := <-kv.applyCh
		kv.mu.Lock()
		if msg.UseSnapshot {
			r := bytes.NewBuffer(msg.Snapshot)
			d := gobWrapper.NewDecoder(r)

			var lastIncludedIndex, lastIncludedTerm int
			d.Decode(&lastIncludedIndex)
			d.Decode(&lastIncludedTerm)
			d.Decode(&kv.data)
			d.Decode(&kv.ack)
		} else {
			// apply operation and send result
			op := msg.Command.(Op)
			result := kv.applyOp(op)
			if ch, ok := kv.resultCh[msg.CommandIndex]; ok {
				select {
				case <-ch: // drain bad data
				default:
				}
			} else {
				kv.resultCh[msg.CommandIndex] = make(chan Result, 1)
			}
			kv.resultCh[msg.CommandIndex] <- result

			// create snapshot if raft state exceeds allowed size
			if kv.maxraftstate != -1 && kv.rf.GetRaftStateSize() > kv.maxraftstate {
				w := new(bytes.Buffer)
				e := gobWrapper.NewEncoder(w)
				e.Encode(kv.data)
				e.Encode(kv.ack)
				go kv.rf.CreateSnapshot(w.Bytes(), msg.CommandIndex)
			}
		}
		kv.mu.Unlock()
	}
}

/*
 * Servers[] contains the ports of the set of servers that will cooperate via Raft to
 form the fault-tolerant key/value service.
 * Me is the index of the current server in servers[].
 * The k/v server should store snapshots with persister.SaveSnapshot(), 
 and Raft should save its state (including log) with persister.SaveRaftState().
 * The k/v server should snapshot when Raft's saved state exceeds maxraftstate bytes,
 in order to allow Raft to garbage-collect its log. if maxraftstate is -1, you don't need to snapshot.
 * StartKVServer() must return quickly, so it should start goroutines for any long-running work.
 */

func StartKVServer(servers []*rpc.ClientEnd, me int, persister *raft.Persister, maxraftstate int) *KVServer {
	// call gobWrapper.Register on structures you want
	// Go's RPC library to marshall/unmarshall.
	gobWrapper.Register(Op{})
	gobWrapper.Register(Result{})

	kv := new(KVServer)
	kv.me = me
	kv.maxraftstate = maxraftstate

	kv.applyCh = make(chan raft.ApplyMsg, 100)
	kv.rf = raft.Make(servers, me, persister, kv.applyCh)

	kv.data = make(map[string]string)
	kv.ack = make(map[int64]int64)
	kv.resultCh = make(map[int]chan Result)

	go kv.Run()
	return kv
}
