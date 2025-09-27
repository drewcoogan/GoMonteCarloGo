import React, { useState } from 'react';

const AddMePage: React.FC = () => {
  const [num1, setNum1] = useState('');
  const [num2, setNum2] = useState('');
  const [loading, setLoading] = useState(false);
  const [result, setResult] = useState<string | null>(null);

  const handleAdd = async () => {
    setLoading(true);
    setResult(null);
    // Placeholder for API request to mc.service
    // Replace with actual endpoint when available
    try {
      // Example: const response = await fetch('/api/add', { ... })
      // Simulate API response
      setTimeout(() => {
        setResult(`Result: ${Number(num1) + Number(num2)}`);
        setLoading(false);
      }, 500);
    } catch (error) {
      setResult('Error occurred');
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
          onClick={handleAdd}
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

export default AddMePage;
