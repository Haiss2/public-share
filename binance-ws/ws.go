package main

import (
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

type WsHandler func(message []byte)

type ErrHandler func(err error)

type BookTicker struct {
	Event           string `json:"e,omitempty"`
	UpdateID        int64  `json:"u"`
	Time            int64  `json:"E,omitempty"`
	TransactionTime int64  `json:"T,omitempty"`
	Symbol          string `json:"s"`
	BestBidPrice    string `json:"b"`
	BestBidQty      string `json:"B"`
	BestAskPrice    string `json:"a"`
	BestAskQty      string `json:"A"`
	ActualTimeUs    int64  `json:"us"`
}

var WsServe = func(params []string, h WsHandler, e ErrHandler) (doneC, stopC chan struct{}, err error) {
	Dialer := websocket.Dialer{
		Proxy:            http.ProxyFromEnvironment,
		HandshakeTimeout: 45 * time.Second,
		// EnableCompression: false,
	}

	L.Infow("Start ws serve", "symbols", params)

	conn, _, err := Dialer.Dial(*wsEndpoint, nil)
	if err != nil {
		return nil, nil, err
	}
	conn.SetReadLimit(655350)
	err = conn.WriteJSON(newMsg(params))
	if err != nil {
		return nil, nil, err
	}

	doneC = make(chan struct{})
	stopC = make(chan struct{})
	go func() {
		// This function will exit either on error from
		// websocket.Conn.ReadMessage or when the stopC channel is
		// closed by the client.
		defer close(doneC)
		keepAlive(conn, time.Minute)
		// Wait for the stopC channel to be closed.  We do that in a
		// separate goroutine because ReadMessage is a blocking
		// operation.
		silent := false
		go func() {
			select {
			case <-stopC:
				silent = true
			case <-doneC:
			}
			conn.Close()
		}()
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				if !silent {
					e(err)
				}
				return
			}
			h(message)
		}
	}()
	return
}

func keepAlive(c *websocket.Conn, timeout time.Duration) {
	ticker := time.NewTicker(timeout)

	lastResponse := time.Now()
	c.SetPongHandler(func(msg string) error {
		lastResponse = time.Now()
		return nil
	})

	go func() {
		defer ticker.Stop()
		for {
			deadline := time.Now().Add(10 * time.Second)
			err := c.WriteControl(websocket.PingMessage, []byte{}, deadline)
			if err != nil {
				return
			}
			<-ticker.C
			if time.Since(lastResponse) > timeout {
				c.Close()
				return
			}
		}
	}()
}

type msg struct {
	Method string   `json:"method"`
	Params []string `json:"params"`
	ID     int64    `json:"id"`
}

func newMsg(params []string) msg {
	return msg{
		Method: "SUBSCRIBE",
		Params: params,
		ID:     time.Now().UnixMilli(),
	}
}
