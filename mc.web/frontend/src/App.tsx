import React from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import './App.css';
import AppLayout from './components';
import { SyncDataPage, QuickLookPage, ScenarioPage, MakeSimulationPage, AboutPage } from './pages';

function App() {
  return (
    <BrowserRouter>
      <Routes>
        <Route path="/" element={<AppLayout />}>
          <Route index element={<SyncDataPage />} />
          <Route path="quick-look" element={<QuickLookPage />} />
          <Route path="scenarios" element={<ScenarioPage />} />
          <Route path="simulate" element={<MakeSimulationPage />} />
          <Route path="about" element={<AboutPage />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Route>
      </Routes>
    </BrowserRouter>
  );
}

export default App;
