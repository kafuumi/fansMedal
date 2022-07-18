package main

import (
	"context"
	"github.com/Hami-Lemon/fans"
	"log"
	"os"
	"os/signal"
)

const version = "0.1.0"

func main() {
	log.Printf("version: %s\n", version)
	file, err := os.Open("config.yaml")
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
	exit := make(chan struct{}, 1)
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
