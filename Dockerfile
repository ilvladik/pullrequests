FROM golang:1.25.2

WORKDIR /app

COPY . .

RUN go build -o /build ./cmd/server \
    && go clean -cache -modcache

EXPOSE 8080

CMD ["/build"]
