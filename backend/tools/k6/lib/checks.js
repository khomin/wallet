import { check } from 'k6';
// import { status as grpcStatus } from 'k6/net/grpc';
import grpc, { StatusOK } from 'k6/net/grpc';

export function checkHttpResponse(res, expectedStatus = StatusOK) {
    return check(res, {
        [`HTTP status is ${expectedStatus}`]: (r) => r.status === expectedStatus,
        'HTTP response has body': (r) => r.body && r.body.length > 0,
        'HTTP wallets field exists': (r) => {
            try {
                const json = r.json();
                return Array.isArray(json.wallets);
            } catch (_) {
                return false;
            }
        },
    });
}

export function checkGrpcResponse(res, expectedStatus = StatusOK) {
    return check(res, {
        [`gRPC status is OK (${expectedStatus})`]: (r) => r.status === expectedStatus,
        'gRPC response has message': (r) => r.message !== null && r.message !== undefined,
        'gRPC wallets field exists': (r) => r.message && Array.isArray(r.message.wallets),
    });
}