// ─── Market / Watchlist Page ───────────────────────────────────────────────
// Browse all supported coins with live prices.

import { useState } from 'react';
import { useCoins, usePrices } from '../hooks/useApi';
import { Spinner, ErrorBlock } from '../components/ui';

// ─── Formatting helpers ──────────────────────────────────────────────────

const fmtUSD = (n: number) =>
  new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 2 }).format(n);

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

// ─── Default symbols to show initially ──────────────────────────────────

const DEFAULT_SYMBOLS = [
  'BTC', 'ETH', 'SOL', 'BNB', 'XRP', 'ADA', 'DOGE', 'MATIC',
  'DOT', 'LINK', 'AVAX', 'UNI', 'SHIB', 'LTC', 'ATOM', 'ETC',
];

// ─── Component ────────────────────────────────────────────────────────────

export default function MarketPage() {
  const [search, setSearch] = useState('');

  const {
    data: coinsData,
    isLoading: coinsLoading,
    isError: coinsError,
    refetch: refetchCoins,
  } = useCoins();

  const allCoins = coinsData?.coins ?? [];

  // Filter coins by search input, or show defaults
  const displayedSymbols =
    search.trim().length > 0
      ? allCoins
          .filter(
            (c) =>
              c.symbol.toLowerCase().includes(search.toLowerCase()) ||
              c.name.toLowerCase().includes(search.toLowerCase()),
          )
          .map((c) => c.symbol)
      : DEFAULT_SYMBOLS;

  const {
    data: pricesData,
    isLoading: pricesLoading,
  } = usePrices(displayedSymbols);

  const prices = pricesData?.price ?? [];

  return (
    <div className="max-w-6xl mx-auto space-y-6">
      <h1 className="text-xl font-semibold">📈 Market</h1>

      {/* Search */}
      <div className="relative">
        <input
          type="text"
          placeholder="Search coins..."
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          className="w-full max-w-sm rounded-lg border border-white/10 bg-white/[0.03] px-4 py-2.5 text-sm text-white
                     placeholder:text-gray-600 focus:outline-none focus:ring-2 focus:ring-purple-500/50"
        />
        {search.trim().length > 0 && coinsLoading && (
          <div className="absolute right-3 top-2.5 h-4 w-4 animate-spin rounded-full border-2 border-purple-500 border-t-transparent" />
        )}
      </div>

      {/* Coins loading (for search) */}
      {coinsLoading && search.trim().length > 0 && (
        <Spinner />
      )}

      {/* Coins error */}
      {coinsError && (
        <ErrorBlock message="Failed to load coins" onRetry={() => refetchCoins()} />
      )}

      {/* Prices table */}
      <div className="rounded-xl border border-white/5 bg-white/[0.03] p-6">
        {pricesLoading && !coinsLoading && <Spinner />}

        {!pricesLoading && prices.length > 0 && (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead>
                <tr className="border-b border-white/5 text-left">
                  <th className="pb-3 text-xs font-medium text-gray-500 uppercase tracking-wider">#</th>
                  <th className="pb-3 text-xs font-medium text-gray-500 uppercase tracking-wider">Coin</th>
                  <th className="pb-3 text-xs font-medium text-gray-500 uppercase tracking-wider">Price</th>
                  <th className="pb-3 text-xs font-medium text-gray-500 uppercase tracking-wider">24h</th>
                  <th className="pb-3 text-xs font-medium text-gray-500 uppercase tracking-wider hidden sm:table-cell">
                    24h High
                  </th>
                  <th className="pb-3 text-xs font-medium text-gray-500 uppercase tracking-wider hidden sm:table-cell">
                    24h Low
                  </th>
                  <th className="pb-3 text-xs font-medium text-gray-500 uppercase tracking-wider hidden md:table-cell">
                    Market Cap
                  </th>
                  <th className="pb-3 text-xs font-medium text-gray-500 uppercase tracking-wider hidden md:table-cell">
                    Volume
                  </th>
                </tr>
              </thead>
              <tbody>
                {prices.map((coin, i) => (
                  <tr
                    key={coin.symbol}
                    className="border-b border-white/[0.02] hover:bg-white/[0.02] transition-colors"
                  >
                    <td className="py-3 pr-4 text-gray-500 text-xs">{i + 1}</td>
                    <td className="py-3 pr-4">
                      <span className="font-medium text-white">{coin.symbol.toUpperCase()}</span>
                      <span className="text-xs text-gray-500 ml-2">{coin.name}</span>
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
                      {fmtUSD(coin.high_24h)}
                    </td>
                    <td className="py-3 pr-4 text-gray-400 font-mono text-xs hidden sm:table-cell">
                      {fmtUSD(coin.low_24h)}
                    </td>
                    <td className="py-3 pr-4 text-gray-400 font-mono text-xs hidden md:table-cell">
                      {fmtLarge(coin.market_cap)}
                    </td>
                    <td className="py-3 text-gray-400 font-mono text-xs hidden md:table-cell">
                      {fmtLarge(coin.total_volume)}
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}

        {!pricesLoading && !coinsLoading && prices.length === 0 && (
          <p className="text-sm text-gray-500 py-8 text-center">
            {search.trim().length > 0 ? 'No coins match your search.' : 'No market data available.'}
          </p>
        )}
      </div>
    </div>
  );
}