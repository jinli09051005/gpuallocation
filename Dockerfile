# FROM nvidia/cuda:12.3.2-devel-ubuntu22.04

FROM registry.cn-hangzhou.aliyuncs.com/jinli09051005/tools:golang-1.21 AS builder
WORKDIR /app
COPY . .
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod tidy && go mod vendor
RUN GOOS=linux GOARCH=amd64 go build -o jinli-dijkstra-api cmd/main.go
RUN apt-get -y update && apt-get -y install cmake
RUN git clone https://github.com/Project-HAMi/HAMi-core.git libvgpu && cd libvgpu && bash ./build.sh

FROM registry.cn-hangzhou.aliyuncs.com/jinli09051005/gpu:cuda-12.3.2-devel-ubuntu22.04
WORKDIR /app
ENV NVIDIA_DISABLE_REQUIRE="true"
ENV NVIDIA_VISIBLE_DEVICES=all
ENV NVIDIA_DRIVER_CAPABILITIES=compute,utility
COPY --from=builder /app/device-plugin  .
COPY --from=builder /app/libvgpu/build/libvgpu.so /usr/local/jinli/libvgpu.so
RUN echo "/usr/local/jinli/libvgpu.so" > /usr/local/jinli/ld.so.preload