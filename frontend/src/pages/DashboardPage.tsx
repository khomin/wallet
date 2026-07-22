// ─── Dashboard Page (Protected) ──────────────────────────────────────────────



// Main wallet portfolio view. Loads wallets from the backend and renders
// summary stats, a wallet table, and add/delete actions.

import { useState } from 'react';
import { useAuth } from '../auth/AuthContext';
import {
  useWallets,
  useCreateWallet,
  useDeleteWallet,

} from '../hooks/useApi';
import { SUPPORTED_CHAINS, type CreateWalletRequest } from '../types/api';

// ─── Formatting helpers ────────────────────────────────────────────────────

const fmtUSD = (n: number) =>
  new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 2 }).format(n);

const fmtCrypto = (n: number) =>
  new Intl.NumberFormat('en-US', { maximumFractionDigits: 6 }).format(n);

const fmtPct = (n: number) => {
  const prefix = n >= 0 ? '+' : '';
  return `${prefix}${n.toFixed(2)}%`;
};

// ─── Component ─────────────────────────────────────────────────────────────

export default function DashboardPage() {
  const { user, logout, accessToken } = useAuth();

  const displayName =
    user?.name ?? user?.preferred_username ?? user?.email ?? 'Whale';




  // ── Data ────────────────────────────────────────────────────────────────
  const {
    data: walletsData,
    isLoading: walletsLoading,
    isError: walletsError,
    refetch: refetchWallets,
  } = useWallets();

  const createWallet = useCreateWallet();
  const deleteWallet = useDeleteWallet();

  // ── Modal state ─────────────────────────────────────────────────────────
  const [showAddModal, setShowAddModal] = useState(false);
  const [deleteConfirmId, setDeleteConfirmId] = useState<string | null>(null);

  // ── Form state ──────────────────────────────────────────────────────────
  const [form, setForm] = useState<CreateWalletRequest>({
    chain: 'ETH',
    address: '',
    token_symbol: '',
    label: '',
  });

  const wallets = walletsData?.wallet ?? [];
  const totalBalance = walletsData?.total_balance_usd ?? 0;
  const walletCount = walletsData?.total ?? 0;

  // Weighted 24h change across all wallets
  const weightedChange24h =
    wallets.length > 0 && totalBalance > 0
      ? wallets.reduce(
        (acc, w) => acc + (w.change_24h_percent * (w.balance_usd / totalBalance)),
        0,
      )
      : 0;

  // ── Handlers ────────────────────────────────────────────────────────────
  const handleAddWallet = async () => {
    if (!form.address.trim() || !form.token_symbol.trim()) return;
    try {
      await createWallet.mutateAsync(form);
      setShowAddModal(false);
      setForm({ chain: 'ETH', address: '', token_symbol: '', label: '' });
    } catch {
      // error shown inline via mutation state
    }
  };

  const handleDeleteWallet = async (id: string) => {
    try {
      await deleteWallet.mutateAsync({ id });
      setDeleteConfirmId(null);
    } catch {
      // error shown inline
    }
  };

  // ── Render ──────────────────────────────────────────────────────────────
  return (
    <div className="min-h-screen bg-gray-950 text-white flex flex-col">
      {/* ── Top nav ─────────────────────────────────────────────────────── */}
      <header className="flex items-center justify-between px-8 py-5 border-b border-white/5">
        <div className="flex items-center gap-2">
          <span className="text-2xl">🐋</span>
          <span className="text-lg font-semibold tracking-tight">WhaleTracker</span>
        </div>

        <div className="flex items-center gap-4">

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


        {/* Stats row */}
        <div className="grid grid-cols-1 sm:grid-cols-3 gap-4 mb-8">















          <StatCard
            label="Total Balance"
            value={walletsLoading ? '$—' : fmtUSD(totalBalance)}
            sub="Across all chains"
          />
          <StatCard
            label="24h Change"
            value={walletsLoading ? '—' : fmtPct(weightedChange24h)}
            sub="Weighted portfolio change"
            highlight={weightedChange24h >= 0 ? 'positive' : 'negative'}
          />
          <StatCard
            label="Wallets"
            value={walletsLoading ? '—' : String(walletCount)}
            sub="Active addresses"
          />
        </div>

        {/* Wallets table */}
        <div className="rounded-xl border border-white/5 bg-white/[0.03] p-6">
          <div className="flex items-center justify-between mb-4">
            <h2 className="text-sm font-semibold text-white">Wallets</h2>












            <button
              onClick={() => setShowAddModal(true)}
              className="rounded-lg bg-purple-600 px-3 py-1.5 text-xs font-medium hover:bg-purple-500 transition-colors cursor-pointer"
            >
              + Add wallet
            </button>
          </div>


















          {/* Loading state */}
          {walletsLoading && (
            <div className="flex items-center justify-center py-16">
              <div className="h-8 w-8 animate-spin rounded-full border-3 border-purple-600 border-t-transparent" />
            </div>
          )}

          {/* Error state */}
          {walletsError && (
            <div className="flex flex-col items-center justify-center py-16 text-center">
              <span className="text-4xl mb-4">⚠️</span>
              <p className="text-gray-400 text-sm font-medium">Failed to load wallets</p>
              <button
                onClick={() => refetchWallets()}
                className="mt-3 text-xs text-purple-400 hover:text-purple-300 cursor-pointer"
              >
                Try again
              </button>
            </div>
          )}

          {/* Empty state */}
          {!walletsLoading && !walletsError && wallets.length === 0 && (
            <div className="flex flex-col items-center justify-center py-16 text-center">
              <span className="text-4xl mb-4">🐋</span>
              <p className="text-gray-400 text-sm font-medium">No wallets yet</p>
              <p className="text-gray-600 text-xs mt-1">
                Add a wallet address to start tracking your portfolio.
              </p>
            </div>
          )}

          {/* Wallet table */}
          {!walletsLoading && !walletsError && wallets.length > 0 && (
            <div className="overflow-x-auto">
              <table className="w-full text-sm">
                <thead>
                  <tr className="border-b border-white/5 text-left">
                    <th className="pb-3 text-xs font-medium text-gray-500 uppercase tracking-wider">Label</th>
                    <th className="pb-3 text-xs font-medium text-gray-500 uppercase tracking-wider">Chain</th>
                    <th className="pb-3 text-xs font-medium text-gray-500 uppercase tracking-wider">Token</th>
                    <th className="pb-3 text-xs font-medium text-gray-500 uppercase tracking-wider">Balance</th>
                    <th className="pb-3 text-xs font-medium text-gray-500 uppercase tracking-wider">USD Value</th>
                    <th className="pb-3 text-xs font-medium text-gray-500 uppercase tracking-wider">24h</th>
                    <th className="pb-3 text-xs font-medium text-gray-500 uppercase tracking-wider"></th>
                  </tr>
                </thead>
                <tbody>
                  {wallets.map((wallet) => (
                    <tr
                      key={wallet.id}
                      className="border-b border-white/[0.02] hover:bg-white/[0.02] transition-colors"
                    >
                      <td className="py-3 pr-4">
                        <div className="flex flex-col">
                          <span className="font-medium text-white">
                            {wallet.label || '—'}
                          </span>
                          <span className="text-xs text-gray-600 font-mono truncate max-w-[120px]">
                            {wallet.address.slice(0, 6)}...{wallet.address.slice(-4)}
                          </span>
                        </div>
                      </td>
                      <td className="py-3 pr-4">
                        <span className="inline-flex items-center gap-1.5 rounded-full border border-white/10 bg-white/5 px-2.5 py-0.5 text-xs text-gray-300">
                          {wallet.chain}
                        </span>
                      </td>
                      <td className="py-3 pr-4">
                        <span className="text-gray-200 font-medium">
                          {wallet.token_symbol}
                        </span>
                      </td>
                      <td className="py-3 pr-4 text-gray-200 font-mono text-xs">
                        {fmtCrypto(wallet.balance_crypto)}
                      </td>
                      <td className="py-3 pr-4 text-gray-200 font-mono text-xs">
                        {fmtUSD(wallet.balance_usd)}
                      </td>
                      <td className="py-3 pr-4">
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
                      <td className="py-3 text-right">
                        <button
                          onClick={() => setDeleteConfirmId(wallet.id)}
                          className="text-gray-600 hover:text-red-400 transition-colors cursor-pointer"
                          title="Delete wallet"
                        >
                          <svg
                            xmlns="http://www.w3.org/2000/svg"
                            className="h-4 w-4"
                            fill="none"
                            viewBox="0 0 24 24"
                            stroke="currentColor"
                            strokeWidth={2}
                          >
                            <path
                              strokeLinecap="round"
                              strokeLinejoin="round"
                              d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16"
                            />
                          </svg>
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
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

      {/* ── Add Wallet Modal ─────────────────────────────────────────────── */}
      {showAddModal && (
        <Modal onClose={() => setShowAddModal(false)} title="Add Wallet">
          <form
            onSubmit={(e) => {
              e.preventDefault();
              handleAddWallet();
            }}
            className="space-y-4"
          >
            {/* Chain selector */}
            <Field label="Chain">
              <select
                value={form.chain}
                onChange={(e) => setForm({ ...form, chain: e.target.value })}
                className="w-full rounded-lg border border-white/10 bg-gray-900 px-3 py-2 text-sm text-white
                           focus:outline-none focus:ring-2 focus:ring-purple-500/50"
              >
                {SUPPORTED_CHAINS.map((c) => (
                  <option key={c.value} value={c.value}>
                    {c.icon} {c.label} ({c.value})
                  </option>
                ))}
              </select>
            </Field>

            {/* Address */}
            <Field label="Wallet Address">
              <input
                type="text"
                placeholder="0x... or CFM..."
                value={form.address}
                onChange={(e) => setForm({ ...form, address: e.target.value })}
                className="w-full rounded-lg border border-white/10 bg-gray-900 px-3 py-2 text-sm text-white font-mono
                           placeholder:text-gray-600 focus:outline-none focus:ring-2 focus:ring-purple-500/50"
              />
            </Field>

            {/* Token symbol */}
            <Field label="Token Symbol">
              <input
                type="text"
                placeholder="e.g. ETH, USDC, XAUT"
                value={form.token_symbol}
                onChange={(e) =>
                  setForm({ ...form, token_symbol: e.target.value.toUpperCase() })
                }
                className="w-full rounded-lg border border-white/10 bg-gray-900 px-3 py-2 text-sm text-white
                           placeholder:text-gray-600 focus:outline-none focus:ring-2 focus:ring-purple-500/50"
              />
            </Field>

            {/* Label */}
            <Field label="Label (optional)">
              <input
                type="text"
                placeholder="My cold wallet"
                value={form.label}
                onChange={(e) => setForm({ ...form, label: e.target.value })}
                className="w-full rounded-lg border border-white/10 bg-gray-900 px-3 py-2 text-sm text-white
                           placeholder:text-gray-600 focus:outline-none focus:ring-2 focus:ring-purple-500/50"
              />
            </Field>

            {/* Error */}
            {createWallet.isError && (
              <p className="text-xs text-red-400">
                {(createWallet.error as Error)?.message || 'Failed to create wallet'}
              </p>
            )}

            {/* Actions */}
            <div className="flex items-center justify-end gap-3 pt-2">
              <button
                type="button"
                onClick={() => setShowAddModal(false)}
                className="rounded-lg border border-white/10 px-4 py-2 text-sm text-gray-400 hover:text-white transition-colors cursor-pointer"
              >
                Cancel
              </button>
              <button
                type="submit"
                disabled={createWallet.isPending || !form.address.trim() || !form.token_symbol.trim()}
                className="rounded-lg bg-purple-600 px-4 py-2 text-sm font-medium text-white
                           hover:bg-purple-500 disabled:opacity-40 disabled:cursor-not-allowed transition-colors cursor-pointer"
              >
                {createWallet.isPending ? 'Adding...' : 'Add Wallet'}
              </button>
            </div>
          </form>
        </Modal>
      )}

      {/* ── Delete Confirmation Modal ────────────────────────────────────── */}
      {deleteConfirmId && (
        <Modal onClose={() => setDeleteConfirmId(null)} title="Delete Wallet">
          <p className="text-sm text-gray-400">
            Are you sure you want to delete this wallet? This action cannot be
            undone.
          </p>

          {deleteWallet.isError && (
            <p className="mt-2 text-xs text-red-400">
              {(deleteWallet.error as Error)?.message || 'Failed to delete wallet'}
            </p>
          )}

          <div className="flex items-center justify-end gap-3 mt-6">
            <button
              onClick={() => setDeleteConfirmId(null)}
              className="rounded-lg border border-white/10 px-4 py-2 text-sm text-gray-400 hover:text-white transition-colors cursor-pointer"
            >
              Cancel
            </button>
            <button
              onClick={() => handleDeleteWallet(deleteConfirmId)}
              disabled={deleteWallet.isPending}
              className="rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white
                         hover:bg-red-500 disabled:opacity-40 disabled:cursor-not-allowed transition-colors cursor-pointer"
            >
              {deleteWallet.isPending ? 'Deleting...' : 'Delete'}
            </button>
          </div>
        </Modal>
      )}
    </div>
  );
}

// ─── Sub-components ───────────────────────────────────────────────────────────

/** Summary stat card in the top row */
function StatCard({
  label,
  value,
  sub,
  highlight,
}: {
  label: string;
  value: string;
  sub: string;
  highlight?: 'positive' | 'negative';
}) {
  const valueColor =
    highlight === 'positive'
      ? 'text-green-400'
      : highlight === 'negative'
        ? 'text-red-400'
        : 'text-white';

  return (
    <div className="rounded-xl border border-white/5 bg-white/[0.03] p-5">
      <p className="text-xs text-gray-500 uppercase tracking-wider mb-1">{label}</p>
      <p className={`text-2xl font-semibold ${valueColor}`}>{value}</p>
      <p className="text-xs text-gray-600 mt-1">{sub}</p>
    </div>
  );
}

/** Generic modal wrapper */
function Modal({
  children,
  onClose,
  title,
}: {
  children: React.ReactNode;
  onClose: () => void;
  title: string;
}) {
  return (
    <div
      className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm"
      onClick={(e) => {
        if (e.target === e.currentTarget) onClose();
      }}
    >
      <div className="w-full max-w-md rounded-2xl border border-white/10 bg-gray-900 p-6 shadow-2xl">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-lg font-semibold text-white">{title}</h3>
          <button
            onClick={onClose}
            className="text-gray-500 hover:text-white transition-colors cursor-pointer"
          >
            <svg
              xmlns="http://www.w3.org/2000/svg"
              className="h-5 w-5"
              fill="none"
              viewBox="0 0 24 24"
              stroke="currentColor"
              strokeWidth={2}
            >
              <path strokeLinecap="round" strokeLinejoin="round" d="M6 18L18 6M6 6l12 12" />
            </svg>
          </button>
        </div>
        {children}
      </div>
    </div>
  );
}

/** Simple labeled form field */
function Field({ label, children }: { label: string; children: React.ReactNode }) {
  return (
    <label className="block">
      <span className="text-xs font-medium text-gray-400 mb-1.5 block">{label}</span>
      {children}
    </label>
  );
}
