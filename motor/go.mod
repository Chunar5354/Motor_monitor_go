module github.com/my/repo

go 1.15

require (
	github.com/go-redis/redis/v8 v8.4.4 // indirect
	github.com/go-sql-driver/mysql v1.5.0 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	motor v0.0.0
)

replace motor => ./motor
