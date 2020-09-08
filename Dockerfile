FROM golang:1.13.5-alpine3.10 AS builder
WORKDIR /build

ENV GOPROXY https://goproxy.cn
COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -o -a fuploadServer .


FROM alpine:3.10 AS final
WORKDIR /app
COPY --from=builder /build/fuploadServer  /app
COPY --from=builder /build/conf/fupload.yml  /app

EXPOSE 8090
ENTRYPOINT ["/app/fuploadServer"]
CMD ["-c","/app/fupload.yml"]