package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
)

var (
	url       = "wss://ascendex.com/0/api/pro/v1/stream"
	sub       = `{ "op": "sub", "id": "abc123", "ch":"bbo:%s" }`
	ping      = `{ "op": "ping" }`
	pingTimer = 15
)

type Connector interface {
	Dial(string) error
	ReadMessage() ([]byte, error)
	WriteMessage([]byte) error
	Close() error
}

type defaultConnector struct {
	con *websocket.Conn
}

func (d *defaultConnector) Dial(url string) error {
	con, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return err
	}
	d.con = con
	return nil
}

func (d *defaultConnector) ReadMessage() ([]byte, error) {
	_, message, err := d.con.ReadMessage()
	return message, err
}

func (d *defaultConnector) WriteMessage(b []byte) error {
	return d.con.WriteMessage(1, b)
}

func (d *defaultConnector) Close() error {
	return d.con.Close()
}

type ASClient struct {
	con  Connector
	done chan struct{}
	log  log.Logger
}

func NewApiClient() *ASClient {
	return &ASClient{
		done: make(chan struct{}),
		con:  &defaultConnector{},
		log:  *log.Default(),
	}
}

func (c *ASClient) Connection() error {
	err := c.con.Dial(url)
	if err != nil {
		return err
	}

	return nil
}

func (c *ASClient) Disconnect() {
	fmt.Println("closing")
	close(c.done)
	c.con.Close()
}

func (c *ASClient) SubscribeToChannel(symbol string) error {
	// Ensure conn
	_, err := c.con.ReadMessage()
	if err != nil {
		c.log.Println("sub.read:", err)
		return err
	}

	// subscribe to chan
	sub := fmt.Sprintf(sub, symbol)
	err = c.con.WriteMessage([]byte(sub))
	if err != nil {
		c.log.Println("sub.write:", err)
		return err
	}

	return nil
}

type message struct {
	M      string `json:"m"`
	Symbol string `json:"symbol"`
	Data   data   `json:"data"`
}

type data struct {
	Ts  int       `json:"ts"`
	Bid [2]string `json:"bid"`
	Ask [2]string `json:"ask"`
}

func (c *ASClient) ReadMessagesFromChannel(ch chan<- BestOrderBook) {
	go func() {
		for {
			msg, err := c.con.ReadMessage()
			if err != nil {
				c.log.Println("ch.read:", err)
				return
			}

			var m message
			err = json.Unmarshal(msg, &m)
			if err != nil {
				c.log.Println("ch.unmarshal:", err)
				return
			}

			switch m.M {
			// the case can be omited
			case "ping":
				c.log.Println("skip ping message")
				continue
			case "bbo":
				bbo, err := extract(&m)
				if err != nil {
					c.log.Println("ch.extract:", err)
					return
				}
				// dont wait until reader reads from chan
				go func() {
					bbo := bbo
					select {
					case <-c.done:
					case ch <- bbo:
					}
				}()
			default:
			}
		}
	}()
}

func (c *ASClient) WriteMessagesToChannel() {
	go func() {
		for {
			select {
			case <-c.done:
				return
			case <-time.After(time.Duration(pingTimer) * time.Second):
				err := c.con.WriteMessage([]byte(ping))
				if err != nil {
					c.log.Println("ch.write:", err)
					return
				}
			}
		}
	}()
}

func extract(m *message) (BestOrderBook, error) {
	am, err := strconv.ParseFloat(m.Data.Ask[0], 64)
	if err != nil {
		return BestOrderBook{}, err
	}

	pr, err := strconv.ParseFloat(m.Data.Ask[1], 64)
	if err != nil {
		return BestOrderBook{}, err
	}

	ask := Order{
		Amount: am,
		Price:  pr,
	}

	am, err = strconv.ParseFloat(m.Data.Bid[0], 64)
	if err != nil {
		return BestOrderBook{}, err
	}

	pr, err = strconv.ParseFloat(m.Data.Bid[1], 64)
	if err != nil {
		return BestOrderBook{}, err
	}

	bid := Order{
		Amount: am,
		Price:  pr,
	}

	return BestOrderBook{
		Ask: ask,
		Bid: bid,
	}, nil
}
