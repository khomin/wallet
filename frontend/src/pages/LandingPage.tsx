// ─── Landing Page ─────────────────────────────────────────────────────────────
// Public page. Shown when the user is not authenticated.
// The "Get Started" / "Log In" CTA kicks off the PKCE login flow.

import { useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { useAuth } from '../auth/AuthContext';

export default function LandingPage() {
  const { login, isAuthenticated, isInitialized } = useAuth();
  const navigate = useNavigate();

  // If already logged in, skip the landing page entirely
  useEffect(() => {
    if (isInitialized && isAuthenticated) {
      navigate('/dashboard', { replace: true });
    }
  }, [isInitialized, isAuthenticated, navigate]);

  return (
    <div className="min-h-screen bg-gray-950 text-white flex flex-col">
      {/* ── Nav ─────────────────────────────────────────────────────────── */}
      <header className="flex items-center justify-between px-8 py-5 border-b border-white/5">
        <div className="flex items-center gap-2">
          <span className="text-2xl">🐋</span>
          <span className="text-lg font-semibold tracking-tight">WhaleTracker</span>
        </div>
        <button
          onClick={login}
          className="rounded-lg border border-purple-500/50 px-4 py-1.5 text-sm font-medium text-purple-300
                     hover:bg-purple-500/10 transition-colors cursor-pointer"
        >
          Log in
        </button>
      </header>

      {/* ── Hero ────────────────────────────────────────────────────────── */}
      <main className="flex flex-1 flex-col items-center justify-center px-6 text-center gap-8">
        {/* Badge */}
        <span
          className="inline-flex items-center gap-1.5 rounded-full border border-purple-500/30
                       bg-purple-500/10 px-3 py-1 text-xs font-medium text-purple-300"
        >
          <span className="h-1.5 w-1.5 rounded-full bg-purple-400 animate-pulse" />
          Multi-chain portfolio tracking
        </span>

        {/* Headline */}
        <h1
          className="max-w-2xl text-5xl font-bold tracking-tight leading-tight
                       bg-gradient-to-br from-white via-gray-200 to-gray-500 bg-clip-text text-transparent"
        >
          Track every wallet.
          <br />
          Across every chain.
        </h1>

        {/* Sub-headline */}
        <p className="max-w-md text-gray-400 text-lg leading-relaxed">
          WhaleTracker gives you a single dashboard for Bitcoin, Ethereum,
          Solana, Tron and more — with live prices and portfolio history.
        </p>

        {/* CTA */}
        <button
          onClick={login}
          className="mt-2 inline-flex items-center gap-2 rounded-xl bg-purple-600 px-8 py-3.5
                     text-base font-semibold shadow-lg shadow-purple-900/40
                     hover:bg-purple-500 active:scale-95 transition-all cursor-pointer"
        >
          Get Started
          <svg
            xmlns="http://www.w3.org/2000/svg"
            className="h-4 w-4"
            viewBox="0 0 20 20"
            fill="currentColor"
          >
            <path
              fillRule="evenodd"
              d="M10.293 3.293a1 1 0 011.414 0l6 6a1 1 0 010 1.414l-6 6a1 1 0 01-1.414-1.414L14.586 11H3a1 1 0 110-2h11.586l-4.293-4.293a1 1 0 010-1.414z"
              clipRule="evenodd"
            />
          </svg>
        </button>

        {/* Trust line */}
        <p className="text-xs text-gray-600">
          Secured by Keycloak · PKCE OAuth2 · No seed phrases stored
        </p>
      </main>

      {/* ── Feature cards ───────────────────────────────────────────────── */}
      <section className="grid grid-cols-1 sm:grid-cols-3 gap-4 px-8 pb-16 max-w-4xl mx-auto w-full">
        {FEATURES.map((f) => (
          <div
            key={f.title}
            className="rounded-xl border border-white/5 bg-white/[0.03] p-5 text-left hover:border-purple-500/30 transition-colors"
          >
            <div className="text-2xl mb-3">{f.icon}</div>
            <h3 className="text-sm font-semibold text-white mb-1">{f.title}</h3>
            <p className="text-xs text-gray-500 leading-relaxed">{f.description}</p>
          </div>
        ))}
      </section>
    </div>
  );
}

const FEATURES = [
  {
    icon: '📡',
    title: 'Live Prices',
    description:
      'Real-time market data powered by CoinGecko and Alchemy for every asset in your portfolio.',
  },
  {
    icon: '📈',
    title: 'Portfolio Overview',
    description:
      'See total balance, per-chain breakdown, and 24h change at a glance.',
  },
  {
    icon: '🔒',
    title: 'Non-custodial',
    description:
      'Read-only wallet monitoring. We never ask for private keys or seed phrases.',
  },
];
