package common

import (
	"context"
	"encoding/json"
	"github.com/gorhill/cronexpr"
	"time"
)

type Job struct {
	Name     string `json:"name"`
	Command  string `json:"command"`
	CronExpr string `json:"cronExpr"`
}

type JobSchedulePlan struct {
	Job      *Job                 // the job information
	Expr     *cronexpr.Expression // the cron expr
	NextTime time.Time            // the next scheduled time
}

type JobExecuteInfo struct {
	Job        *Job               // the job information
	PlanTime   time.Time          // the plan scheduled time
	RealTime   time.Time          // the real scheduled time
	CancelCtx  context.Context    // the job command is context
	CancelFunc context.CancelFunc // the cancel command func
}

type Response struct {
	Errno int         `json:"errno"`
	Msg   string      `json:"msg"`
	Data  interface{} `json:"data"`
}

type JobEvent struct {
	EventType int
	Job       *Job
}

type JobExecuteResult struct {
	ExecuteInfo *JobExecuteInfo // the exec status
	Output      []byte          // the script output
	Err         error           // the script error info
	StartTime   time.Time       // the script start time
	EndTime     time.Time       // the script end time
}

type JobLog struct {
	JobName      string `json:"jobName" bson:"jobName"`
	Command      string `json:"command" bson:"command"`
	Err          string `json:"err" bson:"err"`
	Output       string `json:"output" bson:"output"`
	PlanTime     int64  `json:"planTime" bson:"planTime"`
	ScheduleTime int64  `json:"scheduleTime" bson:"scheduleTime"`
	StartTime    int64  `json:"startTime" bson:"startTime"`
	EndTime      int64  `json:"endTime" bson:"endTime"`
}

type LogBatch struct {
	Logs []interface{} // 多条日志
}

type JobLogFilter struct {
	JobName string `bson:"jobName"`
}

type SortLogByStartTime struct {
	SortOrder int `bson:"startTime"`
}

func BuildResponse(errno int, msg string, data interface{}) (resp []byte, err error) {
	var (
		response Response
	)
	response.Errno = errno
	response.Msg = msg
	response.Data = data
	return json.Marshal(response)
}
