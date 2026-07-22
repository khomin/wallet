// ─── Layout (Protected Shell) ──────────────────────────────────────────────
// Sidebar + top bar that wraps all authenticated pages.

import { NavLink, Outlet } from 'react-router-dom';
import { useAuth } from '../auth/AuthContext';

// ─── Sidebar item definition ──────────────────────────────────────────────

interface NavItem {
  path: string;
  label: string;
  emoji: string;
}

const NAV_ITEMS: NavItem[] = [
  { path: '/dashboard', label: 'Dashboard', emoji: '📊' },
  { path: '/wallets', label: 'Wallets', emoji: '👛' },
  { path: '/alerts', label: 'Alerts', emoji: '🔔' },
  { path: '/market', label: 'Market', emoji: '📈' },
  { path: '/settings', label: 'Settings', emoji: '⚙️' },
];

// ─── Component ────────────────────────────────────────────────────────────

export default function Layout() {
  const { user, logout } = useAuth();

  const displayName =
    user?.name ?? user?.preferred_username ?? user?.email ?? 'Whale';

  return (
    <div className="flex h-screen bg-gray-950 text-white">
      {/* ── Sidebar ──────────────────────────────────────────────────── */}
      <aside className="w-60 shrink-0 border-r border-white/5 flex flex-col">
        {/* Logo */}
        <div className="flex items-center gap-2 px-5 py-5 border-b border-white/5">
          <span className="text-2xl">🐋</span>
          <span className="text-lg font-semibold tracking-tight">WhaleTracker</span>
        </div>

        {/* Nav items */}
        <nav className="flex-1 px-3 py-4 space-y-1">
          {NAV_ITEMS.map((item) => (
            <NavLink
              key={item.path}
              to={item.path}
              className={({ isActive }) =>
                `flex items-center gap-3 rounded-lg px-3 py-2.5 text-sm transition-colors ${isActive
                  ? 'bg-purple-600/20 text-purple-300 font-medium'
                  : 'text-gray-400 hover:bg-white/[0.04] hover:text-gray-200'
                }`
              }
            >
              <span className="text-lg">{item.emoji}</span>
              <span>{item.label}</span>
            </NavLink>
          ))}
        </nav>

        {/* Footer */}
        <div className="px-3 py-4 border-t border-white/5">
          <div className="flex items-center gap-2 rounded-lg px-3 py-2 text-xs text-gray-500">
            <span>🏷️</span>
            <span>v0.1.0</span>
          </div>
        </div>
      </aside>

      {/* ── Main content area ────────────────────────────────────────── */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* Top bar */}
        <header className="flex items-center justify-end px-8 py-4 border-b border-white/5 shrink-0">
          <div className="flex items-center gap-3">
            <div className="flex items-center gap-2 rounded-full border border-white/10 bg-white/5 px-3 py-1.5">
              <div className="h-6 w-6 rounded-full bg-purple-600 flex items-center justify-center text-xs font-bold">
                {displayName.charAt(0).toUpperCase()}
              </div>
              <span className="text-sm text-gray-300">{displayName}</span>
            </div>

            <button
              onClick={logout}
              className="rounded-lg border border-white/10 px-3 py-1.5 text-sm text-gray-400
                         hover:border-red-500/40 hover:text-red-400 transition-colors cursor-pointer"
            >
              Log out
            </button>
          </div>
        </header>

        {/* Page content injected by router */}
        <main className="flex-1 overflow-y-auto p-8">
          <Outlet />
        </main>
      </div>
    </div>
  );
}