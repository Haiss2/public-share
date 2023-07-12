# For future ws testing
go run ./ -gather-data-duration=30s

# For spot ws testing
go run ./ -gather-data-duration=30s -websocket=wss://stream.binance.com:9443/ws -file-prefix=spot
