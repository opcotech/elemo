# Changelog

All notable changes to this project will be documented in this file.

## [unreleased]

**Bug Fixes**

- Ensure tracer is initialized for auth cli command ([683354a](https://github.com/opcotech/elemo/commit/683354a3f2c832b665304c64e840e51b0d1f0dc0))
- Ensure the CI variable is not unbound ([bf444b6](https://github.com/opcotech/elemo/commit/bf444b6b109a2860683fd6df43912f67ab1a58d8))

**Documentation**

- Ensure best practices (#142) ([ff38476](https://github.com/opcotech/elemo/commit/ff38476a65e38ee858292752dcb3fc1365fd2259))
- Replace badges and extend contributing guide ([dec6bb2](https://github.com/opcotech/elemo/commit/dec6bb2b7489390c384ead8544dd0702af98d7fe))
- Improve tone of CODE_OF_CONDUCT.md (#248) ([c60b125](https://github.com/opcotech/elemo/commit/c60b1252a6e9bdc7dfc68129c1a1751f1f77c974))

**Features**

- Initial commit ([35cb9fc](https://github.com/opcotech/elemo/commit/35cb9fccd081c036c901afc4c5affc91aac20bbc))
- Add todo service ([3b4194d](https://github.com/opcotech/elemo/commit/3b4194d392445c15680e680ec95666b2b5d1e42a))
- Add repository response caching (#9) ([af37fbd](https://github.com/opcotech/elemo/commit/af37fbd4a0f20831105daa727f078e0fd79cbcd5))
- Add asynq workers (#35) ([200974b](https://github.com/opcotech/elemo/commit/200974b7cce7c7174a03aeae3080cd1ee7ce243e))
- Implement todo list (#39) ([6faecc1](https://github.com/opcotech/elemo/commit/6faecc10af9c992dd3a8f48c4d3ff2dfdfd15579))
- Add due_date for todo items ([ba2a9da](https://github.com/opcotech/elemo/commit/ba2a9da498aab0dbd067f203a1b43a99879b48ab))
- Sending emails from templates (#40) ([54a3f1b](https://github.com/opcotech/elemo/commit/54a3f1b13101456a3d105e8814c3acb2b96b6a75))
- Add email service and task scheduler (#43) ([470c76a](https://github.com/opcotech/elemo/commit/470c76ab76217bdf7227c18537844179be5d84a4))
- Add organization service (#45) ([88f9857](https://github.com/opcotech/elemo/commit/88f98571adf5d38155f0df722d8de8bdf5ec4d39))
- Add permission service (#62) ([76adc7e](https://github.com/opcotech/elemo/commit/76adc7e2127fe19036870d5c5c5af9f19e1a8714))
- Add role service (#127) ([346bbd4](https://github.com/opcotech/elemo/commit/346bbd408cfe8b921749487dd26ec9796e554ef3))
- In-app notifications (#178) ([f5eb706](https://github.com/opcotech/elemo/commit/f5eb7065e0fdb6c8d216ee9ec78363a3a8ecf567))

**Miscellaneous Tasks**

- Bump front-end dependencies (#34) ([43ac423](https://github.com/opcotech/elemo/commit/43ac4233df123b45ef235d213b7091d804e0b22b))
- Drop unnecessary license decisions (#41) ([37cfce2](https://github.com/opcotech/elemo/commit/37cfce21dddb08f8ba7a0ef89f1b0d3e8e61976a))
- Update dependencies and set update frequency (#42) ([c3014d0](https://github.com/opcotech/elemo/commit/c3014d0d0637098ae6d915d2de0474a2028146d8))
- Ensure commit lint ignores dependencies ([d976979](https://github.com/opcotech/elemo/commit/d976979376df523509268f4747ea6bf09df78251))
- Suppress ld warning ([7ae6258](https://github.com/opcotech/elemo/commit/7ae625888ba6c857c133e153fec5d7b2a411755f))
- Ensure commit lint ignores dependencies ([77ef2fd](https://github.com/opcotech/elemo/commit/77ef2fd4f3a9892c7a44b2ccfca06c0ef27ba132))
- Switch license to AGPL-3.0-or-later (#165) ([2f20bec](https://github.com/opcotech/elemo/commit/2f20bec0e0748a48a0d1c623f44579dc2574b1eb))
- Update dependencies ([3d7b83b](https://github.com/opcotech/elemo/commit/3d7b83bd453431d0bb537881cdab5a8d0550741a))
- Parse quotaValue with ParseUint instead of Atoi (#177) ([b73e816](https://github.com/opcotech/elemo/commit/b73e816b403a406346dc441dc22d1b4c7b791a64))
- Bump @storybook/addon-viewport in /web (#184) ([051bc7e](https://github.com/opcotech/elemo/commit/051bc7eb306a7854ba62dd05d8de95c6acf970e6))
- Bump backend and frontend dependencies (#205) ([d05d616](https://github.com/opcotech/elemo/commit/d05d6168f7b152054d66e82226eb051fd94f117c))

**Refactor**

- Enhance todo list (#44) ([d4798eb](https://github.com/opcotech/elemo/commit/d4798ebd6a03e11070f0f249eac37302732f7cee))
- Revamp UI and related API (#60) ([abc4fe3](https://github.com/opcotech/elemo/commit/abc4fe3197ebde4e7c9504feb3a0875c776239e0))
- Prepare for open sourcing (#128) ([9be5a86](https://github.com/opcotech/elemo/commit/9be5a86de9cfddde1bbfa9c8d983b33370e9e857))
- Update front-end dependencies (#143) ([8996c02](https://github.com/opcotech/elemo/commit/8996c02414a14d22b5b1c15f11ab5fd684ec7b65))
- Separate async tasks from handlers (#153) ([9a3b5f0](https://github.com/opcotech/elemo/commit/9a3b5f0f28a13021b4d9770e6b94608ba0f6db78))
- Simplify getting started for newcomers (#160) ([f9b716b](https://github.com/opcotech/elemo/commit/f9b716beed9370592e9d96ae54ccf20debdd57b2))
- Use the built-in datetime function for updating resources (#181) ([5dd8d6e](https://github.com/opcotech/elemo/commit/5dd8d6eb37382aa62892f86c3c8d8951426b8490))
- Improve contributing experience (#182) ([db5a6cf](https://github.com/opcotech/elemo/commit/db5a6cfd3537f3bb25fd20287d9b785af94d0023))
- Rewrite the cartesian products (#249) ([324b2f1](https://github.com/opcotech/elemo/commit/324b2f12f4cad778685e4071f728f67c5df6b8a9))
- Migrate the frontend to modern foundations (#244) ([0831d61](https://github.com/opcotech/elemo/commit/0831d61ba3e5f3eb03857490394f0db1d18d2225))

**Testing**

- Refactor test env file usage ([2b8d9f4](https://github.com/opcotech/elemo/commit/2b8d9f4b03be2de21c99d7b2452ec80293e2af59))
- Fix linter configuration ([618328b](https://github.com/opcotech/elemo/commit/618328bfb8f5f7af3b33b2d5b18d2bba99b9b243))

**Build**

- Fix config generation and use ([fbe6659](https://github.com/opcotech/elemo/commit/fbe6659f5ae7783cb1b394df2495beeb3c8a7b86))
- Resolve codeql errors ([1fd49b3](https://github.com/opcotech/elemo/commit/1fd49b3285d9008cebee19c88de3b3467465040e))
- Bump go and refactor build pipelines (#250) ([d879ea9](https://github.com/opcotech/elemo/commit/d879ea923de4a4f8d57a82280463f7663648ab80))

**Ci**

- Bump paambaati/codeclimate-action from 3.2.0 to 4.0.0 (#6) ([66ef67b](https://github.com/opcotech/elemo/commit/66ef67bacd57fdee8aabd0722a81f7ca1afc1e67))
- Bump golangci/golangci-lint-action from 3 to 4 (#163) ([ac97102](https://github.com/opcotech/elemo/commit/ac97102a15dde526d51d4b12f0a01f56385ede8d))
- Bump actions/setup-go from 4 to 5 (#162) ([b162b70](https://github.com/opcotech/elemo/commit/b162b70f67c4c66923f9ae7e88f1c57c7319dfb6))
- Bump actions/checkout from 3 to 4 (#161) ([380108d](https://github.com/opcotech/elemo/commit/380108d8540346ca436e27fb438ee5a9622fa44b))
- Bump golangci/golangci-lint-action from 5 to 6 (#183) ([b435210](https://github.com/opcotech/elemo/commit/b43521087f11d20617eea78763e1d9f7c7d0bb74))
- Ensure setup finishes in non-ci environment ([b595bbb](https://github.com/opcotech/elemo/commit/b595bbbfc01e72a63d8b9631760a5de28c97f067))
- Disable dependabot ([19860a2](https://github.com/opcotech/elemo/commit/19860a2ee7a697647defd2e4996471120157ca44))
- Fix upload asset path visibility ([a4db42b](https://github.com/opcotech/elemo/commit/a4db42b37291bb64285710f12294503a55d5b7a3))
- Replace code coverage reporter ([82e8534](https://github.com/opcotech/elemo/commit/82e8534a3e02975c7494a8c1486756c0a101ad68))


