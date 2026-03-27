import React, { useMemo } from 'react';
import { CartesianGrid, Line, LineChart, ResponsiveContainer, Tooltip, XAxis, YAxis } from 'recharts';
import { SAMPLE_PATH_ROLES, SamplePath } from '../../models/simulation-response';

/** Stroke colors for each return percentile band (matches backend ordering). */
const PERCENTILE_STROKE: Record<number, string> = {
  0.05: '#1976d2',
  0.25: '#d32f2f',
  0.5: '#388e3c',
  0.75: '#f57c00',
  0.95: '#7b1fa2',
};

/** Fixed colors for single-path exemplars. */
const ROLE_STROKE: Record<string, string> = {
  [SAMPLE_PATH_ROLES.maxDrawdown]: '#0097a7',
  [SAMPLE_PATH_ROLES.maxVolatility]: '#c2185b',
};

const SAMPLE_ROLE_STYLE = {
  stroke: '#455a64',
  strokeOpacity: 0.42,
  strokeDasharray: '6 4' as const,
};

/** Paths are absolute portfolio value in USD; simulation starts at $100 (see backend `InitialPortfolioValue`). */
function formatUsdWhole(value: number): string {
  if (!Number.isFinite(value)) {
    return '';
  }
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    maximumFractionDigits: 0,
    minimumFractionDigits: 0,
  }).format(value);
}

function approxKey(percentile: number): number | undefined {
  const keys = Object.keys(PERCENTILE_STROKE).map(Number);
  for (const k of keys) {
    if (Math.abs(k - percentile) < 1e-6) {
      return k;
    }
  }
  return undefined;
}

function lineStyleForPath(path: SamplePath): { stroke: string; strokeOpacity?: number; strokeDasharray?: string } {
  if (path.role === SAMPLE_PATH_ROLES.sample) {
    return SAMPLE_ROLE_STYLE;
  }
  if (path.role === SAMPLE_PATH_ROLES.percentile) {
    const k = approxKey(path.percentile);
    const stroke = k !== undefined ? PERCENTILE_STROKE[k] : '#757575';
    return { stroke };
  }
  const byRole = ROLE_STROKE[path.role];
  if (byRole) {
    return { stroke: byRole };
  }
  return { stroke: '#9e9e9e' };
}

type LegendRow = {
  id: string;
  label: string;
  stroke: string;
  strokeOpacity?: number;
  dashed?: boolean;
};

function buildLegendRows(paths: SamplePath[]): LegendRow[] {
  const rows: LegendRow[] = [];
  let randomAdded = false;
  for (const p of paths) {
    if (p.role === SAMPLE_PATH_ROLES.sample) {
      if (!randomAdded) {
        rows.push({
          id: 'random',
          label: 'Random paths (grey, semi-transparent)',
          stroke: SAMPLE_ROLE_STYLE.stroke,
          strokeOpacity: SAMPLE_ROLE_STYLE.strokeOpacity,
          dashed: true,
        });
        randomAdded = true;
      }
      continue;
    }
    const id = p.role === SAMPLE_PATH_ROLES.percentile ? `pct-${p.percentile}` : p.role;
    if (rows.some(r => r.id === id)) {
      continue;
    }
    const style = lineStyleForPath(p);
    rows.push({
      id,
      label: p.label,
      stroke: style.stroke,
      strokeOpacity: style.strokeOpacity,
      dashed: Boolean(style.strokeDasharray),
    });
  }
  return rows;
}

type SeriesKey = {
  dataKey: string;
  label: string;
  path: SamplePath;
};

type ChartRow = Record<string, number | string | null>;

function buildSeries(paths: SamplePath[]): { data: ChartRow[]; keys: SeriesKey[] } {
  if (!paths.length) {
    return { data: [], keys: [] };
  }

  const keys: SeriesKey[] = paths.map((p, i) => ({
    dataKey: `p_${i}`,
    label: p.label || `Path ${i + 1}`,
    path: p,
  }));

  const len = Math.max(...paths.map(p => (Array.isArray(p.values) ? p.values.length : 0)), 0);
  const data: ChartRow[] = [];

  for (let t = 0; t < len; t++) {
    const row: ChartRow = { period: t };
    paths.forEach((p, i) => {
      const v = p.values?.[t];
      row[`p_${i}`] = typeof v === 'number' && Number.isFinite(v) ? v : null;
    });
    data.push(row);
  }

  return { data, keys };
}

type Props = {
  samplePaths: SamplePath[];
};

/**
 * Colors from role (and percentile for `percentile` role). Custom legend collapses all `sample` paths into one entry.
 */
const SamplePathsChart: React.FC<Props> = ({ samplePaths }) => {
  const { data, keys } = useMemo(() => buildSeries(samplePaths), [samplePaths]);
  const legendRows = useMemo(() => buildLegendRows(samplePaths), [samplePaths]);

  if (!keys.length) {
    return (
      <div style={{ padding: 24, color: '#666', textAlign: 'center', background: '#fafafa', borderRadius: 8 }}>
        No sample paths in this result.
      </div>
    );
  }

  return (
    <div style={{ width: '100%' }}>
      <div style={{ width: '100%', height: 360 }}>
        <ResponsiveContainer width="100%" height="100%">
          <LineChart data={data} margin={{ top: 8, right: 24, left: 12, bottom: 8 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="#e0e0e0" />
            <XAxis dataKey="period" tick={{ fontSize: 12 }} label={{ value: 'Period', position: 'insideBottom', offset: -4 }} />
            <YAxis
              width={56}
              tick={{ fontSize: 12 }}
              tickFormatter={(v: number) => formatUsdWhole(v)}
              domain={[(dataMin: number) => Math.min(100, dataMin), 'auto']}
              label={{
                value: 'Portfolio value (US$)',
                angle: -90,
                position: 'center',
                // Nudge left so rotated title clears tick labels (`offset` is ignored for this position).
                dx: -20,
              }}
            />
            <Tooltip
              formatter={(value: number) => (Number.isFinite(value) ? formatUsdWhole(value) : '—')}
              labelFormatter={(label: string | number) => `Period ${label}`}
            />
            {keys.map(k => {
              const style = lineStyleForPath(k.path);
              return (
                <Line
                  key={k.dataKey}
                  type="monotone"
                  dataKey={k.dataKey}
                  name={k.label}
                  stroke={style.stroke}
                  strokeOpacity={style.strokeOpacity}
                  strokeDasharray={style.strokeDasharray}
                  dot={false}
                  strokeWidth={2}
                  isAnimationActive={data.length < 400}
                  connectNulls
                />
              );
            })}
          </LineChart>
        </ResponsiveContainer>
      </div>
      <div
        style={{
          marginTop: 14,
          display: 'flex',
          flexWrap: 'wrap',
          gap: '10px 22px',
          alignItems: 'center',
          fontSize: 13,
          lineHeight: 1.35,
          color: '#424242',
        }}
      >
        {legendRows.map(row => (
          <span key={row.id} style={{ display: 'inline-flex', alignItems: 'center', gap: 8 }}>
            <span
              aria-hidden
              style={{
                width: 26,
                borderTopWidth: 3,
                borderTopStyle: row.dashed ? 'dashed' : 'solid',
                borderTopColor: row.stroke,
                opacity: row.strokeOpacity ?? 1,
                flexShrink: 0,
              }}
            />
            {row.label}
          </span>
        ))}
      </div>
    </div>
  );
};

export default SamplePathsChart;
