import http from 'k6/http';
import {check, group} from 'k6';

import {deepEqual} from '../common/utils.js';

export function getSystemHealth() {
  group('system', function () {
    const res = http.get(`${globalThis.BASE_URL}/v1/system/health`);

    check(res, {
      'status is 200': r => r.status === 200,
      'response body': r => deepEqual(JSON.parse(r.body), {
          'cache_database': 'healthy',
          'graph_database': 'healthy',
          'license': 'healthy',
          'message_queue': 'healthy',
          'relational_database': 'healthy',
      }),
    });
  });
}
