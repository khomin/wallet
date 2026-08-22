// ─── Alerts Page ───────────────────────────────────────────────────────────
// Price-alert CRUD backed by the generated grpc-gateway client.

import { useState } from 'react';
import { Condition } from '../gen/alert/v1/alert_pb';
import { useAlerts, useCoins, useCreateAlert, useDeleteAlert } from '../hooks/useApi';
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
  const [showCreateModal, setShowCreateModal] = useState(false);
  const [deleteConfirmId, setDeleteConfirmId] = useState<string | null>(null);
  const [coinId, setCoinId] = useState('');
  const [condition, setCondition] = useState<Condition>(Condition.ABOVE);
  const [price, setPrice] = useState('');

  const alerts = data?.alerts ?? [];
  const resetForm = () => {
    setCoinId('');
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
                  <td className="py-4 pr-4"><div className="flex items-center gap-2">
                    {coin?.imageUrl ? <img src={coin.imageUrl} alt="" className="h-6 w-6 rounded-full" /> : <span className="flex h-6 w-6 items-center justify-center rounded-full bg-purple-500/20 text-[10px] font-bold text-purple-300">{alert.coinSymbol.slice(0, 2)}</span>}
                    <span className="font-medium text-white">{alert.coinSymbol.toUpperCase()}</span>
                  </div></td>
                  <td className="py-4 pr-4 text-gray-300">{alert.condition === Condition.ABOVE ? 'Rises above' : 'Falls below'}</td>
                  <td className="py-4 pr-4 font-mono text-gray-200">{fmtUSD(alert.price)}</td>
                  <td className="py-4 pr-4"><span className={`rounded-full px-2.5 py-1 text-xs ${triggered ? 'bg-amber-500/10 text-amber-400' : alert.enabled ? 'bg-emerald-500/10 text-emerald-400' : 'bg-white/10 text-gray-500'}`}>{triggered ? 'Triggered' : alert.enabled ? 'Active' : 'Disabled'}</span></td>
                  <td className="py-4 pr-4 text-xs text-gray-500">{fmtDate(alert.createdAt)}</td>
                  <td className="py-4 text-right"><button onClick={() => setDeleteConfirmId(alert.id)} className="text-xs text-gray-500 transition-colors hover:text-red-400">Delete</button></td>
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
              <select value={coinId} onChange={(event) => setCoinId(event.target.value)} className="w-full rounded-xl border border-white/10 bg-black/20 px-3 py-3 text-sm text-white outline-none focus:border-purple-500/60">
                <option value="">Select an asset</option>
                {(coinsData?.token ?? []).map((coin) => <option key={coin.symbol} value={coin.symbol}>{coin.name || coin.symbol} ({coin.symbol.toUpperCase()})</option>)}
              </select>
            </Field>
            <Field label="Notify me when price">
              <div className="flex gap-2"><select value={condition} onChange={(event) => setCondition(Number(event.target.value) as Condition)} className="rounded-xl border border-white/10 bg-black/20 px-3 text-sm text-white outline-none focus:border-purple-500/60"><option value={Condition.ABOVE}>Rises above</option><option value={Condition.BELOW}>Falls below</option></select><input type="number" min="0" step="any" required value={price} onChange={(event) => setPrice(event.target.value)} placeholder="Target price (USD)" className="min-w-0 flex-1 rounded-xl border border-white/10 bg-black/20 px-3 py-3 font-mono text-sm text-white outline-none focus:border-purple-500/60" /></div>
            </Field>
          </div>
          {createAlert.isError && <p className="mt-4 text-xs text-red-400">{createAlert.error.message || 'Failed to create alert'}</p>}
          <div className="mt-7 flex justify-end gap-3 border-t border-white/5 pt-5"><button type="button" onClick={resetForm} className="rounded-lg px-4 py-2 text-sm text-gray-400 hover:text-white">Cancel</button><button type="submit" disabled={createAlert.isPending || !coinId || Number(price) <= 0} className="rounded-xl bg-purple-600 px-4 py-2.5 text-sm font-semibold text-white hover:bg-purple-500 disabled:cursor-not-allowed disabled:opacity-40">{createAlert.isPending ? 'Creating...' : 'Create alert'}</button></div>
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