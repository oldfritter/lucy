FROM golang:1.25.4 AS builder
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn
COPY . /app
RUN cd /app && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o cmd/schedule service/schedule/main.go

FROM alpine
ENV TZ=Asia/Shanghai
LABEL cn.oldfritter.image.authors="leon@oldfritter.cn"
WORKDIR /app
COPY config/img /app/config/img
COPY config/locale /app/config/locale
COPY config/ttf /app/config/ttf
COPY --from=builder /app/cmd/schedule /app/cmd/schedule
ENTRYPOINT /app/cmd/schedule