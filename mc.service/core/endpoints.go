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
	sm "mc.service/models"
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
func jsonResponse[T any](w http.ResponseWriter, statusCode int, data T) {
	payload := sm.GetServiceResponseOk(&data)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(payload)
}

// jsonError writes a JSON error response
func jsonError(w http.ResponseWriter, statusCode int, message string) {
	payload := sm.GetServiceResponseError(message)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(payload)
}

func GetHttpServer(sc ServiceContext) *http.Server {
	r := chi.NewRouter()

	// heartbeat
	r.Get("/api/heartbeat", func(w http.ResponseWriter, r *http.Request) { heartbeat(w, sc) })

	// stock data, syncing, and availability
	r.Route("/api/assets", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) { getAssets(w, sc) })
		r.Post("/sync", func(w http.ResponseWriter, r *http.Request) { syncAsset(w, r, sc) })
	})

	// scenarios, creation, retrieval, updating, and deletion
	r.Route("/api/scenarios", func(r chi.Router) {
		r.Get("/", func(w http.ResponseWriter, r *http.Request) { getScenarios(w, sc) })
		r.Post("/", func(w http.ResponseWriter, r *http.Request) { createScenario(w, r, sc) })
		r.Get("/{id}", func(w http.ResponseWriter, r *http.Request) { getScenario(w, r, sc) })
		r.Put("/{id}", func(w http.ResponseWriter, r *http.Request) { updateScenario(w, r, sc) })
		r.Delete("/{id}", func(w http.ResponseWriter, r *http.Request) { deleteScenario(w, r, sc) })
	})

	// simulation, resource retrieval, and running
	r.Route("/api/simulation", func(r chi.Router) {
		r.Get("/resources", func(w http.ResponseWriter, r *http.Request) { getSimulationResources(w, r, sc) })
		r.Post("/run/{id}", func(w http.ResponseWriter, r *http.Request) { runSimulation(w, r, sc) })
		r.Get("/run-history/{id}", func(w http.ResponseWriter, r *http.Request) { getSimulationRunHistory(w, r, sc) })
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

// GET /api/heartbeat
func heartbeat(w http.ResponseWriter, sc ServiceContext) {
	postgresPing := sc.PostgresConnection.Ping(sc.Context)

	res := map[string]bool{
		"service":  true,
		"database": postgresPing == nil,
	}

	jsonResponse(w, http.StatusOK, res)
}

// GET /api/getAssets
func getAssets(w http.ResponseWriter, sc ServiceContext) {
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

// POST /api/assets/sync
func syncAsset(w http.ResponseWriter, r *http.Request, sc ServiceContext) {
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
		// TODO: what is the error here and what date this this?
		jsonError(w, http.StatusBadRequest, err.Error())
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

	res := ex.FmtShort(md.LastRefreshed)
	jsonResponse(w, http.StatusOK, res)
}

// GET /api/scenarios
func getScenarios(w http.ResponseWriter, sc ServiceContext) {
	scenarios, err := sc.PostgresConnection.GetScenarios(sc.Context)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, fmt.Sprintf("error getting scenarios: %v", err))
		return
	}

	res := make([]sm.ScenarioResponse, len(scenarios))
	for i, scenario := range scenarios {
		res[i] = sm.MapScenarioToResponse(scenario)
	}

	jsonResponse(w, http.StatusOK, res)
}

// POST /api/scenarios
func createScenario(w http.ResponseWriter, r *http.Request, sc ServiceContext) {
	var req sm.ScenarioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	created, status, err := sc.InsertNewScenario(req)
	if err != nil {
		jsonError(w, status, err.Error())
		return
	}

	res := sm.MapScenarioToResponse(created)
	jsonResponse(w, status, res)
}

// GET /api/scenarios/{id}
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

	res := sm.MapScenarioToResponse(scenario)
	jsonResponse(w, http.StatusOK, res)
}

// PUT /api/scenarios/{id}
func updateScenario(w http.ResponseWriter, r *http.Request, sc ServiceContext) {
	scenarioID, err := scenarioIDFromRequest(r)
	if err != nil {
		jsonError(w, http.StatusNotFound, "scenario not found")
		return
	}

	var req sm.ScenarioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	updated, status, err := sc.UpdateScenario(scenarioID, req)
	if err != nil {
		jsonError(w, status, err.Error())
		return
	}

	res := sm.MapScenarioToResponse(updated)
	jsonResponse(w, http.StatusOK, res)
}

// DELETE /api/scenarios/{id}
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

	jsonResponse(w, http.StatusOK, true)
}

// GET /api/simulation/run-history
func getSimulationRunHistory(w http.ResponseWriter, r *http.Request, sc ServiceContext) {
	scenarioID, err := scenarioIDFromRequest(r)
	if err != nil {
		jsonError(w, http.StatusNotFound, "scenario not found")
		return
	}

	history, err := sc.PostgresConnection.GetScenarioRunHistories(sc.Context, scenarioID, 10)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, fmt.Sprintf("error getting scenario run history: %v", err))
		return
	}

	jsonResponse(w, http.StatusOK, history)
}

// GET /api/simulation/resources
func getSimulationResources(w http.ResponseWriter, _ *http.Request, _ ServiceContext) {
	resources := sm.GetSimulationSettingsResources()
	jsonResponse(w, http.StatusOK, resources)
}

// POST /api/simulation/run/{id}
func runSimulation(w http.ResponseWriter, r *http.Request, sc ServiceContext) {
	scenarioID, err := scenarioIDFromRequest(r)
	if err != nil {
		jsonError(w, http.StatusNotFound, "scenario not found")
		return
	}

	var req sm.SimulationRequestSettings
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		jsonError(w, http.StatusBadRequest, err.Error())
		return
	}

	res, err := sc.RunSimulation(scenarioID, req)
	if err != nil {
		jsonError(w, http.StatusInternalServerError, fmt.Sprintf("error running scenario: %v", err))
		return
	}

	jsonResponse(w, http.StatusOK, res)
}

// scenarioIDFromRequest reads and parses the {id} URL param from a Chi route.
func scenarioIDFromRequest(r *http.Request) (int32, error) {
	trimmed := strings.Trim(chi.URLParam(r, "id"), "/")
	if trimmed == "" {
		return 0, fmt.Errorf("scenario id is required")
	}

	id, err := strconv.ParseInt(trimmed, 10, 32)
	if err != nil || id <= 0 {
		return 0, fmt.Errorf("scenario id is invalid")
	}

	return int32(id), nil
}
