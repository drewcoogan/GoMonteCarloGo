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

	slices.SortFunc(res, func(i, j SeriesReturns) int {
		return int(i.SourceID - j.SourceID)
	})

	for i, r := range res {
		res[i].MeanReturn = stat.Mean(r.Returns, nil)
		res[i].StdDev = stat.StdDev(r.Returns, nil)
	}

	return
}
