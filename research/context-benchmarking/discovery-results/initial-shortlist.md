# Demon Docs Context-Injection Benchmark Shortlist

Generated: 2026-07-19T00:06:30.616005+00:00

> These scores estimate benchmark usefulness, not repository or maintainer quality. ‘Documentation gap’ means the repository may expose useful context-retrieval challenges.

## Recommended candidates

### 1. [crossplane-contrib/provider-sql](https://github.com/crossplane-contrib/provider-sql) — 88.3/100

An SQL provider for @crossplane

- Snapshot: `7addde329901acd5306946a1e4c1eb2f9d2f7876`
- Handwritten Go LOC: 31,784; packages: 58; Go files: 172; test files: 27
- Docs: 2 Markdown files; 0 under docs; README: 111 lines; architecture docs: 0
- History sampled: 80 merged PRs; issue-linked: 11; benchmark-sized: 19
- Scores: documentation gap 91.8; structure 90.0; testability 72.0; history 100.0; operations 80.0
- Strengths: large documentation gap, meaningful cross-package structure, strong automated-test surface, good historical benchmark-task inventory, multiple issue-linked merged PRs, database or persistence boundary
- Concerns: contains very large source files

Historical task examples:

- [PR #379](https://github.com/crossplane-contrib/provider-sql/pull/379) — fix(postgresql): make schema optional and conditionally required in DefaultPrivileges (10 files, +574/-28, issue refs: #378, #500)
- [PR #377](https://github.com/crossplane-contrib/provider-sql/pull/377) — fix: add missing ownerRef+ownerSelector field (13 files, +298/-0, issue refs: #372)
- [PR #300](https://github.com/crossplane-contrib/provider-sql/pull/300) — feat: Add MSSQL e2e integration test (8 files, +235/-1, issue refs: #500)
- [PR #290](https://github.com/crossplane-contrib/provider-sql/pull/290) — fix: Namespaced PostgreSQL extension did not use namespaced database reference (6 files, +90/-29, issue refs: #285)
- [PR #361](https://github.com/crossplane-contrib/provider-sql/pull/361) — feat(postgresql): support WITH INHERIT FALSE on role membership grants (PostgreSQL 16+) (11 files, +704/-22, issue refs: #359)

### 2. [shazow/wifitui](https://github.com/shazow/wifitui) — 83.9/100

Fast featureful friendly wifi terminal UI. 🛜✨

- Snapshot: `2026438f2a0d72679edd6f2731c654dd5dac616a`
- Handwritten Go LOC: 8,703; packages: 9; Go files: 49; test files: 18
- Docs: 1 Markdown files; 0 under docs; README: 109 lines; architecture docs: 0
- History sampled: 80 merged PRs; issue-linked: 9; benchmark-sized: 52
- Scores: documentation gap 99.4; structure 64.4; testability 72.0; history 97.0; operations 80.0
- Strengths: large documentation gap, meaningful cross-package structure, strong automated-test surface, good historical benchmark-task inventory, multiple issue-linked merged PRs, external integration boundaries
- Concerns: none identified mechanically

Historical task examples:

- [PR #174](https://github.com/shazow/wifitui/pull/174) — wifi/iwd: Fix iwd backend to use actual D-Bus API (5 files, +374/-201, issue refs: #171)
- [PR #178](https://github.com/shazow/wifitui/pull/178) — Default to empty theme when NO_COLOR is set (2 files, +24/-0, issue refs: #176)
- [PR #173](https://github.com/shazow/wifitui/pull/173) — Check NetworkManager availability, add backend fallback and tests (4 files, +108/-5, issue refs: #171)
- [PR #167](https://github.com/shazow/wifitui/pull/167) — list: (5 APs) -> 📡×5 (5 files, +67/-1, issue refs: #164)
- [PR #163](https://github.com/shazow/wifitui/pull/163) — TUI improvements: Width-aware List/Edit, fit to terminal (12 files, +398/-52, issue refs: #160)

### 3. [openstack-exporter/openstack-exporter](https://github.com/openstack-exporter/openstack-exporter) — 80.1/100

OpenStack Exporter for Prometheus

- Snapshot: `b16526f1ef95c4a5a338d16b867a73b6ea251dad`
- Handwritten Go LOC: 7,626; packages: 8; Go files: 63; test files: 30
- Docs: 3 Markdown files; 0 under docs; README: 1492 lines; architecture docs: 0
- History sampled: 80 merged PRs; issue-linked: 9; benchmark-sized: 41
- Scores: documentation gap 82.1; structure 67.3; testability 72.0; history 97.0; operations 80.0
- Strengths: large documentation gap, meaningful cross-package structure, strong automated-test surface, good historical benchmark-task inventory, multiple issue-linked merged PRs, external integration boundaries, database or persistence boundary
- Concerns: none identified mechanically

Historical task examples:

- [PR #413](https://github.com/openstack-exporter/openstack-exporter/pull/413) — Add feature to get password from Vault (3 files, +94/-1, issue refs: #412)
- [PR #378](https://github.com/openstack-exporter/openstack-exporter/pull/378) — Adding node retired exporting (7 files, +262/-121, issue refs: #377)
- [PR #373](https://github.com/openstack-exporter/openstack-exporter/pull/373) — Add domain_info metric / add tags  (comma separated string) to project_info metric. (2 files, +22/-12, issue refs: #230, #315)
- [PR #323](https://github.com/openstack-exporter/openstack-exporter/pull/323) — Resolve #289 | Add `floating_ips` label to port (3 files, +21/-5, issue refs: #289)
- [PR #541](https://github.com/openstack-exporter/openstack-exporter/pull/541) — fix: use GaugeValue for agent_state metrics (6 files, +6/-6, issue refs: #427)

### 4. [mercuretechnologies/expo-open-ota](https://github.com/mercuretechnologies/expo-open-ota) — 78.7/100

An open-source self-hosted custom updates server implementing the Expo Updates protocol, built for production. Supports cloud storage & CDN.

- Snapshot: `46f8caae7ad85afa3904662dc713ae03fd654cf4`
- Handwritten Go LOC: 9,722; packages: 26; Go files: 76; test files: 21
- Docs: 30 Markdown files; 0 under docs; README: 80 lines; architecture docs: 0
- History sampled: 50 merged PRs; issue-linked: 3; benchmark-sized: 27
- Scores: documentation gap 94.2; structure 74.4; testability 72.0; history 71.5; operations 80.0
- Strengths: large documentation gap, meaningful cross-package structure, strong automated-test surface, good historical benchmark-task inventory, external integration boundaries
- Concerns: none identified mechanically

Historical task examples:

- [PR #71](https://github.com/mercuretechnologies/expo-open-ota/pull/71) — Fix dashboard update count to include only valid updates(#63) (4 files, +33/-0, issue refs: #63)
- [PR #41](https://github.com/mercuretechnologies/expo-open-ota/pull/41) — fix: prevent cold-manifest 500s after OTA publish (8 files, +217/-12, issue refs: #19)
- [PR #60](https://github.com/mercuretechnologies/expo-open-ota/pull/60) — fix: cache expo auth to prevent ENHANCE_YOUR_CALM rate limit (#58) (1 files, +33/-0, issue refs: #58)
- [PR #80](https://github.com/mercuretechnologies/expo-open-ota/pull/80) — fix: run the runtime container as a non-root user (reworked #75) (3 files, +94/-1, issue refs: none)
- [PR #76](https://github.com/mercuretechnologies/expo-open-ota/pull/76) — fix: resolve pnpm projects to pnpm exec instead of the removed pnpx (4 files, +26/-12, issue refs: none)

### 5. [grafana/sobek](https://github.com/grafana/sobek) — 77.4/100

No repository description.

- Snapshot: `267a0e055bb478716bb054818686afa940d32f25`
- Handwritten Go LOC: 70,486; packages: 8; Go files: 124; test files: 47
- Docs: 3 Markdown files; 0 under docs; README: 370 lines; architecture docs: 0
- History sampled: 62 merged PRs; issue-linked: 6; benchmark-sized: 29
- Scores: documentation gap 82.1; structure 67.0; testability 75.0; history 83.5; operations 80.0
- Strengths: large documentation gap, meaningful cross-package structure, strong automated-test surface, good historical benchmark-task inventory, multiple issue-linked merged PRs, external integration boundaries
- Concerns: contains very large source files

Historical task examples:

- [PR #56](https://github.com/grafana/sobek/pull/56) — Update goja (12 files, +157/-34, issue refs: #19, #21)
- [PR #40](https://github.com/grafana/sobek/pull/40) — Updates from goja (12 files, +186/-48, issue refs: #16)
- [PR #12](https://github.com/grafana/sobek/pull/12) — Improve ambiguous import error and position fixes (3 files, +171/-86, issue refs: #11)
- [PR #120](https://github.com/grafana/sobek/pull/120) — Update goja (6 files, +870/-199, issue refs: #114)
- [PR #51](https://github.com/grafana/sobek/pull/51) — Fix errors in async module triggering twice (2 files, +41/-0, issue refs: none)

### 6. [containerd/nri](https://github.com/containerd/nri) — 75.3/100

Node Resource Interface

- Snapshot: `84180b63351d03b54a317a265ba01e3e9db19f24`
- Handwritten Go LOC: 19,729; packages: 30; Go files: 100; test files: 18
- Docs: 17 Markdown files; 1 under docs; README: 559 lines; architecture docs: 0
- History sampled: 80 merged PRs; issue-linked: 3; benchmark-sized: 27
- Scores: documentation gap 71.2; structure 78.0; testability 72.0; history 79.0; operations 80.0
- Strengths: large documentation gap, meaningful cross-package structure, strong automated-test surface, good historical benchmark-task inventory, external integration boundaries
- Concerns: contains very large source files

Historical task examples:

- [PR #264](https://github.com/containerd/nri/pull/264) — api: fix OCI hook ownership tracking. (5 files, +72/-11, issue refs: none)
- [PR #261](https://github.com/containerd/nri/pull/261) — api,adaptation: fix sysctl adjustment collection, add unit test. (3 files, +23/-0, issue refs: none)
- [PR #252](https://github.com/containerd/nri/pull/252) — update wazero/wazero version to v1.10.1 (20 files, +30/-30, issue refs: #12665)

### 7. [kubernetes-sigs/dra-driver-cpu](https://github.com/kubernetes-sigs/dra-driver-cpu) — 73.2/100

CPU DRA Driver

- Snapshot: `1d75acd83ba37d26932723e596ce58b013c3c71e`
- Handwritten Go LOC: 15,146; packages: 28; Go files: 86; test files: 38
- Docs: 14 Markdown files; 4 under docs; README: 808 lines; architecture docs: 1
- History sampled: 80 merged PRs; issue-linked: 6; benchmark-sized: 43
- Scores: documentation gap 46.5; structure 78.0; testability 82.0; history 88.0; operations 80.0
- Strengths: meaningful cross-package structure, strong automated-test surface, good historical benchmark-task inventory, multiple issue-linked merged PRs, external integration boundaries
- Concerns: contains very large source files

Historical task examples:

- [PR #170](https://github.com/kubernetes-sigs/dra-driver-cpu/pull/170) — Fix PrepareResourceClaims allocation commit ordering (2 files, +129/-3, issue refs: #169)
- [PR #180](https://github.com/kubernetes-sigs/dra-driver-cpu/pull/180) — Fix UnprepareResourceClaims CDI removal ordering (2 files, +61/-2, issue refs: #178)
- [PR #190](https://github.com/kubernetes-sigs/dra-driver-cpu/pull/190) — Fail closed on malformed DRA_CPUSET env (6 files, +41/-20, issue refs: #185)
- [PR #177](https://github.com/kubernetes-sigs/dra-driver-cpu/pull/177) — device: move IsPCIeRootName to arch-neutral source so gen-pcie-testdata builds on all arches (4 files, +107/-107, issue refs: #176)
- [PR #171](https://github.com/kubernetes-sigs/dra-driver-cpu/pull/171) — Initialize device lookup maps before publishing (3 files, +335/-105, issue refs: #168)

### 8. [bots-go-framework/bots-fw](https://github.com/bots-go-framework/bots-fw) — 72.9/100

Golang framework to build multilingual bots for messengers (Telegram, FB Messenger, Skype, Line, Kik, WeChat) hosted on AppEngine, Amazon, Azure, Heroku or standalone

- Snapshot: `c46f449a7e2e5be1c44291b21d0722d41a733456`
- Handwritten Go LOC: 13,668; packages: 11; Go files: 164; test files: 51
- Docs: 7 Markdown files; 0 under docs; README: 116 lines; architecture docs: 0
- History sampled: 42 merged PRs; issue-linked: 1; benchmark-sized: 10
- Scores: documentation gap 96.8; structure 76.0; testability 62.0; history 52.5; operations 80.0
- Strengths: large documentation gap, meaningful cross-package structure, strong automated-test surface, external integration boundaries
- Concerns: contains very large source files

Historical task examples:

- [PR #71](https://github.com/bots-go-framework/bots-fw/pull/71) — Fix staticcheck issues in golangci-lint (4 files, +3/-19, issue refs: #3)
- [PR #81](https://github.com/bots-go-framework/bots-fw/pull/81) — fix(botsfw)!: scope bot settings lookup by platform (4 files, +222/-25, issue refs: none)
- [PR #78](https://github.com/bots-go-framework/bots-fw/pull/78) — test: cover botinput and botmsg packages (0% → 100%) (2 files, +85/-0, issue refs: none)
- [PR #74](https://github.com/bots-go-framework/bots-fw/pull/74) — refactor(botsdal): migrate CreatePlatformUserRecord onto dal.InsertRecordWithDataAndID (4 files, +13/-13, issue refs: none)
- [PR #73](https://github.com/bots-go-framework/bots-fw/pull/73) — refactor(botsdal): migrate GetBotChat/GetPlatformUser onto dal.GetRecordWithIDIntoData (5 files, +32/-26, issue refs: none)

### 9. [open-telemetry/opentelemetry-lambda](https://github.com/open-telemetry/opentelemetry-lambda) — 72.7/100

Create your own Lambda Layer in each OTel language using this starter code. Add the Lambda Layer to your Lambda Function to get tracing with OpenTelemetry.

- Snapshot: `58d24723654c5f15984c62ef6d12b93f0b2fe1d7`
- Handwritten Go LOC: 7,871; packages: 21; Go files: 77; test files: 20
- Docs: 20 Markdown files; 1 under docs; README: 158 lines; architecture docs: 2
- History sampled: 80 merged PRs; issue-linked: 4; benchmark-sized: 26
- Scores: documentation gap 63.2; structure 72.7; testability 72.0; history 82.0; operations 80.0
- Strengths: meaningful cross-package structure, strong automated-test surface, good historical benchmark-task inventory, external integration boundaries
- Concerns: contains very large source files

Historical task examples:

- [PR #2288](https://github.com/open-telemetry/opentelemetry-lambda/pull/2288) — fix(collector): use Telemetry API event time as metric timestamp (2 files, +71/-1, issue refs: #2263)
- [PR #2265](https://github.com/open-telemetry/opentelemetry-lambda/pull/2265) — feat(collector): add transform processor to default collector build (7 files, +94/-42, issue refs: #2218)
- [PR #2331](https://github.com/open-telemetry/opentelemetry-lambda/pull/2331) — build(deps): bump the opentelemetry-deps-collector group across 6 directories with 49 updates (12 files, +996/-986, issue refs: #153)

### 10. [GoogleCloudPlatform/guest-agent](https://github.com/GoogleCloudPlatform/guest-agent) — 71.1/100

No repository description.

- Snapshot: `a1ffdafd6424762f1b7bac8ad7bba67c9829c937`
- Handwritten Go LOC: 17,972; packages: 22; Go files: 113; test files: 36
- Docs: 5 Markdown files; 0 under docs; README: 383 lines; architecture docs: 1
- History sampled: 80 merged PRs; issue-linked: 0; benchmark-sized: 40
- Scores: documentation gap 71.9; structure 78.0; testability 62.0; history 70.0; operations 80.0
- Strengths: large documentation gap, meaningful cross-package structure, strong automated-test surface, good historical benchmark-task inventory, external integration boundaries
- Concerns: no issue-linked PRs found in sample

Historical task examples:

- [PR #593](https://github.com/GoogleCloudPlatform/guest-agent/pull/593) — Add packaging changes for locally bundled extensions feature support (4 files, +13/-6, issue refs: none)
- [PR #515](https://github.com/GoogleCloudPlatform/guest-agent/pull/515) — Fix failing builds by updating the dependencies  (2 files, +8/-7, issue refs: none)

## Selection procedure

Before adopting a repository, manually validate its top historical tasks:

1. Check out the parent commit of each accepted PR.
2. Confirm the repository builds and its baseline tests pass at that commit.
3. Recover the original issue text without reading the accepted patch.
4. Confirm the task requires repository discovery rather than isolated syntax work.
5. Confirm demon-docs can select useful context without embedding the expected implementation.
6. Pin the base commit, issue text, verification command, and accepted PR in the benchmark manifest.

## Rejected candidates

- [kelos-dev/kelos](https://github.com/kelos-dev/kelos) — 46.3/100: too large for the initial benchmark
- [markhuangai/dense-mem](https://github.com/markhuangai/dense-mem) — 44.0/100: too large for the initial benchmark
- [escalier-lang/escalier](https://github.com/escalier-lang/escalier) — 33.5/100: too large for the initial benchmark
- [dynatrace-oss/dtctl](https://github.com/dynatrace-oss/dtctl) — 15.3/100: too large for the initial benchmark
- [CrowdStrike/terraform-provider-crowdstrike](https://github.com/CrowdStrike/terraform-provider-crowdstrike) — 11.6/100: too large for the initial benchmark
- [defilantech/LLMKube](https://github.com/defilantech/LLMKube) — 9.4/100: too large for the initial benchmark
- [291-Group/LAN-Orangutan](https://github.com/291-Group/LAN-Orangutan) — 0.0/100: insufficient test surface
- [dannybouwers/trala](https://github.com/dannybouwers/trala) — 0.0/100: insufficient test surface
- [go-go-golems/geppetto](https://github.com/go-go-golems/geppetto) — 0.0/100: clone failed, too little handwritten Go code, too few package boundaries, insufficient test surface
- [Kong/kongctl](https://github.com/Kong/kongctl) — 0.0/100: clone failed, too little handwritten Go code, too few package boundaries, insufficient test surface
