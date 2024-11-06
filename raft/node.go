package raft

import (
	"math/rand"
	"net/rpc"
	"sync"
	"time"
)

func NewRaftNode(id string, peers []string) *RaftNode {
	node := &RaftNode{
		Id:              id,
		State: 					 Follower,
		CurrentTerm:     0,
		VotedFor:        -1,
		Log:             []LogEntry{{Term: 0, Index: 0}}, // Genesis Entry
		CommitIndex:     -1,
		LastApplied:     -1,
		Peers:           peers,
		NextIndex:       make(map[string]int),
		MatchIndex:      make(map[string]int),
		VoteCh:          make(chan bool),
		AppendCh:        make(chan bool),
	}

	return node
}

func (n *RaftNode) ResetElectionTimer() {
	if n.ElectionTimer != nil {
		n.ElectionTimer.Stop()
	}
	// Random election timeout between 150-300ms
	timeout := time.Duration(150+rand.Intn(150)) * time.Millisecond
	n.ElectionTimer = time.AfterFunc(timeout, func() {
			n.startElection()
	})
}

func (n *RaftNode) startElection() {
	n.Lock()
	if n.State == Leader {
		n.Unlock()
		return
	}
	n.State = Candidate
	n.CurrentTerm++
	n.VotedFor = n.Id
	currentTerm := n.CurrentTerm

	// Prepare RequestVote rpc
	args := &RequestVoteArgs{
		Term: currentTerm,
		CandidateId: n.Id,
		LastLogIndex: len(n.Log) - 1,
		LastLogTerm: -1,
	}
	if args.LastLogIndex >= 0 {
		args.LastLogTerm = n.Log[args.LastLogIndex].Term
	}
	n.Unlock()

	// Send RequestVote RPCs to all peers
	votes := 1 // Vote for self
	var voteMutex sync.Mutex

	for _, peer := n.Peers {
		if peer == n.Id {
			continue
		}
		go func(peer string) {
			var reply RequestVoteReply
			client, err := rpc.Dial("tcp", peer)
			if err != nil {
				return
			}
			defer client.Close()
			if err := client.Call("RaftNote.RequestVote", args, &reply); err != nil {
				return
			}
			n.Lock()
			if reply.Term > n.CurrentTerm {
				n.CurrentTerm = reply.Term
				n.State = Follower
				n.VotedFor = ""
				n.Unlock()
				return
			}
			n.Unlock()
			if reply.VoteGranted {
				voteMu.Lock()
				votes++
				if votes > len(n.peers)/2 && n.state == Candidate {
						n.Lock()
						if n.state == Candidate {
								n.state = Leader
								for _, p := range n.peers {
										n.nextIndex[p] = len(n.log)
										n.matchIndex[p] = -1
								}
								n.startHeartbeat()
						}
						n.Unlock()
				}
				voteMu.Unlock()
			}
		}(peer)
	}
}

func (n *RaftNode) StartHeartbeat() {
	if n.heartbeatTimer != nil {
			n.heartbeatTimer.Stop()
	}
	n.heartbeatTimer = time.AfterFunc(100*time.Millisecond, func() {
			n.sendHeartbeat()
			if n.state == Leader {
					n.StartHeartbeat()
			}
	})
}

func (n *RaftNode) sendHeartbeat() {
	n.Lock()
	if n.State != Leader {
		n.Unlock()
		return
	}
	for _, peer := range n.peers {
		if peer == n.id {
				continue
		}
		go func(peer string) {
			for {
					n.Lock()
					prevLogIndex := n.nextIndex[peer] - 1
					prevLogTerm := 0
					if prevLogIndex >= 0 {
							prevLogTerm = n.log[prevLogIndex].Term
					}

					entries := n.log[n.nextIndex[peer]:]
					args := AppendEntriesArgs{
							Term:         n.currentTerm,
							LeaderId:    	n.id,
							PrevLogIndex: prevLogIndex,
							PrevLogTerm:  prevLogTerm,
							Entries:      entries,
							LeaderCommit: n.commitIndex,
					}
					n.Unlock()
					var reply AppendEntriesReply
					client, err := rpc.DialHTTP("tcp", peer)
					if err != nil {
							return
					}
					defer client.Close()
					if err := client.Call("RaftNode.AppendEntries", args, &reply); err != nil {
							return
					}
					n.Lock()
					if reply.Term > n.currentTerm {
							n.currentTerm = reply.Term
							n.state = Follower
							n.votedFor = ""
							n.Unlock()
							return
					}
					if reply.Success {
						n.nextIndex[peer] = prevLogIndex + len(entries) + 1
						n.matchIndex[peer] = n.nextIndex[peer] - 1
						n.Unlock()
						break
					} else {
						if n.nextIndex[peer] > 0 {
								n.nextIndex[peer]--
						}
						n.Unlock()
					}
			}
		}(peer)
	}
	n.Unlock()
}
