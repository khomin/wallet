// ─── Layout (Protected Shell) ──────────────────────────────────────────────
// Sidebar + top bar that wraps all authenticated pages.

import { NavLink, Outlet } from 'react-router-dom';
import { useState, useEffect } from 'react';
import {
  Bell,
  ChartNoAxesCombined,
  LayoutDashboard,
  Menu,
  Settings2,
  Tag,
  WalletCards,
  X,
  type LucideIcon,
} from 'lucide-react';
import { useAuth } from '../auth/AuthContext';

// ─── Sidebar item definition ──────────────────────────────────────────────

interface NavItem {
  path: string;
  label: string;
  icon: LucideIcon;
}

const NAV_ITEMS: NavItem[] = [
  { path: '/dashboard', label: 'Overview', icon: LayoutDashboard },
  { path: '/wallets', label: 'Wallets', icon: WalletCards },
  { path: '/alerts', label: 'Alerts', icon: Bell },
  { path: '/market', label: 'Market', icon: ChartNoAxesCombined },
  { path: '/settings', label: 'Settings', icon: Settings2 },
];

// ─── Component ────────────────────────────────────────────────────────────

export default function Layout() {
  const { user, logout } = useAuth();
  const [sidebarOpen, setSidebarOpen] = useState(false);
  const [isMobile, setIsMobile] = useState(false);

  // Detect mobile viewport
  useEffect(() => {
    const checkMobile = () => setIsMobile(window.innerWidth < 1024);
    checkMobile();
    window.addEventListener('resize', checkMobile);
    return () => window.removeEventListener('resize', checkMobile);
  }, []);

  // Close sidebar on navigation (mobile)
  const handleNavClick = () => {
    if (isMobile) setSidebarOpen(false);
  };

  const displayName =
    user?.name ?? user?.preferred_username ?? user?.email ?? 'Whale';

  return (
    <div className="flex h-screen bg-[#080a12] text-white">
      {/* ── Mobile Overlay ─────────────────────────────────────────────── */}
      {isMobile && sidebarOpen && (
        <div
          className="fixed inset-0 z-40 bg-black/50 backdrop-blur-sm"
          onClick={() => setSidebarOpen(false)}
          aria-hidden="true"
        />
      )}

      {/* ── Sidebar ──────────────────────────────────────────────────── */}
      <aside
        className={`
          flex w-64 shrink-0 flex-col border-r border-white/[0.07] bg-[#0d101a]
          z-50 transition-transform duration-300 ease-in-out
          ${isMobile
            ? 'fixed inset-y-0 left-0 transform lg:translate-x-0'
            : 'translate-x-0'
          }
          ${isMobile && sidebarOpen ? 'translate-x-0' : '-translate-x-full'}
        `}
        aria-label="Main navigation"
      >
        {/* Logo */}
        <div className="flex h-[82px] items-center gap-3 border-b border-white/[0.07] px-6">
          <div className="flex h-9 w-9 items-center justify-center rounded-xl bg-gradient-to-br from-violet-500 to-indigo-600 shadow-lg shadow-violet-950/40">
            <span className="text-lg font-black tracking-tight">W</span>
          </div>
          <div>
            <div className="text-[15px] font-semibold tracking-tight text-white">WhaleTracker</div>
            <div className="mt-0.5 text-[11px] font-medium text-slate-400 truncate">{displayName}</div>
          </div>
        </div>

        {/* Nav items */}
        <nav aria-label="Primary navigation" className="flex-1 px-3 py-7">
          <p className="mb-3 px-3 text-[10px] font-semibold uppercase tracking-[0.18em] text-slate-600">Workspace</p>
          <div className="space-y-1">
            {NAV_ITEMS.map((item) => {
              const Icon = item.icon;

              return (
                <NavLink
                  key={item.path}
                  to={item.path}
                  draggable={false}
                  onClick={handleNavClick}
                  className={({ isActive }) =>
                    `group relative flex select-none items-center gap-3 rounded-xl px-3 py-3 text-[13px] font-medium transition-[background-color,color] duration-200 ${isActive
                      ? 'bg-white/[0.07] text-white'
                      : 'text-slate-400 hover:bg-white/[0.045] hover:text-slate-100'
                    }`
                  }
                >
                  {({ isActive }) => (
                    <>
                      <Icon aria-hidden="true" size={18} strokeWidth={1.8} className="shrink-0 transition-colors" />
                      <span>{item.label}</span>
                      <span
                        aria-hidden="true"
                        className={`ml-auto h-1.5 w-1.5 rounded-full transition-opacity ${isActive ? 'bg-slate-400 opacity-100' : 'bg-slate-600 opacity-0 group-hover:opacity-60'}`}
                      />
                    </>
                  )}
                </NavLink>
              );
            })}
          </div>
        </nav>

        {/* Footer */}
        <div className="border-t border-white/[0.07] px-4 py-5">
          <div className="flex items-center gap-2 px-2 text-[11px] text-slate-600">
            <Tag aria-hidden="true" size={13} />
            <span>WhaleTracker v0.1.0</span>
          </div>
        </div>
      </aside>

      {/* ── Main content area ────────────────────────────────────────── */}
      <div className="flex min-w-0 flex-1 flex-col">
        {/* Top bar - compact, just mobile menu button */}
        <header className="flex shrink-0 items-center border-b border-white/[0.07] px-4 py-2.5 lg:p-0">
          {/* Mobile menu button */}
          <button
            className="lg:hidden inline-flex items-center justify-center p-2 rounded-lg text-gray-400 hover:text-white hover:bg-white/5 transition-colors cursor-pointer"
            onClick={() => setSidebarOpen(true)}
            aria-label="Open navigation menu"
            aria-expanded={sidebarOpen}
            aria-controls="sidebar"
          >
            {sidebarOpen ? (
              <X size={20} aria-hidden="true" />
            ) : (
              <Menu size={20} aria-hidden="true" />
            )}
          </button>
        </header>

        {/* Page content injected by router */}
        <main className="flex-1 overflow-y-auto p-4 sm:p-6 lg:p-8">
          <Outlet />
        </main>
      </div>
    </div>
  );
}