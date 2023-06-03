package raft

//
// this is an outline of the API that raft must expose to
// the service (or tester). see comments below for
// each of these functions for more details.
//
// rf = Make(...)
//   create a new Raft server.
// rf.Start(command interface{}) (index, term, isleader)
//   start agreement on a new log entry
// rf.GetState() (term, isLeader)
//   ask a Raft for its current term, and whether it thinks it is leader
// ApplyMsg
//   each time a new entry is committed to the log, each Raft peer
//   should send an ApplyMsg to the service (or tester)
//   in the same server.
//

import (
	//	"bytes"

	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	//	"6.824/labgob"
	"6.824/labrpc"
)

// as each Raft peer becomes aware that successive log entries are
// committed, the peer should send an ApplyMsg to the service (or
// tester) on the same server, via the applyCh passed to Make(). set
// CommandValid to true to indicate that the ApplyMsg contains a newly
// committed log entry.
//
// in part 2D you'll want to send other kinds of messages (e.g.,
// snapshots) on the applyCh, but set CommandValid to false for these
// other uses.
type ApplyMsg struct {
	CommandValid bool
	Command      interface{}
	CommandIndex int

	// For 2D:
	SnapshotValid bool
	Snapshot      []byte
	SnapshotTerm  int
	SnapshotIndex int
}

// A Go object implementing a single Raft peer.
type Raft struct {
	mu        sync.Mutex          // Lock to protect shared access to this peer's state
	peers     []*labrpc.ClientEnd // RPC end points of all peers
	persister *Persister          // Object to hold this peer's persisted state
	me        int                 // this peer's index into peers[]
	dead      int32               // set by Kill()

	// Your data here (2A, 2B, 2C).
	// Look at the paper's Figure 2 for a description of what
	// state a Raft server must maintain.

	//2A
	//所有服务器---持久性状态
	role        int // 跟随者、候选者、领导者  0、1、2
	currentTerm int // 服务器已知最新的任期（在服务器首次启动时初始化为0，单调递增）
	votedFor    int // 当前任期内收到选票的 candidateId，如果没有投给任何候选人 则为空
	log         int // 日志条目；每个条目包含了用于状态机的命令，以及领导人接收到该条目时的任期（初始索引为1）
	//所有服务器---易失性状态
	commitIndex int //已知已提交的最高的日志条目的索引（初始值为0，单调递增）
	lastApplied int //已经被应用到状态机的最高的日志条目的索引（初始值为0，单调递增）

	votes             int
	leaderHeartsBeats chan struct{}
	// isLeader bool
	// 领导人---易失性状态
	nextIndex  []int //对于每一台服务器，发送到该服务器的下一个日志条目的索引（初始值为领导人最后的日志条目的索引+1）
	matchIndex []int //对于每一台服务器，已知的已经复制到该服务器的最高日志条目的索引（初始值为0，单调递增）

}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {
	var term int
	var isleader bool
	// Your code here (2A).
	rf.mu.Lock()
	term = rf.currentTerm
	// isleader = rf.isLeader
	// 自己会给自己投票
	if rf.role == 2 {
		isleader = true
	}
	rf.mu.Unlock()
	return term, isleader
}

// save Raft's persistent state to stable storage,
// where it can later be retrieved after a crash and restart.
// see paper's Figure 2 for a description of what should be persistent.
func (rf *Raft) persist() {
	// Your code here (2C).
	// Example:
	// w := new(bytes.Buffer)
	// e := labgob.NewEncoder(w)
	// e.Encode(rf.xxx)
	// e.Encode(rf.yyy)
	// data := w.Bytes()
	// rf.persister.SaveRaftState(data)
}

// restore previously persisted state.
func (rf *Raft) readPersist(data []byte) {
	if data == nil || len(data) < 1 { // bootstrap without any state?
		return
	}
	// Your code here (2C).
	// Example:
	// r := bytes.NewBuffer(data)
	// d := labgob.NewDecoder(r)
	// var xxx
	// var yyy
	// if d.Decode(&xxx) != nil ||
	//    d.Decode(&yyy) != nil {
	//   error...
	// } else {
	//   rf.xxx = xxx
	//   rf.yyy = yyy
	// }
}

// A service wants to switch to snapshot.  Only do so if Raft hasn't
// have more recent info since it communicate the snapshot on applyCh.
func (rf *Raft) CondInstallSnapshot(lastIncludedTerm int, lastIncludedIndex int, snapshot []byte) bool {

	// Your code here (2D).

	return true
}

// the service says it has created a snapshot that has
// all info up to and including index. this means the
// service no longer needs the log through (and including)
// that index. Raft should now trim its log as much as possible.
func (rf *Raft) Snapshot(index int, snapshot []byte) {
	// Your code here (2D).

}

// example RequestVote RPC arguments structure.
// field names must start with capital letters!A
type RequestVoteArgs struct {
	// Your data here (2A, 2B).
	Term         int //候选人的任期号
	CandidateId  int // 请求选票的候选人的 ID
	LastLogIndex int //候选人的最后日志条目的索引值
	LastLogTerm  int //候选人最后日志条目的任期号

}

// example RequestVote RPC reply structure.
// field names must start with capital letters!
type RequestVoteReply struct {
	// Your data here (2A).
	Term        int  //当前任期号，以便于候选人去更新自己的任期号
	VoteGranted bool //候选人赢得了此张选票时为真
}

// example RequestVote RPC handler.
func (rf *Raft) RequestVote(args *RequestVoteArgs, reply *RequestVoteReply) {
	// Your code here (2A, 2B).

	//2A
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if rf.role != 0 {
		// 跟随者才能投票
		reply.VoteGranted = false
		DPrintf("me:%v, role:%v, 收到%v投票请求, votes:%v, 拒绝投票:我不是跟随者",
			rf.me, rf.role, args.CandidateId, rf.votes)
		return
	}
	// 当前任期比请求的任期大,直接忽略
	if rf.currentTerm > args.Term {
		DPrintf("me:%v, role:%v, 收到%v投票请求, votes:%v, 拒绝投票",
			rf.me, rf.role, args.CandidateId, rf.votes)
		reply.Term = rf.currentTerm
		reply.VoteGranted = false
		return
	}
	// 已经投过票了
	if rf.currentTerm == args.Term {
		if rf.votedFor != -1 {
			DPrintf("me:%v, role:%v, 收到%v投票请求, votes:%v, 拒绝投票",
				rf.me, rf.role, args.CandidateId, rf.votes)
			reply.Term = rf.currentTerm
			reply.VoteGranted = false
			return
		}
		// 投票给你
		rf.votedFor = args.CandidateId
		reply.Term = rf.currentTerm
		reply.VoteGranted = true
		DPrintf("me:%v, role:%v, 收到%v投票请求, votes:%v, 【给出】投票",
			rf.me, rf.role, args.CandidateId, rf.votes)
		return
	}
	// rf.currentTerm < args.term
	rf.votes = 0
	rf.currentTerm = args.Term
	// 投票给你
	rf.votedFor = args.CandidateId
	reply.Term = rf.currentTerm
	reply.VoteGranted = true
	DPrintf("me:%v, role:%v, 收到%v投票请求, votes:%v, 【给出】投票",
		rf.me, rf.role, args.CandidateId, rf.votes)
	return
}

// example code to send a RequestVote RPC to a server.
// server is the index of the target server in rf.peers[].
// expects RPC arguments in args.
// fills in *reply with RPC reply, so caller should
// pass &reply.
// the types of the args and reply passed to Call() must be
// the same as the types of the arguments declared in the
// handler function (including whether they are pointers).
//
// The labrpc package simulates a lossy network, in which servers
// may be unreachable, and in which requests and replies may be lost.
// Call() sends a request and waits for a reply. If a reply arrives
// within a timeout interval, Call() returns true; otherwise
// Call() returns false. Thus Call() may not return for a while.
// A false return can be caused by a dead server, a live server that
// can't be reached, a lost request, or a lost reply.
//
// Call() is guaranteed to return (perhaps after a delay) *except* if the
// handler function on the server side does not return.  Thus there
// is no need to implement your own timeouts around Call().
//
// look at the comments in ../labrpc/labrpc.go for more details.
//
// if you're having trouble getting RPC to work, check that you've
// capitalized all field names in structs passed over RPC, and
// that the caller passes the address of the reply struct with &, not
// the struct itself.
func (rf *Raft) sendRequestVote(server int, args *RequestVoteArgs, reply *RequestVoteReply) bool {
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	return ok
}

// the service using Raft (e.g. a k/v server) wants to start
// agreement on the next command to be appended to Raft's log. if this
// server isn't the leader, returns false. otherwise start the
// agreement and return immediately. there is no guarantee that this
// command will ever be committed to the Raft log, since the leader
// may fail or lose an election. even if the Raft instance has been killed,
// this function should return gracefully.
//
// the first return value is the index that the command will appear at
// if it's ever committed. the second return value is the current
// term. the third return value is true if this server believes it is
// the leader.
func (rf *Raft) Start(command interface{}) (int, int, bool) {
	index := -1
	term := -1
	isLeader := true

	// Your code here (2B).

	return index, term, isLeader
}

// the tester doesn't halt goroutines created by Raft after each test,
// but it does call the Kill() method. your code can use killed() to
// check whether Kill() has been called. the use of atomic avoids the
// need for a lock.
//
// the issue is that long-running goroutines use memory and may chew
// up CPU time, perhaps causing later tests to fail and generating
// confusing debug output. any goroutine with a long-running loop
// should call killed() to check whether it should stop.
func (rf *Raft) Kill() {
	atomic.StoreInt32(&rf.dead, 1)
	// Your code here, if desired.
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

// 2A 心跳RPC struct
type HeartsBeatsMsg struct {
	Term        int //候选人的任期号
	CandidateId int // 请求选票的候选人的 ID
}
type Empty struct {
}

// 2A 心跳RPC
func (rf *Raft) sendHeartsBeatsMsg(server int, args *HeartsBeatsMsg, empty *Empty) bool {
	ok := rf.peers[server].Call("Raft.DealHeartsBeatsMsg", args, empty)
	return ok
}
func (rf *Raft) DealHeartsBeatsMsg(args *HeartsBeatsMsg, empty *Empty) {
	if _, isLeader := rf.GetState(); isLeader {
		if args.Term < rf.currentTerm {
			return
		}
	}
	rf.mu.Lock()
	defer rf.mu.Unlock()
	// 1、不是leader
	// 2、是leader但是currentTerm小于另一个leader
	rf.role = 0
	rf.votes = 0
	rf.currentTerm = args.Term
	rf.votedFor = args.CandidateId
	rf.leaderHeartsBeats <- struct{}{}
	DPrintf("me:%v, 接收到leader:%v 的心跳,同步term,变为追随者【【【【【【【【【【", rf.me, args.CandidateId)
}

// 2A 心跳机制
func (rf *Raft) heartsbeats() {
	for rf.killed() == false {
		time.Sleep(10 * time.Millisecond)
		if _, isLeader := rf.GetState(); !isLeader {
			continue
		}
		rf.mu.Lock()
		length := len(rf.peers)
		rf.mu.Unlock()
		for i := 0; i < length; i++ {
			rf.mu.Lock()
			args := HeartsBeatsMsg{
				Term:        rf.currentTerm,
				CandidateId: rf.me,
			}
			rf.mu.Unlock()
			if i != rf.me {
				// DPrintf("接收到leader的心跳, leader:%v,【【【【【【【【【【", rf.me)
				rf.sendHeartsBeatsMsg(i, &args, &Empty{})
			}

		}
	}
	if rf.killed() {
		DPrintf("%v: is killed -----------===================-------------", rf.me)
	}
}

// 如果这个peer最近没有收到心跳，那么ticker go例程将开始新的选举。
// The ticker go routine starts a new election if this peer hasn't received
// heartsbeats recently.
func (rf *Raft) ticker() {
	for rf.killed() == false {
		// Your code here to check if a leader election should
		// be started and to randomize sleeping time using
		// time.Sleep().
		rand.Seed(time.Now().UnixMilli())

		select {
		case <-rf.leaderHeartsBeats:
			{

			}
		case <-time.Tick(time.Duration(150+rand.Intn(200)) * time.Millisecond):
			{
				rf.mu.Lock()
				if rf.role == 0 {
					// 跟随者 -> 候选人
					rf.currentTerm++
					rf.role = 1
					// 选择自己
					rf.votes = 1
					rf.votedFor = rf.me
					DPrintf("me:%v, role:%v, 跟随者 -> 候选人 votes:%v",
						rf.me, rf.role, rf.votes)
					// 发出投票
					length := len(rf.peers)
					rf.mu.Unlock()
					for i := 0; i < length; i++ {
						rf.mu.Lock()
						role := rf.role
						rf.mu.Unlock()
						if role != 1 {
							break
						}
						if i != rf.me {
							rf.mu.Lock()
							args := RequestVoteArgs{
								Term:        rf.currentTerm,
								CandidateId: rf.me,
							}
							reply := RequestVoteReply{}
							rf.mu.Unlock()
							rf.sendRequestVote(i, &args, &reply)
							///===================
							rf.mu.Lock()
							if reply.VoteGranted {
								// 成功获取选票
								rf.votes++
								if rf.votes > len(rf.peers)/2 {
									rf.role = 2
									DPrintf("me:%v, role:%v, ===========成为leader============, votes:%v",
										rf.me, rf.role, rf.votes)
								}
								DPrintf("me:%v, role:%v, 【收到】投票, votes:%v",
									rf.me, rf.role, rf.votes)
							} else {
								// 获得选票失败 : 候选人->跟随者
								rf.votedFor = -1
								rf.votes = 0
								// rf.currentTerm = reply.term
								if rf.currentTerm < reply.Term {
									rf.currentTerm = reply.Term
									// 恢复跟随者身份
									rf.role = 0
									DPrintf("me:%v, role:%v, 【没有收到】投票&&currentTerm 太小了-》恢复跟随者身份",
										rf.me, rf.role)
								}
							}
							rf.mu.Unlock()
						}
					}
					rf.mu.Lock()
					if rf.role != 2 {
						rf.role = 0
						rf.votedFor = -1
						rf.votes = 0
						DPrintf("me:%v, 【竞争失败】-》恢复跟随者身份", rf.me)
						// time.Sleep(time.Duration(rand.Intn(10)) * time.Millisecond)
					}
					rf.mu.Unlock()
					continue
				}
				rf.mu.Unlock()
			}
		}

	}
}

// 服务或测试人员想要创建一个Raft服务器。
// 所有Raft服务器的端口(包括这个)都在peer[]中。
// 该服务器的端口是peers[me]。
// 所有 servers' peers[]数组都有相同的顺序。
// Persister是此服务器保存其持久状态的地方，
// 并且最初还保存最近保存的状态(如果有的话)。
// applyCh是一个通道，测试人员或服务希望Raft在该通道上发送ApplyMsg消息。
// Make()必须快速返回，因此它应该为任何长时间运行的工作启动例程。

// the service or tester wants to create a Raft server. the ports
// of all the Raft servers (including this one) are in peers[]. this
// server's port is peers[me]. all the servers' peers[] arrays
// have the same order. persister is a place for this server to
// save its persistent state, and also initially holds the most
// recent saved state, if any. applyCh is a channel on which the
// tester or service expects Raft to send ApplyMsg messages.
// Make() must return quickly, so it should start goroutines
// for any long-running work.
func Make(peers []*labrpc.ClientEnd, me int,
	persister *Persister, applyCh chan ApplyMsg) *Raft {
	rf := &Raft{
		peers:     peers,
		persister: persister,
		me:        me,
		//2A
		votedFor:          -1,
		role:              0,
		leaderHeartsBeats: make(chan struct{}),
	}

	// Your initialization code here (2A, 2B, 2C).
	// 2A

	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	// start ticker goroutine to start elections
	go rf.ticker()
	// 心跳机制
	go rf.heartsbeats()
	return rf
}
