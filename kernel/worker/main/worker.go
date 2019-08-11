package main

import (
	"github.com/common-nighthawk/go-figure"
	"github.com/m9rco/exile/kernel/common"
	"github.com/m9rco/exile/kernel/worker"
	"runtime"
	"sync"
)

func init() {
	// Initialize the environment
	runtime.GOMAXPROCS(runtime.NumCPU())
	var (
		err error
	)
	figure.NewFigure("Exile Client", "", true).Print()

	// Initialize the container
	if err = common.InitContainer(); err != nil {
		goto ERROR
	}
	println("Initialize container success ...")

	// Initialize the serve register
	if err = worker.InitRegister(); err != nil {
		println("Initialize register error ...")
		goto ERROR
	}

	println("Initialize container success ...")
	// Initialize the serve register
	if err = worker.InitLogSink(); err != nil {
		println("Initialize logSink error ...")
		goto ERROR
	}
	println("Initialize logSink success ...")

	// Initialize the executor
	if err = worker.InitExecutor(); err != nil {
		println("Initialize executor error ...")
		goto ERROR
	}
	println("Initialize executor success ...")

	// Initialize the worker job scheduler
	if err = worker.InitScheduler(); err != nil {
		println("Initialize scheduler error ...")
		goto ERROR
	}
	println("Initialize scheduler success ...")

	// Initialize the worker job manager
	if err = worker.InitJobMgr(); err != nil {
		println("Initialize job manager error ...")
		goto ERROR
	}
	println("Initialize  job manager success ...")
ERROR:
	println(err.Error())
}

func main() {
	var wg sync.WaitGroup

	wg.Add(1)
	wg.Wait()
}
