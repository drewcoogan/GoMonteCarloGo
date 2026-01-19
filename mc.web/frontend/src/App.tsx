import React, { useState } from 'react';
import './App.css';
import ScenarioPage from './ScenarioPage';
import SyncDataPage from './SyncDataPage';

type Tab = 'sync' | 'scenarios';

function App() {
  const [activeTab, setActiveTab] = useState<Tab>('sync');

  return (
    <div style={{ minHeight: '100vh', background: '#f5f5f5' }}>
      <div style={{ 
        background: '#fff', 
        boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
        marginBottom: 20
      }}>
        <div style={{ 
          maxWidth: 1200, 
          margin: '0 auto', 
          display: 'flex', 
          gap: 0,
          padding: '0 20px'
        }}>
          <button
            onClick={() => setActiveTab('sync')}
            style={{
              padding: '16px 24px',
              fontSize: 16,
              fontWeight: 'bold',
              background: activeTab === 'sync' ? '#1976d2' : 'transparent',
              color: activeTab === 'sync' ? '#fff' : '#666',
              border: 'none',
              borderBottom: activeTab === 'sync' ? '3px solid #1976d2' : '3px solid transparent',
              cursor: 'pointer',
              transition: 'all 0.2s'
            }}
          >
            Sync Data
          </button>
          <button
            onClick={() => setActiveTab('scenarios')}
            style={{
              padding: '16px 24px',
              fontSize: 16,
              fontWeight: 'bold',
              background: activeTab === 'scenarios' ? '#1976d2' : 'transparent',
              color: activeTab === 'scenarios' ? '#fff' : '#666',
              border: 'none',
              borderBottom: activeTab === 'scenarios' ? '3px solid #1976d2' : '3px solid transparent',
              cursor: 'pointer',
              transition: 'all 0.2s'
            }}
          >
            Scenarios
          </button>
        </div>
      </div>
      <div>
        {activeTab === 'sync' && <SyncDataPage />}
        {activeTab === 'scenarios' && <ScenarioPage />}
      </div>
    </div>
  );
}

export default App;
