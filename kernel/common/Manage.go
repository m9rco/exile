package common

import (
	"fmt"
	"git.kids.qihoo.net/go/riven/kernel/utils"
	"os"
)

var (
	Manage *utils.Container
)

func InitContainer() (err error) {
	// configure the configs
	Manage = utils.NewContainer()
	Manage.SetPrototype("configure", func() (i interface{}, e error) {
		dir, _ := os.Getwd()
		initParser := utils.IniParser{}
		if err != initParser.Load(dir+"/../../../config/app.ini") {
			fmt.Printf("configure fail to read file: %v", err)
			os.Exit(1)
		}
		return initParser, nil
	})

	return
}
