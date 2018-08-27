package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"

	"crawlers/handlers"
	"crawlers/service"

	"github.com/BurntSushi/toml"
)

func main() {

	action := flag.String("action", "", "请输入启动项。")

	configType := flag.String("config", "dev", "配置文件类型，dev,test,prod")

	flag.Parse()
	//获取配置文件
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	absPath := strings.Replace(dir, "\\", "/", -1)
	crawersService := new(service.CrawlersContext)
	defer crawersService.Stop()

	if *configType == "prod" {
		_, err = toml.DecodeFile(filepath.Join(absPath, "prod-config.toml"), &crawersService.Config)
	} else {
		_, err = toml.DecodeFile(filepath.Join(absPath, "dev-config.toml"), &crawersService.Config)
	}

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	if strings.Trim(*action, " ") == "" {
		fmt.Println("启动项不能为空")
		return
	}

	crawersService.Config.Action = *action
	//启动配置服务
	crawersService.Start()

	//退出信号
	signChan := make(chan os.Signal, 1)
	signal.Notify(signChan, os.Interrupt)

	go func() {
		for range signChan {
			crawersService.Logrus.Info("\n程序准备退出")
			crawersService.Stop()
			os.Exit(1)
		}
	}()

	//根据启动项启动相关挖掘器
	switch *action {

	case "wechat":
		handlers.WeiChatHandler(crawersService)
	case "yuqing":
		handlers.YuqingHandler(crawersService)
	default:

	}

}
