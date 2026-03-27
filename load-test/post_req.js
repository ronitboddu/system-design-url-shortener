import http from 'k6/http';
import { sleep, check } from 'k6';

export const options = {
  vus: 1,
  duration: '5s',
  thresholds: {
    http_req_duration: ['p(95)<500'],
  },
};

function generateIP(vu, iter) {
  const lastOctet = (iter % 254) + 1;
  const thirdOctet = (vu % 254) + 1;
  return `10.0.${thirdOctet}.${lastOctet}`;
}

export default function () {
  const url = 'http://localhost:8080/shorten';
  const fakeIP = generateIP(__VU, __ITER);

  const payload = JSON.stringify({
    expTime: 2,
    urlPath: 'https://ww26.0123movie.net/movie/iron-man-3-1712.html',
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'X-Forwarded-For': fakeIP,
      'X-Real-IP': fakeIP,
    },
  };

  const res = http.post(url, payload, params);

  check(res, {
    'status is 200 or 201': (r) => r.status === 200 || r.status === 201,
    'response time < 500ms': (r) => r.timings.duration < 500,
  });

  console.log(
    `VU=${__VU} ITER=${__ITER} IP=${fakeIP} STATUS=${res.status} DURATION_MS=${res.timings.duration}`
  );
  sleep(1);
}
