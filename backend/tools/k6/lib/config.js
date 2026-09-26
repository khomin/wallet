export const CONFIG = {
    HTTP_BASE_URL: __ENV.HTTP_BASE_URL || 'http://localhost:8080',
    GRPC_TARGET: __ENV.GRPC_TARGET || 'localhost:50051',
    AUTH_URL: __ENV.AUTH_URL || 'http://localhost:8080/v1/auth/token',

    IMPORT_PATHS: [
        '../proto'
        // '/Users/panic/Documents/PROJECTS/wallet/proto/price',
        // '../../../../../',
        // '../../',
        // '../../../../',
        // '../../../../proto',
        // '../../../../proto/alert/v1/',
        // '../../../../proto/price/v1/',
        // '../../../../proto/user/v1/',
        // '../../../../../proto'
        // backend/tools/k6/protos/alert/v1/alert.proto
    ],
};

