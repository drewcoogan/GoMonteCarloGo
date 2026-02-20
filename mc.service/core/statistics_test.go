package core

import (
	"math"
	"math/rand/v2"
	"testing"
	"time"

	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/gonum/stat/distuv"
	dm "mc.data/models"
	sm "mc.service/models"
)

const (
	mu_a    = 0.08
	mu_b    = 0.10
	mu_c    = 0.12
	sigma_a = 0.15
	sigma_b = 0.20
	sigma_c = 0.25
	corr_ab = 0.5
	corr_ac = 0.0
	corr_bc = 0.0
)

// TestSupportingGenerators ensures the math is correct for supporting testing functionality
func TestSupportingGenerators(t *testing.T) {
	nSamples := sm.Daily * 500
	returns := generateMockReturns(t, nSamples)

	// verify return correlation
	eval_corr_ab := stat.Correlation(returns[0], returns[1], nil)
	eval_corr_ac := stat.Correlation(returns[0], returns[2], nil)
	eval_corr_bc := stat.Correlation(returns[1], returns[2], nil)

	correlationTolerance := 0.01
	if math.Abs(eval_corr_ab-corr_ab) > correlationTolerance {
		t.Errorf("Corr(Asset A, Asset B): expected %.4f, got %.4f", corr_ab, eval_corr_ab)
	}

	if math.Abs(eval_corr_ac-corr_ac) > correlationTolerance {
		t.Errorf("Corr(Asset A, Asset C): expected %.4f, got %.4f", corr_ac, eval_corr_ac)
	}

	if math.Abs(eval_corr_bc-corr_bc) > correlationTolerance {
		t.Errorf("Corr(Asset B, Asset C): expected %.4f, got %.4f", corr_bc, eval_corr_bc)
	}

	// verify annualized mean
	eval_mu_a := stat.Mean(returns[0], nil) * sm.Daily
	eval_mu_b := stat.Mean(returns[1], nil) * sm.Daily
	eval_mu_c := stat.Mean(returns[2], nil) * sm.Daily
	drift_adjusted_mu_a := calculateDriftAdjustedMu(t, mu_a, sigma_a)
	drift_adjusted_mu_b := calculateDriftAdjustedMu(t, mu_b, sigma_b)
	drift_adjusted_mu_c := calculateDriftAdjustedMu(t, mu_c, sigma_c)

	muTolerance := 0.01
	if math.Abs(eval_mu_a-drift_adjusted_mu_a) > muTolerance {
		t.Errorf("Mu(Asset A): expected %.4f, got %.4f", drift_adjusted_mu_a, eval_mu_a)
	}

	if math.Abs(eval_mu_b-drift_adjusted_mu_b) > muTolerance {
		t.Errorf("Mu(Asset B): expected %.4f, got %.4f", drift_adjusted_mu_b, eval_mu_b)
	}

	if math.Abs(eval_mu_c-drift_adjusted_mu_c) > muTolerance {
		t.Errorf("Mu(Asset C): expected %.4f, got %.4f", drift_adjusted_mu_c, eval_mu_c)
	}

	// verify standard deviation
	eval_sigma_a := stat.StdDev(returns[0], nil) * math.Sqrt(sm.Daily)
	eval_sigma_b := stat.StdDev(returns[1], nil) * math.Sqrt(sm.Daily)
	eval_sigma_c := stat.StdDev(returns[2], nil) * math.Sqrt(sm.Daily)

	sigmaTolerance := 0.01
	if math.Abs(eval_sigma_a-sigma_a) > sigmaTolerance {
		t.Errorf("Sigma(Asset A): expected %.4f, got %.4f", sigma_a, eval_sigma_a)
	}

	if math.Abs(eval_sigma_b-sigma_b) > sigmaTolerance {
		t.Errorf("Sigma(Asset B): expected %.4f, got %.4f", sigma_b, eval_sigma_b)
	}

	if math.Abs(eval_sigma_c-sigma_c) > sigmaTolerance {
		t.Errorf("Sigma(Asset C): expected %.4f, got %.4f", sigma_c, eval_sigma_c)
	}

	prices := generateMockStockPrices(t, returns) // seeds are same, should the same underlying data
	for asset := range returns {
		for day := range len(returns[asset]) {
			calculatedReturn := math.Log(prices[asset][day+1] / prices[asset][day])
			diff := math.Abs(returns[asset][day] - calculatedReturn)
			if diff > 1e-10 {
				t.Errorf("Asset %d, Day %d: return mismatch (diff: %.2e)", asset, day, diff)
			}
		}
	}
}

// TestStatisticalResourcesCalculations verifies that StatisticalResources are created correctly
func TestStatisticalResourcesCalculations(t *testing.T) {
	nSamples := sm.Daily * 500
	returns := GenerateMockSeriesReturns(t, nSamples)
	settings := sm.SimulationRequestSettings{DistType: sm.StandardNormal}

	sr, err := GetStatisticalResources(returns, settings)
	if err != nil {
		t.Fatalf("Failed to create StatisticalResources: %v", err)
	}

	if sr.CovMatrix == nil {
		t.Errorf("Covariance matrix should not be nil")
	}

	if sr.CorrMatrix != nil {
		t.Error("Correlation matrix should be nil for standard normal")
	}

	if sr.CholeskyL == nil {
		t.Error("CholesklyL should not be nil")
	}

	// CholeskyCorrL is used for StandardNormal too (correlated N(0,1) then scale by sigma)
	if sr.CholeskyCorrL == nil {
		t.Error("CholeskyCorrL should not be nil (used for correlated standard normals)")
	}

	if len(sr.Mu) != 3 {
		t.Errorf("Expected 3 assets, got %d", len(sr.Mu))
	}

	expectedMeans := []float64{
		calculateDriftAdjustedMu(t, mu_a, sigma_a),
		calculateDriftAdjustedMu(t, mu_b, sigma_b),
		calculateDriftAdjustedMu(t, mu_c, sigma_c),
	}
	for i, expected := range expectedMeans {
		if math.Abs(sr.Mu[i]-expected) > 0.02 {
			t.Errorf("Asset %d: expected mean ~%.4f, got %.4f", i, expected, sr.Mu[i])
		}
	}

	expectedStds := []float64{sigma_a, sigma_b, sigma_c}
	for i, expected := range expectedStds {
		if math.Abs(sr.Sigma[i]-expected) > 0.02 {
			t.Errorf("Asset %d: expected std ~%.2f, got %.4f", i, expected, sr.Sigma[i])
		}
	}
}

// TestNormalDistributionReturns verifies normal distribution behavior
func TestStatisticalResourcesWorkerCorrelatedReturnsForStandardNormal(t *testing.T) {
	nSamples := sm.Daily * 500
	returns := GenerateMockSeriesReturns(t, nSamples)
	settings := sm.SimulationRequestSettings{DistType: sm.StandardNormal}

	sr, err := GetStatisticalResources(returns, settings)
	if err != nil {
		t.Fatalf("Failed to create StatisticalResources: %v", err)
	}

	worker := NewWorkerResources(sr, 42, 0)

	allReturns := make([][]float64, nSamples)
	for i := range nSamples {
		allReturns[i] = worker.GetCorrelatedReturns(sm.Daily)
	}

	asset_a_returns := make([]float64, nSamples)
	for i := range nSamples {
		asset_a_returns[i] = allReturns[i][0]
	}

	// GetCorrelatedReturns(Daily) returns daily returns; sr.Mu and sr.Sigma are annualized
	dailyMu := stat.Mean(asset_a_returns, nil)
	dailySigma := stat.StdDev(asset_a_returns, nil)
	asset_a_mu := dailyMu * sm.Daily
	asset_a_sigma := dailySigma * math.Sqrt(sm.Daily)

	t.Logf("Asset 0 - Expected mean (annual): %.4f, Simulated (annual): %.4f", sr.Mu[0], asset_a_mu)
	t.Logf("Asset 0 - Expected std (annual): %.4f, Simulated (annual): %.4f", sr.Sigma[0], asset_a_sigma)

	// Allow tolerance for mean and std (Monte Carlo variation; mean has higher SE)
	if math.Abs(asset_a_mu-sr.Mu[0]) > 0.025 {
		t.Errorf("Mean differs too much: expected %.4f, got %.4f", sr.Mu[0], asset_a_mu)
	}
	if math.Abs(asset_a_sigma-sr.Sigma[0]) > 0.02 {
		t.Errorf("StdDev differs too much: expected %.4f, got %.4f", sr.Sigma[0], asset_a_sigma)
	}
}

func TestStatisticalResourcesWorkerCorrelatedReturnsForStudentT(t *testing.T) {
	nSamples := sm.Daily * 500
	returns := GenerateMockSeriesReturns(t, nSamples)

	settings_normal := sm.SimulationRequestSettings{DistType: sm.StandardNormal}
	sr_normal, _ := GetStatisticalResources(returns, settings_normal)
	worker_normal := NewWorkerResources(sr_normal, 42, 0)

	settings_student_t := sm.SimulationRequestSettings{DistType: sm.StudentT, DegreesOfFreedom: 5}
	sr_student_t, _ := GetStatisticalResources(returns, settings_student_t)
	worker_student_t := NewWorkerResources(sr_student_t, 42, 1)

	normalReturns := make([]float64, nSamples)
	tReturns := make([]float64, nSamples)
	for i := range nSamples {
		normalReturns[i] = worker_normal.GetCorrelatedReturns(sm.Daily)[0]
		tReturns[i] = worker_student_t.GetCorrelatedReturns(sm.Daily)[0]
	}

	// TODO: need to finish this at some point, but am going to work on the controller and front end to get some tangible results
}

// Helper: Generate mock series returns
func GenerateMockSeriesReturns(t *testing.T, n int) []*SeriesReturns {
	t.Helper()
	returns := generateMockReturns(t, n)
	res := make([]*SeriesReturns, len(returns))

	for i := range len(returns) {
		res[i] = &SeriesReturns{
			ScenarioConfigurationComponent: dm.ScenarioConfigurationComponent{
				AssetId: int32(i),
				Weight:  1 / float64(len(returns)),
			},
			Returns:             returns[i],
			Dates:               make([]time.Time, 0),
			AnnualizationFactor: sm.Daily,
		}
	}

	return res
}

// Helper: Generate one year of mock daily historical returns
func generateMockReturns(t *testing.T, n int) [][]float64 {
	t.Helper()

	nAssets := 3
	corrData := []float64{
		1.0, corr_ab, corr_ac,
		corr_ab, 1.0, corr_bc,
		corr_ac, corr_bc, 1.0,
	}

	corrMatrix := mat.NewSymDense(nAssets, corrData)
	var chol mat.Cholesky
	if ok := chol.Factorize(corrMatrix); !ok {
		t.Fatalf("Correlation matrix is not positive definite")
	}

	L := new(mat.TriDense)
	chol.LTo(L)

	src := rand.NewPCG(42, 0)
	normalDist := distuv.Normal{Mu: 0, Sigma: 1, Src: src}

	asset_a := make([]float64, n)
	asset_b := make([]float64, n)
	asset_c := make([]float64, n)

	z := make([]float64, nAssets)
	for sim := range n {
		for i := range nAssets {
			z[i] = normalDist.Rand()
		}

		zVec := mat.NewVecDense(nAssets, z)
		correlatedZ := mat.NewVecDense(nAssets, nil)
		correlatedZ.MulVec(L, zVec)

		asset_a[sim] = calculateLogNormalReturn(t, mu_a, sigma_a, correlatedZ.AtVec(0), sm.Daily)
		asset_b[sim] = calculateLogNormalReturn(t, mu_b, sigma_b, correlatedZ.AtVec(1), sm.Daily)
		asset_c[sim] = calculateLogNormalReturn(t, mu_c, sigma_c, correlatedZ.AtVec(2), sm.Daily)
	}

	return [][]float64{asset_a, asset_b, asset_c}
}

// Helper: Centralized way to calculate log normal returns
func calculateLogNormalReturn(t *testing.T, mu, sigma, rng, normalization float64) float64 {
	t.Helper()
	return (mu-0.5*math.Pow(sigma, 2))/normalization + (sigma * rng / math.Sqrt(normalization))
}

// Helper: Generate one year of mock daily historical prices
func generateMockStockPrices(t *testing.T, returns [][]float64) [][]float64 {
	t.Helper()

	nSamples := len(returns[0])
	asset_a := make([]float64, nSamples+1)
	asset_b := make([]float64, nSamples+1)
	asset_c := make([]float64, nSamples+1)

	asset_a[0] = 100 // initial value for asset a
	asset_b[0] = 50  // asset b
	asset_c[0] = 200 // and asset c

	for sim := range nSamples {
		asset_a[sim+1] = asset_a[sim] * math.Exp(returns[0][sim])
		asset_b[sim+1] = asset_b[sim] * math.Exp(returns[1][sim])
		asset_c[sim+1] = asset_c[sim] * math.Exp(returns[2][sim])
	}

	return [][]float64{asset_a, asset_b, asset_c}
}

// Helper: Calculated drift adjusted average returns (mu)
func calculateDriftAdjustedMu(t *testing.T, mu, sigma float64) float64 {
	t.Helper()
	return mu - 0.5*math.Pow(sigma, 2)
}
