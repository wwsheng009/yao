# TUI Event System Refactoring Plan

## Executive Summary

Refactored the TUI event system to eliminate anonymous interfaces in the framework codebase, improving type safety and API discoverability.

## Scope

**Included:** `tui/runtime/`, `tui/framework/`, `tui/framework/component/`, `tui/framework/event/`

**Excluded:** `tui/tui/` (separate project)

## Issues Addressed

### 1. Missing Named Event Handler Interface

The `framework/event` package defined an `Event` interface, but components used anonymous interface assertions:

```go
// Before (anonymous interface):
if handler, ok := currentField.Input.(interface{ HandleEvent(component.Event) bool }); ok {
    // ...
}
```

**Problem:** Prevents compile-time type checking, less discoverable API.

### 2. Event Type Alias Chain

```
component.Event → framework/event.Event (different from runtime/event.Event)
```

This is architecturally sound (framework events are simpler than runtime events).

## Changes Made

### Added `EventComponent` Interface

**File:** `tui/framework/event/event.go`

```go
// EventComponent is the interface for components that handle framework events.
// The Component interface already declares HandleEvent, this serves as a named
// marker for type assertions instead of using anonymous interfaces.
type EventComponent interface {
    Component
}
```

**Note:** Named `EventComponent` instead of `EventHandler` to avoid conflict with the existing `EventHandler` in `handler.go` (which is for callback-style handlers, not components).

### Updated Form Component

**File:** `tui/framework/form/form.go` (lines 791, 814)

```go
// Before:
if handler, ok := currentField.Input.(interface{ HandleEvent(component.Event) bool }); ok {

// After:
if handler, ok := currentField.Input.(event.EventComponent); ok {
```

### Updated E2E Testing

**File:** `tui/framework/testing/e2e.go` (line 136)

```go
// Before:
if handler, ok := ctx.Root.(interface{ HandleEvent(component.Event) bool }); ok {

// After:
if handler, ok := ctx.Root.(event.EventComponent); ok {
```

## Files Modified

| File | Lines | Change |
|------|-------|--------|
| `framework/event/event.go` | 135-140 | Added `EventComponent` interface |
| `framework/form/form.go` | 791, 814 | Replaced anonymous with `event.EventComponent` |
| `framework/testing/e2e.go` | 136 | Replaced anonymous with `event.EventComponent` |

## Test Results

All tests pass:
```bash
$ go test ./tui/framework/...
ok      github.com/yaoapp/yao/tui/framework
ok      github.com/yaoapp/yao/tui/framework/form
ok      github.com/yaoapp/yao/tui/framework/testing
```

## Migration Guide

### For Component Authors

When checking if a component handles events:

```go
// Before (anonymous interface):
if h, ok := input.(interface{ HandleEvent(component.Event) bool }); ok {
    h.HandleEvent(ev)
}

// After (named interface):
if h, ok := input.(event.EventComponent); ok {
    h.HandleEvent(ev)
}
```

### Benefits

- **Type safety:** Named interfaces allow compile-time checking
- **Discoverability:** IDEs can suggest `EventComponent` in completions
- **Documentation:** Clear interface definition in code
- **No breaking changes:** Existing components already satisfy the interface
