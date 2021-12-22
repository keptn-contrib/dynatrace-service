# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

## [0.19.0](https://github.com/keptn-contrib/dynatrace-service/compare/0.18.2...0.19.0) (2021-12-22)

### Release validated with
 | Dynatrace-service: `0.19.0` | Keptn: `0.11.3` | Dynatrace: `1.232` |
 |---|---|---|


### ⚠ BREAKING CHANGES

* Require `dynatrace.conf.yaml` and remove default configuration (#612)

### Features

* add endpoints for readiness and liveness probes ([#635](https://github.com/keptn-contrib/dynatrace-service/issues/635)) ([f943505](https://github.com/keptn-contrib/dynatrace-service/commit/f943505542856451bbe3c995130d486baae0d212))
* Require `dynatrace.conf.yaml` and remove default configuration ([#612](https://github.com/keptn-contrib/dynatrace-service/issues/612)) ([95e8776](https://github.com/keptn-contrib/dynatrace-service/commit/95e87767b9bc1a1f41c9f85c3582225ed5da131b))


### Bug Fixes

* Custom Charting and Data Explorer dashboard tiles that return no data should produce a failed indicator value and be included in SLO objectives ([#610](https://github.com/keptn-contrib/dynatrace-service/issues/610)) ([8df2f95](https://github.com/keptn-contrib/dynatrace-service/commit/8df2f950b1bf218fbdc48b3f141646b38c6dd57a))
* Ensure unsupported dashboard tiles add objectives to SLO files ([#604](https://github.com/keptn-contrib/dynatrace-service/issues/604)) ([86340ff](https://github.com/keptn-contrib/dynatrace-service/commit/86340ff638c172961c6a31ad7b911f0b57d16a87))
* No get-sli.finished event is sent if Dynatrace credentials cannot be found ([#611](https://github.com/keptn-contrib/dynatrace-service/issues/611)) ([3ea2b0c](https://github.com/keptn-contrib/dynatrace-service/commit/3ea2b0c7c0220a28d49d230b660255cf5c76abd6))
* Simplify ProblemsV2Client and SecurityProblemsClient ([#616](https://github.com/keptn-contrib/dynatrace-service/issues/616)) ([c6a6d91](https://github.com/keptn-contrib/dynatrace-service/commit/c6a6d914a5252fda6a8ea6cd601c7553adc8abff))
* Use correct timeframe for SLIs based on Dynatrace SLOs ([#645](https://github.com/keptn-contrib/dynatrace-service/issues/645)) ([032155a](https://github.com/keptn-contrib/dynatrace-service/commit/032155a78221278494312e5e183d5c85f9329dda))


### Other

* Added semantic PR checks ([#615](https://github.com/keptn-contrib/dynatrace-service/issues/615)) ([7609e63](https://github.com/keptn-contrib/dynatrace-service/commit/7609e638e08e8b78aa1e0b8c10d17d8b9d8f4b11))
* Bump k8s.io/api, k8s.io/client-go and k8s.io/apimachinery to 0.23.0 ([#622](https://github.com/keptn-contrib/dynatrace-service/issues/622)) ([044db99](https://github.com/keptn-contrib/dynatrace-service/commit/044db99ca51a317268600fdebe5492eff3430f76))


### Docs

* Refactor documentation ([#632](https://github.com/keptn-contrib/dynatrace-service/issues/632)) ([eadfaee](https://github.com/keptn-contrib/dynatrace-service/commit/eadfaee8e8a4baed7baa9c0e5bf01e3ec256cdd0))
