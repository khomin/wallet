import { useState, useMemo, useEffect, useRef } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Spinner, ErrorBlock, Modal } from '../components/ui';
import { useWallets, useWalletBalances, useUpdateWallet, useDeleteWallet } from '../hooks/useApi';
import { BalancePeriod } from '../gen/wallet/v1/wallet_pb';
import {
    ResponsiveContainer,
    AreaChart,
    Area,
    XAxis,
    YAxis,
    Tooltip,
    CartesianGrid,
} from 'recharts';

const PERIODS: { key: string; label: string; value: BalancePeriod }[] = [
    { key: '1D', label: '1D', value: BalancePeriod.BALANCE_PERIOD_1D },
    { key: '1W', label: '1W', value: BalancePeriod.BALANCE_PERIOD_1W },
    { key: '1M', label: '1M', value: BalancePeriod.BALANCE_PERIOD_1M },
    { key: '6M', label: '6M', value: BalancePeriod.BALANCE_PERIOD_6M },
    { key: '1Y', label: '1Y', value: BalancePeriod.BALANCE_PERIOD_1Y },
    { key: '5Y', label: '5Y', value: BalancePeriod.BALANCE_PERIOD_5Y },
    { key: 'ALL', label: 'ALL', value: BalancePeriod.BALANCE_PERIOD_ALL },
];

const fmtUSD = (n: number) =>
    new Intl.NumberFormat('en-US', { style: 'currency', currency: 'USD', minimumFractionDigits: 2 }).format(n);

function timeToMs(t: any) {
    if (!t) return Date.now();
    if (typeof t === 'string') return Date.parse(t);
    if (typeof t.seconds !== 'undefined') return Number(t.seconds) * 1000 + Math.floor((t.nanos ?? 0) / 1e6);
    try {
        return Date.parse(String(t));
    } catch {
        return Date.now();
    }
}

export default function WalletDetailPage() {
    const { id } = useParams();
    const navigate = useNavigate();
    const { data: walletsData } = useWallets();

    const wallet = walletsData?.wallet.find((w) => w.id === id);

    const updateWallet = useUpdateWallet();
    const deleteWallet = useDeleteWallet();

    const [notify, setNotify] = useState<boolean>(wallet?.notify ?? false);
    useEffect(() => setNotify(wallet?.notify ?? false), [wallet?.notify]);

    const toggleNotify = (next: boolean) => {
        setNotify(next);
        if (!wallet) return;
        updateWallet.mutate({ id: wallet.id, label: wallet.label ?? '', notify: next });
    };

    const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
    const handleDeleteWallet = async () => {
        if (!wallet) return;
        try {
            await deleteWallet.mutateAsync(wallet.id);
            navigate('/wallets');
        } catch {
            // error shown inline via mutation state
        }
    };

    const [period, setPeriod] = useState<BalancePeriod>(BalancePeriod.BALANCE_PERIOD_1W);
    const { data: balancesData, isLoading, isError, refetch } = useWalletBalances(id, period, 100);

    // Refetch balances when the streamed wallet updates its current balance.
    // Debounce to avoid spamming the backend on rapid stream events.
    const refetchDebounceRef = useRef<number | null>(null);
    useEffect(() => {
        if (!wallet || !refetch) return;
        if (wallet.id !== id) return;

        // coalesce rapid updates
        if (refetchDebounceRef.current) {
            clearTimeout(refetchDebounceRef.current);
        }
        refetchDebounceRef.current = window.setTimeout(() => {
            void refetch();
            refetchDebounceRef.current = null;
        }, 500);

        return () => {
            if (refetchDebounceRef.current) {
                clearTimeout(refetchDebounceRef.current);
                refetchDebounceRef.current = null;
            }
        };
    }, [wallet?.balanceUsd, wallet?.balanceCrypto, wallet?.id, id, period, refetch]);

    const points = useMemo(() => {
        const list = (balancesData?.balance ?? []).map((b) => ({
            t: timeToMs((b as any).time),
            usd: (b as any).balanceUsd ?? 0,
            crypto: (b as any).balanceCrypto ?? 0,
        }));
        list.sort((a, b) => a.t - b.t);
        return list;
    }, [balancesData]);

    const data = useMemo(
        () => points.map((p) => ({ t: p.t, usd: Number((p.usd ?? 0).toFixed(6)), crypto: Number((p.crypto ?? 0).toFixed(6)) })),
        [points],
    );

    const lastBalance = useMemo(() => {
        const arr = balancesData?.balance ?? [];
        return arr.length ? arr[arr.length - 1] : undefined;
    }, [balancesData]);

    // If there's only one point, duplicate it with a small time delta so the chart
    // renders a horizontal line instead of a single dot.
    // Merge live wallet current balance (from stream) into historical data so chart updates in real-time
    const mergedData = useMemo(() => {
        const arr = [...(data ?? [])];
        if (wallet) {
            const live = { t: Date.now(), usd: wallet.balanceUsd ?? 0, crypto: wallet.balanceCrypto ?? 0 };
            // If last point is very recent (within 5 seconds), replace it; otherwise append
            const last = arr.length ? arr[arr.length - 1] : undefined;
            if (last && Math.abs(last.t - live.t) < 5_000) {
                arr[arr.length - 1] = live;
            } else {
                arr.push(live);
            }
        }
        arr.sort((a, b) => a.t - b.t);
        return arr;
    }, [data, wallet]);

    const chartData = useMemo(() => {
        if (!mergedData || mergedData.length !== 1) return mergedData;
        const single = mergedData[0];
        const delta = 24 * 60 * 60 * 1000; // 1 day
        return [
            { t: single.t - delta, usd: single.usd, crypto: single.crypto },
            { t: single.t + delta, usd: single.usd, crypto: single.crypto },
        ];
    }, [mergedData]);

    function formatCompactNumber(n: number) {
        try {
            return new Intl.NumberFormat('en-US', { notation: 'compact', maximumFractionDigits: 1 }).format(n);
        } catch {
            return String(n);
        }
    }

    const dataRangeMs = useMemo(() => {
        if (!mergedData || mergedData.length < 2) return 0;
        return mergedData[mergedData.length - 1].t - mergedData[0].t;
    }, [mergedData]);

    function formatXAxisTick(t: number) {
        const d = new Date(Number(t));
        const days = dataRangeMs ? dataRangeMs / (24 * 60 * 60 * 1000) : 0;

        // If data spans only a day or two, show times
        if (days <= 1.5) {
            return d.toLocaleTimeString([], { hour: 'numeric', minute: '2-digit' });
        }

        // If data spans up to ~2 months, show month + day (Sep 6)
        if (days <= 60) {
            return d.toLocaleDateString([], { month: 'short', day: 'numeric' });
        }

        // For longer ranges, show month + full year to avoid ambiguity (Sep 2026)
        return d.toLocaleDateString([], { month: 'short', year: 'numeric' });
    }

    function formatTooltipDate(t: number) {
        const d = new Date(Number(t));
        return d.toLocaleDateString();
    }

    function formatTooltipTime(t: number) {
        const d = new Date(Number(t));
        return d.toLocaleTimeString([], { hour: 'numeric', minute: '2-digit' });
    }

    function CustomTooltip({ active, payload, label, tokenSymbol }: any) {
        if (!active || !payload || !payload.length) return null;
        const p = payload[0].payload || {};
        const usd = typeof p.usd !== 'undefined' ? p.usd : payload[0].value;
        const crypto = typeof p.crypto !== 'undefined' ? p.crypto : p.crypto;
        return (
            <div style={{ backgroundColor: '#0f172a', border: '1px solid #ffffff15', padding: 10, borderRadius: 8, color: '#fff', boxShadow: '0 10px 15px -3px rgba(0,0,0,0.5)' }}>
                <div style={{ color: '#a855f7', fontWeight: 600, marginBottom: 6 }}>{fmtUSD(usd)}</div>
                <div style={{ color: '#94a3b8', marginBottom: 6 }}>{(crypto ?? 0).toFixed(3)} {tokenSymbol ?? ''}</div>
                <div style={{ color: '#94a3b8', fontSize: 12 }}>{formatTooltipDate(label)}</div>
                <div style={{ color: '#64748b', fontSize: 12 }}>{formatTooltipTime(label)}</div>
            </div>
        );
    }

    return (
        <div className="max-w-6xl mx-auto">
            <div className="mb-2 flex items-center justify-between">
                <div>
                    <button
                        onClick={() => navigate(-1)}
                        aria-label="Go back"
                        title="Go back"
                        className="inline-flex items-center gap-2 px-4 py-2 h-10 rounded-full bg-white/3 text-sm text-gray-300 hover:bg-white/8 hover:text-white mr-3 mb-2 transition focus:outline-none">
                        <span className="text-lg leading-none">←</span>
                        <span className="font-medium">Back</span>
                    </button>
                    <h1 className="text-xl font-semibold mb-0">{wallet ? `${wallet.tokenSymbol} · ${wallet.label || wallet.address.slice(0, 6)}` : 'Wallet'}</h1>
                    <p className="text-xs text-gray-500 mt-1">{wallet?.address}</p>
                </div>

                <div className="flex items-center gap-2">
                    <button
                        onClick={() => setShowDeleteConfirm(true)}
                        title="Delete wallet"
                        className="rounded-lg text-gray-400 hover:text-red-400 p-2 transition-colors"
                    >
                        <svg xmlns="http://www.w3.org/2000/svg" className="h-5 w-5" fill="none" viewBox="0 0 24 24" stroke="currentColor" strokeWidth={2}>
                            <path strokeLinecap="round" strokeLinejoin="round" d="M19 7l-.867 12.142A2 2 0 0116.138 21H7.862a2 2 0 01-1.995-1.858L5 7m5 4v6m4-6v6m1-10V4a1 1 0 00-1-1h-4a1 1 0 00-1 1v3M4 7h16" />
                        </svg>
                    </button>
                </div>

            </div>

            <div className="rounded-xl border border-white/5 bg-white/[0.03] p-6">
                <div>
                    {/* Full-bleed chart wrapper: remove horizontal padding by negating card padding */}
                    <div className="-mx-6">
                        <div className="px-6">
                            <div className="mb-4 flex items-start justify-between">
                                <div>
                                    <div className="text-xs text-gray-500">Balance</div>
                                    <div className="text-2xl font-semibold">{fmtUSD(wallet?.balanceUsd ?? lastBalance?.balanceUsd ?? points[points.length - 1]?.usd ?? 0)}</div>
                                    <div className="text-sm text-gray-400 mt-1">{(wallet?.balanceCrypto ?? lastBalance?.balanceCrypto ?? points[points.length - 1]?.crypto ?? 0).toFixed(6)} {wallet?.tokenSymbol ?? ''}</div>
                                </div>
                            </div>
                        </div>

                        <div className="w-full">
                            <div className="relative w-full h-64 sm:h-72 md:h-80">
                                <ResponsiveContainer width="100%" height="100%">
                                    <AreaChart data={chartData} margin={{ top: 0, right: 0, left: 0, bottom: 0 }}>
                                        <defs>
                                            <linearGradient id="colorUv" x1="0" x2="0" y1="0" y2="1">
                                                <stop offset="0%" stopColor="#7c3aed" stopOpacity={0.35} />
                                                <stop offset="100%" stopColor="#7c3aed" stopOpacity={0.03} />
                                            </linearGradient>
                                        </defs>
                                        <CartesianGrid strokeDasharray="3 3" strokeOpacity={0.03} />
                                        <XAxis
                                            dataKey="t"
                                            type="number"
                                            scale="time"
                                            domain={["dataMin", "dataMax"]}
                                            tickFormatter={formatXAxisTick}
                                            tick={{ fill: '#94a3b8', fontSize: 12 }}
                                        />
                                        <YAxis tickFormatter={(v) => formatCompactNumber(Number(v))} tick={{ fill: '#94a3b8', fontSize: 12 }} />
                                        <Tooltip content={<CustomTooltip tokenSymbol={wallet?.tokenSymbol} />} />
                                        <Area type="monotone" dataKey="usd" stroke="#7c3aed" strokeWidth={2.5} dot={false} fillOpacity={1} fill="url(#colorUv)" />
                                    </AreaChart>
                                </ResponsiveContainer>

                                {isLoading && (
                                    <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
                                        <Spinner />
                                    </div>
                                )}

                                {isError && (
                                    <div className="absolute inset-0 flex items-center justify-center p-4">
                                        <ErrorBlock message="Failed to load balances" onRetry={() => refetch()} />
                                    </div>
                                )}
                            </div>
                        </div>
                    </div>

                    <div className="mt-5 flex justify-center gap-4">
                        {PERIODS.map((p) => (
                            <button
                                key={p.key}
                                onClick={() => setPeriod(p.value)}
                                className={`rounded-lg px-3 py-1 text-xs ${period === p.value ? 'bg-purple-600 text-white' : 'text-gray-400 bg-white/5'}`}>
                                {p.label}
                            </button>
                        ))}
                    </div>
                </div>
            </div>

            <div className="mt-8">
                <label className="inline-flex items-center cursor-pointer">
                    <input
                        type="checkbox"
                        checked={notify}
                        onChange={(e) => toggleNotify(e.target.checked)}
                        className="sr-only"
                    />
                    <span className={`w-10 h-6 flex items-center rounded-full p-1 transition-colors ${notify ? 'bg-purple-600' : 'bg-white/8'}`}>
                        <span className={`bg-white w-4 h-4 rounded-full shadow transform transition-transform ${notify ? 'translate-x-4' : ''}`} />
                    </span>
                    <span className="ml-3 text-sm text-gray-300">Notify on balance changes</span>
                </label>
            </div>

            {showDeleteConfirm && (
                <Modal onClose={() => setShowDeleteConfirm(false)} title="Delete Wallet">
                    <p className="text-sm text-gray-400">Are you sure you want to delete this wallet? This action cannot be undone.</p>
                    {deleteWallet.isError && (
                        <p className="mt-2 text-xs text-red-400">{(deleteWallet.error as Error)?.message || 'Failed to delete wallet'}</p>
                    )}
                    <div className="flex items-center justify-end gap-3 mt-6">
                        <button
                            onClick={() => setShowDeleteConfirm(false)}
                            className="rounded-lg border border-white/10 px-4 py-2 text-sm text-gray-400 hover:text-white transition-colors"
                        >
                            Cancel
                        </button>
                        <button
                            onClick={() => handleDeleteWallet()}
                            disabled={deleteWallet.status === 'pending'}
                            className="rounded-lg bg-red-600 px-4 py-2 text-sm font-medium text-white hover:bg-red-500 disabled:opacity-40 disabled:cursor-not-allowed transition-colors"
                        >
                            {deleteWallet.status === 'pending' ? 'Deleting...' : 'Delete'}
                        </button>
                    </div>
                </Modal>
            )}

        </div>
    );
}
