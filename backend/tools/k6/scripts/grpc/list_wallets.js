// import { check } from 'k6';
// import { getGrpcClient } from '../../lib/clients.js';
// import { listWalletsCheck } from '../../lib/checks.js';

// const client = getGrpcClient();

// export const options = {
//     vus: 1, duration: '30s',
//     thresholds: { grpc_req_duration: ['p(95)<200'] },
// };

// export default function () {
//     const res = client.invoke('WalletService/ListWallets', {});
//     check(res, listWalletsCheck);
// }

import grpc from 'k6/net/grpc';
import { CONFIG } from '../../lib/config.js';
import { checkGrpcResponse } from '../../lib/checks.js';
import { connectGrpcClient, getGrpcClient } from '../../lib/clients.js';

const client = getGrpcClient();
let isConnected = false;

export function listWalletsGrpc(token = null) {
    if (!isConnected) {
        connectGrpcClient(client)
        isConnected = true;
    }
    const params = token ? { metadata: { authorization: `Bearer ${token}` } } : {};
    const res = client.invoke('wallet.v1.WalletService/ListWallets', {}, params);

    checkGrpcResponse(res);
    return res;
}

export function closeGrpcClient() {
    client.close();
}