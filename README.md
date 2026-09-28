go run cmd/server/main.go
go run cmd/node/main.go "Node 3" "KEY_3" 9003 127.0.0.1:8080
go run cmd/node/main.go "Node 2" "KEY_2" 9002 127.0.0.1:9003
go run cmd/node/main.go "Node 1" "KEY_1" 9001 127.0.0.1:9002
go run cmd/client/main.go
curl.exe --proxy socks5h://127.0.0.1:10800 http://mysite.shadow
