// ─── Settings Page ─────────────────────────────────────────────────────────
// Notification preferences, currency, Keycloak profile link.

import { useAuth } from '../auth/AuthContext';
import { Settings2 } from 'lucide-react';

export default function SettingsPage() {
  const { user, logout } = useAuth();

  const displayName =
    user?.name ?? user?.preferred_username ?? user?.email ?? 'Whale';

  return (
    <div className="max-w-6xl mx-auto space-y-6">
      <h1 className="text-xl font-semibold flex items-center gap-2">
        <Settings2 className="h-5 w-5" aria-hidden="true" />
        Settings
      </h1>

      {/* Profile card */}
      <div className="rounded-xl border border-white/5 bg-white/[0.03] p-6">
        <h2 className="text-sm font-semibold text-white mb-4">Profile</h2>
        <div className="flex items-center gap-4">
          <div className="h-12 w-12 rounded-full bg-purple-600 flex items-center justify-center text-lg font-bold">
            {displayName.charAt(0).toUpperCase()}
          </div>
          <div>
            <p className="text-white font-medium">{displayName}</p>
            <p className="text-sm text-gray-500">{user?.email ?? '—'}</p>
          </div>
        </div>
      </div>

      {/* Preferences */}
      <div className="rounded-xl border border-white/5 bg-white/[0.03] p-6">
        <h2 className="text-sm font-semibold text-white mb-4">Preferences</h2>

        <div className="space-y-4">
          {/* Wallet alerts */}
          <div className="flex items-center justify-between pt-4 border-t border-white/5">
            <div>
              <p className="text-sm text-white">Wallet movement alerts</p>
              <p className="text-xs text-gray-500">Notify on transactions</p>
            </div>
            <label className="relative inline-flex items-center cursor-pointer">
              <input type="checkbox" className="sr-only peer" checked />
              <div className="w-9 h-5 bg-gray-700 peer-focus:outline-none rounded-full peer peer-checked:after:translate-x-full rtl:peer-checked:after:-translate-x-full peer-checked:after:border-white after:content-[''] after:absolute after:top-[2px] after:start-[2px] after:bg-white after:rounded-full after:h-4 after:w-4 after:transition-all peer-checked:bg-purple-600 opacity-50" />
            </label>
          </div>
        </div>
      </div>

      {/* Danger zone */}
      <div className="rounded-xl border border-red-500/10 bg-red-500/[0.02] p-6">
        <p className="text-xs text-gray-500 mb-4">
          Sign out of your current session
        </p>
        <button
          onClick={logout}
          className="rounded-lg border border-red-500/30 px-4 py-2 text-sm text-red-400
                     hover:bg-red-500/10 transition-colors cursor-pointer"
        >
          Sign out
        </button>
      </div>
    </div>
  );
}