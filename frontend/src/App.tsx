// ─── App – Route declarations ───────────────────────────────────────────────────────────
//
//  /             → LandingPage     (public)
//  /callback     → CallbackPage    (handles Keycloak redirect, then pushes to /dashboard)
//  /dashboard    → DashboardPage   (protected – portfolio summary)
//  /wallets      → WalletsPage     (protected – full wallet management)
//  /alerts       → AlertsPage      (protected – placeholder)
//  /market       → MarketPage      (protected – live coin prices)
//  /settings     → SettingsPage    (protected – preferences & profile)
//

import { BrowserRouter, Route, Routes, Navigate } from 'react-router-dom';
import LandingPage from './pages/LandingPage';
import CallbackPage from './pages/CallbackPage';
import DashboardPage from './pages/DashboardPage';
import WalletsPage from './pages/WalletsPage';
import AlertsPage from './pages/AlertsPage';
import MarketPage from './pages/MarketPage';
import SettingsPage from './pages/SettingsPage';
import ProtectedRoute from './components/ProtectedRoute';
import Layout from './components/Layout';

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        {/* Public – marketing / login entry point */}
        <Route path="/" element={<LandingPage />} />

        {/* OAuth2 redirect target – exchanges code for tokens */}
        <Route path="/callback" element={<CallbackPage />} />

        {/* Protected – all app pages share the sidebar layout */}
        <Route
          element={
            <ProtectedRoute>
              <Layout />
            </ProtectedRoute>
          }
        >
          <Route path="/dashboard" element={<DashboardPage />} />
          <Route path="/wallets" element={<WalletsPage />} />
          <Route path="/alerts" element={<AlertsPage />} />
          <Route path="/market" element={<MarketPage />} />
          <Route path="/settings" element={<SettingsPage />} />
        </Route>

        {/* Catch-all: redirect unknown paths to landing */}
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
}
