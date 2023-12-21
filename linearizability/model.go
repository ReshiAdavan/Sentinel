package linearizability

// Operation represents an operation in the history of a linearizability check.
// It includes both the input to and output from the operation along with their respective timestamps.
type Operation struct {
	Input  interface{} // Input of the operation.
	Call   int64       // Invocation time of the operation.
	Output interface{} // Output of the operation.
	Return int64       // Response time of the operation.
}

// EventKind is a type to distinguish between call and return events.
type EventKind bool

// Constants for differentiating call and return events.
const (
	CallEvent   EventKind = false // Represents a call event.
	ReturnEvent EventKind = true  // Represents a return event.
)

// Event represents either a call or a return event in the history.
type Event struct {
	Kind  EventKind  // Kind of the event (CallEvent or ReturnEvent).
	Value interface{} // Value associated with the event.
	Id    uint       // Unique identifier for the event.
}

// Model defines the structure of the system being checked for linearizability.
// It includes functions for partitioning the history, initializing the system state,
// stepping through the operations, and comparing system states.
type Model struct {
	// Partition functions divide the history into parts, each of which must be linearizable.
	Partition      func(history []Operation) [][]Operation
	PartitionEvent func(history []Event) [][]Event

	// Init initializes the system's state.
	Init func() interface{}

	// Step function takes a state and an operation's input and output,
	// and returns whether the operation is valid in the current state and the new state.
	// It should not mutate the existing state.
	Step func(state interface{}, input interface{}, output interface{}) (bool, interface{})

	// Equal function defines equality for states.
	Equal func(state1, state2 interface{}) bool
}

// NoPartition is a default partitioning function that treats the entire history as a single partition.
func NoPartition(history []Operation) [][]Operation {
	return [][]Operation{history}
}

// NoPartitionEvent is a default partitioning function for event histories, treating the entire history as a single partition.
func NoPartitionEvent(history []Event) [][]Event {
	return [][]Event{history}
}

// ShallowEqual is a default equality function that checks for basic equality between two states.
func ShallowEqual(state1, state2 interface{}) bool {
	return state1 == state2
}
