#!/bin/bash

# Kubernetes 部署脚本 - Oblivious AI Platform
# 用于在 Kubernetes 集群中部署完整的微服务架构

set -e

GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m'

NAMESPACE="oblivious"
KUBE_DIR="./kubernetes"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}  Oblivious K8s 部署工具  ${NC}"
echo -e "${BLUE}========================================${NC}\n"

# ==================== 前置检查 ====================
echo -e "${YELLOW}[1/6] 前置条件检查${NC}\n"

# 检查 kubectl
if command -v kubectl &> /dev/null; then
    echo -e "${GREEN}✓${NC} kubectl 已安装: $(kubectl version --client --short 2>&1 | head -1)"
else
    echo -e "${RED}✗${NC} kubectl 未安装，请先安装 Kubernetes CLI"
    exit 1
fi

# 检查集群连接
if kubectl cluster-info &> /dev/null; then
    echo -e "${GREEN}✓${NC} 集群连接正常"
    kubectl cluster-info | grep "running at"
else
    echo -e "${RED}✗${NC} 无法连接到 Kubernetes 集群"
    echo -e "  请检查 kubeconfig 配置或启动 Minikube/Kind"
    exit 1
fi

# 检查镜像
echo -e "\n${CYAN}检查 Docker 镜像...${NC}"
for image in oblivious-backend:latest oblivious-frontend:latest; do
    if docker images $image | grep -q "$image"; then
        echo -e "${GREEN}✓${NC} $image 存在"
    else
        echo -e "${YELLOW}⚠${NC}  $image 不存在，将在部署时拉取"
    fi
done

# ==================== 创建命名空间 ====================
echo -e "\n${YELLOW}[2/6] 创建命名空间和配置${NC}\n"

if kubectl get namespace $NAMESPACE &> /dev/null; then
    echo -e "${YELLOW}⚠${NC}  命名空间 $NAMESPACE 已存在"
    read -p "是否删除并重新创建？(y/N) " -n 1 -r
    echo
    if [[ $REPLY =~ ^[Yy]$ ]]; then
        echo "删除命名空间 $NAMESPACE ..."
        kubectl delete namespace $NAMESPACE --timeout=60s
        sleep 5
    fi
fi

if ! kubectl get namespace $NAMESPACE &> /dev/null; then
    echo "创建命名空间 $NAMESPACE ..."
    kubectl apply -f $KUBE_DIR/namespace.yaml
    echo -e "${GREEN}✓${NC} 命名空间已创建"
fi

# ==================== 部署数据层 ====================
echo -e "\n${YELLOW}[3/6] 部署数据层服务${NC}\n"

echo "部署 PostgreSQL ..."
kubectl apply -f $KUBE_DIR/postgres.yaml
sleep 2

echo "部署 Redis ..."
kubectl apply -f $KUBE_DIR/redis.yaml
sleep 2

echo -e "${CYAN}等待数据库就绪 (30秒)...${NC}"
kubectl wait --for=condition=ready pod \
    -l app=postgres \
    -n $NAMESPACE \
    --timeout=60s || echo "PostgreSQL 启动超时，继续部署"

kubectl wait --for=condition=ready pod \
    -l app=redis \
    -n $NAMESPACE \
    --timeout=30s || echo "Redis 启动超时，继续部署"

# ==================== 部署后端服务 ====================
echo -e "\n${YELLOW}[4/6] 部署后端微服务${NC}\n"

echo "部署后端服务 (Gateway, User, Chat, Relay, Agent, KB) ..."
kubectl apply -f $KUBE_DIR/backend-services.yaml
sleep 5

echo -e "${CYAN}等待后端服务就绪...${NC}"
for service in gateway user chat relay agent kb; do
    echo -n "  等待 $service ... "
    kubectl wait --for=condition=ready pod \
        -l app=oblivious-$service \
        -n $NAMESPACE \
        --timeout=60s 2>/dev/null && echo -e "${GREEN}✓${NC}" || echo -e "${YELLOW}超时${NC}"
done

# ==================== 部署前端 ====================
echo -e "\n${YELLOW}[5/6] 部署前端服务${NC}\n"

echo "部署前端 ..."
kubectl apply -f $KUBE_DIR/frontend.yaml
sleep 5

echo -e "${CYAN}等待前端就绪...${NC}"
kubectl wait --for=condition=ready pod \
    -l app=oblivious-frontend \
    -n $NAMESPACE \
    --timeout=90s || echo "前端启动超时"

# ==================== 部署 Ingress ====================
echo -e "\n${YELLOW}[6/6] 配置入口${NC}\n"

echo "部署 Ingress ..."
kubectl apply -f $KUBE_DIR/ingress.yaml || echo "Ingress 部署失败，可能需要 Ingress Controller"

# ==================== 部署状态 ====================
echo -e "\n${BLUE}========================================${NC}"
echo -e "${BLUE}  部署完成  ${NC}"
echo -e "${BLUE}========================================${NC}\n"

echo -e "${CYAN}Pod 状态:${NC}"
kubectl get pods -n $NAMESPACE -o wide

echo -e "\n${CYAN}Service 状态:${NC}"
kubectl get svc -n $NAMESPACE

echo -e "\n${CYAN}Ingress 状态:${NC}"
kubectl get ingress -n $NAMESPACE 2>/dev/null || echo "未配置 Ingress"

# ==================== 访问信息 ====================
echo -e "\n${YELLOW}访问信息:${NC}"

# 获取服务端口
GATEWAY_PORT=$(kubectl get svc oblivious-gateway -n $NAMESPACE -o jsonpath='{.spec.ports[0].nodePort}' 2>/dev/null || echo "N/A")
FRONTEND_PORT=$(kubectl get svc oblivious-frontend -n $NAMESPACE -o jsonpath='{.spec.ports[0].nodePort}' 2>/dev/null || echo "N/A")

# 获取节点 IP (Minikube)
if command -v minikube &> /dev/null && minikube status &> /dev/null; then
    NODE_IP=$(minikube ip)
    echo -e "  前端: http://$NODE_IP:$FRONTEND_PORT"
    echo -e "  API:  http://$NODE_IP:$GATEWAY_PORT"
elif kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="ExternalIP")].address}' | grep -q .; then
    NODE_IP=$(kubectl get nodes -o jsonpath='{.items[0].status.addresses[?(@.type=="ExternalIP")].address}')
    echo -e "  前端: http://$NODE_IP:$FRONTEND_PORT"
    echo -e "  API:  http://$NODE_IP:$GATEWAY_PORT"
else
    echo -e "  使用端口转发访问:"
    echo -e "  ${CYAN}kubectl port-forward -n $NAMESPACE svc/oblivious-frontend 3000:3000${NC}"
    echo -e "  ${CYAN}kubectl port-forward -n $NAMESPACE svc/oblivious-gateway 8080:8080${NC}"
fi

# ==================== 管理命令 ====================
echo -e "\n${YELLOW}常用管理命令:${NC}"
echo -e "  查看 Pod: ${CYAN}kubectl get pods -n $NAMESPACE${NC}"
echo -e "  查看日志: ${CYAN}kubectl logs -f <pod-name> -n $NAMESPACE${NC}"
echo -e "  进入容器: ${CYAN}kubectl exec -it <pod-name> -n $NAMESPACE -- sh${NC}"
echo -e "  端口转发: ${CYAN}kubectl port-forward -n $NAMESPACE svc/oblivious-gateway 8080:8080${NC}"
echo -e "  删除部署: ${CYAN}kubectl delete namespace $NAMESPACE${NC}"

echo -e "\n${GREEN}✅ Kubernetes 部署完成！${NC}\n"

# ==================== 可选：健康检查 ====================
read -p "是否执行健康检查？(Y/n) " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Nn]$ ]]; then
    echo -e "\n${YELLOW}执行健康检查...${NC}\n"
    
    # 端口转发 Gateway
    kubectl port-forward -n $NAMESPACE svc/oblivious-gateway 8080:8080 &
    PF_PID=$!
    sleep 3
    
    # 测试 API
    if curl -sf http://localhost:8080/health > /dev/null 2>&1; then
        echo -e "${GREEN}✓${NC} API 网关健康检查通过"
    else
        echo -e "${RED}✗${NC} API 网关健康检查失败"
    fi
    
    # 停止端口转发
    kill $PF_PID 2>/dev/null || true
fi

exit 0
