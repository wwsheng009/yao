# ===========================================
#  Yao Development
#  docker build \
#    --build-arg VERSION="${VERSION}"  \
#    --build-arg ARCH="${ARCH}"  \
#    -t yaoapp/yao-dev:${VERSION}-${ARCH} .
#
#  Build:
#  docker build --platform linux/amd64 --build-arg VERSION=0.10.5 --build-arg ARCH=amd64 -t wwsheng009/yao:0.10.5-amd64-dev .
#  docker build --platform linux/arm64 --build-arg VERSION=0.10.5 --build-arg ARCH=arm64 -t wwsheng009/yao:0.10.5-arm64-dev .
#
#  Tests:
#  docker run --rm wwsheng009/yao:0.10.5-amd64-dev yao version
#  docker run -d -p 5099:5099 wwsheng009/yao:0.10.5-amd64-dev
#
# ===========================================
FROM alpine:latest AS builder
ARG VERSION
ARG ARCH=amd64
RUN apk --no-cache add curl 
# curl -fSL https://github.com/wwsheng009/yao/releases/download/yao-0.10.5/yao-linux-0.10.5-release.tar.gz --output "yao-linux-release.tar.gz"
RUN curl -fSL "https://github.com/wwsheng009/yao/releases/download/yao-${VERSION}/yao-${VERSION}-linux-${ARCH}.tar.gz" \
    --output "yao-linux-release.tar.gz" && \
    tar -xzf "yao-linux-release.tar.gz" && \
    mv yao*-linux-"${ARCH}" /usr/local/bin/yao && \
    chmod +x /usr/local/bin/yao && \
    rm -rf yao-*

FROM alpine:latest
RUN apk add --no-cache tzdata

ENV TZ=Asia/Shanghai
# if the env tz do not work
RUN ln -snf /usr/share/zoneinfo/$TZ /etc/localtime && echo $TZ > /etc/timezone
COPY --from=builder /usr/local/bin/yao /usr/local/bin/yao

RUN mkdir -p /data/app
VOLUME /data/app
WORKDIR /data/app
EXPOSE 5099
CMD ["/usr/local/bin/yao", "start"]