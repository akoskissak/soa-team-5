module api-gateway

go 1.24.5

require (
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.2
	google.golang.org/genproto/googleapis/api v0.0.0-20250818200422-3122310a409c
	google.golang.org/grpc v1.75.0
	google.golang.org/protobuf v1.36.7
)

require github.com/felixge/httpsnoop v1.0.3 // indirect

require (
	github.com/gorilla/handlers v1.5.2
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250818200422-3122310a409c // indirect
)
