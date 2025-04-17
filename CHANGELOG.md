# Changelog

## [0.3.19](https://github.com/leg100/otf/compare/v0.3.18...v0.3.19) (2025-04-16)


### Features

* **docs:** add search ([aa5d43b](https://github.com/leg100/otf/commit/aa5d43b6fdbd8115628b11bd6425bd63d3fc9aa4))
* **ui:** move theme selecter to navbar ([af37fe7](https://github.com/leg100/otf/commit/af37fe7f5b25121b599bf92fec9230b12701f2eb))
* **ui:** re-style workspace main page ([d0084b5](https://github.com/leg100/otf/commit/d0084b595f92a08e788339acc1595521ba1b4301))


### Bug Fixes

* allow email address for username ([ecf3a3d](https://github.com/leg100/otf/commit/ecf3a3ddf09f09cf87864cd6315074245ec94ead))


### Miscellaneous

* tidy up navbar ([7a56987](https://github.com/leg100/otf/commit/7a56987bba515c3c048e77a87b1d0358ddf3e928))
* update docs ([4e971dc](https://github.com/leg100/otf/commit/4e971dcf8d7f004d25be88667f5e7943c135dfb7))

## [0.3.18](https://github.com/leg100/otf/compare/v0.3.17...v0.3.18) (2025-04-13)


### Features

* **ui:** add side menu ([d628f81](https://github.com/leg100/otf/commit/d628f81f16fac260a089d3a1bcef7d3c4ddfe3c9))
* **ui:** add theme selector ([97a4aa1](https://github.com/leg100/otf/commit/97a4aa18371a9cdcd849360329b2cd2f3ccc1b16))
* **ui:** make lock widget nicer ([7df67d9](https://github.com/leg100/otf/commit/7df67d949b5caf92757f5104e03d42be4373de67))


### Bug Fixes

* allow variable set with same name in another org ([8834ba2](https://github.com/leg100/otf/commit/8834ba26718ab2c9e2556afcc4ff55b20bf0353c))
* **ui:** set max width ([cceb0d3](https://github.com/leg100/otf/commit/cceb0d3631404659af8e3b0aac5ce415816a0e1c))
* **ui:** stale statuses on run page ([c4f3a65](https://github.com/leg100/otf/commit/c4f3a65a8867258c2125ef07db6cb169b13c80ea))


### Miscellaneous

* bump playwright ([25313e2](https://github.com/leg100/otf/commit/25313e2e09fea51e22553ad06d43952cce75e8ad))

## [0.3.17](https://github.com/leg100/otf/compare/v0.3.16...v0.3.17) (2025-04-08)


### Features

* **ui:** improve styling ([06a6bce](https://github.com/leg100/otf/commit/06a6bce864225346ca00c6725bac3b66b06f56d9))
* **ui:** tabulate all resources ([#754](https://github.com/leg100/otf/issues/754)) ([fcc78be](https://github.com/leg100/otf/commit/fcc78bec59291b4bd70364dc29e6fcad401be56c))


### Bug Fixes

* add logs index to prevent workspace delete timeout ([e12dec5](https://github.com/leg100/otf/commit/e12dec56e850f2fafbafbf3be1ad8586f41a4b7b))
* **ui:** broken workspace lock widget for non-site-admins ([37c5d1d](https://github.com/leg100/otf/commit/37c5d1d04718922f8f18549c66acd20704fa1226))
* **ui:** express durations &gt; 24h as days not hours ([a37f826](https://github.com/leg100/otf/commit/a37f826e7222b9493bc60fc2765a2b56edaec125))


### Miscellaneous

* bump sqlc version ([96007b3](https://github.com/leg100/otf/commit/96007b331fa64064fac6f4421f6e1f00e626b2ac))
* introduce resource.ID interface ([454209f](https://github.com/leg100/otf/commit/454209fcdc34fe63bdd2150746c53d674fa21de5))
* screenshot test failures ([06443b1](https://github.com/leg100/otf/commit/06443b1034664ed4a4584171d5b6e511909daf64))

## [0.3.16](https://github.com/leg100/otf/compare/v0.3.15...v0.3.16) (2025-03-17)


### Features

* **ui:** tabulate workspace and run listings ([#741](https://github.com/leg100/otf/issues/741)) ([31ab7ce](https://github.com/leg100/otf/commit/31ab7ce8130171ee586a0ddbaff89a6e15248ddf))


### Bug Fixes

* add arm and 386 arch testdata ([#740](https://github.com/leg100/otf/issues/740)) ([e24cdfa](https://github.com/leg100/otf/commit/e24cdfa957d0f6170f891052c7a1233e01747783))
* **ui:** various UI fixes ([a368efe](https://github.com/leg100/otf/commit/a368efe138fd37ed2dea46c4ff67b2df74ad403c))

## [0.3.15](https://github.com/leg100/otf/compare/v0.3.14...v0.3.15) (2025-03-13)


### Features

* order workspaces lexicographically ([66a2b00](https://github.com/leg100/otf/commit/66a2b00665cfe45b81038b4a374ac639e53a0c71))
* **ui:** filter runs by status ([#737](https://github.com/leg100/otf/issues/737)) ([8780849](https://github.com/leg100/otf/commit/8780849e57f45568decdc64e7d4e70a0a1132316))
* **ui:** organization runs view ([#739](https://github.com/leg100/otf/issues/739)) ([dad98f7](https://github.com/leg100/otf/commit/dad98f7ccb1e46ba1c1dd4d507f9eeb3b13a3704))


### Bug Fixes

* **ui:** pagination broken on some pages ([cc6d970](https://github.com/leg100/otf/commit/cc6d970197a8ee80a99aa8ac5ebc1c0fba9f1e2a))

## [0.3.14](https://github.com/leg100/otf/compare/v0.3.13...v0.3.14) (2025-03-07)


### Features

* **ui:** add filter for workspace current run status ([#734](https://github.com/leg100/otf/issues/734)) ([f3b2c85](https://github.com/leg100/otf/commit/f3b2c85dd19973c5afa04a82987fc6a3db468584))
* **ui:** add page size selector ([#733](https://github.com/leg100/otf/issues/733)) ([06efdf5](https://github.com/leg100/otf/commit/06efdf53150145cf07189b33e000c4170d4e433b))


### Bug Fixes

* bump tailwind config to be compat with v4 ([402903d](https://github.com/leg100/otf/commit/402903d42023db6805965fe6b685587a1167eef1))
* gh app webhook panics when no gh app configured ([48a4b24](https://github.com/leg100/otf/commit/48a4b24f1c9de0505dc34282f7b35b424ba25de6))
* report cfg error when starting run without cfg ([a39cc65](https://github.com/leg100/otf/commit/a39cc65f56faeee9dbb1379fc37bb272244d1195))
* **ui:** appropriate colours for run statuses ([#732](https://github.com/leg100/otf/issues/732)) ([d217736](https://github.com/leg100/otf/commit/d2177361cdb9fbcc157c43a4ffd5fda4befa52ad)), closes [#730](https://github.com/leg100/otf/issues/730)
* **ui:** remove spaces from change summary ([45f2329](https://github.com/leg100/otf/commit/45f232931a98eb349bf313eeea888d67fbf689ef))


### Miscellaneous

* bump js libs ([d90c1a4](https://github.com/leg100/otf/commit/d90c1a4ae01f26ded080ac83d14d157c5dd4519f))
* fix(lint): resolve issues raised with latest staticcheck ([39565b9](https://github.com/leg100/otf/commit/39565b9406d83760ab1f3282f76d4537b35bb98f))
* **ui:** nicer icons for run source ([21c3c05](https://github.com/leg100/otf/commit/21c3c0552caa15d1b886ae83de368e55b130e66f))
* use builtin run watcher for integration tests ([e0299d3](https://github.com/leg100/otf/commit/e0299d3b791af3b072ae7cb69ffefe22840c7042))

## [0.3.13](https://github.com/leg100/otf/compare/v0.3.12...v0.3.13) (2025-01-23)


### Bug Fixes

* **runner:** canceling non-running jobs causes deadlock ([bad9fce](https://github.com/leg100/otf/commit/bad9fce0f310f587492c5e2e0d24d6a51aafcbc3))

## [0.3.12](https://github.com/leg100/otf/compare/v0.3.11...v0.3.12) (2025-01-21)


### Bug Fixes

* allocator not accounting for deleted jobs ([44ea11e](https://github.com/leg100/otf/commit/44ea11e504f06ea549bed7dd08c508b6572f7810))
* runner job capacity exceeded again ([f362c6d](https://github.com/leg100/otf/commit/f362c6da10f5914289043125be8b0dac203c3680))

## [0.3.11](https://github.com/leg100/otf/compare/v0.3.10...v0.3.11) (2025-01-21)


### Bug Fixes

* **allocator:** job capacity exceeded on runners ([#723](https://github.com/leg100/otf/issues/723)) ([53cae20](https://github.com/leg100/otf/commit/53cae20e1cb768cb2fc39c8bf476e84d1f2c0ee8))
* don't add plan-only runs to workspace queue ([fc476c5](https://github.com/leg100/otf/commit/fc476c5e5b5ae7b9d939ad81b75180ecc2c3746d))
* dont schedule already-scheduled runs at startup ([91cf7a1](https://github.com/leg100/otf/commit/91cf7a13a55a7bf19529072574c2c89b44e8525b))

## [0.3.10](https://github.com/leg100/otf/compare/v0.3.9...v0.3.10) (2025-01-19)


### Bug Fixes

* log error if job files cannot be deleted ([9c3afcc](https://github.com/leg100/otf/commit/9c3afcca92dbaf68f1cfb4e34066ec75dd67e884))
* **scheduler:** runs stuck in pending state ([#722](https://github.com/leg100/otf/issues/722)) ([3d4306d](https://github.com/leg100/otf/commit/3d4306d12d84f6235e4a5204bedf680c46756939))


### Miscellaneous

* remove unused parameter ([8df5334](https://github.com/leg100/otf/commit/8df53340e0489127cab95307034e5f75b292671e))

## [0.3.9](https://github.com/leg100/otf/compare/v0.3.8...v0.3.9) (2025-01-16)


### Bug Fixes

* **db:** deadlock exhausting max connections ([1d9d18e](https://github.com/leg100/otf/commit/1d9d18ed4f3c7e2198e94fef6ae6eb1db79f3646))
* disable terraform interactive prompts ([272cabb](https://github.com/leg100/otf/commit/272cabb3a0f350e3cb3325944bc64aba08885f4c))
* use generic update func to eliminate tx propagation failure ([#720](https://github.com/leg100/otf/issues/720)) ([900fbe7](https://github.com/leg100/otf/commit/900fbe7d17aef0a7df27c6cc25cfa72a13cd648e))

## [0.3.8](https://github.com/leg100/otf/compare/v0.3.7...v0.3.8) (2025-01-12)


### Features

* add option to skip plans for pull requests ([#714](https://github.com/leg100/otf/issues/714)) ([29455ed](https://github.com/leg100/otf/commit/29455ed335cf556702bca750bcc18e5b53d82a3e))


### Bug Fixes

* download terraform first to avoid test flakiness ([ea4f03b](https://github.com/leg100/otf/commit/ea4f03b0ecb65a35673f956e9919d001ff3e496c))
* insufficient github perms triggering nil pointer err ([da51e5d](https://github.com/leg100/otf/commit/da51e5d6658697d8f1f2b8eae26cb5190bb62b16))


### Miscellaneous

* use new surl v2 lib ([0c5b45c](https://github.com/leg100/otf/commit/0c5b45cb32a7e7743b11631b4a14f65494141039))

## [0.3.7](https://github.com/leg100/otf/compare/v0.3.6...v0.3.7) (2025-01-03)


### Bug Fixes

* bump packages to fix critical vulns ([b53a404](https://github.com/leg100/otf/commit/b53a404f91de2741681cc60e09324338cc4703a8))
* don't assume public schema for postgres migrations ([e7510b0](https://github.com/leg100/otf/commit/e7510b0fabb487f76a6f5122bb6e2708fadb3a75))
* **ui:** bad db query broke runner listing ([d86705c](https://github.com/leg100/otf/commit/d86705c11a6182462b6a3f2303a6bcfe73c752c4))


### Miscellaneous

* bump playwright ([af87baa](https://github.com/leg100/otf/commit/af87baa8054b470ae9c0fcc00fb3d604440c24a6))
* trigger release ([e90e046](https://github.com/leg100/otf/commit/e90e0469c0a46a16112891b19a9e8877c790283b))

## [0.3.6](https://github.com/leg100/otf/compare/v0.3.5...v0.3.6) (2024-11-22)


### Features

* log the reason why agent is retrying api request ([1d41f61](https://github.com/leg100/otf/commit/1d41f61a1257da0bf473af9efc969f0f21ca78f8))


### Bug Fixes

* authorize create-org action on site, not org ([98bd1c6](https://github.com/leg100/otf/commit/98bd1c680e7fef486e622c9d1835d060df444265))
* only log subsystem start when actually started ([a72fbd0](https://github.com/leg100/otf/commit/a72fbd0cc556f6f4bfbb85f79bef255512b40111))
* runner unable to re-register ([#707](https://github.com/leg100/otf/issues/707)) ([41f5669](https://github.com/leg100/otf/commit/41f5669ef362dc42507b8a8c61c074e022c9bb02))


### Miscellaneous

* update screenshots ([b17ffa4](https://github.com/leg100/otf/commit/b17ffa45394cfa23343637529df4ff46bcac97ab))

## [0.3.5](https://github.com/leg100/otf/compare/v0.3.4...v0.3.5) (2024-10-30)


### Bug Fixes

* **ui:** unresponsive agent pool page ([#701](https://github.com/leg100/otf/issues/701)) ([c0853f5](https://github.com/leg100/otf/commit/c0853f51f8ba359e6f96fdd43ccb089ede645598)), closes [#698](https://github.com/leg100/otf/issues/698) [#661](https://github.com/leg100/otf/issues/661)

## [0.3.4](https://github.com/leg100/otf/compare/v0.3.3...v0.3.4) (2024-10-29)


### Features

* run status metrics ([#697](https://github.com/leg100/otf/issues/697)) ([4642e5d](https://github.com/leg100/otf/commit/4642e5d98937ca036f68681ab5abef9ed14379d2))


### Bug Fixes

* log create agent pool db errors ([ece6de7](https://github.com/leg100/otf/commit/ece6de7186d54bbeb390f8d6e7bf5c9e1c0eb543))

## [0.3.3](https://github.com/leg100/otf/compare/v0.3.2...v0.3.3) (2024-10-23)


### Bug Fixes

* make tern migration postgres 12 compat ([#695](https://github.com/leg100/otf/issues/695)) ([c662668](https://github.com/leg100/otf/commit/c6626686d03b8f6e2a8a34d8db1e692bf99f2b05))
* report 409 when cancel or force cancel not allowed ([#693](https://github.com/leg100/otf/issues/693)) ([dbe5668](https://github.com/leg100/otf/commit/dbe5668cb8fab94eb510d64f2962f523efef8b46))

## [0.3.2](https://github.com/leg100/otf/compare/v0.3.1...v0.3.2) (2024-10-22)


### Bug Fixes

* bad git command in chart release step ([3ed91e3](https://github.com/leg100/otf/commit/3ed91e351a23c0d5ba858ff0423035194bc0471b))

## [0.3.1](https://github.com/leg100/otf/compare/v0.3.0...v0.3.1) (2024-10-22)


### Bug Fixes

* make goreleaser config valid for v2 ([0ed8d22](https://github.com/leg100/otf/commit/0ed8d22140570aefefdcabd3a16cdd46eefcebf4))

## [0.3.0](https://github.com/leg100/otf/compare/v0.2.4...v0.3.0) (2024-10-22)


### ⚠ BREAKING CHANGES

* rename --address flag to --url; require scheme
* move to sqlc, tern ([#683](https://github.com/leg100/otf/issues/683))

### refactor

* move to sqlc, tern ([#683](https://github.com/leg100/otf/issues/683)) ([878ebfb](https://github.com/leg100/otf/commit/878ebfba5eea34926b2f3aebe9bce8f826347b20))
* rename --address flag to --url; require scheme ([3e83474](https://github.com/leg100/otf/commit/3e834744e2eff9984dd29255b53f489aa857e2dc))


### Features

* add timeout settings for plans and applies ([#686](https://github.com/leg100/otf/issues/686)) ([797902b](https://github.com/leg100/otf/commit/797902bc8187a8b1682d340abab730a58247c304))
* allow subscription buffer size to be overridden ([#687](https://github.com/leg100/otf/issues/687)) ([d51469d](https://github.com/leg100/otf/commit/d51469d84534e22bd6b4ef3ee55a76582d58aec4))


### Bug Fixes

* avoid hitting Github limit on commit status updates ([#688](https://github.com/leg100/otf/issues/688)) ([029e525](https://github.com/leg100/otf/commit/029e5255a5ea70ea7701f11b16add510fb03b23f))
* don't unnecessarily restart scheduler ([#689](https://github.com/leg100/otf/issues/689)) ([d240965](https://github.com/leg100/otf/commit/d24096536d06da30b035d846c7eb62537e2153a0))
* make linting and tests pass ([ebc1e53](https://github.com/leg100/otf/commit/ebc1e53a9e00412e4db8467a71b299e86b5fb43e))
* pin version of gcp pub-sub emulator docker image ([8048e72](https://github.com/leg100/otf/commit/8048e720d3d7a99c1f244ba797f6276c1dd56938))
* prevent subsystem failure from stopping otfd ([e5061b0](https://github.com/leg100/otf/commit/e5061b04bf051c9ad7f9e4ba7408cc3c4afa1a43))
* use base58 alphabet for resource IDs ([#680](https://github.com/leg100/otf/issues/680)) ([1e7d7a2](https://github.com/leg100/otf/commit/1e7d7a2b2c350c17c29fb49ae0dfbbeb31b2942d))


### Miscellaneous

* bump go ([5663eab](https://github.com/leg100/otf/commit/5663eab55cf22f5d9fe9f8a29c6df9d2908ecb20))
* trigger new version of agent chart upon deploy ([#690](https://github.com/leg100/otf/issues/690)) ([155e026](https://github.com/leg100/otf/commit/155e026252c52d4ec7565f9bb7444b297bca16f7))
* unarchive ([c954c36](https://github.com/leg100/otf/commit/c954c36107c910192941393e366ee1b8a2e290f1))
* upgrade dependencies ([59eb979](https://github.com/leg100/otf/commit/59eb97996785328e78c31f2372cfce889f5ef7cb))
* use forked sse lib's module path ([fc9b138](https://github.com/leg100/otf/commit/fc9b13865844062570ffbeb7d8c57bdf5ae50b91))

## [0.2.4](https://github.com/leg100/otf/compare/v0.2.3...v0.2.4) (2023-12-16)


### Features

* Add Webhook Hostname ([#668](https://github.com/leg100/otf/issues/668)) ([#670](https://github.com/leg100/otf/issues/670)) ([f2dc7e9](https://github.com/leg100/otf/commit/f2dc7e9425dca693cf9adff11aada0217d5cce7e))

## [0.2.3](https://github.com/leg100/otf/compare/v0.2.2...v0.2.3) (2023-12-12)


### Bug Fixes

* gitlab support ([#665](https://github.com/leg100/otf/issues/665)) ([eaf9b15](https://github.com/leg100/otf/commit/eaf9b15556159079d2770064f8d374f627615ea7)), closes [#651](https://github.com/leg100/otf/issues/651)

## [0.2.2](https://github.com/leg100/otf/compare/v0.2.1...v0.2.2) (2023-12-10)


### Bug Fixes

* allocator restarting unnecessarily ([#666](https://github.com/leg100/otf/issues/666)) ([47f8e6f](https://github.com/leg100/otf/commit/47f8e6f74cd7fb36bf2b5eb3885bbd995bcf81c0))
* log max config size exceeded ([#663](https://github.com/leg100/otf/issues/663)) ([e196837](https://github.com/leg100/otf/commit/e196837fe88fc41b3b908537766db0b66530d281)), closes [#652](https://github.com/leg100/otf/issues/652)

## [0.2.1](https://github.com/leg100/otf/compare/v0.2.0...v0.2.1) (2023-12-07)


### Bug Fixes

* add extra case for gitlab repo dir name ([#654](https://github.com/leg100/otf/issues/654)) ([5424565](https://github.com/leg100/otf/commit/542456530d8551c34bd2a2d298f931dee5c52827))
* organization tokens ([#660](https://github.com/leg100/otf/issues/660)) ([be82c55](https://github.com/leg100/otf/commit/be82c559399a0b023aa63fe8f36e61d6fb9a9848))
* various agent pool and job bugs ([#659](https://github.com/leg100/otf/issues/659)) ([ed9b1fd](https://github.com/leg100/otf/commit/ed9b1fdb6485f8ad16f60df19021273376a3bdd4))

## [0.2.0](https://github.com/leg100/otf/compare/v0.1.18...v0.2.0) (2023-12-05)


### ⚠ BREAKING CHANGES

* agent pools ([#653](https://github.com/leg100/otf/issues/653))

### Features

* agent pools ([#653](https://github.com/leg100/otf/issues/653)) ([662bfd9](https://github.com/leg100/otf/commit/662bfd9bbd5aff7a6bc9e94253a5ac525aedc113))


### Bug Fixes

* Add missing CancelRunAction to WorkspaceWriteRole ([#649](https://github.com/leg100/otf/issues/649)) ([599ddcb](https://github.com/leg100/otf/commit/599ddcb5494de845ce6fe8e91240facf3b8fb466))
* docker-compose otfd healthcheck ([c553b58](https://github.com/leg100/otf/commit/c553b5895ff9bc8993991c872a31d74a63bc92f2))

## [0.1.18](https://github.com/leg100/otf/compare/v0.1.17...v0.1.18) (2023-10-30)


### Bug Fixes

* **ci:** charts job needs release info ([f4fef03](https://github.com/leg100/otf/commit/f4fef03c3a594bdb21c63dcbe2d0c9aeef6c242d))
* **ui:** push docs to remote gh-pages branch ([5b3e3f4](https://github.com/leg100/otf/commit/5b3e3f4e3034aeaba6cb27855f2314c12b964112))
* **ui:** workspace listing returning 500 error ([6eb89f4](https://github.com/leg100/otf/commit/6eb89f48c73337426feb1d704a60f4471cf85942))

## [0.1.17](https://github.com/leg100/otf/compare/v0.1.16...v0.1.17) (2023-10-29)


### Features

* **ui:** show running times ([#635](https://github.com/leg100/otf/issues/635)) ([7337c2e](https://github.com/leg100/otf/commit/7337c2ecde3876c51ab77ae477f7664c264f42a3)), closes [#604](https://github.com/leg100/otf/issues/604)


### Bug Fixes

* mike doc versioner flags have changed ([224081c](https://github.com/leg100/otf/commit/224081c5bbff6fd8ea0150365886573b457e25b3))
* publish chart after release not before ([eceab7e](https://github.com/leg100/otf/commit/eceab7efbec1b82070cc02731f734e579d9cdd80))
* **ui:** allow variable to be updated from hcl to non-hcl ([ac0ff5a](https://github.com/leg100/otf/commit/ac0ff5ae654c18525632d41085fce34c8cc36711))


### Miscellaneous

* document some more flags ([e2cc4f2](https://github.com/leg100/otf/commit/e2cc4f271956e737571778de81f8e4d926fe3e55))
* **perf:** pre-allocate slices ([ccc8b6e](https://github.com/leg100/otf/commit/ccc8b6e0a6c3195ef239323574ce3b51aa86bce9))
* remove redundant jsonapiclient interface ([5aa153a](https://github.com/leg100/otf/commit/5aa153a9822d86714845509c5f15c962321382cd))

## [0.1.16](https://github.com/leg100/otf/compare/v0.1.15...v0.1.16) (2023-10-27)


### Bug Fixes

* allow org members to view variable sets ([df9fa53](https://github.com/leg100/otf/commit/df9fa53fad1e51c2c0b9e4d9ac4f493c5be66fb7))
* broken mike python package for docs ([34c50e2](https://github.com/leg100/otf/commit/34c50e2f08b5a460b15ee38d9b319187d34a8516))

## [0.1.15](https://github.com/leg100/otf/compare/v0.1.14...v0.1.15) (2023-10-27)

### Features

* Implement TFE API for Team Tokens ([#624](https://github.com/leg100/otf/issues/624))

### Bug Fixes

* Fix local execution mode ([#627](https://github.com/leg100/otf/issues/627)
* agent error reporting ([#628](https://github.com/leg100/otf/issues/628)) ([76e7dda](https://github.com/leg100/otf/commit/76e7dda7a6d6ca29c8ee1cd8feecb3def0f77c44))
* fixed defect with multiline tfvars not being escaped ([#631](https://github.com/leg100/otf/issues/631)) ([f35dffa](https://github.com/leg100/otf/commit/f35dffa97bec141491c1121fd10f39f5ca7893a1))

## [0.1.14](https://github.com/leg100/otf/compare/v0.1.13...v0.1.14) (2023-10-19)


### Features

* github app: [#617](https://github.com/leg100/otf/issues/617)
* always use latest terraform version ([#616](https://github.com/leg100/otf/issues/616)) ([83469ca](https://github.com/leg100/otf/commit/83469ca998b8756673cc9ff06c8225bd3cc62e61)), closes [#608](https://github.com/leg100/otf/issues/608)


### Bug Fixes

* error 'schema: converter not found for integration.manifest' ([e53ebf2](https://github.com/leg100/otf/commit/e53ebf2e34288e437b11d69eba3e61324b21be22))
* fixed bug where proxy was ignored ([#609](https://github.com/leg100/otf/issues/609)) ([c1ee8d8](https://github.com/leg100/otf/commit/c1ee8d8ea53a05935c7d5d510054a6eaf588aa25))
* prevent modules with no published versions from crashing otf ([#611](https://github.com/leg100/otf/issues/611)) ([84aa299](https://github.com/leg100/otf/commit/84aa2992856b87ad17b6dd582ee4528c01873b69))
* skip reporting runs created via API ([#622](https://github.com/leg100/otf/issues/622)) ([5d4527b](https://github.com/leg100/otf/commit/5d4527b52573c8600d49ed149ea16bdb7f57f141)), closes [#618](https://github.com/leg100/otf/issues/618)


### Miscellaneous

* add note re cloud block to allow CLI apply ([4f03544](https://github.com/leg100/otf/commit/4f03544275ac884073be221f5f8a5f88ada0552d))
* remove unused exchange code response ([4a966cd](https://github.com/leg100/otf/commit/4a966cd8cbfc1c4232c1ebe7b83c62044a2a8af2))
* upgrade vulnerable markdown go mod ([781e0f6](https://github.com/leg100/otf/commit/781e0f6e047abe662336250e679797f1b3ed0752))

## [0.1.13](https://github.com/leg100/otf/compare/v0.1.12...v0.1.13) (2023-09-13)


### Features

* add flags --oidc-username-claim and --oidc-scopes ([#605](https://github.com/leg100/otf/issues/605)) ([87324d0](https://github.com/leg100/otf/commit/87324d00afbf7944516ed091f6014f4b3001c177)), closes [#596](https://github.com/leg100/otf/issues/596)


### Bug Fixes

* restart spooler when broker terminates subscription ([#600](https://github.com/leg100/otf/issues/600)) ([ce41580](https://github.com/leg100/otf/commit/ce41580f1640c282ae89437eb377a8554232c412))
* retrieving state outputs only requires read role ([#603](https://github.com/leg100/otf/issues/603)) ([25c4a99](https://github.com/leg100/otf/commit/25c4a992fac150aca02a51c5d655d6364d159dca))

## [0.1.12](https://github.com/leg100/otf/compare/v0.1.11...v0.1.12) (2023-09-12)


### Features

* **ui:** clickable widgets ([#597](https://github.com/leg100/otf/issues/597)) ([518452e](https://github.com/leg100/otf/commit/518452ede3d458e1bd0105f2a0a46ab5b5cb36c6))


### Bug Fixes

* tfe_outputs resource ([#599](https://github.com/leg100/otf/issues/599)) ([89de01d](https://github.com/leg100/otf/commit/89de01d48c1878982a7f56e436c8904bd3bc0a09)), closes [#595](https://github.com/leg100/otf/issues/595)


### Miscellaneous

* remove unnecessary link from widget heading ([318c390](https://github.com/leg100/otf/commit/318c39052ebcbbee187dbc2a08a0a456dab70352))

## [0.1.11](https://github.com/leg100/otf/compare/v0.1.10...v0.1.11) (2023-09-11)


### Features

* update vcs provider token ([#594](https://github.com/leg100/otf/issues/594)) ([29a0be6](https://github.com/leg100/otf/commit/29a0be667046440aab25efc25c9a7a02720d2f96)), closes [#576](https://github.com/leg100/otf/issues/576)


### Bug Fixes

* dont scrub sensitive variable values for agent ([#591](https://github.com/leg100/otf/issues/591)) ([a333ee6](https://github.com/leg100/otf/commit/a333ee6f7a04c234dbe5c34602a35f1095f35b32)), closes [#590](https://github.com/leg100/otf/issues/590)
* **integration:** prevent -32000 error ([39318f1](https://github.com/leg100/otf/commit/39318f1dd1966f621bfb930bf2f8cbee2c70266d))
* **integration:** wait for alpinejs to load ([346024e](https://github.com/leg100/otf/commit/346024ea87eedabfd086ea536c5ee79d19b531fa))
* resubscribe subsystems when their subscription is terminated ([#593](https://github.com/leg100/otf/issues/593)) ([3195e17](https://github.com/leg100/otf/commit/3195e17fe3e98ec418e0bbef6e4e46bc707a4f6c))

## [0.1.10](https://github.com/leg100/otf/compare/v0.1.9...v0.1.10) (2023-09-06)


### Bug Fixes

* **integration:** ensure text box is visible before focusing ([8d279ae](https://github.com/leg100/otf/commit/8d279aefdc8830b32cb262e8608ff394a2f62880))
* set module status ([#586](https://github.com/leg100/otf/issues/586)) ([8141c6e](https://github.com/leg100/otf/commit/8141c6ed2da175700405cb5c5f34658660cb68e7))
* **ui:** remove undefined css classes ([daf6096](https://github.com/leg100/otf/commit/daf60965418061ff4374689613bc8c2a2ce8efe8))
* **ui:** wrong heading for edit variable set variable page ([cc6f282](https://github.com/leg100/otf/commit/cc6f2827708beefe69d8e6c88d85e83502493a51))
* variable set variables API ([#589](https://github.com/leg100/otf/issues/589)) ([8e29da1](https://github.com/leg100/otf/commit/8e29da191122103dd76eca876c37b419e106e630)), closes [#588](https://github.com/leg100/otf/issues/588)

## [0.1.9](https://github.com/leg100/otf/compare/v0.1.8...v0.1.9) (2023-09-02)


### Features

* variable sets ([#574](https://github.com/leg100/otf/issues/574)) ([419e2fb](https://github.com/leg100/otf/commit/419e2fb81cdb8a3b6b9cc7d91e81ca7af29d3a24)), closes [#306](https://github.com/leg100/otf/issues/306)


### Bug Fixes

* **integration:** stop browser test failing with -32000 error ([27f02cd](https://github.com/leg100/otf/commit/27f02cd9f22f2f94d4427964f64417c0fdec83a0))
* **scheduler:** ignore deleted run events ([60496bb](https://github.com/leg100/otf/commit/60496bb4849d64393c572a89f4969b45257c6b60))
* **ui:** deleting vcs provider no longer breaks module page ([e28b931](https://github.com/leg100/otf/commit/e28b931703848430d0943cf2606d701511b2f003))
* **ui:** make workspace page title use Name, not ID ([#581](https://github.com/leg100/otf/issues/581)) ([8268643](https://github.com/leg100/otf/commit/8268643a6775e2eb492d14c2ddf374c813b86c63))


### Miscellaneous

* add BSL compliance note ([6b537de](https://github.com/leg100/otf/commit/6b537de846d6410d7e765f1c9f73945d0e679090))
* document integration test verbose logging ([75272a4](https://github.com/leg100/otf/commit/75272a4b7842426e2901615f5898d02a515a310b))

## [0.1.8](https://github.com/leg100/otf/compare/v0.1.7...v0.1.8) (2023-08-13)


### Features

* allow terraform apply on connected workspace ([#564](https://github.com/leg100/otf/issues/564)) ([6f90a9c](https://github.com/leg100/otf/commit/6f90a9c0b817f6cb846df1a487606a52a963a7b4)), closes [#231](https://github.com/leg100/otf/issues/231)
* **ui:** add icon in run widget to show source of run ([#563](https://github.com/leg100/otf/issues/563)) ([2e7a0bd](https://github.com/leg100/otf/commit/2e7a0bd71b99556360070337b8e9baad3a021aad))


### Bug Fixes

* return error on stream error for retry (https://github.com/leg100/otf/pull/565)
* cleanup after extracting repo tarball ([bf4758b](https://github.com/leg100/otf/commit/bf4758bead52e6c3bf1e47d1dfe06ebcff0a26a8))
* don't scrub included state output sensitive values ([478e314](https://github.com/leg100/otf/commit/478e314a687f722125653d6aa1010b8c3bf2b060))
* linux/arm64 support ([#562](https://github.com/leg100/otf/issues/562)) ([01a2112](https://github.com/leg100/otf/commit/01a211240e4dca4d18e02d49e3f9d6190754510c)), closes [#311](https://github.com/leg100/otf/issues/311)
* otfd compose healthcheck: curl not installed ([9f52021](https://github.com/leg100/otf/commit/9f52021d7515b4736547d8e978dcabd756d5c263))
* retry run should use existing run properties ([49303ec](https://github.com/leg100/otf/commit/49303ecf42edb106def169ddf68f66df7558b741))
* **tests:** hard link fails when /tmp is separate partition ([cfc7aaa](https://github.com/leg100/otf/commit/cfc7aaa80d31c9b7a0b4e461d13bb09fea9f87bc))
* **ui:** workspace description missing after update ([a579b40](https://github.com/leg100/otf/commit/a579b40ffac4aa2f021afc9417ad9bb8b3b2cc49))
* use png instead of svg for font-based icons ([eae0588](https://github.com/leg100/otf/commit/eae0588d7b6de5fb4cb2e5c2ad7fa483360f308c))


### Miscellaneous

* bump squid version ([7ce3238](https://github.com/leg100/otf/commit/7ce3238f7af3755c317a28690b1dbd8e7efed2b9))
* go 1.21 ([#566](https://github.com/leg100/otf/issues/566)) ([06c13b2](https://github.com/leg100/otf/commit/06c13b250b183c12e0486e69cac2aee1c52b7ed5))
* remove unused cloud team and org sync code ([4e1817d](https://github.com/leg100/otf/commit/4e1817dbbd21093c835e84d921606dd2ae46f871))
* removed unused ca.pem ([799ed25](https://github.com/leg100/otf/commit/799ed25565c155c616e90533b3172bc22f916f6b))
* skip api tests if env vars not present ([5b88474](https://github.com/leg100/otf/commit/5b88474d3c4813897f39f3b463d013cbc831ad64))
* **ui:** make tags less bulbous ([df1645d](https://github.com/leg100/otf/commit/df1645d8de9d4ce021d93e58f03d27911494649f))
* **ui:** pad out buttons on consent page ([1c290e9](https://github.com/leg100/otf/commit/1c290e93248d9620d54c41eb3681065929069cde))
* update docs ([364d183](https://github.com/leg100/otf/commit/364d183dd8635eb0ce73b1e65666475ab0a039ea))
* validate resource names ([c7596fe](https://github.com/leg100/otf/commit/c7596febc1018a546ec2c990ae5087ae297df8c0))

## [0.1.7](https://github.com/leg100/otf/compare/v0.1.6...v0.1.7) (2023-08-05)


### Bug Fixes

* remove unused `groups` OIDC scope ([#558](https://github.com/leg100/otf/issues/558)) ([3dd465a](https://github.com/leg100/otf/commit/3dd465a6992cce43996e712a13af6e84782558e7)), closes [#557](https://github.com/leg100/otf/issues/557)


### Miscellaneous

* chromium bug fixed ([#559](https://github.com/leg100/otf/issues/559)) ([87af2c7](https://github.com/leg100/otf/commit/87af2c74e235c14241987bbcf4f67da70ccd7b4e))

## [0.1.6](https://github.com/leg100/otf/compare/v0.1.5...v0.1.6) (2023-08-02)


### Features

* record who created a run ([#556](https://github.com/leg100/otf/issues/556)) ([57bb9b6](https://github.com/leg100/otf/commit/57bb9b6fad3445cdf830ae782ca3b07b6b024179))
* **ui:** clicking on workspace widget tag filters by that tag ([a7ce9a8](https://github.com/leg100/otf/commit/a7ce9a890dfed4976c42619de09285cf6dd2b70d))
* **ui:** provide more vcs metadata for runs ([#552](https://github.com/leg100/otf/issues/552)) ([18217ce](https://github.com/leg100/otf/commit/18217ce43b357d4107e12b5bd52984346da800a4))


### Miscellaneous

* add organization UI tests ([1c7e3db](https://github.com/leg100/otf/commit/1c7e3dbaba958710d2c07aab7ac6781b950d3b37))
* remove redundant CreateRun magic string ([#555](https://github.com/leg100/otf/issues/555)) ([a2df6d5](https://github.com/leg100/otf/commit/a2df6d5247d1e605fe852eb2ebe4cf7e2b35f795))

## [0.1.5](https://github.com/leg100/otf/compare/v0.1.4...v0.1.5) (2023-08-01)


### Features

* add support for terraform_remote_state ([#550](https://github.com/leg100/otf/issues/550)) ([c2fa0a7](https://github.com/leg100/otf/commit/c2fa0a7b5b6d8d18f842dfa760e4f6d7cd97bc07))

## [0.1.4](https://github.com/leg100/otf/compare/v0.1.3...v0.1.4) (2023-08-01)


### Features

* more workspace VCS settings ([#545](https://github.com/leg100/otf/issues/545)) ([abfc702](https://github.com/leg100/otf/commit/abfc702e8bce25842da08a655e38fee8a4ccc75a))
* **ui:** hide functionality from unpriv persons ([#548](https://github.com/leg100/otf/issues/548)) ([fee491f](https://github.com/leg100/otf/commit/fee491fa0d3c6fee5ce62ecf4c2c3dfd154011ba)), closes [#540](https://github.com/leg100/otf/issues/540)


### Miscellaneous

* downplay legitimate state not found errors ([2d91e31](https://github.com/leg100/otf/commit/2d91e313862d6e412369853f10fb48fb87068337))
* remove demo ([d70c7fd](https://github.com/leg100/otf/commit/d70c7fdfd82ce39ff0e1a1d05b4ee38ba04e0b5b))
* **ui:** make workspace state tabs look nicer ([bbe38b4](https://github.com/leg100/otf/commit/bbe38b4e0ee6808523ac52687b8544e308233a7a))

## [0.1.3](https://github.com/leg100/otf/compare/v0.1.2...v0.1.3) (2023-07-27)


### Features

* **ui:** add tags to workspace widget ([#543](https://github.com/leg100/otf/issues/543)) ([3000c09](https://github.com/leg100/otf/commit/3000c097d50d47f4fdd6c987e1e41a609fa92f16))
* **ui:** show resources and outputs on workspace page ([#542](https://github.com/leg100/otf/issues/542)) ([d792e23](https://github.com/leg100/otf/commit/d792e239733c57d7821957ece8c2704f7e080347)), closes [#308](https://github.com/leg100/otf/issues/308)


### Bug Fixes

* **ui:** style variables table ([ed67d57](https://github.com/leg100/otf/commit/ed67d57d2298e017e1199180341e66ca46fed4be))

## [0.1.2](https://github.com/leg100/otf/compare/v0.1.1...v0.1.2) (2023-07-26)


### Bug Fixes

* agent race error ([#537](https://github.com/leg100/otf/issues/537)) ([6b9e6b1](https://github.com/leg100/otf/commit/6b9e6b1949a0121d5b04558334ce4011fa88a3be))
* handle run-events request from terraform cloud backend ([#534](https://github.com/leg100/otf/issues/534)) ([b1998bd](https://github.com/leg100/otf/commit/b1998bd00450f296a5186c1d0464e93247655e86))
* terraform apply partial state updates ([#539](https://github.com/leg100/otf/issues/539)) ([d25e7e4](https://github.com/leg100/otf/commit/d25e7e4678ca55d49a6dfdf041de077187d5a54a)), closes [#527](https://github.com/leg100/otf/issues/527)


### Miscellaneous

* removed unused config file ([84fe3b1](https://github.com/leg100/otf/commit/84fe3b1a6caf4db7611d912b3316747705209e39))

## [0.1.1](https://github.com/leg100/otf/compare/v0.1.0...v0.1.1) (2023-07-24)


### Bug Fixes

* **ui:** improve dropdown box UX ([d67de76](https://github.com/leg100/otf/commit/d67de7696d21d0bd6c2ef93d9b06ebcfc8190ff7))
* **ui:** new team form missing borders ([0506694](https://github.com/leg100/otf/commit/0506694f862475200f23b137ba39b6af2b755fa0))

## [0.1.0](https://github.com/leg100/otf/compare/v0.0.53...v0.1.0) (2023-07-24)


### ⚠ BREAKING CHANGES

* adding team member creates user if they don't exist ([#525](https://github.com/leg100/otf/issues/525))

### Features

* adding team member creates user if they don't exist ([#525](https://github.com/leg100/otf/issues/525)) ([fbeb789](https://github.com/leg100/otf/commit/fbeb789bc4b5616f7dc395311837423a42535d69))
* organization tokens ([#528](https://github.com/leg100/otf/issues/528)) ([7ddd416](https://github.com/leg100/otf/commit/7ddd416937f6421adfafa59b0ddd60d5f35a05e6))
* **ui:** tag search/dropdown menu ([#523](https://github.com/leg100/otf/issues/523)) ([09b8310](https://github.com/leg100/otf/commit/09b83105e10f882283419b1645d49e2c04929774))


### Bug Fixes

* embed magnifying glass icon ([8a45d51](https://github.com/leg100/otf/commit/8a45d513a436bf42072460d5351bcc2380e5e961))
* run tailwind css on template changes ([e749013](https://github.com/leg100/otf/commit/e7490133ed74bf1278f2b519ab58ebd8a7dd4820))

## [0.0.53](https://github.com/leg100/otf/compare/v0.0.52...v0.0.53) (2023-07-12)


### Bug Fixes

* delete existing unreferenced webhooks too ([6b61b48](https://github.com/leg100/otf/commit/6b61b485198be0b2074bd53c1633649831855588))
* delete webhooks when org or vcs provider is deleted ([#518](https://github.com/leg100/otf/issues/518)) ([0d36ea5](https://github.com/leg100/otf/commit/0d36ea554f1c3a521069426c4643b7c63a73be36))
* **docs:** version using tag not branch name ([8613fe8](https://github.com/leg100/otf/commit/8613fe88ce9d0d8fab939d5784d9bd114bdbf6b1))
* only set not null after populating column ([1da3936](https://github.com/leg100/otf/commit/1da3936e12532170bb6c82d3c96607f53ab50ff4))
* remove trailing slash from requests ([#516](https://github.com/leg100/otf/issues/516)) ([c1ee39e](https://github.com/leg100/otf/commit/c1ee39e73bfe03e2de2b3dcc9a745ea5c99985f5)), closes [#496](https://github.com/leg100/otf/issues/496)
* **ui:** add cache-control header to static files ([061261f](https://github.com/leg100/otf/commit/061261f032aed1d18054ef03c960762695e64aef))


### Miscellaneous

* add hashes to all static urls ([3650926](https://github.com/leg100/otf/commit/36509261c1f9e4c7e574fd22a9d79e6c0b0ee26d))
* test create connected workspace via api ([9bf4bae](https://github.com/leg100/otf/commit/9bf4bae2d7d26c52a169302dca2f7c2ef11c1cde))

## [0.0.52](https://github.com/leg100/otf/compare/v0.0.51...v0.0.52) (2023-07-08)


### Bug Fixes

* helm chart branch name ([b77dc8a](https://github.com/leg100/otf/commit/b77dc8abaa4ff7bc3be0f71f84e14ab7b00dc010))

## [0.0.51](https://github.com/leg100/otf/compare/v0.0.50...v0.0.51) (2023-07-08)


### Bug Fixes

* apply on output changes ([#501](https://github.com/leg100/otf/issues/501)) ([46cd3ef](https://github.com/leg100/otf/commit/46cd3efbffc899d180363e767d7730ee4b473b6a))
* delete unreferenced tags ([#507](https://github.com/leg100/otf/issues/507)) ([d85ac43](https://github.com/leg100/otf/commit/d85ac430faffc2afa1367a96e623001b38a98690)), closes [#502](https://github.com/leg100/otf/issues/502)
* finish events refactor ([#509](https://github.com/leg100/otf/issues/509)) ([096933a](https://github.com/leg100/otf/commit/096933a5affb2e0a33d61dd4503a7793465ea1ac))
* flaky browser tests ([#484](https://github.com/leg100/otf/issues/484)) ([1ce0bd0](https://github.com/leg100/otf/commit/1ce0bd0aa47fde48d9d58f239edb9ee337d1e092))
* prevent empty owners team ([#499](https://github.com/leg100/otf/issues/499)) ([a77c9e9](https://github.com/leg100/otf/commit/a77c9e98aa25f1b3b35041f7680d5298f712f10b))


### Miscellaneous

* Bump default terraform version to v1.5.2 ([#503](https://github.com/leg100/otf/issues/503)) ([67bc3f0](https://github.com/leg100/otf/commit/67bc3f00c2ac9aca11092c5e8c1170f0bccf1216))
