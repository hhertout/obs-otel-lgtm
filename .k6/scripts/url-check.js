import { check } from 'k6';
import http from 'k6/http';

export function checkUrl() {
    const res = http.get('http://test.k6.io/');
    check(res, {
        'is status 200': (r) => r.status === 200,
    });
}