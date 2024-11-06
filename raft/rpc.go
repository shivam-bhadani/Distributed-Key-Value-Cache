package raft

type RequestVoteArgs struct {
	Term         int
	CandidateId  string
	LastLogIndex int
	LastLogTerm  int
}

type RequestVoteReply struct {
	Term        int
	VoteGranted bool
}

type AppendEntriesArgs struct {
	Term         int
	LeaderId     string
	PrevLogIndex int
	PrevLogTerm  int
	Entries      []LogEntry
	LeaderCommit int
}

type AppendEntriesReply struct {
	Term    int
	Success bool
}

func (n *RaftNode) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) error {
	n.Lock()
	defer n.Unlock()
	if args.Term < n.CurrentTerm {
		reply.Term = n.CurrentTerm
		reply.VoteGranted = false
		return nil
	}
	if args.Term > n.CurrentTerm {
		n.CurrentTerm = args.Term
		n.State = Follower
		n.VotedFor = ""
	}
	lastLogIndex := len(n.Log) - 1
	lastLogTerm := 0
	if lastLogIndex >= 0 {
		lastLogTerm = n.Log[lastLogIndex].Term
	}
	upToDate := args.LastLogIndex > lastLogTerm || (args.LastLogTerm == lastLogTerm && args.LastLogIndex >= lastLogIndex)
	if (n.VotedFor == "" || n.VotedFor == args.CandidateId) && upToDate {
		reply.VoteGranted = true
		n.VotedFor = args.CandidateId
		n.ResetElectionTimer()
	} else {
		reply.VoteGranted = false
	}
	reply.Term = n.CurrentTerm
	return nil
}

func (n *RaftNode) AppendEntries(args *AppendEntriesArgs, reply *AppendEntriesReply) error {
	n.Lock()
	defer n.Unlock()
	reply.Success = false
	reply.Term = n.CurrentTerm
	if args.Term < n.CurrentTerm {
		return nil
	}
	if args.Term > n.CurrentTerm {
		n.CurrentTerm = args.Term
		n.State = Follower
		n.VotedFor = ""
	}
	n.ResetElectionTimer()
	if args.PrevLogIndex >= 0 && n.Log[args.PrevLogIndex].Term != args.PrevLogTerm {
		return nil
	}

	// Append new entries, removing conflicting entries
	n.Log = n.Log[:args.PrevLogIndex+1]
	n.Log = append(n.Log, args.Entries...)

	// Update the commit index
	if args.LeaderCommit > n.CommitIndex {
		n.CommitIndex = min(args.LeaderCommit, len(n.Log)-1)
	}
	reply.Success = true
	return nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
