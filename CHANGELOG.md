# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

### [0.23.2](https://github.com/keptn-contrib/dynatrace-service/compare/0.23.1...0.23.2) (2022-08-16)

### Release validated with
 | Dynatrace-service: `0.23.2` | Keptn: `0.16.1` | Dynatrace: `1.247` |
 |---|---|---|


### Bug Fixes

* Support auto space aggregation ([#876](https://github.com/keptn-contrib/dynatrace-service/issues/876)) ([3fb43be](https://github.com/keptn-contrib/dynatrace-service/commit/3fb43be6dad9fd20c725740eb4b03563db7ec912))

### [0.23.1](https://github.com/keptn-contrib/dynatrace-service/compare/0.23.0...0.23.1) (2022-08-03)

### Release validated with
 | Dynatrace-service: `0.23.1` | Keptn: `0.16.1` | Dynatrace: `1.245` |
 |---|---|---|


### Bug Fixes

* Trim whitespace from key-value pairs in markdown tiles ([#868](https://github.com/keptn-contrib/dynatrace-service/issues/868)) ([2449853](https://github.com/keptn-contrib/dynatrace-service/commit/2449853eecbe359fbfce12264130d6aac4921b9c))

## [0.23.0](https://github.com/keptn-contrib/dynatrace-service/compare/0.22.0...0.23.0) (2022-07-13)

### Release validated with
 | Dynatrace-service: `0.23.0` | Keptn: `0.16.1` | Dynatrace: `1.245` |
 |---|---|---|


### ⚠ BREAKING CHANGES

* parsing errors in markdown tiles for SLO configuration will now return `result=fail` in the event payload.
* security problems are no longer considered when evaluating a problem tile of a Dynatrace dashboard

### Features

* Add context to Keptn clients ([#792](https://github.com/keptn-contrib/dynatrace-service/issues/792)) ([2408c78](https://github.com/keptn-contrib/dynatrace-service/commit/2408c78ec4bace58d68065e09f4b9672ff278bfa))
* add new custom property `evaluationHeatmapURL` to evaluation finished events ([#841](https://github.com/keptn-contrib/dynatrace-service/issues/841)) ([dc53e66](https://github.com/keptn-contrib/dynatrace-service/commit/dc53e668b4169ec310c8772a594e4c512566270b))
* Initial support for graceful shutdown ([#789](https://github.com/keptn-contrib/dynatrace-service/issues/789)) ([6daf187](https://github.com/keptn-contrib/dynatrace-service/commit/6daf18788e6aad150dc4343db436222c009930c0))
* send evaluation finished events to PGI level ([#822](https://github.com/keptn-contrib/dynatrace-service/issues/822)) ([6fcb16b](https://github.com/keptn-contrib/dynatrace-service/commit/6fcb16b9288eb0e961d12e983b09f00e9e0c1c13))
* try to push all events to pgi level ([#847](https://github.com/keptn-contrib/dynatrace-service/issues/847)) ([91f6140](https://github.com/keptn-contrib/dynatrace-service/commit/91f614056577ddfab48ef044f63f49eea6e78207))
* Use cp-connector instead of distributor ([#817](https://github.com/keptn-contrib/dynatrace-service/issues/817)) ([9fb6b1e](https://github.com/keptn-contrib/dynatrace-service/commit/9fb6b1eb3ef77bc471b802f20a140149d82548ca))
* Use keptn/go-utils to access Keptn APIs ([#767](https://github.com/keptn-contrib/dynatrace-service/issues/767)) ([a2a798a](https://github.com/keptn-contrib/dynatrace-service/commit/a2a798a62c9ee64de27c80b421082f6a440feee9))
* Use multiple explicit event subscriptions rather than wildcard ([#828](https://github.com/keptn-contrib/dynatrace-service/issues/828)) ([c940513](https://github.com/keptn-contrib/dynatrace-service/commit/c9405136269c9e4b879c163cfa5760a0958c61ab))


### Bug Fixes

* Data Explorer tiles using a management zone should produce an error if no entity type can be determined ([#804](https://github.com/keptn-contrib/dynatrace-service/issues/804)) ([40f7fc7](https://github.com/keptn-contrib/dynatrace-service/commit/40f7fc738acd895beb4d685fb6bd1bff94609f73))
* description cannot be empty for event in release triggered handler ([#820](https://github.com/keptn-contrib/dynatrace-service/issues/820)) ([440b159](https://github.com/keptn-contrib/dynatrace-service/commit/440b159eb0642a9e0b6414605aab83bd21131d5a))
* Disconnect `nats.NatsConnector` on shutdown ([#852](https://github.com/keptn-contrib/dynatrace-service/issues/852)) ([0d9494b](https://github.com/keptn-contrib/dynatrace-service/commit/0d9494b2672b37486e47932f6deeb35154bef00f))
* do not query security problems for problem tiles when retrieving SLIs via dashboards ([#759](https://github.com/keptn-contrib/dynatrace-service/issues/759)) ([963d74b](https://github.com/keptn-contrib/dynatrace-service/commit/963d74b2126f6a3a324537f007696f565bffc53f))
* invalid syntax in dashboard tile titles should return error ([#800](https://github.com/keptn-contrib/dynatrace-service/issues/800)) ([c93c862](https://github.com/keptn-contrib/dynatrace-service/commit/c93c862bd1007054a447cbab35784947697ea643))
* Make K8S_NAMESPACE and K8S_NODE_NAME configureable for cp-connector ([#850](https://github.com/keptn-contrib/dynatrace-service/issues/850)) ([5240f78](https://github.com/keptn-contrib/dynatrace-service/commit/5240f785d35b3b4055073db2478ff486bcf30d14))
* markdown processing returns errors ([#802](https://github.com/keptn-contrib/dynatrace-service/issues/802)) ([e68b556](https://github.com/keptn-contrib/dynatrace-service/commit/e68b5561868c0547a481691ae7831e1db6a094a8))
* Move `GetSLIs()` and `GetShipyard()` to `keptn.ConfigClient` ([#806](https://github.com/keptn-contrib/dynatrace-service/issues/806)) ([9f3e01d](https://github.com/keptn-contrib/dynatrace-service/commit/9f3e01df59b5a544fde564c59d56a5f992bab891))
* Refactor service onboarder ([#755](https://github.com/keptn-contrib/dynatrace-service/issues/755)) ([a773756](https://github.com/keptn-contrib/dynatrace-service/commit/a773756b180d479ba958e6266072eb16b73a622a))
* Refactor service onboarder tests ([#760](https://github.com/keptn-contrib/dynatrace-service/issues/760)) ([4661475](https://github.com/keptn-contrib/dynatrace-service/commit/46614756e3f6930144cdb7eaaca0cc6dc7645036))
* Use correct URLs in Keptn API clients ([#783](https://github.com/keptn-contrib/dynatrace-service/issues/783)) ([479d2f2](https://github.com/keptn-contrib/dynatrace-service/commit/479d2f22dc165c87beff19db4773aa78acb54617))


### Docs

* add troubleshooting for subscriptions ([#756](https://github.com/keptn-contrib/dynatrace-service/issues/756)) ([2adc84a](https://github.com/keptn-contrib/dynatrace-service/commit/2adc84aadfcf32bd3b69aaed0471ddf919ab5e23))
* Added a note to not exclude http usage ([#770](https://github.com/keptn-contrib/dynatrace-service/issues/770)) ([5baa6f5](https://github.com/keptn-contrib/dynatrace-service/commit/5baa6f50c7d5df3275c36a9532263feb0c8a5ee6))
* adding documentation for events on PGI level ([#842](https://github.com/keptn-contrib/dynatrace-service/issues/842)) ([03d84e2](https://github.com/keptn-contrib/dynatrace-service/commit/03d84e271d01a8acfcd80bec43998f693d9498bf)), closes [#848](https://github.com/keptn-contrib/dynatrace-service/issues/848)
* Fix compatibility ([f898006](https://github.com/keptn-contrib/dynatrace-service/commit/f898006b1579b3849c28115e4e82bd411287a157))
* Provide a workaround for Keptn 0.14.1 installation ([#766](https://github.com/keptn-contrib/dynatrace-service/issues/766)) ([211102f](https://github.com/keptn-contrib/dynatrace-service/commit/211102f9d536500ed10acf912140a28efc8ca7ba))
* update compatibility matrix for previous release ([#854](https://github.com/keptn-contrib/dynatrace-service/issues/854)) ([f2fad37](https://github.com/keptn-contrib/dynatrace-service/commit/f2fad377432fe730f8f432e37fef89b85700a40c))
* Update doc to be aligned with official Dynatrace documentation ([#768](https://github.com/keptn-contrib/dynatrace-service/issues/768)) ([461bd13](https://github.com/keptn-contrib/dynatrace-service/commit/461bd13e741191afe2648cdefc4a0e3d0edead15))
* Use `spec_version` ([#758](https://github.com/keptn-contrib/dynatrace-service/issues/758)) ([cbd69f6](https://github.com/keptn-contrib/dynatrace-service/commit/cbd69f693ab1ec84c0928e699818574b304074a2))


### Other

* Clean up Keptn dependencies ([#839](https://github.com/keptn-contrib/dynatrace-service/issues/839)) ([1038696](https://github.com/keptn-contrib/dynatrace-service/commit/10386964574eaa4e28b350882af39e70d8a05ec6))
* remove warning for distributor from README.md ([#855](https://github.com/keptn-contrib/dynatrace-service/issues/855)) ([4cb9035](https://github.com/keptn-contrib/dynatrace-service/commit/4cb90353c92ef2e887612288f80128da9cd6ef2f))
* Switch to golangci-lint ([#837](https://github.com/keptn-contrib/dynatrace-service/issues/837)) ([fdf4503](https://github.com/keptn-contrib/dynatrace-service/commit/fdf45031220e7565390e072f1234ef8d921c492f))
* Update `go-utils`, `cp-common` and `cp-connector` ([#835](https://github.com/keptn-contrib/dynatrace-service/issues/835)) ([bb9eb8f](https://github.com/keptn-contrib/dynatrace-service/commit/bb9eb8f7e178960e349aaa434ee45a27af973957))
* Update to `gopkg.in/yaml.v3` ([#816](https://github.com/keptn-contrib/dynatrace-service/issues/816)) ([9bc11cc](https://github.com/keptn-contrib/dynatrace-service/commit/9bc11cc6234717a836fc36707813f053654c0551))
* Update to `keptn/go-utils v0.16.1-0.20220628141633-eb5fb9ba43e0` ([#840](https://github.com/keptn-contrib/dynatrace-service/issues/840)) ([e22d438](https://github.com/keptn-contrib/dynatrace-service/commit/e22d43884b0bb726571b200073cb40f5575c5a96))
* Update to Go 1.18 ([#833](https://github.com/keptn-contrib/dynatrace-service/issues/833)) ([d7e8dfc](https://github.com/keptn-contrib/dynatrace-service/commit/d7e8dfc7af7a64efa6011ef834a01bc5414b5c01))

## [0.22.0](https://github.com/keptn-contrib/dynatrace-service/compare/0.21.0...0.22.0) (2022-03-23)

### Release validated with
 | Dynatrace-service: `0.22.0` | Keptn: `0.12.6`* | Dynatrace: `1.235` |
 |---|---|---|

&ast; **Note**: to install dynatrace-service 0.22.0 for Keptn 0.14.1 or later, please override the bundled distributor version and target the appropriate Keptn version by setting the Helm chart variable `distributor.image.tag`, i.e. by appending `--set distributor.image.tag=...` during the Helm upgrade.

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
