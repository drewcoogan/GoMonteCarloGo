import React, { useCallback, useEffect, useMemo, useState } from 'react';
import { getScenarios, createScenario } from '../controllers/scenario';
import { getAssets } from '../controllers/asset';
import { Scenario } from '../models/scenario';
import { Asset } from '../models/asset';

const WEIGHT_SUM_TOLERANCE = 0.001;

type ScenarioComponentForm = {
  assetId: number;
  weight: string;
};

const ScenarioPage: React.FC = () => {
  const [assets, setAssets] = useState<Asset[]>([]);
  const [scenarios, setScenarios] = useState<Scenario[]>([]);
  const [loadingAssets, setLoadingAssets] = useState(false);
  const [loadingScenarios, setLoadingScenarios] = useState(false);
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const [name, setName] = useState('');
  const [floatedWeight, setFloatedWeight] = useState(false);
  const [components, setComponents] = useState<ScenarioComponentForm[]>([
    { assetId: 0, weight: '' },
  ]);

  const assetLookup = useMemo(() => {
    const lookup = new Map<number, Asset>();
    assets.forEach(asset => lookup.set(asset.id, asset));
    return lookup;
  }, [assets]);

  const normalizedComponents = useMemo(() => {
    return components
      .filter(component => component.assetId && component.weight !== '')
      .map(component => ({
        assetId: component.assetId,
        weight: Number(component.weight),
      }));
  }, [components]);

  const unassignedWeight = useMemo(() => {
    return components.reduce((sum, component) => {
      if (component.assetId) return sum;
      if (component.weight === '') return sum;
      const value = Number(component.weight);
      return Number.isFinite(value) ? sum + value : sum;
    }, 0);
  }, [components]);

  const totalWeight = useMemo(() => {
    // Must stay consistent with validation (which ignores placeholder assetId=0 rows).
    return normalizedComponents.reduce((sum, component) => {
      return Number.isFinite(component.weight) ? sum + component.weight : sum;
    }, 0);
  }, [normalizedComponents]);

  const fetchAssets = useCallback(async () => {
    setLoadingAssets(true);
    try {
      const data = await getAssets();
      setAssets(data);
    } catch (err: any) {
      setError(`Error loading assets: ${err.message}`);
    } finally {
      setLoadingAssets(false);
    }
  }, []);

  const fetchScenarios = useCallback(async () => {
    setLoadingScenarios(true);
    try {
      const data = await getScenarios();
      setScenarios(data);
    } catch (err: any) {
      setError(`Error loading scenarios: ${err.message}`);
    } finally {
      setLoadingScenarios(false);
    }
  }, []);

  useEffect(() => {
    fetchAssets();
    fetchScenarios();
  }, [fetchAssets, fetchScenarios]);

  const updateComponent = (index: number, patch: Partial<ScenarioComponentForm>) => {
    setComponents(prev =>
      prev.map((component, idx) =>
        idx === index ? { ...component, ...patch } : component
      )
    );
  };

  const addComponent = () => {
    setComponents(prev => [...prev, { assetId: 0, weight: '' }]);
  };

  const removeComponent = (index: number) => {
    setComponents(prev => prev.filter((_, idx) => idx !== index));
  };

  const resetForm = () => {
    setName('');
    setFloatedWeight(false);
    setComponents([{ assetId: 0, weight: '' }]);
  };

  const handleCreateScenario = async () => {
    setError(null);
    setSuccess(null);

    const trimmedName = name.trim();
    if (!trimmedName) {
      setError('Scenario name is required.');
      return;
    }

    if (normalizedComponents.length === 0) {
      setError('Add at least one component.');
      return;
    }

    if (normalizedComponents.some(component => !Number.isFinite(component.weight) || component.weight <= 0)) {
      setError('Component weights must be non zero and positive.');
      return;
    }

    const weightSum = normalizedComponents.reduce((sum, component) => sum + component.weight, 0);
    if (Math.abs(weightSum - 1) > WEIGHT_SUM_TOLERANCE) {
      setError(`Weights must sum to 1.0 (currently ${weightSum.toFixed(4)}).`);
      return;
    }

    const seen = new Set<number>();
    for (const component of normalizedComponents) {
      if (seen.has(component.assetId)) {
        setError('Each asset can only appear once.');
        return;
      }
      seen.add(component.assetId);
    }

    setSaving(true);
    try {
      await createScenario({
        name: trimmedName,
        floatedWeight,
        components: normalizedComponents,
      });

      await fetchScenarios();
      resetForm();
      setSuccess('Scenario created successfully.');
    } catch (err: any) {
      setError(`Error creating scenario: ${err.message}`);
    } finally {
      setSaving(false);
    }
  };

  return (
    <div style={{ maxWidth: 1200, margin: '0 auto', padding: 20 }}>
      <h1 style={{ marginBottom: 8 }}>Scenario Builder</h1>
      <p style={{ color: '#666', marginBottom: 24 }}>
        Create allocation scenarios from synced assets and save them for simulations.
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

      <div style={{ display: 'flex', gap: 24, flexWrap: 'wrap' }}>
        <div style={{ flex: '1 1 420px', background: '#fff', padding: 20, borderRadius: 8, boxShadow: '0 2px 8px #eee' }}>
          <h2 style={{ marginTop: 0 }}>New Scenario</h2>
          <div style={{ display: 'flex', flexDirection: 'column', gap: 12 }}>
            <label style={{ fontWeight: 'bold' }}>Scenario Name</label>
            <input
              type="text"
              value={name}
              onChange={e => setName(e.target.value)}
              placeholder="e.g. Balanced Allocation"
              style={{ padding: 8, fontSize: 16 }}
            />

            <label style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
              <input
                type="checkbox"
                checked={floatedWeight}
                onChange={e => setFloatedWeight(e.target.checked)}
              />
              Floated weights (rebalance during simulation)
            </label>

            <div style={{ marginTop: 8 }}>
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <h3 style={{ margin: 0 }}>Components</h3>
                <button
                  type="button"
                  onClick={addComponent}
                  style={{ padding: '6px 10px', borderRadius: 4, border: '1px solid #1976d2', background: '#fff', color: '#1976d2' }}
                >
                  + Add
                </button>
              </div>

              {components.map((component, index) => (
                <div
                  key={`${component.assetId}-${index}`}
                  style={{ display: 'flex', gap: 8, marginTop: 12, alignItems: 'center' }}
                >
                  <select
                    value={component.assetId}
                    onChange={e => updateComponent(index, { assetId: Number(e.target.value) })}
                    style={{ flex: 2, padding: 8, fontSize: 14 }}
                  >
                    <option value={0}>Select asset</option>
                    {assets.map(asset => (
                      <option key={asset.id} value={asset.id}>
                        {asset.symbol}
                      </option>
                    ))}
                  </select>
                  <input
                    type="number"
                    step="0.01"
                    min="0"
                    value={component.weight}
                    onChange={e => updateComponent(index, { weight: e.target.value })}
                    placeholder="Weight"
                    style={{ flex: 1, padding: 8, fontSize: 14 }}
                  />
                  <button
                    type="button"
                    onClick={() => removeComponent(index)}
                    disabled={components.length === 1}
                    style={{
                      padding: '6px 10px',
                      borderRadius: 4,
                      border: '1px solid #ccc',
                      background: '#fff',
                      color: '#666',
                      cursor: components.length === 1 ? 'not-allowed' : 'pointer',
                    }}
                  >
                    Remove
                  </button>
                </div>
              ))}
            </div>

            <div
              style={{
                marginTop: 8,
                color: Math.abs(totalWeight - 1) <= WEIGHT_SUM_TOLERANCE ? '#2e7d32' : '#ef6c00',
              }}
            >
              Total weight: {totalWeight.toFixed(4)}
            </div>
            {unassignedWeight > 0 && (
              <div style={{ marginTop: 6, color: '#c62828', fontSize: 13 }}>
                Unassigned weight (rows without an asset): {unassignedWeight.toFixed(4)}
              </div>
            )}

            <button
              type="button"
              onClick={handleCreateScenario}
              disabled={saving || loadingAssets}
              style={{
                marginTop: 12,
                padding: 12,
                fontSize: 16,
                background: '#1976d2',
                color: '#fff',
                border: 'none',
                borderRadius: 4,
                cursor: saving ? 'not-allowed' : 'pointer',
                opacity: saving ? 0.7 : 1,
              }}
            >
              {saving ? 'Saving...' : 'Save Scenario'}
            </button>

            {loadingAssets && <div style={{ color: '#666' }}>Loading assets...</div>}
          </div>
        </div>

        <div style={{ flex: '1 1 520px', background: '#fff', padding: 20, borderRadius: 8, boxShadow: '0 2px 8px #eee' }}>
          <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
            <h2 style={{ marginTop: 0 }}>Saved Scenarios</h2>
            <button
              type="button"
              onClick={fetchScenarios}
              style={{ padding: '6px 10px', borderRadius: 4, border: '1px solid #1976d2', background: '#fff', color: '#1976d2' }}
            >
              Refresh
            </button>
          </div>

          {loadingScenarios && <div style={{ color: '#666' }}>Loading scenarios...</div>}

          {!loadingScenarios && scenarios.length === 0 && (
            <div style={{ color: '#666' }}>No scenarios saved yet.</div>
          )}

          {scenarios.map(scenario => (
            <div
              key={scenario.id}
              style={{ borderTop: '1px solid #eee', paddingTop: 12, marginTop: 12 }}
            >
              <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                <div>
                  <strong>{scenario.name}</strong>
                  <div style={{ color: '#666', fontSize: 12 }}>
                    Updated {new Date(scenario.updatedAt).toLocaleString()}
                  </div>
                </div>
                <span style={{ fontSize: 12, color: '#666' }}>
                  {scenario.floatedWeight ? 'Floated' : 'Fixed'} weights
                </span>
              </div>
              <ul style={{ margin: '8px 0 0 18px', padding: 0 }}>
                {scenario.components.map(component => (
                  <li key={`${scenario.id}-${component.assetId}`} style={{ fontSize: 14 }}>
                    {assetLookup.get(component.assetId)?.symbol || `Asset ${component.assetId}`} —{' '}
                    {(component.weight * 100).toFixed(2)}%
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default ScenarioPage;
