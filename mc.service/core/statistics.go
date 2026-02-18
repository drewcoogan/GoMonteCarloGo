package core

import (
	"fmt"
	"math"
	"math/rand/v2"

	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/gonum/stat/distuv"

	ex "mc.data/extensions"
)

const (
	StandardNormal = iota
	StudentT
)

const ( // idk if we need these depending on how front end gets and sends options.
	Daily     = 252
	Weekly    = 52
	Monthly   = 12
	Quarterly = 4
	Yearly    = 1
)

type StatisticalResources struct {
	CovMatrix     *mat.SymDense // covariance matrix for std normal dist
	CorrMatrix    *mat.SymDense // correlation matrix for student t dist
	CholeskyL     *mat.TriDense // cholesky of covariance (std normal dist)
	CholeskyCorrL *mat.TriDense // cholesky of correlation (student t dist)
	AssetWeight   []float64
	Mu            []float64 // annualized
	Sigma         []float64 // annualized
	DistType      int
	Df            int
}

// Used for parallelization, will have shared materials to minimize memory usage
type WorkerResource struct {
	*StatisticalResources // embed read only shared data
	normalDist            distuv.Normal
	tDist                 distuv.StudentsT
}

// Called in the go routine and have seeds respectively set for each
func NewWorkerResources(shared *StatisticalResources, seed, iterable uint64) *WorkerResource {
	rng := new(rand.PCG)
	if seed != 0 {
		rng = rand.NewPCG(seed, iterable)
	}

	tDist := distuv.StudentsT{Mu: 0, Sigma: 1, Nu: float64(shared.Df), Src: rng}
	normalDist := distuv.Normal{Mu: 0, Sigma: 1, Src: rng}

	return &WorkerResource{
		StatisticalResources: shared,
		tDist:                tDist,
		normalDist:           normalDist,
	}
}

func GetStatisticalResources(settings SimulationSettings, seriesReturns []*SeriesReturns) (*StatisticalResources, error) {
	var err error

	sr := &StatisticalResources{
		DistType: settings.DistType,
		Df:       settings.DegreesOfFreedom,
	}

	returns := make([][]float64, len(seriesReturns))
	for i, r := range seriesReturns {
		returns[i] = r.Returns
	}

	sr.CovMatrix = GetCovarianceMatrix(returns)
	sr.CholeskyL, err = GetCholeskyDecomposition(sr.CovMatrix)
	if err != nil {
		return nil, err
	}

	sr.AssetWeight = make([]float64, len(returns))
	sr.Mu = make([]float64, len(returns))
	sr.Sigma = make([]float64, len(returns))
	for i, r := range seriesReturns {
		sr.AssetWeight[i] = r.Weight
		sr.Mu[i] = stat.Mean(r.Returns, nil) * float64(r.AnnualizationFactor)
		sr.Sigma[i] = stat.StdDev(r.Returns, nil) * math.Sqrt(float64(r.AnnualizationFactor))
	}

	// Correlation Cholesky: used for StandardNormal (correlated N(0,1) then scale by sigma)
	// and for StudentT (Gaussian copula). Covariance Cholesky is in daily units so would
	// double-scale if used in CalculateLogNormalReturn.
	sr.CorrMatrix = GetCorrelationMatrix(sr.CovMatrix)
	sr.CholeskyCorrL, err = GetCholeskyDecomposition(sr.CorrMatrix)
	if err != nil {
		return nil, fmt.Errorf("failed to compute correlation Cholesky: %w", err)
	}

	if settings.DistType != StudentT {
		sr.CorrMatrix = nil // leave nil for StandardNormal for API clarity
	}

	return sr, nil
}

// GetCorrelatedReturns generates one set of correlated returns
// This is goroutine-safe as long as each goroutine has its own WorkerResources
func (wr *WorkerResource) GetCorrelatedReturns(simulationUnitOfTime int) []float64 {
	switch wr.DistType {
	case StandardNormal:
		return wr.generateNormalReturns(simulationUnitOfTime)
	case StudentT:
		return wr.generateTReturns(simulationUnitOfTime)
	default:
		return nil
	}
}

// generateNormalReturns generates a single period of correlated normal returns.
// uses Cholesky of correlation matrix so variates are standard normal; then scaled by sigma in CalculateLogNormalReturn.
func (wr *WorkerResource) generateNormalReturns(simulationUnitOfTime int) []float64 {
	n := len(wr.Mu)

	// TODO: is this right? do we use the chol corr l for standard normal?
	correlatedZ := wr.generateCorrelatedRandomVector(n)

	correlatedReturns := make([]float64, n)
	for i := range n {
		correlatedReturns[i] = CalculateLogNormalReturn(wr.Mu[i], wr.Sigma[i], correlatedZ.AtVec(i), simulationUnitOfTime)
	}

	return correlatedReturns
}

// generateTReturns generates correlated Student's t returns using Gaussian copula
func (wr *WorkerResource) generateTReturns(simulationUnitOfTime int) []float64 {
	n := len(wr.Mu)
	correlatedZ := wr.generateCorrelatedRandomVector(n)

	// gaussian copula transformation
	// https://colab.research.google.com/github/tensorflow/probability/blob/main/tensorflow_probability/examples/jupyter_notebooks/Gaussian_Copula.ipynb#scrollTo=1kSHqIp0GaRh
	correlatedReturns := make([]float64, n)
	for i := range n {
		u := wr.normalDist.CDF(correlatedZ.AtVec(i)) // transform to uniform [0,1]
		tValue := wr.tDist.Quantile(u)               // transform to t-distributed
		correlatedReturns[i] = CalculateLogNormalReturn(wr.Mu[i], wr.Sigma[i], tValue, simulationUnitOfTime)
	}

	return correlatedReturns
}

func (wr *WorkerResource) generateCorrelatedRandomVector(n int) *mat.VecDense {
	z := make([]float64, n)
	for i := range n {
		z[i] = wr.normalDist.Rand()
	}

	// L can be either correlated or covariance depending on the distribution
	zVec := mat.NewVecDense(n, z)
	correlatedZ := mat.NewVecDense(n, nil)
	correlatedZ.MulVec(wr.CholeskyCorrL, zVec) // correlated z = chol L * rng variables
	return correlatedZ
}

func CalculateLogNormalReturn(mu, sigma, rng float64, annualizationFactor int) float64 {
	return (mu-0.5*math.Pow(sigma, 2))/float64(annualizationFactor) + (sigma * rng / math.Sqrt(float64(annualizationFactor)))
}

func GetCovarianceMatrix[T ex.Number](data [][]T) *mat.SymDense {
	returnMatrix := ArrToMatrix(data)
	covMatrix := mat.NewSymDense(len(data), nil)
	stat.CovarianceMatrix(covMatrix, returnMatrix, nil)
	return covMatrix
}

// GetCorrelationMatrix builds a correlation matrix from a covariance matrix so diagonal is 1.
// Use this when covMatrix is in the same units (e.g. daily); corr_ij = cov_ij / sqrt(cov_ii*cov_jj).
func GetCorrelationMatrix(covMatrix *mat.SymDense) *mat.SymDense {
	n := covMatrix.SymmetricDim()
	corrMatrix := mat.NewSymDense(n, nil)

	for i := range n {
		for j := range i + 1 {
			corr := covMatrix.At(i, j) / math.Sqrt(covMatrix.At(i, i)*covMatrix.At(j, j))
			corrMatrix.SetSym(i, j, corr)
		}
	}

	return corrMatrix
}

func GetCholeskyDecomposition(covMatrix *mat.SymDense) (*mat.TriDense, error) {
	chol := new(mat.Cholesky)
	if ok := chol.Factorize(covMatrix); !ok {
		return nil, fmt.Errorf("covariance matrix is not positive definite")
	}

	L := new(mat.TriDense)
	chol.LTo(L)

	return L, nil
}

func ArrToMatrix[T ex.Number](data [][]T) *mat.Dense {
	nSymbols := len(data)
	nObservations := len(data[0])
	res := mat.NewDense(nObservations, nSymbols, nil)
	for j, col := range data {
		for i, row := range col {
			res.Set(i, j, float64(row))
		}
	}
	return res
}

func convertFrequencyToString(inp int) string {
	switch inp {
	case Daily:
		return "days"
	case Weekly:
		return "weeks"
	case Monthly:
		return "months"
	case Quarterly:
		return "quarters"
	case Yearly:
		return "years"
	default:
		panic(fmt.Sprintf("%v is not a recognized simulation frequency", inp))
	}
}
