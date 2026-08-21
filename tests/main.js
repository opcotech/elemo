export {monitoring} from './scenarios/monitoring.js';
export {workItemLatency} from './scenarios/work-item-latency.js';

globalThis.PAUSE_MIN = __ENV.PAUSE_MIN || 0.5;
globalThis.PAUSE_MAX = __ENV.PAUSE_MAX || 1.5;
globalThis.BASE_URL = __ENV.BASE_URL || 'http://127.0.0.1:35478';

const configPath = __ENV.K6_CONFIG || './config/test.json';
const testConfig = JSON.parse(open(configPath));

export const options = Object.assign({
  insecureSkipTlsVerify: true,
}, testConfig);

export default function () {
  console.log('No scenarios in test.json. Executing default function...');
}
