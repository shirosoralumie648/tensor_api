# Author: ProgramZmh
# License: Apache-2.0
# Description: Dockerfile for chatnio

FROM --platform=$TARGETPLATFORM golang:1.24-alpine AS backend

WORKDIR /backend
COPY . .

# Set go proxy to https://goproxy.cn (open for vps in China Mainland)
# RUN go env -w GOPROXY=https://goproxy.cn,direct
ARG TARGETARCH
ARG TARGETOS
ENV GOOS=$TARGETOS GOARCH=$TARGETARCH GO111MODULE=on CGO_ENABLED=1
# Go mirrors
ENV GOPROXY=https://goproxy.cn,direct
ENV GOSUMDB=off

# Install build dependencies (with mirror switch and retry)
RUN set -eux; \
    sed -i 's#https://dl-cdn.alpinelinux.org#https://mirrors.tencent.com#g' /etc/apk/repositories || true; \
    for i in 1 2 3; do apk update && break || (sleep 2); done; \
    for i in 1 2 3; do apk add --no-cache \
        gcc \
        musl-dev \
        g++ \
        make \
        linux-headers \
        libwebp-dev \
      && break || (sleep 2); done

# Build backend (disable static linking to avoid CGO static issues)
RUN go env -w GOPROXY=$GOPROXY GOSUMDB=$GOSUMDB && \
    go mod download && \
    go mod tidy && \
    go build -o chat -a .

FROM node:18 AS frontend

WORKDIR /app
COPY ./app .

RUN npm config set registry https://registry.npmmirror.com && \
    npm install -g pnpm && \
    pnpm config set registry https://registry.npmmirror.com && \
    pnpm install && \
    pnpm run build && \
    rm -rf node_modules src


FROM alpine

# Install dependencies (with mirror switch and retry)
RUN set -eux; \
    sed -i 's#https://dl-cdn.alpinelinux.org#https://mirrors.tencent.com#g' /etc/apk/repositories || true; \
    for i in 1 2 3; do apk update && break || (sleep 2); done; \
    for i in 1 2 3; do apk upgrade --no-cache && break || (sleep 2); done; \
    for i in 1 2 3; do apk add --no-cache wget ca-certificates tzdata libwebp && break || (sleep 2); done; \
    update-ca-certificates 2>/dev/null || true

# Set timezone
RUN echo "Asia/Shanghai" > /etc/timezone && \
    ln -sf /usr/share/zoneinfo/Asia/Shanghai /etc/localtime

WORKDIR /

# Copy dist
COPY --from=backend /backend/chat /chat
COPY --from=backend /backend/config.example.yaml /config.example.yaml
COPY --from=backend /backend/utils/templates /utils/templates
COPY --from=backend /backend/addition/article/template.docx /addition/article/template.docx
COPY --from=frontend /app/dist /app/dist

# Volumes
VOLUME ["/config", "/logs", "/storage"]

# Expose port
EXPOSE 8094

# Run application
CMD ["./chat"]
