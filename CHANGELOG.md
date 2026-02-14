# Changelog

## [0.5.14](https://github.com/leg100/otf/compare/v0.5.13...v0.5.14) (2026-02-14)


### Bug Fixes

* force helm chart release ([16fb604](https://github.com/leg100/otf/commit/16fb604008e6cb84ad740eb40ed957ff1e1c3470))


### Miscellaneous

* bump helm charts with app version v0.5.13 ([bdf950f](https://github.com/leg100/otf/commit/bdf950fddbed894c2fe0ebee16e96ac9a6e6ae1c))

## [0.5.13](https://github.com/leg100/otf/compare/v0.5.12...v0.5.13) (2026-02-13)


### Bug Fixes

* broken github workflow ([bf659cc](https://github.com/leg100/otf/commit/bf659cc65d60593ef4c4e95bd6501460d29b1716))
* helm charts not being automatically updated upon release ([bb35d8d](https://github.com/leg100/otf/commit/bb35d8d9fce25f4536577271e00e4c5ed521eaed))


### Miscellaneous

* bump dependencies ([6168747](https://github.com/leg100/otf/commit/616874790518d905ddda66c6b585bedb23780c37))
* bump to go1.26 and adopt new func ([25c12ec](https://github.com/leg100/otf/commit/25c12ec064046154f137d7d2381a1a74082df3ec))

## [0.5.12](https://github.com/leg100/otf/compare/v0.5.11...v0.5.12) (2026-02-12)


### Bug Fixes

* don't append api. to github enterprise URLs ([b8c9b06](https://github.com/leg100/otf/commit/b8c9b062ab21abf7ec6111f484762094f2a5060d))

## [0.5.11](https://github.com/leg100/otf/compare/v0.5.10...v0.5.11) (2026-02-12)


### Bug Fixes

* **api-client:** allow '@' in state output map keys ([79458e6](https://github.com/leg100/otf/commit/79458e648541824970ccc7344b511cc8815b14c8))

## [0.5.10](https://github.com/leg100/otf/compare/v0.5.9...v0.5.10) (2026-02-11)


### Bug Fixes

* allow '@' in state output map keys ([9f838ee](https://github.com/leg100/otf/commit/9f838eed1ed7d3502d6f5f9c193f7c30b559aae7))
* **api-client:** provide more informative error logging ([27d7888](https://github.com/leg100/otf/commit/27d7888da0c2f07283e682cc07d7876b589ed461))
* **api:** log error when response cannot be marshaled ([fef1bc7](https://github.com/leg100/otf/commit/fef1bc79ddc13a6b0d23d820badacdc009b8cf2a))

## [0.5.9](https://github.com/leg100/otf/compare/v0.5.8...v0.5.9) (2026-02-10)


### Bug Fixes

* force otfd helm chart release ([a4f6e6c](https://github.com/leg100/otf/commit/a4f6e6cd363738255984b51e2013de8ce6495a54))
* remove unnecessary string casting ([facada8](https://github.com/leg100/otf/commit/facada8c07bbe4b206b571fd7b2df2639d6b53af))

## [0.5.8](https://github.com/leg100/otf/compare/v0.5.7...v0.5.8) (2026-02-09)


### Features

* **ui:** revamp menu layout ([#884](https://github.com/leg100/otf/issues/884)) ([0510bce](https://github.com/leg100/otf/commit/0510bce29341a11795147387754049b91b714d9a))


### Bug Fixes

* **charts/otfd:** Adjust indentation for volume mounts in deployment.yaml ([#881](https://github.com/leg100/otf/issues/881)) ([67341d3](https://github.com/leg100/otf/commit/67341d38f2decd7a206921595bca217d8efa8462))
* **dockerfile:** bump alpine version and upgrade base packages to fix critical CVE ([#882](https://github.com/leg100/otf/issues/882)) ([3e730d7](https://github.com/leg100/otf/commit/3e730d73468d32cb0e93a9207393de91a2bfe143))
* sync layout templ code ([3cf0505](https://github.com/leg100/otf/commit/3cf0505816df8a5d475488d06ecf290c70856885))

## [0.5.7](https://github.com/leg100/otf/compare/v0.5.6...v0.5.7) (2026-02-01)


### Bug Fixes

* **dockerfile:** helm provider requires home dir ([86bf7b7](https://github.com/leg100/otf/commit/86bf7b7269843e11cc1454e350298c8d8afecbcf))
* tags regex empty string and non-zero trigger patterns should not throw error ([#874](https://github.com/leg100/otf/issues/874)) ([741fc3b](https://github.com/leg100/otf/commit/741fc3bffbdebdd5bab4e9d649f9a48cd398811f))
* **ui:** auto show force cancel button after cool off period ([72b74ca](https://github.com/leg100/otf/commit/72b74caa3a636e7069ebf86dc1773d7c726238a3))


### Miscellaneous

* parameterize ld flags in makefile ([92fd5a4](https://github.com/leg100/otf/commit/92fd5a456eb817e11d6537772e5938b9a497661a))
* remove unnecessary install-linter make task ([dc36aac](https://github.com/leg100/otf/commit/dc36aaca47a3c4595fd626806de02a2e9a031704))

## [0.5.6](https://github.com/leg100/otf/compare/v0.5.5...v0.5.6) (2026-01-28)


### Bug Fixes

* **ci:** skip redundant helm chart test ([8f90e9e](https://github.com/leg100/otf/commit/8f90e9e826d7827915d6d7f1c3a429572918aaf8))

## [0.5.5](https://github.com/leg100/otf/compare/v0.5.4...v0.5.5) (2026-01-28)


### Bug Fixes

* **ci:** support multi platform docker builds ([6af386f](https://github.com/leg100/otf/commit/6af386f9a580465133d6a5cbc6ffdfb01c557e73))

## [0.5.4](https://github.com/leg100/otf/compare/v0.5.3...v0.5.4) (2026-01-28)


### Bug Fixes

* **ci:** disable sbom for docker images ([7d56ea8](https://github.com/leg100/otf/commit/7d56ea8555c5484046919fbbc940b5c90cd990b0))

## [0.5.3](https://github.com/leg100/otf/compare/v0.5.2...v0.5.3) (2026-01-27)


### Bug Fixes

* **ci:** downgrade broken release please action ([f886124](https://github.com/leg100/otf/commit/f886124ea6f560f1b5be94afd1f6d33c203d8b4b))

## [0.5.2](https://github.com/leg100/otf/compare/v0.5.1...v0.5.2) (2026-01-27)


### Bug Fixes

* **ci:** broke release process ([64716f9](https://github.com/leg100/otf/commit/64716f9bf7ce4edfb26d038fae518cb8e0cd24c8))

## [0.5.1](https://github.com/leg100/otf/compare/v0.5.0...v0.5.1) (2026-01-26)


### Features

* docker compose install ([#862](https://github.com/leg100/otf/issues/862)) ([c9946a6](https://github.com/leg100/otf/commit/c9946a6858f1c60d876eb245959174d6526d231b)), closes [#850](https://github.com/leg100/otf/issues/850)
* execute runs in kubernetes jobs ([#867](https://github.com/leg100/otf/issues/867)) ([0d22da8](https://github.com/leg100/otf/commit/0d22da8d893236173f9f6f118d050c45fd41b830)), closes [#866](https://github.com/leg100/otf/issues/866)
* passthrough command-line parameters which affect planning ([#859](https://github.com/leg100/otf/issues/859)) ([b9405d0](https://github.com/leg100/otf/commit/b9405d09b58431a820a0bec94969d929dc5bd436))
* run as non-root user in container ([#823](https://github.com/leg100/otf/issues/823)) ([06ced51](https://github.com/leg100/otf/commit/06ced5159e610998833a5a5ffad36395fbf0ff18))


### Bug Fixes

* **agent:** don't panic when server goes offline ([6017b6e](https://github.com/leg100/otf/commit/6017b6e0a77884709b0c53eea8c14e4ef84fcde8))
* **agent:** dynamic provider creds broken on agents ([acafaee](https://github.com/leg100/otf/commit/acafaee03f50dd86b179605604faa4fac6d70dc9))
* avoid nil ptr ref panics when watching run ([cd5694d](https://github.com/leg100/otf/commit/cd5694dca4b0d5d7b0147ed056508ffeb4d686c9))
* copy clipboard notification clipped ([3703ad4](https://github.com/leg100/otf/commit/3703ad471c0205b9bfec386d9190bcc90ccac302))
* **ui:** cancel run button sometimes doesn't do anything ([d29a0ef](https://github.com/leg100/otf/commit/d29a0ef7ea55abf4060b190365c799ea3cc1c159))


### Miscellaneous

* add more logging and error context ([a7adab5](https://github.com/leg100/otf/commit/a7adab5013f8054b82f821d2485e3c32c185394f))
* bump everything ([73a0f6a](https://github.com/leg100/otf/commit/73a0f6ad2ca3113263ba29036705e83000c8b6c5))
* **ci:** update helm-docs command ([8c6a157](https://github.com/leg100/otf/commit/8c6a1574dd7de7ddabdd2fbe895891c339d38f57))

## [0.5.0](https://github.com/leg100/otf/compare/v0.4.10...v0.5.0) (2025-12-15)


### ⚠ BREAKING CHANGES

* don't ignore custom hostname overrides for VCS providers ([#855](https://github.com/leg100/otf/issues/855))

### Bug Fixes

* don't ignore custom hostname overrides for VCS providers ([#855](https://github.com/leg100/otf/issues/855)) ([c25801e](https://github.com/leg100/otf/commit/c25801e2ce4dc2289ee3d67bd0bf3f0cf62f8302))
* run error tests ([be7d10d](https://github.com/leg100/otf/commit/be7d10d47dcadd6f6cc544b778634d51e99a06f1))

## [0.4.10](https://github.com/leg100/otf/compare/v0.4.9...v0.4.10) (2025-12-14)


### Bug Fixes

* enrich forgejo list repos errors ([9cdf6f2](https://github.com/leg100/otf/commit/9cdf6f23745183d24ae230639d7eb968f3a703b3))


### Miscellaneous

* update Alpine to v3.23 ([#849](https://github.com/leg100/otf/issues/849)) ([1721470](https://github.com/leg100/otf/commit/17214707b969712dcff77d2553b992e875dac1e4))

## [0.4.9](https://github.com/leg100/otf/compare/v0.4.8...v0.4.9) (2025-11-08)


### Features

* delete configs older than a user-specified duration ([#845](https://github.com/leg100/otf/issues/845)) ([b54bd3f](https://github.com/leg100/otf/commit/b54bd3ff2548d0bac7c70e4f3e9409e69d1537fe))

## [0.4.8](https://github.com/leg100/otf/compare/v0.4.7...v0.4.8) (2025-11-02)


### Features

* delete runs after user-specified time period ([#844](https://github.com/leg100/otf/issues/844)) ([fb5547c](https://github.com/leg100/otf/commit/fb5547c75f4f6872c69e59a376b8cf0ddf25cb27)), closes [#834](https://github.com/leg100/otf/issues/834)


### Bug Fixes

* **ui:** polling table not sending page number param ([7591827](https://github.com/leg100/otf/commit/7591827fee43539c1d72eb373899607172abda83))
* workspace sql cascade set null on user or run deletion ([687c35c](https://github.com/leg100/otf/commit/687c35c37cacc499f8b42234083ad1e6fe4b00ec))


### Miscellaneous

* document dynamic credentials test setup ([efecef2](https://github.com/leg100/otf/commit/efecef20d7b5a8539ce5b77c83751a9cc9ae1cef))
* remove unnecessary double html error ([53cb12d](https://github.com/leg100/otf/commit/53cb12d25d00c664daf7ebe7eb7fcd38e609fad5))

## [0.4.7](https://github.com/leg100/otf/compare/v0.4.6...v0.4.7) (2025-10-10)


### Features

* **ui:** let browser cache responses for better perf ([c76a29a](https://github.com/leg100/otf/commit/c76a29a8e4934a78cfccfec253b3587f3471b8b6))
* **ui:** use flash messages for POST errors ([23a969f](https://github.com/leg100/otf/commit/23a969ff1712139fbc233f93ff0e26d3f30fb718))


### Bug Fixes

* exclude websockets from etag middleware ([2d2567c](https://github.com/leg100/otf/commit/2d2567c9cc6763f3f3af841dbdab23651d9e7f47))
* UI crashes when forgejo VCS provider fails to connect ([#838](https://github.com/leg100/otf/issues/838)) ([d4ba67e](https://github.com/leg100/otf/commit/d4ba67e5958ab33e5d87c342f6151e2116473f16)), closes [#837](https://github.com/leg100/otf/issues/837)


### Miscellaneous

* don't raise error unnecessarilyi ([720b93b](https://github.com/leg100/otf/commit/720b93b6a8e9d9e35eacefd91889d91753a64591))
* drop bubblewrap support ([eb747d1](https://github.com/leg100/otf/commit/eb747d16a8a2dd4dd661a8b72647d8ba717a0276))
* fix linting errors ([e7a75a9](https://github.com/leg100/otf/commit/e7a75a99cb128904f8fc7eff54d2b291e2cde5f8))
* remove unused go.mod replace directives ([74505b3](https://github.com/leg100/otf/commit/74505b3714947ed601af303129ae99c089508872))
* use constants where they already exist ([4551222](https://github.com/leg100/otf/commit/45512221b1708a3fbd00881f225d698753b0b83a))

## [0.4.6](https://github.com/leg100/otf/compare/v0.4.5...v0.4.6) (2025-10-01)


### Bug Fixes

* add caching to speed up run listings ([cbfeceb](https://github.com/leg100/otf/commit/cbfeceb054162e6e2a7243391265f30003bea761))
* add indices to boost perf and reduce db load ([04678fc](https://github.com/leg100/otf/commit/04678fc3a8d852d8619305ae181424e47a296090))
* avoid websocket handler infinite loop ([92ea015](https://github.com/leg100/otf/commit/92ea01547fd9c954bb7b56b66ac0e8ba331c6363))
* improve performance of run metrics collector ([d421c16](https://github.com/leg100/otf/commit/d421c16756e1f2bd050b4ecb6cf97e43b57264e0))
* update helm chart lock file ([b5aa360](https://github.com/leg100/otf/commit/b5aa360a255d5d2b7c2f378124bd2b7c1e0d4fd6))


### Miscellaneous

* change default engine notice wording ([620672f](https://github.com/leg100/otf/commit/620672fdabfa5a77095c96d6b8e21be35b49354f))

## [0.4.5](https://github.com/leg100/otf/compare/v0.4.4...v0.4.5) (2025-09-30)


### Bug Fixes

* swap out bitnami postgres helm chart ([539cd9c](https://github.com/leg100/otf/commit/539cd9c8c8fe9f074350c467a3430aee74dfb510))

## [0.4.4](https://github.com/leg100/otf/compare/v0.4.3...v0.4.4) (2025-09-29)


### Bug Fixes

* module markdown rendering ([#831](https://github.com/leg100/otf/issues/831)) ([90700f1](https://github.com/leg100/otf/commit/90700f142a6ec6b81f82ad86ddc040b658a8dceb)), closes [#820](https://github.com/leg100/otf/issues/820)
* remove unnecessary tool install make tasks ([acd9d6f](https://github.com/leg100/otf/commit/acd9d6f19ddd2ba9ceb9ffc6c564ed50571f15c2))


### Miscellaneous

* bump daisyui version ([dfd3904](https://github.com/leg100/otf/commit/dfd3904bffa72a4347cd9c054c32e43bc377de12))
* bump templ version ([15768e0](https://github.com/leg100/otf/commit/15768e0992164e28d86c2a8d3b5ee6266f23787c))

## [0.4.3](https://github.com/leg100/otf/compare/v0.4.2...v0.4.3) (2025-09-27)


### Bug Fixes

* azure dyn cred provider no kid found error ([#828](https://github.com/leg100/otf/issues/828)) ([6d55d8e](https://github.com/leg100/otf/commit/6d55d8e78fcb1e787a150c6a3e9bbb241e1147dc))

## [0.4.2](https://github.com/leg100/otf/compare/v0.4.1...v0.4.2) (2025-09-23)


### Bug Fixes

* AWS dynamic provider credentials invalid apply absolute paths ([#822](https://github.com/leg100/otf/issues/822)) ([fc3162a](https://github.com/leg100/otf/commit/fc3162a35e8b51ff5d155f0eb097c3f96a03786e))
* handle tfe provider agent pool ID empty string gracefully ([48c7f45](https://github.com/leg100/otf/commit/48c7f45a3e2ac729a7ad8051d0dd3ed9629aa9a9))

## [0.4.1](https://github.com/leg100/otf/compare/v0.4.0...v0.4.1) (2025-08-27)


### Bug Fixes

* AWS dynamic provider credentials requires kid field (Key ID) in JWK ([7d1bc8b](https://github.com/leg100/otf/commit/7d1bc8bd149b537d9205e2a6d24dc2411961e353))
* **ui:** remove extra semicolons ([#818](https://github.com/leg100/otf/issues/818)) ([1ee991c](https://github.com/leg100/otf/commit/1ee991cbbd20bb39394f2d4375f9249a9292328c)), closes [#817](https://github.com/leg100/otf/issues/817)

## [0.4.0](https://github.com/leg100/otf/compare/v0.3.27...v0.4.0) (2025-07-24)


### ⚠ BREAKING CHANGES

* 

### Features

* allow user to specify URL per VCS provider ([#813](https://github.com/leg100/otf/issues/813)) ([c792756](https://github.com/leg100/otf/commit/c792756a308514e491e5b0a531ab72556dd98d4e))
* dynamic provider credentials ([#806](https://github.com/leg100/otf/issues/806)) ([aab0a9b](https://github.com/leg100/otf/commit/aab0a9b80ce9d97ec3ed487c1ecfadac3cc14c81))
* extend helm charts values to add volumes, volumeMounts and sidecars ([#807](https://github.com/leg100/otf/issues/807)) ([94147b5](https://github.com/leg100/otf/commit/94147b548cf02e78d0e5b6b59acec603ec89bcfc))


### Bug Fixes

* **ui:** hide sensitive outputs on the workspace page ([#810](https://github.com/leg100/otf/issues/810)) ([87dcc61](https://github.com/leg100/otf/commit/87dcc61bb2e920a1c074578c82a3fc7aacf8a529))
* **ui:** no delete button for variable set variable ([88449cd](https://github.com/leg100/otf/commit/88449cd0a6c8c6f01b8d02af613ed88e7e00a905))
* **ui:** page metadata info hard to see in dark mode ([1ac8cd5](https://github.com/leg100/otf/commit/1ac8cd5a872729b7ce4b227947c3f6c65699f7bc))

## [0.3.27](https://github.com/leg100/otf/compare/v0.3.26...v0.3.27) (2025-07-06)


### Bug Fixes

* **ui:** cannot unassign team permissions ([b4415b4](https://github.com/leg100/otf/commit/b4415b401c0a1367ed4d888c4008936f9c10727e))
* **ui:** edit variable set variable link 404 ([5eed13a](https://github.com/leg100/otf/commit/5eed13a8dbcee53a8ec18560834e2b1a8f48c791))

## [0.3.26](https://github.com/leg100/otf/compare/v0.3.25...v0.3.26) (2025-07-01)


### Features

* **ui:** add workspace overview sub menu link ([892381f](https://github.com/leg100/otf/commit/892381f1b96b00db54151546b63743f3b18b369f))


### Bug Fixes

* github app connected workspaces not sending status updates ([42e2ce7](https://github.com/leg100/otf/commit/42e2ce7ab039709a4c29b11001e642abac251be4))
* **ui:** make page size selector fully visible when set to 100 ([5ec4f84](https://github.com/leg100/otf/commit/5ec4f84132ceef293452d3ed32d1fb5c5e906674))


### Miscellaneous

* bump templ version ([cbd6f3c](https://github.com/leg100/otf/commit/cbd6f3c329a3edfca15c9d8bea266700991a4b15))
* remove debug statements ([c93b542](https://github.com/leg100/otf/commit/c93b5422af1c102c3d94c1dc5fa122e06803beb4))
* remove unnecessary source ptr func ([0ef66e9](https://github.com/leg100/otf/commit/0ef66e9478ad0ffafdf54f1e06a97337cc9646aa))

## [0.3.25](https://github.com/leg100/otf/compare/v0.3.24...v0.3.25) (2025-06-24)


### Features

* **ui:** user avatars ([f35efdd](https://github.com/leg100/otf/commit/f35efdd5a703282645b3239d610421d360bba6be))


### Bug Fixes

* **docs:** fix 404 links ([fb5dda0](https://github.com/leg100/otf/commit/fb5dda0a732c35ce64cad0c9e8716ed84b3e5c9d))

## [0.3.24](https://github.com/leg100/otf/compare/v0.3.23...v0.3.24) (2025-06-15)


### Features

* forgejo provider uses gitea webhooks, for compatibility with both ([#793](https://github.com/leg100/otf/issues/793)) ([9628f8a](https://github.com/leg100/otf/commit/9628f8a3dad6197c66225184f919ae1c4dee63bf)), closes [#792](https://github.com/leg100/otf/issues/792)


### Bug Fixes

* allow overriding source for a vcs kind ([40467e7](https://github.com/leg100/otf/commit/40467e7d2a7b70782baf1512b44c34145e34f8ea))
* **ci:** helm charts copied to wrong dir in otf-charts repo ([94ab37b](https://github.com/leg100/otf/commit/94ab37ba3900dc9788beecb1aad37aaf8af85767))
* **docs:** lists in forgeyo doc rendering items on same line ([db13b7a](https://github.com/leg100/otf/commit/db13b7a29896bafe145af6361f92466f11f052b3))
* **ui:** nil panic when two workspace variables have same key ([4a6ed58](https://github.com/leg100/otf/commit/4a6ed58624df505ab96ec619efce667648cd285c))
* **ui:** various VCS fixes ([#797](https://github.com/leg100/otf/issues/797)) ([99905ce](https://github.com/leg100/otf/commit/99905ceb98fe94d14a0c333665a1d3b97a6fbca0))
* when copying an ID, copy only the ID, not the &lt;span&gt; tag ([#790](https://github.com/leg100/otf/issues/790)) ([1eb4268](https://github.com/leg100/otf/commit/1eb4268ef05cd3e76c5031b700e1a12817968135))


### Miscellaneous

* bump templ ([dd8d027](https://github.com/leg100/otf/commit/dd8d027b6cdfae7a70afce8fef5ae22cc9361972))
* remove unused modd file watcher ([89c422d](https://github.com/leg100/otf/commit/89c422d9f4f81d7b4eb62cae84f5d2d01e23e34b))
* update algorithms ([a303feb](https://github.com/leg100/otf/commit/a303feb49eedd2c0f6afb0f6a96d820674e4ed31))
* update doc screenshots ([#796](https://github.com/leg100/otf/issues/796)) ([0677c07](https://github.com/leg100/otf/commit/0677c0770544f4abb431f077b5406872ed55e290))

## [0.3.23](https://github.com/leg100/otf/compare/v0.3.22...v0.3.23) (2025-05-26)


### Features

* add forgejo vcs provider ([#771](https://github.com/leg100/otf/issues/771)) ([dbb3c0a](https://github.com/leg100/otf/commit/dbb3c0a5221a3c5f78e5244a4fecc74116046a6b))
* **tests:** add browser tracing ([10fd15d](https://github.com/leg100/otf/commit/10fd15dc1a78ab51c4cf4fa651707b9ca9d0444f))


### Bug Fixes

* "Save variable set" button on the variableset edit page ([#778](https://github.com/leg100/otf/issues/778)) ([ed7a3d2](https://github.com/leg100/otf/commit/ed7a3d2d094198a304bd17fc705c8768db94d087))
* a bad background color on the oauth2 confirmation page ([#786](https://github.com/leg100/otf/issues/786)) ([09b3366](https://github.com/leg100/otf/commit/09b33667ecde38db970097f6e24dbef7c2841b89))
* add primary key to events table ([1cc6a21](https://github.com/leg100/otf/commit/1cc6a217197d7daf74bd0a00896db14a0f5763ab))
* drop organization event trigger ([869b82f](https://github.com/leg100/otf/commit/869b82f6d6056c6c9e9d6cbaaa44faf0d93766d3))
* event cleanup failing ([7a76322](https://github.com/leg100/otf/commit/7a76322d17f76de938fa344a96262fab3a5d7626))
* jobs belonging to terminated runners should not restart allocator ([3b8e111](https://github.com/leg100/otf/commit/3b8e11169f7bc912fefa918d87c62c55754cc244))
* permit events larger than 8000 bytes ([e528d9a](https://github.com/leg100/otf/commit/e528d9ac37519f14f09df8080c1089b356c31d2b))
* reduce websocket logging noise ([ed81049](https://github.com/leg100/otf/commit/ed81049f3e253bebbd7fad7e54c4fff1e0ed7c27))
* start subsystems in order to eliminate flaky tests ([63ab291](https://github.com/leg100/otf/commit/63ab291ccde9af7d4bffc31155cd03fa3115ba26))
* support earliest supported version of postgres (v13) ([649f499](https://github.com/leg100/otf/commit/649f499fd4f078c943d15427eb7c2da0a564cf7c))
* **tests:** broken runner registration test ([36ac6a7](https://github.com/leg100/otf/commit/36ac6a77c85f9295bc3675190b050cbceb145fbd))
* **tests:** disable flaky terraform login redirect check ([4f3f6c0](https://github.com/leg100/otf/commit/4f3f6c0c1e310a6d5f2142444c09dd09151fe5ca))
* **tests:** don't use singleton http handler ([eae2149](https://github.com/leg100/otf/commit/eae2149e2e2a90c4b044118c4b38a90764245a7f))
* **tests:** screenshots not being archived upon failure ([3ff867e](https://github.com/leg100/otf/commit/3ff867e119b83705d98c7d005f61b2f368838751))
* **tests:** suppress tls handshake errors ([d3c42ab](https://github.com/leg100/otf/commit/d3c42ab0d022265767c2e140111cfe6266f5866e))
* **tests:** use check instead of click for flaky tests ([e458abe](https://github.com/leg100/otf/commit/e458abe82b789d303ca9f9866756e3c2168ee205))
* **tests:** use workaround for flaky checkbox clicks ([d8dc659](https://github.com/leg100/otf/commit/d8dc659cb77bb0a015911aaddaaa85a84fd4d2af))
* wait for listener to start first ([b49fdb7](https://github.com/leg100/otf/commit/b49fdb73bd7ba970e65e917594254fd881a61218))
* wrong math operator causes logs to be prematurely chunked ([9fcd312](https://github.com/leg100/otf/commit/9fcd31286bf387b5923661945802906d5b9da3ad))


### Miscellaneous

* archive browser traces in github actions ([0310143](https://github.com/leg100/otf/commit/0310143e7a2e6e49e4929b63e572295a7964b9bc))
* debug GHA failing tests ([c7ef8ce](https://github.com/leg100/otf/commit/c7ef8ce261386e8dacae4619aa7fa165019a58f8))
* remove unused list components ([e7166ea](https://github.com/leg100/otf/commit/e7166ea1823eee4283e1a40b50f9c46c27875577))
* require log chunk attributes ([00d61c8](https://github.com/leg100/otf/commit/00d61c83a00ef6df28079b823b6271a6b7165888))
* revert go test make task ([6d0b2a3](https://github.com/leg100/otf/commit/6d0b2a3fbeebb41227877b67f2d3d98239ad7bb8))

## [0.3.22](https://github.com/leg100/otf/compare/v0.3.21...v0.3.22) (2025-05-18)


### Bug Fixes

* add `--allowed-origins` parameter ([#773](https://github.com/leg100/otf/issues/773)) ([b0c83ed](https://github.com/leg100/otf/commit/b0c83edf7db096fb839b405eebc999c256abee93))
* bad background colors on dark themes ([26fa6fc](https://github.com/leg100/otf/commit/26fa6fc3df2aff51d3bf8fc03b4c643f1dd7c193))


### Miscellaneous

* bump playwright version ([94a474e](https://github.com/leg100/otf/commit/94a474e8040dd65be8139570f50d52a6c6fffb15))
* check for differences in CI ([c634586](https://github.com/leg100/otf/commit/c63458677a0942b86963c739709bfeb94d55a568))
* **docs:** remove redundant flag --dev-mode ([ac1d401](https://github.com/leg100/otf/commit/ac1d40159b1a0b5a821d3d12952b1456ba3eea7e))
* **docs:** tweak a couple of pages ([2904f91](https://github.com/leg100/otf/commit/2904f910571729b4d844201febbec7a39ab21c93))
* remove redundant connection param ([#774](https://github.com/leg100/otf/issues/774)) ([5b784c8](https://github.com/leg100/otf/commit/5b784c833335a946aa1f91945626fc510dc7cf5f))
* remove unused watch func ([5bc63b2](https://github.com/leg100/otf/commit/5bc63b2f8d3aa152299ebde321ba8c2430af0207))
* use go tool for templ ([403974f](https://github.com/leg100/otf/commit/403974f5e7cb3d5a44b90b1b104ed76e9693667e))

## [0.3.21](https://github.com/leg100/otf/compare/v0.3.20...v0.3.21) (2025-05-04)


### Features

* new workspace now uses latest engine version ([bd40d22](https://github.com/leg100/otf/commit/bd40d22198f853171afcb2811bbeb22f727417ce))


### Miscellaneous

* remove working comments ([a51716e](https://github.com/leg100/otf/commit/a51716e2cbda281f5433963f8b54692f4f49c9cf))
* upgrade templ ([21a105e](https://github.com/leg100/otf/commit/21a105ec1c6027d10d8be81531f1463acc3fe3f2))

## [0.3.20](https://github.com/leg100/otf/compare/v0.3.19...v0.3.20) (2025-04-25)


### Features

* add support for opentofu ([#763](https://github.com/leg100/otf/issues/763)) ([c5e4742](https://github.com/leg100/otf/commit/c5e474208efec4464bf06383cb83d2321d7d2df3))
* **ui:** filter modules by provider ([8af8e56](https://github.com/leg100/otf/commit/8af8e56af846338177e6658bab3bb62efa33b50c))


### Bug Fixes

* **ui:** some things hard to see in dark mode ([c4aece4](https://github.com/leg100/otf/commit/c4aece40fab20939b6b7af078fb498237b3508f7))


### Miscellaneous

* bump vuln go-slug dep ([927c8b9](https://github.com/leg100/otf/commit/927c8b9a830210b771d958bce600779fd051ec35))
* document tofu support ([9b1fe1b](https://github.com/leg100/otf/commit/9b1fe1bff09ca3930ff41f43412756a6079fd805))
* more tofu tests ([#764](https://github.com/leg100/otf/issues/764)) ([5812449](https://github.com/leg100/otf/commit/581244958f01f0081ee433f467d75767f112f205))
* upgrade all deps ([299c32d](https://github.com/leg100/otf/commit/299c32dfb5bae3641e43ab79c2d79806efcc1410))

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
