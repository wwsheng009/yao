# ===========================================
#  Yao Build Environment (Ubuntu 23.04 AMD64)
#
#  Build:
#  docker build --platform linux/amd64 -t yaoapp/yao-build:0.10.4 .
#
#  Usage:
#  docker run --rm -it -v /local/path/dist:/data yaoapp/yao-build:0.10.4
#
#  Tests:
#  docker run --rm -it yaoapp/yao-build:0.10.4 /bin/bash
#  docker run --rm -it -v ./test:/data yaoapp/yao-build:0.10.4 /bin/bash
#
# ===========================================
FROM ubuntu:24.04
WORKDIR /app
ADD build.sh /app/build.sh
ENV PATH=$PATH:/usr/local/go/bin:/root/go/bin

# RUN cp /etc/apt/sources.list /etc/apt/sources.list.backup && \
#     echo "deb http://archive.ubuntu.com/ubuntu/ jammy main restricted universe multiverse" | tee -a /etc/apt/sources.list && \
#     echo "deb http://archive.ubuntu.com/ubuntu/ jammy-updates main restricted universe multiverse" | tee -a /etc/apt/sources.list && \
#     echo "deb http://archive.ubuntu.com/ubuntu/ jammy-security main restricted universe multiverse" | tee -a /etc/apt/sources.list && \
#     echo "deb http://archive.ubuntu.com/ubuntu/ jammy-backports main restricted universe multiverse" | tee -a /etc/apt/sources.list

RUN apt-get update && \
    apt-get install -y libc6-armel-cross libc6-dev-armel-cross binutils-arm-linux-gnueabi libncurses5-dev build-essential bison flex libssl-dev bc && \
    apt-get install -y gcc-arm-linux-gnueabi g++-arm-linux-gnueabi && \
    apt-get install -y gcc-arm-linux-gnueabihf g++-arm-linux-gnueabihf  && \
    apt-get install -y g++-aarch64-linux-gnu crossbuild-essential-arm64  && \
    apt-get install -y gcc-13-aarch64-linux-gnu  && \
    apt-get install -y g++-13-aarch64-linux-gnu  && \
    apt-get install -y wget  && \
    apt-get install -y curl  && \
    apt-get install -y git && \
    apt-get install -y unzip


RUN wget https://golang.org/dl/go1.21.1.linux-amd64.tar.gz && \
    tar -C /usr/local -xzf go1.21.1.linux-amd64.tar.gz && \
    rm go1.21.1.linux-amd64.tar.gz

RUN curl -fsSL https://deb.nodesource.com/setup_18.x | bash - && \
    apt-get install -y nodejs 

RUN npm install -g pnpm

RUN curl -s https://packagecloud.io/install/repositories/github/git-lfs/script.deb.sh | bash && \
    apt-get install -y git-lfs && \
    git lfs install

# RUN npm install -g pnpm
RUN chmod +x /app/build.sh

VOLUME [ "/data" ]
CMD ["/app/build.sh"]