import http from 'k6/http';
import { sleep, check } from 'k6';

export const options = {
  vus: 1,
  duration: '1s',
  thresholds: {
    http_req_duration: ['p(95)<500'],
  },
};

export default function () {
    const url = 'http://localhost:8080/8w7Amqj'

    const res = http.get(url)

    check(res, {
        'status is 301': (r) => r.status === 301,
        'response time < 500ms': (r) => r.timings.duration < 500,
    });

    sleep(1);
}