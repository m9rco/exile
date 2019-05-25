package master

import (
	"fmt"
	"github.com/gorilla/mux"
	"github.com/m9rco/exile/kernel/common"
	. "github.com/m9rco/exile/kernel/utils"
	"net"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Job http service
type ApiServer struct {
	httpServer *http.Server
}

var (
	err error
	job common.Job
)

/*
// Save the jobs

method POST /job  name=job1&command=echo hello&cronExpr=*\/5 * * * * * *
 */
func handleJobSave(writer http.ResponseWriter, request *http.Request) {
	var (
		oldJob       *common.Job
		bytes        []byte
		jobManageSev JobManager
	)

	if err = request.ParseForm(); err != nil {
		goto ERROR
	}
	job = common.Job{
		Name:     request.PostForm.Get("name"),
		Command:  request.PostForm.Get("command"),
		CronExpr: request.PostForm.Get("cronExpr"),
	}
	jobManageSev = common.Manage.GetSingleton("JobManager").(JobManager)
	if oldJob, err = jobManageSev.SaveJob(&job); err != nil {
		goto ERROR
	}
	if bytes, err = common.BuildResponse(0, "success", oldJob); err == nil {
		writer.Write(bytes)
	}
	return

ERROR:
	// return to the front anomalies. errno -1
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		writer.Write(bytes)
	}
	return
}

/*

// Delete the jobs

method DELETE /job/{name}
*/
func handleJobDelete(writer http.ResponseWriter, request *http.Request) {
	var (
		oldJob       *common.Job
		bytes        []byte
		jobManageSev JobManager
		name         string
		vars         map[string]string
	)
	if err = request.ParseForm(); err != nil {
		goto ERROR
	}
	vars = mux.Vars(request)
	name = vars["name"]
	jobManageSev = common.Manage.GetSingleton("JobManager").(JobManager)
	if oldJob, err = jobManageSev.DeleteJob(name); err != nil {
		goto ERROR
	}
	if bytes, err = common.BuildResponse(0, "success", oldJob); err == nil {
		writer.Write(bytes)
	}
	return
ERROR:
	// return to the front anomalies. errno -1
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		writer.Write(bytes)
	}
	return
}

/*

// List the jobs

method GET  /job
*/
func handleJobList(writer http.ResponseWriter, _ *http.Request) {
	var (
		jobList      []*common.Job
		bytes        []byte
		jobManageSev JobManager
	)

	jobManageSev = common.Manage.GetSingleton("JobManager").(JobManager)
	if jobList, err = jobManageSev.ListJobs(); err != nil {
		goto ERROR
	}
	if bytes, err = common.BuildResponse(0, "success", jobList); err == nil {
		writer.Write(bytes)
	}
	return
ERROR:
	// return to the front anomalies. errno -1
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		writer.Write(bytes)
	}
	return
}

/**
// Kill the jobs
PUT /job/{name}

 */
func handleJobKill(writer http.ResponseWriter, request *http.Request) {
	var (
		bytes        []byte
		jobManageSev JobManager
		name         string
		vars         map[string]string
	)
	if err = request.ParseForm(); err != nil {
		goto ERROR
	}
	vars = mux.Vars(request)
	name = vars["name"]
	jobManageSev = common.Manage.GetSingleton("JobManager").(JobManager)
	if err = jobManageSev.KillJob(name); err != nil {
		goto ERROR
	}
	if bytes, err = common.BuildResponse(0, "success", nil); err == nil {
		writer.Write(bytes)
	}
	return
ERROR:
	// return to the front anomalies. errno -1
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		writer.Write(bytes)
	}
	return
}

func handleJobLog(writer http.ResponseWriter, request *http.Request) {
	var (
		err        error
		name       string
		skipParam  string
		limitParam string
		skip       int
		limit      int
		logArr     []*common.JobLog
		bytes      []byte
	)

	if err = request.ParseForm(); err != nil {
		goto ERR
	}

	name = request.Form.Get("name")
	skipParam = request.Form.Get("skip")
	limitParam = request.Form.Get("limit")
	if skip, err = strconv.Atoi(skipParam); err != nil {
		skip = 0
	}
	if limit, err = strconv.Atoi(limitParam); err != nil {
		limit = 20
	}

	if logArr, err = G_logMgr.ListLog(name, skip, limit); err != nil {
		goto ERR
	}

	// 正常应答
	if bytes, err = common.BuildResponse(0, "success", logArr); err == nil {
		writer.Write(bytes)
	}
	return

ERR:
	if bytes, err = common.BuildResponse(-1, err.Error(), nil); err == nil {
		writer.Write(bytes)
	}
}

func handleWorker(_ http.ResponseWriter, _ *http.Request) {

}

// Initialize the service
func InitApiServer() (err error) {
	var (
		httpServer      *http.Server
		listener        net.Listener
		configureSource interface{}
	)
	if configureSource, err = common.Manage.GetPrototype("configure"); err != nil {
		fmt.Printf("fail to read file: %v", err)
		os.Exit(1)
	}
	configure := configureSource.(IniParser)
	router := mux.NewRouter().StrictSlash(true)

	// Configure the routers
	router.HandleFunc(common.API_JOB_CREATE, handleJobSave).Methods("POST")
	router.HandleFunc(common.API_JOB_DELETE, handleJobDelete).Methods("DELETE")
	router.HandleFunc(common.API_JOB_LIST, handleJobList).Methods("GET")
	router.HandleFunc(common.API_JOB_KILL, handleJobKill).Methods("PUT")
	router.HandleFunc(common.API_JOB_LOG, handleJobLog).Methods("GET")
	router.HandleFunc(common.API_WORK_LIST, handleWorker).Methods("GET")

	// Start TCP listener
	if listener, err = net.Listen(
		configure.GetString("server", "protocol"), configure.GetString("server", "port"));
		err != nil {
		return
	}
	// Create http servers
	httpServer = &http.Server{
		Addr:              "",
		Handler:           router,
		TLSConfig:         nil,
		ReadTimeout:       time.Duration(configure.GetInt32("server", "read_timeout")) * time.Millisecond,
		ReadHeaderTimeout: 0,
		WriteTimeout:      time.Duration(configure.GetInt32("server", "write_timeout")) * time.Millisecond,
		IdleTimeout:       0,
		MaxHeaderBytes:    0,
		TLSNextProto:      nil,
		ConnState:         nil,
		ErrorLog:          nil,
	}
	common.Manage.SetSingleton("ApiServer", &ApiServer{
		httpServer: httpServer,
	})
	go httpServer.Serve(listener)
	return
}
