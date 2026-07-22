// ─── API Hooks (TanStack Query) ──────────────────────────────────────────────
// Custom hooks for every backend endpoint. Each hook manages caching,
// loading / error states, and automatic refetching.

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import api from '../services/api';
import type {
  WalletsResponse,
  WalletResponse,
  CreateWalletRequest,
  DeleteWalletRequest,
  DeleteWalletResponse,
  CoinsResponse,
  PricesResponse,
} from '../types/api';

// ─── Query key factory ─────────────────────────────────────────────────────
export const queryKeys = {
  wallets: ['wallets'] as const,
  coins: ['coins'] as const,
  prices: (symbols: string[]) => ['prices', ...symbols] as const,
};

// ─── Wallets ───────────────────────────────────────────────────────────────

/** Fetch all wallets for the current user */
export function useWallets() {
  return useQuery<WalletsResponse>({
    queryKey: queryKeys.wallets,
    queryFn: async () => {
      const { data } = await api.get<WalletsResponse>('/api/v1/wallets');
      return data;
    },
    // Refetch every 30s so balances stay reasonably fresh
    refetchInterval: 30_000,
  });
}

/** Create a new wallet */
export function useCreateWallet() {
  const qc = useQueryClient();
  return useMutation<WalletResponse, Error, CreateWalletRequest>({
    mutationFn: async (req) => {
      const { data } = await api.post<WalletResponse>('/api/v1/wallets', req);
      return data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.wallets });
    },
  });
}

/** Delete a wallet */
export function useDeleteWallet() {
  const qc = useQueryClient();
  return useMutation<DeleteWalletResponse, Error, DeleteWalletRequest>({
    mutationFn: async (req) => {
      const { data } = await api.delete<DeleteWalletResponse>('/api/v1/wallets', {
        data: req,
      });
      return data;
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.wallets });
    },
  });
}

// ─── Coins ─────────────────────────────────────────────────────────────────

/** Fetch the list of all supported coins */
export function useCoins() {
  return useQuery<CoinsResponse>({
    queryKey: queryKeys.coins,
    queryFn: async () => {
      const { data } = await api.get<CoinsResponse>('/api/v1/coins');
      return data;
    },
    // Coin list rarely changes – keep it fresh enough but not aggressively
    staleTime: 5 * 60 * 1000,
  });
}

// ─── Prices ────────────────────────────────────────────────────────────────

/** Fetch prices for specific symbols */
export function usePrices(symbols: string[]) {
  const joined = symbols.map((s) => s.toLowerCase()).join(',');
  return useQuery<PricesResponse>({
    queryKey: queryKeys.prices(symbols),
    queryFn: async () => {
      const { data } = await api.get<PricesResponse>('/api/v1/prices', {
        params: { symbols: joined },
      });
      return data;
    },
    enabled: symbols.length > 0,
    refetchInterval: 30_000,
  });
}