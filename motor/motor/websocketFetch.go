package motor

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
} // use default options

func responseWebsocketFetch(c *websocket.Conn, mt int, request []byte) {
	req := ParseFetchMessage(request)
	t1 := time.Now()
	
	var response string
	// 如果只查询故障信息就不需要访问Redis
	if req.OnlyFault == "true" {
		response = handleFetchSql(req.SerialNumber, req.Start, req.OnlyFault, req.Parameters)
	} else {
		response = haldleRedis(req.SerialNumber, req.Start, req.Parameters)
		if response == "" {
			response = handleFetchSql(req.SerialNumber, req.Start, req.OnlyFault, req.Parameters)
		}
	}

	err := c.WriteMessage(mt, []byte(response))
	if err != nil {
		// log.Println("write:", err)
		Error.Println("write:", err)
		return
	}
	// log.Println("during: ", time.Since(t1))
	Info.Println("during: ", time.Since(t1))
}

func websocketHandleFetch(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		// log.Println("Websocket upgrade:", err)
		Error.Println("Websocket upgrade:", err)
		return
	}
	defer c.Close()

	ch := make(chan struct{})

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
		mt, req, err := c.ReadMessage()
		if err != nil {
			// log.Println("read:", err)
			Error.Println("read:", err)
			break
		}
		ch <- struct{}{}
		responseWebsocketFetch(c, mt, req)
	}
}

func WebSocketRun(addr string) {
	http.HandleFunc("/", websocketHandleFetch)
	// log.Fatal(http.ListenAndServe(addr, nil))
	Info.Fatal(http.ListenAndServe(addr, nil))
}
