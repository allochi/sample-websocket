http:
	caddy file-server --root public

server:
	go run .
