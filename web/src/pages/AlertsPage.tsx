// ─── Alerts Page ───────────────────────────────────────────────────────────
// Price-alert CRUD backed by the generated grpc-gateway client.

import { useState } from 'react';
import type { Token } from '../gen/price/v1/price_pb';
import type { Alert } from '../gen/alert/v1/alert_pb';
import { Condition } from '../gen/alert/v1/alert_pb';
import {
  useAlerts,
  useCoins,
  useCreateAlert,
  useDeleteAlert,
  usePauseAlert,
  useResumeAlert,
  useUpdateAlert,
  usePrices,
} from '../hooks/useApi';
import { EmptyBlock, ErrorBlock, Field, Modal, Spinner } from '../components/ui';

const fmtUSD = (value: number) =>
  new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', maximumFractionDigits: 8 }).format(value);

const fmtDate = (timestamp?: { seconds: bigint | number } | undefined) => {
  if (!timestamp) return '—';
  return new Date(Number(timestamp.seconds) * 1000).toLocaleString();
};

export default function AlertsPage() {
  const { data, isLoading, isError, refetch } = useAlerts();
  const { data: coinsData } = useCoins();
  const createAlert = useCreateAlert();
  const deleteAlert = useDeleteAlert();
  const pauseAlert = usePauseAlert();
  const resumeAlert = useResumeAlert();
  const updateAlert = useUpdateAlert();
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [editingAlert, setEditingAlert] = useState<Alert | null>(null);
  const [editCondition, setEditCondition] = useState<Condition>(Condition.ABOVE);
  const [editPrice, setEditPrice] = useState('');
  const [deleteConfirmId, setDeleteConfirmId] = useState<string | null>(null);
  const [coinId, setCoinId] = useState('');
  const [coinSearch, setCoinSearch] = useState('');
  const [condition, setCondition] = useState<Condition>(Condition.ABOVE);
  const [price, setPrice] = useState('');
  const { data: pricesData, isLoading: pricesLoading } = usePrices(coinId ? [coinId] : []);

  const alerts = data?.alerts ?? [];
  const resetForm = () => {
    setCoinId('');
    setCoinSearch('');
    setCondition(Condition.ABOVE);
    setPrice('');
    setShowCreateModal(false);
    createAlert.reset();
  };

  const handleCreate = async () => {
    const targetPrice = Number(price);
    if (!coinId || !Number.isFinite(targetPrice) || targetPrice <= 0) return;
    try {
      // The alert API calls this field coin_id; the price catalog identifies
      // supported assets by symbol, so the selected symbol is sent as its ID.
      await createAlert.mutateAsync({ coinId, condition, price: targetPrice });
      resetForm();
    } catch {
      // The mutation error is rendered in the modal.
    }
  };

  const selectedCoin = coinsData?.token.find((coin) => coin.symbol === coinId);
  const currentPrice = pricesData?.price.find(
    (item) => item.symbol.toLowerCase() === coinId.toLowerCase(),
  )?.priceUsd;
  const filteredCoins = (coinsData?.token ?? []).filter((coin) =>
    `${coin.name} ${coin.symbol}`.toLowerCase().includes(coinSearch.toLowerCase()),
  );

  const chooseCoin = (coin: Token) => {
    setCoinId(coin.symbol);
    setCoinSearch('');
  };

  const openEdit = (alert: Alert) => {
    setEditingAlert(alert);
    setEditCondition(alert.condition);
    setEditPrice(String(alert.price));
    updateAlert.reset();
  };

  const closeEdit = () => {
    setEditingAlert(null);
    updateAlert.reset();
  };

  const handleUpdate = async () => {
    if (!editingAlert) return;
    const targetPrice = Number(editPrice);
    if (!Number.isFinite(targetPrice) || targetPrice <= 0) return;
    try {
      await updateAlert.mutateAsync({
        id: editingAlert.id,
        condition: editCondition,
        price: targetPrice,
      });
      closeEdit();
    } catch {
      // The mutation error is rendered in the edit modal.
    }
  };

  const handleToggle = async (id: string, enabled: boolean) => {
    try {
      if (enabled) await pauseAlert.mutateAsync(id);
      else await resumeAlert.mutateAsync(id);
    } catch {
      // The list refetches after a successful mutation; errors stay local.
    }
  };

  const handleReactivate = async (id: string) => {
    try {
      await resumeAlert.mutateAsync(id);
    } catch {
      // The list refetches after a successful mutation; errors stay local.
    }
  };

  const handleDelete = async (id: string) => {
    try {
      await deleteAlert.mutateAsync(id);
      setDeleteConfirmId(null);
    } catch {
      // The mutation error is rendered in the confirmation modal.
    }
  };

  return (
    <div className="max-w-6xl mx-auto">
      <div className="mb-6 flex items-center justify-between">
        <div>
          <h1 className="text-xl font-semibold">Price alerts</h1>
          <p className="mt-1 text-xs text-gray-500">Get notified when an asset reaches your target price.</p>
        </div>
        <button onClick={() => setShowCreateModal(true)} className="rounded-lg bg-purple-600 px-3 py-1.5 text-xs font-medium transition-colors hover:bg-purple-500">
          + New alert
        </button>
      </div>

      <div className="rounded-xl border border-white/5 bg-white/[0.03] p-6">
        {isLoading && <Spinner />}
        {isError && <ErrorBlock message="Failed to load alerts" onRetry={() => refetch()} />}
        {!isLoading && !isError && alerts.length === 0 && (
          <EmptyBlock emoji="🔔" title="No price alerts" subtitle="Create an alert to start watching a coin." />
        )}
        {!isLoading && !isError && alerts.length > 0 && (
          <div className="overflow-x-auto">
            <table className="w-full text-sm">
              <thead><tr className="border-b border-white/5 text-left">
                <th className="pb-3 text-xs font-medium uppercase tracking-wider text-gray-500">Asset</th>
                <th className="pb-3 text-xs font-medium uppercase tracking-wider text-gray-500">Condition</th>
                <th className="pb-3 text-xs font-medium uppercase tracking-wider text-gray-500">Target price</th>
                <th className="pb-3 text-xs font-medium uppercase tracking-wider text-gray-500">Status</th>
                <th className="pb-3 text-xs font-medium uppercase tracking-wider text-gray-500">Created</th>
                <th className="pb-3" />
              </tr></thead>
              <tbody>{alerts.map((alert) => {
                const coin = coinsData?.token.find((item) => item.symbol.toLowerCase() === alert.coinSymbol.toLowerCase());
                const triggered = Boolean(alert.triggeredAt);
                return <tr key={alert.id} className="border-b border-white/[0.02] hover:bg-white/[0.02]">
                  <td className="py-4 pr-4"><div className="flex items-center gap-2 ml-2">
                    {coin?.imageUrl ? <img src={coin.imageUrl} alt="" className="h-6 w-6 rounded-full" /> : <span className="flex h-6 w-6 items-center justify-center rounded-full bg-purple-500/20 text-[10px] font-bold text-purple-300">{alert.coinSymbol.slice(0, 2)}</span>}
                    <span className="font-medium text-white">{alert.coinSymbol.toUpperCase()}</span>
                  </div></td>
                  <td className="py-4 pr-4 text-gray-300">{alert.condition === Condition.ABOVE ? 'Rises above' : 'Falls below'}</td>
                  <td className="py-4 pr-4 font-mono text-gray-200">{fmtUSD(alert.price)}</td>
                  <td className="py-4 pr-4">
                    <button
                      onClick={() => triggered ? void handleReactivate(alert.id) : void handleToggle(alert.id, alert.enabled)}
                      disabled={pauseAlert.isPending || resumeAlert.isPending}
                      title={triggered ? 'Activate alert again' : alert.enabled ? 'Pause alert' : 'Resume alert'}
                      aria-label={triggered ? `Activate ${alert.coinSymbol} alert again` : alert.enabled ? `Pause ${alert.coinSymbol} alert` : `Resume ${alert.coinSymbol} alert`}
                      className={`group inline-flex items-center gap-2 rounded-full border px-2.5 py-1 text-xs transition-all disabled:cursor-wait disabled:opacity-50 ${triggered ? 'border-amber-500/20 bg-amber-500/10 text-amber-400 hover:border-amber-400/50 hover:bg-amber-500/20' : alert.enabled ? 'border-emerald-500/20 bg-emerald-500/10 text-emerald-400 hover:border-emerald-400/50 hover:bg-emerald-500/20' : 'border-white/10 bg-white/[0.06] text-gray-400 hover:border-purple-400/40 hover:bg-purple-500/10 hover:text-purple-300'}`}
                    >
                      <span className={`h-1.5 w-1.5 rounded-full ${triggered ? 'bg-amber-400' : alert.enabled ? 'bg-emerald-400 shadow-[0_0_7px_rgba(52,211,153,0.8)]' : 'bg-gray-500'}`} />
                      {triggered ? 'Triggered' : alert.enabled ? 'Active' : 'Paused'}
                      <span className="ml-0.5 text-[10px] opacity-60">{triggered ? '↻' : alert.enabled ? 'Ⅱ' : '▶'}</span>
                    </button>
                  </td>
                  <td className="py-4 pr-4 text-xs text-gray-500">{fmtDate(alert.createdAt)}</td>
                  <td className="py-4 pr-4 text-right">
                    <div className="flex items-center justify-end gap-3">
                      <button onClick={() => openEdit(alert)} className="text-xs text-gray-500 transition-colors hover:text-purple-300">Edit</button>
                      <button onClick={() => setDeleteConfirmId(alert.id)} className="text-xs text-gray-500 transition-colors hover:text-red-400">Delete</button>
                    </div>
                  </td>
                </tr>;
              })}</tbody>
            </table>
          </div>
        )}
      </div>

      {showCreateModal && <Modal onClose={resetForm} title="Create price alert">
        <form onSubmit={(event) => { event.preventDefault(); void handleCreate(); }}>
          <div className="space-y-4">
            <Field label="Asset">
              {selectedCoin ? (
                <div className="flex items-center gap-3 rounded-xl border border-purple-500/40 bg-purple-500/[0.08] p-3">
                  {selectedCoin.imageUrl ? <img src={selectedCoin.imageUrl} alt="" className="h-7 w-7 rounded-full" /> : <span className="flex h-7 w-7 items-center justify-center rounded-full bg-purple-500/20 text-[10px] font-bold text-purple-300">{selectedCoin.symbol.slice(0, 2)}</span>}
                  <span className="min-w-0 flex-1"><span className="block truncate text-sm font-medium text-gray-100">{selectedCoin.name || selectedCoin.symbol}</span><span className="text-xs text-gray-500">{selectedCoin.symbol.toUpperCase()}</span></span>
                  <span className="bg-purple-500/20 text-[10px] font-bold text-purple-300">
                    {pricesLoading ? 'Loading...' : currentPrice !== undefined ? `Current ${fmtUSD(currentPrice)}` : 'Unavailable'}
                  </span>
                  <button type="button" onClick={() => setCoinId('')} className="text-xs font-medium text-purple-300 hover:text-purple-200">Change</button>
                </div>
              ) : (
                <>
                  <div className="relative mb-3">
                    <svg className="pointer-events-none absolute left-3 top-3 h-4 w-4 text-gray-500" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><circle cx="11" cy="11" r="7" /><path d="m20 20-4-4" /></svg>
                    <input autoFocus value={coinSearch} onChange={(event) => setCoinSearch(event.target.value)} placeholder="Search crypto asset..." className="w-full rounded-xl border border-white/10 bg-black/20 py-2.5 pl-10 pr-3 text-xs text-gray-300 outline-none placeholder:text-gray-600 focus:border-purple-500/60 focus:ring-2 focus:ring-purple-500/20" />
                  </div>
                  <div className="dark-scrollbar h-56 space-y-1 overflow-y-auto pr-1">
                    {filteredCoins.map((coin) => <button type="button" key={coin.symbol} onClick={() => chooseCoin(coin)} className="group flex w-full items-center gap-3 rounded-xl border border-transparent p-3 text-left transition hover:border-white/10 hover:bg-white/[0.06]">
                      {coin.imageUrl ? <img src={coin.imageUrl} alt="" className="h-7 w-7 rounded-full bg-white/10" /> : <span className="flex h-7 w-7 items-center justify-center rounded-full bg-purple-500/20 text-[10px] font-bold text-purple-300">{coin.symbol.slice(0, 2)}</span>}
                      <span className="min-w-0 flex-1"><span className="block truncate text-sm font-medium text-gray-100">{coin.name || coin.symbol}</span><span className="text-xs text-gray-500">{coin.symbol.toUpperCase()}</span></span>
                    </button>)}
                    {filteredCoins.length === 0 && <p className="py-6 text-center text-xs text-gray-500">No supported assets found.</p>}
                  </div>
                </>
              )}
            </Field>
            <Field label="Notify me when price">
              <div className="flex gap-2">
                <div className="flex shrink-0 rounded-xl border border-white/10 bg-gray-950/40 p-1">
                  <button type="button" onClick={() => setCondition(Condition.ABOVE)} className={`rounded-lg px-3 py-2 text-xs transition-colors ${condition === Condition.ABOVE ? 'bg-white/[0.08] text-gray-200' : 'text-gray-500 hover:text-gray-300'}`}>Above</button>
                  <button type="button" onClick={() => setCondition(Condition.BELOW)} className={`rounded-lg px-3 py-2 text-xs transition-colors ${condition === Condition.BELOW ? 'bg-white/[0.08] text-gray-200' : 'text-gray-500 hover:text-gray-300'}`}>Below</button>
                </div>
                <div className="relative min-w-0 flex-1"><span className="pointer-events-none absolute left-3 top-3 text-sm text-gray-600">$</span><input type="number" min="0" step="any" required value={price} onChange={(event) => setPrice(event.target.value)} placeholder="Target price" className="w-full [appearance:textfield] rounded-xl border border-white/10 bg-gray-950/40 py-3 pl-7 pr-3 font-mono text-sm text-gray-300 outline-none placeholder:text-gray-600 focus:border-purple-500/60 focus:ring-2 focus:ring-purple-500/20 [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none" /></div>
              </div>
            </Field>
          </div>
          {createAlert.isError && <p className="mt-4 text-xs text-red-400">{createAlert.error.message || 'Failed to create alert'}</p>}
          <div className="mt-7 flex justify-end gap-3 border-t border-white/5 pt-5"><button type="button" onClick={resetForm} className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-white">Cancel</button><button type="submit" disabled={createAlert.isPending || !coinId || Number(price) <= 0} className="rounded-xl bg-purple-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-purple-500 disabled:cursor-not-allowed disabled:opacity-40">{createAlert.isPending ? 'Creating...' : 'Create alert'}</button></div>
        </form>
      </Modal>}

      {editingAlert && <Modal onClose={closeEdit} title="Tune your alert">
        <form onSubmit={(event) => { event.preventDefault(); void handleUpdate(); }}>
          <div className="mb-5 rounded-xl border border-purple-500/20 bg-purple-500/[0.06] px-4 py-3">
            <p className="text-xs text-gray-500">Watching</p>
            <p className="mt-1 font-medium text-white">{editingAlert.coinSymbol.toUpperCase()} <span className="font-normal text-gray-500">price alert</span></p>
          </div>
          <Field label="Notify me when price">
            <div className="flex gap-2">
              <div className="flex shrink-0 rounded-xl border border-white/10 bg-gray-950/40 p-1">
                <button type="button" onClick={() => setEditCondition(Condition.ABOVE)} className={`rounded-lg px-3 py-2 text-xs transition-colors ${editCondition === Condition.ABOVE ? 'bg-white/[0.08] text-gray-200' : 'text-gray-500 hover:text-gray-300'}`}>Above</button>
                <button type="button" onClick={() => setEditCondition(Condition.BELOW)} className={`rounded-lg px-3 py-2 text-xs transition-colors ${editCondition === Condition.BELOW ? 'bg-white/[0.08] text-gray-200' : 'text-gray-500 hover:text-gray-300'}`}>Below</button>
              </div>
              <div className="relative min-w-0 flex-1"><span className="pointer-events-none absolute left-3 top-3 text-sm text-gray-600">$</span><input type="number" min="0" step="any" required value={editPrice} onChange={(event) => setEditPrice(event.target.value)} className="w-full [appearance:textfield] rounded-xl border border-white/10 bg-gray-950/40 py-3 pl-7 pr-3 font-mono text-sm text-gray-300 outline-none placeholder:text-gray-600 focus:border-purple-500/60 focus:ring-2 focus:ring-purple-500/20 [&::-webkit-inner-spin-button]:appearance-none [&::-webkit-outer-spin-button]:appearance-none" /></div>
            </div>
          </Field>
          {updateAlert.isError && <p className="mt-4 text-xs text-red-400">{updateAlert.error.message || 'Failed to update alert'}</p>}
          <div className="mt-7 flex justify-end gap-3 border-t border-white/5 pt-5"><button type="button" onClick={closeEdit} className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-white">Cancel</button><button type="submit" disabled={updateAlert.isPending || Number(editPrice) <= 0} className="rounded-xl bg-purple-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-purple-500 disabled:cursor-not-allowed disabled:opacity-40">{updateAlert.isPending ? 'Saving...' : 'Save changes'}</button></div>
        </form>
      </Modal>}

      {deleteConfirmId && <Modal onClose={() => setDeleteConfirmId(null)} title="Delete alert">
        <p className="text-sm text-gray-400">Are you sure you want to delete this alert?</p>
        {deleteAlert.isError && <p className="mt-3 text-xs text-red-400">{deleteAlert.error.message || 'Failed to delete alert'}</p>}
        <div className="mt-6 flex justify-end gap-3"><button onClick={() => setDeleteConfirmId(null)} className="rounded-lg border border-white/10 px-4 py-2 text-sm text-gray-400 hover:text-white">Cancel</button><button onClick={() => void handleDelete(deleteConfirmId)} disabled={deleteAlert.isPending} className="rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-500 disabled:opacity-40">{deleteAlert.isPending ? 'Deleting...' : 'Delete'}</button></div>
      </Modal>}
    </div>
  );
}