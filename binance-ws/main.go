package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"time"

	"go.uber.org/zap"
)

var (
	debug              = flag.Bool("debug", true, "Enable debug logs")
	wsEndpoint         = flag.String("websocket", "wss://fstream.binance.com/ws", "Websocket API endpoint")
	gatherDataDuration = flag.Duration("gather-data-duration", time.Hour, "Gather data duration")
	storagePath        = flag.String("storage-path", "data/", "Path to storage output file")
	L                  *zap.SugaredLogger
	params             = []string{
		"btcusdt@bookTicker",
		"ethusdt@bookTicker",
		"bnbusdt@bookTicker",
		"maticusdt@bookTicker",
		"imxusdt@bookTicker",
		"arbusdt@bookTicker",
		"aptusdt@bookTicker",
		"aaveusdt@bookTicker",
		"avaxusdt@bookTicker",
		"solusdt@bookTicker",
	}
)

type Event struct {
	Data   []byte
	TimeUs int64
}

func main() {
	flag.Parse()
	L = setupLogger(*debug)
	L.Infow("Start testing", "symbols", params, "wsEndpoint", wsEndpoint,
		"gatherDataDuration", gatherDataDuration, "storagePath", storagePath)

	singleSymbolData := make([]Event, 0)
	multiSymbolData := make([]Event, 0)

	handlerMulti := func(data []byte) {
		multiSymbolData = append(multiSymbolData, Event{json.RawMessage(data), time.Now().UnixMicro()})
	}
	handlerSingle := func(data []byte) {
		singleSymbolData = append(singleSymbolData, Event{json.RawMessage(data), time.Now().UnixMicro()})
	}

	errH := func(e error) {
		L.Errorw("Received err", "err", e)
	}

	go WsServe(params, handlerMulti, errH)
	for _, p := range params {
		go WsServe([]string{p}, handlerSingle, errH)
	}

	time.Sleep(*gatherDataDuration)

	// store data
	now := time.Now().UnixMilli()
	saveData(singleSymbolData, fmt.Sprintf("single_%d.json", now))
	saveData(multiSymbolData, fmt.Sprintf("multi_%d.json", now))
}
