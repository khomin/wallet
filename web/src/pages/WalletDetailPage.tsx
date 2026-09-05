import { useState, useMemo } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Spinner, ErrorBlock } from '../components/ui';
import { useWallets, useWalletBalances, useCoins } from '../hooks/useApi';
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
    const { data: coinsData } = useCoins();

    const wallet = walletsData?.wallet.find((w) => w.id === id);

    const [period, setPeriod] = useState<BalancePeriod>(BalancePeriod.BALANCE_PERIOD_1W);
    const { data: balancesData, isLoading, isError, refetch } = useWalletBalances(id, period, 1000);

    const points = useMemo(() => {
        const list = (balancesData?.balance ?? []).map((b) => ({
            t: timeToMs((b as any).time),
            v: b.balanceUsd ?? b.balanceCrypto ?? 0,
        }));
        list.sort((a, b) => a.t - b.t);
        return list;
    }, [balancesData]);

    const data = useMemo(
        () => points.map((p) => ({ t: p.t, value: Number((p.v ?? 0).toFixed(6)) })),
        [points],
    );

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

            </div>

            <div className="rounded-xl border border-white/5 bg-white/[0.03] p-6">
                {isLoading && <Spinner />}
                {isError && <ErrorBlock message="Failed to load balances" onRetry={() => refetch()} />}
                {!isLoading && !isError && (
                    <div>
                        {/* Full-bleed chart wrapper: remove horizontal padding by negating card padding */}
                        <div className="-mx-6">
                            <div className="px-6">
                                <div className="mb-4 flex items-start justify-between">
                                    <div>
                                        <div className="text-xs text-gray-500">Balance</div>
                                        <div className="text-2xl font-semibold">{fmtUSD(points[points.length - 1]?.v ?? 0)}</div>
                                    </div>
                                </div>
                            </div>

                            <div className="w-full">
                                <ResponsiveContainer width="100%" height={260}>
                                    <AreaChart data={data} margin={{ top: 0, right: 0, left: 0, bottom: 0 }}>
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
                                            hide
                                        />
                                        {/* hide Y axis labels/lines to remove left/right legend */}
                                        <YAxis hide />
                                        <Tooltip
                                            contentStyle={{
                                                backgroundColor: '#0f172a',
                                                borderColor: '#ffffff15',
                                                borderRadius: 8,
                                                color: '#fff',
                                                boxShadow: '0 10px 15px -3px rgba(0, 0, 0, 0.5)',
                                            }}
                                            itemStyle={{ color: '#a855f7' }}
                                            labelStyle={{ color: '#94a3b8', fontSize: 12 }}
                                            labelFormatter={(t) => new Date(Number(t)).toLocaleString()}
                                            formatter={(v: any) => [fmtUSD(v)]}
                                        />
                                        <Area type="monotone" dataKey="value" stroke="#7c3aed" strokeWidth={2.5} dot={false} fillOpacity={1} fill="url(#colorUv)" />
                                    </AreaChart>
                                </ResponsiveContainer>
                            </div>
                        </div>

                        <div className="mt-4 flex justify-center gap-2">
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
                )}
            </div>
        </div>
    );
}
