package main

import (
	"fmt"
	"os"
	"os/signal"
)

var symbol = `ASD/USDT`

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	ch := make(chan BestOrderBook)
	// defer close(ch)

	c := NewApiClient()
	err := c.Connection()
	if err != nil {
		panic(err)
	}

	go func() {
		<-interrupt
		c.Disconnect()
		close(ch)
		fmt.Println("gracefuly exited")
	}()

	// defer c.Disconnect()
	err = c.SubscribeToChannel(symbol)
	if err != nil {
		panic(err)
	}

	c.ReadMessagesFromChannel(ch)
	c.WriteMessagesToChannel()

	for bbo := range ch {
		fmt.Println("Bbo: ", bbo)
	}

	// url := "wss://ascendex.com/0/api/pro/v1/stream"

	// // u := url.URL{Scheme: "ws", Host: addr, Path: "/echo"}
	// c, _, err := websocket.DefaultDialer.Dial(url, nil)
	// if err != nil {
	// 	log.Fatal("dial:", err)
	// }
	// defer c.Close()

	// // insure connection
	// mt, message, err := c.ReadMessage()
	// if err != nil {
	// 	log.Println("read:", err)
	// 	return
	// }
	// log.Printf("recv: %s, mt: %v", message, mt)

	// data := `{ "op": "sub", "id": "abc123", "ch":"bbo:BTC/USDT" }`

	// // subscribe to chan
	// err = c.WriteMessage(1, []byte(data))
	// if err != nil {
	// 	log.Println("write:", err)
	// 	return
	// }

	// done := make(chan struct{})
	// go func() {
	// 	defer close(done)
	// 	for {
	// 		mt, message, err := c.ReadMessage()
	// 		if err != nil {
	// 			log.Println("read:", err)
	// 			return
	// 		}
	// 		log.Printf("recv: %s, mt: %v", message, mt)
	// 	}
	// }()

	// time.Sleep(50 * time.Second)

}
