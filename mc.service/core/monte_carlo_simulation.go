package core

import (
	"fmt"
	"log"
	"maps"
	"math"
	"slices"
	"time"

	ex "mc.data/extensions"
	m "mc.data/models"
)

const (
	Workers   = 8
	BatchSize = 10_000
)

// TODO: this is being abstracted out a bit to the simulation controller
// not need this exactly, or what that will look like, but in theory this is going to juts be called from the opt controller
type SimulationSettings struct {
	MaxLookback time.Duration `json:"maxlookback"` // what is this going to look like in api req?

	Iterations int   `json:"iterations"`
	Seed       int64 `json:"seed"`     // ^^
	DistType   int   `json:"disttype"` // standar normal, student t

	SimulationUnitOfTime int `json:"simulationunitoftime"` // daily, weekly, monthly, quarterly, yearly
	SimulationDuration   int `json:"simulationduration"`   // number of units of time to simulate
	DegreesOfFreedom     int `json:"degreesoffreedom"`     // degrees of freedom for student t distribution
}

type SimulationResult struct {
	PathMetrics
	PathValues []float64
}

type PathMetrics struct {
	FinalValue           float64
	TotalReturn          float64
	AnnualizedReturn     float64
	AnnualizedVolatility float64
	MaxDrawdown          float64
}

type SeriesReturns struct {
	m.ScenarioConfigurationComponent
	Returns             []float64
	Dates               []time.Time
	AnnualizationFactor int
}

type job struct {
	index int
	start int
	end   int
}

func (sc *ServiceContext) RunMonteCarloSimulation(scenario *m.Scenario, simulation SimulationSettings) ([]*SimulationResult, error) {
	res := make([]*SimulationResult, simulation.Iterations)
	seriesReturns, err := sc.getSeriesReturns(scenario, simulation.MaxLookback)
	if err != nil {
		return res, err
	}

	statisticalResources, err := GetStatisticalResources(simulation, seriesReturns)
	if err != nil {
		return res, err
	}

	nJobs := int(math.Ceil(float64(simulation.Iterations) / BatchSize / Workers))
	if nJobs == 0 && simulation.Iterations > 0 {
		nJobs = 1
	}

	log.Println("Starting monte carlo simulation:")
	log.Printf("\t Simulation duration: %v %s", simulation.SimulationDuration, convertFrequencyToString(simulation.SimulationUnitOfTime))
	log.Printf("\t Simulation paths: %v", simulation.Iterations)
	log.Printf("\t Simulation batch size: %v", BatchSize)
	log.Printf("\t Workers: %v", Workers)

	workerCount := ex.Min(nJobs, Workers)
	statisticalResource := NewWorkerResources(statisticalResources, uint64(simulation.Seed), 0)

	jobs := make(chan job, nJobs) // TODO: if njobs is less than workers, take the minimum
	done := make(chan bool, ex.Min(nJobs, Workers))

	worker := func(wr *WorkerResource) {
		for j := range jobs { // this will loop over available jobs, and will reup if a job finishes and there are more jobs
			for sim := j.start; sim < j.end; sim++ { // this will loop over the iterations
				portfolioValue := 100.0
				pathValues := make([]float64, simulation.SimulationDuration+1)
				pathValues[0] = portfolioValue

				for period := range simulation.SimulationDuration {
					correlatedReturns := wr.GetCorrelatedReturns(simulation.SimulationUnitOfTime)
					portfolioReturn, err := ex.DotProduct(statisticalResources.AssetWeight, correlatedReturns)
					if err != nil {
						log.Printf("error calculating dot product in resource worker for simulation %d: %v", sim, err)
						return // TODO: how to handle errors in channels? <-- good and important question
					}

					portfolioValue *= math.Exp(portfolioReturn)
					pathValues[period+1] = portfolioValue
				}

				pathMetrics := calculatePathMetrics(pathValues, simulation.SimulationUnitOfTime)

				res[sim] = &SimulationResult{
					PathMetrics: pathMetrics,
					PathValues:  pathValues,
				}
			}
		}
		done <- true
	}

	// starts the workers
	for i := range workerCount {
		go worker(workerResources[i])
	}

	// allocate the jobs and the respective dist index, start and end iteration indicies for result allocation
	for i := range nJobs {
		start := i * BatchSize
		end := ex.Min(start+BatchSize, simulation.Iterations)
		if start != end {
			jobs <- job{index: i, start: start, end: end}
		}
	}
	close(jobs) // close the job channel, there isnt anything else being added to it

	// this will loop until all of the jobs are complete
	for range workerCount {
		<-done
	}

	return res, nil
}

func calculatePathMetrics(pathValues []float64, simulationUnitOfTime int) PathMetrics {
	n := len(pathValues)

	var sumReturns, sumSquaredReturns, maxDrawdown, peak float64
	logReturns := make([]float64, n-1)

	for i := range n {
		if i != 0 {
			logReturn := math.Log(pathValues[i] / pathValues[i-1])
			logReturns[i-1] = logReturn
			sumReturns += logReturn
			sumSquaredReturns += logReturn * logReturn
		}

		if pathValues[i] > peak {
			peak = pathValues[i]
		}

		drawdown := (peak - pathValues[i]) / peak
		if drawdown > maxDrawdown {
			maxDrawdown = drawdown
		}
	}

	initialValue := pathValues[0]
	finalValue := pathValues[n-1]
	totalReturn := (finalValue - initialValue) / initialValue

	// general required factors for annualization
	numPeriods := float64(n - 1)
	periodsPerYear := float64(simulationUnitOfTime)

	// annualized return: geometric mean of returns, spelling this out to be explicit
	totalLogReturn := math.Log(finalValue / initialValue)
	annualizedReturn := math.Exp(totalLogReturn*periodsPerYear/numPeriods) - 1.0

	// annualized volatility: sample standard deviation of log returns, spelling this out to be explicit
	meanReturn := sumReturns / numPeriods
	variance := (sumSquaredReturns - numPeriods*meanReturn*meanReturn) / (numPeriods - 1)
	periodVolatility := math.Sqrt(variance)
	annualizedVolatility := periodVolatility * math.Sqrt(periodsPerYear)

	return PathMetrics{
		FinalValue:           finalValue,
		TotalReturn:          totalReturn,
		AnnualizedReturn:     annualizedReturn,
		AnnualizedVolatility: annualizedVolatility,
		MaxDrawdown:          maxDrawdown,
	}
}

func (sc *ServiceContext) getSeriesReturns(scenario *m.Scenario, maxLookback time.Duration) ([]*SeriesReturns, error) {
	tickerLookup := make(map[int32]m.ScenarioConfigurationComponent, len(scenario.Components))
	for _, component := range scenario.Components {
		tickerLookup[component.AssetId] = component
	}

	returns, err := sc.PostgresConnection.GetTimeSeriesReturns(sc.Context, slices.Collect(maps.Keys(tickerLookup)), maxLookback)
	if err != nil {
		return nil, fmt.Errorf("error getting time series returns: %v", err)
	}

	agg := make(map[int32]*SeriesReturns, len(scenario.Components))
	for _, ret := range returns {
		if agg[ret.Id] == nil {
			agg[ret.Id] = &SeriesReturns{
				ScenarioConfigurationComponent: tickerLookup[ret.Id],
				Returns:                        []float64{},
				Dates:                          []time.Time{},
				AnnualizationFactor:            Weekly, // TODO: leaving as hard coded for now, need to verify this works with other than weekly
			}
		}

		agg[ret.Id].Returns = append(agg[ret.Id].Returns, ret.LogReturn)
		agg[ret.Id].Dates = append(agg[ret.Id].Dates, ret.Timestamp)
	}

	res := make([]*SeriesReturns, 0, len(agg))
	for _, tickerAgg := range agg {
		res = append(res, tickerAgg)
	}

	// sorts on source id for consistency, useful for testing
	slices.SortFunc(res, func(i, j *SeriesReturns) int {
		return int(i.AssetId - j.AssetId)
	})

	if err := verifySeriesReturnIntegrity(res); err != nil {
		return nil, err
	}

	return res, nil
}

func verifySeriesReturnIntegrity(data []*SeriesReturns) error {
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

func getTimeRange(data *SeriesReturns) (time.Time, time.Time, int) {
	// i dont think we need to keep the dates in the same order... but i dont want to find out
	dates := slices.Clone(data.Dates)
	slices.SortFunc(dates, func(i, j time.Time) int {
		return i.Compare(j)
	})

	first := dates[0]
	length := len(dates)
	last := dates[length-1]

	if last.Before(first) {
		log.Println("you dummy you missed the multipler in getTimeRange() sort compare")
	}

	return first, last, length
}
