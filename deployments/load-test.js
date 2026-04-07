import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Rate, Trend, Counter, Gauge } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const requestDuration = new Trend('request_duration');
const throughput = new Counter('requests');
const activeConnections = new Gauge('active_connections');

export const options = {
  stages: [
    { duration: '30s', target: 100 },    // Warm up
    { duration: '2m', target: 1000 },    // Ramp to 1K
    { duration: '5m', target: 5000 },    // Ramp to 5K
    { duration: '5m', target: 10000 },   // Peak load: 10K
    { duration: '2m', target: 5000 },    // Wind down to 5K
    { duration: '1m', target: 0 },       // Cool down
  ],
  thresholds: {
    'http_req_duration': ['p(99)<500', 'p(95)<300', 'p(50)<100'],
    'errors': ['rate<0.05'],  // < 5% error rate
    'request_duration': ['avg<250'],
  },
  ext: {
    loadimpact: {
      projectID: 3356104,
      name: 'TokMan Load Test'
    }
  }
};

const API_ENDPOINT = __ENV.API_ENDPOINT || 'http://localhost:8083';
const API_KEY = __ENV.API_KEY || 'test-key';

// Test data samples
const codeSamples = [
  // Small function
  `function fibonacci(n) {
    if (n <= 1) return n;
    return fibonacci(n-1) + fibonacci(n-2);
  }`,

  // Medium code with imports
  `import { useState, useEffect } from 'react';
  import axios from 'axios';

  export function UserDashboard() {
    const [users, setUsers] = useState([]);
    const [loading, setLoading] = useState(true);

    useEffect(() => {
      fetchUsers();
    }, []);

    const fetchUsers = async () => {
      const response = await axios.get('/api/users');
      setUsers(response.data);
      setLoading(false);
    };

    return (
      <div>
        {loading ? <p>Loading...</p> : <UserList users={users} />}
      </div>
    );
  }`,

  // Large class-based code
  `class DataProcessor {
    private cache: Map<string, any>;
    private logger: Logger;
    private mutex: Mutex;

    constructor() {
      this.cache = new Map();
      this.logger = new Logger('DataProcessor');
      this.mutex = new Mutex();
    }

    async process(data: Buffer): Promise<Buffer> {
      await this.mutex.lock();
      try {
        const cacheKey = this.hash(data);
        if (this.cache.has(cacheKey)) {
          this.logger.info('Cache hit', { key: cacheKey });
          return this.cache.get(cacheKey);
        }

        const result = await this.transform(data);
        this.cache.set(cacheKey, result);
        this.logger.info('Processed', { size: result.length });
        return result;
      } finally {
        await this.mutex.unlock();
      }
    }

    private async transform(data: Buffer): Promise<Buffer> {
      return Buffer.from(data.toString().toUpperCase());
    }

    private hash(data: Buffer): string {
      return data.toString('hex').substring(0, 16);
    }
  }`,

  // ML training logs
  `Epoch 1/100
1000/1000 [==============================] - 42s 42ms/step - loss: 0.5234 - accuracy: 0.7823 - val_loss: 0.4521 - val_accuracy: 0.7945
Epoch 2/100
1000/1000 [==============================] - 38s 38ms/step - loss: 0.4156 - accuracy: 0.8234 - val_loss: 0.3892 - val_accuracy: 0.8421
Epoch 3/100
1000/1000 [==============================] - 39s 39ms/step - loss: 0.3421 - accuracy: 0.8567 - val_loss: 0.3245 - val_accuracy: 0.8756`,

  // DevOps output
  `$ kubectl get pods -n production
NAME                               READY   STATUS    RESTARTS   AGE
tokman-api-5d4cb4d7f9-2x4h8       1/1     Running   0          2d
tokman-api-5d4cb4d7f9-4j9kl       1/1     Running   0          2d
tokman-api-5d4cb4d7f9-8m3np       1/1     Running   0          2d
tokman-dashboard-7b8d9c3f-9q2ws   1/1     Running   0          1d
tokman-workers-6c5d4e2a-5p7rt     1/1     Running   0          3h`,
];

export default function () {
  activeConnections.add(1);

  // Select random code sample
  const code = codeSamples[Math.floor(Math.random() * codeSamples.length)];
  const params = {
    headers: {
      'Authorization': `Bearer ${API_KEY}`,
      'Content-Type': 'application/json',
    },
  };

  group('Analyze Endpoint', () => {
    const payload = JSON.stringify({
      code: code,
      language: 'javascript',
      compression_level: 'aggressive'
    });

    const startTime = new Date();
    const response = http.post(`${API_ENDPOINT}/analyze`, payload, params);
    const duration = new Date() - startTime;

    requestDuration.add(duration);
    throughput.add(1);

    check(response, {
      'status is 200': (r) => r.status === 200,
      'has tokens_saved': (r) => r.json('tokens_saved') > 0,
      'has compression_ratio': (r) => r.json('compression_ratio') > 0,
      'compression_ratio is valid': (r) => {
        const ratio = r.json('compression_ratio');
        return ratio >= 0 && ratio <= 1;
      },
      'response time < 500ms': () => duration < 500,
    }) || errorRate.add(1);
  });

  group('Batch Analyze Endpoint', () => {
    const payload = JSON.stringify({
      files: [
        { name: 'file1.js', code: codeSamples[0] },
        { name: 'file2.js', code: codeSamples[1] },
        { name: 'file3.js', code: codeSamples[2] },
      ],
      compression_level: 'moderate'
    });

    const response = http.post(`${API_ENDPOINT}/analyze-batch`, payload, params);

    check(response, {
      'status is 200': (r) => r.status === 200,
      'has results array': (r) => r.json('results').length === 3,
      'all results have compression_ratio': (r) => {
        const results = r.json('results');
        return results.every(item => item.compression_ratio !== undefined);
      },
    }) || errorRate.add(1);
  });

  // Rate limiting test (some requests should be rate limited on free tier)
  group('Rate Limiting Behavior', () => {
    const response = http.get(`${API_ENDPOINT}/health`, params);

    check(response, {
      'health check succeeds': (r) => r.status === 200,
      'has rate limit headers': (r) => r.headers['X-RateLimit-Limit'] !== undefined,
    });
  });

  activeConnections.add(-1);
  sleep(Math.random() * 2); // Random think time between 0-2 seconds
}

export function teardown() {
  console.log(`Load test completed:
    Total requests: ${throughput.value}
    Error rate: ${(errorRate.value * 100).toFixed(2)}%
    Avg duration: ${requestDuration.value.toFixed(0)}ms`);
}
