// ─── Alerts Page (Placeholder) ────────────────────────────────────────────

export default function AlertsPage() {
  return (
    <div className="max-w-6xl mx-auto">
      <h1 className="text-xl font-semibold mb-6">🔔 Alerts</h1>

      <div className="rounded-xl border border-white/5 bg-white/[0.03] p-6">
        <div className="flex flex-col items-center justify-center py-16 text-center">
          <span className="text-4xl mb-4">🚧</span>
          <p className="text-gray-400 text-sm font-medium">Coming soon</p>
          <p className="text-gray-600 text-xs mt-1 max-w-sm">
            Price alerts and wallet activity notifications will appear here.
            Configure thresholds for balance changes, whale movements, and price swings.
          </p>
        </div>
      </div>
    </div>
  );
}