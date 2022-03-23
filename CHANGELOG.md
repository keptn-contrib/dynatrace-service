# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

## [0.22.0](https://github.com/keptn-contrib/dynatrace-service/compare/0.21.0...0.22.0) (2022-03-23)

### Release validated with
 | Dynatrace-service: `0.22.0` | Keptn: `0.12.6` | Dynatrace: `1.235` |
 |---|---|---|


### Features

* Improve handling of SLI queries that don't produce a single value ([#733](https://github.com/keptn-contrib/dynatrace-service/issues/733)) ([0c4208c](https://github.com/keptn-contrib/dynatrace-service/commit/0c4208c15b6acf2fa076ae5f477527b8af1cdee1))


### Bug Fixes

* Delay calls to Dynatrace APIs such that required data is available ([#723](https://github.com/keptn-contrib/dynatrace-service/issues/723)) ([05467e8](https://github.com/keptn-contrib/dynatrace-service/commit/05467e80b0eb87a7f0e621580cfa3236663819bd))
* Explicitly use tile properties to derive SLO definitions ([#747](https://github.com/keptn-contrib/dynatrace-service/issues/747)) ([9a4a27b](https://github.com/keptn-contrib/dynatrace-service/commit/9a4a27b458274ca3c4e83e4e5b5eb15367e26415))
* Remove helm chart option for optional dynatrace-service container deployment ([#720](https://github.com/keptn-contrib/dynatrace-service/issues/720)) ([4096213](https://github.com/keptn-contrib/dynatrace-service/commit/409621374c9232c6699bce52dae3718621090d09))
* Return an error if multiple dashboards match query ([#743](https://github.com/keptn-contrib/dynatrace-service/issues/743)) ([d390219](https://github.com/keptn-contrib/dynatrace-service/commit/d3902198c4552a0e6c4713d79ce727416e81296a))


### Other

* Update CI badge ([dfc3a89](https://github.com/keptn-contrib/dynatrace-service/commit/dfc3a899c2ff67a9336e7aa44dfed580bd9d0327))
* Update CODEOWNERS ([b6d06ff](https://github.com/keptn-contrib/dynatrace-service/commit/b6d06ff72e8396c626dfef5913355a2f600b9740))


### Docs

* Add cross-links to `dashboard` entry in `dynatrace/dynatrace.conf.yaml` ([#727](https://github.com/keptn-contrib/dynatrace-service/issues/727)) ([c68e604](https://github.com/keptn-contrib/dynatrace-service/commit/c68e6049ec225f1f6631f75183b57a31351f00bf))
* Add initial troubleshooting guide to documentation ([#726](https://github.com/keptn-contrib/dynatrace-service/issues/726)) ([dd34da4](https://github.com/keptn-contrib/dynatrace-service/commit/dd34da41c5211d361d36279d59b7a2b5892b2300))
* Improve troubleshooting content for evaluation failed ([#734](https://github.com/keptn-contrib/dynatrace-service/issues/734)) ([4fbd86a](https://github.com/keptn-contrib/dynatrace-service/commit/4fbd86a28d861ddb01e986ea6d063680d6dbd271))

## [0.21.0](https://github.com/keptn-contrib/dynatrace-service/compare/0.20.0...0.21.0) (2022-02-21)

### Release validated with
 | Dynatrace-service: `0.21.0` | Keptn: `0.12.2` | Dynatrace: `1.234` |
 |---|---|---|


### Features

* Add support for the line chart visualization type in USQL queries and dashboard tiles ([#716](https://github.com/keptn-contrib/dynatrace-service/issues/716)) ([29ea965](https://github.com/keptn-contrib/dynatrace-service/commit/29ea965634f58efc04f8b78214ac95edc5eb71db))
* Support placeholders in all SLIs defined in `dynatrace/sli.yaml` files ([#681](https://github.com/keptn-contrib/dynatrace-service/issues/681)) ([6dd69d9](https://github.com/keptn-contrib/dynatrace-service/commit/6dd69d91bc31c64cbf8820836badac299c073331))


### Bug Fixes

* Ensure problem tile processing always produces indicators ([#707](https://github.com/keptn-contrib/dynatrace-service/issues/707)) ([1027225](https://github.com/keptn-contrib/dynatrace-service/commit/102722563dfe3885f6d82f741bfde4d5a54aae76))
* Ensure SLO tile processing always produces an indicator ([#706](https://github.com/keptn-contrib/dynatrace-service/issues/706)) ([0f9870a](https://github.com/keptn-contrib/dynatrace-service/commit/0f9870a961cbf113c4e8a3c20ce8960b6a8c9ac1))
* Ensure USQL tile processing always produces an indicator ([#710](https://github.com/keptn-contrib/dynatrace-service/issues/710)) ([b3ca3d1](https://github.com/keptn-contrib/dynatrace-service/commit/b3ca3d1a9f6dbd429970ad49384e9a7353bf3417))
* error messages are no longer attached to indicator from event if dashboard processing fails ([#687](https://github.com/keptn-contrib/dynatrace-service/issues/687)) ([51e1e9a](https://github.com/keptn-contrib/dynatrace-service/commit/51e1e9a0d45ea2911b3d3d247ff6ecda888849b6))


### Other

* Improve CI pipeline and make unit tests reusable ([#675](https://github.com/keptn-contrib/dynatrace-service/issues/675)) ([f9c1bec](https://github.com/keptn-contrib/dynatrace-service/commit/f9c1bec32a1bd7e56903f992c80ff61ec8679f77))


### Docs

* Document Dynatrace API token scopes ([#701](https://github.com/keptn-contrib/dynatrace-service/issues/701)) ([5e93933](https://github.com/keptn-contrib/dynatrace-service/commit/5e93933d1dd20a851e7e929f953c374bc6a2e3de))
* Update documentation links ([#702](https://github.com/keptn-contrib/dynatrace-service/issues/702)) ([20cd1d0](https://github.com/keptn-contrib/dynatrace-service/commit/20cd1d0383c170ab803a35f507138c2167451637))
* Use KEPTN_ENDPOINT in installation documentation ([#700](https://github.com/keptn-contrib/dynatrace-service/issues/700)) ([88860fd](https://github.com/keptn-contrib/dynatrace-service/commit/88860fdff039e9170f1e570da9f1c59f1c9cd7d3))

## [0.20.0](https://github.com/keptn-contrib/dynatrace-service/compare/0.19.0...0.20.0) (2022-01-17)

### Release validated with
 | Dynatrace-service: `0.20.0` | Keptn: `0.11.4` | Dynatrace: `1.233` |
 |---|---|---|


### Features

* Forward all Dynatrace problem details ([#665](https://github.com/keptn-contrib/dynatrace-service/issues/665)) ([dc04c6d](https://github.com/keptn-contrib/dynatrace-service/commit/dc04c6dccc2fc0eb89991855beb095a5519552ac))


### Bug Fixes

*  Ensure problem notifications created using `keptn configure monitoring` refer to a valid project ([#671](https://github.com/keptn-contrib/dynatrace-service/issues/671)) ([fc9bdc5](https://github.com/keptn-contrib/dynatrace-service/commit/fc9bdc5b3c88ac87bdbbfa8998fb7e9e372e6f6f))
*  Only support Keptn placeholders in values in dynatrace/dynatrace.conf.yaml where it makes sense ([#654](https://github.com/keptn-contrib/dynatrace-service/issues/654)) ([ce16c01](https://github.com/keptn-contrib/dynatrace-service/commit/ce16c01937330f02f5c34ba8bde44ed78ade8198))
* `ProblemEventHandler` sends `sh.keptn.event.[STAGE].remediation.triggered` event even if stage is not set  ([#672](https://github.com/keptn-contrib/dynatrace-service/issues/672)) ([ac06bf8](https://github.com/keptn-contrib/dynatrace-service/commit/ac06bf86cb297401236dffa11761e0fe6c3bdb66))
* Improve errors when unable to process events ([#679](https://github.com/keptn-contrib/dynatrace-service/issues/679)) ([b0f024c](https://github.com/keptn-contrib/dynatrace-service/commit/b0f024c6b91f4f4d0e1bd51e54546e87ac74ed1a))
* ProblemEventHandler forwards wrong events  ([#664](https://github.com/keptn-contrib/dynatrace-service/issues/664)) ([1663b77](https://github.com/keptn-contrib/dynatrace-service/commit/1663b77a7fc3cf3ecff6c2da639cc9a4eacc76bf))
* Remove automatic configure monitoring for new projects  ([#661](https://github.com/keptn-contrib/dynatrace-service/issues/661)) ([e5405eb](https://github.com/keptn-contrib/dynatrace-service/commit/e5405eb02f48539fa87c635eb865f156ba8fa62f))
* Service entities tagged with multiple `keptn_service` tags should produce an error ([#673](https://github.com/keptn-contrib/dynatrace-service/issues/673)) ([d542669](https://github.com/keptn-contrib/dynatrace-service/commit/d54266985ab74583bda38fbc042c171c027de0e1))
* Use event type as task for non-task events ([#670](https://github.com/keptn-contrib/dynatrace-service/issues/670)) ([b112d2c](https://github.com/keptn-contrib/dynatrace-service/commit/b112d2cd24dcca9f325376adc032e88b87a49888))
* USQL processing will not panic in case of errors ([#677](https://github.com/keptn-contrib/dynatrace-service/issues/677)) ([111d569](https://github.com/keptn-contrib/dynatrace-service/commit/111d569e2598c1db481faff4eced499db57c15b1))


### Refactoring

* Move `HTTPGetHandler` to `health` package ([#660](https://github.com/keptn-contrib/dynatrace-service/issues/660)) ([26a06d0](https://github.com/keptn-contrib/dynatrace-service/commit/26a06d0b55da5ddad3290630f36fc6ef25ffa2dd))


### Docs

* Fix management zone criterion in PV2 entity selector ([#662](https://github.com/keptn-contrib/dynatrace-service/issues/662)) ([bd5955f](https://github.com/keptn-contrib/dynatrace-service/commit/bd5955ff57dce6dfea8c876120c18ad505343c97))


### Other

* Re-use docker-build action from keptn/gh-automation ([#676](https://github.com/keptn-contrib/dynatrace-service/issues/676)) ([745885e](https://github.com/keptn-contrib/dynatrace-service/commit/745885e5bedf4f0ff53f888acb1fde9cf7537641))

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
