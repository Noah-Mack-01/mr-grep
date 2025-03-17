package mr

import (
	"encoding/json"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"log"
	"net/rpc"
	"os"
	"sort"
	"time"
)

// Map functions return a slice of KeyValue.
type KeyValue struct {
	Key   string
	Value string
}

type ByKey []KeyValue

// for sorting by key.
func (a ByKey) Len() int           { return len(a) }
func (a ByKey) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByKey) Less(i, j int) bool { return a[i].Key < a[j].Key }

// use ihash(key) % NReduce to choose the reduce
// task number for each KeyValue emitted by Map.
func ihash(key string) int {
	h := fnv.New32a()
	h.Write([]byte(key))
	return int(h.Sum32() & 0x7fffffff)
}

// main/mrworker.go calls this function.
func Worker(mapf func(string, string) []KeyValue,
	reducef func(string, []string) string, regex string) {
	// Your worker implementation here.
	// uncomment to send the Example RPC to the coordinator.
	// CallExample()
	id := 0
	for {
		var taskResponse TaskResponseArgs
		var err error

		taskResponse, err = RequestNewWork(id)
		if taskResponse.JobType == "WAIT" {
			time.Sleep(1000)
			continue
		} else if err != nil {
			log.Fatalf("Encountered error: %v", err)
			break
		} else if taskResponse.JobType == "" {
			time.Sleep(2000)
			continue
		}
		id = taskResponse.Job.WorkerId
		if taskResponse.JobType == "Map" {
			filename := taskResponse.Job.InputFilePath
			file, err := os.Open(filename)
			if err != nil {
				log.Fatalf("cannot open %v", filename)
			}

			content, err := ioutil.ReadAll(file)
			if err != nil {
				log.Fatalf("cannot read %v", filename)
			}
			file.Close()
			kva := mapf(regex, string(content))
			kvaPartitions := make(map[int][]KeyValue) // each file gets a keyvalue[]
			for i := 0; i < taskResponse.Job.NReduce; i++ {
				kvaPartitions[i] = make([]KeyValue, 0)
			}
			for _, kv := range kva {
				fileNum := ihash(kv.Key) % taskResponse.Job.NReduce
				kvaPartitions[fileNum] = append(kvaPartitions[fileNum], kv)
			}
			for i := 0; i < taskResponse.Job.NReduce; i++ {
				filename := fmt.Sprintf("mr-%d-%d", i, taskResponse.Job.JobId)
				file, err := os.CreateTemp("", filename)
				if err != nil {
					log.Fatalf("Error creating temp file mr-%d-%d", i, taskResponse.Job.JobId)
				}
				enc := json.NewEncoder(file)
				for _, kv := range kvaPartitions[i] {
					enc.Encode(&kv)
				}
				err = file.Close()
				if err != nil {
					log.Fatal("error closing file")
				}
				err = os.Rename(file.Name(), filename)
				if err != nil {
					log.Fatal("error performing atomic rewrite")
				}
			}
		} else {
			intermediate := make([]KeyValue, 0)
			format := fmt.Sprintf("mr-out-%d", taskResponse.Job.JobId)
			ofile, err := os.CreateTemp("", format)
			for i := 0; i < taskResponse.Job.MapPartitions; i++ {
				fileFormat := fmt.Sprintf("mr-%d-%d", taskResponse.Job.JobId, i)
				inputFile, err := os.Open(fileFormat)
				if err != nil {
					log.Fatalf("Failed to open file %v", fileFormat)
				}
				dec := json.NewDecoder(inputFile)
				for {
					var kv KeyValue
					if err := dec.Decode(&kv); err != nil {
						break
					}
					intermediate = append(intermediate, kv)
				}

			}
			sort.Sort(ByKey(intermediate))
			i := 0
			for i < len(intermediate) {
				j := i + 1
				for j < len(intermediate) && intermediate[j].Key == intermediate[i].Key {
					j++
				}
				values := []string{}
				for k := i; k < j; k++ {
					values = append(values, intermediate[k].Value)
				}
				output := reducef(intermediate[i].Key, values)

				// this is the correct format for each line of Reduce output.
				fmt.Fprintf(ofile, "%v %v\n", intermediate[i].Key, output)

				i = j
			}

			ofile.Close()
			if err = os.Rename(ofile.Name(), format); err == nil {
			}
			//if err = os.Remove(taskResponse.Job.InputFilePath); err == nil {	}
		}
	}
}

// Request new task to complete
func RequestNewWork(workerId int) (TaskResponseArgs, error) {
	args := WorkerCallArgs{workerId}
	reply := TaskResponseArgs{}
	ok := call("Coordinator.RequestNewWork", &args, &reply)
	if ok || reply.JobType == "WAIT" {
		return reply, nil
	} else {
		return TaskResponseArgs{}, fmt.Errorf("error requesting new work via rpc")
	}
}

// example function to show how to make an RPC call to the coordinator.
//
// the RPC argument and reply types are defined in rpc.go.
func CallExample() {

	// declare an argument structure.
	args := ExampleArgs{}

	// fill in the argument(s).
	args.X = 99

	// declare a reply structure.
	reply := ExampleReply{}

	// send the RPC request, wait for the reply.
	// the "Coordinator.Example" tells the
	// receiving server that we'd like to call
	// the Example() method of struct Coordinator.
	ok := call("Coordinator.Example", &args, &reply)
	if ok {
		// reply.Y should be 100.
		fmt.Printf("reply.Y %v\n", reply.Y)
	} else {
		fmt.Printf("call failed!\n")
	}
}

// send an RPC request to the coordinator, wait for the response.
// usually returns true.
// returns false if something goes wrong.
func call(rpcname string, args interface{}, reply interface{}) bool {
	// c, err := rpc.DialHTTP("tcp", "127.0.0.1"+":1234")
	sockname := coordinatorSock()
	c, err := rpc.DialHTTP("unix", sockname)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	defer c.Close()

	err = c.Call(rpcname, args, reply)
	if err == nil {
		return true
	}

	return false
}
