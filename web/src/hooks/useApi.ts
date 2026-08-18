// ─── API Hooks (TanStack Query) ──────────────────────────────────────────────
// Custom hooks for every backend endpoint. Each hook manages caching,
// loading / error states, and automatic refetching.
//
// The underlying client now talks to the grpc-gateway JSON API and uses the
// protobuf types generated into /src/gen.

import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useEffect } from 'react';
import { create } from '@bufbuild/protobuf';
import type { MessageInitShape } from '@bufbuild/protobuf';
import { walletService, priceService, alertService } from '../services/grpcGateway';
import { CreateWalletRequestSchema } from '../gen/wallet/v1/wallet_pb';
import { CreateAlertRequestSchema } from '../gen/alert/v1/alert_pb';
import { GetPricesResponseSchema } from '../gen/price/v1/price_pb';
import type {
  ListWalletsResponse,
  CreateWalletResponse,
  DeleteWalletResponse,
  ListCoinsResponse,
  GetPricesResponse,
  ListAlertsResponse,
  Alert,
  DeleteAlertResponse,
} from '../services/grpcGateway';

// ─── Query key factory ─────────────────────────────────────────────────────
export const queryKeys = {
  wallets: ['wallets'] as const,
  coins: ['coins'] as const,
  alerts: ['alerts'] as const,
  prices: (symbols: string[]) => ['prices', ...symbols] as const,
};

// ─── Wallets ───────────────────────────────────────────────────────────────

/** Fetch all wallets for the current user */
export function useWallets() {
  return useQuery<ListWalletsResponse>({
    queryKey: queryKeys.wallets,
    queryFn: () => walletService.listWallets(),
    // Refetch every 30s so balances stay reasonably fresh
    refetchInterval: 30_000,
  });
}

/** Create a new wallet */
export function useCreateWallet() {
  const qc = useQueryClient();
  return useMutation<
    CreateWalletResponse,
    Error,
    MessageInitShape<typeof CreateWalletRequestSchema>
  >({
    mutationFn: async (req) => {
      const message = create(CreateWalletRequestSchema, req);
      return walletService.createWallet(message);
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.wallets });
    },
  });
}

/** Delete a wallet */
export function useDeleteWallet() {
  const qc = useQueryClient();
  return useMutation<DeleteWalletResponse, Error, string>({
    mutationFn: (id) => walletService.deleteWallet(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.wallets });
    },
  });
}

// ─── Coins ─────────────────────────────────────────────────────────────────

/** Fetch the list of all supported coins */
export function useCoins() {
  return useQuery<ListCoinsResponse>({
    queryKey: queryKeys.coins,
    queryFn: () => priceService.listCoins(),
    // Coin list rarely changes – keep it fresh enough but not aggressively
    staleTime: 5 * 60 * 1000,
  });
}

// ─── Prices ────────────────────────────────────────────────────────────────

/** Fetch prices for specific symbols */
export function usePrices(symbols: string[]) {
  const queryClient = useQueryClient();
  const symbolsKey = symbols.join(',');
  const query = useQuery<GetPricesResponse>({
    queryKey: queryKeys.prices(symbols),
    queryFn: () => priceService.getPrices(symbols),
    enabled: symbols.length > 0,
    refetchInterval: 30_000,
  });

  // Keep the initial REST response above as the source of truth for loading
  // and errors, then merge each streamed update into the same query cache.
  useEffect(() => {
    if (!symbolsKey) return;

    const controller = new AbortController();
    let mounted = true;

    const run = async () => {
      while (mounted) {
        try {
          for await (const update of priceService.streamPrices(
            symbolsKey.split(','),
            controller.signal,
          )) {
            if (!mounted) return;
            queryClient.setQueryData<GetPricesResponse>(
              queryKeys.prices(symbolsKey.split(',')),
              (current) => {
                const streamedBySymbol = new Map(
                  update.price.map((price) => [price.symbol.toLowerCase(), price]),
                );
                const currentPrices = current?.price ?? [];
                const merged = currentPrices.map(
                  (price) => streamedBySymbol.get(price.symbol.toLowerCase()) ?? price,
                );
                const existingSymbols = new Set(
                  currentPrices.map((price) => price.symbol.toLowerCase()),
                );
                for (const price of update.price) {
                  if (!existingSymbols.has(price.symbol.toLowerCase())) merged.push(price);
                }
                return current
                  ? { ...current, total: merged.length, price: merged }
                  : create(GetPricesResponseSchema, { total: merged.length, price: merged });
              },
            );
          }
        } catch {
          if (controller.signal.aborted || !mounted) return;
        }
        // Reconnect if the server closes the stream or a transient error occurs.
        await new Promise((resolve) => setTimeout(resolve, 1000));
      }
    };

    void run();
    return () => {
      mounted = false;
      controller.abort();
    };
  }, [queryClient, symbolsKey]);

  return query;
}

// ─── Alerts ────────────────────────────────────────────────────────────────

/** Fetch all alerts for the current user */
export function useAlerts() {
  return useQuery<ListAlertsResponse>({
    queryKey: queryKeys.alerts,
    queryFn: () => alertService.listAlerts(),
    refetchInterval: 30_000,
  });
}

/** Create a new alert */
export function useCreateAlert() {
  const qc = useQueryClient();
  return useMutation<
    Alert,
    Error,
    MessageInitShape<typeof CreateAlertRequestSchema>
  >({
    mutationFn: async (req) => {
      const message = create(CreateAlertRequestSchema, req);
      return alertService.createAlert(message);
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.alerts });
    },
  });
}

/** Delete an alert */
export function useDeleteAlert() {
  const qc = useQueryClient();
  return useMutation<DeleteAlertResponse, Error, string>({
    mutationFn: (id) => alertService.deleteAlert(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.alerts });
    },
  });
}