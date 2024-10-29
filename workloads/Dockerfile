FROM registry.cn-hangzhou.aliyuncs.com/jinli09051005/tools:golang-1.22 AS builder
WORKDIR /app
COPY . .
ENV GO111MODULE=on
ENV GOPROXY=https://goproxy.cn,direct
RUN go mod tidy && go mod vendor
RUN GOOS=linux GOARCH=amd64 go build -o gpu-app workloads/main.go

FROM registry.cn-hangzhou.aliyuncs.com/jinli09051005/gpu:cuda-12.4.1-base-ubuntu22.04
WORKDIR /app
ENV ASSUME_NO_MOVING_GC_UNSAFE_RISK_IT_WITH=go1.22
COPY --from=builder /app/gpu-app  .