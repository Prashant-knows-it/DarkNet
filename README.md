go run cmd/server/main.go

go run cmd/node/main.go "Node 3" "KEY_3" 9003 127.0.0.1:8080

go run cmd/node/main.go "Node 2" "KEY_2" 9002 127.0.0.1:9003

go run cmd/node/main.go "Node 1" "KEY_1" 9001 127.0.0.1:9002

go run cmd/client/main.go
OR
Run the clinet.exe on client laptop, instead

curl.exe --proxy socks5h://127.0.0.1:10800 http://mysite.shadow


--------------------------
			node1Conn, err := net.Dial("tcp", "fcufk-27-61-117-202.run.pinggy-free.link:36063")
Change line number 90 of client to use different pinggy url
then type 
go build -o client.exe cmd/client/main.go
this will generate clinet.exe which you can share to any client
Check previous git version to see line number 90 of clinet if you want to test everything in single laptop in localhost. That will help

