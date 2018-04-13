package main

import (
	"github.com/urfave/cli"
	"os"
	"runtime"
	"github.com/go-toon/phpbear/cmd"
	"fmt"
)

const APP_VER = "1.0.1"

func init() {
	runtime.GOMAXPROCS(runtime.NumCPU())
}

func main() {
	app := cli.NewApp()
	app.Name  = "phpbear"
	app.Usage = "php团队项目构建工具"
	app.Version = APP_VER
	app.Commands = []cli.Command{
		cmd.CmdInit,
		cmd.CmdBuild,
	}
	app.Flags = append(app.Flags, []cli.Flag{}...)

	err := app.Run(os.Args)
	if err != nil {
		fmt.Println(err.Error())
	}
}