package motor

import (
	"encoding/json"
	"fmt"
	"net"
	"strings"
	"time"
)

// parse the request message
func ParseUploadMessage(text []byte) map[string]string {
	req := make(map[string]string)
	if err := json.Unmarshal(text, &req); err != nil {
		// log.Fatalf("JSON unmarshaling failed: %s", err)
		Error.Fatalf("JSON unmarshaling failed: %s", err)
	}
	return req
}

// connect to mysql&redis and insert data
func dbUpload(req map[string]string) error {
	serial_number := req["serial_number"]
	create_time := req["create_time"]
	delete(req, "serial_number") // delete non-parament keys
	delete(req, "create_time")
	create_time = strings.Replace(create_time, "_", ":", -1)
	err := mysqlUpload(req, serial_number, create_time)
	if err != nil {
		return err
	}
	err = redisUpload(req, serial_number, create_time)
	if err != nil {
		return err
	}
	return nil
}

// write response to client
func responseUpload(c net.Conn, text []byte) {
	t1 := time.Now()
	req := ParseUploadMessage(text)
	// fmt.Println(strings.TrimRight(req["temp_water"], "\n")) // 去掉数据后面的换行符
	fmt.Fprintln(c, "Successfully uploaded")
	err := dbUpload(req)
	fmt.Println("during: ", time.Since(t1))
	if err != nil {
		// log.Println("Something wrong with database:", err)
		Error.Println("Something wrong with database:", err)
	}
}

func handleUpload(c net.Conn) {
	defer c.Close()
	ch := make(chan struct{})
	var buf = make([]byte, 256)
	var req []byte

	// timing
	go func() {
		for {
			select {
			case <-time.After(60 * time.Second): // if no message in one minute, close the connect
				// log.Println("connect finished")
				Error.Println("connect finished")
				c.Close()
				return
			case <-ch:
			}
		}
	}()

	// log.Println("start to read from conn")
	Info.Println("start to read from conn")
	for {
		req = nil
		for {
			// read from the connection
			n, err := c.Read(buf)
			if err != nil {
				// log.Println("conn read error:", err)
				Error.Println("conn read error:", err)
				return
			}
			req = append(req, buf[:n]...)
			if buf[n-1] == '}' { // check if the reveiving data is end
				break
			}
		}
		ch <- struct{}{}
		responseUpload(c, req)
	}
}

// run upload server
func UploadRun(addr string) {
	// create tcp connection
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		// log.Fatal(err)
		Error.Fatal(err)
	}

	// handle socket connection
	for {
		conn, err := listener.Accept()
		if err != nil {
			// log.Print(err) // e.g., connection aborted
			Error.Println(err) // e.g., connection aborted
			continue
		}
		go handleUpload(conn) // handle one connection at a time
	}
}
