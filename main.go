package main

import (
	"geth/cmon"
	"sort"
	"time"
)

func main() {
	multiSMap := make(map[int64]*WsBookTickerEvent)
	singleSMap := make(map[int64]*WsBookTickerEvent)

	params := []string{
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
	handler1 := func(event *WsBookTickerEvent) {
		if event.Symbol == "SOLUSDT" {
			event.ActualTimeUs = time.Now().UnixMicro()
			multiSMap[event.UpdateID] = event
		}
	}

	handler2 := func(event *WsBookTickerEvent) {
		event.ActualTimeUs = time.Now().UnixMicro()
		singleSMap[event.UpdateID] = event
	}
	errH := func(e error) {
		cmon.L.Errorw("Received err", "err", e)
	}

	go WsServe(params, handler1, errH)
	WsServe([]string{"solusdt@bookTicker"}, handler2, errH)

	time.Sleep(60 * time.Second)

	// Compare result
	cmon.L.Infow("Multi symbols test event length", "len", len(multiSMap))
	cmon.L.Infow("Single symbol test event length", "len", len(singleSMap))

	intersectionNum := 0
	multiFasterArr := make([]int64, 0)
	singleFasterArr := make([]int64, 0)

	var (
		maxMSDiff, maxMSID int64
		maxSMDiff, maxSMID int64
	)
	for updateID, mEvent := range multiSMap {
		if sEvent, ok := singleSMap[updateID]; ok {
			intersectionNum++
			if mEvent.ActualTimeUs > sEvent.ActualTimeUs {
				diff := mEvent.ActualTimeUs - sEvent.ActualTimeUs
				if diff > maxMSDiff {
					maxMSDiff = diff
					maxMSID = updateID
				}
				multiFasterArr = append(multiFasterArr, diff)
			} else {
				diff := sEvent.ActualTimeUs - mEvent.ActualTimeUs
				if diff > maxSMDiff {

					maxSMDiff = diff
					maxSMID = updateID
				}
				singleFasterArr = append(singleFasterArr, diff)
			}
		}
	}
	cmon.L.Infow("Intersection count!", "count", intersectionNum)

	// statistic max diff
	sort.Slice(multiFasterArr, func(i, j int) bool {
		return multiFasterArr[i] < multiFasterArr[j]
	})
	sort.Slice(singleFasterArr, func(i, j int) bool {
		return singleFasterArr[i] < singleFasterArr[j]
	})

	cmon.L.Infow("Multi faster than single", "count", len(multiFasterArr),
		"meanUs", cmon.Mean(multiFasterArr),
		"medianUs", multiFasterArr[len(multiFasterArr)/2],
		"maxUs", multiFasterArr[len(multiFasterArr)-1],
		"minUs", multiFasterArr[0],
	)
	cmon.L.Infow("Single faster than multi ", "count", len(singleFasterArr),
		"meanUs", cmon.Mean(singleFasterArr),
		"medianUs", singleFasterArr[len(singleFasterArr)/2],
		"maxUs", singleFasterArr[len(singleFasterArr)-1],
		"minUs", singleFasterArr[0],
	)

	cmon.L.Infow("Biggest diff that multi > single", "multi", multiSMap[maxMSID], "single", singleSMap[maxMSID], "diff", maxMSDiff)
	cmon.L.Infow("Biggest diff that single > multi", "multi", multiSMap[maxSMID], "single", singleSMap[maxSMID], "diff", maxSMDiff)
}
