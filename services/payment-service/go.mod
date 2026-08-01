module github.com/t-kiattisak/event-driven-shortlink-payment/services/payment-service

go 1.22

replace github.com/t-kiattisak/event-driven-shortlink-payment/proto => ../../proto

require (
	github.com/gofiber/fiber/v2 v2.52.4
	github.com/gofiber/template/html/v2 v2.1.1
	github.com/google/uuid v1.6.0
	github.com/segmentio/kafka-go v0.4.47
	github.com/skip2/go-qrcode v0.0.0-20200617195104-da1b6568686e
	google.golang.org/grpc v1.65.0
	gorm.io/driver/postgres v1.5.7
	gorm.io/gorm v1.25.10
)

require (
	golang.org/x/net v0.25.0 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20240528184218-531527333157 // indirect
	google.golang.org/protobuf v1.34.1 // indirect
)

require (
	github.com/andybalholm/brotli v1.0.5 // indirect
	github.com/gofiber/template v1.8.3 // indirect
	github.com/gofiber/utils v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/pgx/v5 v5.4.3 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/klauspost/compress v1.17.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mattn/go-runewidth v0.0.15 // indirect
	github.com/pierrec/lz4/v4 v4.1.15 // indirect
	github.com/rivo/uniseg v0.2.0 // indirect
	github.com/t-kiattisak/event-driven-shortlink-payment/proto v0.0.0
	github.com/valyala/bytebufferpool v1.0.0 // indirect
	github.com/valyala/fasthttp v1.51.0 // indirect
	github.com/valyala/tcplisten v1.0.0 // indirect
	golang.org/x/crypto v0.23.0 // indirect
	golang.org/x/sys v0.20.0 // indirect
	golang.org/x/text v0.15.0 // indirect
)
