package main

// TODO: make a plugin for vscode
// This generates the go code from .proto files
// For simple cases like this it avoids the need to have a Makefile

//go:generate protoc --go_opt=module=dlmsprocessor --go-grpc_opt=module=dlmsprocessor --go_out=./ --go-grpc_out=./ --proto_path=./ dlmsprocessor.proto

//go:generate protoc --go_opt=module=dlmsprocessor --go-grpc_opt=module=dlmsprocessor --go_out=../dlms_consumer/ --go-grpc_out=../dlms_consumer/ --proto_path=./ dlmsprocessor.proto

// Proto to TS for use in UI
// This generates JS code from .proto files
//go:generate protoc --ts_opt=no_namespace --ts_opt=unary_rpc_promise=true --ts_opt=target=web --ts_out=../dlms_ui/src/proto/ --proto_path=./ dlmsprocessor.proto

// Replace grpc-js with grpc-web for browser compatibility
//go:generate sh -c "sed -i 's|@grpc/grpc-js|grpc-web|g' ../dlms_ui/proto/dlmsprocessor.ts"

// Remove server-side abstract service class completely (between the class declaration and its closing brace)
//go:generate sh -c "sed -i '/^export abstract class.*Service {$/,/^}$/c\\// Server-side service class removed for client-side compatibility' ../dlms_ui/proto/dlmsprocessor.ts"

// Remove lines containing server-side types that don't exist in grpc-web
//go:generate sh -c "sed -i '/UntypedHandleCall/d; /ServerWritableStream/d; /ServerUnaryCall/d; /sendUnaryData/d' ../dlms_ui/proto/dlmsprocessor.ts"

// Add @ts-nocheck at the beginning to suppress any remaining TypeScript errors
//go:generate sh -c "sed -i '1i\\// @ts-nocheck' ../dlms_ui/proto/dlmsprocessor.ts"

// // go:generate sqlc -f db/scripts/sqlc.yaml generate
