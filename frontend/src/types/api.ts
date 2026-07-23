// ─── API Response Types ────────────────────────────────────────────────────────
// Mirrors the Go DTOs in backend/internal/api/dto/

export interface WalletResponse {
  id: string;
  address: string;
  chain: string;
  token_symbol: string;
  label: string;
  created_at: string;
  updated_at: string;
  balance_crypto: number;
  balance_usd: number;
  change_24h_percent: number;
  has_error: boolean;
  error_msg: string;
}

export interface WalletsResponse {
  wallet: WalletResponse[];
  total: number;
  total_balance_usd: number;
}

export interface CreateWalletRequest {
  chain: string;
  address: string;
  token_symbol: string;
  label?: string;
}

export interface DeleteWalletRequest {
  id: string;
}

export interface DeleteWalletResponse {
  id: string;
}

export interface CoinResponse {
  symbol: string;
  name: string;
  image_url: string;
}

export interface CoinsResponse {
  total: number;
  coins: CoinResponse[];
}

export interface PriceResponse {
  symbol: string;
  name: string;
  price_usd: number;
  market_cap: number;
  total_volume: number;
  high_24h: number;
  low_24h: number;
  price_change_24h: number;
  price_change_percentage_24h: number;
  market_cap_change_24h: number;
  market_cap_change_percentage_24h: number;
  last_updated: string;
}

export interface PricesResponse {
  total: number;
  price: PriceResponse[];
}

export interface ErrorResponse {
  code: string;
  message: string;
}

// ─── Supported Chains (matches backend token_registry keys) ──────────────────

// TODO: pull from API

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