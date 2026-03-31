import http from 'k6/http';
import { check, sleep } from 'k6';

const existingShortCodes = [
  'lXY6YIaRrO',
  'lXY6ZTGZcA',
  'lXY713vsYM',
  'lXY724ePN6',
  'lXY73cUYfe',
  'lXY74bnuOQ',
  'lXY75a7FII',
  'lXY768RQCA',
  'lXY778baAo',
  'lXY787uxKo',
];

export const options = {
  scenarios: {
    reads: {
      executor: 'constant-arrival-rate',
      rate: 400,
      timeUnit: '1s',
      duration: '30s',
      preAllocatedVUs: 5000,
      exec: 'readFlow',
    },
    writes: {
      executor: 'constant-arrival-rate',
      rate: 50,
      timeUnit: '1s',
      duration: '30s',
      preAllocatedVUs: 1500,
      exec: 'writeFlow',
    },
  },
  summaryTrendStats: ['min', 'avg', 'med', 'p(90)', 'p(95)', 'p(99)', 'max'],
  thresholds: {
    http_req_duration: ['p(95)<500', 'p(99)<1000'],
    http_req_failed: ['rate<0.01'],
  },
};

export function readFlow() {
  const code = existingShortCodes[__ITER % existingShortCodes.length];
  const res = http.get(`http://localhost:8080/${code}`, {
    redirects: 0,
  });

  check(res, {
    'read status is 302': (r) => r.status === 302,
    'read latency < 500ms': (r) => r.timings.duration < 500,
  });

  sleep(0.1);
}

export function writeFlow() {
  const url = 'http://localhost:8080/shorten';

  const payload = JSON.stringify({
    expTime: 2,
    urlPath: `https://example.com/write/${__VU}-${__ITER}-${Date.now()}`,
  });

  const fakeIP = `10.${(__VU % 200) + 1}.${(__ITER % 200) + 1}.${((__ITER + __VU) % 254) + 1}`;

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'X-Forwarded-For': fakeIP,
      'X-Real-IP': fakeIP,
    },
  };

  const res = http.post(url, payload, params);

  check(res, {
    'write status ok': (r) => r.status === 200 || r.status === 201,
    'write latency < 500ms': (r) => r.timings.duration < 500,
  });

  sleep(0.1);
}
