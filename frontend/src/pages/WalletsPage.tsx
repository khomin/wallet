// ─── Wallets Page ──────────────────────────────────────────────────────────
// Full wallet management: stats row, wallet table, add/delete modals.

import { useState } from 'react';
import { useWallets, useCreateWallet, useDeleteWallet, useCoins } from '../hooks/useApi';
import { StatCard, Modal, Field, Spinner, ErrorBlock, EmptyBlock } from '../components/ui';
import { SUPPORTED_CHAINS, type CreateWalletRequest } from '../types/api';

// ─── Formatting helpers ──────────────────────────────────────────────────

const fmtUSD = (n: number) =>
  new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 2 }).format(n);

const fmtCrypto = (n: number) =>
  new Intl.NumberFormat('en-US', { maximumFractionDigits: 6 }).format(n);

const fmtPct = (n: number) => {
  const prefix = n >= 0 ? '+' : '';
  return `${prefix}${n.toFixed(2)}%`;
};

// ─── Component ────────────────────────────────────────────────────────────

export default function WalletsPage() {
  // ── Data ──────────────────────────────────────────────────────────────
  const {
    data: walletsData,
    isLoading: walletsLoading,
    isError: walletsError,
    refetch: refetchWallets,
  } = useWallets();

  const createWallet = useCreateWallet();
  const deleteWallet = useDeleteWallet();
  const { data: coinsData } = useCoins();

  // Build a lookup map: symbol → image_url
  const coinImageMap: Record<string, string> = {};
  for (const c of coinsData?.coins ?? []) {
    coinImageMap[c.symbol.toLowerCase()] = c.image_url;
  }

  // ── Modal state ───────────────────────────────────────────────────────
  const [showAddModal, setShowAddModal] = useState(false);
  const [deleteConfirmId, setDeleteConfirmId] = useState<string | null>(null);

  // ── Form state ────────────────────────────────────────────────────────
  const [form, setForm] = useState<CreateWalletRequest>({
    chain: 'ETH',
    address: '',
    token_symbol: '',
    label: '',
  });

  const wallets = walletsData?.wallet ?? [];
  const totalBalance = walletsData?.total_balance_usd ?? 0;
  const walletCount = walletsData?.total ?? 0;

  // Weighted 24h change
  const weightedChange24h =
    wallets.length > 0 && totalBalance > 0
      ? wallets.reduce(
        (acc, w) => acc + (w.change_24h_percent * (w.balance_usd / totalBalance)),
        0,
      )
      : 0;

  // ── Handlers ──────────────────────────────────────────────────────────
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

  // ── Render ────────────────────────────────────────────────────────────
  return (
    <div className="max-w-6xl mx-auto">
      <h1 className="text-xl font-semibold mb-6">👛 Wallets</h1>

      {/* Wallets table card */}
      <div className="rounded-xl border border-white/5 bg-white/[0.03] p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-sm font-semibold text-white">Your Wallets</h2>
          <button
            onClick={() => setShowAddModal(true)}
            className="rounded-lg bg-purple-600 px-3 py-1.5 text-xs font-medium hover:bg-purple-500 transition-colors cursor-pointer"
          >
            + Add wallet
          </button>
        </div>

        {/* Loading */}
        {walletsLoading && <Spinner />}

        {/* Error */}
        {walletsError && (
          <ErrorBlock message="Failed to load wallets" onRetry={() => refetchWallets()} />
        )}

        {/* Empty */}
        {!walletsLoading && !walletsError && wallets.length === 0 && (
          <EmptyBlock
            emoji="🐋"
            title="No wallets yet"
            subtitle="Add a wallet address to start tracking your portfolio."
          />
        )}

        {/* Table */}
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
                      <div className="flex items-center gap-1.5">
                        <img
                          src={coinImageMap[wallet.token_symbol.toLowerCase()]}
                          alt={wallet.token_symbol}
                          className="w-5 h-5 rounded-full"
                          onError={(e) => {
                            (e.currentTarget as HTMLImageElement).style.display = 'none';
                          }}
                        />
                        <span className="text-gray-200 font-medium">{wallet.token_symbol}</span>
                      </div>
                    </td>
                    <td className="py-3 pr-4 text-gray-200 font-mono text-xs">
                      {fmtCrypto(wallet.balance_crypto)}
                    </td>
                    <td className="py-3 pr-4 text-gray-200 font-mono text-xs">
                      {fmtUSD(wallet.balance_usd)}
                    </td>
                    <td className="py-3 pr-4">
                      <span className={wallet.change_24h_percent >= 0 ? 'text-green-400' : 'text-red-400'}>
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

      {/* ── Add Wallet Modal ───────────────────────────────────────────── */}
      {showAddModal && (
        <Modal onClose={() => setShowAddModal(false)} title="Add Wallet">
          <form
            onSubmit={(e) => {
              e.preventDefault();
              handleAddWallet();
            }}
            className="space-y-4"
          >
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

            <Field label="Token Symbol">
              <input
                type="text"
                placeholder="e.g. ETH, USDC, XAUT"
                value={form.token_symbol}
                onChange={(e) => setForm({ ...form, token_symbol: e.target.value.toUpperCase() })}
                className="w-full rounded-lg border border-white/10 bg-gray-900 px-3 py-2 text-sm text-white
                           placeholder:text-gray-600 focus:outline-none focus:ring-2 focus:ring-purple-500/50"
              />
            </Field>

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

            {createWallet.isError && (
              <p className="text-xs text-red-400">
                {(createWallet.error as Error)?.message || 'Failed to create wallet'}
              </p>
            )}

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

      {/* ── Delete Confirmation Modal ──────────────────────────────────── */}
      {deleteConfirmId && (
        <Modal onClose={() => setDeleteConfirmId(null)} title="Delete Wallet">
          <p className="text-sm text-gray-400">
            Are you sure you want to delete this wallet? This action cannot be undone.
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