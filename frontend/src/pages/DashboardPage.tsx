// ─── Dashboard Page ────────────────────────────────────────────────────────
// Portfolio summary: stats row, popular coin prices, and a compact
// wallet snapshot. Full wallet management lives on /wallets.

import { useNavigate } from 'react-router-dom';
import { useWallets, usePrices, useCoins } from '../hooks/useApi';
import { StatCard, Spinner } from '../components/ui';

// ─── Formatting helpers ──────────────────────────────────────────────────

const fmtUSD = (n: number) =>
  new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD',
    minimumFractionDigits: 2,
  }).format(n);

const fmtCryptoCompact = (n: number) =>
  new Intl.NumberFormat('en-US', { maximumFractionDigits: 4 }).format(n);

const fmtPct = (n: number) => {
  const prefix = n >= 0 ? '+' : '';
  return `${prefix}${n.toFixed(2)}%`;
};

const fmtLarge = (n: number) => {
  if (n >= 1e12) return `$${(n / 1e12).toFixed(2)}T`;
  if (n >= 1e9) return `$${(n / 1e9).toFixed(2)}B`;
  if (n >= 1e6) return `$${(n / 1e6).toFixed(2)}M`;
  return fmtUSD(n);
};

// ─── Popular coins to always show ────────────────────────────────────────

const POPULAR_SYMBOLS = ['BTC', 'ETH', 'SOL', 'BNB', 'XRP', 'ADA', 'DOGE', 'MATIC'];

// ─── Component ────────────────────────────────────────────────────────────

export default function DashboardPage() {
  const navigate = useNavigate();

  const { data: walletsData, isLoading: walletsLoading } = useWallets();
  const { data: pricesData, isLoading: pricesLoading } = usePrices(POPULAR_SYMBOLS);
  const { data: coinsData } = useCoins();

  // Build a lookup map: symbol → image_url
  const coinImageMap: Record<string, string> = {};
  for (const c of coinsData?.coins ?? []) {
    coinImageMap[c.symbol.toLowerCase()] = c.image_url;
  }

  const wallets = walletsData?.wallet ?? [];
  const totalBalance = walletsData?.total_balance_usd ?? 0;
  const walletCount = walletsData?.total ?? 0;
  const prices = pricesData?.price ?? [];

  const weightedChange24h =
    wallets.length > 0 && totalBalance > 0
      ? wallets.reduce(
        (acc, w) =>
          acc + w.change_24h_percent * (w.balance_usd / totalBalance),
        0,
      )
      : 0;

  const topWallets = [...wallets]
    .sort((a, b) => b.balance_usd - a.balance_usd)
    .slice(0, 5);

  return (
    <div className="max-w-6xl mx-auto space-y-8">
      <h1 className="text-xl font-semibold">📊 Dashboard</h1>

      {/* Stats row */}
      <div className="grid grid-cols-1 sm:grid-cols-3 gap-4">
        <StatCard
          label="Total Balance"
          value={walletsLoading ? '$—' : fmtUSD(totalBalance)}
          sub="Across all wallets"
        />
        <StatCard
          label="24h Change"
          value={walletsLoading ? '—' : fmtPct(weightedChange24h)}
          sub="Weighted portfolio change"
          highlight={weightedChange24h >= 0 ? 'positive' : 'negative'}
        />
        <StatCard
          label="Tracked Wallets"
          value={walletsLoading ? '—' : String(walletCount)}
          sub="Active addresses"
        />
      </div>

      {/* Market Overview */}
      <div className="rounded-xl border border-white/5 bg-white/[0.03] p-6">
        <h2 className="text-sm font-semibold text-white mb-4">
          📈 Market Overview
        </h2>

        {pricesLoading && <Spinner />}

        {!pricesLoading && prices.length > 0 && (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-white/5 text-left">
                  <th className="pb-3 text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Coin
                  </th>
                  <th className="pb-3 text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Price
                  </th>
                  <th className="pb-3 text-xs font-medium text-gray-500 uppercase tracking-wider">
                    24h Change
                  </th>
                  <th className="pb-3 text-xs font-medium text-gray-500 uppercase tracking-wider hidden sm:table-cell">
                    Market Cap
                  </th>
                  <th className="pb-3 text-xs font-medium text-gray-500 uppercase tracking-wider hidden sm:table-cell">
                    Volume (24h)
                  </th>
                </tr>
              </thead>
              <tbody>
                {prices.map((coin) => (
                  <tr
                    key={coin.symbol}
                    className="border-b border-white/[0.02] hover:bg-white/[0.02] transition-colors"
                  >
                    <td className="py-3 pr-4">
                      <div className="flex items-center gap-2">
                        <img
                          src={coinImageMap[coin.symbol.toLowerCase()]}
                          alt={coin.symbol}
                          className="w-5 h-5 rounded-full"
                          onError={(e) => {
                            (e.currentTarget as HTMLImageElement).style.display = 'none';
                          }}
                        />
                        <span className="font-medium text-white">
                          {coin.symbol.toUpperCase()}
                        </span>
                        <span className="text-xs text-gray-500">{coin.name}</span>
                      </div>
                    </td>
                    <td className="py-3 pr-4 text-gray-200 font-mono text-xs">
                      {fmtUSD(coin.price_usd)}
                    </td>
                    <td className="py-3 pr-4">
                      <span
                        className={
                          coin.price_change_percentage_24h >= 0
                            ? 'text-green-400'
                            : 'text-red-400'
                        }
                      >
                        {fmtPct(coin.price_change_percentage_24h)}
                      </span>
                    </td>
                    <td className="py-3 pr-4 text-gray-400 font-mono text-xs hidden sm:table-cell">
                      {fmtLarge(coin.market_cap)}
                    </td>
                    <td className="py-3 text-gray-400 font-mono text-xs hidden sm:table-cell">
                      {fmtLarge(coin.total_volume)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {!pricesLoading && prices.length === 0 && (
          <p className="text-sm text-gray-500 py-8 text-center">
            Market data unavailable.
          </p>
        )}
      </div>

      {/* Wallet Snapshot */}
      <div className="rounded-xl border border-white/5 bg-white/[0.03] p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-sm font-semibold text-white">
            👛 Wallet Snapshot
          </h2>
          <button
            onClick={() => navigate('/wallets')}
            className="text-xs text-purple-400 hover:text-purple-300 transition-colors cursor-pointer"
          >
            View all →
          </button>
        </div>

        {walletsLoading && <Spinner />}

        {!walletsLoading && wallets.length === 0 && (
          <div className="text-center py-10">
            <p className="text-gray-500 text-sm mb-3">
              No wallets connected yet.
            </p>
            <button
              onClick={() => navigate('/wallets')}
              className="rounded-lg bg-purple-600 px-4 py-2 text-xs font-medium hover:bg-purple-500 transition-colors cursor-pointer"
            >
              + Add your first wallet
            </button>
          </div>
        )}

        {!walletsLoading && topWallets.length > 0 && (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-white/5 text-left">
                  <th className="pb-3 text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Label
                  </th>
                  <th className="pb-3 text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Token
                  </th>
                  <th className="pb-3 text-xs font-medium text-gray-500 uppercase tracking-wider">
                    Balance
                  </th>
                  <th className="pb-3 text-xs font-medium text-gray-500 uppercase tracking-wider">
                    USD Value
                  </th>
                  <th className="pb-3 text-xs font-medium text-gray-500 uppercase tracking-wider">
                    24h
                  </th>
                </tr>
              </thead>
              <tbody>
                {topWallets.map((wallet) => (
                  <tr
                    key={wallet.id}
                    className="border-b border-white/[0.02] hover:bg-white/[0.02] transition-colors"
                  >
                    <td className="py-3 pr-4">
                      <div className="flex flex-col">
                        <span className="font-medium text-white">
                          {wallet.label || '—'}
                        </span>
                        <span className="text-xs text-gray-600 font-mono">
                          {wallet.address.slice(0, 6)}...
                          {wallet.address.slice(-4)}
                        </span>
                      </div>
                    </td>
                    <td className="py-3 pr-4">
                      <div className="flex items-center gap-1.5">
                        <img
                          src={coinImageMap[wallet.token_symbol.toLowerCase()]}
                          alt={wallet.token_symbol}
                          className="w-5 h-5 rounded-full"
                          onError={(e) => {
                            (e.currentTarget as HTMLImageElement).style.display = 'none';
                          }}
                        />
                        <span className="text-gray-200 font-medium">
                          {wallet.token_symbol}
                        </span>
                        <span className="text-xs text-gray-500">({wallet.chain})</span>
                      </div>
                    </td>
                    <td className="py-3 pr-4 text-gray-200 font-mono text-xs">
                      {fmtCryptoCompact(wallet.balance_crypto)}
                    </td>
                    <td className="py-3 pr-4 text-gray-200 font-mono text-xs">
                      {fmtUSD(wallet.balance_usd)}
                    </td>
                    <td className="py-3">
                      <span
                        className={
                          wallet.change_24h_percent >= 0
                            ? 'text-green-400'
                            : 'text-red-400'
                        }
                      >
                        {fmtPct(wallet.change_24h_percent)}
                      </span>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
    </div>
  );
}
