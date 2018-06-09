package mapreduce

import (
	"fmt"
	"sync"
)

//
// schedule() starts and waits for all tasks in the given phase (mapPhase
// or reducePhase). the mapFiles argument holds the names of the files that
// are the inputs to the map phase, one per map task. nReduce is the
// number of reduce tasks. the registerChan argument yields a stream
// of registered workers; each item is the worker's RPC address,
// suitable for passing to call(). registerChan will yield all
// existing registered workers (if any) and new ones as they register.
//
func schedule(jobName string, mapFiles []string, nReduce int, phase jobPhase,
	registerChan chan string) {
	var ntasks int
	var n_other int // number of inputs (for reduce) or outputs (for map)
	switch phase {
	case mapPhase:
		ntasks = len(mapFiles)
		n_other = nReduce
	case reducePhase:
		ntasks = nReduce
		n_other = len(mapFiles)
	}

	fmt.Printf("Schedule: %v %v tasks (%d I/Os)\n", ntasks, phase, n_other)

	// All ntasks tasks have to be scheduled on workers. Once all tasks
	// have completed successfully, schedule() should return.
	//
	// Your code here (Part III, Part IV).
	//

	var nrJobs sync.WaitGroup
	nrJobs.Add(ntasks)

	wq := make(chan int)
	go func() {
		for wkRpcAddr := range registerChan {
			go func(wkRpcAddr string) {
				for tskId := range wq {
					isSuccess := call(
						wkRpcAddr,
						"Worker.DoTask",
						DoTaskArgs{jobName, mapFiles[tskId], phase, tskId, n_other},
						nil,
					)
					if (isSuccess) {
						nrJobs.Done()
					} else {
						// Remove this worker and then redo the task
						wq <- tskId
						fmt.Printf("Schedule: worker (%s) died and redo task %d\n", wkRpcAddr, tskId)
						break
					}
				}
			}(wkRpcAddr)
		}
	}()

	for tskId := 0; tskId < ntasks; tskId++ {
		wq <- tskId
	}

	nrJobs.Wait()

	fmt.Printf("Schedule: %v done\n", phase)
}
