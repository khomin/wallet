// ─── gRPC-HTTP Gateway Client ────────────────────────────────────────────────
// Typed fetch client for the grpc-gateway JSON API generated into /src/gen.
// It replaces the old axios REST client (services/api.ts) and keeps the same
// responsibilities:
//   1. Base URL from Vite config (default http://localhost:8080)
//   2. Attaches the Keycloak Bearer token on every request
//   3. Serialises requests / deserialises responses with protobuf-es schemas
//   4. Centralises 401 + gateway error handling

import { fromJsonString, toJsonString, create } from '@bufbuild/protobuf';
import type { DescMessage, MessageShape } from '@bufbuild/protobuf';
import { API_CONFIG } from '../config/api';

// ─── Generated message schemas + types ─────────────────────────────────────

import {
  ListWalletsResponseSchema,
  WalletUpdateSchema,
  WalletSchema,
  CreateWalletRequestSchema,
  CreateWalletResponseSchema,
  DeleteWalletResponseSchema,
  type ListWalletsResponse,
  type WalletUpdate,
  type CreateWalletRequest,
  type CreateWalletResponse,
  type DeleteWalletResponse,
} from '../gen/wallet/v1/wallet_pb';

import {
  ListCoinsResponseSchema,
  GetPricesResponseSchema,
  PriceUpdateSchema,
  type ListCoinsResponse,
  type GetPricesResponse,
  type PriceUpdate,
} from '../gen/price/v1/price_pb';

import {
  ListAlertsResponseSchema,
  CreateAlertRequestSchema,
  UpdateAlertRequestSchema,
  DeleteAlertResponseSchema,
  AlertSchema,
  type ListAlertsResponse,
  type CreateAlertRequest,
  type UpdateAlertRequest,
  type PauseAlertRequest,
  type ResumeAlertRequest,
  type DeleteAlertResponse,
  type Alert,
} from '../gen/alert/v1/alert_pb';

// ─── Types ─────────────────────────────────────────────────────────────────

type HttpMethod = 'GET' | 'POST' | 'PATCH' | 'DELETE';

export class GatewayError extends Error {
  /** grpc-gateway status code (3 = INVALID_ARGUMENT, 5 = NOT_FOUND, ...) */
  readonly grpcCode: number;
  /** HTTP status code returned by the gateway */
  readonly status: number;

  constructor(message: string, status: number, grpcCode: number) {
    super(message);
    this.name = 'GatewayError';
    this.status = status;
    this.grpcCode = grpcCode;
  }
}

// ─── Transport helpers ─────────────────────────────────────────────────────

async function requestJson<Res extends DescMessage>(
  method: HttpMethod,
  path: string,
  resSchema: Res,
  query?: URLSearchParams,
): Promise<MessageShape<Res>> {
  return doFetch(method, path, resSchema, undefined, undefined, query);
}

async function requestJsonWithBody<Req extends DescMessage, Res extends DescMessage>(
  method: HttpMethod,
  path: string,
  reqSchema: Req,
  req: MessageShape<Req>,
  resSchema: Res,
): Promise<MessageShape<Res>> {
  return doFetch(method, path, resSchema, reqSchema, req, undefined);
}

async function doFetch<Req extends DescMessage, Res extends DescMessage>(
  method: HttpMethod,
  path: string,
  resSchema: Res,
  reqSchema?: Req,
  req?: MessageShape<Req>,
  query?: URLSearchParams,
): Promise<MessageShape<Res>> {
  const url = new URL(path, API_CONFIG.baseUrl);
  if (query) url.search = query.toString();

  const headers: Record<string, string> = {};
  const token = sessionStorage.getItem('kc_access_token');
  if (token) headers.Authorization = `Bearer ${token}`;

  let body: string | undefined;
  if (reqSchema && req) {
    headers['Content-Type'] = 'application/json';
    // grpc-gateway is configured with UseProtoNames, so emit snake_case fields
    body = toJsonString(reqSchema, req, { useProtoFieldName: true });
  }

  const response = await fetch(url.toString(), { method, headers, body });

  // Token expired/invalid – mirror the old axios interceptor behaviour
  if (response.status === 401) {
    sessionStorage.clear();
    window.location.href = '/';
    throw new GatewayError('Unauthorized', 401, 16);
  }

  const text = await response.text();

  if (!response.ok) {
    throw parseGatewayError(text, response.status);
  }

  if (!text) {
    return create(resSchema);
  }

  return fromJsonString(resSchema, text, { ignoreUnknownFields: true });
}

function parseGatewayError(text: string, status: number): GatewayError {
  let message = `Request failed with status ${status}`;
  let grpcCode = 13; // INTERNAL
  try {
    const parsed = JSON.parse(text) as { message?: string; code?: number };
    if (parsed.message) message = parsed.message;
    if (typeof parsed.code === 'number') grpcCode = parsed.code;
  } catch {
    // Non-JSON error body – keep the generic message
  }
  return new GatewayError(message, status, grpcCode);
}

// ─── Wallet service ────────────────────────────────────────────────────────

export const walletService = {
  listWallets: () =>
    requestJson('GET', 'v1/wallets', ListWalletsResponseSchema),

  /**
   * Read the grpc-gateway server stream. Server-streaming responses are
   * newline-delimited JSON messages, just like the price stream above.
   */
  streamWallets: async function* (symbols: string[] = [], signal?: AbortSignal): AsyncGenerator<WalletUpdate> {
    const url = new URL('v1/wallets/stream', API_CONFIG.baseUrl);
    for (const symbol of symbols) {
      url.searchParams.append('symbols', symbol.toLowerCase());
    }

    const headers: Record<string, string> = { Accept: 'application/json' };
    const token = sessionStorage.getItem('kc_access_token');
    if (token) headers.Authorization = `Bearer ${token}`;

    const response = await fetch(url.toString(), { headers, signal });
    if (!response.ok) {
      const text = await response.text();
      throw parseGatewayError(text, response.status);
    }
    if (!response.body) throw new Error('Wallet stream is not available');

    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let buffered = '';

    const consume = (line: string): WalletUpdate | undefined => {
      const trimmed = line.trim();
      if (!trimmed) return undefined;

      const update = fromJsonString(WalletUpdateSchema, trimmed, {
        ignoreUnknownFields: true,
      });
      if (update.wallet) return update;

      // Some stream handlers serialize the WalletUpdate payload directly
      // instead of wrapping it as { wallet: ... }. Normalize that shape too.
      try {
        const raw = JSON.parse(trimmed)
        if (raw.result != undefined) {
          const payload = raw.result.wallet ?? raw;
          const wallet = fromJsonString(WalletSchema, JSON.stringify(payload), {
            ignoreUnknownFields: true,
          });
          return create(WalletUpdateSchema, { wallet });
        }
      } catch {
        return update;
      }
    };

    while (true) {
      const { done, value } = await reader.read();
      buffered += decoder.decode(value, { stream: !done });
      const lines = buffered.split(/\r?\n/);
      buffered = lines.pop() ?? '';
      for (const line of lines) {
        const update = consume(line);
        if (update) yield update;
      }
      if (done) {
        const update = consume(buffered);
        if (update) yield update;
        break;
      }
    }
  },

  createWallet: (req: MessageShape<typeof CreateWalletRequestSchema>) =>
    requestJsonWithBody(
      'POST',
      'v1/wallets',
      CreateWalletRequestSchema,
      req,
      CreateWalletResponseSchema,
    ),

  deleteWallet: (id: string) =>
    requestJson(
      'DELETE',
      `v1/wallets/${encodeURIComponent(id)}`,
      DeleteWalletResponseSchema,
    ),
};

// ─── Price service ─────────────────────────────────────────────────────────

export const priceService = {
  listCoins: () =>
    requestJson('GET', 'v1/coins', ListCoinsResponseSchema),

  getPrices: (symbols: string[]) => {
    const query = new URLSearchParams();
    for (const symbol of symbols) {
      query.append('symbols', symbol.toLowerCase());
    }
    return requestJson('GET', 'v1/prices', GetPricesResponseSchema, query);
  },

  /**
   * Read the grpc-gateway server stream. grpc-gateway represents a
   * server-streaming response as newline-delimited JSON messages.
   */
  streamPrices: async function* (symbols: string[], signal?: AbortSignal): AsyncGenerator<PriceUpdate> {
    const url = new URL('v1/prices/stream', API_CONFIG.baseUrl);
    for (const symbol of symbols) {
      url.searchParams.append('symbols', symbol.toLowerCase());
    }

    const headers: Record<string, string> = { Accept: 'application/json' };
    const token = sessionStorage.getItem('kc_access_token');
    if (token) headers.Authorization = `Bearer ${token}`;

    const response = await fetch(url.toString(), { headers, signal });
    if (!response.ok) {
      const text = await response.text();
      throw parseGatewayError(text, response.status);
    }
    if (!response.body) throw new Error('Price stream is not available');

    const reader = response.body.getReader();
    const decoder = new TextDecoder();
    let buffered = '';

    const consume = (line: string): PriceUpdate | undefined => {
      const trimmed = line.trim();
      if (!trimmed) return undefined;
      return fromJsonString(PriceUpdateSchema, trimmed, { ignoreUnknownFields: true });
    };

    while (true) {
      const { done, value } = await reader.read();
      buffered += decoder.decode(value, { stream: !done });
      const lines = buffered.split(/\r?\n/);
      buffered = lines.pop() ?? '';
      for (const line of lines) {
        const update = consume(line);
        if (update) yield update;
      }
      if (done) {
        const update = consume(buffered);
        if (update) yield update;
        break;
      }
    }
  },
};

// ─── Alert service ─────────────────────────────────────────────────────────

export const alertService = {
  listAlerts: () =>
    requestJson('GET', 'v1/alerts', ListAlertsResponseSchema),

  createAlert: (req: MessageShape<typeof CreateAlertRequestSchema>) =>
    requestJsonWithBody(
      'POST',
      'v1/alerts',
      CreateAlertRequestSchema,
      req,
      AlertSchema,
    ),

  updateAlert: (req: MessageShape<typeof UpdateAlertRequestSchema>) =>
    requestJsonWithBody(
      'PATCH',
      `v1/alerts/${encodeURIComponent(req.id)}`,
      UpdateAlertRequestSchema,
      req,
      AlertSchema,
    ),

  pauseAlert: (id: string) =>
    requestJson('POST', `v1/alerts/${encodeURIComponent(id)}:pause`, AlertSchema),

  resumeAlert: (id: string) =>
    requestJson('POST', `v1/alerts/${encodeURIComponent(id)}:resume`, AlertSchema),

  deleteAlert: (id: string) =>
    requestJson(
      'DELETE',
      `v1/alerts/${encodeURIComponent(id)}`,
      DeleteAlertResponseSchema,
    ),
};

// ─── Re-export generated types for convenience ─────────────────────────────

export type {
  ListWalletsResponse,
  CreateWalletRequest,
  CreateWalletResponse,
  DeleteWalletResponse,
  WalletUpdate,
  ListCoinsResponse,
  GetPricesResponse,
  PriceUpdate,
  ListAlertsResponse,
  CreateAlertRequest,
  UpdateAlertRequest,
  PauseAlertRequest,
  ResumeAlertRequest,
  DeleteAlertResponse,
  Alert,
};
