package raft

import "sync"

// Persister is used to store and manage the persistent state of Raft and kvraft.
type Persister struct {
	mu        sync.Mutex // Mutex for protecting concurrent access to the state
	raftstate []byte     // Byte slice to store Raft's persistent state (like log entries)
	snapshot  []byte     // Byte slice to store a snapshot of the key-value server's state
}

// MakePersister creates and returns a new Persister instance.
func MakePersister() *Persister {
	return &Persister{}
}

// Copy creates a deep copy of the current Persister's state.
func (ps *Persister) Copy() *Persister {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	np := MakePersister()
	np.raftstate = ps.raftstate // Copy Raft state
	np.snapshot = ps.snapshot   // Copy snapshot
	return np
}

// SaveRaftState saves the given Raft state into the Persister.
func (ps *Persister) SaveRaftState(state []byte) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.raftstate = state
}

// ReadRaftState returns the current Raft state stored in the Persister.
func (ps *Persister) ReadRaftState() []byte {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return ps.raftstate
}

// RaftStateSize returns the size of the stored Raft state.
func (ps *Persister) RaftStateSize() int {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return len(ps.raftstate)
}

// SaveStateAndSnapshot atomically saves both the Raft state and the key-value server snapshot.
// This helps ensure that the Raft state and the snapshot do not get out of sync.
func (ps *Persister) SaveStateAndSnapshot(state []byte, snapshot []byte) {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	ps.raftstate = state   // Save Raft state
	ps.snapshot = snapshot // Save snapshot
}

// ReadSnapshot returns the current snapshot of the key-value server's state stored in the Persister.
func (ps *Persister) ReadSnapshot() []byte {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return ps.snapshot
}

// SnapshotSize returns the size of the stored snapshot.
func (ps *Persister) SnapshotSize() int {
	ps.mu.Lock()
	defer ps.mu.Unlock()
	return len(ps.snapshot)
}
