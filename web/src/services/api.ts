// ─── Axios API Client ──────────────────────────────────────────────────────────
// Pre-configured axios instance that:
//  1. Sets the base URL from Vite env (defaults to http://localhost:8080)
//  2. Attaches the Keycloak Bearer token on every request
//  3. Centralises error handling

import axios from 'axios';
import { API_CONFIG } from '../config/api';

const api = axios.create({
  baseURL: API_CONFIG.baseUrl,
  headers: { 'Content-Type': 'application/json' },
});

// ── Request interceptor – attach Bearer token if present ────────────────────
api.interceptors.request.use((config) => {
  const token = sessionStorage.getItem('kc_access_token');
  if (token) {
    config.headers.Authorization = `Bearer ${token}`;
  }
  return config;
});

// ── Response interceptor – normalise errors ─────────────────────────────────
api.interceptors.response.use(
  (response) => response,
  (error) => {
    if (error.response?.status === 401) {
      // Token expired or invalid – clear session and redirect to landing
      sessionStorage.clear();
      window.location.href = '/';
    }
    return Promise.reject(error);
  },
);

export default api;