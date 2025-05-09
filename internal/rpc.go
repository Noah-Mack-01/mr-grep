package internal

//
// RPC definitions.
//
// remember to capitalize all names.
//

import (
	"os"
	"strconv"
)

type WorkerCallArgs struct {
	WorkerId int
}

// new Task response from Coordinator
type TaskResponseArgs struct {
	Job     Job
	JobType string
}

type Job struct {
	InputFilePath, Status                   string
	WorkerId, JobId, NReduce, MapPartitions int
}

// Cook up a unique-ish UNIX-domain socket name
// in /var/tmp, for the coordinator.
// Can't use the current directory since
// Athena AFS doesn't support UNIX-domain sockets.
func coordinatorSock() string {
	s := "/var/tmp/5840-mr-"
	s += strconv.Itoa(os.Getuid())
	return s
}
