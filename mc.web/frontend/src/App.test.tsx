import React from 'react';
import { render, screen } from '@testing-library/react';
import App from './App';

test('renders main navigation tabs', () => {
  render(<App />);
  const syncTab = screen.getByText(/sync data/i);
  const scenariosTab = screen.getByText(/scenarios/i);
  expect(syncTab).toBeInTheDocument();
  expect(scenariosTab).toBeInTheDocument();
});
