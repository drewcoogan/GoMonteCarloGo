package core

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
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
	r := chi.NewRouter()

	// heartbeat
	r.Get("/api/ping", http.HandlerFunc(ping))

	// stock data, syncing and availability
	r.Post("/api/syncStockData", func(w http.ResponseWriter, r *http.Request) { syncStockData(w, r, sc) })
	r.Get("/api/assets", func(w http.ResponseWriter, r *http.Request) { listAssets(w, sc) })

	// scenarios, collection and item management
	r.Route("/api/scenarios", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) { listScenarios(w, sc) })
		r.Post("/", func(w http.ResponseWriter, r *http.Request) { createScenario(w, r, sc) })
		r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) { getScenario(w, r, sc) })
		r.Put("/{id}", func(w http.ResponseWriter, r *http.Request) { updateScenario(w, r, sc) })
		r.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) { deleteScenario(w, r, sc) })
		r.Post("/run/{id}", func(w http.ResponseWriter, r *http.Request) { runScenario(w, r, sc) })
	})

	handler := getHandler(r)

	return &http.Server{
		Addr:           DefaultAddr,
		Handler:        handler,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		MaxHeaderBytes: 1 << 20,
	}
}

// heartbeat
func ping(w http.ResponseWriter, r *http.Request) {
	jsonResponse(w, http.StatusOK, map[string]string{
		"message": "pong",
	})
}

// sync stock data
func syncStockData(w http.ResponseWriter, r *http.Request, sc ServiceContext) {
	var req struct {
		Symbol string `json:"symbol"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	if strings.TrimSpace(req.Symbol) == "" {
		jsonError(w, http.StatusBadRequest, "symbol is required")
		return
	}

	lastUpdateTime, err := sc.SyncSymbolTimeSeriesData(req.Symbol)
	if err != nil {
		if lastUpdateTime.IsZero() {
			jsonError(w, http.StatusBadRequest, err.Error())
			return
		}
		jsonResponse(w, http.StatusBadRequest, map[string]any{
			"date":    ex.FmtShort(lastUpdateTime),
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

	jsonResponse(w, http.StatusOK, map[string]string{
		"date": ex.FmtShort(md.LastRefreshed),
	})
}

func listAssets(w http.ResponseWriter, sc ServiceContext) {
	assets, err := sc.PostgresConnection.GetAllMetaData(sc.Context)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, fmt.Sprintf("error getting assets: %v", err))
		return
	}

	type AssetSummary struct {
		Id            int32     `json:"id"`
		Symbol        string    `json:"symbol"`
		LastRefreshed time.Time `json:"lastRefreshed"`
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

// scenarios, collection and item management
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

// listScenarios handles GET /api/scenarios
func listScenarios(w http.ResponseWriter, sc ServiceContext) {
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
}

// createScenario handles POST /api/scenarios
func createScenario(w http.ResponseWriter, r *http.Request, sc ServiceContext) {
	var req ScenarioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	created, status, err := sc.InsertNewScenario(req)
	if err != nil {
		jsonError(w, status, err.Error())
		return
	}

	jsonResponse(w, status, toScenarioResponse(created))
}

// getScenario handles GET /api/scenarios/{id}
func getScenario(w http.ResponseWriter, r *http.Request, sc ServiceContext) {
	scenarioID, err := scenarioIDFromRequest(r)
	if err != nil {
		jsonError(w, http.StatusNotFound, "scenario not found")
		return
	}

	scenario, err := sc.PostgresConnection.GetScenarioByID(sc.Context, scenarioID)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, fmt.Sprintf("error getting scenario: %v", err))
		return
	}

	jsonResponse(w, http.StatusOK, toScenarioResponse(scenario))
}

// updateScenario handles PUT /api/scenarios/{id}
func updateScenario(w http.ResponseWriter, r *http.Request, sc ServiceContext) {
	scenarioID, err := scenarioIDFromRequest(r)
	if err != nil {
		jsonError(w, http.StatusNotFound, "scenario not found")
		return
	}

	var req ScenarioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	updated, status, err := sc.UpdateScenario(scenarioID, req)
	if err != nil {
		jsonError(w, status, err.Error())
		return
	}
	
	jsonResponse(w, http.StatusOK, toScenarioResponse(updated))
}

// deleteScenario handles DELETE /api/scenarios/{id}
func deleteScenario(w http.ResponseWriter, r *http.Request, sc ServiceContext) {
	scenarioID, err := scenarioIDFromRequest(r)
	if err != nil {
		jsonError(w, http.StatusNotFound, "scenario not found")
		return
	}

	if err := sc.PostgresConnection.DeleteScenario(sc.Context, scenarioID); err != nil {
		if strings.Contains(err.Error(), "not found") {
			jsonError(w, http.StatusNotFound, "scenario not found")
			return
		}
		jsonError(w, http.StatusInternalServerError, fmt.Sprintf("error deleting scenario: %v", err))
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// runScenario handles POST /api/scenarios/run/{id}
func runScenario(w http.ResponseWriter, r *http.Request, _ ServiceContext) {
	scenarioID, err := scenarioIDFromRequest(r)
	if err != nil {
		jsonError(w, http.StatusNotFound, "scenario not found")
		return
	}

	_ = scenarioID
}

// scenarioIDFromRequest reads and parses the {id} URL param from a Chi route.
func scenarioIDFromRequest(r *http.Request) (int32, error) {
	trimmed := strings.Trim(chi.URLParam(r, "id"), "/")
	if trimmed == "" || strings.Contains(trimmed, "/") {
		return 0, fmt.Errorf("invalid scenario id")
	}

	id, err := strconv.ParseInt(trimmed, 10, 32)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("invalid scenario id")
	}

	return int32(id), nil
}

func mapScenarioRequest(req ScenarioRequest) m.Scenario {
	components := make([]m.ScenarioConfigurationComponent, len(req.Components))
	for i, c := range req.Components {
		components[i] = m.ScenarioConfigurationComponent {
			AssetId: c.AssetId,
			Weight:  c.Weight,
		}
	}

	return m.Scenario{
		ScenarioConfiguration: m.ScenarioConfiguration {
			Name:          req.Name,
			FloatedWeight: req.FloatedWeight,
		},
		Components: components,
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