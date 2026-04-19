#!/bin/bash

# Tok Deployment Script
# Automates deployment to staging and production environments

set -e  # Exit on error

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'  # No Color

# Configuration
ENVIRONMENT=${1:-staging}
REGION=${2:-us-central1}
PROJECT_ID=${3:-tok-project}
NAMESPACE=tok-${ENVIRONMENT}

# Functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[✓]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[!]${NC} $1"
}

log_error() {
    echo -e "${RED}[✗]${NC} $1"
}

check_prerequisites() {
    log_info "Checking prerequisites..."

    # Check required tools
    for tool in docker kubectl gcloud; do
        if ! command -v $tool &> /dev/null; then
            log_error "$tool is not installed"
            exit 1
        fi
    done

    log_success "All prerequisites met"
}

build_docker_image() {
    log_info "Building Docker image..."

    # Build image
    docker build \
        -t gcr.io/${PROJECT_ID}/tok:latest \
        -t gcr.io/${PROJECT_ID}/tok:$(git rev-parse --short HEAD) \
        -f Dockerfile .

    log_success "Docker image built"
}

push_docker_image() {
    log_info "Pushing Docker image to GCR..."

    # Authenticate with GCR
    gcloud auth configure-docker

    # Push images
    docker push gcr.io/${PROJECT_ID}/tok:latest
    docker push gcr.io/${PROJECT_ID}/tok:$(git rev-parse --short HEAD)

    log_success "Docker image pushed"
}

create_namespace() {
    log_info "Creating Kubernetes namespace..."

    if kubectl get namespace $NAMESPACE &> /dev/null; then
        log_warning "Namespace $NAMESPACE already exists"
    else
        kubectl create namespace $NAMESPACE
        log_success "Namespace $NAMESPACE created"
    fi

    kubectl config set-context --current --namespace=$NAMESPACE
}

create_secrets() {
    log_info "Creating Kubernetes secrets..."

    # Check if secrets already exist
    if kubectl get secret tok-secrets -n $NAMESPACE &> /dev/null; then
        log_warning "Secrets already exist, skipping creation"
        return
    fi

    # Get secrets from environment
    API_KEY=${TOK_API_KEY:-""}
    DB_PASSWORD=${TOK_DB_PASSWORD:-""}
    JWT_SECRET=${TOK_JWT_SECRET:-""}

    if [ -z "$API_KEY" ] || [ -z "$DB_PASSWORD" ]; then
        log_error "Missing required environment variables"
        log_error "Set: TOK_API_KEY, TOK_DB_PASSWORD, TOK_JWT_SECRET"
        exit 1
    fi

    # Create secrets
    kubectl create secret generic tok-secrets \
        --from-literal=api-key="$API_KEY" \
        --from-literal=db-password="$DB_PASSWORD" \
        --from-literal=jwt-secret="$JWT_SECRET" \
        -n $NAMESPACE

    log_success "Kubernetes secrets created"
}

create_configmap() {
    log_info "Creating Kubernetes ConfigMap..."

    # Check if ConfigMap already exists
    if kubectl get configmap tok-config -n $NAMESPACE &> /dev/null; then
        log_warning "ConfigMap already exists, updating..."
        kubectl delete configmap tok-config -n $NAMESPACE
    fi

    # Create ConfigMap
    kubectl create configmap tok-config \
        --from-literal=log-level=info \
        --from-literal=compression-level=aggressive \
        --from-literal=cache-ttl=3600 \
        --from-literal=environment=$ENVIRONMENT \
        -n $NAMESPACE

    log_success "Kubernetes ConfigMap created"
}

deploy_kubernetes() {
    log_info "Deploying to Kubernetes..."

    # Update image in deployment manifest
    sed -e "s|IMAGE|gcr.io/${PROJECT_ID}/tok:$(git rev-parse --short HEAD)|g" \
        deployments/kubernetes/tok-deployment.yaml \
        | kubectl apply -f - -n $NAMESPACE

    log_success "Kubernetes manifests applied"
}

wait_for_deployment() {
    log_info "Waiting for deployment to be ready..."

    # Wait for rollout
    kubectl rollout status deployment/tok-api -n $NAMESPACE --timeout=5m

    log_success "Deployment is ready"
}

verify_deployment() {
    log_info "Verifying deployment..."

    # Get service IP
    SERVICE_IP=$(kubectl get svc tok-api -n $NAMESPACE \
        -o jsonpath='{.status.loadBalancer.ingress[0].ip}' 2>/dev/null || echo "pending")

    if [ "$SERVICE_IP" == "pending" ]; then
        log_warning "Service IP still pending (this is normal for new deployments)"
        log_info "Run: kubectl get svc -n $NAMESPACE"
        return
    fi

    # Test health endpoint
    log_info "Testing health endpoint..."

    if curl -s http://${SERVICE_IP}:8083/health | grep -q "ok"; then
        log_success "Health check passed"
    else
        log_warning "Health check failed, but deployment may still be initializing"
    fi
}

run_integration_tests() {
    if [ "$ENVIRONMENT" != "production" ]; then
        log_info "Running integration tests..."

        # Run tests
        go test -v -timeout 30s ./internal/integration/...

        log_success "Integration tests passed"
    fi
}

run_load_test() {
    if [ "$ENVIRONMENT" == "staging" ]; then
        read -p "Run load test? (y/n) " -n 1 -r
        echo

        if [[ $REPLY =~ ^[Yy]$ ]]; then
            log_info "Running load test..."

            # Get service IP
            SERVICE_IP=$(kubectl get svc tok-api -n $NAMESPACE \
                -o jsonpath='{.status.loadBalancer.ingress[0].ip}')

            # Run k6 load test
            k6 run deployments/load-test.js \
                -e API_ENDPOINT="http://${SERVICE_IP}:8083"

            log_success "Load test completed"
        fi
    fi
}

setup_monitoring() {
    log_info "Setting up monitoring..."

    # Check if Prometheus is already deployed
    if kubectl get deployment prometheus -n $NAMESPACE &> /dev/null; then
        log_warning "Prometheus already deployed"
        return
    fi

    # Deploy monitoring stack (optional)
    # This would deploy Prometheus, Grafana, etc.
    log_info "Monitoring setup skipped (deploy separately if needed)"
}

print_summary() {
    log_info "=========================================="
    log_info "Deployment Summary"
    log_info "=========================================="
    log_info "Environment: $ENVIRONMENT"
    log_info "Namespace: $NAMESPACE"
    log_info "Project: $PROJECT_ID"
    log_info "Region: $REGION"
    log_info ""

    # Get deployment info
    log_info "Deployment Status:"
    kubectl get deployment -n $NAMESPACE

    log_info ""
    log_info "Pod Status:"
    kubectl get pods -n $NAMESPACE

    log_info ""
    log_info "Service Status:"
    kubectl get svc -n $NAMESPACE

    log_info ""
    log_info "Next steps:"
    log_info "1. Monitor deployment: kubectl logs -f -l app=tok -n $NAMESPACE"
    log_info "2. Access dashboard: kubectl port-forward svc/tok-dashboard 3000:3000 -n $NAMESPACE"
    log_info "3. View metrics: kubectl port-forward svc/prometheus 9090:9090 -n $NAMESPACE"
    log_info "=========================================="
}

rollback() {
    log_warning "Rolling back deployment..."

    kubectl rollout undo deployment/tok-api -n $NAMESPACE
    kubectl rollout status deployment/tok-api -n $NAMESPACE

    log_success "Rollback completed"
}

# Main deployment flow
main() {
    log_info "Starting Tok deployment to $ENVIRONMENT"

    # Pre-flight checks
    check_prerequisites

    # Build and push image
    build_docker_image
    push_docker_image

    # Kubernetes setup
    create_namespace
    create_secrets
    create_configmap

    # Deploy
    deploy_kubernetes
    wait_for_deployment

    # Verify
    verify_deployment
    run_integration_tests
    run_load_test
    setup_monitoring

    # Summary
    print_summary

    log_success "Deployment completed successfully!"
}

# Error handling
trap 'log_error "Deployment failed!"; exit 1' ERR

# Run main function
if [ "${BASH_SOURCE[0]}" == "${0}" ]; then
    main "$@"
fi
