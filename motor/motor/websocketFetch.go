package motor

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{} // use default options

func responseWebsocketFetch(c *websocket.Conn, mt int, req []byte) {
	request := ParseFetchMessage(req)
	t1 := time.Now()

	response := haldleRedis(request.SerialNumber, request.Start, request.Parameters)
	if response == "" {
		response = handleFetchSql(request.SerialNumber, request.Start, request.Parameters)
	}

	err := c.WriteMessage(mt, []byte(response))
	if err != nil {
		log.Println("write:", err)
		return
	}
	log.Println("during: ", time.Since(t1))
}

func websocketHandleFetch(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Websocket upgrade:", err)
		return
	}
	defer c.Close()

	ch := make(chan struct{})

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
		mt, req, err := c.ReadMessage()
		if err != nil {
			log.Println("read:", err)
			break
		}
		ch <- struct{}{}
		responseWebsocketFetch(c, mt, req)
	}
}

func WebSocketRun(addr string) {
	http.HandleFunc("/", websocketHandleFetch)
	log.Fatal(http.ListenAndServe(addr, nil))
}
