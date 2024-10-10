# FROM nvidia/cuda:12.3.2-devel-ubuntu22.04

FROM registry.cn-hangzhou.aliyuncs.com/jinli09051005/tools:golang-1.21 AS builder
WORKDIR /app
COPY . .
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod tidy && go mod vendor
RUN GOOS=linux GOARCH=amd64 go build -o jinli-dijkstra-api cmd/main.go

FROM registry.cn-hangzhou.aliyuncs.com/jinli09051005/gpu:cuda-12.3.2-devel-ubuntu22.04
WORKDIR /app
ENV NVIDIA_DISABLE_REQUIRE="true"
ENV NVIDIA_VISIBLE_DEVICES=all
ENV NVIDIA_DRIVER_CAPABILITIES=compute,utility
COPY --from=builder /app/device-plugin  .