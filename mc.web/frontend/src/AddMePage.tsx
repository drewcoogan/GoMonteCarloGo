import React, { useState } from 'react';

const API_BASE = 'http://localhost:8080';

const AddMePage: React.FC = () => {
  const [num1, setNum1] = useState('');
  const [num2, setNum2] = useState('');
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<string | null>(null);

  const handleAddByGet = async () => {
    setLoading(true);
    setResult(null);
    try {
      const params = new URLSearchParams();
      params.append('number1', num1);
      params.append('number2', num2);
      setTimeout(async () => {
        const response = await fetch(`${API_BASE}/api/addByGet?${params.toString()}`);
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
        const response = await fetch(`${API_BASE}/api/addByPost`, {
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
