# TokMan Quick Start Guide

## 🚀 Local Development (5 minutes)

```bash
# 1. Clone and navigate
git clone https://github.com/GrayCodeAI/tokman
cd tokman

# 2. Start all services
docker-compose -f deployments/docker-compose.yaml up -d

# 3. Verify services
curl http://localhost:8083/health          # API
open http://localhost:3000                 # Dashboard
open http://localhost:9090                 # Prometheus
```

## 📊 Environment Variables

```bash
# Create .env file
cat > .env << EOF
# Database
DATABASE_URL=postgres://user:password@localhost:5432/tokman

# API
API_PORT=8083
API_ENDPOINT=http://localhost:8083

# Tokens & Auth
TOKMAN_API_KEY=dev-key-local
TOKMAN_TEAM_ID=team-dev

# Features
COMPRESSION_LEVEL=aggressive
CACHE_TTL=3600
LOG_LEVEL=info

# Cloud (optional)
GCP_PROJECT_ID=tokman-project
AWS_REGION=us-central1
EOF

# Source environment
export $(cat .env | xargs)
```

## 🧪 Testing

### Run All Tests
```bash
# Unit tests
go test ./internal/...

# Integration tests
go test ./internal/integration/...

# Load tests (requires k6)
k6 run deployments/load-test.js
```

### Quick Integration Test
```bash
# Start server
go run cmd/server/main.go

# In another terminal
curl -X POST http://localhost:8083/analyze \
  -H "Authorization: Bearer test-key" \
  -H "Content-Type: application/json" \
  -d '{"code":"func test() {}","language":"go"}'
```

## 📦 Deployment

### Staging (Kubernetes)
```bash
# One-command deploy
./scripts/deploy.sh staging us-central1 tokman-project

# Or manually:
kubectl create namespace tokman-staging
kubectl apply -f deployments/kubernetes/ -n tokman-staging
kubectl rollout status deployment/tokman-api -n tokman-staging
```

### Production
```bash
# Same script, different environment
./scripts/deploy.sh production us-central1 tokman-project

# Verify
kubectl get pods -n tokman-production
kubectl logs -f deployment/tokman-api -n tokman-production
```

## 📈 Monitoring

### Prometheus Metrics
```bash
# Port forward
kubectl port-forward svc/prometheus 9090:9090

# Query key metrics
curl http://localhost:9090/api/v1/query?query=tokman_tokens_processed_total
```

### Grafana Dashboards
```bash
# Port forward
kubectl port-forward svc/grafana 3001:3000

# Login: admin/admin
# Add Prometheus datasource: http://prometheus:9090
```

### View Logs
```bash
# Tail logs from deployment
kubectl logs -f deployment/tokman-api -n tokman-production

# Or search across all pods
kubectl logs -l app=tokman --all-containers=true -n tokman-production
```

## 🐛 Troubleshooting

### Pod Not Starting
```bash
# Check pod events
kubectl describe pod <pod-name> -n tokman-production

# Check logs
kubectl logs <pod-name> -n tokman-production

# Check resources
kubectl top pods -n tokman-production
```

### High Response Times
```bash
# Check database
kubectl exec -it <pod> -- psql $DATABASE_URL
select * from pg_stat_statements order by mean_exec_time desc;

# Check cache
kubectl exec -it <pod> -- redis-cli info stats
```

### Memory Issues
```bash
# Check memory usage
kubectl top nodes
kubectl top pods -n tokman-production

# Increase limits in deployment
kubectl set resources deployment tokman-api \
  -n tokman-production \
  --limits=memory=2Gi,cpu=2000m
```

## 💡 Common Tasks

### Check Service Status
```bash
# All pods
kubectl get pods -n tokman-production

# All services
kubectl get svc -n tokman-production

# Deployment details
kubectl describe deployment tokman-api -n tokman-production
```

### Scale Deployment
```bash
# Manual scale
kubectl scale deployment tokman-api --replicas=5 -n tokman-production

# Or edit deployment
kubectl edit deployment tokman-api -n tokman-production
```

### Update Image
```bash
# New image version
kubectl set image deployment/tokman-api \
  tokman=gcr.io/tokman-project/tokman:v1.1.0 \
  -n tokman-production

# Check rollout status
kubectl rollout status deployment/tokman-api -n tokman-production

# Rollback if needed
kubectl rollout undo deployment/tokman-api -n tokman-production
```

### Access Dashboard
```bash
# Port forward
kubectl port-forward svc/tokman-dashboard 3000:3000 -n tokman-production

# Open browser
open http://localhost:3000
```

## 🔑 API Examples

### Analyze Code
```bash
curl -X POST https://api.tokman.dev/analyze \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "func fibonacci(n int) int { return n }",
    "language": "go",
    "compression_level": "aggressive"
  }'
```

### Batch Analysis
```bash
curl -X POST https://api.tokman.dev/analyze-batch \
  -H "Authorization: Bearer YOUR_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "files": [
      {"name": "file1.go", "code": "..."},
      {"name": "file2.go", "code": "..."}
    ]
  }'
```

### Get Analytics
```bash
curl -X GET 'https://api.tokman.dev/analytics/stats?start_date=2026-04-01' \
  -H "Authorization: Bearer YOUR_API_KEY"
```

## 📚 Key Files

### Source Code
- `cmd/server/main.go` - Main API server
- `internal/core/pipeline.go` - Compression pipeline
- `internal/learning/engine.go` - Adaptive learning
- `cmd/dashboard/` - React frontend

### Configuration
- `deployments/docker-compose.yaml` - Local development
- `deployments/kubernetes/` - Production k8s manifests
- `deployments/load-test.js` - Load testing script

### Documentation
- `DEPLOY.md` - Full deployment guide
- `BUSINESS_MODEL.md` - Financial projections
- `LAUNCH_CHECKLIST.md` - Launch planning
- `PROJECT_STATUS.md` - Complete status report

## 🎯 Useful Commands

```bash
# Check health
curl http://localhost:8083/health

# Get metrics
curl http://localhost:8083/metrics

# Watch deployment
kubectl rollout status deployment/tokman-api -n tokman-production -w

# Port forward API
kubectl port-forward svc/tokman-api 8083:8083 -n tokman-production

# Port forward Dashboard
kubectl port-forward svc/tokman-dashboard 3000:3000 -n tokman-production

# View recent logs
kubectl logs deployment/tokman-api -n tokman-production --tail=100

# Get all resources
kubectl get all -n tokman-production

# Describe service
kubectl describe svc tokman-api -n tokman-production

# Check ingress
kubectl get ingress -n tokman-production

# Scale up/down
kubectl scale deployment tokman-api --replicas=10 -n tokman-production
```

## 🚨 Emergency Procedures

### Rollback Failed Deployment
```bash
# Immediately rollback
kubectl rollout undo deployment/tokman-api -n tokman-production

# Wait for rollback
kubectl rollout status deployment/tokman-api -n tokman-production
```

### Restart Pod
```bash
# Delete pod (k8s will restart)
kubectl delete pod <pod-name> -n tokman-production
```

### Check Resource Exhaustion
```bash
# Node resources
kubectl top nodes

# Pod resources
kubectl top pods -n tokman-production

# Increase if needed
kubectl patch deployment tokman-api -p \
  '{"spec":{"template":{"spec":{"containers":[{"name":"tokman","resources":{"limits":{"memory":"2Gi"}}}]}}}}'  \
  -n tokman-production
```

### Database Issues
```bash
# Connect to database
kubectl exec -it <pod> -- psql $DATABASE_URL

# Check migrations
\dt  # list tables

# Rollback migrations (if needed)
go run cmd/server/main.go migrate:rollback
```

## 📞 Getting Help

### Documentation
- 📖 Full guide: [DEPLOY.md](./deployments/DEPLOY.md)
- 📋 Launch plan: [LAUNCH_CHECKLIST.md](./LAUNCH_CHECKLIST.md)
- 💼 Business model: [BUSINESS_MODEL.md](./BUSINESS_MODEL.md)
- 📊 Project status: [PROJECT_STATUS.md](./PROJECT_STATUS.md)

### Support
- 📧 Email: dev@tokman.dev
- 💬 Discord: https://discord.gg/tokman
- 🐛 Issues: https://github.com/GrayCodeAI/tokman/issues
- 📖 Docs: https://tokman.dev/docs

---

**Ready to launch! For detailed information, see the comprehensive guides above.** 🚀
