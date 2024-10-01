# 使用 CentOS 作为基础镜像
FROM centos:latest

# 设置环境变量
ENV GO_VERSION=1.22.7
ENV GO_TAR=go${GO_VERSION}.linux-amd64.tar.gz
ENV GO_URL=https://golang.org/dl/${GO_TAR}
ENV GOPATH=/go
ENV PATH=$PATH:/usr/local/go/bin:$GOPATH/bin

# 安装必要的工具
RUN yum -y update && \
    yum -y install wget tar

# 下载并安装 Go
RUN wget ${GO_URL} && \
    tar -C /usr/local -xzf ${GO_TAR} && \
    rm -f ${GO_TAR}

# 创建 Go 工作目录
RUN mkdir /app

# 设置工作目录
WORKDIR /app

COPY ./main /app/main

# 验证 Go 安装
RUN go version

# 运行一个空的命令以保持容器运行
CMD ["go ", "run", "main"]