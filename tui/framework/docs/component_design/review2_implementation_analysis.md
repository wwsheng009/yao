# TUI Framework Implementation Analysis vs Design Document

## Overview

This document analyzes the current implementation of the Yao TUI framework (`tui/framework` and `tui/runtime`) against the design principles described in `review2.md`.

---

## Design Document Summary

The design document describes a multi-layered TUI engine architecture with the following key systems:

| Layer | Description | Goal |
|-------|-------------|------|
| **1. Dependency Tracking** | State → Component dependency graph | Precise dirty marking |
| **2. Priority Scheduling** | StateZone + DirtyLevel | Time-sliced rendering |
| **3. Incremental Layout** | LayoutDirty + cache | O(affected path) complexity |
| **4. Layer System** | Independent refresh regions | Isolated high-frequency updates |
| **5. GPU Batching** | DrawCmd + style state machine | Minimize terminal IO |
| **6. Remote Optimization** | Frame buffering + delta encoding | SSH-friendly streaming |

---

## Implementation Status Analysis

### ✅ Layer 1: State Management (IMPLEMENTED)

**Design:** `State.Set(key)` → mark dependent components dirty

**Implementation:**
- `ReactiveStore` (`tui/framework/binding/store.go`) - Full reactive store with:
  - Path-based state access (`user.name`)
  - Observer pattern (`Subscribe`, `SubscribeGlobal`)
  - Batch updates (`BeginBatch`/`EndBatch`)
  - Computed values (`StoreComputed`)
  - Context-aware store with cleanup

**Verdict:** ✅ **FULLY IMPLEMENTED** - Reactive state management is solid.

---

### ❌ Layer 1a: Dependency Graph (MISSING)

**Design:**
```go
deps["user.name"] = [Label#42, Button#15]  // Key → Component list
```

When `state.Set("user.name", "Tom")` happens, system automatically marks those components dirty.

**Current Implementation:**
- `ReactiveStore` has observer pattern but observers are manual callbacks
- No automatic registration of which components depend on which state paths
- `Prop.GetDependencies()` exists but doesn't connect back to component marking

**Gap:** State changes → component dirty marking is not automatic. Developer must manually subscribe and mark dirty.

**Verdict:** ❌ **MISSING** - No automatic state → component dependency graph.

---

### ✅ Layer 2: Dirty Marking (IMPLEMENTED)

**Design:** Cell-level and region-level dirty tracking with flood-fill optimization

**Implementation:**
- `DirtyTracker` (`tui/runtime/paint/dirty.go`) implements:
  - Cell-level dirty marking (`MarkCell`)
  - Rectangle-level dirty marking (`MarkRect`)
  - Flood-fill region extraction (`extractRegion`)
  - Region merging (`mergeDirtyRegions`)
  - Diff between buffers (`Diff`)

**Verdict:** ✅ **FULLY IMPLEMENTED** - Advanced dirty tracking with optimization.

---

### ⚠️ Layer 2: Single Dirty Flag (PARTIAL)

**Design:** Two types of dirty flags:
- `dirty` - needs repaint
- `layoutDirty` - needs relayout (size changed)

**Current Implementation:**
- `LayoutNode` has single `dirty` flag (`tui/runtime/node.go:51`)
- No separation between layout dirty and paint dirty

**Gap:** When text length changes, both layout and paint need refresh. Currently both are bundled.

**Verdict:** ⚠️ **PARTIAL** - Single dirty flag, no `layoutDirty` vs `dirty` separation.

---

### ❌ Layer 3: Priority Scheduling (MISSING)

**Design:**
```go
type DirtyLevel int
const (
    DirtyLow DirtyLevel = iota    // Background logs
    DirtyNormal                    // Data tables
    DirtyHigh                      // Input focus
)

type StateZone int
const (
    ZoneUI StateZone = iota
    ZoneData
    ZoneBackground
)
```

Renderer processes dirty nodes by priority level with time budget:
```go
func RenderFrame() {
    budget := 2 * time.Millisecond
    processDirty(root, DirtyHigh, budget)   // Input first
    processDirty(root, DirtyNormal, budget) // Data second
    processDirty(root, DirtyLow, budget)    // Logs last
}
```

**Current Implementation:** Not found. No priority levels or time-sliced rendering.

**Gap:** High-frequency updates (logs, streaming data) can block input responsiveness.

**Verdict:** ❌ **MISSING** - No priority-based rendering or time slicing.

---

### ✅ Layer 4: Layout Cache (IMPLEMENTED)

**Design:** Cache layout results with SHA256 hashing

**Implementation:**
- `Cache` (`tui/runtime/layout/cache.go`) implements:
  - SHA256-based node hashing
  - Constraint-based cache keys
  - Hit count tracking
  - Cache eviction (oldest first)

**Verdict:** ✅ **FULLY IMPLEMENTED** - Layout caching with proper invalidation.

---

### ⚠️ Layer 5: Layer System (BASIC ONLY)

**Design:**
```go
type Layer struct {
    buffer   ScreenBuffer
    dirty    bool
    zIndex   int
    rect     Rect
}
```

Each layer has independent buffer and dirty tracking.

**Current Implementation:**
- `ZIndex` exists in `Cell` (`tui/runtime/runtime.go:121`)
- Z-index comparison in `SetContent` (`tui/runtime/runtime.go:180-182`)
- No independent layer buffers or per-layer dirty tracking

**Gap:** All components render to same buffer. No layer-level isolation for high-frequency updates.

**Verdict:** ⚠️ **PARTIAL** - Z-index exists but no proper layer system.

---

### ❌ Layer 6: GPU-Style Command Batching (MISSING)

**Design:**
```go
type DrawCmd struct {
    X, Y   int
    Text   string
    Style  Style
}
```

Paint generates commands → batch → single flush:
```go
// Before: Multiple moves
Move(1,1) Write("H")
Move(2,1) Write("e")

// After: Single command
Move(1,1) Write("Hel")
```

**Current Implementation:**
- Direct buffer write via `CellBuffer.SetContent()`
- No draw command queue
- No style state machine to minimize VT codes
- No command batching

**Gap:** Each cell write generates terminal output. Could be optimized with batching.

**Verdict:** ❌ **MISSING** - No command batching or style state machine.

---

### ❌ Layer 7: Remote Rendering Optimization (MISSING)

**Design:**
- 16ms frame ticker
- Delta frame encoding
- Scroll region optimization (`ESC[nS`)

**Current Implementation:** Not found.

**Gap:** SSH performance may suffer with high-frequency updates.

**Verdict:** ❌ **MISSING** - No frame buffering or remote optimization.

---

### ✅ Additional Implemented Features

The implementation includes features beyond the design doc:

1. **Expression Parsing** (`tui/framework/binding/expression.go`)
   - Lexer + parser for `{{ price * quantity }}` expressions
   - Dependency extraction
   - AST evaluation

2. **Property System** (`tui/framework/binding/prop.go`)
   - Generic `Prop[T]` for static/bound/expression properties
   - Automatic type detection from strings

3. **State Snapshot/Time Travel** (`tui/runtime/state/`)
   - `Tracker` for undo/redo
   - `Snapshot` for complete state capture
   - `Diff` for state comparison

4. **Selection System**
   - Text selection with mouse
   - Clipboard integration
   - Selection rendering with reverse video

5. **Focus Management**
   - Focus path system
   - Focus scopes for modals
   - Keyboard navigation

---

## Summary Table

| Feature | Design | Implementation | Status |
|---------|--------|----------------|--------|
| Reactive Store | ✅ | ✅ `ReactiveStore` | Complete |
| Dependency Graph | ✅ | ❌ Not automatic | Missing |
| Dirty Tracking | ✅ | ✅ `DirtyTracker` | Complete |
| LayoutDirty vs PaintDirty | ✅ | ⚠️ Single flag | Partial |
| Priority Scheduling | ✅ | ❌ None | Missing |
| Time Sliced Rendering | ✅ | ❌ None | Missing |
| Layout Cache | ✅ | ✅ `layout/cache.go` | Complete |
| Layer System | ✅ | ⚠️ ZIndex only | Partial |
| Command Batching | ✅ | ❌ None | Missing |
| Style State Machine | ✅ | ❌ None | Missing |
| Remote Optimization | ✅ | ❌ None | Missing |
| Expression Parsing | - | ✅ `expression.go` | Bonus |
| State Snapshots | - | ✅ `state/` | Bonus |

---

## Key Gaps Summary

### High Priority (Performance Impact)

1. **No Dependency Graph** - State changes don't auto-mark components dirty
2. **No Priority Scheduling** - High-frequency updates can block UI
3. **No Command Batching** - Excessive terminal IO

### Medium Priority (Architecture Completeness)

4. **Single Dirty Flag** - Can't distinguish layout vs paint needs
5. **Basic Layer System** - No independent refresh regions

### Low Priority (Edge Cases)

6. **No Remote Optimization** - SSH performance not addressed

---

## Files Referenced

**Design Document:**
- `review2.md` - Architecture design specification

**Implementation:**
- `tui/framework/binding/store.go` - Reactive state management
- `tui/framework/binding/prop.go` - Property system
- `tui/framework/binding/expression.go` - Expression parser
- `tui/runtime/paint/dirty.go` - Dirty region tracking
- `tui/runtime/layout/cache.go` - Layout caching
- `tui/runtime/node.go` - Layout node with dirty flag
- `tui/runtime/runtime.go` - Core runtime with CellBuffer
- `tui/runtime/state/tracker.go` - State tracking
- `tui/runtime/state/snapshot.go` - State snapshots
- `tui/runtime/state/diff.go` - State diffing
