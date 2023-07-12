package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

var (
	debug              = flag.Bool("debug", true, "Enable debug logs")
	wsEndpoint         = flag.String("websocket", "wss://fstream.binance.com/ws", "Websocket API endpoint")
	gatherDataDuration = flag.Duration("gather-data-duration", time.Hour, "Gather data duration")
	filePrefix         = flag.String("file-prefix", "future", "File name prefix")
	storagePath        = flag.String("storage-path", "data/", "Path to storage output file")
	L                  *zap.SugaredLogger
	symbols            = flag.String("symbols",
		"btcusdt,ethusdt,bnbusdt,maticusdt,imxusdt,arbusdt,aptusdt,aaveusdt,avaxusdt,solusdt",
		"symbols to receive bookticker ws events")
)

type Event struct {
	Data   []byte
	TimeUs int64
}

func (e Event) toBookTicker() (BookTicker, error) {
	var b BookTicker
	err := json.Unmarshal(e.Data, &b)
	if err != nil {
		return b, err
	}
	b.ActualTimeUs = e.TimeUs
	return b, nil
}

func toSubscription(s string) string {
	return s + "@bookTicker"
}

var errH = func(e error) {
	L.Errorw("Received err", "err", e)
}

func main() {
	flag.Parse()
	L = setupLogger(*debug)
	L.Infow("Start testing", "symbols", symbols, "wsEndpoint", wsEndpoint,
		"gatherDataDuration", gatherDataDuration, "storagePath", storagePath)

	symbolSlice := strings.Split(*symbols, ",")
	params := make([]string, len(symbolSlice))
	for id, s := range symbolSlice {
		params[id] = toSubscription(s)
	}

	ctx, cancel := context.WithTimeout(context.Background(), *gatherDataDuration)
	defer cancel()

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		handleConnection(ctx, params, "multi")
	}()
	for _, s := range symbolSlice {
		wg.Add(1)
		go func(s string) {
			defer wg.Done()
			handleConnection(ctx, []string{toSubscription(s)}, s)
		}(s)
	}

	// wait for gathering!
	wg.Wait()
}

func handleConnection(ctx context.Context, params []string, file string) {
	data := make([]Event, 0)
	handler := func(e []byte) {
		data = append(data, Event{e, time.Now().UnixMicro()})
	}

	_, stopC, err := WsServe(params, handler, errH)
	if err != nil {
		L.Errorw("Fail to WsServe", "err", err, "symbols", params)
	}

	<-ctx.Done()
	stopC <- struct{}{}

	now := time.Now().UnixMilli()
	saveData(parseData(data), fmt.Sprintf("%s_%s_%d.json", *filePrefix, file, now))
}

func parseData(events []Event) []BookTicker {
	res := make([]BookTicker, len(events))
	var err error
	for id, e := range events {
		res[id], err = e.toBookTicker()
		if err != nil {
			L.Errorw("Fail to Unmarshal event", "err", err, "event", e)
		}
	}
	return res
}
