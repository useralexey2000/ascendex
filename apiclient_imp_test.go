package main

import (
	"ascendex/mocks"
	"errors"
	"log"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ApiTestSuite struct {
	suite.Suite
	ctl     *gomock.Controller
	conmock *mocks.MockConnector
	c       *ASClient
}

func (ts *ApiTestSuite) BeforeTest(suitName, testName string) {
	ctl := gomock.NewController(ts.T())
	conmock := mocks.NewMockConnector(ctl)
	ts.ctl = ctl
	ts.conmock = conmock

	c := &ASClient{
		con:  conmock,
		done: make(chan struct{}),
		log:  *log.Default(),
	}
	ts.c = c
}

func (ts *ApiTestSuite) AfterTest(suitName, testName string) {
	ts.ctl.Finish()
}

func (ts *ApiTestSuite) Test_ClientConnection() {
	ts.conmock.EXPECT().Dial(url).Return(nil)

	// success case
	err := ts.c.Connection()
	assert.Nil(ts.T(), err)

	// err case
	ts.conmock.EXPECT().Dial(url).Return(errors.New("err"))
	err = ts.c.Connection()
	assert.NotNil(ts.T(), err)
}

func (ts *ApiTestSuite) Test_ClientDisconnect() {
	ts.conmock.EXPECT().Close()
	ts.c.Disconnect()
}

func (ts *ApiTestSuite) Test_SubscribeToChannel() {
	// success case
	ts.conmock.EXPECT().ReadMessage().Return(nil, nil)
	ts.conmock.EXPECT().WriteMessage(gomock.Any()).Return(nil)

	err := ts.c.SubscribeToChannel("chan")
	assert.Nil(ts.T(), err)

	// err read
	ts.conmock.EXPECT().ReadMessage().Return(nil, errors.New("err"))
	err = ts.c.SubscribeToChannel("chan")
	assert.NotNil(ts.T(), err)

	// write err
	ts.conmock.EXPECT().ReadMessage().Return(nil, nil)
	ts.conmock.EXPECT().WriteMessage(gomock.Any()).Return(errors.New("err"))

	err = ts.c.SubscribeToChannel("chan")
	assert.NotNil(ts.T(), err)
}

func (ts *ApiTestSuite) Test_ReadMessagesFromChannel() {
	msg := `{
		"m": "bbo",
		"symbol": "BTC/USDT",
		"data": {
			"ts": 1573068442532,
			"bid": [
				"9309.11",
				"0.0197172"
			],
			"ask": [
				"9309.12",
				"0.8851266"
			]
		}
	}`
	expected := BestOrderBook{
		Ask: Order{
			Amount: 9309.12,
			Price:  0.8851266,
		},
		Bid: Order{
			Amount: 9309.11,
			Price:  0.0197172,
		},
	}

	// success case
	ts.conmock.EXPECT().ReadMessage().Return([]byte(msg), nil)
	ts.conmock.EXPECT().ReadMessage().Return(nil, errors.New("close err"))

	ch := make(chan BestOrderBook)

	ts.c.ReadMessagesFromChannel(ch)
	actual := <-ch

	assert.Equal(ts.T(), expected, actual)

	// // marshal err
	ts.conmock.EXPECT().ReadMessage().Return([]byte("garbage"), nil)

	ts.c.ReadMessagesFromChannel(ch)
	// wait for read call
	time.Sleep(1 * time.Second)

	// extract err
	msg = `{
		"m": "bbo",
		"symbol": "BTC/USDT",
		"data": {
			"ts": 1573068442532,
			"bid": [
				"9309.11",
				"0.0197172"
			],
			"ask": [
				"garbage",
				"0.8851266"
			]
		}
	}`

	ts.conmock.EXPECT().ReadMessage().Return([]byte(msg), nil)

	ts.c.ReadMessagesFromChannel(ch)
	// wait for read call
	time.Sleep(1 * time.Second)
}

func (ts *ApiTestSuite) Test_WriteMessagesToChannel() {
	ts.conmock.EXPECT().WriteMessage([]byte(ping)).Return(nil)
	ts.conmock.EXPECT().WriteMessage(gomock.Any()).Return(errors.New("close err"))

	pingTimer = 0
	ts.c.WriteMessagesToChannel()

	time.Sleep(1 * time.Second)

	// done return
	close(ts.c.done)
	ts.c.WriteMessagesToChannel()
}

func TestApiTestSuite(t *testing.T) {
	suite.Run(t, new(ApiTestSuite))
}

func Test_extract(t *testing.T) {

}
