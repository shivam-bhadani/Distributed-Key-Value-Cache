package raft

import (
	"sync"
	"time"
)

type NodeState int

const (
	Follower NodeState = iota
	Candidate
	Leader
)

type LogEntry struct {
	Term    int
	Command interface{}
}

type RaftNode struct {
	sync.Mutex
	Id          string
	State       NodeState
	CurrentTerm int
	VotedFor    string
	Log         []LogEntry

	// Volatile State
	CommitIndex int
	LastApplied int

	// Leader State
	NextIndex  map[string]int
	MatchIndex map[string]int

	// Cluster State
	Peers []string // tcp addresses

	// Channels for communication
	VoteCh   chan bool
	AppendCh chan bool

	// Election timeouts
	ElectionTimer  time.Duration
	HeartbeatTimer time.Duration
}
