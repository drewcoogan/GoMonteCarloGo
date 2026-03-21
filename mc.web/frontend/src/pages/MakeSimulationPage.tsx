import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { getScenarios } from '../controllers/scenario';
import SimulationResultModal from '../components/simulation/SimulationResultModal';
import { SimulationResources } from '../models/simulation-resources';
import {
  getSimulationResources,
  getSimulationResult,
  getSimulationRunHistory,
  runSimulation,
} from '../controllers/simulation';
import { Scenario } from '../models/scenario';
import { DEFAULT_SIMULATION_REQUEST_SETTINGS, SimulationRequestSettings } from '../models/simulation-request-settings';
import { SimulationResponse } from '../models/simulation-response';
import { SimulationRun } from '../models/simulation-run';
import { SettingsLine, runHasViewableResult, settingsLinesFromLive, settingsLinesFromRun } from '../utilities/simulation-result-view';
import { DaysToNanoseconds, NanosecondsToDays } from '../utilities/time';

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
  const [modal, setModal] = useState<{ result: SimulationResponse; settingsLines: SettingsLine[] } | null>(null);
  const [lastCompleted, setLastCompleted] = useState<{ result: SimulationResponse; settingsLines: SettingsLine[] } | null>(
    null
  );

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

  const unitOfTimeOptions = useMemo(() => {
    if (!simulationResources?.simulationUnitOfTime) {
        return [];
    }
    return Object.entries(simulationResources.simulationUnitOfTime);
  }, [simulationResources]);

  const durationOptions = useMemo(() => {
    if (!simulationResources?.simulationDuration) {
        return [];
    }
    return Object.entries(simulationResources.simulationDuration);
  }, [simulationResources]);

  const fetchData = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [resources, scenarioList] = await Promise.all([
        getSimulationResources(), // resources for the simulation settings
        getScenarios(), // list of to choose from
      ]);

      setSimulationResources(resources);
      setScenarios(scenarioList);

      if (scenarioList.length > 0) {
        setSelectedScenarioId(prev => (prev === 0 ? scenarioList[0].id : prev));
      }

      // set the default settings to the first available option
      // make sure the users will see the string representation for these, not the numeric keys
      if (resources?.distributionType && Object.keys(resources.distributionType).length > 0) {
        const firstDist = Object.values(resources.distributionType)[0];
        setSettings((prev) => ({ ...prev, distributionType: firstDist }));
      }

      if (resources?.simulationUnitOfTime && Object.keys(resources.simulationUnitOfTime).length > 0) {
        const firstUnit = Object.values(resources.simulationUnitOfTime)[0];
        setSettings((prev) => ({ ...prev, simulationUnitOfTime: firstUnit }));
      }

      if (resources?.simulationDuration && Object.keys(resources.simulationDuration).length > 0) {
        const firstDur = Object.values(resources.simulationDuration)[0];
        setSettings((prev) => ({ ...prev, simulationDuration: firstDur }));
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

    setRunning(true);
    try {
      const data = await runSimulation(selectedScenarioId, settings);
      const lines = settingsLinesFromLive(selectedScenarioName, settings, simulationResources);
      const snapshot = { result: data, settingsLines: lines };
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
      setModal({ result: data, settingsLines: settingsLinesFromRun(run) });
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

            <div style={settingsGridItem}>
              <label style={{ display: 'block', fontWeight: 'bold', marginBottom: 6 }}>Unit of Time</label>
              <select
                value={settings.simulationUnitOfTime}
                onChange={(e) => updateSettings({ simulationUnitOfTime: Number(e.target.value) })}
                style={{ width: '100%', padding: 8, fontSize: 14 }}
              >
                {unitOfTimeOptions.map(([label, value]) => (
                  <option key={label} value={value}>
                    {label}
                  </option>
                ))}
              </select>
            </div>

            <div style={settingsGridItem}>
              <label style={{ display: 'block', fontWeight: 'bold', marginBottom: 6 }}>Duration (periods)</label>
              <select
                value={settings.simulationDuration}
                onChange={(e) => updateSettings({ simulationDuration: Number(e.target.value) })}
                style={{ width: '100%', padding: 8, fontSize: 14 }}
              >
                {durationOptions.map(([label, value]) => (
                  <option key={label} value={value}>
                    {value} {label}
                  </option>
                ))}
              </select>
            </div>

            <div style={settingsGridItem}>
              <label style={{ display: 'block', fontWeight: 'bold', marginBottom: 6 }}>Max Lookback (days)</label>
              <input
                type="number"
                min={1}
                max={3650}
                value={NanosecondsToDays(settings.maxLookback)}
                onChange={(e) => {
                  const days = Number(e.target.value);
                  if (!Number.isNaN(days) && days > 0) {
                    updateSettings({ maxLookback: DaysToNanoseconds(days) });
                  }
                }}
                style={textInputStyle}
              />
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
                        Finished {ended} · {run.iterations} iterations
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
          onClose={() => setModal(null)}
        />
      )}
    </div>
  );
};

export default MakeSimulationPage;
