package common

const (
	API_JOB_LIST   string = "/job"        // get job list
	API_JOB_CREATE        = "/job"        // create the jobs
	API_JOB_DELETE        = "/job/{name}" // delete the jobs
	API_JOB_FETCH         = "/job/{name}" // kill the jobs
	API_JOB_KILL          = "/job/{name}" // get job log
	API_JOB_LOG           = "/job/log"    // get 
	API_WORK_LIST         = "/worker"     // get job work
)

const (
	JOB_SAVE_DIR   string = "/cron/jobs/"
	JOB_KILLER_DIR        = "/cron/killer/"
	JOB_LOCK_DIR          = "/cron/lock/"
	JOB_WORKER_DIR        = "/cron/workers/"
)

const (
	JOB_EVENT_DELETE = iota // delete the job event
	JOB_EVENT_SAVE          // save the job event
	JOB_EVENT_KILL          // killer the job event
)
