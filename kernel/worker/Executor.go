package worker

import (
	"github.com/m9rco/exile/kernel/common"
	"math/rand"
	"os/exec"
	"time"
)

type Executor struct {
}

// initialize the job executor
func InitExecutor() (err error) {
	common.Manage.SetSingleton("Executor", Executor{})
	return
}

// execute the once job
func (executor *Executor) ExecuteJob(info *common.JobExecuteInfo) {
	go func() {
		var (
			cmd     *exec.Cmd
			err     error
			output  []byte
			result  *common.JobExecuteResult
			jobLock *JobLock
			jobSev  JobManager
		)
		result = &common.JobExecuteResult{
			ExecuteInfo: info,
			Output:      make([]byte, 0),
		}
		jobSev = common.Manage.GetSingleton("JobManager").(JobManager)

		// initialize lock(txn)
		jobLock = jobSev.CreateJobLock(info.Job.Name)

		// initialize start time.
		result.StartTime = time.Now()

		// locked and rand sleep (0~1s)
		time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

		err = jobLock.TryLock()
		defer jobLock.Unlock()

		if err != nil { // locked fail
			result.Err = err
			result.EndTime = time.Now()
		} else {
			result.StartTime = time.Now()
			// exec shell
			cmd = exec.CommandContext(info.CancelCtx, "/bin/bash", "-c", info.Job.Command)
			output, err = cmd.CombinedOutput()
			// record the execution time
			result.EndTime = time.Now()
			result.Output = output
			result.Err = err
		}
		// result push job
		scheduler := common.Manage.GetSingleton("Scheduler").(Scheduler)
		scheduler.PushJobResult(result)
	}()
}
