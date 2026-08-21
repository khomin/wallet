// ─── Frontend API / domain types ──────────────────────────────────────────────
// Protobuf-generated API types now live in /src/gen and are re-exported from
// src/services/grpcGateway.ts. This file only keeps UI-specific constants/types.

// Supported assets and their networks are loaded from the backend with useCoins.

/** Form state used by the Add Wallet modal. */
export interface CreateWalletFormState {
  chains: string[];
  address: string;
  tokenSymbol: string;
  label: string;
}