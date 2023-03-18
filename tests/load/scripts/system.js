import http from 'k6/http';
import {check, group} from 'k6';

import {deepEqual} from '../common/utils.js';

export function getSystemHealth() {
  group('System', function () {
    const res = http.get(`${globalThis.BASE_URL}/v1/system/health`);

    check(res, {
      'status is 200': r => r.status === 200,
      'protocol is HTTP/2': r => r.proto === 'HTTP/2.0',
      'response body': r => deepEqual(JSON.parse(r.body), {
        "graph_database": "healthy",
        "relational_database": "healthy"
      }),
    });
  });
}
