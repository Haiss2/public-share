package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	fix "github.com/KyberNetwork/binance_fix_api"
	"github.com/KyberNetwork/go-binance/v2"
	"github.com/quickfixgo/enum"
	"golang.org/x/sync/errgroup"
)

const (
	outputPath         = "./"
	configFilePath     = "./fix.conf"
	apiKey             = ""
	privateKeyFilePath = ""
	binanceApiKey      = ""
	binanceSecretKey   = ""
)

func main() {
	logger := SetupLogger()

	settings, err := fix.LoadQuickfixSettings(configFilePath)
	if err != nil {
		logger.Errorw("Failed to LoadQuickfixSettings", "err", err)
		return
	}

	conf := fix.Config{
		APIKey:             apiKey,
		PrivateKeyFilePath: privateKeyFilePath,
		Settings:           settings,
	}
	fixClient, err := fix.NewClient(
		context.Background(), logger, conf,
		fix.WithResponseModeOpt(fix.ResponseModeOnlyAcks),
		fix.WithMessageHandlingOpt(fix.MessageHandlingUnordered),
	)
	if err != nil {
		logger.Errorw("Failed to init client", "err", err)
		return
	}

	defer fixClient.Stop()

	logger.Info("Everything is ready!")

	// Prepare for CSV
	header := []string{"symbol", "qty", "price", "side", "tif", "start_at",
		"fix_response_at", "rest_response_at", "fix_transact_time", "rest_transact_time",
		"fix_latency", "rest_latency", "fix_response_time", "rest_response_time"}
	data := [][]string{}

	// TRY TO PLACE ORDER
	restClient := binance.NewClient(binanceApiKey, binanceSecretKey)
	type test struct {
		Symbol     string
		Price, Qty float64
	}
	var tests []test

	serverTime, _ := restClient.NewServerTimeService().Do(context.Background())
	delta := serverTime - time.Now().UnixMilli()

	tickers, err := restClient.NewListPriceChangeStatsService().Do(context.Background())
	if err != nil {
		logger.Errorw("Failed to get binance ticker", "err", err)
		return
	}

	count := 0
	for _, ticker := range tickers {
		if count >= 50 {
			break
		}
		if ticker.Symbol != "DOCKUSDT" && strings.HasSuffix(ticker.Symbol, "USDT") {
			lastPx := StringToFloat(ticker.LastPrice) * 0.95
			lastPx = round(lastPx)
			if lastPx == 0 {
				continue
			}
			rawQty := 55 / lastPx
			tests = append(tests, test{
				Symbol: ticker.Symbol,
				Price:  round(lastPx),
				Qty:    round(rawQty),
			})
			count += 1
		}
	}

	logger.Infow("Setup test data", "data", tests)
	for _, test := range tests {
		var (
			now                               = time.Now().UnixMilli()
			eg                                errgroup.Group
			fixRespTime, restRespTime         int64
			fixTransactTime, restTransactTime int64
		)

		// place FIX order
		eg.Go(func() error {
			order, err := fixClient.NewOrderSingleService().
				Symbol(test.Symbol).
				Side(enum.Side_BUY).
				Type(enum.OrdType_LIMIT).
				TimeInForce(enum.TimeInForce_IMMEDIATE_OR_CANCEL).
				Price(test.Price).
				Quantity(test.Qty).
				Do(context.Background())
			fixRespTime = time.Now().UnixMilli()
			if err != nil {
				logger.Errorw("Failed to place fix order", "err", err)
				return err
			}
			fixTransactTime = order.TransactTime.UnixMilli()
			return nil
		})

		// place rest API order
		eg.Go(func() error {
			order, err := restClient.NewCreateOrderService().
				Symbol(test.Symbol).
				Side(binance.SideTypeBuy).
				Type(binance.OrderTypeLimit).
				TimeInForce(binance.TimeInForceTypeIOC).
				Price(FloatToString(test.Price)).
				Quantity(FloatToString(test.Qty)).
				NewOrderRespType(binance.NewOrderRespTypeFULL).
				Do(context.Background())
			restRespTime = time.Now().UnixMilli()
			if err != nil {
				logger.Errorw("Failed to place rest order", "err", err)
				return err
			}
			restTransactTime = order.TransactTime
			return nil
		})
		if err := eg.Wait(); err != nil {
			logger.Errorw("Failed to init client", "err", err)
			return
		} else {
			// "Symbol", "Quantity", "Price", "Side", "TimeInForce", "StartTime",
			// "FixRespTime", "RestRespTime", "FixTransactTime", "RestTransactTime", "FixLatency", "RestLatency", "FixDiffRespTime", "RestDiffRespTime"
			data = append(data, []string{
				test.Symbol, FloatToString(test.Qty), FloatToString(test.Price), "BUY", "IOC", IntToString(now),
				IntToString(fixRespTime), IntToString(restRespTime),
				IntToString(fixTransactTime), IntToString(restTransactTime),
				IntToString(fixTransactTime - now - delta), IntToString(restTransactTime - now - delta),
				IntToString(fixRespTime - now), IntToString(restRespTime - now),
			})

			time.Sleep(time.Duration(rand.Intn(1000)+1) * time.Millisecond)
		}
	}

	if err := WriteCSV(header, data); err != nil {
		logger.Errorw("Failed to WriteCSV", "err", err)
		return
	}

	logger.Info("CSV file written successfully")
}

func WriteCSV(header []string, data [][]string) error {
	// Create a new CSV file
	file, err := os.Create(fmt.Sprintf("%s/test_%d.csv", outputPath, time.Now().Unix()))
	if err != nil {
		return err
	}
	defer file.Close()

	// Create a CSV writer
	writer := csv.NewWriter(file)
	defer writer.Flush()

	if err := writer.Write(header); err != nil {
		return err
	}

	for _, record := range data {
		if err := writer.Write(record); err != nil {
			return err
		}
	}

	return nil
}
