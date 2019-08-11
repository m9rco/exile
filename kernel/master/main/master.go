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
	myFigure := figure.NewFigure("exile", "", true)
	myFigure.Print()

	// Initialize the container
	if err = common.InitContainer(); err != nil {
		fmt.Printf("init container fail: %v", err)
		goto ERROR
	}

	// Initialize the master logger manager
	if err = master.InitWorkerMgr(); err != nil {
		fmt.Printf("init Job worker fail: %v", err)
		goto ERROR
	}

	// Initialize the master logger manager
	if err = master.InitLogMgr(); err != nil {
		fmt.Printf("init Job Manager fail: %v", err)
		goto ERROR
	}

	// Initialize the master api serve
	if err = master.InitApiServer(); err != nil {
		fmt.Printf("init Api Serve fail: %v", err)
		goto ERROR
	}
ERROR:
}

func main() {
	var wg sync.WaitGroup

	wg.Add(1)
	wg.Wait()
}
