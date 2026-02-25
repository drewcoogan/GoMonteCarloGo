import React from 'react';
import { NavLink, Outlet } from 'react-router-dom';

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
            gap: 0,
            padding: '0 20px',
          }}
        >
          <NavLink to="/" end style={navLinkStyle}>
            Sync Data
          </NavLink>
          <NavLink to="/scenarios" style={navLinkStyle}>
            Scenarios
          </NavLink>
        </nav>
      </header>
      <main>
        <Outlet />
      </main>
    </div>
  );
};

export default AppLayout;
