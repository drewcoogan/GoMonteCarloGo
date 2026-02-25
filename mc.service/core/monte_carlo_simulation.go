package core

import (
	"fmt"
	"log"
	"maps"
	"math"
	"slices"
	"time"

	"golang.org/x/sync/errgroup"

	ex "mc.data/extensions"
	dm "mc.data/models"
	ms "mc.service/models"
	sm "mc.service/models"
)

const (
	Workers               = 8
	BatchSize             = 10_000
	InitialPortfolioValue = 100.0
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
	dm.ScenarioConfigurationComponent
	Returns             []float64
	Dates               []time.Time
	AnnualizationFactor int
}

type job struct {
	start int
	end   int
}

func GetNumberOfJobsAndWorkers(iterations int, batchSize int, workers int) ([]job, int) {
	// we want to figure out how to divide the work into batches and how many workers to use given the max number of workers and batch size
	// take the total number of simulations and divide it by the batch size, round up to the nearest int to get total number of batches
	nJobs := int(math.Ceil(float64(iterations) / float64(batchSize)))

	// we have a max number of workers, so we take the minimum of the number of jobs and the number of workers
	nWorkers := ex.Min(nJobs, workers)

	// jobs will store what index the job starts and ends at, truncating the last job to number of iterations if needed
	jobs := make([]job, nJobs)
	for i := range nJobs {
		jobs[i] = job{
			start: i * batchSize,
			end:   ex.Min((i+1)*batchSize, iterations) - 1, // -1 because a batch would be 0 -> batch size - 1
		}
	}

	return jobs, nWorkers
}

// RunMonteCarloSimulation runs the monte carlo simulation, abstracted out the
func (sc *ServiceContext) RunMonteCarloSimulation(statisticalResources *StatisticalResources, simulationSettings sm.SimulationRequestSettings) ([]*SimulationResult, error) {
	res := make([]*SimulationResult, simulationSettings.Iterations)

	jobs, nWorkers := GetNumberOfJobsAndWorkers(simulationSettings.Iterations, BatchSize, Workers)

	log.Println("Starting monte carlo simulation:")
	log.Printf("\t Simulation duration: %v %s", simulationSettings.SimulationDuration, ms.ConvertFrequencyToString(simulationSettings.SimulationUnitOfTime))
	log.Printf("\t Simulation paths: %v", simulationSettings.Iterations)
	log.Printf("\t Simulation batch size: %v", BatchSize)
	log.Printf("\t Workers: %v", Workers)

	// this is the channel that will hold all the jobs to be processed, workers will steal jobs from this channel as they process other jobs
	jobsChannel := make(chan job, len(jobs))
	for _, v := range jobs {
		jobsChannel <- v
	}
	close(jobsChannel) // close the job channel, there isnt anything else being added to it

	// using the service context to DERIVE the err group context here will allow for a few things:
	// if a user cancels the request, the simulations will also be cancelled
	// if a worker errors, it wont take down the user's context
	group, ctx := errgroup.WithContext(sc.Context)

	for i := range nWorkers {
		workerResource := NewWorkerResources(statisticalResources, uint64(simulationSettings.Seed), uint64(i+1))
		group.Go(func() error {
			// this will loop over available jobs, and will reup if a job finishes and there are more jobs
			for j := range jobsChannel {
				// "select" is a golang keyword that is used primarily for channels
				// here, we are checking if the context is done, and if it is, we return the error
				// if it is not done, we continue with the loop
				select {
				case <-ctx.Done():
					return ctx.Err()
				default:
				}

				for sim := j.start; sim <= j.end; sim++ { // this will loop over the iterations
					portfolioValue := InitialPortfolioValue
					pathValues := make([]float64, simulationSettings.SimulationDuration+1)
					pathValues[0] = portfolioValue

					for period := range simulationSettings.SimulationDuration { // this will loop over the time steps for the duration by the unit of time
						correlatedReturns := workerResource.GetCorrelatedReturns(simulationSettings.SimulationUnitOfTime)
						portfolioReturn, err := ex.DotProduct(statisticalResources.AssetWeight, correlatedReturns)
						if err != nil {
							log.Printf("error calculating dot product in resource worker for simulation %d: %v", sim, err)
							return err
						}

						portfolioValue *= math.Exp(portfolioReturn)
						pathValues[period+1] = portfolioValue
					}

					pathMetrics := calculatePathMetrics(pathValues, simulationSettings.SimulationUnitOfTime)

					res[sim] = &SimulationResult{
						PathMetrics: pathMetrics,
						PathValues:  pathValues,
					}
				}
			}

			return nil
		})
	}

	if err := group.Wait(); err != nil {
		return nil, err
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

func (sc *ServiceContext) getSeriesReturns(scenario *dm.Scenario, maxLookback time.Time) ([]*SeriesReturns, error) {
	tickerLookup := make(map[int32]dm.ScenarioConfigurationComponent, len(scenario.Components))
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
				AnnualizationFactor:            ms.Weekly, // TODO: leaving as hard coded for now, need to verify this works with other than weekly
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
	for i, v := range data {
		first, last, length := getTimeRange(v)
		firstDates[i] = first
		lastDates[i] = last
		lengths[i] = length
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
