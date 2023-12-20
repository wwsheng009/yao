# ===========================================
#  Yao Development
#  docker build \
#    --build-arg VERSION="${VERSION}"  \
#    --build-arg ARCH="${ARCH}"  \
#    -t yaoapp/yao-dev:${VERSION}-${ARCH} .
#
#  Build:
#  docker build --platform linux/amd64 --build-arg VERSION=0.10.4 --build-arg ARCH=amd64 -t wwsheng009/yao:0.10.4-amd64-dev .
#  docker build --platform linux/arm64 --build-arg VERSION=0.10.4 --build-arg ARCH=arm64 -t wwsheng009/yao:0.10.4-arm64-dev .
#
#  Tests:
#  docker run --rm wwsheng009/yao:0.10.4-amd64-dev yao version
#  docker run -d -p 5099:5099 wwsheng009/yao:0.10.4-amd64-dev
#
# ===========================================
FROM ubuntu:20.04
ARG VERSION
ARG ARCH
RUN groupadd -r yao && useradd -r -g yao yao && \
    apt-get update && \
    apt-get install -y curl sudo procps net-tools
# https://github.com/wwsheng009/yao/releases/download/yao-0.10.4/yao-linux-0.10.4-release.tar.gz
RUN curl -fsSL "https://github.com/wwsheng009/yao/releases/download/yao-${VERSION}/yao-linux-${VERSION}-release.tar.gz" \
    --output "yao-linux-release.tar.gz"

RUN tar -xzf "yao-linux-release.tar.gz" && \
    mv "yao*-linux-${ARCH}" > /usr/local/bin/yao && \
    chmod +x /usr/local/bin/yao && \
    mkdir -p /data/app
# Clean up
RUN rm yao-linux-release.tar.gz

VOLUME /data/app
WORKDIR /data/app
EXPOSE 5099
CMD ["tail", "-f","/dev/null"]
