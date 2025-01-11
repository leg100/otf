# Changelog

## [0.4.0](https://github.com/leg100/otf/compare/v0.3.8...v0.4.0) (2025-01-11)


### ⚠ BREAKING CHANGES

* rename --address flag to --url; require scheme
* move to sqlc, tern ([#683](https://github.com/leg100/otf/issues/683))
* agent pools ([#653](https://github.com/leg100/otf/issues/653))
* adding team member creates user if they don't exist ([#525](https://github.com/leg100/otf/issues/525))
* add team membership tfe api endpoints ([#492](https://github.com/leg100/otf/issues/492))
* don't reveal info on healthz endpoint ([#489](https://github.com/leg100/otf/issues/489))
* terraform login protocol ([#424](https://github.com/leg100/otf/issues/424))
* replace zerolog with slog for logging ([#396](https://github.com/leg100/otf/issues/396))

### refactor

* move to sqlc, tern ([#683](https://github.com/leg100/otf/issues/683)) ([878ebfb](https://github.com/leg100/otf/commit/878ebfba5eea34926b2f3aebe9bce8f826347b20))
* rename --address flag to --url; require scheme ([3e83474](https://github.com/leg100/otf/commit/3e834744e2eff9984dd29255b53f489aa857e2dc))


### Features

* add color to logo ([#409](https://github.com/leg100/otf/issues/409)) ([0e1b7e3](https://github.com/leg100/otf/commit/0e1b7e3003cec9646007b1318bcf2cb8f24872f9))
* add flags --oidc-username-claim and --oidc-scopes ([#605](https://github.com/leg100/otf/issues/605)) ([87324d0](https://github.com/leg100/otf/commit/87324d00afbf7944516ed091f6014f4b3001c177)), closes [#596](https://github.com/leg100/otf/issues/596)
* add metadata to pubsub notifications ([#481](https://github.com/leg100/otf/issues/481)) ([19597a8](https://github.com/leg100/otf/commit/19597a8d897f455b18427dcba374349a1c4e202c))
* add more links ([200f8ec](https://github.com/leg100/otf/commit/200f8ec2ee061942036697e98f7f80c4d5669706))
* add option to skip plans for pull requests ([#714](https://github.com/leg100/otf/issues/714)) ([29455ed](https://github.com/leg100/otf/commit/29455ed335cf556702bca750bcc18e5b53d82a3e))
* add plan summary to ui widget ([#408](https://github.com/leg100/otf/issues/408)) ([39d84f8](https://github.com/leg100/otf/commit/39d84f8ee828d937afa7a5aa7948756a02f2bc44))
* add support for terraform_remote_state ([#550](https://github.com/leg100/otf/issues/550)) ([c2fa0a7](https://github.com/leg100/otf/commit/c2fa0a7b5b6d8d18f842dfa760e4f6d7cd97bc07))
* add team membership tfe api endpoints ([#492](https://github.com/leg100/otf/issues/492)) ([059e970](https://github.com/leg100/otf/commit/059e97094eede3f80d8c97f7c48748c68101d073))
* add tfe-api OAuth Client endpoints ([#493](https://github.com/leg100/otf/issues/493)) ([15a70df](https://github.com/leg100/otf/commit/15a70df7a18373d9d9bac19508a22ab55a6c48d8))
* add timeout settings for plans and applies ([#686](https://github.com/leg100/otf/issues/686)) ([797902b](https://github.com/leg100/otf/commit/797902bc8187a8b1682d340abab730a58247c304))
* Add Webhook Hostname ([#668](https://github.com/leg100/otf/issues/668)) ([#670](https://github.com/leg100/otf/issues/670)) ([f2dc7e9](https://github.com/leg100/otf/commit/f2dc7e9425dca693cf9adff11aada0217d5cce7e))
* adding team member creates user if they don't exist ([#525](https://github.com/leg100/otf/issues/525)) ([fbeb789](https://github.com/leg100/otf/commit/fbeb789bc4b5616f7dc395311837423a42535d69))
* agent pools ([#653](https://github.com/leg100/otf/issues/653)) ([662bfd9](https://github.com/leg100/otf/commit/662bfd9bbd5aff7a6bc9e94253a5ac525aedc113))
* allow subscription buffer size to be overridden ([#687](https://github.com/leg100/otf/issues/687)) ([d51469d](https://github.com/leg100/otf/commit/d51469d84534e22bd6b4ef3ee55a76582d58aec4))
* allow terraform apply on connected workspace ([#564](https://github.com/leg100/otf/issues/564)) ([6f90a9c](https://github.com/leg100/otf/commit/6f90a9c0b817f6cb846df1a487606a52a963a7b4)), closes [#231](https://github.com/leg100/otf/issues/231)
* always use latest terraform version ([#616](https://github.com/leg100/otf/issues/616)) ([83469ca](https://github.com/leg100/otf/commit/83469ca998b8756673cc9ff06c8225bd3cc62e61)), closes [#608](https://github.com/leg100/otf/issues/608)
* builtin livereload functionality ([#451](https://github.com/leg100/otf/issues/451)) ([f800fb9](https://github.com/leg100/otf/commit/f800fb9de658bafd26853bc82084b421299cfb33))
* create run using config from repo ([#466](https://github.com/leg100/otf/issues/466)) ([885f786](https://github.com/leg100/otf/commit/885f78647c60825078cba7cd57b635e890b8244e))
* force unlock tfc api endpoint ([#404](https://github.com/leg100/otf/issues/404)) ([fd72212](https://github.com/leg100/otf/commit/fd722122856a366680d738f859051ece70edea76))
* get team by id ([#487](https://github.com/leg100/otf/issues/487)) ([6edd315](https://github.com/leg100/otf/commit/6edd315437c517a8e132df82fd3a6d2aba326c44))
* include outputs via workspace API ([#411](https://github.com/leg100/otf/issues/411)) ([c443fc1](https://github.com/leg100/otf/commit/c443fc1194eada28536c05733a36416fad5f7bad))
* log the reason why agent is retrying api request ([1d41f61](https://github.com/leg100/otf/commit/1d41f61a1257da0bf473af9efc969f0f21ca78f8))
* more workspace VCS settings ([#545](https://github.com/leg100/otf/issues/545)) ([abfc702](https://github.com/leg100/otf/commit/abfc702e8bce25842da08a655e38fee8a4ccc75a))
* notification configurations ([#428](https://github.com/leg100/otf/issues/428)) ([60d78c6](https://github.com/leg100/otf/commit/60d78c68b056e308cdf177acd3518da11b61eea7))
* organization tokens ([#528](https://github.com/leg100/otf/issues/528)) ([7ddd416](https://github.com/leg100/otf/commit/7ddd416937f6421adfafa59b0ddd60d5f35a05e6))
* pass go-tfe organization tests ([#495](https://github.com/leg100/otf/issues/495)) ([5cb3cb1](https://github.com/leg100/otf/commit/5cb3cb1e34f28a6b2ab5d7faa940e76656585831))
* pass go-tfe team integration tests ([#494](https://github.com/leg100/otf/issues/494)) ([77a9c7e](https://github.com/leg100/otf/commit/77a9c7e8422958c17e9881d9798f0ba7148d748c))
* queue destroy plan in UI ([#410](https://github.com/leg100/otf/issues/410)) ([4aff44c](https://github.com/leg100/otf/commit/4aff44c42e863adc6165b93269a6aa6bd3c549de))
* record who created a run ([#556](https://github.com/leg100/otf/issues/556)) ([57bb9b6](https://github.com/leg100/otf/commit/57bb9b6fad3445cdf830ae782ca3b07b6b024179))
* replace zerolog with slog for logging ([#396](https://github.com/leg100/otf/issues/396)) ([527f3bf](https://github.com/leg100/otf/commit/527f3bf57c636ee55cb3c5a3f04ab1acef5d1bfe))
* retry run via UI ([#438](https://github.com/leg100/otf/issues/438)) ([532df3d](https://github.com/leg100/otf/commit/532df3d62c9b67d5074cc4190d720eec54cf66f1))
* run status metrics ([#697](https://github.com/leg100/otf/issues/697)) ([4642e5d](https://github.com/leg100/otf/commit/4642e5d98937ca036f68681ab5abef9ed14379d2))
* screenshots for documentation ([#441](https://github.com/leg100/otf/issues/441)) ([9ce60a8](https://github.com/leg100/otf/commit/9ce60a8a38d42bdd1ea8153a0932fe81eb86ab87))
* support creating plan-only runs ([#465](https://github.com/leg100/otf/issues/465)) ([3f9c31e](https://github.com/leg100/otf/commit/3f9c31edb33f4d9941062b74acf3a34b9997990d))
* terraform login protocol ([#424](https://github.com/leg100/otf/issues/424)) ([2e627ca](https://github.com/leg100/otf/commit/2e627cabd462b3e511c9660d021457944c084b7b))
* ui improvements ([#406](https://github.com/leg100/otf/issues/406)) ([7a41215](https://github.com/leg100/otf/commit/7a41215f3ff3fcceee88d9e0c649ac672a7235cc))
* **ui:** add icon in run widget to show source of run ([#563](https://github.com/leg100/otf/issues/563)) ([2e7a0bd](https://github.com/leg100/otf/commit/2e7a0bd71b99556360070337b8e9baad3a021aad))
* **ui:** add tags to workspace widget ([#543](https://github.com/leg100/otf/issues/543)) ([3000c09](https://github.com/leg100/otf/commit/3000c097d50d47f4fdd6c987e1e41a609fa92f16))
* **ui:** clickable widgets ([#597](https://github.com/leg100/otf/issues/597)) ([518452e](https://github.com/leg100/otf/commit/518452ede3d458e1bd0105f2a0a46ab5b5cb36c6))
* **ui:** clicking on workspace widget tag filters by that tag ([a7ce9a8](https://github.com/leg100/otf/commit/a7ce9a890dfed4976c42619de09285cf6dd2b70d))
* **ui:** hide functionality from unpriv persons ([#548](https://github.com/leg100/otf/issues/548)) ([fee491f](https://github.com/leg100/otf/commit/fee491fa0d3c6fee5ce62ecf4c2c3dfd154011ba)), closes [#540](https://github.com/leg100/otf/issues/540)
* **ui:** provide more vcs metadata for runs ([#552](https://github.com/leg100/otf/issues/552)) ([18217ce](https://github.com/leg100/otf/commit/18217ce43b357d4107e12b5bd52984346da800a4))
* **ui:** show resources and outputs on workspace page ([#542](https://github.com/leg100/otf/issues/542)) ([d792e23](https://github.com/leg100/otf/commit/d792e239733c57d7821957ece8c2704f7e080347)), closes [#308](https://github.com/leg100/otf/issues/308)
* **ui:** show running times ([#635](https://github.com/leg100/otf/issues/635)) ([7337c2e](https://github.com/leg100/otf/commit/7337c2ecde3876c51ab77ae477f7664c264f42a3)), closes [#604](https://github.com/leg100/otf/issues/604)
* **ui:** tag search/dropdown menu ([#523](https://github.com/leg100/otf/issues/523)) ([09b8310](https://github.com/leg100/otf/commit/09b83105e10f882283419b1645d49e2c04929774))
* update vcs provider token ([#594](https://github.com/leg100/otf/issues/594)) ([29a0be6](https://github.com/leg100/otf/commit/29a0be667046440aab25efc25c9a7a02720d2f96)), closes [#576](https://github.com/leg100/otf/issues/576)
* variable sets ([#574](https://github.com/leg100/otf/issues/574)) ([419e2fb](https://github.com/leg100/otf/commit/419e2fb81cdb8a3b6b9cc7d91e81ca7af29d3a24)), closes [#306](https://github.com/leg100/otf/issues/306)
* workspace search on UI ([#461](https://github.com/leg100/otf/issues/461)) ([4f30539](https://github.com/leg100/otf/commit/4f3053967f74d8cfe7e92dd1c0d9ccb94279c469))


### Bug Fixes

* add arm64 support for terraform binary download ([#430](https://github.com/leg100/otf/issues/430)) ([cf7046b](https://github.com/leg100/otf/commit/cf7046ba75d2515ac3aa27a451518b63388cc3b0))
* add extra case for gitlab repo dir name ([#654](https://github.com/leg100/otf/issues/654)) ([5424565](https://github.com/leg100/otf/commit/542456530d8551c34bd2a2d298f931dee5c52827))
* Add missing CancelRunAction to WorkspaceWriteRole ([#649](https://github.com/leg100/otf/issues/649)) ([599ddcb](https://github.com/leg100/otf/commit/599ddcb5494de845ce6fe8e91240facf3b8fb466))
* add notification actions to workspace roles ([204704c](https://github.com/leg100/otf/commit/204704cb268f7c375d58b09e371e8d85c2b3d08a))
* add trailing slash to discovery URLs ([#475](https://github.com/leg100/otf/issues/475)) ([698f8e9](https://github.com/leg100/otf/commit/698f8e9a0ba048972bfbb3c04c01815791610d0d))
* agent error reporting ([#628](https://github.com/leg100/otf/issues/628)) ([76e7dda](https://github.com/leg100/otf/commit/76e7dda7a6d6ca29c8ee1cd8feecb3def0f77c44))
* agent race error ([#537](https://github.com/leg100/otf/issues/537)) ([6b9e6b1](https://github.com/leg100/otf/commit/6b9e6b1949a0121d5b04558334ce4011fa88a3be))
* allocator restarting unnecessarily ([#666](https://github.com/leg100/otf/issues/666)) ([47f8e6f](https://github.com/leg100/otf/commit/47f8e6f74cd7fb36bf2b5eb3885bbd995bcf81c0))
* allow org members to view variable sets ([df9fa53](https://github.com/leg100/otf/commit/df9fa53fad1e51c2c0b9e4d9ac4f493c5be66fb7))
* allow updating notification url ([#485](https://github.com/leg100/otf/issues/485)) ([cf1dbac](https://github.com/leg100/otf/commit/cf1dbac005367d78712001127200899af593948f))
* always use relative links in docs ([c019ff3](https://github.com/leg100/otf/commit/c019ff3068e8c8357a929416e06e205755e6b59b))
* apply on output changes ([#501](https://github.com/leg100/otf/issues/501)) ([46cd3ef](https://github.com/leg100/otf/commit/46cd3efbffc899d180363e767d7730ee4b473b6a))
* authorize create-org action on site, not org ([98bd1c6](https://github.com/leg100/otf/commit/98bd1c680e7fef486e622c9d1835d060df444265))
* avoid hitting Github limit on commit status updates ([#688](https://github.com/leg100/otf/issues/688)) ([029e525](https://github.com/leg100/otf/commit/029e5255a5ea70ea7701f11b16add510fb03b23f))
* bad git command in chart release step ([3ed91e3](https://github.com/leg100/otf/commit/3ed91e351a23c0d5ba858ff0423035194bc0471b))
* broken mike python package for docs ([34c50e2](https://github.com/leg100/otf/commit/34c50e2f08b5a460b15ee38d9b319187d34a8516))
* bump packages to fix critical vulns ([b53a404](https://github.com/leg100/otf/commit/b53a404f91de2741681cc60e09324338cc4703a8))
* caching failures are non-fatal ([#457](https://github.com/leg100/otf/issues/457)) ([5916e73](https://github.com/leg100/otf/commit/5916e73a82f60f97c713d5b7322361292a2a6774)), closes [#453](https://github.com/leg100/otf/issues/453)
* **ci:** charts job needs release info ([f4fef03](https://github.com/leg100/otf/commit/f4fef03c3a594bdb21c63dcbe2d0c9aeef6c242d))
* cleanup after extracting repo tarball ([bf4758b](https://github.com/leg100/otf/commit/bf4758bead52e6c3bf1e47d1dfe06ebcff0a26a8))
* connected workspace access for unpriv user ([#435](https://github.com/leg100/otf/issues/435)) ([f1471c2](https://github.com/leg100/otf/commit/f1471c2d33183092b1a3dd34a00bf2e80709f8ea))
* copy content to clipboard without whitespace ([#447](https://github.com/leg100/otf/issues/447)) ([1b1ef10](https://github.com/leg100/otf/commit/1b1ef10a25ca68a1c70a888b84a4d9b220dd9b45))
* delete existing unreferenced webhooks too ([6b61b48](https://github.com/leg100/otf/commit/6b61b485198be0b2074bd53c1633649831855588))
* delete unreferenced tags ([#507](https://github.com/leg100/otf/issues/507)) ([d85ac43](https://github.com/leg100/otf/commit/d85ac430faffc2afa1367a96e623001b38a98690)), closes [#502](https://github.com/leg100/otf/issues/502)
* delete webhooks when org or vcs provider is deleted ([#518](https://github.com/leg100/otf/issues/518)) ([0d36ea5](https://github.com/leg100/otf/commit/0d36ea554f1c3a521069426c4643b7c63a73be36))
* docker-compose otfd healthcheck ([c553b58](https://github.com/leg100/otf/commit/c553b5895ff9bc8993991c872a31d74a63bc92f2))
* docs grammar ([377c17e](https://github.com/leg100/otf/commit/377c17efed8207eec3f93a0aed0c837518c6baec))
* **docs:** version using tag not branch name ([8613fe8](https://github.com/leg100/otf/commit/8613fe88ce9d0d8fab939d5784d9bd114bdbf6b1))
* don't assume public schema for postgres migrations ([e7510b0](https://github.com/leg100/otf/commit/e7510b0fabb487f76a6f5122bb6e2708fadb3a75))
* don't reveal info on healthz endpoint ([#489](https://github.com/leg100/otf/issues/489)) ([af595d4](https://github.com/leg100/otf/commit/af595d409f38742aab53187951045c4733b3ea94))
* don't scrub included state output sensitive values ([478e314](https://github.com/leg100/otf/commit/478e314a687f722125653d6aa1010b8c3bf2b060))
* don't unnecessarily restart scheduler ([#689](https://github.com/leg100/otf/issues/689)) ([d240965](https://github.com/leg100/otf/commit/d24096536d06da30b035d846c7eb62537e2153a0))
* don't use chromium for browser-based tests ([#478](https://github.com/leg100/otf/issues/478)) ([ade0579](https://github.com/leg100/otf/commit/ade05796049d85b8518deaf78d7d710c2a786241))
* dont scrub sensitive variable values for agent ([#591](https://github.com/leg100/otf/issues/591)) ([a333ee6](https://github.com/leg100/otf/commit/a333ee6f7a04c234dbe5c34602a35f1095f35b32)), closes [#590](https://github.com/leg100/otf/issues/590)
* download terraform first to avoid test flakiness ([6370660](https://github.com/leg100/otf/commit/637066069650cf9430341660ce3356720b7f2a91))
* embed magnifying glass icon ([8a45d51](https://github.com/leg100/otf/commit/8a45d513a436bf42072460d5351bcc2380e5e961))
* enable livereload on error page too ([e5a2f01](https://github.com/leg100/otf/commit/e5a2f01ca840edc688daa4c4cffe573523bb2f86))
* error 'schema: converter not found for integration.manifest' ([e53ebf2](https://github.com/leg100/otf/commit/e53ebf2e34288e437b11d69eba3e61324b21be22))
* finish events refactor ([#509](https://github.com/leg100/otf/issues/509)) ([096933a](https://github.com/leg100/otf/commit/096933a5affb2e0a33d61dd4503a7793465ea1ac))
* fixed bug where proxy was ignored ([#609](https://github.com/leg100/otf/issues/609)) ([c1ee8d8](https://github.com/leg100/otf/commit/c1ee8d8ea53a05935c7d5d510054a6eaf588aa25))
* fixed defect with multiline tfvars not being escaped ([#631](https://github.com/leg100/otf/issues/631)) ([f35dffa](https://github.com/leg100/otf/commit/f35dffa97bec141491c1121fd10f39f5ca7893a1))
* flaky browser tests ([#484](https://github.com/leg100/otf/issues/484)) ([1ce0bd0](https://github.com/leg100/otf/commit/1ce0bd0aa47fde48d9d58f239edb9ee337d1e092))
* github PR updates not handled ([#399](https://github.com/leg100/otf/issues/399)) ([eb4d587](https://github.com/leg100/otf/commit/eb4d58777fd2bf94f00227e23f3b1a45c24963ce))
* gitlab support ([#665](https://github.com/leg100/otf/issues/665)) ([eaf9b15](https://github.com/leg100/otf/commit/eaf9b15556159079d2770064f8d374f627615ea7)), closes [#651](https://github.com/leg100/otf/issues/651)
* handle run-events request from terraform cloud backend ([#534](https://github.com/leg100/otf/issues/534)) ([b1998bd](https://github.com/leg100/otf/commit/b1998bd00450f296a5186c1d0464e93247655e86))
* helm chart branch name ([b77dc8a](https://github.com/leg100/otf/commit/b77dc8abaa4ff7bc3be0f71f84e14ab7b00dc010))
* incorrect workspace queue report ([00d04b0](https://github.com/leg100/otf/commit/00d04b03dfb507247420c6551efa9663535dffba))
* inproper setting of max cache size ([3b3ece4](https://github.com/leg100/otf/commit/3b3ece4df6b3082d3c7bad1b18162f924fc6ca02))
* insufficient github perms triggering nil pointer err ([da51e5d](https://github.com/leg100/otf/commit/da51e5d6658697d8f1f2b8eae26cb5190bb62b16))
* **integration:** ensure text box is visible before focusing ([8d279ae](https://github.com/leg100/otf/commit/8d279aefdc8830b32cb262e8608ff394a2f62880))
* **integration:** prevent -32000 error ([39318f1](https://github.com/leg100/otf/commit/39318f1dd1966f621bfb930bf2f8cbee2c70266d))
* **integration:** stop browser test failing with -32000 error ([27f02cd](https://github.com/leg100/otf/commit/27f02cd9f22f2f94d4427964f64417c0fdec83a0))
* **integration:** wait for alpinejs to load ([346024e](https://github.com/leg100/otf/commit/346024ea87eedabfd086ea536c5ee79d19b531fa))
* internal migration broke dev mode ([ea306b6](https://github.com/leg100/otf/commit/ea306b6495c7603e3d64efe7956035f453e88018))
* iterating map without mutex ([22f4fdc](https://github.com/leg100/otf/commit/22f4fdcb4d633d287941781b993284b471c6c928))
* linux/arm64 support ([#562](https://github.com/leg100/otf/issues/562)) ([01a2112](https://github.com/leg100/otf/commit/01a211240e4dca4d18e02d49e3f9d6190754510c)), closes [#311](https://github.com/leg100/otf/issues/311)
* log create agent pool db errors ([ece6de7](https://github.com/leg100/otf/commit/ece6de7186d54bbeb390f8d6e7bf5c9e1c0eb543))
* log max config size exceeded ([#663](https://github.com/leg100/otf/issues/663)) ([e196837](https://github.com/leg100/otf/commit/e196837fe88fc41b3b908537766db0b66530d281)), closes [#652](https://github.com/leg100/otf/issues/652)
* log user spec correctly ([57bfd35](https://github.com/leg100/otf/commit/57bfd3507cc4521f88fc9d0c3bae5de68ddba7fa))
* make goreleaser config valid for v2 ([0ed8d22](https://github.com/leg100/otf/commit/0ed8d22140570aefefdcabd3a16cdd46eefcebf4))
* make linting and tests pass ([ebc1e53](https://github.com/leg100/otf/commit/ebc1e53a9e00412e4db8467a71b299e86b5fb43e))
* make tern migration postgres 12 compat ([#695](https://github.com/leg100/otf/issues/695)) ([c662668](https://github.com/leg100/otf/commit/c6626686d03b8f6e2a8a34d8db1e692bf99f2b05))
* mike doc versioner flags have changed ([224081c](https://github.com/leg100/otf/commit/224081c5bbff6fd8ea0150365886573b457e25b3))
* new users see all orgs ([#397](https://github.com/leg100/otf/issues/397)) ([fe767d9](https://github.com/leg100/otf/commit/fe767d9fbdd0b48b9dd00121aae8ef7c381bfdb5))
* OIDC doc missing information on required scopes ([#444](https://github.com/leg100/otf/issues/444)) ([72191cc](https://github.com/leg100/otf/commit/72191cc19b48ef4b6a3227ef1ab0ae3a8186e15f))
* only log subsystem start when actually started ([a72fbd0](https://github.com/leg100/otf/commit/a72fbd0cc556f6f4bfbb85f79bef255512b40111))
* only set not null after populating column ([1da3936](https://github.com/leg100/otf/commit/1da3936e12532170bb6c82d3c96607f53ab50ff4))
* organization tokens ([#660](https://github.com/leg100/otf/issues/660)) ([be82c55](https://github.com/leg100/otf/commit/be82c559399a0b023aa63fe8f36e61d6fb9a9848))
* otf-agent auth failure ([#446](https://github.com/leg100/otf/issues/446)) ([5889626](https://github.com/leg100/otf/commit/5889626969e3be175c75ce6d69370dd2ab9c53d7))
* otfd compose healthcheck: curl not installed ([9f52021](https://github.com/leg100/otf/commit/9f52021d7515b4736547d8e978dcabd756d5c263))
* permit workspace write role to delete variable ([186f904](https://github.com/leg100/otf/commit/186f904d251ba7ccfe93bfe31e9ce63c21ecb95b))
* pin version of gcp pub-sub emulator docker image ([8048e72](https://github.com/leg100/otf/commit/8048e720d3d7a99c1f244ba797f6276c1dd56938))
* prevent empty owners team ([#499](https://github.com/leg100/otf/issues/499)) ([a77c9e9](https://github.com/leg100/otf/commit/a77c9e98aa25f1b3b35041f7680d5298f712f10b))
* prevent modules with no published versions from crashing otf ([#611](https://github.com/leg100/otf/issues/611)) ([84aa299](https://github.com/leg100/otf/commit/84aa2992856b87ad17b6dd582ee4528c01873b69))
* prevent subsystem failure from stopping otfd ([e5061b0](https://github.com/leg100/otf/commit/e5061b04bf051c9ad7f9e4ba7408cc3c4afa1a43))
* produce doc screenshots only when specified ([#460](https://github.com/leg100/otf/issues/460)) ([dd49975](https://github.com/leg100/otf/commit/dd49975fd2af2707c28b35135ccd7e371bd2aad5))
* publish chart after release not before ([eceab7e](https://github.com/leg100/otf/commit/eceab7efbec1b82070cc02731f734e579d9cdd80))
* publishing multiple notifications ([96f9a85](https://github.com/leg100/otf/commit/96f9a855f973ad6b5ff23b7e4a040b07012797c0))
* qemu needed for building multi-arch images ([1aa8cf8](https://github.com/leg100/otf/commit/1aa8cf87bdb6954a01abaaef8ae26be8a0dbfa7d))
* real-time run listing updates ([#467](https://github.com/leg100/otf/issues/467)) ([07ef459](https://github.com/leg100/otf/commit/07ef459384d989f919d75829d8ae35eb80136cc2))
* redirect expired ajax requests correctly ([b0ce44c](https://github.com/leg100/otf/commit/b0ce44cda392ad05fc40ffe2e634b3628a19f457))
* remove owners from permission-assignable teams ([422da90](https://github.com/leg100/otf/commit/422da90231c91d7148c8e6883c30607ca2fec60f))
* remove trailing slash from requests ([#516](https://github.com/leg100/otf/issues/516)) ([c1ee39e](https://github.com/leg100/otf/commit/c1ee39e73bfe03e2de2b3dcc9a745ea5c99985f5)), closes [#496](https://github.com/leg100/otf/issues/496)
* remove unused `groups` OIDC scope ([#558](https://github.com/leg100/otf/issues/558)) ([3dd465a](https://github.com/leg100/otf/commit/3dd465a6992cce43996e712a13af6e84782558e7)), closes [#557](https://github.com/leg100/otf/issues/557)
* remove versions endpoint causing 404 ([0fd7451](https://github.com/leg100/otf/commit/0fd7451f09450688b29ee7bc552d58ba44222c9e))
* report 409 when cancel or force cancel not allowed ([#693](https://github.com/leg100/otf/issues/693)) ([dbe5668](https://github.com/leg100/otf/commit/dbe5668cb8fab94eb510d64f2962f523efef8b46))
* restart spooler when broker terminates subscription ([#600](https://github.com/leg100/otf/issues/600)) ([ce41580](https://github.com/leg100/otf/commit/ce41580f1640c282ae89437eb377a8554232c412))
* resubscribe subsystems when their subscription is terminated ([#593](https://github.com/leg100/otf/issues/593)) ([3195e17](https://github.com/leg100/otf/commit/3195e17fe3e98ec418e0bbef6e4e46bc707a4f6c))
* retrieving state outputs only requires read role ([#603](https://github.com/leg100/otf/issues/603)) ([25c4a99](https://github.com/leg100/otf/commit/25c4a992fac150aca02a51c5d655d6364d159dca))
* retry run should use existing run properties ([49303ec](https://github.com/leg100/otf/commit/49303ecf42edb106def169ddf68f66df7558b741))
* run tailwind css on template changes ([e749013](https://github.com/leg100/otf/commit/e7490133ed74bf1278f2b519ab58ebd8a7dd4820))
* run w/o config ID pull config from repo ([#482](https://github.com/leg100/otf/issues/482)) ([0b53365](https://github.com/leg100/otf/commit/0b53365174d7a09acb6fc3907a53fa1f55511666))
* runner unable to re-register ([#707](https://github.com/leg100/otf/issues/707)) ([41f5669](https://github.com/leg100/otf/commit/41f5669ef362dc42507b8a8c61c074e022c9bb02))
* **scheduler:** ignore deleted run events ([60496bb](https://github.com/leg100/otf/commit/60496bb4849d64393c572a89f4969b45257c6b60))
* set module status ([#586](https://github.com/leg100/otf/issues/586)) ([8141c6e](https://github.com/leg100/otf/commit/8141c6ed2da175700405cb5c5f34658660cb68e7))
* set version correctly via go-releaser too ([4345b9c](https://github.com/leg100/otf/commit/4345b9ccce5ed3e56cd32c0c628f249e9338b38e))
* skip goreleaser git checks ([943490e](https://github.com/leg100/otf/commit/943490e9b4b8acf0297173adfe0a07285306231d))
* skip reporting runs created via API ([#622](https://github.com/leg100/otf/issues/622)) ([5d4527b](https://github.com/leg100/otf/commit/5d4527b52573c8600d49ed149ea16bdb7f57f141)), closes [#618](https://github.com/leg100/otf/issues/618)
* space out all UI forms ([5f2d7b5](https://github.com/leg100/otf/commit/5f2d7b5ea6fc37a7dda15592cf5106fe915a2a1c))
* state output values are json ([#477](https://github.com/leg100/otf/issues/477)) ([c2b60c0](https://github.com/leg100/otf/commit/c2b60c09822dd49e3aceb6527eb4813eff8fc2e0))
* state version output API ([#422](https://github.com/leg100/otf/issues/422)) ([9adb486](https://github.com/leg100/otf/commit/9adb486d6cfd132f1035def8e09b45e0f19774c2))
* terraform apply partial state updates ([#539](https://github.com/leg100/otf/issues/539)) ([d25e7e4](https://github.com/leg100/otf/commit/d25e7e4678ca55d49a6dfdf041de077187d5a54a)), closes [#527](https://github.com/leg100/otf/issues/527)
* **tests:** hard link fails when /tmp is separate partition ([cfc7aaa](https://github.com/leg100/otf/commit/cfc7aaa80d31c9b7a0b4e461d13bb09fea9f87bc))
* tf cli confusing locked ws for blocked queuea# ([f62ffd3](https://github.com/leg100/otf/commit/f62ffd3588bc6090295104ad7cf27c7488cd6361))
* tfe_outputs resource ([#599](https://github.com/leg100/otf/issues/599)) ([89de01d](https://github.com/leg100/otf/commit/89de01d48c1878982a7f56e436c8904bd3bc0a09)), closes [#595](https://github.com/leg100/otf/issues/595)
* tone down bold text in UI ([8d47ca1](https://github.com/leg100/otf/commit/8d47ca180b0a3c56e453ea9ec433e7f5b092d3cc))
* ui workspace permissions selector ([#405](https://github.com/leg100/otf/issues/405)) ([b5143d7](https://github.com/leg100/otf/commit/b5143d7f54215eadfd208a3e7a32c570a2a9206e))
* **ui:** add cache-control header to static files ([061261f](https://github.com/leg100/otf/commit/061261f032aed1d18054ef03c960762695e64aef))
* **ui:** allow variable to be updated from hcl to non-hcl ([ac0ff5a](https://github.com/leg100/otf/commit/ac0ff5ae654c18525632d41085fce34c8cc36711))
* **ui:** bad db query broke runner listing ([d86705c](https://github.com/leg100/otf/commit/d86705c11a6182462b6a3f2303a6bcfe73c752c4))
* **ui:** deleting vcs provider no longer breaks module page ([e28b931](https://github.com/leg100/otf/commit/e28b931703848430d0943cf2606d701511b2f003))
* **ui:** improve dropdown box UX ([d67de76](https://github.com/leg100/otf/commit/d67de7696d21d0bd6c2ef93d9b06ebcfc8190ff7))
* **ui:** make workspace page title use Name, not ID ([#581](https://github.com/leg100/otf/issues/581)) ([8268643](https://github.com/leg100/otf/commit/8268643a6775e2eb492d14c2ddf374c813b86c63))
* **ui:** new team form missing borders ([0506694](https://github.com/leg100/otf/commit/0506694f862475200f23b137ba39b6af2b755fa0))
* **ui:** push docs to remote gh-pages branch ([5b3e3f4](https://github.com/leg100/otf/commit/5b3e3f4e3034aeaba6cb27855f2314c12b964112))
* **ui:** remove undefined css classes ([daf6096](https://github.com/leg100/otf/commit/daf60965418061ff4374689613bc8c2a2ce8efe8))
* **ui:** style variables table ([ed67d57](https://github.com/leg100/otf/commit/ed67d57d2298e017e1199180341e66ca46fed4be))
* **ui:** unresponsive agent pool page ([#701](https://github.com/leg100/otf/issues/701)) ([c0853f5](https://github.com/leg100/otf/commit/c0853f51f8ba359e6f96fdd43ccb089ede645598)), closes [#698](https://github.com/leg100/otf/issues/698) [#661](https://github.com/leg100/otf/issues/661)
* **ui:** workspace description missing after update ([a579b40](https://github.com/leg100/otf/commit/a579b40ffac4aa2f021afc9417ad9bb8b3b2cc49))
* **ui:** workspace listing returning 500 error ([6eb89f4](https://github.com/leg100/otf/commit/6eb89f48c73337426feb1d704a60f4471cf85942))
* **ui:** wrong heading for edit variable set variable page ([cc6f282](https://github.com/leg100/otf/commit/cc6f2827708beefe69d8e6c88d85e83502493a51))
* use `mktemp` instead of `tempdir` ([#432](https://github.com/leg100/otf/issues/432)) ([f81b893](https://github.com/leg100/otf/commit/f81b8931448d5f298b2677e3e4e4f842bdbbbc37))
* use base58 alphabet for resource IDs ([#680](https://github.com/leg100/otf/issues/680)) ([1e7d7a2](https://github.com/leg100/otf/commit/1e7d7a2b2c350c17c29fb49ae0dfbbeb31b2942d))
* use png instead of svg for font-based icons ([eae0588](https://github.com/leg100/otf/commit/eae0588d7b6de5fb4cb2e5c2ad7fa483360f308c))
* variable set variables API ([#589](https://github.com/leg100/otf/issues/589)) ([8e29da1](https://github.com/leg100/otf/commit/8e29da191122103dd76eca876c37b419e106e630)), closes [#588](https://github.com/leg100/otf/issues/588)
* variables listing overflowing ([3c65474](https://github.com/leg100/otf/commit/3c654742674688d82fa4fa998fd99b028858a46e))
* various agent pool and job bugs ([#659](https://github.com/leg100/otf/issues/659)) ([ed9b1fd](https://github.com/leg100/otf/commit/ed9b1fdb6485f8ad16f60df19021273376a3bdd4))
* vcs event postgres publish error ([#403](https://github.com/leg100/otf/issues/403)) ([265bf49](https://github.com/leg100/otf/commit/265bf49d3088063275cb61aaf40eec63bf60bb02))
* version not showing in footer ([21dfac4](https://github.com/leg100/otf/commit/21dfac4a591546936c88c5c72e9921ec93346ae6))
* webhook updates ([#454](https://github.com/leg100/otf/issues/454)) ([9b411ce](https://github.com/leg100/otf/commit/9b411cefb5bb41e25314063e5c02d457de98df47))
* workspace listing unnecessary next page link ([e836a95](https://github.com/leg100/otf/commit/e836a956969915e47fe447b966797680828a2cda))
* workspace lock UI issues ([0130a05](https://github.com/leg100/otf/commit/0130a052c7457ee863c77f48b0c619f81153a46e))
* wrong doc content ([34631db](https://github.com/leg100/otf/commit/34631dbe893c1e92211bc236abe6dfb14693af05))


### Miscellaneous

* add BSL compliance note ([6b537de](https://github.com/leg100/otf/commit/6b537de846d6410d7e765f1c9f73945d0e679090))
* add hashes to all static urls ([3650926](https://github.com/leg100/otf/commit/36509261c1f9e4c7e574fd22a9d79e6c0b0ee26d))
* add note re cloud block to allow CLI apply ([4f03544](https://github.com/leg100/otf/commit/4f03544275ac884073be221f5f8a5f88ada0552d))
* add organization UI tests ([1c7e3db](https://github.com/leg100/otf/commit/1c7e3dbaba958710d2c07aab7ac6781b950d3b37))
* Bump default terraform version to v1.5.2 ([#503](https://github.com/leg100/otf/issues/503)) ([67bc3f0](https://github.com/leg100/otf/commit/67bc3f00c2ac9aca11092c5e8c1170f0bccf1216))
* bump go ([5663eab](https://github.com/leg100/otf/commit/5663eab55cf22f5d9fe9f8a29c6df9d2908ecb20))
* bump log threshold for frequent actions ([7e2b6af](https://github.com/leg100/otf/commit/7e2b6af7866d89e95b370624067e15bb4a09068b))
* bump playwright ([af87baa](https://github.com/leg100/otf/commit/af87baa8054b470ae9c0fcc00fb3d604440c24a6))
* bump squid version ([7ce3238](https://github.com/leg100/otf/commit/7ce3238f7af3755c317a28690b1dbd8e7efed2b9))
* chromium bug fixed ([#559](https://github.com/leg100/otf/issues/559)) ([87af2c7](https://github.com/leg100/otf/commit/87af2c74e235c14241987bbcf4f67da70ccd7b4e))
* document --max-config-size flag ([b4b9ff3](https://github.com/leg100/otf/commit/b4b9ff38d24c90b8f663a63e1ad328014b5e71c0))
* document integration test verbose logging ([75272a4](https://github.com/leg100/otf/commit/75272a4b7842426e2901615f5898d02a515a310b))
* document some more flags ([e2cc4f2](https://github.com/leg100/otf/commit/e2cc4f271956e737571778de81f8e4d926fe3e55))
* downplay legitimate state not found errors ([2d91e31](https://github.com/leg100/otf/commit/2d91e313862d6e412369853f10fb48fb87068337))
* embed team ID in workspace policy ([#436](https://github.com/leg100/otf/issues/436)) ([0411174](https://github.com/leg100/otf/commit/0411174e97b68208120885a79ea84cad391e7a50))
* favicon ([b7dae88](https://github.com/leg100/otf/commit/b7dae8869604ec6e41a2639ded8a46fb475411b3))
* favicon for docs ([c07622c](https://github.com/leg100/otf/commit/c07622c20c621aaecc7562bf47bfa362d59c6bd3))
* go 1.21 ([#566](https://github.com/leg100/otf/issues/566)) ([06c13b2](https://github.com/leg100/otf/commit/06c13b250b183c12e0486e69cac2aee1c52b7ed5))
* make subsystems more resilient ([#459](https://github.com/leg100/otf/issues/459)) ([de017c5](https://github.com/leg100/otf/commit/de017c50d814063d65ecf57424c6ac22fb1869c2))
* **master:** release 0.0.51 ([#514](https://github.com/leg100/otf/issues/514)) ([36ee226](https://github.com/leg100/otf/commit/36ee226bdec4c5512bcfd3dc678017bcfa201874))
* **master:** release 0.0.52 ([#515](https://github.com/leg100/otf/issues/515)) ([dfbc041](https://github.com/leg100/otf/commit/dfbc04134c2c207c7a9782de58a5835007fb5dd9))
* **master:** release 0.0.53 ([#517](https://github.com/leg100/otf/issues/517)) ([fce9285](https://github.com/leg100/otf/commit/fce9285484a4e6ee94875ca0b018067438aa3487))
* **master:** release 0.1.0 ([#519](https://github.com/leg100/otf/issues/519)) ([75a9eef](https://github.com/leg100/otf/commit/75a9eefc8ca58f9f5851a3712508246ae0d13252))
* **master:** release 0.1.1 ([#530](https://github.com/leg100/otf/issues/530)) ([3e9321b](https://github.com/leg100/otf/commit/3e9321be6cb85e29dd0bf0b121d2669ba1f5e3c1))
* **master:** release 0.1.10 ([#587](https://github.com/leg100/otf/issues/587)) ([fc64a73](https://github.com/leg100/otf/commit/fc64a7350a81555f44163ea83b6ca0ddb9493d0a))
* **master:** release 0.1.11 ([#592](https://github.com/leg100/otf/issues/592)) ([3596ad9](https://github.com/leg100/otf/commit/3596ad95bf5a4b8cebf55b5c1d9087a0772f26c1))
* **master:** release 0.1.12 ([#598](https://github.com/leg100/otf/issues/598)) ([cc057e8](https://github.com/leg100/otf/commit/cc057e820f9767d05cea737847392db67bfd4307))
* **master:** release 0.1.13 ([#602](https://github.com/leg100/otf/issues/602)) ([cb025ed](https://github.com/leg100/otf/commit/cb025ed248c7126d158833ccc4bbf4202f01bb92))
* **master:** release 0.1.14 ([#606](https://github.com/leg100/otf/issues/606)) ([abb7f0f](https://github.com/leg100/otf/commit/abb7f0f6d749ecf8219547944104503e2faebc6b))
* **master:** release 0.1.15 ([#629](https://github.com/leg100/otf/issues/629)) ([9085f5b](https://github.com/leg100/otf/commit/9085f5bc6d91e6d39430d2da819572a0cd9bcbaf))
* **master:** release 0.1.16 ([#632](https://github.com/leg100/otf/issues/632)) ([899a371](https://github.com/leg100/otf/commit/899a37145770100010a07101a083ce2f2f25cfa4))
* **master:** release 0.1.17 ([#634](https://github.com/leg100/otf/issues/634)) ([bf1b3d4](https://github.com/leg100/otf/commit/bf1b3d4a6585788f71be09e744ecc70d232d2aae))
* **master:** release 0.1.18 ([#637](https://github.com/leg100/otf/issues/637)) ([dc7c7cc](https://github.com/leg100/otf/commit/dc7c7ccf82314e3f9d8cb6b084cf9cce235d8ee8))
* **master:** release 0.1.2 ([#531](https://github.com/leg100/otf/issues/531)) ([2b42dfd](https://github.com/leg100/otf/commit/2b42dfd9f2ac009b393609538ee10f9a93f57ae4))
* **master:** release 0.1.3 ([#541](https://github.com/leg100/otf/issues/541)) ([4dd3e03](https://github.com/leg100/otf/commit/4dd3e0358e461a86db88f5a37d24f3fa5dbc6f0c))
* **master:** release 0.1.4 ([#547](https://github.com/leg100/otf/issues/547)) ([19148a6](https://github.com/leg100/otf/commit/19148a670ab910e157715b7c70aa263e2629a458))
* **master:** release 0.1.5 ([#551](https://github.com/leg100/otf/issues/551)) ([9dbe599](https://github.com/leg100/otf/commit/9dbe599dc88659349b8130005a47262746fd3e52))
* **master:** release 0.1.6 ([#554](https://github.com/leg100/otf/issues/554)) ([031d673](https://github.com/leg100/otf/commit/031d6735f9ff3eb65f69c4fbc3a2c677e8fc86ba))
* **master:** release 0.1.7 ([#560](https://github.com/leg100/otf/issues/560)) ([6331830](https://github.com/leg100/otf/commit/63318309ce125b25ca12231a310795a3bdafcd9d))
* **master:** release 0.1.8 ([#561](https://github.com/leg100/otf/issues/561)) ([9c6d0f6](https://github.com/leg100/otf/commit/9c6d0f65628ee5786a815957df8d0df339ba9b47))
* **master:** release 0.1.8 ([#568](https://github.com/leg100/otf/issues/568)) ([4a1ef8d](https://github.com/leg100/otf/commit/4a1ef8de6899f318e6cb4449bac876ff95af1184))
* **master:** release 0.1.9 ([#569](https://github.com/leg100/otf/issues/569)) ([24038a8](https://github.com/leg100/otf/commit/24038a8e03f9f5a336811292f0f6bd1cadf48762))
* **master:** release 0.2.0 ([#644](https://github.com/leg100/otf/issues/644)) ([a0c5b7a](https://github.com/leg100/otf/commit/a0c5b7a91360c1d3e67ebdff925992826ce62d1a))
* **master:** release 0.2.1 ([#657](https://github.com/leg100/otf/issues/657)) ([2915697](https://github.com/leg100/otf/commit/291569786da3613f5864e52afd438f130e19a9a9))
* **master:** release 0.2.2 ([#664](https://github.com/leg100/otf/issues/664)) ([3c62ff1](https://github.com/leg100/otf/commit/3c62ff16ccde81fb8e192d7fd678882f44210e1b))
* **master:** release 0.2.3 ([#667](https://github.com/leg100/otf/issues/667)) ([b518d44](https://github.com/leg100/otf/commit/b518d44f2b200261204faa7904d1610360bdc369))
* **master:** release 0.2.4 ([#674](https://github.com/leg100/otf/issues/674)) ([a2d4583](https://github.com/leg100/otf/commit/a2d4583fc314f8f77b14bc604aca6ae36b21d4ec))
* **master:** release 0.3.0 ([#682](https://github.com/leg100/otf/issues/682)) ([59d22ba](https://github.com/leg100/otf/commit/59d22baa29902cac265fe07ff6bb572d8b3efb6b))
* **master:** release 0.3.1 ([#691](https://github.com/leg100/otf/issues/691)) ([7b24720](https://github.com/leg100/otf/commit/7b24720f8c76af424ba9cce52d652afe562437a2))
* **master:** release 0.3.2 ([#692](https://github.com/leg100/otf/issues/692)) ([0d9269f](https://github.com/leg100/otf/commit/0d9269f98ed769e48d6ca943bd4e78411253ba6a))
* **master:** release 0.3.3 ([#694](https://github.com/leg100/otf/issues/694)) ([45f115e](https://github.com/leg100/otf/commit/45f115e56815ebb3de0e390f30f2436129fd3b86))
* **master:** release 0.3.4 ([#700](https://github.com/leg100/otf/issues/700)) ([c8cdd4a](https://github.com/leg100/otf/commit/c8cdd4a261c90c723bcdd81448707f9b0dbadccb))
* **master:** release 0.3.5 ([#702](https://github.com/leg100/otf/issues/702)) ([1035d18](https://github.com/leg100/otf/commit/1035d18ae58ae413f7c57ada40d983f6455690e7))
* **master:** release 0.3.6 ([#706](https://github.com/leg100/otf/issues/706)) ([3511d5f](https://github.com/leg100/otf/commit/3511d5f66209fa767de0d841ccfc3d47da155551))
* **master:** release 0.3.7 ([#712](https://github.com/leg100/otf/issues/712)) ([231de97](https://github.com/leg100/otf/commit/231de974a89b8dece68a8b4afd0aa1fc7a3d74fd))
* **master:** release 0.3.8 ([#715](https://github.com/leg100/otf/issues/715)) ([36702fa](https://github.com/leg100/otf/commit/36702faf158bbb700f5e48ae9ad7886a05da1173))
* **perf:** pre-allocate slices ([ccc8b6e](https://github.com/leg100/otf/commit/ccc8b6e0a6c3195ef239323574ce3b51aa86bce9))
* remove client rate limiter ([#412](https://github.com/leg100/otf/issues/412)) ([40ece6b](https://github.com/leg100/otf/commit/40ece6b8ed446b43be06ec9027fbf77b0297e523))
* remove demo ([d70c7fd](https://github.com/leg100/otf/commit/d70c7fdfd82ce39ff0e1a1d05b4ee38ba04e0b5b))
* remove redundant CreateRun magic string ([#555](https://github.com/leg100/otf/issues/555)) ([a2df6d5](https://github.com/leg100/otf/commit/a2df6d5247d1e605fe852eb2ebe4cf7e2b35f795))
* remove redundant jsonapiclient interface ([5aa153a](https://github.com/leg100/otf/commit/5aa153a9822d86714845509c5f15c962321382cd))
* remove unnecessary link from widget heading ([318c390](https://github.com/leg100/otf/commit/318c39052ebcbbee187dbc2a08a0a456dab70352))
* remove unused cloud team and org sync code ([4e1817d](https://github.com/leg100/otf/commit/4e1817dbbd21093c835e84d921606dd2ae46f871))
* remove unused exchange code response ([4a966cd](https://github.com/leg100/otf/commit/4a966cd8cbfc1c4232c1ebe7b83c62044a2a8af2))
* removed unused ca.pem ([799ed25](https://github.com/leg100/otf/commit/799ed25565c155c616e90533b3172bc22f916f6b))
* removed unused config file ([84fe3b1](https://github.com/leg100/otf/commit/84fe3b1a6caf4db7611d912b3316747705209e39))
* skip api tests if env vars not present ([5b88474](https://github.com/leg100/otf/commit/5b88474d3c4813897f39f3b463d013cbc831ad64))
* test create connected workspace via api ([9bf4bae](https://github.com/leg100/otf/commit/9bf4bae2d7d26c52a169302dca2f7c2ef11c1cde))
* trigger new version of agent chart upon deploy ([#690](https://github.com/leg100/otf/issues/690)) ([155e026](https://github.com/leg100/otf/commit/155e026252c52d4ec7565f9bb7444b297bca16f7))
* trigger release ([e90e046](https://github.com/leg100/otf/commit/e90e0469c0a46a16112891b19a9e8877c790283b))
* **ui:** make tags less bulbous ([df1645d](https://github.com/leg100/otf/commit/df1645d8de9d4ce021d93e58f03d27911494649f))
* **ui:** make workspace state tabs look nicer ([bbe38b4](https://github.com/leg100/otf/commit/bbe38b4e0ee6808523ac52687b8544e308233a7a))
* **ui:** pad out buttons on consent page ([1c290e9](https://github.com/leg100/otf/commit/1c290e93248d9620d54c41eb3681065929069cde))
* unarchive ([c954c36](https://github.com/leg100/otf/commit/c954c36107c910192941393e366ee1b8a2e290f1))
* update docs ([364d183](https://github.com/leg100/otf/commit/364d183dd8635eb0ce73b1e65666475ab0a039ea))
* update screenshots ([b17ffa4](https://github.com/leg100/otf/commit/b17ffa45394cfa23343637529df4ff46bcac97ab))
* upgrade dependencies ([59eb979](https://github.com/leg100/otf/commit/59eb97996785328e78c31f2372cfce889f5ef7cb))
* upgrade vulnerable markdown go mod ([781e0f6](https://github.com/leg100/otf/commit/781e0f6e047abe662336250e679797f1b3ed0752))
* use forked sse lib's module path ([fc9b138](https://github.com/leg100/otf/commit/fc9b13865844062570ffbeb7d8c57bdf5ae50b91))
* use new surl v2 lib ([0c5b45c](https://github.com/leg100/otf/commit/0c5b45cb32a7e7743b11631b4a14f65494141039))
* validate resource names ([c7596fe](https://github.com/leg100/otf/commit/c7596febc1018a546ec2c990ae5087ae297df8c0))

## [0.3.8](https://github.com/leg100/otf/compare/v0.3.7...v0.3.8) (2025-01-11)


### Features

* add option to skip plans for pull requests ([#714](https://github.com/leg100/otf/issues/714)) ([29455ed](https://github.com/leg100/otf/commit/29455ed335cf556702bca750bcc18e5b53d82a3e))


### Bug Fixes

* download terraform first to avoid test flakiness ([6370660](https://github.com/leg100/otf/commit/637066069650cf9430341660ce3356720b7f2a91))
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
