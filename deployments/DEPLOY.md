# TokMan Production Deployment Guide

## Quick Start (5 minutes)

### Local Development
```bash
cd tokman
docker-compose -f deployments/docker-compose.yaml up -d

# Verify services
curl http://localhost:8083/health
open http://localhost:3000  # Dashboard
open http://localhost:9090  # Prometheus
open http://localhost:3001  # Grafana
```

## Staging Deployment (Kubernetes)

### Prerequisites
```bash
# Install tools
brew install kubectl helm
gcloud auth login  # or: aws configure, az login
gcloud container clusters get-credentials tokman-staging --zone us-central1-a
```

### Deploy to GKE
```bash
# 1. Create namespace
kubectl create namespace tokman-staging
kubectl config set-context --current --namespace=tokman-staging

# 2. Build & push image
docker build -t gcr.io/PROJECT_ID/tokman:latest .
docker push gcr.io/PROJECT_ID/tokman:latest

# 3. Deploy manifests
kubectl apply -f deployments/kubernetes/

# 4. Verify deployment
kubectl rollout status deployment/tokman-api
kubectl get pods -l app=tokman
kubectl get svc tokman-api

# 5. Check endpoints
API_IP=$(kubectl get svc tokman-api -o jsonpath='{.status.loadBalancer.ingress[0].ip}')
curl http://$API_IP:8083/health
```

### Expose Dashboard
```bash
# Option 1: Port forward (dev only)
kubectl port-forward svc/tokman-dashboard 3000:3000

# Option 2: Ingress (production)
kubectl apply -f - <<EOF
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: tokman-ingress
spec:
  rules:
  - host: tokman-staging.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: tokman-dashboard
            port:
              number: 3000
EOF
```

## Load Testing (10K Concurrent Users)

### Setup Load Test Environment
```bash
# 1. Install Grafana k6
brew install k6

# 2. Create test scenario
cat > deployments/load-test.js << 'EOF'
import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  stages: [
    { duration: '2m', target: 100 },   // Ramp-up to 100 users
    { duration: '5m', target: 1000 },  // Ramp-up to 1K users
    { duration: '5m', target: 10000 }, // Ramp-up to 10K users
    { duration: '5m', target: 10000 }, // Stay at 10K
    { duration: '2m', target: 0 },     // Ramp-down to 0
  ],
  thresholds: {
    http_req_duration: ['p(99)<500'],  // 99% under 500ms
    http_req_failed: ['rate<0.1'],     // Error rate < 0.1%
  },
};

export default function () {
  let response = http.post('http://localhost:8083/analyze', {
    code: `func fibonacci(n int) int {
      if n <= 1 { return n }
      return fibonacci(n-1) + fibonacci(n-2)
    }`,
  }, {
    headers: { 'Authorization': 'Bearer test-token' },
  });

  check(response, {
    'status is 200': (r) => r.status === 200,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });

  sleep(1);
}
EOF

# 3. Run load test
k6 run deployments/load-test.js

# 4. View results
# Check Grafana dashboard for metrics
open http://localhost:3001 
# Login: admin/admin (change in production!)
# View: TokMan > Request Metrics > Duration/Throughput
```

## Monitoring Setup

### Prometheus (Auto-scraped)
```bash
# Metrics available at:
# http://localhost:9091/metrics

# Key metrics to monitor:
# - tokman_tokens_processed_total
# - tokman_compression_ratio_average
# - tokman_requests_duration_seconds
# - tokman_active_users
# - tokman_cache_hit_ratio
```

### Grafana Dashboards
```bash
# 1. Login to Grafana
open http://localhost:3001
# Default: admin/admin

# 2. Add Prometheus datasource
# Configuration > Data Sources > Add Prometheus
# URL: http://prometheus:9090

# 3. Import dashboard
# Dashboards > Import > ID: 1860 (Prometheus)
# Create custom dashboard with:
#   - Request Rate (req/s)
#   - Response Time (p50/p95/p99)
#   - Token Metrics (processed/saved)
#   - Compression Ratio
#   - Error Rate
#   - Active Connections
```

### Alerting Rules
```yaml
groups:
  - name: tokman.rules
    rules:
    - alert: HighErrorRate
      expr: rate(tokman_errors_total[5m]) > 0.05
      for: 5m
      annotations:
        summary: "High error rate detected"

    - alert: SlowResponse
      expr: histogram_quantile(0.99, rate(tokman_requests_duration_seconds[5m])) > 1
      for: 5m
      annotations:
        summary: "P99 response time > 1s"

    - alert: LowCacheHitRate
      expr: tokman_cache_hit_ratio < 0.75
      for: 10m
      annotations:
        summary: "Cache hit rate < 75%"

    - alert: NearQuotaLimit
      expr: tokman_token_budget_used > 0.8
      for: 5m
      annotations:
        summary: "Team using 80% of token quota"
```

## Production Deployment (AWS/GCP/Azure)

### AWS EKS
```bash
# Create cluster
eksctl create cluster --name tokman-prod --region us-east-1 --nodes 3 --node-type t3.large

# Deploy
kubectl apply -f deployments/kubernetes/

# Configure autoscaling
kubectl autoscale deployment tokman-api --min=5 --max=50 --cpu-percent=70
```

### GCP Cloud Run
```bash
# Build image
gcloud builds submit --tag gcr.io/PROJECT_ID/tokman:latest

# Deploy
gcloud run deploy tokman-api \
  --image gcr.io/PROJECT_ID/tokman:latest \
  --platform managed \
  --region us-central1 \
  --memory 2Gi \
  --cpu 2 \
  --concurrency 100 \
  --min-instances 5 \
  --max-instances 50
```

### Azure Container Instances
```bash
# Create container
az container create \
  --resource-group tokman \
  --name tokman-api \
  --image myregistry.azurecr.io/tokman:latest \
  --cpu 2 \
  --memory 2 \
  --ports 8083
```

## Database Migration

### PostgreSQL Setup
```bash
# Create database
createdb tokman_production

# Run migrations
psql tokman_production < internal/tracking/migrations.sql

# Create read replicas (high availability)
# Via cloud console for RDS/Cloud SQL/Azure Database for PostgreSQL
```

## Secrets Management

### Kubernetes Secrets
```bash
# Create secrets
kubectl create secret generic tokman-secrets \
  --from-literal=api-key=$API_KEY \
  --from-literal=db-password=$DB_PASSWORD \
  --from-literal=jwt-secret=$JWT_SECRET

# Verify
kubectl get secrets
```

### Environment Configuration
```bash
kubectl create configmap tokman-config \
  --from-literal=log-level=info \
  --from-literal=compression-level=aggressive \
  --from-literal=cache-ttl=3600
```

## Health Checks & Readiness

### Liveness Probe
```bash
curl -f http://localhost:8083/health || exit 1
```

### Readiness Probe
```bash
curl -f http://localhost:8083/ready || exit 1
```

These should both return 200 OK before serving traffic.

## Backup & Disaster Recovery

### Database Backups
```bash
# Automated daily backups via cloud provider
# Retention: 30 days
# RPO: 24 hours
# RTO: 1 hour (restore from backup)

# Test restore
pg_dump tokman_production > backup.sql
psql tokman_test < backup.sql
```

### Config Sync Recovery
```bash
# Device sync conflicts resolved via:
# 1. Latest version hash comparison
# 2. User manual resolution UI
# 3. Fallback to last known good config
```

## Performance Tuning

### Connection Pooling
```go
// cmd/server/main.go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

### Redis Caching
```go
// Every filter result cached for TTL
cache.Set(ctx, cacheKey, result, 1*time.Hour)
```

### Compression Optimization
- SIMD acceleration for pattern matching (if available)
- Lazy evaluation of low-confidence filters
- Early exit on threshold met

## Rollout Strategy

### Canary Deployment
```bash
# 1. Deploy new version to 5% of traffic
kubectl set image deployment/tokman-api \
  tokman=gcr.io/PROJECT_ID/tokman:v1.1.0 \
  --record

# 2. Monitor error rate for 5 minutes
# If < 0.1% errors → proceed to 25%
# Otherwise → rollback

# 3. Gradual rollout: 5% → 25% → 50% → 100%

# 4. Automatic rollback on alert
kubectl rollout undo deployment/tokman-api
```

### Blue-Green Deployment
```bash
# Deploy new version alongside old
# Switch traffic via ingress
kubectl patch service tokman-api -p '{"spec":{"selector":{"version":"v1.1"}}}'
```

## Monitoring Checklist

- [ ] All pods running and ready
- [ ] Load balancer distributing traffic
- [ ] Database replication lag < 100ms
- [ ] Cache hit ratio > 80%
- [ ] P99 response time < 500ms
- [ ] Error rate < 0.1%
- [ ] CPU utilization 30-70%
- [ ] Memory utilization 40-60%
- [ ] Disk space > 20% free
- [ ] Log streams flowing to aggregator

## Troubleshooting

### Pods not starting
```bash
kubectl describe pod <pod-name>
kubectl logs <pod-name>
```

### High response times
```bash
# Check database performance
kubectl exec -it <pod> -- psql $DATABASE_URL
select * from pg_stat_statements order by mean_exec_time desc limit 5;

# Check cache hit rate
curl http://localhost:9091/metrics | grep cache_hit
```

### Memory leaks
```bash
# Check pprof profiles
go tool pprof http://localhost:6060/debug/pprof/heap
```

---

**Next: Publish VSCode Extension & GitHub Action, Launch Beta (Week 3-4)**
