package main

import (
	"context"
	"github.com/lifei6671/goapollo"
	"log"
	"os"
	"time"
)

func main() {
	host := os.Getenv("APOLLO_HOST")
	c := goapollo.New(host, "6e77bd897fe903ad", "default")

	c.AddNamespace("application").
		AddNamespaceWithSerializer("wechat", goapollo.NewGobSerializer()).
		AddNamespaceWithSerializerWithPath("mongo", goapollo.NewJsonSerializer(), "/Users/minho/wx_Lifeilin/github.com/lifei6671/goapollo/testdata/mongo.json")

	if err := c.Run(context.Background()); err != nil {
		log.Fatalf("启动 Apollo 失败 -> %s", err)
	}
	val, _ := c.GetValue("zookeeper_timeout")
	log.Printf("zookeeper_timeout:%s", val)
	log.Printf("keys -> %+v", c.AllKeys("application"))
	timer := time.NewTicker(time.Second * 5)

	for {
		select {
		case change := <-c.WatchUpdate():
			log.Printf("配置更新通知%s\n", change)
		case <-timer.C:
			val, _ := c.GetValueWithNamespace("wechat", "TIMEOUT")
			log.Printf("TIMEOUT=%s", val)
		}
	}
}
