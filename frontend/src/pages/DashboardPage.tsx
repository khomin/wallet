// ─── Dashboard Page (Protected) ──────────────────────────────────────────────
// Only reachable when the user is authenticated (enforced by ProtectedRoute).
// This is the main wallet portfolio view – shell / skeleton for now.

import { useAuth } from '../auth/AuthContext';

export default function DashboardPage() {
  const { user, logout, accessToken } = useAuth();

  const displayName =
    user?.name ?? user?.preferred_username ?? user?.email ?? 'Whale';

  return (
    <div className="min-h-screen bg-gray-950 text-white flex flex-col">
      {/* ── Top nav ─────────────────────────────────────────────────────── */}
      <header className="flex items-center justify-between px-8 py-5 border-b border-white/5">
        <div className="flex items-center gap-2">
          <span className="text-2xl">🐋</span>
          <span className="text-lg font-semibold tracking-tight">WhaleTracker</span>
        </div>

        <div className="flex items-center gap-4">
          {/* User badge */}
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

      {/* ── Main content ────────────────────────────────────────────────── */}
      <main className="flex-1 p-8 max-w-6xl mx-auto w-full">

        {/* Stats row – skeleton placeholders */}
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-8">
          {STAT_CARDS.map((card) => (
            <div
              key={card.label}
              className="rounded-xl border border-white/5 bg-white/[0.03] p-5"
            >
              <p className="text-xs text-gray-500 uppercase tracking-wider mb-1">
                {card.label}
              </p>
              <p className="text-2xl font-semibold text-white">{card.value}</p>
              <p className="text-xs text-gray-600 mt-1">{card.sub}</p>
            </div>
          ))}
        </div>

        {/* Wallets table – placeholder */}
        <div className="rounded-xl border border-white/5 bg-white/[0.03] p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-sm font-semibold text-white">Wallets</h2>
            <button className="rounded-lg bg-purple-600 px-3 py-1.5 text-xs font-medium hover:bg-purple-500 transition-colors cursor-pointer">
              + Add wallet
            </button>
          </div>

          {/* Empty state */}
          <div className="flex flex-col items-center justify-center py-16 text-center">
            <span className="text-4xl mb-4">🐋</span>
            <p className="text-gray-400 text-sm font-medium">No wallets yet</p>
            <p className="text-gray-600 text-xs mt-1">
              Add a wallet address to start tracking your portfolio.
            </p>
          </div>
        </div>

        {/* Dev helper: show token for testing API calls */}
        {import.meta.env.DEV && accessToken && (
          <details className="mt-8 rounded-xl border border-yellow-500/20 bg-yellow-500/5 p-4">
            <summary className="text-xs text-yellow-500/70 cursor-pointer select-none">
              DEV – Access Token (click to expand)
            </summary>
            <pre className="mt-2 text-[10px] text-yellow-400/50 break-all whitespace-pre-wrap">
              {accessToken}
            </pre>
          </details>
        )}
      </main>
    </div>
  );
}

const STAT_CARDS = [
  { label: 'Total Balance', value: '$—', sub: 'Across all chains' },
  { label: '24h Change', value: '— %', sub: 'vs yesterday' },
  { label: 'Wallets', value: '0', sub: 'Active addresses' },
];
