# Motor_monitor_go

基于Go编写的点击远程监测服务器程序

在motor/maintest.go中给出了运行方式，需要修改其中的mysql和redis信息，在运行时可以指定端口，如

```
$ go run maintest.go -port 8090
```

其中的UploadRun，SocketFetchRun和WebSocketRun函数分别对应数据上传（socket方式），读取数据并基于socket传输，读取数据并基于websocket传输的实现
