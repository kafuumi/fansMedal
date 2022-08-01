package main

import (
	"context"
	"fmt"
	"github.com/Hami-Lemon/fans"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
)

const version = "0.1.4"

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
	//列出所有的粉丝牌
	if len(os.Args) == 2 && os.Args[1] == "list" {
		listMedals(cfg)
		return
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

func listMedals(cfg fans.Config) {
	bili := fans.NewBili(cfg.AccessKey, 0)
	medals, err := bili.GetMedals()
	if err != nil {
		log.Fatalln(err)
	}
	sort.Slice(medals, func(i, j int) bool {
		return medals[i].Level() > medals[j].Level()
	})
	set := make(fans.Set[int64])
	set.Add(cfg.List...)
	var typeStr string
	if cfg.Type {
		typeStr = "[处理]"
	} else {
		typeStr = "[不处理]"
	}
	fmt.Printf("共有粉丝牌：%d个\n", len(medals))
	fmt.Println("等级 | 粉丝牌 | 主播")
	for _, medal := range medals {
		str := fmt.Sprintf("[%d] %s %s(%d) ",
			medal.Level(), medal.Name(), medal.AnchorName(), medal.TargetId())
		if set.Contains(medal.TargetId()) {
			str += typeStr
		}
		fmt.Println(str)
	}
}
