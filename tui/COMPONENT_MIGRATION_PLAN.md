# TUI Component Migration Plan

## Current Status (2026-01-23)

### ✅ Completed
- Runtime Core Engine (Measure → Layout → Render)
- Native Runtime Components: Input, Button, Text, Header, Footer, List
- Layout Components: Row, Column, Flex
- DSL Factory with props binding
- Component Registry
- Dynamic State Synchronization
- Template Binding ({{key}})
- Test Coverage: 8 Runtime tests all passing

### 🎯 Next Priority: Table Component

## Phase 1: Complex Component Migrations

### 1.1 Table Component (HIGH PRIORITY)
**Status:** 🔄 In Progress
**Complexity:** High
**Value:** High - Essential for data-heavy applications

**Features:**
- Multi-column layout with headers
- Cell rendering with various content types
- Vertical and horizontal scrolling
- Column sizing (fixed, percentage, auto)
- Row selection (single/multi)
- Sorting hooks
- Filtering hooks

**Files:**
- `tui/ui/components/table.go` (NEW, ~700 lines)
- `tui/runtime/dsl/factory.go` - add `applyTableProps()`, `ParseTableItems()`
- `tui/runtime/registry/registry.go` - register "table"
- `tui/model_runtime_integration.go` - add state sync for tables
- `tui/runtime_e2e_test.go` - add `TestRuntimeTableComponent`

**Estimated:** 2-3 hours

---

### 1.2 Form Component
**Status:** ⏳ Pending
**Complexity:** Medium
**Value:** High - Essential for data entry

**Features:**
- Container for form fields
- Validation state management
- Submit/cancel actions
- Error display
- Field grouping
- Disabled state for all fields

**Files:**
- `tui/ui/components/form.go` (NEW, ~400 lines)
- `tui/runtime/dsl/factory.go` - add `applyFormProps()`
- Test: `TestRuntimeFormComponent`

**Estimated:** 1.5-2 hours

---

### 1.3 Modal/Dialog Component
**Status:** ⏳ Pending
**Complexity:** Medium
**Value:** Medium - Important for overlays

**Features:**
- Overlay rendering (z-index)
- Title and content
- Action buttons (confirm/cancel)
- Backdrop dimming
- Keyboard shortcuts (ESC to close)
- Focus trapping

**Files:**
- `tui/ui/components/modal.go` (NEW, ~350 lines)
- `tui/runtime/dsl/factory.go` - add `applyModalProps()`
- Test: `TestRuntimeModalComponent`

**Estimated:** 1.5 hours

---

### 1.4 Tabs Component
**Status:** ⏳ Pending
**Complexity:** Medium
**Value:** Medium - Common navigation pattern

**Features:**
- Tab labels with icons
- Active tab highlighting
- Tab content panels
- Keyboard navigation (Ctrl+Tab, arrow keys)
- Scrollable tab bar
- Closable tabs (optional)

**Files:**
- `tui/ui/components/tabs.go` (NEW, ~400 lines)
- `tui/runtime/dsl/factory.go` - add `applyTabsProps()`
- Test: `TestRuntimeTabsComponent`

**Estimated:** 2 hours

---

## Phase 2: Supporting Components

### 2.1 Progress Component
**Complexity:** Low
**Value:** Low - Nice to have

**Features:**
- Horizontal progress bar
- Percentage display
- Indeterminate mode (spinner)
- Color gradients
- Custom labels

**Files:**
- `tui/ui/components/progress.go` (NEW, ~200 lines)
- `tui/runtime/dsl/factory.go` - add `applyProgressProps()`

**Estimated:** 1 hour

---

### 2.2 Spinner Component
**Complexity:** Low
**Value:** Low - Loading states

**Features:**
- Multiple spinner types (dots, line, arrows)
- Custom frames
- Speed control
- Color support
- Text label

**Files:**
- `tui/ui/components/spinner.go` (NEW, ~150 lines)
- `tui/runtime/dsl/factory.go` - add `applySpinnerProps()`

**Estimated:** 0.5-1 hour

---

## Phase 3: Architecture Improvements

### 3.1 Enhanced Rendering with Lipgloss
**Status:** ⏳ Pending
**Priority:** Medium

**Tasks:**
1. Create `tui/runtime/render/lipgloss_adapter.go`
   - Convert lipgloss.Style to CellStyle
   - Parse ANSI colors
   - Handle text decorations

2. Update `tui/runtime/runtime_impl.go`
   - Use lipgloss in `renderComponent()`
   - Support multi-line text wrapping
   - Apply styles correctly

3. Update `tui/runtime/runtime.go`
   - `CellBuffer.String()` with ANSI escape codes
   - Proper terminal output

4. Create `tui/runtime/render/diff.go`
   - Frame-to-frame diffing
   - Dirty region tracking
   - Minimal terminal updates

**Estimated:** 3-4 hours

---

### 3.2 Event System (Phase 3)
**Status:** ⏳ Pending
**Priority:** High

**Tasks:**
1. Create `tui/runtime/event/hittest.go`
   - Geometry-based hit testing
   - Component bounds checking
   - Z-index support

2. Create `tui/runtime/event/dispatch.go`
   - Event routing based on geometry
   - Event bubbling/capturing
   - Event delegation

3. Enhanced Mouse Support
   - Click detection
   - Drag and drop hooks
   - Scroll wheel handling

4. Enhanced Keyboard Support
   - Global key bindings
   - Key conflict resolution
   - Shortcut system

**Estimated:** 4-5 hours

---

### 3.3 Focus Management Enhancement
**Status:** ⏳ Pending
**Priority:** Medium

**Tasks:**
1. Create `tui/runtime/focus/focus.go`
   - Focus order specification
   - Focus groups
   - Focus history
   - Tab cycle customization

2. Focus Indicators
   - Visual feedback
   - Configurable styles
   - Animation support

**Estimated:** 2-3 hours

---

### 3.4 Performance Optimization (Phase 4)
**Status:** ⏳ Pending
**Priority:** Medium

**Tasks:**
1. Diff Rendering
   - Only re-render changed regions
   - Component-level caching
   - Layout result caching

2. Smart Invalidation
   - Dirty flag propagation
   - Subtree re-rendering
   - State change batching

3. Profiling Tools
   - Render timing
   - Layout timing
   - Memory usage tracking

**Estimated:** 3-4 hours

---

## Phase 4: Documentation & Examples

### 4.1 Component Documentation
- Create examples for each component
- DSL configuration samples
- Usage patterns
- Best practices

### 4.2 Migration Guide
- Legacy → Runtime migration steps
- Component mapping
- Breaking changes
- Performance tips

### 4.3 Architecture Documentation
- Runtime design decisions
- Three-phase rendering explained
- Layout algorithm details
- Event flow diagrams

---

## Execution Order

### Sprint 1: Core Components (Current)
1. ✅ List Component (DONE)
2. 🔄 Table Component (IN PROGRESS)
3. ⏳ Form Component
4. ⏳ Modal Component

### Sprint 2: Navigation & Feedback
5. ⏳ Tabs Component
6. ⏳ Progress Component
7. ⏳ Spinner Component

### Sprint 3: Architecture
8. ⏳ Enhanced Rendering (Lipgloss)
9. ⏳ Event System
10. ⏳ Focus Management

### Sprint 4: Polish
11. ⏳ Performance Optimization
12. ⏳ Documentation
13. ⏳ Examples

---

## Success Criteria

- ✅ All legacy components have native Runtime equivalents
- ✅ No legacy dependencies in Runtime components
- ✅ All tests passing (target: 15+ Runtime tests)
- ✅ Performance parity or better than legacy
- ✅ Full documentation coverage
- ✅ Migration guide available

---

## Current Blockers

None identified.

---

## Notes

- All Runtime components are pure native implementations
- No dependencies on legacy `tui/components/` package
- Each component includes comprehensive test coverage
- DSL factory provides unified API for all components
- State synchronization works for all component types
