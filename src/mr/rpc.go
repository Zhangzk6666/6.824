package mr

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
)

//
// example to show how to declare the arguments
// and reply for an RPC.
//

// type ExampleArgs struct {
// 	X int
// }

//	type ExampleReply struct {
//		Y int
//	}
type JobType int

const (
	MapJob        JobType = 1
	ReduceJob     JobType = 2
	MapJobDone    JobType = 3
	ReduceJobDone JobType = 4
	Done          JobType = 5
	CreateFiles   JobType = 6
	SleepType     JobType = 7
)

type GetJobRequest struct {
	Type     JobType
	WorkerId int
}

type SendJobResponse struct {
	Type         JobType
	FileNameList []string
	Num          int
	NReduce      int
}

type FinishMapJobNotify struct {
	WorkerId     int
	FileNameList []string
}

type FinishReduceJobNotify struct {
	ReduceNum int
	FileName  string
}
type Empty struct {
}

type DoneStatus struct {
	IsDone bool
}

// Add your RPC definitions here.

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/824-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
