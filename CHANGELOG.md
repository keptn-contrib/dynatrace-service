# Changelog

All notable changes to this project will be documented in this file. See [standard-version](https://github.com/conventional-changelog/standard-version) for commit guidelines.

## [0.26.0](https://github.com/keptn-contrib/dynatrace-service/compare/0.25.0...0.26.0) (2022-12-21)


### Bug Fixes

* Disallow duplicate SLI names and display names for dashboard SLIs ([#969](https://github.com/keptn-contrib/dynatrace-service/issues/969)) ([ec26fb8](https://github.com/keptn-contrib/dynatrace-service/commit/ec26fb8e674f1c4c00b1f4249cf5a35d7d77c081))
* Support display units for metrics with `Count` and `Unspecified` units ([#971](https://github.com/keptn-contrib/dynatrace-service/issues/971)) ([6da09cc](https://github.com/keptn-contrib/dynatrace-service/commit/6da09cc27797b44675d6429224526329e0f42666))
* Warnings from informational SLOs should not affect overall result of `get-sli.finished` events ([#974](https://github.com/keptn-contrib/dynatrace-service/issues/974)) ([dc9b415](https://github.com/keptn-contrib/dynatrace-service/commit/dc9b41542fe670471e5a23911045913bb2a13e45))


### Other

* release 0.26.0 ([#983](https://github.com/keptn-contrib/dynatrace-service/issues/983)) ([4d5991f](https://github.com/keptn-contrib/dynatrace-service/commit/4d5991f87889a9f0cd64818baf3f0603d3acb80c))
* Remove automatic problem remediation SLI and SLO from `get-sli.triggered` handler ([#973](https://github.com/keptn-contrib/dynatrace-service/issues/973)) ([3d9f666](https://github.com/keptn-contrib/dynatrace-service/commit/3d9f6660bfabdcdd15585e10d054404deca6aec2))
* Update tests for errors in get-sli.triggered handling ([#980](https://github.com/keptn-contrib/dynatrace-service/issues/980)) ([9dc318b](https://github.com/keptn-contrib/dynatrace-service/commit/9dc318b8676b9ee63a18d179f5871c26b8b5aa2e))

## [0.25.0](https://github.com/keptn-contrib/dynatrace-service/compare/0.24.0...0.25.0) (2022-11-15)

### Release validated with
 | Dynatrace-service: `0.25.0` | Keptn: `0.19.3` | Dynatrace: `1.254` |
 |---|---|---|


### Features

* Use management zone names rather than IDs ([#957](https://github.com/keptn-contrib/dynatrace-service/issues/957)) ([418c854](https://github.com/keptn-contrib/dynatrace-service/commit/418c8540eaaa9a2ca53e693293227eb7758f55fe))


### Bug Fixes

* Correctly encode requests made by `dynatrace.EntitiesClient` ([#955](https://github.com/keptn-contrib/dynatrace-service/issues/955)) ([93a01f7](https://github.com/keptn-contrib/dynatrace-service/commit/93a01f75df6d199258be740c75b87d65167aa9d7))
* Improve Data explorer unit support ([#959](https://github.com/keptn-contrib/dynatrace-service/issues/959)) ([6fa0e94](https://github.com/keptn-contrib/dynatrace-service/commit/6fa0e948d9cbc7c1cd6fa9e493c8377e5fae1d25))


### Docs

* Clarify unit conversion for SLIs ([#951](https://github.com/keptn-contrib/dynatrace-service/issues/951)) ([694742f](https://github.com/keptn-contrib/dynatrace-service/commit/694742f96ca363109224d097d053e885698f44f3))

## [0.24.0](https://github.com/keptn-contrib/dynatrace-service/compare/0.23.0...0.24.0) (2022-10-25)

### Release validated with
 | Dynatrace-service: `0.24.0` | Keptn: `0.19.1` | Dynatrace: `1.252` |
 |---|---|---|


### ⚠ BREAKING CHANGES

* Support resolution parameter (#916)
* add support for keptn 0.19 (#914)
* Use data explorer metric expressions (#883)
* Generate SLO pass and warning criteria from Data Explorer thresholds (#846)
* Remove automatic conversion of `builtin:service.response.time` as well as microsecond and byte units (#867)
* Do not create `dynatrace/sli.yaml` when retrieving SLIs from a dashboard (#866)
* Change dashboard tile title parsing logic (#844)

### Features

* Add dimension name to display name for USQL tiles ([#904](https://github.com/keptn-contrib/dynatrace-service/issues/904)) ([6cfa146](https://github.com/keptn-contrib/dynatrace-service/commit/6cfa1460bb380887deb1163d97f12509dfe3ac9f))
* add support for keptn 0.19 ([#914](https://github.com/keptn-contrib/dynatrace-service/issues/914)) ([d4a1f1a](https://github.com/keptn-contrib/dynatrace-service/commit/d4a1f1a103d26cb2535fc23820bccd773369fe3d))
* Change dashboard tile title parsing logic ([#844](https://github.com/keptn-contrib/dynatrace-service/issues/844)) ([48721ea](https://github.com/keptn-contrib/dynatrace-service/commit/48721ea51354b7fa172d620d8a46644887937c79))
* Do not create `dynatrace/sli.yaml` when retrieving SLIs from a dashboard ([#866](https://github.com/keptn-contrib/dynatrace-service/issues/866)) ([c90efc2](https://github.com/keptn-contrib/dynatrace-service/commit/c90efc27904f5b6ddb28cb935946b458d754c3d0))
* Generate SLO pass and warning criteria from Data Explorer thresholds ([#846](https://github.com/keptn-contrib/dynatrace-service/issues/846)) ([68f95ee](https://github.com/keptn-contrib/dynatrace-service/commit/68f95eed77f440bc35e56e7fbab389a3ed2a36ae))
* Remove automatic conversion of `builtin:service.response.time` as well as microsecond and byte units ([#867](https://github.com/keptn-contrib/dynatrace-service/issues/867)) ([9c17484](https://github.com/keptn-contrib/dynatrace-service/commit/9c17484d9ff47fd32b5548f5a845f990c5aeeaa8))
* Support resolution parameter ([#916](https://github.com/keptn-contrib/dynatrace-service/issues/916)) ([1c1e07d](https://github.com/keptn-contrib/dynatrace-service/commit/1c1e07dbf26a32cd1b4c0e35428a8727808f2b7b))
* Support units in Data Explorer and Custom Charting tiles ([#939](https://github.com/keptn-contrib/dynatrace-service/issues/939)) ([e32dd8a](https://github.com/keptn-contrib/dynatrace-service/commit/e32dd8a3134f9e9d085d21a5d4922692eb437cc4))
* Use data explorer metric expressions ([#883](https://github.com/keptn-contrib/dynatrace-service/issues/883)) ([9012758](https://github.com/keptn-contrib/dynatrace-service/commit/9012758156f4906d7e6edbe2a5ad1551699992fc))


### Bug Fixes

* Explicitly ignore `sh.keptn.event.get-sli.triggered` events not for Dynatrace ([#863](https://github.com/keptn-contrib/dynatrace-service/issues/863)) ([5f10284](https://github.com/keptn-contrib/dynatrace-service/commit/5f10284bfef1d5d09704e4a5d60d490257d52328))
* Generate specific error message if no SLIs are requested when using file-based SLIs ([#901](https://github.com/keptn-contrib/dynatrace-service/issues/901)) ([8722c36](https://github.com/keptn-contrib/dynatrace-service/commit/8722c368c07830bfacf3bf9e6b42c522a072340b))
* Support auto space aggregation ([#875](https://github.com/keptn-contrib/dynatrace-service/issues/875)) ([ddfadcf](https://github.com/keptn-contrib/dynatrace-service/commit/ddfadcff9a656974fe4e0d81c54c94602de3fce8))
* Update create release pr pull request title pattern ([#946](https://github.com/keptn-contrib/dynatrace-service/issues/946)) ([8cecf92](https://github.com/keptn-contrib/dynatrace-service/commit/8cecf92ca490c226edc062f23c99cd9a9334cc54))
* Update links to point to Keptn 0.16.x docs ([#865](https://github.com/keptn-contrib/dynatrace-service/issues/865)) ([173a66d](https://github.com/keptn-contrib/dynatrace-service/commit/173a66d0843f06724bf3309ec7748dbd8b91d33d))


### Docs

* fix broken link in README.md ([#860](https://github.com/keptn-contrib/dynatrace-service/issues/860)) ([93922de](https://github.com/keptn-contrib/dynatrace-service/commit/93922de9cfeb0d35b80f33d62e8d59fc1a96f1a3))
* removes duplicate sections ([#861](https://github.com/keptn-contrib/dynatrace-service/issues/861)) ([5f9b1ed](https://github.com/keptn-contrib/dynatrace-service/commit/5f9b1ed9a9552942542b7784a827d2abdc2b5d44))
* Update documentation for 0.24.0 release ([#941](https://github.com/keptn-contrib/dynatrace-service/issues/941)) ([fc3e1f3](https://github.com/keptn-contrib/dynatrace-service/commit/fc3e1f346e984a42d61fb7f58dcc9733180e5428))


### Other

* Add keptn contrib bot to codeowners ([#943](https://github.com/keptn-contrib/dynatrace-service/issues/943)) ([f730ad1](https://github.com/keptn-contrib/dynatrace-service/commit/f730ad190422b3116bd66a4278012aa27c619ad3))
* Add tests for handling of whitespace in key-value pairs in dashboard markdown tiles ([#871](https://github.com/keptn-contrib/dynatrace-service/issues/871)) ([ca2680d](https://github.com/keptn-contrib/dynatrace-service/commit/ca2680d47efb53256ce325aa74166b67543bdc45))
* **deps:** add renovate.json ([9c1716b](https://github.com/keptn-contrib/dynatrace-service/commit/9c1716b04df3860b57e25934f46276364d11b982))
* Remove keptn-contrib-bot from codeowners ([#945](https://github.com/keptn-contrib/dynatrace-service/issues/945)) ([86f80b1](https://github.com/keptn-contrib/dynatrace-service/commit/86f80b1521ddf2c41d99d363a33c89cd4df527d9))
* Update CODEOWNERS ([#872](https://github.com/keptn-contrib/dynatrace-service/issues/872)) ([e29e896](https://github.com/keptn-contrib/dynatrace-service/commit/e29e8966fed391cc434f368fcf43bd595882a009))
* Update codeowners ([#927](https://github.com/keptn-contrib/dynatrace-service/issues/927)) ([8ddbd6e](https://github.com/keptn-contrib/dynatrace-service/commit/8ddbd6e4af73b04ec82fcf77e5464167dc65b663))

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
