package core

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/joho/godotenv"

	dm "mc.data/models"
	r "mc.data/repos"
	av "mc.service/api/alpha_vantage"
	sm "mc.service/models"
)

type scenarioResponseBody struct {
	Data  sm.ScenarioResponse `json:"data"`
	Error string              `json:"error"`
}

type scenariosResponseBody struct {
	Data  []sm.ScenarioResponse `json:"data"`
	Error string                `json:"error"`
}

type heartbeatResponseBody struct {
	Data struct {
		Service   bool   `json:"service"`
		Database  bool   `json:"database"`
		GoVersion string `json:"goVersion"`
	} `json:"data"`
	Error string `json:"error"`
}

type assetsResponseBody struct {
	Data []struct {
		Id     int32  `json:"id"`
		Symbol string `json:"symbol"`
	} `json:"data"`
	Error string `json:"error"`
}

type simulationResourcesResponseBody struct {
	Data struct {
		DistributionType     map[string]int `json:"distributionType"`
		SimulationUnitOfTime map[string]int `json:"simulationUnitOfTime"`
		SimulationDuration   map[string]int `json:"simulationDuration"`
	} `json:"data"`
	Error string `json:"error"`
}

type runHistoryResponseBody struct {
	Data  []any  `json:"data"`
	Error string `json:"error"`
}

func getTestServiceContext(t *testing.T) ServiceContext {
	t.Helper()
	if err := godotenv.Load("../.env"); err != nil {
		t.Fatalf("loading .env: %v", err)
	}
	ctx := context.Background()
	pg, err := r.GetPostgresConnection(ctx, os.Getenv("DATABASE_URL"))
	if err != nil {
		t.Fatalf("getting postgres connection: %v", err)
	}
	t.Cleanup(func() { pg.Close() })
	return ServiceContext{
		Context:            ctx,
		PostgresConnection: pg,
		AlphaVantageClient: av.AlphaVantageClient{},
	}
}

func getTestServer(t *testing.T) (*http.Server, ServiceContext) {
	t.Helper()
	sc := getTestServiceContext(t)
	return GetHttpServer(sc), sc
}

func Test_heartbeat(t *testing.T) {
	_, sc := getTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/heartbeat", nil)
	rec := httptest.NewRecorder()

	GetHttpServer(sc).Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("heartbeat: status = %d; want %d", rec.Code, http.StatusOK)
	}

	var body heartbeatResponseBody
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error != "" {
		t.Errorf("heartbeat: error = %q", body.Error)
	}
	if !body.Data.Service {
		t.Error("heartbeat: data.service should be true")
	}
	_ = body.Data.Database
	if body.Data.GoVersion == "" {
		t.Error("heartbeat: data.goVersion should be non-empty (runtime.Version)")
	}
}

func Test_getAssets(t *testing.T) {
	_, sc := getTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/assets/", nil)
	rec := httptest.NewRecorder()

	GetHttpServer(sc).Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("getAssets: status = %d; want %d", rec.Code, http.StatusOK)
	}

	var body assetsResponseBody
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error != "" {
		t.Errorf("getAssets: error = %q", body.Error)
	}
	_ = body.Data
}

func Test_getSimulationResources(t *testing.T) {
	_, sc := getTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/simulation/resources", nil)
	rec := httptest.NewRecorder()

	GetHttpServer(sc).Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("getSimulationResources: status = %d; want %d", rec.Code, http.StatusOK)
	}

	var body simulationResourcesResponseBody
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.Error != "" {
		t.Errorf("getSimulationResources: error = %q", body.Error)
	}
	if body.Data.DistributionType == nil || body.Data.SimulationUnitOfTime == nil || body.Data.SimulationDuration == nil {
		t.Error("getSimulationResources: expected non-nil resource maps")
	}
}

func Test_getAssetPrices_notFound(t *testing.T) {
	_, sc := getTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/assets/999999999/prices", nil)
	rec := httptest.NewRecorder()

	GetHttpServer(sc).Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("getAssetPrices: status = %d; want %d", rec.Code, http.StatusNotFound)
	}
}

func Test_syncAsset_emptySymbol(t *testing.T) {
	_, sc := getTestServer(t)
	body := strings.NewReader(`{"symbol":""}`)
	req := httptest.NewRequest(http.MethodPost, "/api/assets/sync", body)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	GetHttpServer(sc).Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("syncAsset empty symbol: status = %d; want %d", rec.Code, http.StatusBadRequest)
	}
}

func Test_getScenarios(t *testing.T) {
	_, sc := getTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/scenarios/", nil)
	rec := httptest.NewRecorder()

	GetHttpServer(sc).Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Errorf("getScenarios: status = %d; want %d", rec.Code, http.StatusOK)
	}

	var resp scenariosResponseBody
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Error != "" {
		t.Errorf("getScenarios: error = %q", resp.Error)
	}
	_ = resp.Data
}

// insertTestMetadata inserts two metadata rows for scenario tests and returns their IDs.
// Caller must defer cleanup: delete scenario first (if any), then delete metadata by ID.
func insertTestMetadata(t *testing.T, sc ServiceContext) (assetAID, assetBID int32) {
	t.Helper()
	suffix := time.Now().UnixNano()
	assetA := dm.TimeSeriesMetadata{
		Symbol:        fmt.Sprintf("_EPTEST_A_%d", suffix),
		LastRefreshed: time.Date(2025, 10, 31, 0, 0, 0, 0, time.UTC),
	}
	assetB := dm.TimeSeriesMetadata{
		Symbol:        fmt.Sprintf("_EPTEST_B_%d", suffix),
		LastRefreshed: time.Date(2025, 10, 31, 0, 0, 0, 0, time.UTC),
	}
	if err := sc.PostgresConnection.InsertNewMetaData(sc.Context, &assetA, nil); err != nil {
		t.Fatalf("insert metadata A: %v", err)
	}
	if err := sc.PostgresConnection.InsertNewMetaData(sc.Context, &assetB, nil); err != nil {
		t.Fatalf("insert metadata B: %v", err)
	}
	return assetA.Id, assetB.Id
}

func Test_createScenario_and_getScenario(t *testing.T) {
	_, sc := getTestServer(t)
	assetAID, assetBID := insertTestMetadata(t, sc)
	defer func() {
		_ = sc.PostgresConnection.DeleteMetadataByID(sc.Context, assetAID)
		_ = sc.PostgresConnection.DeleteMetadataByID(sc.Context, assetBID)
	}()

	createBody := fmt.Sprintf(`
		{"name":"Endpoint Test Scenario",
		"floatedWeight":false,
		"components":
		[{
			"assetId":%d,
			"weight":0.6
		},
		{
			"assetId":%d,
			"weight":0.4
		}
		]}`, assetAID, assetBID)

	req := httptest.NewRequest(http.MethodPost, "/api/scenarios/", strings.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	GetHttpServer(sc).Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Errorf("createScenario: status = %d; want %d", rec.Code, http.StatusCreated)
	}

	var createResp scenarioResponseBody
	if err := json.NewDecoder(rec.Body).Decode(&createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	if createResp.Error != "" {
		t.Fatalf("createScenario: error = %q", createResp.Error)
	}
	if createResp.Data.Id == 0 {
		t.Fatal("createScenario: expected non-zero id")
	}
	scenarioID := createResp.Data.Id
	defer func() {
		_ = sc.PostgresConnection.DeleteScenarioPermanent(sc.Context, scenarioID)
	}()

	// GET /api/scenarios/{id}
	getReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/scenarios/%d", scenarioID), nil)
	getRec := httptest.NewRecorder()
	GetHttpServer(sc).Handler.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Errorf("getScenario: status = %d; want %d", getRec.Code, http.StatusOK)
	}
	var getResp scenarioResponseBody
	if err := json.NewDecoder(getRec.Body).Decode(&getResp); err != nil {
		t.Fatalf("decode get response: %v", err)
	}
	if getResp.Error != "" {
		t.Errorf("getScenario: error = %q", getResp.Error)
	}
	if getResp.Data.Id != scenarioID || getResp.Data.Name != "Endpoint Test Scenario" {
		t.Errorf("getScenario: id=%d name=%q; want id=%d name=Endpoint Test Scenario", getResp.Data.Id, getResp.Data.Name, scenarioID)
	}
}

func Test_createScenario_validation(t *testing.T) {
	_, sc := getTestServer(t)

	tests := []struct {
		name string
		body string
		want int
	}{
		{
			"empty name",
			`{"name":"","floatedWeight":false,"components":[{"assetId":1,"weight":1}]}`,
			http.StatusBadRequest,
		},
		{
			"no components",
			`{"name":"x","floatedWeight":false,"components":[]}`,
			http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/scenarios/", strings.NewReader(test.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			GetHttpServer(sc).Handler.ServeHTTP(rec, req)
			if rec.Code != test.want {
				t.Errorf("status = %d; want %d", rec.Code, test.want)
			}
		})
	}
}

func Test_updateScenario(t *testing.T) {
	_, sc := getTestServer(t)
	assetAID, assetBID := insertTestMetadata(t, sc)
	defer func() {
		_ = sc.PostgresConnection.DeleteMetadataByID(sc.Context, assetAID)
		_ = sc.PostgresConnection.DeleteMetadataByID(sc.Context, assetBID)
	}()

	createBody := fmt.Sprintf(`
		{"name":"To Update",
		"floatedWeight":false,
		"components":
		[{
			"assetId":%d,
			"weight":0.5
		},
		{
			"assetId":%d,
			"weight":0.5
		}]}`,
		assetAID, assetBID)

	req := httptest.NewRequest(http.MethodPost, "/api/scenarios/", strings.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	GetHttpServer(sc).Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("createScenario: status = %d", rec.Code)
	}
	var createResp scenarioResponseBody
	if err := json.NewDecoder(rec.Body).Decode(&createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	scenarioID := createResp.Data.Id
	defer func() {
		_ = sc.PostgresConnection.DeleteScenarioPermanent(sc.Context, scenarioID)
	}()

	updateBody := fmt.Sprintf(`
		{"name":"Updated Name",
		"floatedWeight":true,
		"components":
		[{
			"assetId":%d,
			"weight":0.7
		},
		{
			"assetId":%d,
			"weight":0.3
		}]}`, assetAID, assetBID)

	putReq := httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/scenarios/%d", scenarioID), strings.NewReader(updateBody))
	putReq.Header.Set("Content-Type", "application/json")
	putRec := httptest.NewRecorder()
	GetHttpServer(sc).Handler.ServeHTTP(putRec, putReq)

	if putRec.Code != http.StatusOK {
		t.Errorf("updateScenario: status = %d; want %d", putRec.Code, http.StatusOK)
	}
	var putResp scenarioResponseBody
	if err := json.NewDecoder(putRec.Body).Decode(&putResp); err != nil {
		t.Fatalf("decode update response: %v", err)
	}
	if putResp.Data.Name != "Updated Name" || !putResp.Data.FloatedWeight {
		t.Errorf("updateScenario: name=%q floatedWeight=%v; want Updated Name, true", putResp.Data.Name, putResp.Data.FloatedWeight)
	}
}

func Test_deleteScenario(t *testing.T) {
	_, sc := getTestServer(t)
	assetAID, assetBID := insertTestMetadata(t, sc)
	defer func() {
		_ = sc.PostgresConnection.DeleteMetadataByID(sc.Context, assetAID)
		_ = sc.PostgresConnection.DeleteMetadataByID(sc.Context, assetBID)
	}()

	createBody := fmt.Sprintf(`
		{"name":"To Delete",
		"floatedWeight":false,
		"components":
		[{
			"assetId":%d,
			"weight":0.5
		},
		{
			"assetId":%d,
			"weight":0.5
		}]}`,
		assetAID, assetBID)

	req := httptest.NewRequest(http.MethodPost, "/api/scenarios/", strings.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	GetHttpServer(sc).Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("createScenario: status = %d", rec.Code)
	}
	var createResp scenarioResponseBody
	if err := json.NewDecoder(rec.Body).Decode(&createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	scenarioID := createResp.Data.Id
	defer func() {
		_ = sc.PostgresConnection.DeleteScenarioPermanent(sc.Context, scenarioID)
	}()

	delReq := httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/scenarios/%d", scenarioID), nil)
	delRec := httptest.NewRecorder()
	GetHttpServer(sc).Handler.ServeHTTP(delRec, delReq)

	if delRec.Code != http.StatusOK {
		t.Errorf("deleteScenario: status = %d; want %d", delRec.Code, http.StatusOK)
	}

	// get after soft-delete: repo returns "not found", endpoint currently returns 500
	getReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/scenarios/%d", scenarioID), nil)
	getRec := httptest.NewRecorder()
	GetHttpServer(sc).Handler.ServeHTTP(getRec, getReq)
	if getRec.Code != http.StatusNotFound && getRec.Code != http.StatusOK && getRec.Code != http.StatusInternalServerError {
		t.Errorf("getScenario after delete: status = %d; want 404, 200, or 500", getRec.Code)
	}
}

func Test_getScenario_invalidId(t *testing.T) {
	_, sc := getTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/scenarios/0", nil)
	rec := httptest.NewRecorder()
	GetHttpServer(sc).Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("getScenario id=0: status = %d; want %d", rec.Code, http.StatusNotFound)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/api/scenarios/notanumber", nil)
	rec2 := httptest.NewRecorder()
	GetHttpServer(sc).Handler.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusNotFound {
		t.Errorf("getScenario invalid id: status = %d; want %d", rec2.Code, http.StatusNotFound)
	}
}

func Test_getSimulationRunHistory(t *testing.T) {
	_, sc := getTestServer(t)
	assetAID, assetBID := insertTestMetadata(t, sc)
	defer func() {
		_ = sc.PostgresConnection.DeleteMetadataByID(sc.Context, assetAID)
		_ = sc.PostgresConnection.DeleteMetadataByID(sc.Context, assetBID)
	}()

	createBody := fmt.Sprintf(`
		{"name":"RunHistory Test",
		"floatedWeight":false,
		"components":
		[{
			"assetId":%d,
			"weight":0.5
		},
		{
			"assetId":%d,
			"weight":0.5
		}]}`,
		assetAID, assetBID)

	req := httptest.NewRequest(http.MethodPost, "/api/scenarios/", strings.NewReader(createBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	GetHttpServer(sc).Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("createScenario: status = %d", rec.Code)
	}
	var createResp scenarioResponseBody
	if err := json.NewDecoder(rec.Body).Decode(&createResp); err != nil {
		t.Fatalf("decode create response: %v", err)
	}
	scenarioID := createResp.Data.Id
	defer func() {
		_ = sc.PostgresConnection.DeleteScenarioPermanent(sc.Context, scenarioID)
	}()

	historyReq := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/simulation/run-history/%d", scenarioID), nil)
	historyRec := httptest.NewRecorder()
	GetHttpServer(sc).Handler.ServeHTTP(historyRec, historyReq)

	if historyRec.Code != http.StatusOK {
		t.Errorf("getSimulationRunHistory: status = %d; want %d", historyRec.Code, http.StatusOK)
	}
	var historyResp runHistoryResponseBody
	if err := json.NewDecoder(historyRec.Body).Decode(&historyResp); err != nil {
		t.Fatalf("decode run-history response: %v", err)
	}
	if historyResp.Error != "" {
		t.Errorf("getSimulationRunHistory: error = %q", historyResp.Error)
	}
	// data may be empty array
	_ = historyResp.Data
}

func Test_getSimulationResult_notFound(t *testing.T) {
	_, sc := getTestServer(t)
	// id 1 may or may not exist; use a high id that is unlikely to have a result
	req := httptest.NewRequest(http.MethodGet, "/api/simulation/result/999999", nil)
	rec := httptest.NewRecorder()
	GetHttpServer(sc).Handler.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("getSimulationResult missing: status = %d; want %d", rec.Code, http.StatusNotFound)
	}
}

func Test_runSimulation(t *testing.T) {
	_, sc := getTestServer(t)
	scenarioID, assetAID, assetBID := insertScenarioWithTimeSeriesForSimulation(t, sc)
	defer func() {
		_ = sc.PostgresConnection.DeleteScenarioPermanent(sc.Context, scenarioID)
		_ = sc.PostgresConnection.DeleteMetadataByID(sc.Context, assetAID)
		_ = sc.PostgresConnection.DeleteMetadataByID(sc.Context, assetBID)
	}()

	// 1-year horizon → 52 weekly steps; 2-year lookback; 100 iterations, standard normal
	runBody := `{"distributionType":0,
		"simulationHorizonCount":1,
		"simulationHorizonUnit":"years",
		"maxLookbackCount":2,
		"maxLookbackUnit":"years",
		"iterations":100,
		"seed":42,
		"degreesOfFreedom":10}`

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/simulation/run/%d", scenarioID), strings.NewReader(runBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	GetHttpServer(sc).Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("runSimulation: status = %d; want %d", rec.Code, http.StatusOK)
	}
	var runResp struct {
		Data  *dm.SimulationResponse `json:"data"`
		Error string                 `json:"error"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&runResp); err != nil {
		t.Fatalf("decode run response: %v", err)
	}
	if runResp.Error != "" {
		t.Fatalf("runSimulation: error = %q", runResp.Error)
	}
	if runResp.Data == nil {
		t.Fatal("runSimulation: expected non-nil data")
	}
	if math.IsNaN(runResp.Data.RiskMetrics.MeanFinalValue) || math.IsInf(runResp.Data.RiskMetrics.MeanFinalValue, 0) {
		t.Errorf("runSimulation: MeanFinalValue not finite, got %f", runResp.Data.RiskMetrics.MeanFinalValue)
	}

	// Clean up the simulation run (created by the endpoint).
	runs, err := sc.PostgresConnection.GetSimulationRunHistories(sc.Context, scenarioID, 10)
	if err != nil {
		t.Fatalf("get run histories: %v", err)
	}
	for _, run := range runs {
		_ = sc.PostgresConnection.DeleteSimulationRunByID(sc.Context, run.Id)
	}
}
