package core

import (
	"fmt"
	"maps"
	"slices"
	"time"

	//"math/rand/v2"

	//"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat"
	//"gonum.org/v1/gonum/stat/distuv"

	ex "mc.data/extensions"
)

type SimulationAllocation struct {
	Id     int32   `json:"id"`
	Ticker string  `json:"ticker"`
	Weight float64 `json:"weight"`
}

type SimulationRequest struct {
	Allocations []SimulationAllocation `json:"allocations"`
	Iterations  int                    `json:"iterations"`
	MaxLookback time.Duration
	// other stuff?
}

type SeriesReturns struct {
	SourceID   int32
	Ticker     string
	Returns    []float64
	Dates      []time.Time
	MeanReturn float64
	StdDev     float64
}

func (sc *ServiceContext) RunEquityMonteCarloWithCovarianceMartix(request SimulationRequest) error {
	res, err := sc.getSeriesReturns(request)

	if err != nil {
		return err
	}

	return nil
}

func (sc *ServiceContext) getSeriesReturns(request SimulationRequest) (res []SeriesReturns, err error) {
	tickerLookup := make(map[int32]string, len(request.Allocations))
	for _, allocation := range request.Allocations {
		tickerLookup[allocation.Id] = allocation.Ticker
	}

	returns, err := sc.PostgresConnection.GetTimeSeriesReturns(sc.Context, slices.Collect(maps.Keys(tickerLookup)), request.MaxLookback)
	if err != nil {
		return res, fmt.Errorf("error getting time series returns: %v", err)
	}

	agg := make(map[int32]*SeriesReturns, len(request.Allocations))
	for _, ret := range returns {
		if agg[ret.Id] == nil {
			agg[ret.Id] = &SeriesReturns{
				SourceID: ret.Id,
				Ticker:   tickerLookup[ret.Id],
				Returns:  []float64{},
				Dates:    []time.Time{},
			}
		}

		agg[ret.Id].Returns = append(agg[ret.Id].Returns, ret.LogReturn)
		agg[ret.Id].Dates = append(agg[ret.Id].Dates, ret.Timestamp)
	}

	for _, tickerAgg := range agg {
		res = append(res, *tickerAgg)
	}

	// sorts on source id for consistency, maybe sort on weight? could also arrange that in front end.
	slices.SortFunc(res, func(i, j SeriesReturns) int {
		return int(i.SourceID - j.SourceID)
	})

	for i, r := range res {
		res[i].MeanReturn = stat.Mean(r.Returns, nil)
		res[i].StdDev = stat.StdDev(r.Returns, nil)
	}

	if err = verifyDataIntegrity(res); err != nil {
		return
	}

	return
}

func verifyDataIntegrity(data []SeriesReturns) error {
	firstDates := make([]time.Time, len(data))
	lastDates := make([]time.Time, len(data))
	lengths := make([]int, len(data))
	for _, v := range data {
		first, last, length := getTimeRange(v)
		firstDates = append(firstDates, first)
		lastDates = append(lastDates, last)
		lengths = append(lengths, length)
	}

	if ex.AreAllEqual(firstDates) {
		return fmt.Errorf("data validation failed, first dates in range do not align")
	}

	if ex.AreAllEqual(lastDates) {
		return fmt.Errorf("data validation failed, last dates in range do not align")
	}

	if ex.AreAllEqual(lengths) {
		return fmt.Errorf("data validation failed, length of dates in range do not align")
	}

	return nil
}

func getTimeRange(data SeriesReturns) (first, last time.Time, length int) {
	maxUnixSeconds := int64(1<<63 - 1 - 62135596801)
	first = time.Unix(maxUnixSeconds, 999999999)
	for _, v := range data.Dates {
		if v.Before(first) {
			first = v
		}
		if v.After(last) {
			last = v
		}
		length++
	}
	return
}
