import React from 'react';
import { NavLink, Outlet } from 'react-router-dom';
import ServiceHealthIndicator from './ServiceHealthIndicator';

const navLinkStyle = ({ isActive }: { isActive: boolean }) => ({
  padding: '16px 24px',
  fontSize: 16,
  fontWeight: 'bold' as const,
  background: isActive ? '#1976d2' : 'transparent',
  color: isActive ? '#fff' : '#666',
  border: 'none',
  borderBottom: isActive ? '3px solid #1976d2' : '3px solid transparent',
  cursor: 'pointer',
  transition: 'all 0.2s',
  textDecoration: 'none',
});

const AppLayout: React.FC = () => {
  return (
    <div style={{ minHeight: '100vh', background: '#f5f5f5' }}>
      <header
        style={{
          background: '#fff',
          boxShadow: '0 2px 4px rgba(0,0,0,0.1)',
          marginBottom: 20,
        }}
      >
        <nav
          style={{
            maxWidth: 1200,
            margin: '0 auto',
            display: 'flex',
            alignItems: 'center',
            flexWrap: 'wrap',
            gap: 0,
            padding: '0 20px',
          }}
        >
          <div
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: 10,
              marginRight: 16,
              padding: '12px 0',
            }}
          >
            <span
              style={{
                fontSize: 20,
                fontWeight: 800,
                letterSpacing: '0.06em',
                color: '#111',
              }}
            >
              GMCG
            </span>
            <ServiceHealthIndicator />
          </div>
          <div style={{ display: 'flex', flex: 1, minWidth: 0 }}>
          <NavLink to="/" end style={navLinkStyle}>
            Sync Data
          </NavLink>
          <NavLink to="/scenarios" style={navLinkStyle}>
            Scenarios
          </NavLink>
          <NavLink to="/simulate" style={navLinkStyle}>
            Simulate
          </NavLink>
          </div>
        </nav>
      </header>
      <main>
        <Outlet />
      </main>
    </div>
  );
};

export default AppLayout;
