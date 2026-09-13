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