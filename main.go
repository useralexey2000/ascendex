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
}
