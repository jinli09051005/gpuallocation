# FROM nvidia/cuda:12.4.1-devel-ubuntu22.04

FROM registry.cn-hangzhou.aliyuncs.com/jinli09051005/tools:golang-1.22 AS builder1
WORKDIR /app
COPY . .
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod tidy && go mod vendor
RUN GOOS=linux GOARCH=amd64 go build -o device-plugin cmd/main.go

FROM registry.cn-hangzhou.aliyuncs.com/jinli09051005/gpu:cuda-12.4.1-devel-ubuntu22.04 AS builder2
WORKDIR /app
RUN apt-get -y update && apt-get -y install cmake git
RUN git clone https://github.com/Project-HAMi/HAMi-core.git libvgpu && cd libvgpu && bash ./build.sh

FROM registry.cn-hangzhou.aliyuncs.com/jinli09051005/gpu:cuda-12.4.1-devel-ubuntu22.04
WORKDIR /app
ENV NVIDIA_DISABLE_REQUIRE="true"
ENV NVIDIA_VISIBLE_DEVICES=all
ENV NVIDIA_DRIVER_CAPABILITIES=compute,utility
COPY --from=builder1 /app/device-plugin  .
COPY --from=builder2 /app/libvgpu/build/libvgpu.so /usr/local/jinli/libvgpu.so
RUN echo "/usr/local/jinli/libvgpu.so" > /usr/local/jinli/ld.so.preload