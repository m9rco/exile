package main

import (
	"fmt"
	"github.com/common-nighthawk/go-figure"
	"github.com/m9rco/exile/kernel/common"
	"github.com/m9rco/exile/kernel/master"
	"runtime"
	"sync"
)

func init() {
	// Initialize the environment
	runtime.GOMAXPROCS(runtime.NumCPU())
	var (
		err error
	)
	figure.NewFigure("Exile Server", "", true).Print()
	println(" ")

	// Initialize the container
	if err = common.InitContainer(); err != nil {
		fmt.Printf("init container fail: %v", err)
		goto ERROR
	}
	println("Initialize container success ...")

	// Initialize the master logger manager
	if err = master.InitWorkerMgr(); err != nil {
		fmt.Printf("init Job worker fail: %v", err)
		goto ERROR
	}
	println("Initialize job worker success ...")

	// Initialize the master logger manager
	if err = master.InitLogMgr(); err != nil {
		fmt.Printf("init log Manager fail: %v", err)
		goto ERROR
	}
	println("Initialize log manager success ...")

	// Initialize the job  manager
	if err = master.InitJobMgr(); err != nil {
		fmt.Printf("init job Manager fail: %v", err)
		goto ERROR
	}
	println("Initialize job manager success ...")

	// Initialize the master api serve
	if err = master.InitApiServer(); err != nil {
		fmt.Printf("init api server fail: %v", err)
		goto ERROR
	}
	println("Initialize api server success ...")

ERROR:
	if err != nil {
		println(err.Error())
	}
}

func main() {
	var wg sync.WaitGroup

	wg.Add(1)
	wg.Wait()
}
