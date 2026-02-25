import React from 'react';
import { render, screen } from '@testing-library/react';
import App from './App';

test('renders main navigation links', () => {
  render(<App />);
  const syncLink = screen.getByRole('link', { name: /sync data/i });
  const scenariosLink = screen.getByRole('link', { name: /scenarios/i });
  expect(syncLink).toBeInTheDocument();
  expect(scenariosLink).toBeInTheDocument();
});
