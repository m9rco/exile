package worker

import (
	"fmt"
	"github.com/m9rco/exile/kernel/common"
	"time"
)

type Scheduler struct {
	jobEventChan      chan *common.JobEvent
	jobPlanTable      map[string]*common.JobSchedulePlan
	jobExecutingTable map[string]*common.JobExecuteInfo
	jobResultChan     chan *common.JobExecuteResult
}

// initialize the InitScheduler
func InitScheduler() (err error) {
	common.Manage.SetSingleton("Scheduler", Scheduler{
		jobEventChan:      make(chan *common.JobEvent, 1000),
		jobPlanTable:      make(map[string]*common.JobSchedulePlan),
		jobExecutingTable: make(map[string]*common.JobExecuteInfo),
		jobResultChan:     make(chan *common.JobExecuteResult, 1000),
	})
	scheduler := common.Manage.GetSingleton("JobManager").(Scheduler)
	go scheduler.scheduleLoop() // start scheduler
	return
}

// scheduler event
func (scheduler *Scheduler) handleJobEvent(jobEvent *common.JobEvent) {
	var (
		jobSchedulePlan *common.JobSchedulePlan
		jobExecuteInfo  *common.JobExecuteInfo
		jobExecuting    bool
		jobExisted      bool
		err             error
	)
	switch jobEvent.EventType {
	case common.JOB_EVENT_SAVE: // save job event
		if jobSchedulePlan, err = common.BuildJobSchedulePlan(jobEvent.Job); err != nil {
			return
		}
		scheduler.jobPlanTable[jobEvent.Job.Name] = jobSchedulePlan
	case common.JOB_EVENT_DELETE: // delete job event
		if jobSchedulePlan, jobExisted = scheduler.jobPlanTable[jobEvent.Job.Name]; jobExisted {
			delete(scheduler.jobPlanTable, jobEvent.Job.Name)
		}
	case common.JOB_EVENT_KILL: // kill job event
		// cancel command exec, to determine whether a task in the execution
		if jobExecuteInfo, jobExecuting = scheduler.jobExecutingTable[jobEvent.Job.Name]; jobExecuting {
			jobExecuteInfo.CancelFunc() // trigger the command kill shell child process
		}
	}
}

// try start the jobs
func (scheduler *Scheduler) TryStartJob(jobPlan *common.JobSchedulePlan) {
	var (
		jobExecuteInfo *common.JobExecuteInfo
		jobExecuting   bool
		executor       Executor
	)
	// ! job may be running for a long time，to prevent concurrent
	if jobExecuteInfo, jobExecuting = scheduler.jobExecutingTable[jobPlan.Job.Name]; jobExecuting {
		return
	}

	// build the execution status information
	jobExecuteInfo = common.BuildJobExecuteInfo(jobPlan)

	// save execution status
	scheduler.jobExecutingTable[jobPlan.Job.Name] = jobExecuteInfo

	// exec
	executor = common.Manage.GetSingleton("Executor").(Executor)
	//fmt.Println("exec job :", jobExecuteInfo.Job.Name, jobExecuteInfo.PlanTime, jobExecuteInfo.RealTime)
	executor.ExecuteJob(jobExecuteInfo)
}

// to recalculate the job scheduling
func (scheduler *Scheduler) TrySchedule() (scheduleAfter time.Duration) {
	var (
		jobPlan  *common.JobSchedulePlan
		now      time.Time
		nearTime *time.Time
	)

	// if the job list is empty，sleep
	if len(scheduler.jobPlanTable) == 0 {
		scheduleAfter = 1 * time.Second
		return
	}
	now = time.Now()

	// iterate through all the job
	for _, jobPlan = range scheduler.jobPlanTable {
		if jobPlan.NextTime.Before(now) || jobPlan.NextTime.Equal(now) {
			scheduler.TryStartJob(jobPlan)
			jobPlan.NextTime = jobPlan.Expr.Next(now)
		}

		// statistics recently a task time to expiration
		if nearTime == nil || jobPlan.NextTime.Before(*nearTime) {
			nearTime = &jobPlan.NextTime
		}
	}
	// next schedule interval（ recent schedule time - now time ）
	scheduleAfter = (*nearTime).Sub(now)
	return
}

func (scheduler *Scheduler) handleJobResult(result *common.JobExecuteResult) {
	var (
		jobLog *common.JobLog
	)
	// delete executing log
	delete(scheduler.jobExecutingTable, result.ExecuteInfo.Job.Name)

	// create exec job logs
	if result.Err != common.ERROR_LOCK_ALREADY_REQUIRED {
		jobLog = &common.JobLog{
			JobName:      result.ExecuteInfo.Job.Name,
			Command:      result.ExecuteInfo.Job.Command,
			Output:       string(result.Output),
			PlanTime:     result.ExecuteInfo.PlanTime.UnixNano() / 1000 / 1000,
			ScheduleTime: result.ExecuteInfo.RealTime.UnixNano() / 1000 / 1000,
			StartTime:    result.StartTime.UnixNano() / 1000 / 1000,
			EndTime:      result.EndTime.UnixNano() / 1000 / 1000,
		}
		if result.Err != nil {
			jobLog.Err = result.Err.Error()
		} else {
			jobLog.Err = ""
		}
		common.Manage.GetSingleton("LogManager").(LogManager).Append(jobLog)
	}
	fmt.Println("job execute is finished:", result.ExecuteInfo.Job.Name, string(result.Output), result.Err)
}

// schedule loop
func (scheduler *Scheduler) scheduleLoop() {
	var (
		jobEvent      *common.JobEvent
		scheduleAfter time.Duration
		scheduleTimer *time.Timer
		jobResult     *common.JobExecuteResult
	)

	// initialize (1s)
	scheduleAfter = scheduler.TrySchedule()

	// scheduling delay timer
	scheduleTimer = time.NewTimer(scheduleAfter)

	// crontab common.Job
	for {
		select {
		case jobEvent = <-scheduler.jobEventChan: // listen job change event
			scheduler.handleJobEvent(jobEvent)
		case <-scheduleTimer.C:
		case jobResult = <-scheduler.jobResultChan:
			scheduler.handleJobResult(jobResult)
		}
		// try scheduler
		scheduleAfter = scheduler.TrySchedule()
		// reset scheduler
		scheduleTimer.Reset(scheduleAfter)
	}
}

// push job change event
func (scheduler *Scheduler) PushJobEvent(jobEvent *common.JobEvent) {
	scheduler.jobEventChan <- jobEvent
}

// callback job result
func (scheduler *Scheduler) PushJobResult(jobResult *common.JobExecuteResult) {
	scheduler.jobResultChan <- jobResult
}
