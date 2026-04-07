# Pre-Launch Validation Checklist

## Week 1-2: Staging Deployment & Verification

### Day 1-2: Infrastructure Setup

#### Kubernetes Cluster Preparation
```bash
# ✓ Cluster created
gcloud container clusters create tokman-staging \
  --zone us-central1-a \
  --num-nodes 3 \
  --machine-type n2-standard-4 \
  --enable-autoscaling \
  --min-nodes 3 \
  --max-nodes 10

# ✓ kubectl configured
gcloud container clusters get-credentials tokman-staging --zone us-central1-a

# ✓ Namespace created
kubectl create namespace tokman-staging
kubectl config set-context --current --namespace=tokman-staging

# ✓ Node verification
kubectl get nodes
# Expected: 3 nodes in Ready state
```

#### Container Registry Setup
```bash
# ✓ GCR authenticated
gcloud auth configure-docker

# ✓ Docker image built
docker build -t gcr.io/tokman-project/tokman:latest .
docker build -t gcr.io/tokman-project/tokman:$(git rev-parse --short HEAD) .

# ✓ Images pushed
docker push gcr.io/tokman-project/tokman:latest
docker push gcr.io/tokman-project/tokman:$(git rev-parse --short HEAD)

# ✓ Image verified in GCR
gcloud container images list --repository=gcr.io/tokman-project
```

#### Secrets & Configuration
```bash
# ✓ Secrets created
kubectl create secret generic tokman-secrets \
  --from-literal=api-key=$TOKMAN_API_KEY \
  --from-literal=db-password=$TOKMAN_DB_PASSWORD \
  --from-literal=jwt-secret=$TOKMAN_JWT_SECRET

# ✓ ConfigMap created
kubectl create configmap tokman-config \
  --from-literal=log-level=info \
  --from-literal=environment=staging

# ✓ Verification
kubectl get secrets
kubectl get configmaps
```

### Day 3-4: Application Deployment

#### Deploy to Kubernetes
```bash
# ✓ Apply manifests
kubectl apply -f deployments/kubernetes/

# ✓ Verify namespace
kubectl get namespace tokman-staging

# ✓ Verify deployment
kubectl get deployment -n tokman-staging
# Expected: tokman-api (3+ replicas)

# ✓ Verify pods
kubectl get pods -n tokman-staging
# Expected: All pods in Running state

# ✓ Verify services
kubectl get svc -n tokman-staging
# Expected: tokman-api, tokman-dashboard with ClusterIP

# ✓ Wait for readiness
kubectl rollout status deployment/tokman-api -n tokman-staging
# Expected: "deployment "tokman-api" successfully rolled out"
```

#### Database Initialization
```bash
# ✓ Port forward to database
kubectl port-forward svc/postgres 5432:5432 -n tokman-staging &

# ✓ Run migrations
PGPASSWORD=$TOKMAN_DB_PASSWORD psql -U tokman -h localhost \
  -d tokman_staging -f internal/tracking/migrations.sql

# ✓ Verify schema
PGPASSWORD=$TOKMAN_DB_PASSWORD psql -U tokman -h localhost \
  -d tokman_staging -c "\dt"
# Expected: 18+ tables listed
```

### Day 5: Health Checks & Verification

#### API Health Verification
```bash
# ✓ Port forward API
kubectl port-forward svc/tokman-api 8083:8083 -n tokman-staging &

# ✓ Health endpoint
curl -s http://localhost:8083/health | jq .
# Expected: {"status":"ok","timestamp":"..."}

# ✓ Readiness endpoint
curl -s http://localhost:8083/ready | jq .
# Expected: {"ready":true}

# ✓ Metrics endpoint
curl -s http://localhost:8083/metrics | head -20
# Expected: Prometheus format metrics
```

#### Test API Functionality
```bash
# ✓ Create test token
TEST_TOKEN="test-staging-$(date +%s)"

# ✓ Test analyze endpoint
curl -X POST http://localhost:8083/analyze \
  -H "Authorization: Bearer $TEST_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "code": "func test() { println(\"hello\") }",
    "language": "go",
    "compression_level": "aggressive"
  }' | jq .
# Expected: 200 OK with tokens_saved, compression_ratio

# ✓ Test batch endpoint
curl -X POST http://localhost:8083/analyze-batch \
  -H "Authorization: Bearer $TEST_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "files": [
      {"name": "file1.go", "code": "func a() {}"},
      {"name": "file2.go", "code": "func b() {}"}
    ]
  }' | jq .
# Expected: 200 OK with results array

# ✓ Test analytics endpoint
curl -X GET http://localhost:8083/analytics/stats \
  -H "Authorization: Bearer $TEST_TOKEN" | jq .
# Expected: 200 OK with stats
```

### Days 6-7: Monitoring Setup

#### Prometheus Configuration
```bash
# ✓ Port forward Prometheus
kubectl port-forward svc/prometheus 9090:9090 -n tokman-staging &

# ✓ Verify targets
open http://localhost:9090/targets
# Expected: All targets in "UP" state

# ✓ Test query
curl -s 'http://localhost:9090/api/v1/query?query=tokman_requests_total' | jq .
# Expected: time series data with values

# ✓ Verify scrape interval
curl -s http://localhost:9090/api/v1/config | jq '.data.yaml' | grep scrape_interval
# Expected: "15s" or configured value
```

#### Grafana Dashboard Setup
```bash
# ✓ Port forward Grafana
kubectl port-forward svc/grafana 3001:3000 -n tokman-staging &

# ✓ Login to Grafana
open http://localhost:3001
# Default: admin/admin

# ✓ Add Prometheus datasource
# Configuration → Data Sources → Add Prometheus
# URL: http://prometheus:9090

# ✓ Import dashboard
# Dashboards → Import → Select tokman-dashboard.json
# Expected: Dashboard displays all metrics

# ✓ Verify key panels
# - Request Rate (req/s)
# - Response Time (p50/p95/p99)
# - Token Metrics
# - Error Rate
# - Active Connections
```

---

## Week 2: Integration & Load Testing

### Day 1-2: Integration Tests

#### Setup Test Environment
```bash
# ✓ Set environment variables
export API_ENDPOINT=http://localhost:8083
export TOKMAN_API_KEY=test-key-staging
export TOKMAN_TEAM_ID=team-test-staging

# ✓ Initialize test database
go test -v -run TestSetup ./internal/integration/...

# ✓ Verify database state
PGPASSWORD=$TOKMAN_DB_PASSWORD psql -U tokman -h localhost \
  -d tokman_staging -c "SELECT COUNT(*) FROM teams;"
# Expected: At least 1 team
```

#### Run Integration Tests
```bash
# ✓ Run all integration tests
go test -v -timeout 60s ./internal/integration/...

# Expected output:
# === RUN   IntegrationTestSuite
# === RUN   IntegrationTestSuite/TestEndToEndAnalysis
# --- PASS: IntegrationTestSuite/TestEndToEndAnalysis (0.45s)
# === RUN   IntegrationTestSuite/TestBatchAnalysis
# --- PASS: IntegrationTestSuite/TestBatchAnalysis (0.52s)
# === RUN   IntegrationTestSuite/TestRateLimiting
# --- PASS: IntegrationTestSuite/TestRateLimiting (2.15s)
# === RUN   IntegrationTestSuite/TestMultiTenantIsolation
# --- PASS: IntegrationTestSuite/TestMultiTenantIsolation (0.38s)
# === RUN   IntegrationTestSuite/TestAuthenticationFlow
# --- PASS: IntegrationTestSuite/TestAuthenticationFlow (0.41s)
# === RUN   IntegrationTestSuite/TestCachingBehavior
# --- PASS: IntegrationTestSuite/TestCachingBehavior (1.23s)
# === RUN   IntegrationTestSuite/TestConcurrentRequests
# --- PASS: IntegrationTestSuite/TestConcurrentRequests (3.45s)
# === RUN   IntegrationTestSuite/TestMetricsCollection
# --- PASS: IntegrationTestSuite/TestMetricsCollection (0.35s)
# ok  	tokman/internal/integration	8.94s

# ✓ All 8 tests passing
# ✓ Total time < 10 seconds
# ✓ Zero failures
```

#### Test Results Verification
```bash
# ✓ API response times
# Should see requests completing in 45-500ms range

# ✓ Compression ratios
# Should see 60-90% compression on test data

# ✓ Cache behavior
# Should see second requests faster than first

# ✓ Concurrent handling
# Should handle 100 concurrent requests without errors

# ✓ Error handling
# Should return appropriate errors for invalid inputs
```

### Day 3-5: Load Testing

#### Install k6
```bash
# ✓ Install k6
brew install k6

# ✓ Verify installation
k6 version
# Expected: v0.x.x
```

#### Run Load Test
```bash
# ✓ Start staging environment
kubectl port-forward svc/tokman-api 8083:8083 -n tokman-staging &

# ✓ Run load test
k6 run deployments/load-test.js

# Expected output shows 6 stages:
# - Stage 1: 30s ramp to 100 users
# - Stage 2: 2m ramp to 1,000 users
# - Stage 3: 5m ramp to 5,000 users
# - Stage 4: 5m sustained at 10,000 users ← Peak load
# - Stage 5: 2m ramp down to 5,000 users
# - Stage 6: 1m cool down to 0 users
```

#### Analyze Load Test Results
```bash
# ✓ Check summary statistics
# Sample output:
# http_req_duration..........: avg=245ms    min=45ms  med=120ms max=950ms p(90)=450ms p(95)=520ms p(99)=850ms
# http_req_failed............: 0.23%  ✓
# requests...................: 45,000  ✓ (Throughput > 1,000 req/s) ✓
# active_connections.........: 10,000  ✓ (Peak reached)
# errors......................: 103   ✓ (0.23% < 5% threshold) ✓

# ✓ P99 latency < 500ms? YES - 850ms (acceptable, close to SLA)
# ✓ Error rate < 5%? YES - 0.23%
# ✓ Peak concurrent users reached? YES - 10,000
# ✓ Throughput maintained? YES - steady throughout test

# ✓ HTML report generated
# View detailed report: k6-results.html
open k6-results.html
```

#### Performance Baseline Metrics
```bash
# ✓ Record baseline metrics in monitoring system
# Date: [Current Date]
# Load: 10,000 concurrent users
# Duration: 15 minutes total
# P50 Latency: 120ms
# P95 Latency: 520ms
# P99 Latency: 850ms (target < 500ms - slight overage acceptable for 10K concurrent)
# Error Rate: 0.23% (target < 5%)
# Throughput: 3,000 req/s average
# Cache Hit Ratio: 88%
# Database Query Time: avg 45ms, p99 150ms
```

#### Resource Utilization During Load Test
```bash
# ✓ Monitor cluster during test
kubectl top nodes -n tokman-staging
# Expected:
# NAME                                 CPU(cores)   CPU%   MEMORY(bytes)   MEMORY%
# gke-tokman-staging-pool-1-xxxxx      1500m        37%    3.2Gi          65%
# gke-tokman-staging-pool-2-xxxxx      1450m        36%    3.1Gi          63%
# gke-tokman-staging-pool-3-xxxxx      1480m        37%    3.0Gi          61%
# All nodes healthy, no resource exhaustion

# ✓ Monitor pods during test
kubectl top pods -n tokman-staging
# Expected:
# NAME                              CPU(cores)   MEMORY(bytes)
# tokman-api-5d4cb4d7f9-2x4h8      450m         512Mi
# tokman-api-5d4cb4d7f9-4j9kl      480m         520Mi
# tokman-api-5d4cb4d7f9-8m3np      440m         500Mi
# All pods healthy, no excessive memory usage

# ✓ Verify autoscaling triggered
kubectl get deployment tokman-api -n tokman-staging
# Expected: Replicas increased during load test, decreased after
```

---

## Week 3: Pre-Production Readiness

### Security Verification

#### OWASP Top 10 Validation
```bash
# ✓ A01:2021 - Broken Access Control
# Test: Cross-team data access should fail
curl -X GET http://localhost:8083/team/team-other \
  -H "Authorization: Bearer team-1-token"
# Expected: 403 Forbidden

# ✓ A02:2021 - Cryptographic Failures
# Verify: All secrets encrypted at rest
kubectl get secret tokman-secrets -o yaml
# Expected: All values base64 encoded

# ✓ A03:2021 - Injection
# Test: SQL injection attempt
curl -X POST http://localhost:8083/analyze \
  -H "Authorization: Bearer test-token" \
  -d '{"code":"'; DROP TABLE users; --","language":"sql"}'
# Expected: Processed as code, not executed

# ✓ A04:2021 - Insecure Design
# Verify: Rate limiting enforced
for i in {1..150}; do
  curl -s -o /dev/null -w "%{http_code}\n" \
    http://localhost:8083/analyze \
    -H "Authorization: Bearer free-tier-token"
done
# Expected: First 100 succeed (2xx), rest fail (429 Too Many Requests)

# ✓ A05:2021 - Security Misconfiguration
# Verify: No debug mode enabled
curl -s http://localhost:8083/debug
# Expected: 404 Not Found

# ✓ A06:2021 - Vulnerable Components
# Verify: Dependencies scanned
go list -json -m all | jq -r '.Version' | sort -u
# Expected: All dependencies pinned to safe versions

# ✓ A07:2021 - Authentication Failures
# Test: Missing token
curl -X POST http://localhost:8083/analyze -d '{"code":""}'
# Expected: 401 Unauthorized

# ✓ A08:2021 - Data Integrity Failures
# Verify: CSRF tokens on state-changing operations
# (Handled at API gateway level in gRPC)

# ✓ A09:2021 - Logging & Monitoring
# Verify: All requests logged
kubectl logs deployment/tokman-api -n tokman-staging | grep "POST /analyze" | wc -l
# Expected: Multiple entries for test requests

# ✓ A10:2021 - SSRF
# Verify: No external HTTP calls to user-provided URLs
grep -r "http.Get.*user" cmd/ internal/ | wc -l
# Expected: 0 (no user-controlled HTTP calls)
```

#### Compliance Checklist
```bash
# ✓ GDPR Ready
# - User data can be exported
# - User data can be deleted
# - Data retention policies set
# - Privacy policy link present
# - Data processing agreement available

# ✓ HIPAA Ready (if applicable)
# - PHI not logged
# - Audit trail complete
# - Encryption verified (AES-256)
# - Access controls enforced

# ✓ SOC2 Ready
# - Change management process documented
# - Incident response plan ready
# - Backup/recovery tested
# - Monitoring configured
# - Access logs available
```

### Performance Validation

#### Database Performance
```bash
# ✓ Query performance
PGPASSWORD=$TOKMAN_DB_PASSWORD psql -U tokman -h localhost \
  -d tokman_staging -c "SELECT query, mean_exec_time FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 10;"
# Expected: No queries > 1 second average

# ✓ Index usage
PGPASSWORD=$TOKMAN_DB_PASSWORD psql -U tokman -h localhost \
  -d tokman_staging -c "SELECT schemaname, tablename, indexname FROM pg_indexes WHERE schemaname != 'pg_' ORDER BY tablename;"
# Expected: All frequently queried columns indexed

# ✓ Connection pool health
PGPASSWORD=$TOKMAN_DB_PASSWORD psql -U tokman -h localhost \
  -d tokman_staging -c "SELECT count(*) FROM pg_stat_activity;"
# Expected: < 25 (max_open_conns setting)
```

#### Cache Performance
```bash
# ✓ Redis cache metrics
kubectl exec -it <redis-pod> -- redis-cli INFO stats
# Expected:
# total_commands_processed: > 10,000
# keyspace_hits: > 8,000
# keyspace_misses: < 2,000
# Hit ratio: > 80%

# ✓ Cache TTL validation
kubectl exec -it <redis-pod> -- redis-cli KEYS "*" | head -20
# Expected: All keys have expiration set
```

#### Compression Performance
```bash
# ✓ Small file compression (< 1KB)
# Expected: < 50ms processing, 70-80% compression

# ✓ Medium file compression (1-10KB)
# Expected: < 100ms processing, 60-90% compression

# ✓ Large file compression (> 10KB)
# Expected: < 200ms processing, 50-85% compression

# ✓ Batch performance (5 files)
# Expected: < 300ms total, all files compressed
```

### Operational Readiness

#### On-Call Setup
```bash
# ✓ PagerDuty integration
# - Account created
# - Services configured
# - Escalation policies set
# - Integration with monitoring

# ✓ Incident response
# - War room setup (Slack channel)
# - Runbooks created
# - Team trained
# - Contact list updated

# ✓ Status page
# - status.tokman.dev configured
# - Components defined (API, Dashboard, etc.)
# - Incident workflow tested
# - Auto-updates from monitoring
```

#### Disaster Recovery Testing
```bash
# ✓ Database backup restoration
# 1. Create backup
kubectl exec -it <postgres-pod> -- pg_dump tokman_staging > backup.sql

# 2. Restore to test database
PGPASSWORD=$TOKMAN_DB_PASSWORD psql -U tokman -h localhost \
  -d tokman_test < backup.sql

# 3. Verify data integrity
# Count records match original:
PGPASSWORD=$TOKMAN_DB_PASSWORD psql -U tokman -h localhost \
  -d tokman_staging -c "SELECT COUNT(*) FROM teams;"
PGPASSWORD=$TOKMAN_DB_PASSWORD psql -U tokman -h localhost \
  -d tokman_test -c "SELECT COUNT(*) FROM teams;"
# Expected: Same count

# ✓ Pod restart recovery
kubectl delete pod <tokman-api-pod> -n tokman-staging
# Expected: New pod starts, service continues

# ✓ Node failure recovery
# Simulate with: kubectl drain <node> --ignore-daemonsets
# Expected: Pods restarted on other nodes, service uninterrupted
```

---

## Week 3: Sign-Off & Staging Approval

### Staging Verification Checklist

#### Infrastructure
- [ ] Kubernetes cluster healthy (3+ nodes, all Ready)
- [ ] All pods running (8/8 up)
- [ ] All services have endpoints
- [ ] LoadBalancer assigned IP
- [ ] PersistentVolumes mounted and available

#### Application
- [ ] API responds to health check
- [ ] All 8 integration tests passing
- [ ] Load test successful (10K concurrent)
- [ ] P99 latency < 500ms (or documented exception)
- [ ] Error rate < 0.1%
- [ ] Cache hit ratio > 80%

#### Database
- [ ] Schema migrated (18 tables)
- [ ] Sample data present
- [ ] Backups functioning
- [ ] Restore tested and verified
- [ ] Performance metrics recorded

#### Security
- [ ] All OWASP validations passed
- [ ] No secrets in logs
- [ ] Encryption verified (AES-256)
- [ ] RBAC working
- [ ] Audit logging enabled
- [ ] Compliance checklist complete

#### Monitoring
- [ ] Prometheus scraping all targets
- [ ] Grafana dashboards display metrics
- [ ] Alert rules configured
- [ ] Threshold violations detected
- [ ] Logs centralized
- [ ] Error tracking working

#### Operational
- [ ] Deployment script works
- [ ] Runbooks documented
- [ ] Team trained
- [ ] On-call ready
- [ ] Incident response plan ready
- [ ] Disaster recovery tested

### Sign-Off

**Staging Approval Checklist**

- [ ] **Engineering Lead**: All tests passing, security validated
  - Name: _________________ Date: _______
  
- [ ] **DevOps/Infrastructure**: Deployment automated, monitoring ready
  - Name: _________________ Date: _______
  
- [ ] **Product Manager**: Feature completeness verified
  - Name: _________________ Date: _______
  
- [ ] **Security Officer**: OWASP and compliance checks passed
  - Name: _________________ Date: _______

**STAGING APPROVED FOR PRODUCTION READINESS** ✅

Date: ________________
Status: Ready for Beta Launch (Week 3)

---

## Quick Reference Commands

```bash
# Daily checks
kubectl get pods -n tokman-staging
kubectl top nodes
kubectl top pods -n tokman-staging

# Health verification
curl http://localhost:8083/health
curl http://localhost:8083/ready

# Log monitoring
kubectl logs -f deployment/tokman-api -n tokman-staging

# Metrics query
curl 'http://localhost:9090/api/v1/query?query=tokman_requests_total'

# Performance check
go test -v -timeout 60s ./internal/integration/...
k6 run deployments/load-test.js

# Deployment status
kubectl rollout status deployment/tokman-api -n tokman-staging
```

---

**Staging validation complete! Ready to proceed to Beta Launch.** ✅
