import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
  vus: 1,
  iterations: 10,
};

export default function () {
  const idx = __ITER + 1;
  const url = 'http://localhost:8080/shorten';

  const payload = JSON.stringify({
    expTime: 2,
    urlPath: `https://example.com/seed/${idx}`,
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
      'X-Forwarded-For': `10.10.0.${idx}`,
      'X-Real-IP': `10.10.0.${idx}`,
    },
  };

  const res = http.post(url, payload, params);

  check(res, {
    'seed status ok': (r) => r.status === 200 || r.status === 201,
  });

  let shortURL = '';
  try {
    shortURL = res.json('short_url');
  } catch (e) {
    shortURL = 'unable-to-parse-short-url';
  }

  console.log(`SEED idx=${idx} status=${res.status} short_url=${shortURL}`);
  sleep(0.2);
}
