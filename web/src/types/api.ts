// ─── Frontend API / domain types ──────────────────────────────────────────────
// Protobuf-generated API types now live in /src/gen and are re-exported from
// src/services/grpcGateway.ts. This file only keeps UI-specific constants/types.

// ─── Supported Chains (matches backend token_registry keys) ──────────────────

// TODO: pull from API

// btc 3LYJfcfHPXYJreMsASk2jkn69LWEYKzexb

export const SUPPORTED_CHAINS = [
  { value: 'ETH', label: 'Ethereum', icon: 'Ξ' },
  { value: 'ARB', label: 'Arbitrum', icon: '🔷' },
  { value: 'BASE', label: 'Base', icon: '🔵' },
  { value: 'POLYGON', label: 'Polygon', icon: '🟣' },
  { value: 'BNB', label: 'BNB Chain', icon: '🟡' },
  { value: 'SOL', label: 'Solana', icon: '◎' },
  { value: 'TRX', label: 'Tron', icon: '🔴' },
] as const;

export type ChainValue = (typeof SUPPORTED_CHAINS)[number]['value'];

/** Form state used by the Add Wallet modal. */
export interface CreateWalletFormState {
  chain: string;
  address: string;
  tokenSymbol: string;
  label: string;
}