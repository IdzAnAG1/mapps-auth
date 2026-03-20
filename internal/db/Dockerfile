FROM golang:latest as build

RUN GOOS="linux" GOARCH="amd64" CGO_ENABLED=0 go install github.com/pressly/goose/v3/cmd/goose@latest

FROM --platform=linux/amd64 alpine

COPY --from=build /go/bin/linux_amd64/goose /usr/local/bin/goose

COPY ./internal/db/migrations ./migrations

ENV GOOSE_DRIVER=postgres

CMD ["goose", "-dir", "./migrations", "up"]
