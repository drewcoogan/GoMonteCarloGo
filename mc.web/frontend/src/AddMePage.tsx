import React, { useState } from 'react';

const API_BASE = 'http://localhost:8080';

const AddMePage: React.FC = () => {
  const [num1, setNum1] = useState('');
  const [num2, setNum2] = useState('');
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<string | null>(null);
  const [pingStatus, setPingStatus] = useState<'idle' | 'loading' | 'success' | 'error'>('idle');

  const handleAddByGet = async () => {
    setLoading(true);
    setResult(null);
    try {
      const params = new URLSearchParams();
      params.append('number1', num1);
      params.append('number2', num2);
      setTimeout(async () => {
        const response = await fetch(`${API_BASE}/api/test/addByGet?${params.toString()}`);
        console.log('Response:', response);
        if (!response.ok) {
          throw new Error(`Response status: ${response.status}`);
        }

        const json = await response.json();
        console.log('Response json:', JSON.stringify(json));
        setResult(`Result: ${json.result}`);
        setLoading(false);
      }, 500);
    } catch (error: any) {
      setResult(`Error occurred: ${error.message}`);
      setLoading(false);
    }
  };

  const handleAddByPost = async () => {
    setLoading(true);
    setResult(null);
    try {
      const numbers = new NumbersToSum(num1, num2);
      setTimeout(async () => {
        const response = await fetch(`${API_BASE}/api/test/addByPost`, {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
          },
          body: JSON.stringify(numbers),
        });
        if (!response.ok) {
          throw new Error(`Response status: ${response.status}`);
        }
        const json = await response.json();
        console.log('Response json:', JSON.stringify(json));
        setResult(`Result: ${json.result}`);
        setLoading(false);
      }, 500);
    }
    catch (error: any) {
      setResult(`Error occurred: ${error.message}`);
      setLoading(false);
    }
  };

  const handlePing = async () => {
    setPingStatus('loading');
    try {
      const response = await fetch(`${API_BASE}/api/ping`);
      if (response.ok) {
        setPingStatus('success');
      } else {
        setPingStatus('error');
      }
    } catch (error) {
      setPingStatus('error');
    }
  };

  return (
    <div style={{ maxWidth: 400, margin: '40px auto', padding: 24, boxShadow: '0 2px 8px #eee', borderRadius: 8 }}>
      <h1 style={{ textAlign: 'center' }}>Add me</h1>
      <div style={{ display: 'flex', flexDirection: 'column', gap: 16 }}>
        <input
          type="number"
          placeholder="Enter first number"
          value={num1}
          onChange={e => setNum1(e.target.value)}
          style={{ padding: 8, fontSize: 16 }}
        />
        <input
          type="number"
          placeholder="Enter second number"
          value={num2}
          onChange={e => setNum2(e.target.value)}
          style={{ padding: 8, fontSize: 16 }}
        />
        <button
          onClick={handleAddByPost}
          disabled={loading || !num1 || !num2}
          style={{ padding: 10, fontSize: 16, background: '#1976d2', color: '#fff', border: 'none', borderRadius: 4 }}
        >
          {loading ? 'Adding...' : 'Add'}
        </button>
        {result && <div style={{ marginTop: 16, fontWeight: 'bold', textAlign: 'center' }}>{result}</div>}
        
        <div style={{ marginTop: 24, paddingTop: 24, borderTop: '1px solid #eee' }}>
          <h2 style={{ textAlign: 'center', fontSize: 18, marginBottom: 16 }}>API Status</h2>
          <button
            onClick={handlePing}
            disabled={pingStatus === 'loading'}
            style={{ 
              padding: 10, 
              fontSize: 16, 
              background: '#4caf50', 
              color: '#fff', 
              border: 'none', 
              borderRadius: 4,
              width: '100%',
              cursor: pingStatus === 'loading' ? 'not-allowed' : 'pointer',
              opacity: pingStatus === 'loading' ? 0.6 : 1
            }}
          >
            {pingStatus === 'loading' ? 'Pinging...' : 'Ping API'}
          </button>
          {pingStatus === 'success' && (
            <div style={{ 
              marginTop: 16, 
              padding: 12,
              background: '#4caf50', 
              color: '#fff',
              borderRadius: 4,
              textAlign: 'center',
              fontWeight: 'bold'
            }}>
              ✓ API is responding
            </div>
          )}
          {pingStatus === 'error' && (
            <div style={{ 
              marginTop: 16, 
              padding: 12,
              background: '#f44336', 
              color: '#fff',
              borderRadius: 4,
              textAlign: 'center',
              fontWeight: 'bold'
            }}>
              ✗ API is not responding
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

class NumbersToSum {
  constructor(number1: any, number2: any) {
    this.number1 = Number(number1);
    this.number2 = Number(number2);
  }

  number1;
  number2;
}

export default AddMePage;
