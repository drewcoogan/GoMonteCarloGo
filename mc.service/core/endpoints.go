package core

import (
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"strconv"
	"strings"
	"time"

	ex "mc.data/extensions"
	m "mc.data/models"
)

const (
	DefaultAddr = ":8080"
)

func getHandler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Accept, Authorization")
		w.Header().Set("Access-Control-Expose-Headers", "Content-Length")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "43200") // 12 hours in seconds

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

// jsonResponse writes a JSON response with the given status code and data
func jsonResponse(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(data)
}

// jsonError writes a JSON error response
func jsonError(w http.ResponseWriter, statusCode int, message string) {
	jsonResponse(w, statusCode, map[string]string{"error": message})
}

func GetHttpServer(sc ServiceContext) *http.Server {
	mux := http.NewServeMux()

	// heartbeat route
	mux.HandleFunc("/api/ping", ping)

	// core functionality routes
	mux.HandleFunc("/api/syncStockData", func(w http.ResponseWriter, r *http.Request) {
		syncStockData(w, r, sc)
	})
	mux.HandleFunc("/api/assets", func(w http.ResponseWriter, r *http.Request) {
		listAssets(w, r, sc)
	})
	mux.HandleFunc("/api/scenarios", func(w http.ResponseWriter, r *http.Request) {
		scenariosHandler(w, r, sc)
	})
	mux.HandleFunc("/api/scenarios/", func(w http.ResponseWriter, r *http.Request) {
		scenariosHandler(w, r, sc)
	})

	// basic testing routes, will remove eventually
	mux.HandleFunc("/api/test/addByGet", addByGet)
	mux.HandleFunc("/api/test/addByPost", addByPost)

	handler := getHandler(mux)

	return &http.Server{
		Addr:           DefaultAddr,
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}

// heartbeat routes
func ping(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}
	jsonResponse(w, http.StatusOK, map[string]string{"message": "pong"})
}

// core functionalty routes
func GetConfigurationResources(w http.ResponseWriter, r *http.Request, sc ServiceContext) {
	if r.Method != http.MethodGet {
		jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}

}

type SyncStockDataRequest struct {
	Symbol string `json:"symbol"`
}

func syncStockData(w http.ResponseWriter, r *http.Request, sc ServiceContext) {
	if r.Method != http.MethodPost {
		jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req SyncStockDataRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	if req.Symbol == "" {
		jsonError(w, http.StatusBadRequest, "symbol is required")
		return
	}

	lut, err := sc.SyncSymbolTimeSeriesData(req.Symbol)
	if err != nil {
		if lut.IsZero() {
			jsonError(w, http.StatusBadRequest, err.Error())
			return
		}
		jsonResponse(w, http.StatusBadRequest, map[string]any{
			"date":    ex.FmtShort(lut),
			"message": err.Error(),
		})
		return
	}

	// Get the updated metadata to return the last refreshed date
	md, err := sc.PostgresConnection.GetMetaDataBySymbol(sc.Context, req.Symbol)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, fmt.Sprintf("error getting metadata: %v", err))
		return
	}

	if md == nil {
		jsonError(w, http.StatusInternalServerError, "metadata not found after sync")
		return
	}

	jsonResponse(w, http.StatusOK, map[string]string{"date": ex.FmtShort(md.LastRefreshed)})
}

type AssetSummary struct {
	Id            int32     `json:"id"`
	Symbol        string    `json:"symbol"`
	LastRefreshed time.Time `json:"lastRefreshed"`
}

func listAssets(w http.ResponseWriter, r *http.Request, sc ServiceContext) {
	if r.Method != http.MethodGet {
		jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	assets, err := sc.PostgresConnection.GetAllMetaData(sc.Context)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, fmt.Sprintf("error getting assets: %v", err))
		return
	}

	res := make([]AssetSummary, 0, len(assets))
	for _, asset := range assets {
		res = append(res, AssetSummary{
			Id:            asset.Id,
			Symbol:        asset.Symbol,
			LastRefreshed: asset.LastRefreshed,
		})
	}

	jsonResponse(w, http.StatusOK, res)
}

type ScenarioComponentPayload struct {
	AssetId int32   `json:"assetId"`
	Weight  float64 `json:"weight"`
}

type ScenarioRequest struct {
	Name          string                     `json:"name"`
	FloatedWeight bool                       `json:"floatedWeight"`
	Components    []ScenarioComponentPayload `json:"components"`
}

type ScenarioResponse struct {
	Id            int32                      `json:"id"`
	Name          string                     `json:"name"`
	FloatedWeight bool                       `json:"floatedWeight"`
	CreatedAt     time.Time                  `json:"createdAt"`
	UpdatedAt     time.Time                  `json:"updatedAt"`
	Components    []ScenarioComponentPayload `json:"components"`
}

func scenariosHandler(w http.ResponseWriter, r *http.Request, sc ServiceContext) {
	basePath := "/api/scenarios"
	path := strings.TrimPrefix(r.URL.Path, basePath)

	if path == "" || path == "/" {
		handleScenarioCollection(w, r, sc)
		return
	}

	scenarioID, err := parseScenarioID(path)
	if err != nil {
		jsonError(w, http.StatusNotFound, "scenario not found")
		return
	}

	handleScenarioItem(w, r, sc, scenarioID)
}

func handleScenarioCollection(w http.ResponseWriter, r *http.Request, sc ServiceContext) {
	switch r.Method {
	case http.MethodGet:
		scenarios, err := sc.PostgresConnection.GetScenarios(sc.Context)
		if err != nil {
			jsonError(w, http.StatusInternalServerError, fmt.Sprintf("error getting scenarios: %v", err))
			return
		}

		res := make([]ScenarioResponse, 0, len(scenarios))
		for _, scenario := range scenarios {
			res = append(res, toScenarioResponse(scenario))
		}

		jsonResponse(w, http.StatusOK, res)
	case http.MethodPost:
		var req ScenarioRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := validateScenarioRequest(req); err != nil {
			jsonError(w, http.StatusBadRequest, err.Error())
			return
		}

		newScenario := mapScenarioRequest(req)
		created, err := sc.PostgresConnection.InsertNewScenario(sc.Context, newScenario, nil)
		if err != nil {
			jsonError(w, http.StatusInternalServerError, fmt.Sprintf("error creating scenario: %v", err))
			return
		}

		jsonResponse(w, http.StatusCreated, toScenarioResponse(created))
	default:
		jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func handleScenarioItem(w http.ResponseWriter, r *http.Request, sc ServiceContext, scenarioID int32) {
	switch r.Method {
	case http.MethodGet:
		scenario, err := sc.PostgresConnection.GetScenarioByID(sc.Context, scenarioID)
		if err != nil {
			jsonError(w, http.StatusInternalServerError, fmt.Sprintf("error getting scenario: %v", err))
			return
		}
		if scenario == nil {
			jsonError(w, http.StatusNotFound, "scenario not found")
			return
		}

		jsonResponse(w, http.StatusOK, toScenarioResponse(scenario))
	case http.MethodPut:
		var req ScenarioRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			jsonError(w, http.StatusBadRequest, err.Error())
			return
		}

		if err := validateScenarioRequest(req); err != nil {
			jsonError(w, http.StatusBadRequest, err.Error())
			return
		}

		updateScenario := mapScenarioRequest(req)
		updated, err := sc.PostgresConnection.UpdateExistingScenario(sc.Context, scenarioID, updateScenario, nil)
		if err != nil {
			if strings.Contains(err.Error(), "not found") {
				jsonError(w, http.StatusNotFound, "scenario not found")
				return
			}
			jsonError(w, http.StatusInternalServerError, fmt.Sprintf("error updating scenario: %v", err))
			return
		}

		jsonResponse(w, http.StatusOK, toScenarioResponse(updated))
	case http.MethodDelete:
		if err := sc.PostgresConnection.DeleteScenario(sc.Context, scenarioID, nil); err != nil {
			if strings.Contains(err.Error(), "not found") {
				jsonError(w, http.StatusNotFound, "scenario not found")
				return
			}
			jsonError(w, http.StatusInternalServerError, fmt.Sprintf("error deleting scenario: %v", err))
			return
		}

		w.WriteHeader(http.StatusNoContent)
	default:
		jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func parseScenarioID(path string) (int32, error) {
	trimmed := strings.Trim(path, "/")
	if trimmed == "" || strings.Contains(trimmed, "/") {
		return 0, fmt.Errorf("invalid scenario id")
	}

	id, err := strconv.ParseInt(trimmed, 10, 32)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid scenario id")
	}

	return int32(id), nil
}

func validateScenarioRequest(req ScenarioRequest) error {
	if strings.TrimSpace(req.Name) == "" {
		return fmt.Errorf("name is required")
	}
	if len(req.Components) == 0 {
		return fmt.Errorf("at least one component is required")
	}

	seen := make(map[int32]bool, len(req.Components))
	weightSum := 0.0
	for _, component := range req.Components {
		if component.AssetId == 0 {
			return fmt.Errorf("assetId must be provided")
		}
		if component.Weight <= 0 {
			return fmt.Errorf("component weights must be positive")
		}
		if seen[component.AssetId] {
			return fmt.Errorf("duplicate assetId %d", component.AssetId)
		}
		seen[component.AssetId] = true
		weightSum += component.Weight
	}

	if math.Abs(weightSum-1.0) > 0.001 {
		return fmt.Errorf("component weights must sum to 1.0, got %.4f", weightSum)
	}

	return nil
}

func mapScenarioRequest(req ScenarioRequest) m.NewScenario {
	components := make([]m.NewComponent, len(req.Components))
	for i, c := range req.Components {
		components[i] = m.NewComponent{
			AssetId: c.AssetId,
			Weight:  c.Weight,
		}
	}

	return m.NewScenario{
		Name:          req.Name,
		FloatedWeight: req.FloatedWeight,
		Components:    components,
	}
}

func toScenarioResponse(scenario *m.Scenario) ScenarioResponse {
	res := ScenarioResponse{
		Id:            scenario.Id,
		Name:          scenario.Name,
		FloatedWeight: scenario.FloatedWeight,
		CreatedAt:     scenario.CreatedAt,
		UpdatedAt:     scenario.UpdatedAt,
		Components:    make([]ScenarioComponentPayload, 0, len(scenario.Components)),
	}

	for _, c := range scenario.Components {
		res.Components = append(res.Components, ScenarioComponentPayload{
			AssetId: c.AssetId,
			Weight:  c.Weight,
		})
	}

	return res
}

// Testing endpoints below to ensure functionality
type NumbersToSum struct {
	Number1 int `json:"number1"`
	Number2 int `json:"number2"`
}

// AddByGet adds two numbers via a GET request
func addByGet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	number1Str := r.URL.Query().Get("number1")
	number2Str := r.URL.Query().Get("number2")

	number1, err1 := strconv.Atoi(number1Str)
	number2, err2 := strconv.Atoi(number2Str)

	if err1 != nil || err2 != nil {
		jsonError(w, http.StatusBadRequest, "Invalid numbers")
		return
	}

	result := number1 + number2
	jsonResponse(w, http.StatusOK, map[string]int{"result": result})
}

// AddByPost adds two numbers via a POST request
func addByPost(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		jsonError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var nums NumbersToSum
	if err := json.NewDecoder(r.Body).Decode(&nums); err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}
	result := nums.Number1 + nums.Number2
	jsonResponse(w, http.StatusOK, map[string]int{"result": result})
}
