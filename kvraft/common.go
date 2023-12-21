package raftkv

// Constants defining possible error states.
const (
	OK       = "OK"       // Indicates successful operation.
	ErrNoKey = "ErrNoKey" // Indicates that the requested key does not exist in the key-value store.
)

// Err is a custom type representing an error string.
type Err string

// PutAppendArgs defines the arguments structure for Put and Append operations.
type PutAppendArgs struct {
	Key       string // Key in the key-value store.
	Value     string // Value to be associated with the key.
	Command   string // Operation type: "Put" or "Append".
	ClientId  int64  // Unique client identifier to differentiate requests.
	RequestId int64  // Unique request identifier for idempotency.
}

// PutAppendReply defines the reply structure for Put and Append operations.
type PutAppendReply struct {
	WrongLeader bool // Flag to indicate if the operation reached a non-leader server.
	Err         Err  // Error status of the operation.
}

// GetArgs defines the arguments structure for Get operation.
type GetArgs struct {
	Key       string // Key to retrieve from the key-value store.
	ClientId  int64  // Unique client identifier.
	RequestId int64  // Unique request identifier.
}

// GetReply defines the reply structure for Get operation.
type GetReply struct {
	WrongLeader bool   // Flag to indicate if the operation reached a non-leader server.
	Err         Err    // Error status of the operation.
	Value       string // The value retrieved for the key, if any.
}
