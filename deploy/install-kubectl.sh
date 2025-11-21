#!/bin/bash

# kubectl 安装脚本 - Linux

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${YELLOW}安装 kubectl...${NC}\n"

# 方法 1: 使用 snap（推荐，最简单）
if command -v snap &> /dev/null; then
    echo "使用 snap 安装..."
    sudo snap install kubectl --classic
    
    if kubectl version --client &> /dev/null; then
        echo -e "${GREEN}✅ kubectl 安装成功（via snap）${NC}"
        kubectl version --client
        exit 0
    fi
fi

# 方法 2: 使用 apt
echo "使用 apt 安装..."

# 添加 K8s 仓库
sudo apt-get update
sudo apt-get install -y apt-transport-https ca-certificates curl gnupg

# 添加 K8s GPG key
sudo mkdir -p -m 755 /etc/apt/keyrings
curl -fsSL https://pkgs.k8s.io/core:/stable:/v1.31/deb/Release.key | sudo gpg --dearmor -o /etc/apt/keyrings/kubernetes-apt-keyring.gpg

# 添加 K8s 仓库
echo 'deb [signed-by=/etc/apt/keyrings/kubernetes-apt-keyring.gpg] https://pkgs.k8s.io/core:/stable:/v1.31/deb/ /' | sudo tee /etc/apt/sources.list.d/kubernetes.list

# 安装 kubectl
sudo apt-get update
sudo apt-get install -y kubectl

# 验证
if kubectl version --client &> /dev/null; then
    echo -e "${GREEN}✅ kubectl 安装成功（via apt）${NC}"
    kubectl version --client
    exit 0
else
    echo -e "${RED}❌ kubectl 安装失败${NC}"
    exit 1
fi
