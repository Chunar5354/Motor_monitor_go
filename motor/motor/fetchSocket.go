package motor

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

// the format of request message
type Request struct { // 被解析的字段必须要大写
	SerialNumber string   `json:"serial_number"`
	Parameters   []string `json:"parameter"`
	Limit        string
	Start        string
	End          string
}

// parse the request message
func ParseFetchMessage(text []byte) Request {
	var req Request
	if err := json.Unmarshal(text, &req); err != nil {
		log.Fatalf("JSON unmarshaling failed: %s", err)
	}
	return req
}

// write response to client
func responseFetch(c net.Conn, request []byte) {
	// fmt.Fprintf(c, "%4d%v", length, string(request))
	t1 := time.Now()
	req := ParseFetchMessage(request) // get information from request
	// firstly fetch from redis, if data not in redis, then fetcj from mysql
	response := haldleRedis(req.SerialNumber, req.Start, req.Parameters)
	if response == "" {
		response = handleFetchSql(req.SerialNumber, req.Start, req.Parameters)
	}

	fmt.Fprintf(c, "%v", response)
	log.Println("during: ", time.Since(t1))
}

func socketHandleFetch(c net.Conn) {
	defer c.Close()
	ch := make(chan struct{})
	var buf = make([]byte, 256)
	var req []byte

	// timing
	go func() {
		for {
			select {
			case <-time.After(60 * time.Second): // if no message in one minute, close the connect
				log.Println("connect finished")
				c.Close()
				return
			case <-ch:
			}
		}
	}()

	log.Println("start to read from conn")
	for {
		req = nil
		for {
			// read from the connection
			n, err := c.Read(buf)
			if err != nil {
				log.Println("conn read error:", err)
				return
			}
			req = append(req, buf[:n]...)
			if buf[n-1] == '}' { // check if the reveiving data is end
				break
			}
		}
		ch <- struct{}{}
		responseFetch(c, req)
	}
}

func SocketFetchRun(addr string) {
	// create connection
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		log.Fatal(err)
	}

	// handle socket connection
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err) // e.g., connection aborted
			continue
		}
		go socketHandleFetch(conn) // handle one connection at a time
	}
}
