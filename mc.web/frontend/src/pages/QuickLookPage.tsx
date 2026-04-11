import React, { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import {
  Bar,
  BarChart,
  CartesianGrid,
  Legend,
  Line,
  LineChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
  YAxis,
} from 'recharts';
import { sampleCorrelation } from 'simple-statistics';
import { getAssetPrices, getAssets } from '../controllers/asset';
import { AssetPricePoint } from '../models/asset-prices';
import { Asset } from '../models/asset';

const CHART_COLORS = ['#1976d2', '#d32f2f', '#388e3c', '#f57c00', '#7b1fa2', '#0097a7'];

function formatYMD(d: Date): string {
  const y = d.getFullYear();
  const m = String(d.getMonth() + 1).padStart(2, '0');
  const day = String(d.getDate()).padStart(2, '0');
  return `${y}-${m}-${day}`;
}

function threeYearsAgoYMD(): string {
  const d = new Date();
  d.setFullYear(d.getFullYear() - 3);
  return formatYMD(d);
}

function getDateBoundsFromLoaded(loaded: { symbol: string; points: AssetPricePoint[] }[]): { min: string; max: string } | null {
  if (loaded.length === 0) {
    return null;
  }
  let min = '';
  let max = '';
  for (const L of loaded) {
    for (const p of L.points) {
      if (!min || p.date < min) {
        min = p.date;
      }
      if (!max || p.date > max) {
        max = p.date;
      }
    }
  }
  if (!min || !max) {
    return null;
  }
  return { min, max };
}

function normalizeDateRange(start: string, end: string): { start: string; end: string } {
  if (!start || !end) {
    return { start, end };
  }
  return start <= end ? { start, end } : { start: end, end: start };
}

function filterPointsByDateRange(points: AssetPricePoint[], start: string, end: string): AssetPricePoint[] {
  const { start: a, end: b } = normalizeDateRange(start, end);
  return points.filter(p => p.date >= a && p.date <= b);
}

function correlationPair(a: number[], b: number[]): number | null {
  if (a.length < 2 || a.length !== b.length) {
    return null;
  }
  try {
    const r = sampleCorrelation(a, b);
    return Number.isFinite(r) ? r : null;
  } catch {
    return null;
  }
}

function toDateMap(points: AssetPricePoint[]): Map<string, AssetPricePoint> {
  const m = new Map<string, AssetPricePoint>();
  for (const p of points) {
    m.set(p.date, p);
  }
  return m;
}

/** One row per calendar date; each symbol column is adjusted close (or undefined if no trade). */
function mergeAdjustedClose(loaded: { symbol: string; points: AssetPricePoint[] }[]) {
  const maps = loaded.map(L => ({ symbol: L.symbol, map: toDateMap(L.points) }));
  const dates = new Set<string>();
  for (const { map } of maps) {
    map.forEach((_v, d) => {
      dates.add(d);
    });
  }
  const sorted = Array.from(dates).sort();
  return sorted.map(date => {
    const row: Record<string, string | number | undefined> = { date };
    for (const { symbol, map } of maps) {
      const pt = map.get(date);
      row[symbol] = pt !== undefined ? pt.adjustedClose : undefined;
    }
    return row;
  });
}

function dailyReturnsFromApi(points: AssetPricePoint[]): { date: string; value: number }[] {
  const out: { date: string; value: number }[] = [];
  for (const p of points) {
    if (p.dailyReturn != null && Number.isFinite(p.dailyReturn)) {
      out.push({ date: p.date, value: p.dailyReturn });
    }
  }
  return out;
}

function rolling5DayReturnsFromApi(points: AssetPricePoint[]): { date: string; value: number }[] {
  const out: { date: string; value: number }[] = [];
  for (const p of points) {
    if (p.rolling5DayReturn != null && Number.isFinite(p.rolling5DayReturn)) {
      out.push({ date: p.date, value: p.rolling5DayReturn });
    }
  }
  return out;
}

/** Dates where every series has a finite dailyReturn (for correlation). */
function commonDatesForDailyReturns(loaded: { symbol: string; points: AssetPricePoint[] }[]): string[] {
  const maps = loaded.map(L => {
    const m = new Map<string, number>();
    for (const p of L.points) {
      if (p.dailyReturn != null && Number.isFinite(p.dailyReturn)) {
        m.set(p.date, p.dailyReturn);
      }
    }
    return m;
  });
  let dates = Array.from(maps[0].keys());
  for (let i = 1; i < maps.length; i++) {
    dates = dates.filter(d => maps[i].has(d));
  }
  return dates.sort();
}

function buildCorrelationMatrix(loaded: { symbol: string; points: AssetPricePoint[] }[]): {
  symbols: string[];
  matrix: (number | null)[][] | null;
  overlapDays: number;
} | null {
  if (loaded.length < 2) {
    return null;
  }
  const symbols = loaded.map(l => l.symbol);
  const maps = loaded.map(L => {
    const m = new Map<string, number>();
    for (const p of L.points) {
      if (p.dailyReturn != null && Number.isFinite(p.dailyReturn)) {
        m.set(p.date, p.dailyReturn);
      }
    }
    return m;
  });
  const dates = commonDatesForDailyReturns(loaded);
  if (dates.length < 2) {
    return { symbols, matrix: null, overlapDays: dates.length };
  }
  const vectors = maps.map(m => dates.map(d => m.get(d) as number));
  const n = symbols.length;
  const matrix: (number | null)[][] = Array.from({ length: n }, () => Array(n).fill(null));
  for (let i = 0; i < n; i++) {
    matrix[i][i] = 1;
    for (let j = i + 1; j < n; j++) {
      const r = correlationPair(vectors[i], vectors[j]);
      matrix[i][j] = r;
      matrix[j][i] = r;
    }
  }
  return { symbols, matrix, overlapDays: dates.length };
}

function mergeReturnSeries(loaded: { symbol: string; series: { date: string; value: number }[] }[]) {
  const maps = loaded.map(L => {
    const map = new Map<string, number>();
    for (const x of L.series) {
      map.set(x.date, x.value);
    }
    return { symbol: L.symbol, map };
  });
  const dates = new Set<string>();
  for (const { map } of maps) {
    map.forEach((_v, d) => {
      dates.add(d);
    });
  }
  const sorted = Array.from(dates).sort();
  return sorted.map(date => {
    const row: Record<string, string | number | undefined> = { date };
    for (const { symbol, map } of maps) {
      const v = map.get(date);
      row[symbol] = v;
    }
    return row;
  });
}

function pctTooltipFormatter(value: unknown): string {
  if (typeof value !== 'number' || !Number.isFinite(value)) {
    return '';
  }
  return `${(value * 100).toFixed(2)}%`;
}

function usdTooltipFormatter(value: unknown): string {
  if (typeof value !== 'number' || !Number.isFinite(value)) {
    return '';
  }
  return value.toLocaleString('en-US', { style: 'currency', currency: 'USD', maximumFractionDigits: 2 });
}

function CorrelationTableBlock({
  symbols,
  matrix,
}: {
  symbols: string[];
  matrix: (number | null)[][];
}): React.ReactElement {
  return (
    <div
      style={{
        overflowX: 'auto',
        background: '#fff',
        borderRadius: 8,
        padding: 16,
        boxShadow: '0 2px 8px #eee',
      }}
    >
      <table style={{ borderCollapse: 'collapse', fontSize: 14, margin: '0 auto' }}>
        <thead>
          <tr>
            <th style={{ padding: '8px 12px', borderBottom: '2px solid #ddd' }} />
            {symbols.map(sym => (
              <th key={sym} style={{ padding: '8px 12px', borderBottom: '2px solid #ddd', fontWeight: 700 }}>
                {sym}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>
          {symbols.map((rowSym, i) => (
            <tr key={rowSym}>
              <th
                style={{
                  padding: '8px 12px',
                  textAlign: 'left',
                  borderBottom: '1px solid #eee',
                  fontWeight: 700,
                }}
              >
                {rowSym}
              </th>
              {matrix[i].map((cell, j) => (
                <td
                  key={`${rowSym}-${symbols[j]}`}
                  style={{
                    padding: '8px 12px',
                    borderBottom: '1px solid #eee',
                    textAlign: 'right',
                    fontFamily: 'ui-monospace, monospace',
                  }}
                >
                  {cell === null ? '—' : cell.toFixed(3)}
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}

const QuickLookPage: React.FC = () => {
  const [catalog, setCatalog] = useState<Asset[]>([]);
  const [catalogError, setCatalogError] = useState<string | null>(null);
  const [dropdownId, setDropdownId] = useState<string>('');
  const [selected, setSelected] = useState<Asset[]>([]);
  const [seriesById, setSeriesById] = useState<
    Record<number, { symbol: string; points: AssetPricePoint[] } | 'loading' | 'error'>
  >({});
  const [chartRangeStart, setChartRangeStart] = useState('');
  const [chartRangeEnd, setChartRangeEnd] = useState('');
  const chartRangeInitRef = useRef(false);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const list = await getAssets();
        if (!cancelled) {
          setCatalog(list);
        }
      } catch (e) {
        if (!cancelled) {
          setCatalogError(e instanceof Error ? e.message : 'Failed to load assets');
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  const loadSeries = useCallback(async (asset: Asset) => {
    setSeriesById(prev => ({ ...prev, [asset.id]: 'loading' }));
    try {
      const payload = await getAssetPrices(asset.id);
      setSeriesById(prev => ({ ...prev, [asset.id]: { symbol: payload.symbol, points: payload.points } }));
    } catch {
      setSeriesById(prev => ({ ...prev, [asset.id]: 'error' }));
    }
  }, []);

  useEffect(() => {
    for (const a of selected) {
      const cur = seriesById[a.id];
      if (cur === undefined) {
        loadSeries(a);
      }
    }
  }, [selected, seriesById, loadSeries]);

  const loaded = useMemo(() => {
    const rows: { symbol: string; points: AssetPricePoint[] }[] = [];
    for (const a of selected) {
      const s = seriesById[a.id];
      if (s && s !== 'loading' && s !== 'error' && s.points.length > 0) {
        rows.push({ symbol: s.symbol, points: s.points });
      }
    }
    return rows;
  }, [selected, seriesById]);

  const dataDateBounds = useMemo(() => getDateBoundsFromLoaded(loaded), [loaded]);

  useEffect(() => {
    if (loaded.length === 0) {
      chartRangeInitRef.current = false;
      setChartRangeStart('');
      setChartRangeEnd('');
      return;
    }
    if (chartRangeInitRef.current) {
      return;
    }
    const b = getDateBoundsFromLoaded(loaded);
    if (!b) {
      return;
    }
    const end = b.max;
    const startCandidate = threeYearsAgoYMD();
    const start = startCandidate <= end ? startCandidate : b.min;
    setChartRangeStart(start);
    setChartRangeEnd(end);
    chartRangeInitRef.current = true;
  }, [loaded]);

  const applyDefaultDateRange = useCallback(() => {
    const b = getDateBoundsFromLoaded(loaded);
    if (!b) {
      return;
    }
    const end = b.max;
    const startCandidate = threeYearsAgoYMD();
    const start = startCandidate <= end ? startCandidate : b.min;
    setChartRangeStart(start);
    setChartRangeEnd(end);
  }, [loaded]);

  const loadedInRange = useMemo(() => {
    if (!chartRangeStart || !chartRangeEnd) {
      return loaded;
    }
    const { start, end } = normalizeDateRange(chartRangeStart, chartRangeEnd);
    return loaded.map(L => ({
      symbol: L.symbol,
      points: filterPointsByDateRange(L.points, start, end),
    }));
  }, [loaded, chartRangeStart, chartRangeEnd]);

  const addFromDropdown = () => {
    const id = Number(dropdownId);
    if (!id) {
      return;
    }
    const asset = catalog.find(x => x.id === id);
    if (!asset) {
      return;
    }
    if (selected.some(s => s.id === id)) {
      return;
    }
    setSelected(prev => [...prev, asset]);
  };

  const removeAsset = (id: number) => {
    setSelected(prev => prev.filter(a => a.id !== id));
  };

  const priceData = useMemo(() => mergeAdjustedClose(loadedInRange), [loadedInRange]);
  const dailyData = useMemo(
    () =>
      mergeReturnSeries(
        loadedInRange.map(L => ({
          symbol: L.symbol,
          series: dailyReturnsFromApi(L.points),
        })),
      ),
    [loadedInRange],
  );
  const weeklyData = useMemo(
    () =>
      mergeReturnSeries(
        loadedInRange.map(L => ({
          symbol: L.symbol,
          series: rolling5DayReturnsFromApi(L.points),
        })),
      ),
    [loadedInRange],
  );

  const correlation = useMemo(() => buildCorrelationMatrix(loadedInRange), [loadedInRange]);

  const symbols = loadedInRange.map(l => l.symbol);

  return (
    <div style={{ maxWidth: 1200, margin: '0 auto', padding: '24px 20px 48px' }}>
      <h1 style={{ textAlign: 'center', marginBottom: 8 }}>Quick look</h1>
      <p style={{ textAlign: 'center', color: '#555', marginBottom: 24 }}>
        Choose synced assets and compare price, daily returns, and five-day rolling returns.
      </p>

      <div
        style={{
          display: 'flex',
          flexWrap: 'wrap',
          gap: 12,
          alignItems: 'center',
          marginBottom: 20,
          padding: 16,
          background: '#fff',
          borderRadius: 8,
          boxShadow: '0 2px 8px #eee',
        }}
      >
        <select
          value={dropdownId}
          onChange={e => setDropdownId(e.target.value)}
          style={{ padding: '8px 12px', fontSize: 15, minWidth: 200 }}
        >
          <option value="">Select an asset…</option>
          {catalog.map(a => (
            <option key={a.id} value={a.id}>
              {a.symbol}
            </option>
          ))}
        </select>
        <button
          type="button"
          onClick={addFromDropdown}
          disabled={!dropdownId}
          style={{
            padding: '8px 16px',
            fontSize: 15,
            background: '#1976d2',
            color: '#fff',
            border: 'none',
            borderRadius: 4,
            cursor: dropdownId ? 'pointer' : 'not-allowed',
            opacity: dropdownId ? 1 : 0.6,
          }}
        >
          Add to chart
        </button>
        {catalogError && <span style={{ color: '#c62828' }}>{catalogError}</span>}
      </div>

      {selected.length > 0 && (
        <div style={{ display: 'flex', flexWrap: 'wrap', gap: 8, marginBottom: 20 }}>
          {selected.map(a => {
            const st = seriesById[a.id];
            const label =
              st === 'loading'
                ? `${a.symbol} …`
                : st === 'error'
                  ? `${a.symbol} (load failed)`
                  : typeof st === 'object' && st !== null && 'points' in st
                    ? `${st.symbol} (${st.points.length} days)`
                    : a.symbol;
            return (
              <button
                key={a.id}
                type="button"
                onClick={() => removeAsset(a.id)}
                style={{
                  padding: '6px 12px',
                  borderRadius: 16,
                  border: '1px solid #ccc',
                  background: '#fafafa',
                  cursor: 'pointer',
                  fontSize: 14,
                }}
              >
                {label} ×
              </button>
            );
          })}
        </div>
      )}

      {selected.length === 0 && (
        <p style={{ textAlign: 'center', color: '#888' }}>Add at least one asset to see charts.</p>
      )}

      {selected.length > 0 && loaded.length === 0 && (
        <p style={{ textAlign: 'center', color: '#888' }}>
          {Object.values(seriesById).some(s => s === 'loading')
            ? 'Loading series…'
            : 'No price data yet — sync an asset on Sync Data first.'}
        </p>
      )}

      {loaded.length > 0 && chartRangeStart && chartRangeEnd && dataDateBounds && (
        <>
          <div
            style={{
              display: 'flex',
              flexWrap: 'wrap',
              alignItems: 'center',
              gap: 12,
              marginBottom: 20,
              padding: 16,
              background: '#fff',
              borderRadius: 8,
              boxShadow: '0 2px 8px #eee',
            }}
          >
            <span style={{ fontWeight: 700, fontSize: 15 }}>Chart date range</span>
            <label style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 14 }}>
              Start
              <input
                type="date"
                value={chartRangeStart}
                min={dataDateBounds.min}
                max={chartRangeEnd}
                onChange={e => setChartRangeStart(e.target.value)}
                style={{ padding: 6, fontSize: 14 }}
              />
            </label>
            <label style={{ display: 'flex', alignItems: 'center', gap: 8, fontSize: 14 }}>
              End
              <input
                type="date"
                value={chartRangeEnd}
                min={chartRangeStart}
                max={dataDateBounds.max}
                onChange={e => setChartRangeEnd(e.target.value)}
                style={{ padding: 6, fontSize: 14 }}
              />
            </label>
            <button
              type="button"
              onClick={applyDefaultDateRange}
              style={{
                padding: '8px 14px',
                fontSize: 14,
                background: '#fff',
                color: '#1976d2',
                border: '1px solid #1976d2',
                borderRadius: 4,
                cursor: 'pointer',
              }}
            >
              Reset to default
            </button>
            <span style={{ fontSize: 13, color: '#666' }}>
              Default: three years before today through the latest date in the loaded data.
            </span>
          </div>

          {priceData.length === 0 && (
            <p style={{ textAlign: 'center', color: '#888', marginBottom: 16 }}>
              No price observations in the selected range — widen the date range or sync more history.
            </p>
          )}

          <section style={{ marginBottom: 32 }}>
            <h2 style={{ fontSize: 18, marginBottom: 8 }}>Adjusted close</h2>
            <div style={{ width: '100%', height: 360, background: '#fff', borderRadius: 8, padding: 8, boxSizing: 'border-box' }}>
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={priceData} margin={{ top: 8, right: 16, left: 8, bottom: 8 }}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="date" tick={{ fontSize: 11 }} interval="preserveStartEnd" minTickGap={40} />
                  <YAxis tick={{ fontSize: 11 }} tickFormatter={v => (typeof v === 'number' ? v.toFixed(0) : '')} />
                  <Tooltip formatter={usdTooltipFormatter} labelFormatter={d => String(d)} />
                  <Legend />
                  {symbols.map((sym, i) => (
                    <Line
                      key={sym}
                      type="monotone"
                      dataKey={sym}
                      stroke={CHART_COLORS[i % CHART_COLORS.length]}
                      dot={false}
                      strokeWidth={1.5}
                      connectNulls
                    />
                  ))}
                </LineChart>
              </ResponsiveContainer>
            </div>
          </section>

          <section style={{ marginBottom: 32 }}>
            <h2 style={{ fontSize: 18, marginBottom: 8 }}>Day-over-day returns</h2>
            <p style={{ fontSize: 14, color: '#666', marginBottom: 8 }}>Bar chart (adjusted close).</p>
            <div style={{ width: '100%', height: 360, background: '#fff', borderRadius: 8, padding: 8, boxSizing: 'border-box' }}>
              <ResponsiveContainer width="100%" height="100%">
                <BarChart data={dailyData} margin={{ top: 8, right: 16, left: 8, bottom: 8 }}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="date" tick={{ fontSize: 11 }} interval="preserveStartEnd" minTickGap={40} />
                  <YAxis tick={{ fontSize: 11 }} tickFormatter={v => (typeof v === 'number' ? `${(v * 100).toFixed(1)}%` : '')} />
                  <Tooltip formatter={pctTooltipFormatter} labelFormatter={d => String(d)} />
                  <Legend />
                  {symbols.map((sym, i) => (
                    <Bar key={sym} dataKey={sym} fill={CHART_COLORS[i % CHART_COLORS.length]} opacity={0.85} />
                  ))}
                </BarChart>
              </ResponsiveContainer>
            </div>
          </section>

          <section>
            <h2 style={{ fontSize: 18, marginBottom: 8 }}>Rolling five trading-day returns</h2>
            <p style={{ fontSize: 14, color: '#666', marginBottom: 8 }}>Total return over the prior five sessions (adjusted close).</p>
            <div style={{ width: '100%', height: 360, background: '#fff', borderRadius: 8, padding: 8, boxSizing: 'border-box' }}>
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={weeklyData} margin={{ top: 8, right: 16, left: 8, bottom: 8 }}>
                  <CartesianGrid strokeDasharray="3 3" />
                  <XAxis dataKey="date" tick={{ fontSize: 11 }} interval="preserveStartEnd" minTickGap={40} />
                  <YAxis tick={{ fontSize: 11 }} tickFormatter={v => (typeof v === 'number' ? `${(v * 100).toFixed(1)}%` : '')} />
                  <Tooltip formatter={pctTooltipFormatter} labelFormatter={d => String(d)} />
                  <Legend />
                  {symbols.map((sym, i) => (
                    <Line
                      key={sym}
                      type="monotone"
                      dataKey={sym}
                      stroke={CHART_COLORS[i % CHART_COLORS.length]}
                      dot={false}
                      strokeWidth={1.5}
                      connectNulls
                    />
                  ))}
                </LineChart>
              </ResponsiveContainer>
            </div>
          </section>

          {correlation && (
            <section style={{ marginTop: 32 }}>
              <h2 style={{ fontSize: 18, marginBottom: 8 }}>Daily return correlation</h2>
              <p style={{ fontSize: 14, color: '#666', marginBottom: 12 }}>
                Pearson correlation of daily simple returns (via{' '}
                <code style={{ fontSize: 13 }}>simple-statistics</code>
                ) in the selected chart range, on dates where every selected asset has a return (
                {correlation.overlapDays} overlapping day{correlation.overlapDays === 1 ? '' : 's'}).
              </p>
              {correlation.matrix === null ? (
                <p style={{ color: '#888' }}>
                  Not enough overlapping daily returns (need at least two common trading days with returns).
                </p>
              ) : (
                <CorrelationTableBlock symbols={correlation.symbols} matrix={correlation.matrix} />
              )}
            </section>
          )}
        </>
      )}
    </div>
  );
};

export default QuickLookPage;
