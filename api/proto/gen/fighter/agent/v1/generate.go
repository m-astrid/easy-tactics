package fapiv1

// Proto files location:
// - fighter-agent.proto: api/proto/fighter-agent.proto
// - ai-service.proto: python/ai-service/ai-service.proto
// Run manually:
//   cd api/proto/gen/fighter/agent/v1
//   protoc --go_out=. --go_opt=paths=source_relative --go_grpc_out=. --go-grpc_opt=paths=source_relative ../../../fighter-agent.proto
//   protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ../../../../python/ai-service/ai-service.proto
