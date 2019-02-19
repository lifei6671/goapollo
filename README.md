# 携程 Apollo 配置系统客户端SDK

## 安装

```bash
go get github.com/lifei6671/goapollo
```

## 使用

```go
	ctx,cancel := context.WithCancel(context.Background())
	client := goapollo.NewApolloClient(ctx,
		"http://dev-config.xin.com/",goapollo.ApolloConfig{AppId: "6e77bd897fe903ac", Cluster: "default"},
		"dev_business.nginx")

	client.AddNamespace("op.php")
	client.AddNamespace("proxy.yml")
	client.OnChangeEvent(func(event *goapollo.ChangeEvent) {
		log.Printf("变更事件已执行 -> %+v", event)
	})

	client.Watch()

	time.AfterFunc(time.Minute * 2, func() {
		log.Printf("删除命名空间")
		client.RemoveNamespace("op.php")
	})
```

## 功能

- 实时同步配置
- 配置文件容灾
- 零依赖
- 支持多namespace