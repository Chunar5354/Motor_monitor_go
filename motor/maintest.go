package main

import (
	"flag"
	"motor"
)

// set the address to get port from commadline
type Addr struct{ a string }

func (a *Addr) Set(s string) error {
	a.a = "your_ip:" + s
	return nil
}

func (a *Addr) String() string {
	return "your_ip:" + a.a
}

func main() {
	motor.MysqlInit("mysql_user_name", "mysql_password", "mysql_port")
	motor.RedisInit("redis_password", "redis_port", 0) // 0 stands for the database 0
	// set port from command line
	address := Addr{"default_ip:default_port"}
	flag.CommandLine.Var(&address, "port", "set the port")
	flag.Parse()
	// motor.SocketFetchRun(address.a)
	// motor.UploadRun(address.a)
	motor.WebSocketRun(address.a)
}
