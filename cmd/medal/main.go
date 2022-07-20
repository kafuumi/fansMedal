package main

import (
	"context"
	"github.com/Hami-Lemon/fans"
	"log"
	"os"
	"os/signal"
	"path/filepath"
)

const version = "0.1.2"

func main() {
	log.Printf("version: %s\n", version)
	configPath := filepath.Dir(os.Args[0]) + "/config.yaml"
	log.Printf("使用配置文件：%s\n", configPath)
	file, err := os.Open(configPath)
	if err != nil {
		log.Fatalln(err)
	}
	cfg, err := fans.ReadConfig(file)
	_ = file.Close()
	if err != nil {
		log.Fatalln()
	}
	worker := fans.CreateWorker(cfg)
	if worker == nil {
		log.Println("出现错误")
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	exit := make(chan struct{}, 2)
	worker.Start(ctx, exit)
	go func() {
		e := make(chan os.Signal)
		signal.Notify(e, os.Interrupt, os.Kill)
		<-e
		exit <- struct{}{}
		log.Printf("主动退出")
	}()
	<-exit
	cancel()
	log.Println("任务结束")
}
