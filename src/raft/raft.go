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

	"log"
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

type logEntry struct {
	Index   int //索引
	Term    int
	Command interface{}
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
	role        int        // 跟随者、候选者、领导者  0、1、2
	currentTerm int        // 服务器已知最新的任期（在服务器首次启动时初始化为0，单调递增）
	votedFor    int        // 当前任期内收到选票的 candidateId，如果没有投给任何候选人 则为空
	log         []logEntry // 日志条目；每个条目包含了用于状态机的命令，以及领导人接收到该条目时的任期（初始索引为1）
	//所有服务器---易失性状态
	commitIndex int //已知已提交的最高的日志条目的索引（初始值为0，单调递增）
	lastApplied int //已经被应用到状态机的最高的日志条目的索引（初始值为0，单调递增）

	votes             int
	leaderHeartsBeats chan struct{}
	followerOvertime  chan struct{} //跟随者身份续期
	applyCh           chan ApplyMsg
	// isLeader bool
	// 领导人---易失性状态
	nextIndex  map[int]int //对于每一台服务器，发送到该服务器的下一个日志条目的索引（初始值为领导人最后的日志条目的索引+1）
	matchIndex map[int]int //对于每一台服务器，已知的已经复制到该服务器的最高日志条目的索引（初始值为0，单调递增）

}

// return currentTerm and whether this server
// believes it is the leader.
func (rf *Raft) GetState() (int, bool) {
	var term int
	var isleader bool
	// Your code here (2A).
	rf.mu.Lock()
	term = rf.currentTerm
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
	if rf.killed() {
		Errorf("already killed...........", rf)
		return
	}
	//2A
	rf.mu.Lock()
	defer rf.mu.Unlock()
	if rf.role != 0 {
		// 跟随者才能投票
		reply.VoteGranted = false
		Infof(" 收到%d投票请求, votes:%d, 拒绝投票:我不是跟随者",
			rf, args.CandidateId, rf.votes)
		return
	}
	// 当前任期比请求的任期大,直接忽略
	if rf.currentTerm > args.Term {
		Infof(" 收到%d投票请求, votes:%d, 拒绝投票",
			rf, args.CandidateId, rf.votes)
		reply.Term = rf.currentTerm
		reply.VoteGranted = false
		return
	}
	// 已经投过票了
	if rf.currentTerm == args.Term {
		if rf.votedFor != -1 {
			Infof(" 收到%d投票请求, votes:%d, 拒绝投票||||已经投过票了",
				rf, args.CandidateId, rf.votes)
			reply.Term = rf.currentTerm
			reply.VoteGranted = false
			return
		}
		// 投票给你
		rf.votedFor = args.CandidateId
		reply.Term = rf.currentTerm
		reply.VoteGranted = true
		Infof("【给出】投票 S%d ---> S%d",
			rf, rf.me, args.CandidateId)
		return
	}
	// rf.currentTerm < args.term
	rf.votes = 0
	rf.currentTerm = args.Term
	// 投票给你
	rf.votedFor = args.CandidateId
	reply.Term = rf.currentTerm
	reply.VoteGranted = true
	go func() {
		rf.followerOvertime <- struct{}{}
	}()
	Infof("【给出】投票 S%d ---> S%d",
		rf, rf.me, args.CandidateId)
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
	if rf.killed() {
		Errorf("already killed...........", rf)
		return false
	}
	Infof("【Vote】 call() S%d -----> S%d ", rf, rf.me, server)
	ok := rf.peers[server].Call("Raft.RequestVote", args, reply)
	if !ok {
		Errorf("【Vote】 call() 【res】 S%d --×××××--> S%d ", rf, rf.me, server)
	} else {
		Infof("【Vote】 call()  【res】 S%d --√√√√√√--> S%d ", rf, rf.me, server)
	}
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
	if rf.killed() {
		Errorf("already killed...........", rf)
		return index, term, false
	}
	term, isLeader = rf.GetState()
	if isLeader {
		rf.mu.Lock()
		index = len(rf.log)
		rf.log = append(rf.log, logEntry{
			Index:   index,
			Term:    term,
			Command: command,
		})
		rf.mu.Unlock()
		Infof("Start().. command追加,当前 rf.log:%v commitIndex: %v   command内容: %v", rf, rf.log, rf.commitIndex, command)
	}
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
	log.Println("-------------killed-------------", rf)
}

func (rf *Raft) killed() bool {
	z := atomic.LoadInt32(&rf.dead)
	return z == 1
}

// // 2A 心跳RPC struct
//
//	type HeartsBeatsMsg struct {
//		Term        int //候选人的任期号
//		CandidateId int // 请求选票的候选人的 ID
//	}
//
// type Empty struct {
// }
// 2A  2B
// 追加条目 || 也被当做心跳使用
type AppendEntriesArgs struct {
	Term         int        //领导人的任期
	LeaderId     int        //领导人 ID 因此跟随者可以对客户端进行重定向（译者注：跟随者根据领导人
	ID           int        //把客户端的请求重定向到领导人，比如有时客户端把请求发给了跟随者而不是领导人）
	PrevLogIndex int        //紧邻新日志条目之前的那个日志条目的索引
	PrevLogTerm  int        //紧邻新日志条目之前的那个日志条目的任期
	Entries      []logEntry //需要被保存的日志条目（被当做心跳使用时，则日志条目内容为空；为了提高效率可能一次性发送多个）
	LeaderCommit int        //领导人的已知已提交的最高的日志条目的索引
}

type AppendEntriesReply struct {
	Term    int  //当前任期，对于领导人而言 它会更新自己的任期
	Success bool //如果跟随者所含有的条目和 prevLogIndex 以及 prevLogTerm 匹配上了，则为 true
}

// 2A 心跳RPC
func (rf *Raft) sendHeartsBeatsMsg(server int, args *AppendEntriesArgs, reply *AppendEntriesReply) bool {
	if rf.killed() {
		Errorf("already killed...........", rf)
		return false
	}
	Infof("【HeartsBeats】 call()   S%d -----> S%d ", rf, rf.me, server)
	ok := rf.peers[server].Call("Raft.DealHeartsBeatsMsg", args, reply)
	if !ok {
		Errorf("【HeartsBeats】 call()  【res】 S%d --×××××--> S%d ", rf, rf.me, server)
	} else {
		Infof("【HeartsBeats】 call()  【res】 S%d --√√√√√√--> S%d ", rf, rf.me, server)
	}
	return ok
}
func (rf *Raft) DealHeartsBeatsMsg(args *AppendEntriesArgs, reply *AppendEntriesReply) {
	if rf.killed() {
		Errorf("already killed...........", rf)
		return
	}

	term, isLeader := rf.GetState()
	reply.Term = term
	reply.Success = false
	if isLeader {
		// if args.Term < term {
		// 	return
		// }
		if args.Term <= term {
			// leader 且 term > args.Term
			// 无效的心跳
			// reply.Term = -1
			// Errorf("i am leader,无效心跳", rf)
			return
		}
	}
	if args.Term < term {
		// args.Term < term
		// 无效的心跳
		// reply.Term = -1
		Errorf("任期太小,拒绝.....", rf)
		return
	}
	rf.mu.Lock()
	// 1、不是leader
	// 2、是leader但是currentTerm小于另一个leader
	// DPrintf("%d", args.Term == rf.currentTerm)
	rf.role = 0
	rf.votes = 0
	rf.currentTerm = args.Term
	rf.votedFor = args.LeaderId
	rf.mu.Unlock()
	go func(rf *Raft) {
		rf.leaderHeartsBeats <- struct{}{}
		// DPrintf(" 接收到leader:%d 的心跳,同步term,变为追随者【【【【【【【【【【", rf, args.CandidateId)
	}(rf)
	rf.mu.Lock()
	//  log 处理
	prevLogIndex := args.PrevLogIndex
	prevLogTerm := args.PrevLogTerm
	flag := false
	expectLogLength := 0
	if len(args.Entries) != 0 {
		if len(rf.log) == 0 {
			flag = true
		} else {
			for i := 0; i < len(rf.log) && !flag && !rf.killed(); i++ {
				if rf.log[i].Index == prevLogIndex && rf.log[i].Term == prevLogTerm {
					flag = true
					expectLogLength = i + 1
					// Infof("--------- %v ", rf, expectLogLength)
					break
				}
			}
		}

	}

	if flag {
		//  记录日志 || 验证日志
		rf.log = append(rf.log[:expectLogLength], args.Entries...)
		// rf.commitIndex = prevLogIndex + 1
		Infof("add log %v ", rf, args.Entries)
		// reply
		reply.Term = rf.currentTerm
		reply.Success = true
	}
	// commitIndex 更新
	rf.commitIndex = args.LeaderCommit
	if rf.lastApplied > rf.commitIndex {
		rf.lastApplied = rf.commitIndex - 1
		if rf.lastApplied < 0 {
			rf.lastApplied = 0
		}
	}
	DPrintf("========rf.log:  %v ||  rf.commitIndex:%v", rf, rf.log, rf.commitIndex)
	rf.mu.Unlock()
	rf.applyEntries()
}

// 应用日志
func (rf *Raft) applyEntries() {
	rf.mu.Lock()
	commitIndex := rf.commitIndex
	lastApplied := rf.lastApplied
	rf.mu.Unlock()
	if lastApplied != commitIndex {
		rf.mu.Lock()
		for i := rf.lastApplied + 1; i <= rf.commitIndex && i < len(rf.log) && !rf.killed(); i++ {
			applyMsg := ApplyMsg{
				CommandValid: true,
				Command:      rf.log[i].Command,
				CommandIndex: rf.log[i].Index,
			}
			rf.mu.Unlock()
			rf.applyCh <- applyMsg
			rf.mu.Lock()
			rf.lastApplied = i
			Infof("applied....: %v", rf, applyMsg)
		}
		rf.mu.Unlock()
	}
}
func (rf *Raft) autoApplyEntries() {
	for rf.killed() == false {
		rf.applyEntries()
		time.Sleep(10 * time.Millisecond)
	}
}
func (rf *Raft) heartbeatGroup(wg *sync.WaitGroup, i int) {
	if rf.killed() {
		Errorf("already killed...........", rf)
		return
	}
	defer wg.Wait()
	rf.heartbeat(i)
}

func (rf *Raft) heartbeat(i int) {
	rf.mu.Lock()
	currentIndex := rf.nextIndex[i] // ???????
	prevLogIndex := rf.nextIndex[i] - 1
	if prevLogIndex <= 0 {
		prevLogIndex = 0
	}
	prevLogTerm := 0
	var args AppendEntriesArgs
	if len(rf.log) != 0 && rf.nextIndex[i] > rf.log[len(rf.log)-1].Index {
		args = AppendEntriesArgs{
			Term:         rf.currentTerm,
			LeaderId:     rf.me,
			ID:           rf.me,
			PrevLogIndex: prevLogIndex,
			PrevLogTerm:  prevLogTerm,
			Entries:      make([]logEntry, 0),
			LeaderCommit: rf.commitIndex,
		}
	} else {
		// currentLog := make([]logEntry, 0)
		entries := make([]logEntry, 0)
		for j := 0; j < len(rf.log) && !rf.killed(); j++ {
			// if rf.log[j].Index > prevLogIndex {
			// 	Errorf("%v", rf, rf.log[j])
			// 	break
			// }
			if rf.log[j].Index == prevLogIndex {
				prevLogTerm = rf.log[j].Term
				next := j + 1
				if next < len(rf.log) {
					// length := 10
					// if len(rf.log)-next <= length {
					// 	length = len(rf.log) - next
					// }
					entries = append(entries, rf.log[next])
					// rf.nextIndex[i] = rf.log[len(rf.log)-1].Index
					// currentIndex = rf.nextIndex[i]

					// currentIndex = rf.nextIndex[i]
					// currentLog = append(currentLog, rf.log[currentIndex:]...)
					// currentIndex = rf.log[len(rf.log)-1].Index
				}
				break
			}
		}
		// 找不到再往前的log
		if prevLogTerm == 0 && len(rf.log) > 1 {
			// 只要初始化日志 日志有默认值就不会进来
			// 假设一次最多携带日志长度 length
			// length := 10
			// if len(rf.log) <= length {
			// 	length = len(rf.log)
			// }
			//  rf.log[endLogIndex-1]
			// endLogIndex := length
			// currentLog = append(currentLog, rf.log[1])
			// currentIndex = rf.log[1].Index
		}

		Infof("同步日志,>> S%v   rf.commitIndex :%v    rf.nextIndex[i] :%v  ", rf, i, rf.commitIndex, rf.nextIndex[i])
		args = AppendEntriesArgs{
			Term:         rf.currentTerm,
			LeaderId:     rf.me,
			ID:           rf.me,
			PrevLogIndex: prevLogIndex,
			PrevLogTerm:  prevLogTerm,
			Entries:      entries,
			LeaderCommit: rf.commitIndex,
		}
	}

	rf.mu.Unlock()
	reply := AppendEntriesReply{}
	// DPrintf("发出心跳,】】】】】】】】】", rf)
	rf.sendHeartsBeatsMsg(i, &args, &reply)
	DPrintf("rf.log:%v", rf, rf.log)
	if reply.Success {
		// TODO
		rf.mu.Lock()
		if reply.Term == rf.currentTerm {
			if rf.nextIndex[i] == currentIndex &&
				rf.log[len(rf.log)-1].Index >= rf.nextIndex[i] {
				rf.nextIndex[i]++
			}

		}
		rf.mu.Unlock()
		go rf.applyEntries()
	} else {
		// if reply.Term != 0 {
		rf.mu.Lock()

		if reply.Term > rf.currentTerm {
			// 回归跟随者
			rf.role = 0
			rf.votedFor = -1
			rf.votes = 0
			// rand.Seed(time.Now().UnixMilli())
			// rf.followerTimeOut = time.Now().Add(time.Duration(75+rand.Intn(45)) * time.Millisecond)
			Errorf("【任期太小】-》恢复跟随者身份", rf)
			// } else if reply.Term == -2 {
			// 	// 普通心跳
			// 	// Infof("普通心跳.......", rf)
			// } else if reply.Term == -3 {
			// 	// 普通心跳
			// 	Infof("无效心跳.......", rf)
		} else if reply.Term == rf.currentTerm {
			if rf.nextIndex[i] == currentIndex {
				rf.nextIndex[i]--
			}
			if rf.nextIndex[i] <= 1 {
				rf.nextIndex[i] = 1
			}
		}

		rf.mu.Unlock()
	}

	// 检查 nextIndex
	go func(rf *Raft) {
		rf.mu.Lock()
		defer rf.mu.Unlock()
		checkIndex := rf.commitIndex + 1
		replyCounts := 1
		for peerIndex := 0; peerIndex < len(rf.peers) && !rf.killed(); peerIndex++ {
			if peerIndex == rf.me {
				continue
			}
			if rf.nextIndex[peerIndex] > checkIndex {
				replyCounts++
			}
		}
		if replyCounts > len(rf.peers)/2 {
			if rf.commitIndex < checkIndex {
				rf.commitIndex = checkIndex
			}
			go rf.applyEntries()
		}

	}(rf)

}

// 2A 心跳机制
func (rf *Raft) autoHeartbeats() {
	for rf.killed() == false {
		normalHeartBeats := make(chan struct{}, 0)
		if _, isLeader := rf.GetState(); !isLeader {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		go func(rf *Raft) {
			rf.mu.Lock()
			length := len(rf.peers)
			rf.mu.Unlock()
			wg := sync.WaitGroup{}
			for i := 0; i < length && !rf.killed(); i++ {
				if i != rf.me {
					wg.Add(1)
					go rf.heartbeatGroup(&wg, i)
				}
			}
			wg.Wait()
			normalHeartBeats <- struct{}{}
		}(rf)

		rand.Seed(time.Now().UnixMilli())
		select {
		case <-normalHeartBeats:
			{
				Infof("心跳周期结束", rf)
				time.Sleep(10 * time.Millisecond)
			}
			// 心跳超时处理
		case <-time.Tick(time.Duration(50) * time.Millisecond):
			{
				Errorf("心跳超时", rf)
			}
		}

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
		time.Sleep(time.Duration(150+rand.Intn(150)) * time.Millisecond)
		rand.Seed(time.Now().UnixMilli())
		select {
		case <-rf.followerOvertime:
			{
				Infof("跟随者身份续期", rf)
			}
		case <-rf.leaderHeartsBeats:
		default:
			{
				askForVotesChan := make(chan struct{}, 0)
				leaderChan := make(chan struct{}, 0)
				go func(askForVotesChan chan struct{}, leaderChan chan struct{}) {
					rf.mu.Lock()
					role := rf.role
					rf.mu.Unlock()
					if role == 2 {
						leaderChan <- struct{}{}
						return
					}
					if role == 0 {
						rf.mu.Lock()
						// 跟随者 -> 候选人
						rf.currentTerm++
						rf.role = 1
						// 选择自己
						rf.votes = 1
						rf.votedFor = rf.me
						DPrintf("  跟随者 -> 候选人 votes:%d",
							rf, rf.votes)
						// 发出投票
						length := len(rf.peers)
						rf.mu.Unlock()

						wg := sync.WaitGroup{}
						for i := 0; i < length && !rf.killed(); i++ {
							rf.mu.Lock()
							role := rf.role
							rf.mu.Unlock()
							if role != 1 {
								break
							}
							if i != rf.me {
								wg.Add(1)
								go func(wg *sync.WaitGroup, i int, rf *Raft) {
									defer wg.Done()
									rf.mu.Lock()
									args := RequestVoteArgs{
										Term:        rf.currentTerm,
										CandidateId: rf.me,
									}
									reply := RequestVoteReply{}
									rf.mu.Unlock()

									rf.sendRequestVote(i, &args, &reply)

									rf.mu.Lock()
									if reply.VoteGranted && reply.Term == rf.currentTerm {
										// 成功获取选票
										rf.votes++
										DPrintf("【收到】投票", rf)
										if rf.votes > len(rf.peers)/2 && rf.role != 2 {
											rf.role = 2
											DPrintf("===========成为leader============ ", rf)
										}
									} else {
										// 获得选票失败 : 候选人->跟随者
										// rf.currentTerm = reply.term
										if rf.currentTerm < reply.Term {
											rf.currentTerm = reply.Term
											// 恢复跟随者身份   候选人->跟随者
											rf.role = 0
											rf.votedFor = -1
											rf.votes = 0
											DPrintf("【没有收到】投票&&currentTerm 太小了-》恢复跟随者身份",
												rf)
										}
									}
									rf.mu.Unlock()
								}(&wg, i, rf)

							}
						}
						wg.Wait()
					}
					askForVotesChan <- struct{}{}
				}(askForVotesChan, leaderChan)
				rand.Seed(time.Now().UnixMilli())
				select {
				case <-askForVotesChan:
					{
						rf.mu.Lock()
						if rf.role == 1 {
							Infof("选举正常结束", rf)
							rf.role = 0
							rf.votedFor = -1
							rf.votes = 0
							Errorf("me: %d, 【竞争失败】-》恢复跟随者身份", rf, rf.me)
						}
						rf.mu.Unlock()
					}
				case <-leaderChan:
					// DPrintf("我是leader不需要选举", rf)
				case <-time.Tick(time.Duration(150+rand.Intn(150)) * time.Millisecond):
					{
						rf.mu.Lock()
						if rf.role == 1 {
							rf.role = 0
							rf.votedFor = -1
							rf.votes = 0
							Errorf("me: %d, 【选举超时】-【竞争失败】-》恢复跟随者身份", rf, rf.me)
						}
						rf.mu.Unlock()
					}
				}
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
		leaderHeartsBeats: make(chan struct{}, 1),

		//2B
		lastApplied:      0,
		log:              make([]logEntry, 1),
		nextIndex:        make(map[int]int),
		matchIndex:       make(map[int]int),
		followerOvertime: make(chan struct{}),
		applyCh:          applyCh,
	}

	// Your initialization code here (2A, 2B, 2C).
	// 2A
	// initialize from state persisted before a crash
	rf.readPersist(persister.ReadRaftState())

	// start ticker goroutine to start elections
	go rf.ticker()
	// 心跳机制
	go rf.autoHeartbeats()
	// 应用日志
	go rf.autoApplyEntries()
	return rf
}
