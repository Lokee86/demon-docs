# Candidate Historical Task Inventory

Generated: 2026-07-19T00:09:11.906258+00:00

> Base commits are the pull request base SHAs reported by GitHub. Before using a task, verify that the issue was open against that exact state and run baseline tests at the pinned commit.

## crossplane-contrib/provider-sql

### [PR #379](https://github.com/crossplane-contrib/provider-sql/pull/379) — fix(postgresql): make schema optional and conditionally required in DefaultPrivileges

- Task quality: 90/100
- Base commit: `9665c434424129d54f2bc265f5921fa3dab29581`
- Patch: 10 files, +574/-28
- Test paths: pkg/controller/cluster/postgresql/default_privileges/reconciler_test.go, pkg/controller/namespaced/postgresql/default_privileges/reconciler_test.go
- Strengths: has original issue context, issue contains substantive reproduction or requirements, bounded multi-file change, accepted patch changes tests, requires broader repository context
- Concerns: none

Linked issues:

- [#378](https://github.com/crossplane-contrib/provider-sql/issues/378) — DefaultPrivileges: objectType: schema generates invalid SQL: ### What happened? <!-- Please let us know what behaviour you expected and how Crossplane diverged from that behaviour. --> Creating a `DefaultPrivileges` resource with `objectType: schema` always fails. The provider generates an `ALTER DEFAULT PRIVILEGES ...
- Issue #500: retrieval failed

Changed paths:

- `apis/cluster/postgresql/v1alpha1/default_privileges_types.go`
- `apis/namespaced/postgresql/v1alpha1/default_privileges_types.go`
- `examples/cluster/postgresql/defaultprivileges.yaml`
- `examples/namespaced/postgresql/defaultprivileges.yaml`
- `package/crds/postgresql.sql.crossplane.io_defaultprivileges.yaml`
- `package/crds/postgresql.sql.m.crossplane.io_defaultprivileges.yaml`
- `pkg/controller/cluster/postgresql/default_privileges/reconciler.go`
- `pkg/controller/cluster/postgresql/default_privileges/reconciler_test.go`
- `pkg/controller/namespaced/postgresql/default_privileges/reconciler.go`
- `pkg/controller/namespaced/postgresql/default_privileges/reconciler_test.go`

### [PR #290](https://github.com/crossplane-contrib/provider-sql/pull/290) — fix: Namespaced PostgreSQL extension did not use namespaced database reference

- Task quality: 80/100
- Base commit: `57020959ab138fb2a0ebe0fe16e466855bf52b66`
- Patch: 6 files, +90/-29
- Test paths: none detected
- Strengths: has original issue context, issue contains substantive reproduction or requirements, bounded multi-file change, reviewable implementation size, requires broader repository context
- Concerns: accepted patch has no obvious test change, generated-file churn may dominate the patch

Linked issues:

- [#285](https://github.com/crossplane-contrib/provider-sql/issues/285) — postgresql: Namespaced Extensions are not properly resolving databaseRef: ### What happened? During testing of a namespaced Extension I was unable to get databaseRef to function as it did under a clustered Extension. After a little digging it seems the namespaced Extension is not using the namespaced Reference [here](https://github.

Changed paths:

- `apis/namespaced/postgresql/v1alpha1/extension_types.go`
- `apis/namespaced/postgresql/v1alpha1/zz_generated.deepcopy.go`
- `apis/namespaced/postgresql/v1alpha1/zz_generated.resolvers.go`
- `cluster/local/postgresdb_functions.sh`
- `examples/namespaced/postgresql/extension.yaml`
- `package/crds/postgresql.sql.m.crossplane.io_extensions.yaml`

### [PR #361](https://github.com/crossplane-contrib/provider-sql/pull/361) — feat(postgresql): support WITH INHERIT FALSE on role membership grants (PostgreSQL 16+)

- Task quality: 80/100
- Base commit: `b0fd0dad529c6629be01792bdd765d4660d086d0`
- Patch: 11 files, +704/-22
- Test paths: pkg/controller/cluster/postgresql/grant/reconciler_test.go, pkg/controller/namespaced/postgresql/grant/reconciler_test.go
- Strengths: has original issue context, issue contains substantive reproduction or requirements, bounded multi-file change, accepted patch changes tests, requires broader repository context
- Concerns: generated-file churn may dominate the patch

Linked issues:

- [#359](https://github.com/crossplane-contrib/provider-sql/issues/359) — feat(postgresql): support WITH INHERIT FALSE on role membership grants (PostgreSQL 16+): ## Summary The PostgreSQL `Grant` managed resource currently supports `withOption` for privilege grants (`WITH GRANT OPTION`), but has no way to express `WITH INHERIT FALSE` on role membership grants — a feature introduced in PostgreSQL 16. ## Background Postg

Changed paths:

- `apis/cluster/postgresql/v1alpha1/grant_types.go`
- `apis/cluster/postgresql/v1alpha1/zz_generated.deepcopy.go`
- `apis/namespaced/postgresql/v1alpha1/grant_types.go`
- `apis/namespaced/postgresql/v1alpha1/zz_generated.deepcopy.go`
- `examples/cluster/postgresql/grant-with-inherit-false.yaml`
- `package/crds/postgresql.sql.crossplane.io_grants.yaml`
- `package/crds/postgresql.sql.m.crossplane.io_grants.yaml`
- `pkg/controller/cluster/postgresql/grant/reconciler.go`
- `pkg/controller/cluster/postgresql/grant/reconciler_test.go`
- `pkg/controller/namespaced/postgresql/grant/reconciler.go`
- `pkg/controller/namespaced/postgresql/grant/reconciler_test.go`

### [PR #377](https://github.com/crossplane-contrib/provider-sql/pull/377) — fix: add missing ownerRef+ownerSelector field

- Task quality: 60/100
- Base commit: `a76bde9446e13719cfee548bc2aefd47230c5550`
- Patch: 13 files, +298/-0
- Test paths: none detected
- Strengths: has original issue context, issue contains substantive reproduction or requirements, reviewable implementation size, requires broader repository context
- Concerns: accepted patch has no obvious test change, generated-file churn may dominate the patch

Linked issues:

- [#372](https://github.com/crossplane-contrib/provider-sql/issues/372) — Support ownerRef for postgres database: ### What problem are you facing? Currently we expect the role name to be set in the `forProvider.owner` for a `Database.postgresql.sql` resource. It is not possible to reference by label, using the well-known *NameSelector or *NameRef patterns. ### How could C

Changed paths:

- `apis/cluster/postgresql/v1alpha1/database_types.go`
- `apis/cluster/postgresql/v1alpha1/zz_generated.deepcopy.go`
- `apis/cluster/postgresql/v1alpha1/zz_generated.resolvers.go`
- `apis/namespaced/postgresql/v1alpha1/database_types.go`
- `apis/namespaced/postgresql/v1alpha1/zz_generated.deepcopy.go`
- `apis/namespaced/postgresql/v1alpha1/zz_generated.resolvers.go`
- `cluster/local/postgresdb_functions.sh`
- `examples/cluster/postgresql/database.yaml`
- `examples/cluster/postgresql/role.yaml`
- `examples/namespaced/postgresql/database.yaml`
- `examples/namespaced/postgresql/role.yaml`
- `package/crds/postgresql.sql.crossplane.io_databases.yaml`
- `package/crds/postgresql.sql.m.crossplane.io_databases.yaml`

## mercuretechnologies/expo-open-ota

### [PR #71](https://github.com/mercuretechnologies/expo-open-ota/pull/71) — Fix dashboard update count to include only valid updates(#63)

- Task quality: 100/100
- Base commit: `ce45f2d3fdf738888f5b435fe5e52b240fa943f7`
- Patch: 4 files, +33/-0
- Test paths: test/dashboard_test.go
- Strengths: has original issue context, issue contains substantive reproduction or requirements, bounded multi-file change, reviewable implementation size, accepted patch changes tests, requires broader repository context
- Concerns: none

Linked issues:

- [#63](https://github.com/mercuretechnologies/expo-open-ota/issues/63) — Dashboard update count may include incomplete updates after failed publish attempts: Hi! I’m currently doing a POC for my company with this project in a local Docker-based setup, and I noticed what looks like a small dashboard inconsistency. ## Summary I found a possible dashboard inconsistency after several failed publish attempts in a local

Changed paths:

- `internal/bucket/gcsBucket.go`
- `internal/bucket/localBucket.go`
- `internal/bucket/s3Bucket.go`
- `test/dashboard_test.go`

### [PR #41](https://github.com/mercuretechnologies/expo-open-ota/pull/41) — fix: prevent cold-manifest 500s after OTA publish

- Task quality: 100/100
- Base commit: `223108b95edc22f8c1430d563a0f7aed63edcf50`
- Patch: 8 files, +217/-12
- Test paths: test/channel_mapping_cache_test.go, test/manifest_test.go, test/requestUpload_test.go
- Strengths: has original issue context, issue contains substantive reproduction or requirements, bounded multi-file change, reviewable implementation size, accepted patch changes tests, requires broader repository context
- Concerns: none

Linked issues:

- [#19](https://github.com/mercuretechnologies/expo-open-ota/issues/19) — /manifest timing out (>9s) after releasing an update: Hello, we've recently started using this package and when there are no updates released or if user has latest version, /manifest API responds in 150ms on average. However right after releasing an update, /manifest starts timing out (takes longer than 10s) Is t

Changed paths:

- `internal/handlers/dashboard_handler.go`
- `internal/handlers/upload_handler.go`
- `internal/services/expo.go`
- `internal/update/prewarm.go`
- `internal/update/updates.go`
- `test/channel_mapping_cache_test.go`
- `test/manifest_test.go`
- `test/requestUpload_test.go`

### [PR #60](https://github.com/mercuretechnologies/expo-open-ota/pull/60) — fix: cache expo auth to prevent ENHANCE_YOUR_CALM rate limit (#58)

- Task quality: 68/100
- Base commit: `60af6804573de2b94a14ca265b9a1fbf324c0b2e`
- Patch: 1 files, +33/-0
- Test paths: none detected
- Strengths: has original issue context, issue contains substantive reproduction or requirements, reviewable implementation size
- Concerns: mostly localized task, accepted patch has no obvious test change

Linked issues:

- [#58](https://github.com/mercuretechnologies/expo-open-ota/issues/58) — ENHANCE_YOUR_CALM rate limit from api.expo.dev on every /uploadLocalFile request due to missing auth caching: **Problem** When publishing updates with a large number of assets using eoas publish, the server hits Expo's API rate limit (ENHANCE_YOUR_CALM) during file uploads, causing all uploads to fail with 401. **Error** Error validating expo auth: Post "https://api.e

Changed paths:

- `internal/services/expo.go`

## openstack-exporter/openstack-exporter

### [PR #378](https://github.com/openstack-exporter/openstack-exporter/pull/378) — Adding node retired exporting

- Task quality: 100/100
- Base commit: `add5ae7c707ee08006720a7a2cd914638a7c3cee`
- Patch: 7 files, +262/-121
- Test paths: exporters/ironic_test.go
- Strengths: has original issue context, issue contains substantive reproduction or requirements, bounded multi-file change, reviewable implementation size, accepted patch changes tests, requires broader repository context
- Concerns: none

Linked issues:

- [#377](https://github.com/openstack-exporter/openstack-exporter/issues/377) — Support node.Retired and node.RetiredReason: Issue that needs to be addressed first in gophercloud is here with an attached PR: https://github.com/gophercloud/gophercloud/pull/3136 The ironic code: https://github.com/openstack/ironic/blob/1e52143f07266b53ad47d49a5be04cbf85cccc8b/ironic/objects/node.py#L1

Changed paths:

- `exporters/exporter.go`
- `exporters/fixtures/ironic_nodes.json`
- `exporters/ironic.go`
- `exporters/ironic_test.go`
- `exporters/utils.go`
- `go.mod`
- `go.sum`

### [PR #323](https://github.com/openstack-exporter/openstack-exporter/pull/323) — Resolve #289 | Add `floating_ips` label to port

- Task quality: 100/100
- Base commit: `132c31c6d17e68b138e559a1ef821f885bfaed8f`
- Patch: 3 files, +21/-5
- Test paths: exporters/neutron_test.go
- Strengths: has original issue context, issue contains substantive reproduction or requirements, bounded multi-file change, reviewable implementation size, accepted patch changes tests
- Concerns: none

Linked issues:

- [#289](https://github.com/openstack-exporter/openstack-exporter/issues/289) — [Feature Request] Add IP Address Label To Neutron Port: Hello. I hope that we have fixed_ip_address label for neutron port metric, we really need it. Thank you.

Changed paths:

- `exporters/fixtures/neutron_ports.json`
- `exporters/neutron.go`
- `exporters/neutron_test.go`

### [PR #541](https://github.com/openstack-exporter/openstack-exporter/pull/541) — fix: use GaugeValue for agent_state metrics

- Task quality: 90/100
- Base commit: `8da53eedc88e1c59433ce9f3aa3011bb1c36173d`
- Patch: 6 files, +6/-6
- Test paths: exporters/cinder_test.go, exporters/neutron_test.go, exporters/nova_test.go
- Strengths: has original issue context, issue contains substantive reproduction or requirements, bounded multi-file change, accepted patch changes tests, requires broader repository context
- Concerns: none

Linked issues:

- [#427](https://github.com/openstack-exporter/openstack-exporter/issues/427) — There are some metrics defined `Counter` should be `Gauge`: Within the metrics that represent instantaneous states and are not monotonically increasing, there are some defined as Counters. Since these metrics are annotated as TYPE: Counter in the exporter's output, certain scrape implementations automatically append su

Changed paths:

- `exporters/cinder.go`
- `exporters/cinder_test.go`
- `exporters/neutron.go`
- `exporters/neutron_test.go`
- `exporters/nova.go`
- `exporters/nova_test.go`

### [PR #413](https://github.com/openstack-exporter/openstack-exporter/pull/413) — Add feature to get password from Vault

- Task quality: 80/100
- Base commit: `6b5671234b4fbc6da47512b9623f42a1d97a13df`
- Patch: 3 files, +94/-1
- Test paths: none detected
- Strengths: has original issue context, bounded multi-file change, reviewable implementation size, requires broader repository context
- Concerns: accepted patch has no obvious test change

Linked issues:

- [#412](https://github.com/openstack-exporter/openstack-exporter/issues/412) — Add feature to get  password from Vault, instead of config file.: It would be nice to implement security feature to get some credentials from Vault, e.g. password.

Changed paths:

- `go.mod`
- `go.sum`
- `main.go`

## shazow/wifitui

### [PR #174](https://github.com/shazow/wifitui/pull/174) — wifi/iwd: Fix iwd backend to use actual D-Bus API

- Task quality: 100/100
- Base commit: `8a244a4ba99253ab63173e3d7c37479f85027e79`
- Patch: 5 files, +374/-201
- Test paths: backend_linux_test.go, wifi/networkmanager/networkmanager_test.go
- Strengths: has original issue context, issue contains substantive reproduction or requirements, bounded multi-file change, reviewable implementation size, accepted patch changes tests, requires broader repository context
- Concerns: none

Linked issues:

- [#171](https://github.com/shazow/wifitui/issues/171) — Error: The name is not activatable: Trying on Arch Linux with Omarchy. Installed with `eget shazow/wifitui` and with `brew install wifitui`. Same result: Both `wifitui` and `wifitui list` return "Error: The name is not activatable"

Changed paths:

- `backend_linux.go`
- `backend_linux_test.go`
- `wifi/iwd/iwd.go`
- `wifi/networkmanager/networkmanager.go`
- `wifi/networkmanager/networkmanager_test.go`

### [PR #173](https://github.com/shazow/wifitui/pull/173) — Check NetworkManager availability, add backend fallback and tests

- Task quality: 100/100
- Base commit: `8a244a4ba99253ab63173e3d7c37479f85027e79`
- Patch: 4 files, +108/-5
- Test paths: backend_linux_test.go, wifi/networkmanager/networkmanager_test.go
- Strengths: has original issue context, issue contains substantive reproduction or requirements, bounded multi-file change, reviewable implementation size, accepted patch changes tests, requires broader repository context
- Concerns: none

Linked issues:

- [#171](https://github.com/shazow/wifitui/issues/171) — Error: The name is not activatable: Trying on Arch Linux with Omarchy. Installed with `eget shazow/wifitui` and with `brew install wifitui`. Same result: Both `wifitui` and `wifitui list` return "Error: The name is not activatable"

Changed paths:

- `backend_linux.go`
- `backend_linux_test.go`
- `wifi/networkmanager/networkmanager.go`
- `wifi/networkmanager/networkmanager_test.go`

### [PR #163](https://github.com/shazow/wifitui/pull/163) — TUI improvements: Width-aware List/Edit, fit to terminal

- Task quality: 100/100
- Base commit: `4583f965beaac68ed8de4cdfebd614645fcbac8a`
- Patch: 12 files, +398/-52
- Test paths: internal/tui/edit_test.go, internal/tui/focus_test.go, internal/tui/list_layout_test.go, internal/tui/list_test.go, internal/tui/tui_test.go, internal/tui/window_test.go
- Strengths: has original issue context, issue contains substantive reproduction or requirements, bounded multi-file change, reviewable implementation size, accepted patch changes tests, requires broader repository context
- Concerns: none

Linked issues:

- [#160](https://github.com/shazow/wifitui/issues/160) — ui: Span border full window width: Right now the window border shrinks to fit the content rather than the full width of the window. Ideally we want to span the full window, and we can also adjust the columns dynamically to take advantage of the increased width when available. There's a WIP PR t

Changed paths:

- `internal/tui/edit.go`
- `internal/tui/edit_test.go`
- `internal/tui/focus_test.go`
- `internal/tui/list.go`
- `internal/tui/list_layout_test.go`
- `internal/tui/list_test.go`
- `internal/tui/scanschedule.go`
- `internal/tui/stack.go`
- `internal/tui/tui.go`
- `internal/tui/tui_test.go`
- `internal/tui/window.go`
- `internal/tui/window_test.go`

### [PR #167](https://github.com/shazow/wifitui/pull/167) — list: (5 APs) -> 📡×5

- Task quality: 100/100
- Base commit: `86e3912f192617d5cbbc001f0ca059f710ddbe3d`
- Patch: 5 files, +67/-1
- Test paths: internal/tui/list_test.go, internal/tui/theme_test.go
- Strengths: has original issue context, issue contains substantive reproduction or requirements, bounded multi-file change, reviewable implementation size, accepted patch changes tests, requires broader repository context
- Concerns: none

Linked issues:

- [#164](https://github.com/shazow/wifitui/issues/164) — ui: Improve list view "(3 APs)" annotation: I kinda don't love seeing things like "(3 APs)" in the list view. <img width="50%" height="50%" alt="Image" src="https://github.com/user-attachments/assets/2ecce519-5a37-44fd-8ded-b6842d538a2a" /> Some ideas of how to improve them: 1. Emojis? Max out on three

Changed paths:

- `internal/tui/list.go`
- `internal/tui/list_test.go`
- `internal/tui/theme.go`
- `internal/tui/theme_test.go`
- `theme.toml`

### [PR #178](https://github.com/shazow/wifitui/pull/178) — Default to empty theme when NO_COLOR is set

- Task quality: 90/100
- Base commit: `111f53dee103724c0bbafd155cdcb51f8ab2a731`
- Patch: 2 files, +24/-0
- Test paths: none detected
- Strengths: has original issue context, issue contains substantive reproduction or requirements, bounded multi-file change, reviewable implementation size, requires broader repository context
- Concerns: accepted patch has no obvious test change

Linked issues:

- [#176](https://github.com/shazow/wifitui/issues/176) — feature: NO_COLOR support : Hi, I’d like to suggest adding support for the NO_COLOR standard (https://no-color.org/). When the NO_COLOR environment variable is set, the application would disable colored output. This is a widely adopted convention across CLI tools and helps improve access

Changed paths:

- `internal/tui/theme.go`
- `main.go`
