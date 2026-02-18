package core

import (
	"testing"
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
	if jobs[3].end != 3_500 {
		t.Errorf("Expected last job to end at 3_500, got %d", jobs[3].end)
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
	if jobs[0].end != 10 {
		t.Errorf("Expected last job to end at 10, got %d", jobs[0].end)
	}
}
