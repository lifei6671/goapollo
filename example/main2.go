package main

import (
	"context"
	"github.com/lifei6671/goapollo"
	"log"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := goapollo.Run(ctx); err != nil {
		log.Fatalf("启动客户端失败->%s", err)
	}
	for {
		select {
		case change := <-goapollo.WatchUpdate():
			log.Printf("配置更新通知%s\n", change)
		}
	}
}
