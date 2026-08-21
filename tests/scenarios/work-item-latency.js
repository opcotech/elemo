import {runIssueListBenchmark, setupIssueListBenchmark} from '../scripts/work-item-list.js';

let benchmarkData;

export function workItemLatency() {
  if (!benchmarkData) {
    benchmarkData = setupIssueListBenchmark();
  }
  runIssueListBenchmark(benchmarkData);
}
