package raft

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"
)

// // Debugging
// const Debug = true

// func DPrintf(format string, a ...interface{}) (n int, err error) {
// 	if Debug {
// 		log.Printf(format, a...)
// 	}
// 	return
// }

// Retrieve the verbosity level from an environment variable
func getVerbosity() int {
	v := os.Getenv("VERBOSE")
	level := 0
	if v != "" {
		var err error
		level, err = strconv.Atoi(v)
		if err != nil {
			log.Fatalf("Invalid verbosity %v", v)
		}
	}
	return level
}

type LogTopic string

const (
	dClient    LogTopic = "CLNT"
	dCommit    LogTopic = "CMIT"
	dDrop      LogTopic = "DROP"
	dError     LogTopic = "ERRO"
	dInfo      LogTopic = "INFO"
	dLeader    LogTopic = "LEAD"
	dCandidate LogTopic = "CAND"
	dLog       LogTopic = "LOG1"
	dLog2      LogTopic = "LOG2"
	dPersist   LogTopic = "PERS"
	dSnap      LogTopic = "SNAP"
	dTerm      LogTopic = "TERM"
	dTest      LogTopic = "TEST"
	dTimer     LogTopic = "TIMR"
	dTrace     LogTopic = "TRCE"
	dVote      LogTopic = "VOTE"
	dWarn      LogTopic = "WARN"
)

var debugStart time.Time
var debugVerbosity int

func init() {
	debugVerbosity = getVerbosity()
	debugStart = time.Now()

	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))
}

func PrettyDebug(topic LogTopic, format string, a ...interface{}) {
	if debugVerbosity >= 1 {
		t := time.Since(debugStart).Microseconds()
		t /= 100
		prefix := fmt.Sprintf("%06d %v ", t, string(topic))
		format = prefix + format
		log.Printf(format, a...)
	}
}

type Log struct {
	raft *Raft
}

func DPrintf(template string, rf *Raft, args ...interface{}) {
	prefix := fmt.Sprintf("S%d ,T%d ,votes: %d, role:%v rf.votedFor:%v  len(rf.peers):%v ", 
	rf.me, rf.currentTerm, rf.votes, rf.role, rf.votedFor, len(rf.peers))
	template = prefix + template + "\n"
	PrettyDebug(dTimer, template, args...)
}

func Infof(template string, rf *Raft, args ...interface{}) {
	// prefix := fmt.Sprintf("S%d ,T%d ", l.raft.me, l.raft.currentTerm)
	// template = prefix + template + "\n"
	// PrettyDebug(dLog, template, args)
	prefix := fmt.Sprintf("S%d ,T%d ,votes: %d, role:%v rf.votedFor:%v  len(rf.peers):%v ", rf.me, rf.currentTerm, rf.votes, rf.role, rf.votedFor, len(rf.peers))
	template = prefix + template + "\n"
	PrettyDebug(dLog, template, args...)
}

func Errorf(template string, rf *Raft, args ...interface{}) {
	// prefix := fmt.Sprintf("S%d ,T%d ", l.raft.me, l.raft.currentTerm)
	// template = prefix + template + "\n"
	// PrettyDebug(dTrace, template, args)
	prefix := fmt.Sprintf("S%d ,T%d ,votes: %d, role:%v rf.votedFor:%v  len(rf.peers):%v ", rf.me, rf.currentTerm, rf.votes, rf.role, rf.votedFor, len(rf.peers))
	template = prefix + template + "\n"
	PrettyDebug(dTrace, template, args...)
}

// func (l *Log) Printf(topic LogTopic, template string) {
// 	prefix := fmt.Sprintf("S%d ,T%d ", l.raft.me, l.raft.currentTerm)
// 	template = prefix + template + "\n"
// 	PrettyDebug(topic, template)
// }

func MakeLogger(raft *Raft) Log {
	r := Log{raft: raft}
	return r
}
