REPO=https://github.com/IdzAnAG1/mapps-contracts.git#branch=main

buf_gen:
	buf generate $(REPO) --template buf.gen.yaml --path proto/auth/v1

local:
	go run cmd/main/main.go

sql_gen:
	sqlc generate
