// ─── Wallets Page ──────────────────────────────────────────────────────────
// Full wallet management: stats row, wallet table, add/delete modals.

import { useState } from 'react';
import type { Token } from '../gen/price/v1/price_pb';
import WAValidator from 'multicoin-address-validator';
import { useWallets, useCreateWallet, useDeleteWallet, useCoins } from '../hooks/useApi';
import { Modal, Field, Spinner, ErrorBlock, EmptyBlock } from '../components/ui';
import type { CreateWalletFormState } from '../types/api';

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
  for (const c of coinsData?.token ?? []) {
    coinImageMap[c.symbol.toLowerCase()] = c.imageUrl;
  }

  // ── Modal state ───────────────────────────────────────────────────────
  const [showAddModal, setShowAddModal] = useState(false);
  const [deleteConfirmId, setDeleteConfirmId] = useState<string | null>(null);
  const [addStep, setAddStep] = useState<1 | 2 | 3>(1);
  const [assetSearch, setAssetSearch] = useState('');
  const [selectedAssetSymbol, setSelectedAssetSymbol] = useState('');

  // ── Form state ────────────────────────────────────────────────────────
  const [form, setForm] = useState<CreateWalletFormState>({
    chains: [],
    address: '',
    tokenSymbol: '',
    label: '',
  });

  const wallets = walletsData?.wallet ?? [];

  // ── Handlers ──────────────────────────────────────────────────────────
  const selectedAsset = (coinsData?.token ?? []).find(
    (coin) => coin.symbol.toLowerCase() === selectedAssetSymbol.toLowerCase(),
  );

  const chooseAsset = (coin: Token) => {
    setSelectedAssetSymbol(coin.symbol);
    setForm({
      // Native assets use their symbol as the chain identifier. Their
      // metadata intentionally has no entries in `chains`.
      chains: coin.isNative ? [coin.symbol] : [],
      address: '',
      tokenSymbol: coin.symbol.toUpperCase(),
      label: `My ${coin.name || coin.symbol} Wallet`,
    });
    setAddStep(2);
  };

  const chooseChain = (chain: string) => {
    setForm((current) => ({ ...current, chains: [chain], address: '' }));
    setAddStep(3);
  };

  const networkLabel = (chain: string) => {
    chain = chain.toUpperCase()
    const labels: Record<string, string> = { ETH: 'Ethereum ERC-20', SOL: 'Solana SPL', BTC: 'Bitcoin', TRX: 'Tron TRC-20' };
    return labels[chain] || chain;
  };

  const isAddressValid = (address: string, chain: string) => {
    const value = address.trim();
    if (!value) return false;
    return WAValidator.validate(address, chain.toLowerCase())
  };

  const handleAddWallet = async () => {
    if (form.chains.length == 0) return
    var targetChain = form.chains[0];
    if (!form.address.trim() || !form.tokenSymbol.trim() || !isAddressValid(form.address, targetChain)) return;
    try {
      await createWallet.mutateAsync({
        chain: targetChain,
        address: form.address,
        tokenSymbol: form.tokenSymbol,
        label: form.label,
      });
      setShowAddModal(false);
      setAddStep(1);
      setSelectedAssetSymbol('');
      setAssetSearch('');
      setForm({ chains: [], address: '', tokenSymbol: '', label: '' });
    } catch {
      // error shown inline via mutation state
    }
  };

  const closeAddModal = () => {
    setShowAddModal(false);
    setAddStep(1);
    setSelectedAssetSymbol('');
    setAssetSearch('');
    setForm({ chains: [], address: '', tokenSymbol: '', label: '' });
  };

  const handleDeleteWallet = async (id: string) => {
    try {
      await deleteWallet.mutateAsync(id);
      setDeleteConfirmId(null);
    } catch {
      // error shown inline
    }
  };

  const chain = form.chains[0] || '';
  const isValidAddress = isAddressValid(form.address, chain);
  const isAddressStep = selectedAsset?.isNative ? addStep === 2 : addStep === 3;
  const totalAddSteps = selectedAsset?.isNative ? 2 : 3;

  // ── Render ────────────────────────────────────────────────────────────
  return (
    < div className="max-w-6xl mx-auto" >
      <h1 className="text-xl font-semibold mb-6">👛 Wallets</h1>

      {/* Wallets table card */}
      <div className="rounded-xl border border-white/5 bg-white/[0.03] p-6">
        <div className="flex items-center justify-between mb-4">
          <h2 className="text-sm font-semibold text-white">Your Wallets</h2>
          <button
            onClick={() => { setAddStep(1); setAssetSearch(''); setSelectedAssetSymbol(''); setShowAddModal(true); }}
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
                  <th className="pb-3 text-xs font-medium text-gray-500 uppercase tracking-wider">Token</th>
                  <th className="pb-3 text-xs font-medium text-gray-500 uppercase tracking-wider">USD Value</th>
                  <th className="pb-3 text-xs font-medium text-gray-500 uppercase tracking-wider">Balance</th>
                  <th className="pb-3 text-xs font-medium text-gray-500 uppercase tracking-wider">24h</th>
                  <th className="pb-3 text-xs font-medium text-gray-500 uppercase tracking-wider">Chain</th>
                  <th className="pb-3 text-xs font-medium text-gray-500 uppercase tracking-wider">Label</th>
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
                      <div className="flex items-center gap-1.5">
                        <img
                          src={coinImageMap[wallet.tokenSymbol.toLowerCase()]}
                          alt={wallet.tokenSymbol}
                          className="w-5 h-5 rounded-full"
                          onError={(e) => {
                            (e.currentTarget as HTMLImageElement).style.display = 'none';
                          }}
                        />
                        <span className="text-gray-200 font-medium">{wallet.tokenSymbol}</span>
                      </div>
                    </td>

                    <td className="py-3 pr-4 font-mono text-xs">
                      {wallet.hasError ? (
                        <span className="text-amber-500/80 text-3xl" title={wallet.errorMsg}>
                          ⚠
                        </span>
                      ) : (
                        <span className="text-gray-200">{fmtUSD(wallet.balanceUsd)}</span>
                      )}
                    </td>

                    <td className="py-3 pr-4 font-mono text-xs">
                      {wallet.hasError ? (
                        <span className="text-amber-500/80 text-3xl" title={wallet.errorMsg}>
                          ⚠
                        </span>
                      ) : (
                        <span className="text-gray-200">{fmtCrypto(wallet.balanceCrypto)}</span>
                      )}
                    </td>

                    <td className="py-3 pr-4">
                      <span className={(wallet.price?.priceChangePercentage24h ?? 0) >= 0 ? 'text-green-400' : 'text-red-400'}>
                        {fmtPct(wallet.price?.priceChangePercentage24h ?? 0)}
                      </span>
                    </td>

                    <td className="py-3 pr-4">
                      <span className="inline-flex items-center gap-1.5 rounded-full border border-white/10 bg-white/5 px-2.5 py-0.5 text-xs text-gray-300">
                        {wallet.chain}
                      </span>
                    </td>

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

                    <td className="py-3 pr-3 text-right">
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
      {
        showAddModal && (
          <Modal onClose={closeAddModal} title="Add a wallet">
            <div className="mb-6 flex items-center gap-3">
              <div className="flex flex-1 items-center gap-2">
                {Array.from({ length: totalAddSteps }, (_, index) => index + 1).map((step) => (
                  <div key={step} className={`h-1 flex-1 rounded-full transition-colors ${step <= addStep ? 'bg-purple-500' : 'bg-white/10'}`} />
                ))}
              </div>
              <span className="whitespace-nowrap text-[11px] font-medium uppercase tracking-wider text-gray-500">Step {addStep} of {totalAddSteps}</span>
            </div>

            <form onSubmit={(e) => { e.preventDefault(); handleAddWallet(); }}>
              {addStep === 1 ? (
                <div className="animate-[fadeIn_.2s_ease-out]">
                  <div className="mb-4">
                    <h4 className="text-base font-semibold text-white">Select crypto asset</h4>
                    <p className="mt-1 text-xs text-gray-500">Choose the asset and network you want to track.</p>
                  </div>
                  <div className="relative mb-3">
                    <svg className="pointer-events-none absolute left-3 top-3 h-4 w-4 text-gray-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><circle cx="11" cy="11" r="7" /><path d="m20 20-4-4" /></svg>
                    <input autoFocus value={assetSearch} onChange={(e) => setAssetSearch(e.target.value)} placeholder="Search crypto asset, e.g. Ethereum, Solana, USDC..." className="w-full rounded-xl border border-white/10 bg-black/20 py-2.5 pl-10 pr-3 text-xs text-white outline-none transition focus:border-purple-500/60 focus:ring-2 focus:ring-purple-500/20" />
                  </div>
                  <div className="dark-scrollbar max-h-72 space-y-1 overflow-y-auto pr-1">
                    {(coinsData?.token ?? []).filter((coin) => `${coin.name} ${coin.symbol} ${coin.chains.join(' ')}`.toLowerCase().includes(assetSearch.toLowerCase())).map((coin) => (
                      <button type="button" key={`${coin.symbol}-${coin.chains.join('-')}`} onClick={() => chooseAsset(coin)} className="group flex w-full items-center gap-3 rounded-xl border border-transparent p-3 text-left transition hover:border-white/10 hover:bg-white/[0.06]">
                        {coin.imageUrl ? <img src={coin.imageUrl} alt="" className="h-6 w-6 rounded-full bg-white/10" /> : <span className="flex h-6 w-6 items-center justify-center rounded-full bg-purple-500/20 text-[10px] font-bold text-purple-300">{coin.symbol.slice(0, 2)}</span>}
                        <span className="min-w-0 flex-1"><span className="block truncate text-sm font-medium text-gray-100">{coin.name || coin.symbol}</span><span className="text-xs text-gray-500">{coin.symbol.toUpperCase()}</span></span>
                      </button>
                    ))}
                    {coinsData?.token && coinsData.token.length === 0 && <p className="py-8 text-center text-xs text-gray-500">No supported assets found.</p>}
                  </div>
                </div>
              ) : addStep === 2 && !selectedAsset?.isNative ? (
                <div className="animate-[fadeIn_.2s_ease-out]">
                  <div className="mb-5">
                    <h4 className="text-base font-semibold text-white">Select network</h4>
                    <p className="mt-1 text-xs text-gray-500">Choose the chain for this asset.</p>
                  </div>
                  <div className="space-y-2">
                    {selectedAsset?.chains.map((chain) => (
                      <button type="button" key={`${chain.symbol}}`} onClick={() => chooseChain(chain.symbol)} className="group flex w-full items-center gap-3 rounded-xl border border-transparent p-3 text-left transition hover:border-white/10 hover:bg-white/[0.06]">
                        {chain.imageUrl ? <img src={chain.imageUrl} alt="" className="h-6 w-6 rounded-full bg-white/10" /> : <span className="flex h-6 w-6 items-center justify-center rounded-full bg-purple-500/20 text-[10px] font-bold text-purple-300">{chain.symbol.slice(0, 2)}</span>}
                        <span className="min-w-0 flex-1"><span className="block truncate text-sm font-medium text-gray-100">{chain.name || chain.symbol}</span><span className="text-xs text-gray-500">{chain.symbol.toUpperCase()}</span></span>
                      </button>
                    ))}
                  </div>
                </div>
              ) : (
                <div className="animate-[fadeIn_.2s_ease-out]">
                  <div className="mb-5 flex items-center justify-between rounded-xl border border-white/10 bg-white/[0.04] p-3">
                    <div className="flex items-center gap-3">{selectedAsset?.imageUrl ? <img src={selectedAsset.imageUrl} alt="" className="h-8 w-8 rounded-full" /> : <span className="flex h-8 w-8 items-center justify-center rounded-full bg-purple-500/20 text-xs text-purple-300">{form.tokenSymbol.slice(0, 2)}</span>}<div><p className="text-sm font-semibold text-white">{selectedAsset?.name || form.tokenSymbol}</p><p className="text-[11px] text-gray-500">{form.tokenSymbol} · {networkLabel(chain)}</p></div></div>
                    <button type="button" onClick={() => setAddStep(1)} className="text-xs font-medium text-purple-400 hover:text-purple-300">Change</button>
                  </div>
                  <Field label="Wallet address">
                    <div className="relative"><input autoFocus type="text" placeholder="Paste your wallet address" value={form.address} onChange={(e) => setForm({ ...form, address: e.target.value })} className={`w-full rounded-xl border bg-black/20 py-3 pl-3 pr-20 font-mono text-xs text-white outline-none transition placeholder:text-gray-600 focus:ring-2 ${form.address ? (isValidAddress ? 'border-emerald-500/60 focus:ring-emerald-500/20' : 'border-red-500/50 focus:ring-red-500/20') : 'border-white/10 focus:border-purple-500/60 focus:ring-purple-500/20'}`} /><button type="button" onClick={async () => setForm({ ...form, address: await navigator.clipboard.readText() })} className="absolute right-2 top-1.5 rounded-lg px-2 py-1.5 text-xs font-medium text-purple-300 hover:bg-white/10">Paste</button></div>
                    {form.address && <p className={`mt-2 flex items-center gap-1.5 text-xs ${isValidAddress ? 'text-emerald-400' : 'text-red-400'}`}>{isValidAddress ? '✓ Valid wallet address' : '× Check this address format'}</p>}
                  </Field>
                  {isValidAddress && <div className="mt-5 rounded-xl border border-white/5 bg-white/[0.025] p-3"><p className="mb-2 text-[10px] font-semibold uppercase tracking-wider text-gray-500">Auto-detected metadata</p><div className="flex gap-2"><span className="rounded-lg bg-white/[0.07] px-2.5 py-1.5 text-xs text-gray-300">Token <b className="ml-1 text-white">{form.tokenSymbol}</b></span><span className="rounded-lg bg-white/[0.07] px-2.5 py-1.5 text-xs text-gray-300">Decimals <b className="ml-1 text-white">{/SOL/i.test(chain) ? 9 : /BTC/i.test(chain) ? 8 : 18}</b></span></div></div>}
                  <div className="mt-5"><Field label="Wallet label (optional)"><input type="text" value={form.label} onChange={(e) => setForm({ ...form, label: e.target.value })} className="w-full rounded-xl border border-white/10 bg-black/20 px-3 py-3 text-sm text-white outline-none focus:border-purple-500/60 focus:ring-2 focus:ring-purple-500/20" /></Field></div>
                </div>
              )}
              {createWallet.isError && <p className="mt-4 text-xs text-red-400">{(createWallet.error as Error)?.message || 'Failed to create wallet'}</p>}
              <div className="mt-7 flex items-center justify-between border-t border-white/5 pt-5"><button type="button" onClick={() => isAddressStep ? setAddStep(selectedAsset?.isNative ? 1 : 2) : addStep === 2 ? setAddStep(1) : closeAddModal()} className="rounded-lg px-3 py-2 text-sm text-gray-500 hover:bg-white/5 hover:text-white">{addStep === 1 ? 'Cancel' : 'Back'}</button>{isAddressStep && <button type="submit" disabled={createWallet.isPending || !isValidAddress} className="flex items-center gap-2 rounded-xl bg-purple-600 px-4 py-2.5 text-sm font-semibold text-white shadow-lg shadow-purple-900/20 transition hover:bg-purple-500 disabled:cursor-not-allowed disabled:opacity-40">{createWallet.isPending && <span className="h-3.5 w-3.5 animate-spin rounded-full border-2 border-white/40 border-t-white" />}{createWallet.isPending ? 'Adding...' : 'Add Wallet'}</button>}</div>
            </form>
          </Modal>
        )
      }

      {/* ── Delete Confirmation Modal ──────────────────────────────────── */}
      {
        deleteConfirmId && (
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
        )
      }
    </div >
  );
}