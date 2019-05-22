package main

import (
	"fmt"
	"github.com/m9rco/exile/kernel/common"
	"github.com/m9rco/exile/kernel/master"
	"runtime"
	"time"
)

func main() {
	// Initialize the environment
	runtime.GOMAXPROCS(runtime.NumCPU())
	var (
		err error
	)
	if err = common.InitContainer(); err != nil {
		fmt.Printf("init container fail: %v", err)
		goto ERROR
	}

	if err = master.InitJobMgr(); err != nil {
		fmt.Printf("init Job Manager fail: %v", err)
		goto ERROR
	}

	if err = master.InitApiServer(); err != nil {
		fmt.Printf("init Api Serve fail: %v", err)
		goto ERROR
	}
	for {
		time.Sleep(1 * time.Second)
	}
ERROR:
}
