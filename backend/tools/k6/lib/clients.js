import grpc from 'k6/net/grpc';
import { CONFIG } from './config.js';

export function getGrpcClient() {
    const client = new grpc.Client();
    client.load(CONFIG.IMPORT_PATHS, 'wallet/v1/wallet.proto');
    return client;
}

export function connectGrpcClient(client) {
    client.connect('localhost:50051', { plaintext: true });
    return client;
}


// client.connect('localhost:50051', { plaintext: true });
// client.load(['../../protos/wallet/v1'], 'wallet.proto');
// client.load(['../../protos/alert/v1'], 'alert.proto');
// client.load(['../../protos/price/v1'], 'price.proto');
// client.load(['../../protos/user/v1'], 'user.proto');
// client.load(['../../../../../proto/wallet/v1'], 'wallet.proto');
// client.load(['../../protos/google/api']);
// client.loadProtoset('../../protos/')
// client.connect('localhost:50051', { plaintext: true });