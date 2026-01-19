import React from 'react';
import { render, screen } from '@testing-library/react';
import App from './App';

test('renders main navigation tabs', () => {
  render(<App />);
  // "Sync Data" appears both as a tab label and as a page button; ensure we assert on the tab.
  const syncTab = screen.getAllByRole('button', { name: /sync data/i })[0];
  const scenariosTab = screen.getByRole('button', { name: /scenarios/i });
  expect(syncTab).toBeInTheDocument();
  expect(scenariosTab).toBeInTheDocument();
});
