# 携程 apollo 项目 golang 客户端

方便Golang快捷接入携程配置中心框架 [Apollo](https://github.com/ctripcorp/apollo) 所开发的Golang版本客户端。


## 功能

- 实时同步配置
- 灰度配置
- 客户端容灾
- 多命名空间支持
- 无多余依赖
- 支持自定义日志打印


 ## 安装
 
 **Golang Module**
 在项目中引入，执行下面命令即可：
 ```bash
go mod tidy
```

**其他包管理**

```bash
go get github.com/lifei6671/goapollo
```

## 使用

### 自定义客户端
```go
package main

import (
	"context"
	"github.com/lifei6671/goapollo"
	"log"
	"os"
)

func main() {
	//1. 初始化一个客户端
    c := goapollo.New("host", "6e77bd897fe903ad", "default")
    //2. 添加多个命名空间
	c.AddNamespace("application").AddNamespace("wechat")
    //3. 启动并运行客户端
	c.Run(context.Background())
	val,_ := c.GetValue("zookeeper_timeout")
	log.Printf("zookeeper_timeout:%s",val)
	log.Printf("keys -> %+v",c.AllKeys("application"))
	//4.监控变更通知
	for {
		select {
		case change := <-c.WatchUpdate():
			log.Printf("配置更新通知%s\n", change)
		}
	}
}
```

### 使用环境变量初始化默认客户端

```go
package main

import (
	"context"
	"github.com/lifei6671/goapollo"
	"log"
)

func main() {
	ctx,cancel := context.WithCancel(context.Background())
	defer cancel()
	if err := goapollo.Run(ctx);err != nil {
		log.Fatalf("启动客户端失败->%s",err)
	}
	for {
		select {
		case change := <-goapollo.WatchUpdate():
			log.Printf("配置更新通知%s\n", change)
		}
	}
}
```

其中，需要设置如下环境变量：

- `APOLLO_HOST` Apollo 服务器地址
- `APOLLO_APP_ID` 需要监听的APPID
- `APOLLO_CLUSTER` 需要今天的集群
- `APOLLO_NAMESPACE` 需要监听的命名空间，多个可用`;`分隔

## 自定义序列化器

系统支持自定义序列化器，方便接入时根据实际需求来序列化和反序列化配置信息。

目前系统内置了两种序列化器：json 和 gob。

默认情况下，使用 json 做序列化器，也可以在添加命名空间时指定自己实现的序列化器。

