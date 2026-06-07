# AUDIT: Parallelization & Concurrency Optimization

# go-plm / Markdown-PLM v2.0

**Date:** 2026-06-07
**Auditor:** Principal Software Engineer — Performance & Concurrency
**Language:** Go 1.26
**Scope:** 22 packages, ~2300 LOC production code

---

# 1. Executive Summary

## Key Findings

The go-plm codebase is a **clean, well-structured Go application** with three systemic sequential bottlenecks:

1. **Triple-write pattern** — `SaveObject` + `Index.UpsertObject` + `AppendHistory` execute sequentially in `CreateObject`, `RunTransition`, and `AttachArtifact`. All three are independent writes to different backends (filesystem, SQLite, event log) sharing only read-only input. **Fix: fan out with `errgroup`.**

2. **Sequential object reads in loops** — `ListObjects`, `GetBOM`, `CheckReadiness`, `GenerateManifest`, `FindDuplicates` all iterate over collections calling `GetObject` one-by-one. Each read is independent. **Fix: bounded worker pool + errgroup for bulk reads.**

3. **SQLite single-connection bottleneck** — `SetMaxOpenConns(1)` prevents all intra-transaction parallelism. Three independent DELETE statements in `DeleteObject` and `RebuildIndex` run sequentially despite targeting different tables. **Fix: batch INSERTs, parallelize DELETEs within WAL mode.**

## What CAN be safely parallelized

| Priority | Operation | Strategy | Gain | Risk |
|----------|-----------|----------|------|------|
| **P0** | `CreateObject`/`RunTransition` triple-write | `errgroup` fan-out | **3x** write throughput | Low |
| **P0** | `ListObjects` — parallel `GetObject` | Worker pool (limit=8) | **N/8x** for large projects | Low |
| **P1** | `CheckReadiness`/`GenerateManifest` — bulk reads | Worker pool | **linear** with scope size | Low |
| **P1** | `walkBOM` — parallel child reads | Per-level errgroup | **linear** with BOM width | Low |
| **P1** | `RebuildIndex` — batch INSERTs | SQL bulk insert (500/batch) | **10-50x** rebuild speed | Low |
| **P2** | `FindDuplicates` — parallel comparisons | Worker pool | **linear** with library size | Medium |
| **P2** | `ValidateProject` — parallel object validation | Worker pool | **N/8x** | Low |

## What CANNOT be parallelized

| Operation | Reason |
|-----------|--------|
| FSM lifecycle transitions (`draft → approved → released`) | Sequential state machine by design |
| Naming sequence generation (`NextID`) | Monotonic counter — must be sequential |
| Git commit chain | Hash-chain dependency between commits |
| Transaction plan execution order | Phases ordered: write → rename → delete |
| `GetBOM` flat rollup | Grouping operation is inherently sequential |

## Where maximum gain is expected

1. **Triple-write fan-out** (`command.go`): 3 sequential I/O calls → 1 parallel fan-out. Measurable in every object mutation. Affects: `CreateObject`, `RunTransition`, `AttachArtifact`.

2. **Bulk object reads** (`fsrepo.go`, `release.go`, `bom.go`): Any operation reading N objects does N sequential file I/O calls. With a worker pool of 8: 1000 objects = 125 sequential windows instead of 1000.

3. **Index rebuild batching** (`index.go`): Single INSERT per object → bulk INSERT 500/batch. Index rebuild for 10K objects drops from ~5 seconds to ~0.5 seconds.

## Critical risks

- **SQLite WAL + parallelism**: SQLite in WAL mode supports concurrent readers but only one writer. All parallel writes MUST go through the same connection.
- **File descriptor exhaustion**: Parallel file reads can exhaust `ulimit -n`. Worker pools MUST be bounded.
- **Map iteration non-determinism** (`validation/engine.go:52`): `for _, rule := range e.Registry.rules` iterates a Go map in random order. Currently harmless because rules are independent, but **must be fixed before any rule depends on another's output**.

---

# 2. Candidate Operations for Parallelization

| # | Location | Current Behavior | Parallelization Potential | Strategy | Risk | Confidence |
|---|----------|-----------------|--------------------------|----------|------|------------|
| 1 | `command.go:86-106` | SaveObject → UpsertObject → AppendHistory sequential | **high** | errgroup fan-out | low | high |
| 2 | `command.go:131-147` | Same pattern in RunTransition | **high** | errgroup fan-out | low | high |
| 3 | `command.go:222-227` | SaveObject → UpsertObject in AttachArtifact | **high** | errgroup fan-out | low | high |
| 4 | `fsrepo.go:108-124` | Sequential GetObject per directory entry | **high** | Worker pool (limit=8) | low | high |
| 5 | `release.go:104-119` | Sequential GetObject per scope object | **high** | Worker pool | low | high |
| 6 | `release.go:136-139` | Same in GenerateManifest | **high** | Worker pool | low | high |
| 7 | `bom.go:120-126` | Sequential GetObject per child in BOM walk | **high** | Per-level errgroup | low | high |
| 8 | `index.go:169-183` | 3 DELETEs + sequential INSERTs | **high** | Parallel DELETEs + batch INSERT | low | high |
| 9 | `standardparts.go:111-117` | Sequential comparison per existing part | **high** | Worker pool | medium | medium |
| 10 | `engine.go:62-69` | Sequential object validation in ValidateProject | **medium** | Worker pool | low | high |
| 11 | `bom.go:142-146` | O(N) visited map clone per recursion | **medium** | Backtracking (delete from map) | low | high |
| 12 | `transaction.go:101-151` | Sequential ops within each phase | **medium** | errgroup within phase | medium | medium |
| 13 | `index.go:110-114` | 3 sequential DELETEs in transaction | **medium** | Parallel (same tx) | medium | medium |
| 14 | `bom.go:185-198` | Sequential BOM row validation | **low** | Worker pool | low | high |
| 15 | `index.go:191-193` | 3 sequential COUNT queries | **low** | errgroup | low | high |
| 16 | `engine.go:52-59` | Non-deterministic map iteration over rules | **needs fix** | Ordered slice/map | low | high |

---

# 3. Detailed Findings

---

### Finding 1: Triple-write pattern in command layer (P0)

**Location:**
- `internal/app/command/command.go`
- Functions: `CreateObject` (lines 86–106), `RunTransition` (lines 131–147), `AttachArtifact` (lines 222–227)

**Current behavior:**
```go
// CreateObject — 3 sequential I/O operations
s.Repo.SaveObject(ctx, obj)       // Step 1: filesystem write
s.Index.UpsertObject(ctx, obj)    // Step 2: SQLite index update (waits for Step 1)
s.Repo.AppendHistory(...)         // Step 3: event log append (waits for Step 2)
```

**Dependency analysis:**
- Step 1 depends on: `obj` (built in memory)
- Step 2 depends on: `obj` only (NOT on Step 1 result)
- Step 3 depends on: `evtJSON` only (NOT on Step 1 or Step 2)
- **Shared state:** No — `obj` is read-only after construction
- **Order required:** No — all three writes are to independent backends
- **Side effects:** Filesystem write, SQLite write, event log append — all idempotent with `operation_id`
- **Three completely independent operations masquerading as a sequential pipeline.**

**Parallelization potential: HIGH**

**Recommended strategy:** `errgroup` with 3 goroutines

```go
g, ctx := errgroup.WithContext(ctx)
g.Go(func() error { return s.Repo.SaveObject(ctx, obj) })
g.Go(func() error { return s.Index.UpsertObject(ctx, obj) })
g.Go(func() error { return s.Repo.AppendHistory(ctx, obj.ID, string(evtJSON)) })
return g.Wait()
```

**Expected impact:** **HIGH** — 3 sequential I/O calls of ~1-5ms each become 1 parallel fan-out bounded by the slowest (~5ms total vs ~15ms). This affects EVERY object mutation.

**Risks:** Low. All three operations are to different backends. Failures are independent — if `AppendHistory` fails, the object is still saved and indexed (acceptable: history is best-effort). No shared mutable state.

**Required tests:**
- Unit: `TestCreateObjectParallel` — verify all 3 writes complete
- Unit: `TestCreateObjectSaveFails` — verify error from SaveObject propagates
- Unit: `TestCreateObjectPartialFailure` — verify behavior when Index fails but Save succeeds
- Race detector: `go test -race ./internal/app/command/`

**Confidence: HIGH**

---

### Finding 2: Sequential object reads in ListObjects (P0)

**Location:**
- `internal/store/fsrepo/fsrepo.go`
- Function: `ListObjects` (lines 108–124)

**Current behavior:**
```go
for _, entry := range entries {
    obj, err := r.GetObject(ctx, id)  // Read file + parse YAML — one at a time
    objects = append(objects, obj)
}
```

**Dependency analysis:**
- Each `GetObject` call: reads a different file, parses different YAML
- No data dependency between iterations
- Order of result slice matters for `sort.Slice` callers — but we can collect results then sort
- **Shared state:** `objects` slice — needs mutex or channel-based collection

**Parallelization potential: HIGH**

**Recommended strategy:** Worker pool (limit=8) + result collector

```go
sem := make(chan struct{}, 8)
var mu sync.Mutex
for _, entry := range entries {
    sem <- struct{}{}
    go func(id object.ID) {
        defer func() { <-sem }()
        obj, err := r.GetObject(ctx, id)
        mu.Lock()
        objects = append(objects, obj)
        mu.Unlock()
    }(object.ID(entry.Name()))
}
```

**Expected impact: HIGH** — for 500 objects: 500 sequential file reads → 500/8 = ~63 windows of 8 parallel reads.

**Risks:** Low. File descriptor limit — worker pool of 8 stays safely under `ulimit -n` (typically 1024). Order is preserved via sort after collection.

**Required tests:**
- `TestListObjectsParallel` — N objects, verify all N returned
- `TestListObjectsRace` — `go test -race`
- `TestListObjectsFileDescriptorLimit` — 1000 objects, verify no `too many open files`

**Confidence: HIGH**

---

### Finding 3: Sequential reads in Release scope check (P1)

**Location:**
- `internal/modules/release/release.go`
- Functions: `CheckReadiness` (lines 104–119), `GenerateManifest` (lines 136–139)

**Current behavior:**
```go
for _, id := range scope.Objects {
    obj, err := s.Objects.GetObject(ctx, object.ID(id))  // One at a time
    // ... check state, collect blockers
}
```

**Identical pattern to Finding 2.** Release scope of 500 objects = 500 sequential reads.

**Parallelization potential: HIGH**

**Recommended strategy:** Same worker pool pattern as Finding 2.

**Expected impact: HIGH** — linear with scope size.

**Risks:** Low. Same as Finding 2.

**Confidence: HIGH**

---

### Finding 4: Sequential child reads in BOM walk (P1)

**Location:**
- `internal/modules/bom/bom.go`
- Function: `walkBOM` (lines 120–126)

**Current behavior:**
```go
for i, rel := range rels {
    childObj, err := s.Objects.GetObject(ctx, object.ID(rel.ToID))  // Sequential per child
    // ... build BOM row
}
```

**Parallelization potential: HIGH** — all children at the same level are independent reads.

**Recommended strategy:** Per-level errgroup — fetch all children at current level concurrently, then recurse.

**Expected impact: HIGH** — a BOM node with 50 children does 50 parallel reads instead of 50 sequential.

**Risks:** Low within the same level. Must NOT parallelize across levels (recursion depends on level completion).

**Confidence: HIGH**

---

### Finding 5: Index rebuild — sequential INSERTs (P1)

**Location:**
- `internal/store/index/index.go`
- Function: `RebuildIndex` (lines 169–183)

**Current behavior:**
```go
for _, obj := range objects {
    tx.ExecContext(ctx, "INSERT INTO objects ...", ...)  // One INSERT per object
}
```

**Parallelization potential: HIGH**

**Recommended strategy:** Batch INSERT with multiple VALUES clauses:

```go
const batchSize = 500
for i := 0; i < len(objects); i += batchSize {
    batch := objects[i:min(i+batchSize, len(objects))]
    // Build: INSERT INTO objects VALUES (?,?,?,...), (?,?,?,...), ...
    tx.ExecContext(ctx, query, args...)
}
```

**Expected impact: HIGH** — 10K sequential INSERTs → 20 batch INSERTs. SQLite batch INSERT is 10-50x faster than individual INSERTs in a transaction.

**Risks:** Low. SQLite has a practical limit of ~999 parameters per statement (500 rows × 10 columns = 5000 params — fine).

**Confidence: HIGH**

---

### Finding 6: Standard parts duplicate comparison (P2)

**Location:**
- `internal/modules/standardparts/standardparts.go`
- Function: `FindDuplicates` (lines 111–117)

**Parallelization potential: HIGH** — each comparison is independent CPU-bound operation.

**Recommended strategy:** Worker pool for large libraries (>1000 parts). For smaller libraries, overhead outweighs gain.

**Expected impact: MEDIUM** — only significant when library >1000 parts.

**Risks:** Medium. `classifyMatch` is pure function — safe to parallelize. But worker pool overhead may dominate for small N.

**Confidence: MEDIUM**

---

### Finding 7: Visited map clone in BOM walk (P2)

**Location:**
- `internal/modules/bom/bom.go`
- `walkBOM` (lines 142–146)

**Current behavior:**
```go
newVisited := make(map[string]bool)
for k, v := range visited { newVisited[k] = v }  // O(N) clone per recursion level
```

**Parallelization potential: MEDIUM** — optimization, not parallelization.

**Recommended strategy:** Use backtracking (like `detectCycles` does):
```go
visited[id] = true
defer delete(visited, id)  // Backtrack: remove after recursion returns
```

**Expected impact: MEDIUM** — eliminates O(depth × N) map allocation. `detectCycles` already uses this pattern correctly.

**Risks:** Low — pure code optimization.

**Confidence: HIGH**

---

### Finding 8: Non-deterministic validation rule order (P2 — correctness fix)

**Location:**
- `internal/validation/engine/engine.go`
- Function: `ValidateObject` (lines 52–59)

**Current behavior:**
```go
for _, rule := range e.Registry.rules {  // map iteration — NON-DETERMINISTIC order
    diags, err := rule.Check(ctx, target)
    allDiags = append(allDiags, diags...)
}
```

**Dependency analysis:**
- Rules are stored in `map[string]Rule` — iteration order is random in Go
- `DefaultRegistry` registers `ValidName` then `RequiredMetadata` with intent of ordering
- Currently harmless: rules are independent
- **BUT: if any future rule depends on another rule's output, this becomes a bug**

**Recommended fix:** Replace `map[string]Rule` with ordered slice or `[]Rule`:
```go
type Registry struct {
    rules []Rule
    index map[string]int  // for fast Get()
}
```

**Expected impact: LOW now, HIGH later** — correctness fix before it becomes a bug.

**Risks:** Low — internal change, API compatible.

**Confidence: HIGH**

---

# 4. Unsafe or Risky Operations

### UNSAFE-1: SQLite WAL writes across multiple connections

**Location:** `internal/store/index/index.go:40` — `db.SetMaxOpenConns(1)`

**Why unsafe:** SQLite in WAL mode allows concurrent READERS but only ONE writer. Multiple goroutines calling `UpsertObject` on the same DB must go through the same connection. The current `SetMaxOpenConns(1)` is correct — do NOT remove it.

**Recommendation:** Keep `MaxOpenConns=1`. For write parallelism, batch writes into single transactions instead of concurrent connections.

---

### UNSAFE-2: FSM lifecycle transitions — inherently sequential

**Location:** `internal/process/fsm/fsm.go`, `internal/app/command/command.go:RunTransition`

**Why unsafe to parallelize:** `draft → submit_review → approve → release` is a sequential state machine by design. Each state depends on the previous. Parallelizing transitions on the same object would cause state corruption.

**Recommendation:** Do NOT parallelize. The sequential FSM is correct by design.

---

### UNSAFE-3: Git commit chain — hash-chain dependency

**Location:** `internal/gitops/gitops.go:Checkpoint`

**Why unsafe to parallelize:** Each commit's parent hash is the previous commit's hash. The chain is mathematically sequential.

**Recommendation:** Do NOT parallelize.

---

### UNSAFE-4: Transaction plan execution — phase ordering

**Location:** `internal/store/transaction/transaction.go:Execute`

**Why partially unsafe:** The six phases (write → copy → rename → delete) MUST execute in order. Within each phase, operations on different files are independent and can be parallelized. But cross-phase ordering is critical.

**Recommendation:** Parallelize within phases, keep phase ordering sequential.

---

### UNSAFE-5: Naming sequence counter — monotonic invariant

**Location:** `internal/naming/naming.go:MemorySequenceStore.Next`

**Why unsafe to parallelize:** Sequence counter (`prt: 0 → 1 → 2 → 3`) is monotonic. Two goroutines calling `Next("prt")` concurrently would produce duplicate IDs.

**Recommendation:** Do NOT parallelize. The `sync.Mutex` in `MemorySequenceStore` is correct and required. For high-throughput ID generation, consider pre-allocating ID ranges to workers (advanced pattern, not needed for MVP).

---

# 5. Recommended Implementation Plan

### Phase 1 — Low-risk wins (Week 1-2)

| # | Action | Location | Effort |
|---|--------|----------|--------|
| 1.1 | Fan out triple-write with errgroup | `command.go` CreateObject/RunTransition/AttachArtifact | 1h |
| 1.2 | Batch INSERT in RebuildIndex | `index.go` RebuildIndex (500/batch) | 30m |
| 1.3 | Backtracking in walkBOM (visited map) | `bom.go` walkBOM | 15m |
| 1.4 | Add `golang.org/x/sync/errgroup` dependency | `go.mod` | 5m |

### Phase 2 — Medium-risk improvements (Week 2-3)

| # | Action | Location | Effort |
|---|--------|----------|--------|
| 2.1 | Worker pool for ListObjects | `fsrepo.go` ListObjects (limit=8) | 2h |
| 2.2 | Worker pool for CheckReadiness | `release.go` CheckReadiness | 1h |
| 2.3 | Worker pool for GenerateManifest | `release.go` GenerateManifest | 30m |
| 2.4 | Per-level errgroup for BOM child reads | `bom.go` walkBOM | 2h |
| 2.5 | Fix validation rule registry to ordered slice | `engine.go` Registry | 1h |
| 2.6 | Worker pool for FindDuplicates | `standardparts.go` (if library >1000) | 1h |

### Phase 3 — Architectural changes (Post-MVP)

| # | Action | Location | Effort |
|---|--------|----------|--------|
| 3.1 | Event-driven write pipeline: Command → WriteQueue → (FS + Index + History) | New `internal/app/pipeline/` | 3d |
| 3.2 | Read-through cache for GetObject | New `internal/store/cache/` | 2d |
| 3.3 | Pre-fetching for BOM traversal | `modules/bom/` | 1d |
| 3.4 | Connection pool for parallel index reads | `store/index/` (read-only replicas) | 2d |

---

# 6. Required Tests

### For Phase 1 changes

| Change | Test |
|--------|------|
| Triple-write errgroup | `TestCreateObject_AllWritesSucceed`, `TestCreateObject_SaveFails_ErrorPropagated`, `TestCreateObject_IndexFails_SavePersisted`, `TestRunTransition_Parallel` |
| Batch INSERT | `TestRebuildIndex_BatchInsert`, `TestRebuildIndex_LargeDataset(10000)` |
| Backtracking | `TestBOMWalk_BacktrackingVisited`, `TestBOMWalk_DeepNesting(50)` |

### For Phase 2 changes

| Change | Test |
|--------|------|
| Worker pool ListObjects | `TestListObjects_Parallel(1000)`, `TestListObjects_Race`, `TestListObjects_FileDescriptorLimit` |
| Worker pool Readiness | `TestCheckReadiness_Parallel(500)` |
| BOM parallel reads | `TestBOMWalk_ParallelChildren`, `TestBOMWalk_DeepHierarchy` |
| Ordered rule registry | `TestRegistry_DeterministicOrder`, `TestRegistry_PreservesRegistrationOrder` |

### Concurrency tests (all phases)

```bash
go test -race -count=100 ./internal/app/command/
go test -race -count=100 ./internal/store/fsrepo/
go test -race -count=100 ./internal/modules/release/
```

### Stress tests

```bash
# Create 10K objects, verify no races
go test -run TestStressCreate -count=1 -timeout=300s

# Validate 10K objects in parallel
go test -run TestStressValidate -count=1 -timeout=300s
```

---

# 7. Metrics to Measure

### Before changes (baseline)

```bash
# Create 1000 objects sequentially
time go test -run TestBaselineSeqCreate -count=1

# List 1000 objects
time go test -run TestBaselineSeqList -count=1

# BOM walk with 100 children
time go test -run TestBaselineSeqBOM -count=1
```

### After changes

| Metric | Before (est.) | After (est.) | Improvement |
|--------|--------------|--------------|-------------|
| CreateObject latency | ~15ms (3 sequential I/O) | ~5ms (1 parallel fan-out) | **3x** |
| ListObjects (1000 objs) | ~1000ms | ~125ms (8 workers) | **8x** |
| RebuildIndex (10K objs) | ~5000ms | ~500ms (batch INSERT) | **10x** |
| BOM walk (50 children/level) | ~250ms | ~50ms (parallel reads) | **5x** |
| CheckReadiness (500 scope) | ~500ms | ~65ms (8 workers) | **8x** |
| FindDuplicates (10K library) | ~100ms | ~15ms (8 workers) | **7x** |
| Memory (ListObjects 1000) | ~50MB | ~80MB (+30MB for goroutines) | acceptable |
| File descriptors (peak) | ~5 | ~13 (8 workers + 5 base) | safe (<<1024) |

---

# 8. Final Recommendations

### Do first (this week)

1. **Fan out triple-write with errgroup** (`command.go`) — highest impact, lowest risk, 3 files changed. Affects every object mutation. Estimated: 1 hour coding + 30 min testing.

2. **Batch INSERT in RebuildIndex** (`index.go`) — simple SQL optimization. 500 rows per INSERT. Estimated: 30 min.

3. **Fix visited map cloning** (`bom.go`) — use backtracking pattern from `detectCycles`. Copy-paste fix. Estimated: 15 min.

### Do next (next week)

4. **Worker pool for ListObjects** (`fsrepo.go`) — bounded goroutines, mutex-protected result slice. Template pattern reusable for Findings 5, 6. Estimated: 2 hours.

5. **Apply same worker pool to CheckReadiness + GenerateManifest** (`release.go`) — identical pattern. Estimated: 1 hour.

6. **Per-level parallel BOM reads** (`bom.go`) — errgroup for all children at current recursion level. Estimated: 2 hours.

7. **Fix rule registry ordering** (`engine.go`) — replace map with ordered slice. Prevents future bug. Estimated: 1 hour.

### Do NOT parallelize

- FSM lifecycle transitions (sequential by design)
- Naming sequence generation (monotonic counter)
- Git commit chain (hash-chain dependency)
- Transaction phase ordering (write before rename before delete)
- SQLite writes across multiple connections (keep `MaxOpenConns=1`)

### Add tests for

- Race detector on all parallelized functions
- Partial failure scenarios (save succeeds, index fails)
- File descriptor exhaustion (1000 parallel reads with bounded pool)
- Deterministic rule execution order

### Measure

Run `go test -bench=. -benchmem` before and after each phase. Track:
- Allocations per operation
- Bytes allocated per operation
- Wall-clock time for bulk operations (>100 objects)

---

## Appendix A: Dependency Map

```
command.go
  CreateObject ─── SaveObject ─── UpsertObject ─── AppendHistory
                    │ independent │ independent │
                    └───── all three can run in parallel ─────┘

release.go
  CheckReadiness ─── for each scope object ─── GetObject
                      │ all independent │
                      └── worker pool ──┘

bom.go
  walkBOM ─── for each child at level N ─── GetObject ─── recurse level N+1
               │ all independent │               │ must wait │
               └── errgroup ──────┘               └── sequential ──┘

index.go
  RebuildIndex ─── DELETE artifacts ─── DELETE relations ─── DELETE objects
                    │ independent │
                    └── parallel DELETEs ──┘
                  ─── INSERT obj1 ─── INSERT obj2 ─── ... (sequential)
                    │ all independent │
                    └── batch INSERT 500/batch ──┘
```

## Appendix B: Go Concurrency Patterns Reference

```go
// Pattern 1: errgroup fan-out (for Findings 1, 2, 3)
g, ctx := errgroup.WithContext(ctx)
g.Go(func() error { return op1(ctx) })
g.Go(func() error { return op2(ctx) })
g.Go(func() error { return op3(ctx) })
return g.Wait()

// Pattern 2: Bounded worker pool (for Findings 4, 5, 6, 9, 10)
sem := make(chan struct{}, 8)
var mu sync.Mutex
for _, item := range items {
    sem <- struct{}{}
    go func(it Item) {
        defer func() { <-sem }()
        result := process(it)
        mu.Lock()
        results = append(results, result)
        mu.Unlock()
    }(item)
}
// Drain semaphore
for i := 0; i < cap(sem); i++ { sem <- struct{}{} }

// Pattern 3: Batch SQL INSERT (for Finding 8)
const batchSize = 500
for i := 0; i < len(rows); i += batchSize {
    end := min(i+batchSize, len(rows))
    batch := rows[i:end]
    query := buildBatchInsert(batch)
    tx.ExecContext(ctx, query, flattenParams(batch)...)
}
```
