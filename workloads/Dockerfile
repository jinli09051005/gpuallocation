FROM registry.cn-hangzhou.aliyuncs.com/jinli09051005/gpu:pytorch-22.04-py3
WORKDIR /app
COPY ./workloads/ .
RUN echo 'deb http://mirrors.aliyun.com/ubuntu/ jammy main' >> /etc/apt/sources.list && apt update && apt install libc6 -y
RUN pip3 install -r requires.txt
