package common

const (
	API_JOB_LIST   string = "/job"        // GET
	API_JOB_CREATE        = "/job"        // POST
	API_JOB_DELETE        = "/job/{name}" // DELETE
	API_JOB_FETCH         = "/job/{name}" // GET
	API_JOB_KILL          = "/job/{name}" // PUT
	API_JOB_LOG           = "/job/log"    // GET
	API_WORK_LIST         = "/worker"     // GET
)

const (
	JOB_SAVE_DIR   string = "/cron/jobs/"
	JOB_KILLER_DIR        = "/cron/killer/"
	JOB_LOCK_DIR          = "/cron/lock/"
	JOB_WORKER_DIR        = "/cron/workers/"
)

const (
	JOB_EVENT_DELETE = iota // 删除任务事件
	JOB_EVENT_SAVE          // 保存任务事件
	JOB_EVENT_KILL          // 强杀任务事件
)
