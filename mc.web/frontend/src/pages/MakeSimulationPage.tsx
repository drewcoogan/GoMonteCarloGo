import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { getAssets } from '../controllers/asset';
import { getScenarios } from '../controllers/scenario';
import SimulationResultModal from '../components/simulation/SimulationResultModal';
import { SimulationResources } from '../models/simulation-resources';
import {
  getSimulationResources,
  getSimulationResult,
  getSimulationRunHistory,
  runSimulation,
} from '../controllers/simulation';
import { Asset } from '../models/asset';
import { Scenario } from '../models/scenario';
import {
  DEFAULT_SIMULATION_REQUEST_SETTINGS,
  SimulationRequestSettings,
} from '../models/simulation-request-settings';
import { SimulationResponse } from '../models/simulation-response';
import { SimulationRun } from '../models/simulation-run';
import { simulationWallTimeSettingsLine } from '../utilities/format-duration';
import {
  SettingsLine,
  WeightAtRunRow,
  buildWeightAtRunRows,
  runHasViewableResult,
  settingsLinesFromLive,
  settingsLinesFromRun,
} from '../utilities/simulation-result-view';
import { formatDurationFromNanos, wallClockNsFromRun } from '../utilities/format-duration';
import { maxHorizonCount, maxLookbackCount, type HorizonUnit, type LookbackUnit } from '../utilities/simulation-horizon';

type SimulationResultModalState = {
  result: SimulationResponse;
  settingsLines: SettingsLine[];
  weightsAtRun: WeightAtRunRow[];
};

const MakeSimulationPage: React.FC = () => {
  const [simulationResources, setSimulationResources] = useState<SimulationResources | null>(null);
  const [scenarios, setScenarios] = useState<Scenario[]>([]);
  const [loading, setLoading] = useState(true);
  const [running, setRunning] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const [selectedScenarioId, setSelectedScenarioId] = useState<number>(0);
  const [settings, setSettings] = useState<SimulationRequestSettings>(() => ({ ...DEFAULT_SIMULATION_REQUEST_SETTINGS }));
  const [runHistory, setRunHistory] = useState<SimulationRun[]>([]);
  const [loadingHistory, setLoadingHistory] = useState(false);
  const [viewingRunId, setViewingRunId] = useState<number | null>(null);
  const [modal, setModal] = useState<SimulationResultModalState | null>(null);
  const [lastCompleted, setLastCompleted] = useState<SimulationResultModalState | null>(null);
  const [assets, setAssets] = useState<Asset[]>([]);
  /** While focused, raw text; on blur we clamp into `settings` and clear. */
  const [horizonDraft, setHorizonDraft] = useState<string | null>(null);
  const [lookbackDraft, setLookbackDraft] = useState<string | null>(null);

  /*
    Use memo has a method that will run only when the dependency changes, this is the second parameter
    So if simulationResources changes, the function will run again and if it doesn't change, the function will return the cached result
  */
  const distTypeOptions = useMemo(() => {
    if (!simulationResources?.distributionType) {
        return [];
    }
    return Object.entries(simulationResources.distributionType);
  }, [simulationResources]);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [resources, scenarioList, assetList] = await Promise.all([
        getSimulationResources(), // resources for the simulation settings
        getScenarios(), // list of to choose from
        getAssets(),
      ]);

      setSimulationResources(resources);
      setScenarios(scenarioList);
      setAssets(assetList);

      if (scenarioList.length > 0) {
        setSelectedScenarioId(prev => (prev === 0 ? scenarioList[0].id : prev));
      }

      // set the default settings to the first available option
      // make sure the users will see the string representation for these, not the numeric keys
      if (resources?.distributionType && Object.keys(resources.distributionType).length > 0) {
        const firstDist = Object.values(resources.distributionType)[0];
        setSettings((prev) => ({ ...prev, distributionType: firstDist }));
      }

    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to load resources');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchData();
  }, [fetchData]);

  useEffect(() => {
    if (!selectedScenarioId) {
      setRunHistory([]);
      return;
    }
    let cancelled = false;
    (async () => {
      setLoadingHistory(true);
      try {
        const rows = await getSimulationRunHistory(selectedScenarioId);
        if (!cancelled) {
          setRunHistory(rows);
        }
      } catch (err: unknown) {
        if (!cancelled) {
          setError(err instanceof Error ? err.message : 'Failed to load run history');
        }
      } finally {
        if (!cancelled) {
          setLoadingHistory(false);
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, [selectedScenarioId]);

  const selectedScenarioName = useMemo(() => {
    return scenarios.find(s => s.id === selectedScenarioId)?.name ?? '';
  }, [scenarios, selectedScenarioId]);

  const symbolByAssetId = useMemo(() => {
    const m = new Map<number, string>();
    for (const a of assets) {
      m.set(a.id, a.symbol);
    }
    return m;
  }, [assets]);

  const selectedScenario = useMemo(
    () => scenarios.find(s => s.id === selectedScenarioId),
    [scenarios, selectedScenarioId]
  );

  const updateSettings = (patch: Partial<SimulationRequestSettings>) => {
    setSettings((prev) => ({ ...prev, ...patch }));
  };

  const handleRunSimulation = async () => {
    setError(null);
    setSuccess(null);
    setModal(null);
    setLastCompleted(null);

    if (!selectedScenarioId) {
      setError('Please select a scenario.');
      return;
    }

    if (scenarios.length === 0) {
      setError('No scenarios available. Create a scenario first.');
      return;
    }

    if (settings.iterations < 1 || settings.iterations > 100000) {
      setError('Iterations must be between 1 and 100,000.');
      return;
    }

    const hMax = maxHorizonCount(settings.simulationHorizonUnit);
    if (settings.simulationHorizonCount < 1 || settings.simulationHorizonCount > hMax) {
      setError(
        settings.simulationHorizonUnit === 'years'
          ? 'Horizon must be between 1 and 10 years.'
          : 'Horizon must be between 1 and 120 months.'
      );
      return;
    }

    const lbMax = maxLookbackCount(settings.maxLookbackUnit);
    if (settings.maxLookbackCount < 1 || settings.maxLookbackCount > lbMax) {
      setError(
        settings.maxLookbackUnit === 'years'
          ? 'Max lookback must be between 1 and 5 years.'
          : 'Max lookback must be between 1 and 60 months.'
      );
      return;
    }

    setRunning(true);
    try {
      const data = await runSimulation(selectedScenarioId, settings);
      const baseLines = settingsLinesFromLive(selectedScenarioName, settings, simulationResources);
      const wallLine = simulationWallTimeSettingsLine(data);
      const lines = wallLine ? [...baseLines, wallLine] : baseLines;
      const weightsAtRun = buildWeightAtRunRows(selectedScenario?.components, symbolByAssetId);
      const snapshot: SimulationResultModalState = { result: data, settingsLines: lines, weightsAtRun };
      setLastCompleted(snapshot);
      setModal(snapshot);
      setSuccess('Simulation completed successfully.');
      try {
        const rows = await getSimulationRunHistory(selectedScenarioId);
        setRunHistory(rows);
      } catch {
        /* list refresh is best-effort */
      }
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Simulation failed');
    } finally {
      setRunning(false);
    }
  };

  const openResultFromHistory = async (run: SimulationRun) => {
    setError(null);
    setViewingRunId(run.id);
    try {
      const data = await getSimulationResult(run.id);
      const weightsAtRun = buildWeightAtRunRows(run.components, symbolByAssetId);
      const baseLines = settingsLinesFromRun(run);
      const wallLine = simulationWallTimeSettingsLine(data, run);
      const lines = wallLine ? [...baseLines, wallLine] : baseLines;
      setModal({
        result: data,
        settingsLines: lines,
        weightsAtRun,
      });
    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : 'Failed to load simulation result');
    } finally {
      setViewingRunId(null);
    }
  };

  const studentTDistEntry = useMemo(
    () => distTypeOptions.find(([label]) => label.toLowerCase().includes('student')),
    [distTypeOptions]
  );
  const isStudentTSelected =
    studentTDistEntry !== undefined && settings.distributionType === studentTDistEntry[1];

  const settingsGridItem: React.CSSProperties = { minWidth: 0 };

  const studentTOnlyInputStyle = (enabled: boolean): React.CSSProperties => ({
    width: '100%',
    padding: 8,
    fontSize: 14,
    boxSizing: 'border-box',
    opacity: enabled ? 1 : 0.55,
    cursor: enabled ? 'text' : 'not-allowed',
    background: enabled ? undefined : '#f5f5f5',
  });

  const textInputStyle: React.CSSProperties = {
    width: '100%',
    padding: 8,
    fontSize: 14,
    boxSizing: 'border-box',
  };

  if (loading) {
    return (
      <div style={{ maxWidth: 900, margin: '0 auto', padding: 20 }}>
        <div style={{ color: '#666' }}>Loading simulation resources...</div>
      </div>
    );
  }

  return (
    <div style={{ maxWidth: 900, margin: '0 auto', padding: 20 }}>
      <h1 style={{ marginBottom: 8 }}>Run Monte Carlo Simulation</h1>
      <p style={{ color: '#666', marginBottom: 24 }}>
        Select a scenario and configure simulation settings to run a Monte Carlo simulation.
      </p>

      {error && (
        <div
          style={{
            marginBottom: 16,
            padding: 12,
            background: '#ffebee',
            color: '#c62828',
            borderRadius: 4,
            fontWeight: 'bold',
          }}
        >
          {error}
        </div>
      )}

      {success && (
        <div
          style={{
            marginBottom: 16,
            padding: 12,
            background: '#e8f5e9',
            color: '#2e7d32',
            borderRadius: 4,
            fontWeight: 'bold',
          }}
        >
          {success}
        </div>
      )}

      <div style={{ display: 'flex', flexDirection: 'column', gap: 24 }}>
        <div style={{ background: '#fff', padding: 20, borderRadius: 8, boxShadow: '0 2px 8px #eee' }}>
          <h2 style={{ marginTop: 0, marginBottom: 16 }}>Simulation Settings</h2>

          <div
            style={{
              display: 'grid',
              gridTemplateColumns: 'repeat(auto-fill, minmax(220px, 280px))',
              columnGap: 24,
              rowGap: 20,
              justifyContent: 'start',
              alignItems: 'start',
            }}
          >
            <div style={settingsGridItem}>
              <label style={{ display: 'block', fontWeight: 'bold', marginBottom: 6 }}>Scenario</label>
              <select
                value={selectedScenarioId}
                onChange={(e) => setSelectedScenarioId(Number(e.target.value))}
                style={{ width: '100%', padding: 8, fontSize: 14 }}
              >
                <option value={0}>Select scenario</option>
                {scenarios.map((s) => (
                  <option key={s.id} value={s.id}>
                    {s.name}
                  </option>
                ))}
              </select>
            </div>

            <div style={settingsGridItem}>
              <label style={{ display: 'block', fontWeight: 'bold', marginBottom: 6 }}>Distribution Type</label>
              <select
                value={settings.distributionType}
                onChange={(e) => updateSettings({ distributionType: Number(e.target.value) })}
                style={{ width: '100%', padding: 8, fontSize: 14 }}
              >
                {distTypeOptions.map(([label, value]) => (
                  <option key={label} value={value}>
                    {label.replace(/([A-Z])/g, ' $1').replace(/^./, (s) => s.toUpperCase())}
                  </option>
                ))}
              </select>
            </div>

            <div style={{ ...settingsGridItem, gridColumn: '1 / -1' }}>
              <p style={{ margin: '0 0 8px', fontSize: 13, color: '#555' }}>
                Each simulated path advances <strong>one week per step</strong> (aligned with weekly historical returns).
                Choose how far into the future the path should run using months or years (up to 10 years).
              </p>
            </div>

            <div style={settingsGridItem}>
              <label style={{ display: 'block', fontWeight: 'bold', marginBottom: 6 }}>Horizon length</label>
              <input
                type="text"
                inputMode="numeric"
                autoComplete="off"
                value={horizonDraft !== null ? horizonDraft : String(settings.simulationHorizonCount)}
                onFocus={() => setHorizonDraft(String(settings.simulationHorizonCount))}
                onChange={(e) => setHorizonDraft(e.target.value)}
                onBlur={() => {
                  const raw = (horizonDraft ?? '').trim();
                  let n = parseInt(raw.replace(/\D/g, ''), 10);
                  if (Number.isNaN(n) || n < 1) {
                    n = 1;
                  }
                  const cap = maxHorizonCount(settings.simulationHorizonUnit);
                  if (n > cap) {
                    n = cap;
                  }
                  updateSettings({ simulationHorizonCount: n });
                  setHorizonDraft(null);
                }}
                style={textInputStyle}
              />
              <span style={{ display: 'block', marginTop: 4, fontSize: 12, color: '#666' }}>
                Up to {maxHorizonCount(settings.simulationHorizonUnit)}{' '}
                {settings.simulationHorizonUnit === 'years' ? 'years' : 'months'}
              </span>
            </div>

            <div style={settingsGridItem}>
              <label style={{ display: 'block', fontWeight: 'bold', marginBottom: 6 }}>Horizon unit</label>
              <select
                value={settings.simulationHorizonUnit}
                onChange={(e) => {
                  const unit = e.target.value as HorizonUnit;
                  setHorizonDraft(null);
                  setSettings((prev) => {
                    const cap = maxHorizonCount(unit);
                    const count = Math.min(prev.simulationHorizonCount, cap);
                    return { ...prev, simulationHorizonUnit: unit, simulationHorizonCount: Math.max(1, count) };
                  });
                }}
                style={{ width: '100%', padding: 8, fontSize: 14 }}
              >
                <option value="months">Months</option>
                <option value="years">Years</option>
              </select>
            </div>

            <div style={settingsGridItem}>
              <label style={{ display: 'block', fontWeight: 'bold', marginBottom: 6 }}>Max lookback length</label>
              <input
                type="text"
                inputMode="numeric"
                autoComplete="off"
                value={lookbackDraft !== null ? lookbackDraft : String(settings.maxLookbackCount)}
                onFocus={() => setLookbackDraft(String(settings.maxLookbackCount))}
                onChange={(e) => setLookbackDraft(e.target.value)}
                onBlur={() => {
                  const raw = (lookbackDraft ?? '').trim();
                  let n = parseInt(raw.replace(/\D/g, ''), 10);
                  if (Number.isNaN(n) || n < 1) {
                    n = 1;
                  }
                  const cap = maxLookbackCount(settings.maxLookbackUnit);
                  if (n > cap) {
                    n = cap;
                  }
                  updateSettings({ maxLookbackCount: n });
                  setLookbackDraft(null);
                }}
                style={textInputStyle}
              />
              <span style={{ display: 'block', marginTop: 4, fontSize: 12, color: '#666' }}>
                Up to {maxLookbackCount(settings.maxLookbackUnit)}{' '}
                {settings.maxLookbackUnit === 'years' ? 'years' : 'months'}
              </span>
            </div>

            <div style={settingsGridItem}>
              <label style={{ display: 'block', fontWeight: 'bold', marginBottom: 6 }}>Max lookback unit</label>
              <select
                value={settings.maxLookbackUnit}
                onChange={(e) => {
                  const unit = e.target.value as LookbackUnit;
                  setLookbackDraft(null);
                  setSettings((prev) => {
                    const cap = maxLookbackCount(unit);
                    const count = Math.min(prev.maxLookbackCount, cap);
                    return { ...prev, maxLookbackUnit: unit, maxLookbackCount: Math.max(1, count) };
                  });
                }}
                style={{ width: '100%', padding: 8, fontSize: 14 }}
              >
                <option value="months">Months</option>
                <option value="years">Years</option>
              </select>
            </div>

            <div style={settingsGridItem}>
              <label style={{ display: 'block', fontWeight: 'bold', marginBottom: 6 }}>Iterations</label>
              <input
                type="number"
                min={100}
                max={100000}
                step={100}
                value={settings.iterations}
                onChange={(e) => updateSettings({ iterations: Number(e.target.value) || 1000 })}
                style={textInputStyle}
              />
            </div>

            <div style={settingsGridItem}>
              <label style={{ display: 'block', fontWeight: 'bold', marginBottom: 6 }}>Seed</label>
              <input
                type="number"
                value={settings.seed}
                onChange={(e) => updateSettings({ seed: Number(e.target.value) || 0 })}
                style={textInputStyle}
              />
            </div>

            {studentTDistEntry && (
              <div style={settingsGridItem}>
                <label
                  style={{
                    display: 'block',
                    fontWeight: 'bold',
                    marginBottom: 6,
                    color: isStudentTSelected ? undefined : '#888',
                  }}
                >
                  Degrees of Freedom
                </label>
                <input
                  type="number"
                  min={2}
                  max={100}
                  value={settings.degreesOfFreedom}
                  disabled={!isStudentTSelected}
                  onChange={(e) => updateSettings({ degreesOfFreedom: Number(e.target.value) || 5 })}
                  style={studentTOnlyInputStyle(isStudentTSelected)}
                />
              </div>
            )}
          </div>

          <div style={{ display: 'flex', flexWrap: 'wrap', gap: 12, alignItems: 'center', marginTop: 20 }}>
            <button
              type="button"
              onClick={handleRunSimulation}
              disabled={running || !selectedScenarioId}
              style={{
                padding: '12px 20px',
                fontSize: 16,
                background: '#1976d2',
                color: '#fff',
                border: 'none',
                borderRadius: 4,
                cursor: running ? 'not-allowed' : 'pointer',
                opacity: running || !selectedScenarioId ? 0.7 : 1,
              }}
            >
              {running ? 'Running...' : 'Run simulation'}
            </button>
            {lastCompleted && (
              <button type="button" className="mc-btn mc-btn--outline" onClick={() => setModal(lastCompleted)}>
                View latest results
              </button>
            )}
          </div>
        </div>

        {selectedScenarioId > 0 && (
          <div style={{ background: '#fff', padding: 20, borderRadius: 8, boxShadow: '0 2px 8px #eee' }}>
            <h2 style={{ marginTop: 0, marginBottom: 8 }}>Recent runs for this scenario</h2>
            <p style={{ marginTop: 0, marginBottom: 16, color: '#666', fontSize: 14 }}>
              Open a saved result to see settings, VaR / CVaR, and sample paths (same payload as the run API).
            </p>
            {loadingHistory && <div style={{ color: '#666' }}>Loading history…</div>}
            {!loadingHistory && runHistory.length === 0 && (
              <div style={{ color: '#666' }}>No runs yet for this scenario.</div>
            )}
            {!loadingHistory &&
              runHistory.map(run => {
                const canView = runHasViewableResult(run);
                const ended = run.endTimeUtc ? new Date(run.endTimeUtc).toLocaleString() : '—';
                const wallNs = wallClockNsFromRun(run);
                const durationLabel = wallNs != null ? formatDurationFromNanos(wallNs) : null;
                return (
                  <div
                    key={run.id}
                    style={{
                      borderTop: '1px solid #eee',
                      paddingTop: 12,
                      marginTop: 12,
                      display: 'flex',
                      flexWrap: 'wrap',
                      gap: 12,
                      alignItems: 'center',
                      justifyContent: 'space-between',
                    }}
                  >
                    <div>
                      <div style={{ fontWeight: 600 }}>Run #{run.id}</div>
                      <div style={{ fontSize: 13, color: '#666' }}>
                        Finished {ended}
                        {durationLabel ? ` · ${durationLabel}` : ''} · {run.iterations} iterations
                      </div>
                      {run.errorMessage ? (
                        <div style={{ fontSize: 13, color: '#c62828', marginTop: 4 }}>{run.errorMessage}</div>
                      ) : null}
                    </div>
                    <button
                      type="button"
                      className="mc-btn mc-btn--primary"
                      disabled={!canView || viewingRunId === run.id}
                      onClick={() => openResultFromHistory(run)}
                    >
                      {viewingRunId === run.id ? 'Loading…' : 'View results'}
                    </button>
                  </div>
                );
              })}
          </div>
        )}
      </div>

      {modal && (
        <SimulationResultModal
          result={modal.result}
          settingsLines={modal.settingsLines}
          weightsAtRun={modal.weightsAtRun}
          onClose={() => setModal(null)}
        />
      )}
    </div>
  );
};

export default MakeSimulationPage;
