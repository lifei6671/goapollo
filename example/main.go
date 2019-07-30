package main

import (
	"context"
	"github.com/lifei6671/goapollo"
	"log"
	"os"
)

func main() {
	host := os.Getenv("APOLLO_HOST")
	c := goapollo.New(host, "6e77bd897fe903ad", "default")

	c.AddNamespace("application").AddNamespace("wechat")

	c.Run(context.Background())
	val, _ := c.GetValue("zookeeper_timeout")
	log.Printf("zookeeper_timeout:%s", val)
	log.Printf("keys -> %+v", c.AllKeys("application"))
	for {
		select {
		case change := <-c.WatchUpdate():
			log.Printf("配置更新通知%s\n", change)
		}
	}
}
