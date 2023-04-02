The following folders and files are included in the starter templates:

### common

The `common` folder serves as a suitable location for storing re-usable functions.

An example of such a function can be found in `utils.js`, namely `randomIntBetween` (note the use of the `export`
keyword to allow it to be imported elsewhere). This function can be used to add a random `sleep` "think-time"
in `scripts/getSystemHealth.js`.

### config

It is often useful to separate test configuration from JavaScript code, particularly when the code is being kept under
source control.

The contents of `test.json` will be merged with `options` set directly in the entry-point script (`main.js`). One of the
more useful properties to store in this JSON file are [Scenario](https://k6.io/docs/using-k6/scenarios/) definitions.

The current `test.json` contains a single scenario that will run the exported function in `scenarios/scenario1.js` as
determined by the scenario's `exec` property that expects the name of an exported function to execute.

Note that it is necessary to re-export this function in the entry-point script (`main.js`) in order to make it
accessible.

### scenarios

Scripts in the `scenarios` folder contain exported functions that can be used as entry-points for Scenarios (i.e.
the `exec` property in `test.json`). These functions will in turn call exported functions provided by the `scripts`
themselves. This layer of abstraction allows for some customization of what should be executed; perhaps there is a need
to loop over a set of API functions - this might not be desirable to define in the scripts themselves. The order in
which script functions are called can also be changed more easily if this code is separated out.

### scripts

The `scripts` folder can be used to store code that performs actual automation.

In the example script `getSystemHealth.js`, a GET request is made and a [check](https://k6.io/docs/using-k6/checks/) is
used to validate the response status code. A subsequent `sleep` statement is then used to apply "think-time" based on a
random value between the global `PAUSE_MIN` and `PAUSE_MAX` variables defined in the entry-point script `main.js`.

### Examples

The v2 branch of the [k6-example-woocommerce](https://github.com/grafana/k6-example-woocommerce/tree/v2) repository
implements the `vanilla` template.
