package internal

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"net/rpc"
	"os"
	"sync"
	"time"
)

type Coordinator struct {
	tasks        map[int]Job       //
	lastPinged   map[int]time.Time // task ID ->
	currentJob   map[int]int       // job ID worker is currently working on
	nextWorkerId int               // Next Worker ID to assign
	// startTime time.Time // for if we do backups
	mutex   sync.Mutex
	mapJobs int
}

// Update our coordinator state to reflect finalized task if it exists.
// Regardless, assign new task to worker.
func (c *Coordinator) RequestNewWork(stub *WorkerCallArgs, reply *TaskResponseArgs) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	workerId := stub.WorkerId
	if workerId == 0 {
		workerId = c.nextWorkerId
		c.nextWorkerId += 1
	} else {
		jobId, jobOk := c.currentJob[workerId]
		if jobOk {
			var job, ok = c.tasks[jobId]
			if !ok {
				log.Fatal("no Job found!")
			} else if job.Status != "IP" {
				log.Fatalf("Job Not in Progress! %v", job.Status)
			} else {
				// we've validated job is in progress for worker id provided.
				// now reflect worker is unassigned.
				job.Status = "D" // D for DONE
				c.tasks[jobId] = job
				delete(c.currentJob, workerId)
			}
		}
	}
	c.lastPinged[workerId] = time.Now()
	nextJob, jobType, jobId := c.findNextUnfinishedJob()
	if jobType != "WAIT" && jobType != "" {
		nextJob.Status = "IP"
		reply.JobType = jobType
		reply.Job = nextJob
		reply.Job.WorkerId = workerId
		c.tasks[jobId] = reply.Job
		c.currentJob[workerId] = jobId
	} else {
		reply.JobType = jobType
	}
	return nil
}

// returns next job necessary for completion.
func (c *Coordinator) findNextUnfinishedJob() (Job, string, int) {
	var nextJob Job
	var jobType string
	var jobId int
	for id := 0; id < len(c.tasks); id++ { //this is hacky but i doubt we'll r-m > 100
		job, ok := c.tasks[id]
		if !ok {
			continue
		}

		if job.Status == "Ready" {
			if id >= c.mapJobs {
				for i := 0; i < c.mapJobs; i++ {
					if job1 := c.tasks[i]; job1.Status != "D" {
						return Job{}, "WAIT", 0
					}
				}
				jobType = "Reduce"
			} else {
				jobType = "Map"
			}
			nextJob = job
			jobId = id
			break
		}
	}
	return nextJob, jobType, jobId
}

// start a thread that listens for RPCs from worker.go
func (c *Coordinator) server() {
	rpc.Register(c)
	rpc.HandleHTTP()
	//l, e := net.Listen("tcp", ":1234")
	sockname := coordinatorSock()
	os.Remove(sockname)
	l, e := net.Listen("unix", sockname)
	if e != nil {
		log.Fatal("listen error:", e)
	}
	go http.Serve(l, nil)
}

// main/mrcoordinator.go calls Done() periodically to find out
// if the entire job has finished.
func (c *Coordinator) Done() bool {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for k, v := range c.lastPinged {
		if time.Since(v).Seconds() < 10 {
		} else {
			for i, job := range c.tasks {
				if i < c.mapJobs && job.WorkerId == k {
					job.Status = "Ready"
					job.WorkerId = 0
					c.tasks[i] = job
				}
			}
			delete(c.lastPinged, k)
		}
	}

	for _, j := range c.tasks {
		if j.Status != "D" {
			return false
		}
	}
	return true
}

// create a Coordinator.
// main/mrcoordinator.go calls this function.
// nReduce is the number of reduce tasks to use.
func MakeCoordinator(files []string, nReduce int) *Coordinator {
	c := Coordinator{}
	c.tasks = make(map[int]Job)
	c.lastPinged = make(map[int]time.Time)
	c.currentJob = make(map[int]int)
	c.nextWorkerId = 1
	c.mapJobs = len(files)
	index1 := 0
	for index, file := range files {
		c.tasks[index] = Job{InputFilePath: file, Status: "Ready", JobId: index, NReduce: nReduce, MapPartitions: len(files)}
		index1 += 1
	}

	for i := 0; i < nReduce; i++ {
		c.tasks[index1] = Job{
			InputFilePath: fmt.Sprintf("mr-%d-", i),
			Status:        "Ready",
			JobId:         i,
			NReduce:       nReduce,
			MapPartitions: len(files),
		}
		index1 += 1
	}
	c.server()
	return &c
}
