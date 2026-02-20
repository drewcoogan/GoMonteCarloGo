package core

import (
	"context"
	"testing"
	"time"

	sm "mc.service/models"
)

func TestJobsAndIterationsLogicIsCorrect(t *testing.T) {
	iterations := 10_000
	batchSize := 1_000
	maximumAvailableWorkers := 4

	jobs, nWorkers := GetNumberOfJobsAndWorkers(iterations, batchSize, maximumAvailableWorkers)

	if len(jobs) != 10 {
		t.Errorf("Expected 10 jobs, got %d", len(jobs))
	}
	if nWorkers != 4 {
		t.Errorf("Expected 4 workers, got %d", nWorkers)
	}
	if jobs[0].end == jobs[1].start {
		t.Errorf("Expected second job to start after the first job ends")
	}

	// should have 4 jobs (last one should have 500 iterations)
	iterations = 3_500
	batchSize = 1_000
	maximumAvailableWorkers = 4

	jobs, nWorkers = GetNumberOfJobsAndWorkers(iterations, batchSize, maximumAvailableWorkers)

	// TODO make sure there is no overlap, like end of job 1 is 1000, start of job 2 is 1001, etc.

	if len(jobs) != 4 {
		t.Errorf("Expected 4 jobs, got %d", len(jobs))
	}
	if nWorkers != 4 {
		t.Errorf("Expected 4 workers, got %d", nWorkers)
	}
	if jobs[3].end != 3_500-1 {
		t.Errorf("Expected last job to end at 3_499 (remember, index 0), got %d", jobs[3].end)
	}

	iterations = 10
	batchSize = 1_000
	maximumAvailableWorkers = 4

	jobs, nWorkers = GetNumberOfJobsAndWorkers(iterations, batchSize, maximumAvailableWorkers)

	if len(jobs) != 1 {
		t.Errorf("Expected 1 job, got %d", len(jobs))
	}
	if nWorkers != 1 {
		t.Errorf("Expected 1 worker, got %d", nWorkers)
	}
	if jobs[0].start != 0 {
		t.Errorf("Expected first job to start at 0, got %d", jobs[0].start)
	}
	if jobs[0].end != 9 {
		t.Errorf("Expected last job to end at 10, got %d", jobs[0].end)
	}
}

// TestRunMonteCarloSimulation_SingleAsset runs the simulation with one asset, fixed seed, and enough iterations to use workers. Uses standard normal.
func TestRunMonteCarloSimulation_SingleAsset(t *testing.T) {
	nSamples := sm.Daily * 500
	allReturns := GenerateMockSeriesReturns(t, nSamples) // from statistics_test.go
	singleAssetReturns := allReturns[:1]

	settings := sm.SimulationRequestSettings{
		DistType:             sm.StandardNormal,
		SimulationUnitOfTime: sm.Daily,
		SimulationDuration:   252,
		Iterations:           25_000, // > BatchSize to utilize multiple workers
		Seed:                 42,
	}

	sr, err := GetStatisticalResources(singleAssetReturns, settings)
	if err != nil {
		t.Fatalf("GetStatisticalResources: %v", err)
	}

	sc := &ServiceContext{Context: context.Background()}

	start := time.Now()
	res, err := sc.RunMonteCarloSimulation(sr, settings)
	elapsed := time.Since(start)
	t.Logf("RunMonteCarloSimulation (1 asset, %d iterations): %v", settings.Iterations, elapsed)

	if err != nil {
		t.Fatalf("RunMonteCarloSimulation: %v", err)
	}
	if len(res) != settings.Iterations {
		t.Errorf("expected %d results, got %d", settings.Iterations, len(res))
	}
	for i, r := range res {
		if r == nil {
			t.Errorf("result[%d] is nil", i)
			break
		}
		if len(r.PathValues) != settings.SimulationDuration+1 {
			t.Errorf("result[%d]: expected PathValues length %d, got %d", i, settings.SimulationDuration+1, len(r.PathValues))
		}
		if r.PathValues[0] != InitialPortfolioValue {
			t.Errorf("result[%d]: expected initial value %f, got %v", i, InitialPortfolioValue, r.PathValues[0])
		}
	}
}
