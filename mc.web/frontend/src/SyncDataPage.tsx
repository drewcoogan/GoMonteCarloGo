import React, { useState } from 'react';

const API_BASE = '';

const SyncDataPage: React.FC = () => {
  const [symbol, setSymbol] = useState('');
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<string | null>(null);
  const [error, setError] = useState<string | null>(null);

  const handleSync = async () => {
    setLoading(true);
    setResult(null);
    setError(null);
    
    try {
      const response = await fetch(`${API_BASE}/api/syncStockData`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ symbol }),
      });

      const json = await response.json();
      
      if (!response.ok) {
        setError(json.error || json.message || 'An error occurred');
        setLoading(false);
        return;
      }

      if (json.date) {
        const date = new Date(json.date);
        setResult(`Success! Last refreshed: ${date.toLocaleString()}`);
      } else {
        setError('Unexpected response format');
      }
      
      setLoading(false);
    } catch (err: any) {
      setError(`Error occurred: ${err.message}`);
      setLoading(false);
    }
  };

  return (
    <div style={{ maxWidth: 500, margin: '40px auto', padding: 24, boxShadow: '0 2px 8px #eee', borderRadius: 8 }}>
      <h1 style={{ textAlign: 'center' }}>Sync Stock Data</h1>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
        <input
          type="text"
          placeholder="Enter stock symbol (e.g., AAPL)"
          value={symbol}
          onChange={e => setSymbol(e.target.value.toUpperCase())}
          style={{ padding: 8, fontSize: 16 }}
          onKeyPress={e => {
            if (e.key === 'Enter' && !loading && symbol) {
              handleSync();
            }
          }}
        />
        <button
          onClick={handleSync}
          disabled={loading || !symbol}
          style={{ 
            padding: 10, 
            fontSize: 16, 
            background: '#1976d2', 
            color: '#fff', 
            border: 'none', 
            borderRadius: 4,
            cursor: loading || !symbol ? 'not-allowed' : 'pointer',
            opacity: loading || !symbol ? 0.6 : 1
          }}
        >
          {loading ? 'Syncing...' : 'Sync Data'}
        </button>
        {result && (
          <div style={{ 
            marginTop: 16, 
            padding: 12,
            background: '#e8f5e9', 
            color: '#2e7d32',
            borderRadius: 4,
            fontWeight: 'bold' 
          }}>
            {result}
          </div>
        )}
        {error && (
          <div style={{ 
            marginTop: 16, 
            padding: 12,
            background: '#ffebee', 
            color: '#c62828',
            borderRadius: 4,
            fontWeight: 'bold' 
          }}>
            {error}
          </div>
        )}
      </div>
    </div>
  );
};

export default SyncDataPage;

