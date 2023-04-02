export {monitoring} from './scenarios/monitoring.js';

globalThis.PAUSE_MIN = __ENV.PAUSE_MIN || 0;
globalThis.PAUSE_MAX = __ENV.PAUSE_MAX || 5;
globalThis.BASE_URL = __ENV.BASE_URL || 'https://0.0.0.0:35478';

const testConfig = JSON.parse(open('./config/test.json'));

export const options = Object.assign({
  insecureSkipTlsVerify: true,
}, testConfig);

export default function () {
  console.log('No scenarios in test.json. Executing default function...');
}
