module soa/blog-service

go 1.24.5

require (
	github.com/golang-jwt/jwt/v5 v5.3.0
	github.com/google/uuid v1.6.0
	github.com/joho/godotenv v1.5.1
	gorm.io/driver/postgres v1.6.0
	gorm.io/gorm v1.30.0
)

require github.com/felixge/httpsnoop v1.0.3 // indirect

require (
	api-gateway v0.0.0-00010101000000-000000000000
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.27.2 // indirect
	github.com/stretchr/testify v1.10.0 // indirect
	golang.org/x/net v0.41.0 // indirect
	golang.org/x/sys v0.33.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20250818200422-3122310a409c // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20250818200422-3122310a409c // indirect
	google.golang.org/protobuf v1.36.7 // indirect
)

require (
	github.com/gorilla/handlers v1.5.2
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/pgx/v5 v5.6.0 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/lib/pq v1.10.9
	golang.org/x/crypto v0.39.0 // indirect
	golang.org/x/sync v0.16.0 // indirect
	golang.org/x/text v0.28.0 // indirect
	google.golang.org/grpc v1.75.0
)

replace api-gateway => ../api-gateway
