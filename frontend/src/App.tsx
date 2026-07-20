// ─── App – Route declarations ───────────────────────────────────────────────────────────
//
//  /             → LandingPage   (public)
//  /callback     → CallbackPage  (handles Keycloak redirect, then pushes to /dashboard)
//  /dashboard    → DashboardPage (protected – requires valid token)
//

import { BrowserRouter, Route, Routes, Navigate } from 'react-router-dom';
import LandingPage from './pages/LandingPage';
import CallbackPage from './pages/CallbackPage';
import DashboardPage from './pages/DashboardPage';
import ProtectedRoute from './components/ProtectedRoute';

export default function App() {
  return (
    <BrowserRouter>
      <Routes>
        {/* Public – marketing / login entry point */}
        <Route path="/" element={<LandingPage />} />

        {/* OAuth2 redirect target – exchanges code for tokens */}
        <Route path="/callback" element={<CallbackPage />} />

        {/* Protected – only accessible when authenticated */}
        <Route
          path="/dashboard"
          element={
            <ProtectedRoute>
              <DashboardPage />
            </ProtectedRoute>
          }
        />

        {/* Catch-all: redirect unknown paths to landing */}
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
}
