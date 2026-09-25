import { check } from 'k6';
import { getGrpcClient } from '../../lib/clients.js';
import { listWalletsCheck } from '../../lib/checks.js';

const client = getGrpcClient();

export const options = {
    vus: 1, duration: '30s',
    thresholds: { grpc_req_duration: ['p(95)<200'] },
};

export default function () {
    const res = client.invoke('WalletService/ListWallets', {});
    check(res, listWalletsCheck);
}