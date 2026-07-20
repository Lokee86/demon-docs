---
author: brian
created: "2026-07-19"
document_id: 019f7d55-31e4-719a-b726-af194111f3a1
document_type: general
policy_exempt: false
summary: This document defines the implemented .ddocs object repository, record namespace, deterministic sharding and codec, root publication, transaction semantics, concurrency control, corruption behavior, and ownership boundaries.
---
# Private Object Repository

Parent index: [Architecture](./INDEX.md)

## Purpose

This document defines the implemented `.ddocs` object repository, record namespace, deterministic sharding and codec, root publication, transaction semantics, concurrency control, corruption behavior, and ownership boundaries.

## Overview

Demon Docs stores durable reconciliation state in a private bare Git object database rooted at `.ddocs/`.

The object repository is a small transactional key/value layer built on Git blobs, trees, and one state reference:

```text
refs/ddocs/state -> root Git tree
                    -> up to 16 shard blobs
                       -> named byte records
```

It does not create a commit for every state change. A successful transaction writes any changed shard blobs, writes a new root tree, and compare-and-set updates `refs/ddocs/state` from the transaction's original root to the new root.

Consumers such as link reconciliation and reverse indexes define record schemas and meanings. `internal/ddrepo` owns storage integrity and publication, not domain interpretation.

## Primary ownership

The private object repository owns:

- creating or opening the bare Git repository below `.ddocs/`;
- initializing `refs/ddocs/state` to an empty root tree;
- validating record names;
- deterministically assigning records to one of 16 shards;
- deterministic binary encoding and decoding of shard contents;
- lazy transaction reads;
- copy-safe record values;
- write, delete, and prefix-name operations;
- tracking changed shards;
- reusing unchanged shard hashes;
- writing changed blobs and the replacement root tree;
- detecting stale transaction bases;
- compare-and-set publication of the state reference; and
- surfacing malformed objects, missing state, absent records, closed transactions, and conflicts.

## Explicit non-ownership

`internal/ddrepo` does not own:

- the semantic schema of record values;
- version fields inside consumer records;
- migration from one consumer schema to another;
- authored Markdown or code files;
- generated-file replacement;
- review-history events under `refs/ddocs/review`;
- repository-demon runtime files under `.ddocs/runtime/`;
- repository configuration under `.ddocs/config.toml`;
- Git history for the user's working repository;
- retry policy after a transaction conflict; or
- determining whether state can be safely rebuilt from the current filesystem.

Callers must decode their own record formats and decide whether missing, incompatible, or stale domain state should be rebuilt, migrated, or treated as an error.

## Physical layout

An initialized Demon Docs repository contains a bare Git repository at `.ddocs/`.

The durable state reference is:

```text
refs/ddocs/state
```

The reference points directly to a Git tree object. It does not point to a commit.

Each entry in the root tree represents one shard:

```text
0
1
2
3
4
5
6
7
8
9
a
b
c
d
e
f
```

Only non-empty shards need entries. Every root entry must:

- have a one-character lowercase hexadecimal name;
- use regular-file mode; and
- point to a blob containing the encoded shard.

Any other entry name or mode makes the root invalid and causes transaction start to fail.

## Private object-store initialization

`ddrepo.Init` accepts either a repository root or a path already ending in `.ddocs`.

Private object-store initialization performs this sequence:

```text
resolve .ddocs path
-> initialize a bare Git repository
-> encode and write an empty Git tree
-> publish refs/ddocs/state to that tree hash
-> return the repository handle
```

An empty state therefore has a valid reference and root object even though it has no shard entries.

Private object-store initialization does not write product configuration, runtime ownership files, review history, or consumer records. Those are owned by their respective application layers.

## Repository opening

`ddrepo.Open` resolves the same `.ddocs` path and opens the existing bare Git repository.

Opening the object database does not prove that `refs/ddocs/state` exists or is valid. State validity is checked when reading the current reference or beginning a transaction.

The repository handle stores:

- the go-git object/reference storage;
- a process-local mutex; and
- the resolved storage path.

The mutex serializes state-reference reads and commits through that handle. Compare-and-set publication remains the final concurrency authority when multiple handles or processes operate on the same storage.

## Record namespace

Record names are UTF-8 strings with slash-separated logical segments.

Valid names must satisfy all of these conditions:

- non-empty;
- valid UTF-8;
- not Unix-absolute;
- not drive-absolute such as `C:/...`;
- no backslash;
- no NUL byte;
- no empty path segment;
- no `.` segment; and
- no `..` segment.

Examples of valid logical names:

```text
links/files
links/sources/4f8a...
reverse/root/src
schema/version
```

Record names are not operating-system paths. Forward slashes provide a stable logical hierarchy across platforms.

Validation occurs during reads, writes, deletes, shard encoding, and shard decoding. Malformed stored names are rejected rather than exposed to callers.

## Deterministic shard assignment

A record's shard is the first hexadecimal nibble of the SHA-256 digest of its UTF-8 name.

```text
sha256(record name)
-> first digest byte
-> first lowercase hexadecimal character
-> shard 0-f
```

The value does not affect shard placement.

This gives exactly 16 possible shards and keeps a one-record update localized to one shard unless another changed record maps elsewhere.

Shard assignment is deterministic across processes and platforms. There is no configurable shard count or runtime balancing.

## Shard codec

Each shard blob uses a private deterministic binary format.

The header contains:

```text
magic:   DDOC
version: 1 byte, currently 1
count:   4-byte big-endian record count
```

Each record then contains:

```text
name length:  4-byte big-endian unsigned length
value length: 4-byte big-endian unsigned length
name bytes
value bytes
```

Before encoding, record names are sorted lexicographically. The same record map therefore produces the same byte stream and Git blob hash regardless of Go map iteration order.

Values are arbitrary byte slices at this layer. The codec does not inspect JSON, schema versions, or domain content.

## Decode validation

Shard decoding rejects:

- truncated headers;
- an invalid magic header;
- an unsupported shard codec version;
- truncated name or value length fields;
- lengths larger than the remaining blob;
- invalid record names;
- duplicate names; and
- trailing bytes after the declared record count.

The decoder returns copied value slices in a new record map.

A malformed shard is a hard state error. The object repository does not skip the shard, guess its contents, or fall back to a partial state view.

## Root tree encoding

A root tree is built from the current shard-name-to-blob-hash map.

Root entries are sorted by shard name before Git tree encoding. Invalid shard names are rejected before writing.

Git object storage naturally reuses identical blobs and trees by hash. If a changed transaction produces bytes identical to an existing object, writing the object does not create a distinct logical version.

## Transaction start

`Repository.Begin` acquires the repository mutex and reads `refs/ddocs/state`.

When the state reference exists, beginning a transaction:

1. loads the referenced root tree;
2. validates every root entry;
3. records the existing shard hashes without decoding shard blobs; and
4. stores the original reference as the transaction base.

Shard contents remain lazy. A shard blob is decoded only when a transaction operation needs it.

When the state reference is missing, `Begin` creates an empty transaction with no base reference. This permits a caller to publish a replacement state into an otherwise opened object database, although normal private object-store initialization creates the empty reference first.

Errors reading the root tree or validating entries abort transaction creation.

## Transaction state

A transaction tracks:

```text
repo    owning repository handle
base    state reference observed at Begin
shards  original shard hashes from the root tree
loaded  decoded or newly created shard maps
dirty   shards whose logical records changed
closed  whether the transaction can still be used
```

The transaction is an in-memory mutable view over the original root.

It does not lock the state reference for its full lifetime. Other transactions may begin and commit after it starts. Staleness is detected at commit.

## Reads

`Read` and `Get`:

1. verify the transaction is open;
2. validate the record name;
3. derive and lazy-load its shard;
4. locate the record; and
5. return a copy of the value.

Returning a copy prevents callers from mutating transaction state without using `Write`.

A missing record returns `ErrRecordAbsent` with the requested name.

## Writes

`Write` and `Put`:

1. verify the transaction is open;
2. validate the record name;
3. lazy-load the target shard;
4. compare the existing bytes when the record already exists;
5. copy the supplied value into the shard map; and
6. mark the shard dirty only when bytes changed.

Writing the same bytes is a no-op. It does not mark the shard dirty and does not force a new root publication.

The transaction copies the supplied value so later caller mutation does not alter pending state.

## Deletes

`Delete` validates the name and lazy-loads its shard.

Deleting an absent record succeeds without marking the shard dirty.

Deleting an existing record removes it from the in-memory shard map and marks that shard dirty.

If the shard becomes empty, commit removes its entry from the replacement root tree rather than writing an empty shard blob.

## Prefix enumeration

`Names(prefix)` returns sorted record names beginning with the supplied string.

Because shard assignment is hash-based rather than prefix-based, prefix enumeration must load every shard known to the transaction plus any newly created in-memory shards.

The prefix itself is not interpreted as a path and is not validated as a complete record name.

`Names` is deterministic but potentially more expensive than a single-record read. Consumers should not treat the record namespace as a database index.

## Callback transactions

`Repository.Transaction` wraps the common lifecycle:

```text
Begin
-> invoke caller callback
-> commit when callback succeeds and transaction remains open
```

A nil callback closes the transaction and returns an error.

When the callback returns an error, the transaction is closed and no commit is attempted.

If the callback explicitly committed the transaction, the wrapper sees it as closed and does not commit again.

There is no rollback method because uncommitted changes exist only in transaction memory and newly written unreachable Git objects. The state reference remains authoritative until commit succeeds.

## Commit algorithm

Commit first marks the transaction to become closed when the method returns, regardless of success.

It then acquires the repository mutex and performs:

```text
read current refs/ddocs/state
-> compare current reference with transaction base
-> copy original shard-hash map
-> for each dirty shard in sorted order:
     remove root entry when shard is empty
     otherwise encode shard and write blob
-> encode and write replacement root tree
-> return success immediately when current root already equals replacement root
-> compare-and-set refs/ddocs/state from current reference to replacement root
```

Only dirty shards are re-encoded. Unchanged root entries retain their original blob hashes.

A transaction that makes no logical changes may still construct the current root hash, then return without changing the reference when it matches.

## Conflict detection

Before writing replacement state, commit compares the current state reference with the transaction's original base.

The comparison includes:

- both references being absent; or
- the same reference name and hash.

When another transaction has changed `refs/ddocs/state` since `Begin`, commit returns `ErrConflict`.

After this explicit check, final publication uses go-git's `CheckAndSetReference`. This protects against a reference change that occurs between the check and update, including changes made through another repository handle or process.

Callers own retry behavior. A safe retry must begin a new transaction, reread current records, reapply the intended logical operation, and commit against the new base. Reusing the closed stale transaction is invalid.

## Atomicity boundary

The transaction's authoritative publication point is the state-reference update.

Before that update, newly written blobs and trees may exist as unreachable objects. They do not become current state.

After a successful compare-and-set, the new root and all referenced shard blobs become visible together through `refs/ddocs/state`.

The object repository therefore provides atomic publication of its record set, but it does not atomically include:

- authored filesystem writes;
- `refs/ddocs/review` updates;
- runtime JSON and heartbeat files;
- configuration writes; or
- another external system.

Higher-level mutation flows must document ordering and recovery across those separate boundaries.

## Transaction closure

Any commit attempt closes the transaction, including:

- success;
- conflict;
- codec failure;
- object-write failure; or
- reference-publication failure.

A callback error through `Repository.Transaction` also closes it.

Later reads, writes, deletes, name enumeration, or commits return `ErrClosed`.

Closing after failure prevents accidental reuse of a state view whose base or partial object writes may no longer be meaningful.

## Process and cross-process concurrency

The repository mutex serializes operations performed through one `Repository` handle.

It does not prevent:

- another `Repository` handle in the same process;
- another Demon Docs process; or
- a direct writer to the same Git storage

from attempting state publication.

Base-reference comparison and `CheckAndSetReference` provide optimistic concurrency across those writers.

There is no record-level merge inside `ddrepo`. Two transactions that modify different shards still conflict if one publishes first, because the state reference changed. The losing caller must reopen current state and retry its logical operation.

## Error model

The package exposes these primary sentinel errors:

- `ErrConflict` — the state reference no longer matches the transaction base;
- `ErrMissingState` — the state reference or repository state is unavailable;
- `ErrRecordAbsent` — a requested record does not exist; and
- `ErrClosed` — a transaction has already ended.

Additional errors wrap context for:

- invalid record names;
- invalid root entries;
- missing or malformed Git objects;
- malformed or unsupported shard data;
- object encoding or writing; and
- state-reference reads and publication.

Callers should use `errors.Is` for sentinel behavior and preserve contextual error text for diagnostics.

## Corruption behavior

The private object repository fails closed on structural corruption.

It does not:

- skip a malformed shard;
- retain only records that decoded before an error;
- ignore an invalid root entry;
- accept an unknown shard codec version;
- rewrite a damaged root automatically; or
- infer domain state from partial records.

Recovery belongs to the consuming subsystem and operational tooling. Depending on the affected records, current filesystem state may be rebuildable while historical identity or review information may not be recoverable from authored files alone.

See [Repository State and Transactions](repository-state-and-transactions.md) and [Recovery and Troubleshooting](../operations/recovery-and-troubleshooting.md) before deleting private state.

## State-domain boundaries

Several `.ddocs` surfaces are intentionally separate:

```text
.ddocs/config.toml
  selected product configuration

refs/ddocs/state
  transactional consumer records through internal/ddrepo

refs/ddocs/review
  append-only review event history managed by internal/review

.ddocs/runtime/
  disposable demon owner, feeder, shutdown, and log files
```

A successful transaction under `refs/ddocs/state` does not imply that any other surface changed.

The private bare repository may physically contain objects reachable from multiple private references, but each reference has its own publication and schema contract.

## Consumer schemas and compatibility

`ddrepo` stores opaque bytes. Consumer packages own:

- record keys;
- serialization format;
- schema version fields;
- compatibility checks;
- import of legacy state;
- defaulting missing records;
- deciding whether a record is rebuildable; and
- determining when an incompatible state blocks mutation.

A shard codec version change is different from a consumer schema change.

- Shard codec version controls whether `ddrepo` can decode the record container.
- Consumer schema version controls whether a package can interpret one decoded value.

Do not change either without explicit compatibility and recovery documentation.

## Invariants

The private object repository must preserve these invariants:

- `refs/ddocs/state` points to a valid root tree or state access fails.
- Root entries are regular files named with one lowercase hexadecimal character.
- Every record name maps deterministically to exactly one shard.
- Shard bytes are deterministic for the same record set.
- Values crossing the transaction API are copied.
- Unchanged records do not dirty their shard.
- Only dirty shards are replaced in a normal commit.
- Empty dirty shards disappear from the root.
- A stale transaction cannot replace newer state.
- One reference update publishes the complete replacement record set.
- Failed publication does not make newly written unreachable objects authoritative.
- Closed transactions cannot be reused.
- Domain meaning and migration remain outside the storage layer.

## Extension rules

### Adding a record

Choose a stable logical name owned by the consumer subsystem. Use forward slashes, avoid encoding current filesystem paths unless path identity is the intended schema, and document rebuildability and compatibility in the consumer architecture.

### Changing a record schema

Version the consumer payload or provide an unambiguous compatibility check. Define migration, fallback, and failure behavior before shipping a writer that older readers cannot interpret.

Do not change the shard codec merely to evolve one consumer record.

### Adding transactional multi-record behavior

Read and write all related `refs/ddocs/state` records in one transaction when they must become visible together. Handle `ErrConflict` by repeating the logical operation against a fresh transaction rather than replaying raw stale bytes blindly.

### Changing the shard codec

A new codec version requires explicit decode compatibility or a repository migration path. The current decoder rejects unknown versions.

### Adding another private reference

Use a separate reference only when the data requires an independent publication or history model, as review events do. Document cross-reference ordering because `ddrepo` cannot make two references atomic together.

## Verification

Focused verification:

```bash
go test ./internal/ddrepo -count=1
```

The focused tests protect:

- deterministic shard encoding and round-trip decoding;
- rejection of malformed shard data;
- transaction persistence across reopen;
- stale-transaction rejection; and
- one-record updates changing only their assigned shard.

Consumers require their own tests for record schemas, migration, retry, and higher-level mutation ordering.

## Code map

Primary implementation:

- `internal/ddrepo/codec.go` — record-name validation, shard assignment, deterministic binary encoding, and strict decoding.
- `internal/ddrepo/objects.go` — Git blob and root-tree reads and writes.
- `internal/ddrepo/repository.go` — `.ddocs` path resolution, bare repository initialization/opening, state reference access, and callback transaction wrapper.
- `internal/ddrepo/transaction.go` — lazy shard views, record operations, dirty tracking, optimistic conflict detection, and compare-and-set publication.

Focused tests:

- `internal/ddrepo/codec_test.go`
- `internal/ddrepo/repository_test.go`

Important consumers:

- `internal/links/` — durable file identities, link inventory, history, and generated-write state.
- `internal/reverseindex/` — deterministic reverse-index state.
- other current packages that store opaque records below `refs/ddocs/state`.

Separate private-state owners:

- `internal/review/` owns `refs/ddocs/review`.
- `internal/demon/` owns `.ddocs/runtime/`.
- `internal/config/` owns `.ddocs/config.toml` interpretation and mutation.

## Related docs

- [Repository State and Transactions](repository-state-and-transactions.md)
- [Managed Files and State](../reference/managed-files-and-state.md)
- [Review Ledger](review-ledger.md)
- [Repository Demon Lease Protocol](repository-demon-lease-protocol.md)
- [Repository Scope and Worktrees](repository-scope-and-worktrees.md)
- [Compatibility and Migrations](../reference/compatibility-and-migrations.md)
- [Recovery and Troubleshooting](../operations/recovery-and-troubleshooting.md)

## Notes

The use of Git objects is an internal storage mechanism. Demon Docs does not expose the root tree or shard blobs as a public editing interface, and users should not repair records by modifying object data manually.
