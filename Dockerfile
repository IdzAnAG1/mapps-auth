FROM golang:latest as build

RUN mkdir /app

COPY . /app

WORKDIR /app

RUN GOOS="linux" GOARCH="arm64" CGO_ENABLED=0 go build -o mappsAuth ./cmd/main/main.go

FROM --platform=linux/arm64 alpine

COPY --from=build /app/mappsAuth /app/mappsAuth

WORKDIR /app

EXPOSE 8080

CMD ["/app/mappsAuth"]
