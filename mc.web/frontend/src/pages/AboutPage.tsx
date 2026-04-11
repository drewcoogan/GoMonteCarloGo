import React, { useEffect, useState } from 'react';
import { getHeartbeatPayload } from '../controllers/heartbeat';

const GITHUB_REPO = 'https://github.com/drewcoogan/GoMonteCarloGo';

const AboutPage: React.FC = () => {
  const [goRuntimeVersion, setGoRuntimeVersion] = useState<string | null>(null);
  const [goVersionError, setGoVersionError] = useState(false);

  useEffect(() => {
    let cancelled = false;
    (async () => {
      try {
        const hb = await getHeartbeatPayload();
        if (!cancelled) {
          setGoRuntimeVersion(hb.goVersion || null);
          setGoVersionError(false);
        }
      } catch {
        if (!cancelled) {
          setGoRuntimeVersion(null);
          setGoVersionError(true);
        }
      }
    })();
    return () => {
      cancelled = true;
    };
  }, []);

  return (
    <div style={{ maxWidth: 720, margin: '0 auto', padding: '24px 20px 48px' }}>
      <h1 style={{ textAlign: 'center', marginBottom: 12 }}>About</h1>
      <p style={{ textAlign: 'center', color: '#555', marginBottom: 28 }}>
        GoMonteCarloGo (GMCG) — Monte Carlo portfolio simulation with synced market data.
      </p>

      <div
        style={{
          padding: 24,
          background: '#fff',
          borderRadius: 8,
          boxShadow: '0 2px 8px #eee',
          marginBottom: 24,
        }}
      >
        <h2 style={{ fontSize: 18, marginTop: 0, marginBottom: 12 }}>Source code</h2>
        <p style={{ margin: 0, lineHeight: 1.6 }}>
          <a
            href={GITHUB_REPO}
            target="_blank"
            rel="noopener noreferrer"
            style={{ color: '#1976d2', fontWeight: 600 }}
          >
            {GITHUB_REPO}
          </a>
        </p>
      </div>

      <div
        style={{
          padding: 24,
          background: '#fff',
          borderRadius: 8,
          boxShadow: '0 2px 8px #eee',
        }}
      >
        <h2 style={{ fontSize: 18, marginTop: 0, marginBottom: 16 }}>Tech stack</h2>
        <p style={{ fontSize: 14, color: '#666', marginBottom: 16 }}>
          Frontend versions follow <code>package.json</code>. The Go version below is reported live by the API process (
          <code>runtime.Version()</code>); <code>go.mod</code> sets the minimum language version for{' '}
          <code>mc.service</code> and <code>mc.data</code> (same toolchain when built together).
        </p>

        <h3 style={{ fontSize: 15, marginBottom: 8 }}>Backend</h3>
        <ul style={{ margin: '0 0 20px 0', paddingLeft: 22, lineHeight: 1.7 }}>
          <li>
            <strong>Go (runtime)</strong>{' '}
            {goRuntimeVersion ? (
              <code>{goRuntimeVersion}</code>
            ) : goVersionError ? (
              <span style={{ color: '#c62828' }}>(could not reach API — start mc.service to load)</span>
            ) : (
              <span style={{ color: '#888' }}>…</span>
            )}
          </li>
          <li>
            <strong>chi</strong> v5 — HTTP routing (<code>github.com/go-chi/chi/v5</code>)
          </li>
          <li>
            <strong>pgx/v5</strong> — PostgreSQL driver
          </li>
          <li>
            <strong>gonum</strong> — numerical work in simulations
          </li>
          <li>REST API (JSON), CORS for local dev</li>
        </ul>

        <h3 style={{ fontSize: 15, marginBottom: 8 }}>Data &amp; persistence</h3>
        <ul style={{ margin: '0 0 20px 0', paddingLeft: 22, lineHeight: 1.7 }}>
          <li>PostgreSQL — time series and scenario metadata</li>
          <li>Embedded SQL via <code>go:embed</code> in <code>mc.data</code></li>
        </ul>

        <h3 style={{ fontSize: 15, marginBottom: 8 }}>Frontend</h3>
        <ul style={{ margin: '0 0 20px 0', paddingLeft: 22, lineHeight: 1.7 }}>
          <li>
            <strong>React</strong> {React.version} (see <code>package.json</code> for pinned deps)
          </li>
          <li>
            <strong>TypeScript</strong> 4.x
          </li>
          <li>
            <strong>react-router-dom</strong> — client routing
          </li>
          <li>
            <strong>Recharts</strong> — charts
          </li>
          <li>
            <strong>simple-statistics</strong> — correlation helpers on Quick look
          </li>
          <li>Create React App (<code>react-scripts</code>)</li>
        </ul>

        <h3 style={{ fontSize: 15, marginBottom: 8 }}>External APIs</h3>
        <ul style={{ margin: 0, paddingLeft: 22, lineHeight: 1.7 }}>
          <li>Alpha Vantage — market data sync (API key via env)</li>
        </ul>
      </div>
    </div>
  );
};

export default AboutPage;
