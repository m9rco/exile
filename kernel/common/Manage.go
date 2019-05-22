package common

import (
	"fmt"
	"github.com/m9rco/exile/kernel/utils"
	"os"
)

var (
	Manage *utils.Container
)

func InitContainer() (err error) {
	Manage = utils.NewContainer()
	// configure the configs
	Manage.SetPrototype("configure", func() (i interface{}, e error) {
		dir, _ := os.Getwd()
		initParser := utils.IniParser{}
		if err != initParser.Load(dir+"/config/app.ini") {
			fmt.Printf("fail to read file: %v", err)
			os.Exit(1)
		}
		return initParser, nil
	})

	return
}
