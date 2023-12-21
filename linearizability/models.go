package linearizability

// KvInput represents the input for a key-value store operation.
// It includes the operation type (get, put, append), key, and value.
type KvInput struct {
	Op    uint8  // Operation type: 0 => get, 1 => put, 2 => append
	Key   string // Key in the key-value store
	Value string // Value to be used in the operation
}

// KvOutput represents the output of a get operation in the key-value store.
type KvOutput struct {
	Value string // Value retrieved from the key-value store
}

// KvModel returns a Model specific to a key-value store. This model can be used
// to check linearizability of operations on a key-value store.
func KvModel() Model {
	return Model{
		// Partition partitions the operations by key. Each key's operations
		// are considered a separate history for linearizability checks.
		Partition: func(history []Operation) [][]Operation {
			m := make(map[string][]Operation)
			for _, v := range history {
				key := v.Input.(KvInput).Key
				m[key] = append(m[key], v)
			}
			var ret [][]Operation
			for _, v := range m {
				ret = append(ret, v)
			}
			return ret
		},
		// Init initializes the model state. For a key-value store model,
		// the state is represented as a string (value of a key).
		Init: func() interface{} {
			// Note: This model represents a single key's value; partitioning by key makes this valid.
			return ""
		},
		// Step defines how the model transitions from one state to another
		// given an input and an expected output.
		Step: func(state, input, output interface{}) (bool, interface{}) {
			inp := input.(KvInput)
			out := output.(KvOutput)
			st := state.(string)
			switch inp.Op {
			case 0: // get operation
				return out.Value == st, state
			case 1: // put operation
				return true, inp.Value
			case 2: // append operation
				return true, st + inp.Value
			}
			// Default case: should not happen in correct usage
			return false, state
		},
		// Equal defines how to determine if two states of the model are equal.
		Equal: ShallowEqual,
	}
}
