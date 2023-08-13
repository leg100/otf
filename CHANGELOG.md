# Changelog

## [0.1.8](https://github.com/leg100/otf/compare/v0.1.7...v0.1.8) (2023-08-13)


### Features

* allow terraform apply on connected workspace ([#564](https://github.com/leg100/otf/issues/564)) ([6f90a9c](https://github.com/leg100/otf/commit/6f90a9c0b817f6cb846df1a487606a52a963a7b4)), closes [#231](https://github.com/leg100/otf/issues/231)
* **ui:** add icon in run widget to show source of run ([#563](https://github.com/leg100/otf/issues/563)) ([2e7a0bd](https://github.com/leg100/otf/commit/2e7a0bd71b99556360070337b8e9baad3a021aad))


### Bug Fixes

* cleanup after extracting repo tarball ([bf4758b](https://github.com/leg100/otf/commit/bf4758bead52e6c3bf1e47d1dfe06ebcff0a26a8))
* don't scrub included state output sensitive values ([478e314](https://github.com/leg100/otf/commit/478e314a687f722125653d6aa1010b8c3bf2b060))
* linux/arm64 support ([#562](https://github.com/leg100/otf/issues/562)) ([01a2112](https://github.com/leg100/otf/commit/01a211240e4dca4d18e02d49e3f9d6190754510c)), closes [#311](https://github.com/leg100/otf/issues/311)
* otfd compose healthcheck: curl not installed ([9f52021](https://github.com/leg100/otf/commit/9f52021d7515b4736547d8e978dcabd756d5c263))
* qemu needed for building multi-arch images ([1aa8cf8](https://github.com/leg100/otf/commit/1aa8cf87bdb6954a01abaaef8ae26be8a0dbfa7d))
* retry run should use existing run properties ([49303ec](https://github.com/leg100/otf/commit/49303ecf42edb106def169ddf68f66df7558b741))
* **tests:** hard link fails when /tmp is separate partition ([cfc7aaa](https://github.com/leg100/otf/commit/cfc7aaa80d31c9b7a0b4e461d13bb09fea9f87bc))
* **ui:** workspace description missing after update ([a579b40](https://github.com/leg100/otf/commit/a579b40ffac4aa2f021afc9417ad9bb8b3b2cc49))
* use png instead of svg for font-based icons ([eae0588](https://github.com/leg100/otf/commit/eae0588d7b6de5fb4cb2e5c2ad7fa483360f308c))


### Miscellaneous

* bump squid version ([7ce3238](https://github.com/leg100/otf/commit/7ce3238f7af3755c317a28690b1dbd8e7efed2b9))
* go 1.21 ([#566](https://github.com/leg100/otf/issues/566)) ([06c13b2](https://github.com/leg100/otf/commit/06c13b250b183c12e0486e69cac2aee1c52b7ed5))
* **master:** release 0.1.8 ([#561](https://github.com/leg100/otf/issues/561)) ([9c6d0f6](https://github.com/leg100/otf/commit/9c6d0f65628ee5786a815957df8d0df339ba9b47))
* remove unused cloud team and org sync code ([4e1817d](https://github.com/leg100/otf/commit/4e1817dbbd21093c835e84d921606dd2ae46f871))
* removed unused ca.pem ([799ed25](https://github.com/leg100/otf/commit/799ed25565c155c616e90533b3172bc22f916f6b))
* skip api tests if env vars not present ([5b88474](https://github.com/leg100/otf/commit/5b88474d3c4813897f39f3b463d013cbc831ad64))
* **ui:** make tags less bulbous ([df1645d](https://github.com/leg100/otf/commit/df1645d8de9d4ce021d93e58f03d27911494649f))
* **ui:** pad out buttons on consent page ([1c290e9](https://github.com/leg100/otf/commit/1c290e93248d9620d54c41eb3681065929069cde))
* update docs ([364d183](https://github.com/leg100/otf/commit/364d183dd8635eb0ce73b1e65666475ab0a039ea))
* validate resource names ([c7596fe](https://github.com/leg100/otf/commit/c7596febc1018a546ec2c990ae5087ae297df8c0))

## [0.1.8](https://github.com/leg100/otf/compare/v0.1.7...v0.1.8) (2023-08-13)


### Features

* allow terraform apply on connected workspace ([#564](https://github.com/leg100/otf/issues/564)) ([6f90a9c](https://github.com/leg100/otf/commit/6f90a9c0b817f6cb846df1a487606a52a963a7b4)), closes [#231](https://github.com/leg100/otf/issues/231)
* **ui:** add icon in run widget to show source of run ([#563](https://github.com/leg100/otf/issues/563)) ([2e7a0bd](https://github.com/leg100/otf/commit/2e7a0bd71b99556360070337b8e9baad3a021aad))


### Bug Fixes

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
