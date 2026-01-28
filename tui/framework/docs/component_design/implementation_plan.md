# TUI Framework Performance Optimization Implementation Plan

## Overview

This plan implements the missing performance optimization features from the TUI Framework design document (`tui/framework/docs/component_design/review2.md`).

**Priority Order:**
1. Separate `layoutDirty` vs `paintDirty` flags (Foundation)
2. Automatic dependency graph (Core)
3. Priority scheduling system (Performance)
4. GPU-style command batching (Performance)
5. Independent layer buffers (Advanced)
6. Remote rendering optimization (Edge case)

---

## Phase 1: Separate layoutDirty vs paintDirty Flags

### Goal
Allow the runtime to distinguish between components that need re-layout (size changed) vs components that only need repainting (content changed but size same).

### Files to Modify
- `tui/runtime/node.go` - Add separate dirty flags
- `tui/runtime/layout/engine.go` - Use layoutDirty flag
- `tui/runtime/render/renderer.go` - Use paintDirty flag

### Implementation

**1. Update `LayoutNode` in `tui/runtime/node.go`:**

```go
type LayoutNode struct {
    // ... existing fields ...

    // Dirty flags - separated for optimization
    layoutDirty bool  // Needs measure/layout phase
    paintDirty  bool  // Needs render phase only

    // cacheKey is used for measurement caching
    cacheKey string
}
```

**2. Add new methods:**

```go
// MarkLayoutDirty marks this node as needing layout
// Layout dirtiness propagates to ancestors
func (n *LayoutNode) MarkLayoutDirty() {
    if n == nil || n.layoutDirty {
        return
    }
    n.layoutDirty = true
    // Size change may affect parent layout
    if n.Parent != nil {
        n.Parent.MarkLayoutDirty()
    }
}

// MarkPaintDirty marks this node as needing repaint only
// Paint dirtiness does NOT propagate to ancestors
func (n *LayoutNode) MarkPaintDirty() {
    if n == nil {
        return
    }
    n.paintDirty = true
}

// IsLayoutDirty returns true if node needs layout
func (n *LayoutNode) IsLayoutDirty() bool {
    return n != nil && n.layoutDirty
}

// IsPaintDirty returns true if node needs paint
func (n *LayoutNode) IsPaintDirty() bool {
    return n != nil && n.paintDirty
}

// ClearLayoutDirty clears the layout dirty flag
func (n *LayoutNode) ClearLayoutDirty() {
    if n != nil {
        n.layoutDirty = false
    }
}

// ClearPaintDirty clears the paint dirty flag
func (n *LayoutNode) ClearPaintDirty() {
    if n != nil {
        n.paintDirty = false
    }
}

// ClearDirty clears both dirty flags
func (n *LayoutNode) ClearDirty() {
    if n != nil {
        n.layoutDirty = false
        n.paintDirty = false
    }
}
```

**3. Keep legacy `MarkDirty()` for backward compatibility:**

```go
// MarkDirty marks node as both layout and paint dirty
// This is the "conservative" default when unsure
func (n *LayoutNode) MarkDirty() {
    if n == nil {
        return
    }
    n.layoutDirty = true
    n.paintDirty = true
    for _, child := range n.Children {
        child.MarkDirty()
    }
}
```

---

## Phase 2: Automatic Dependency Graph

### Goal
When state changes, automatically mark dependent components dirty without manual subscriptions.

### Files to Create
- `tui/framework/binding/dependency.go` - Dependency graph tracker

### Files to Modify
- `tui/framework/binding/store.go` - Integrate with dependency tracking
- `tui/framework/binding/prop.go` - Register dependencies during Resolve
- `tui/runtime/node.go` - Add component reference for dependency registration

### Implementation

**1. Create `tui/framework/binding/dependency.go`:**

```go
package binding

import (
    "sync"
)

// DependencyGraph tracks state → component dependencies
type DependencyGraph struct {
    mu   sync.RWMutex
    // deps maps state key → list of dependent node IDs
    deps map[string][]string
    // reverse maps node ID → list of state keys it depends on
    reverse map[string][]string
}

// NewDependencyGraph creates a new dependency graph
func NewDependencyGraph() *DependencyGraph {
    return &DependencyGraph{
        deps:    make(map[string][]string),
        reverse: make(map[string][]string),
    }
}

// Register registers a dependency: nodeID depends on stateKey
func (g *DependencyGraph) Register(nodeID, stateKey string) {
    g.mu.Lock()
    defer g.mu.Unlock()

    // Add to deps map
    g.deps[stateKey] = append(g.deps[stateKey], nodeID)

    // Add to reverse map
    g.reverse[nodeID] = append(g.reverse[nodeID], stateKey)
}

// Unregister removes all dependencies for a node
func (g *DependencyGraph) Unregister(nodeID string) {
    g.mu.Lock()
    defer g.mu.Unlock()

    // Get all keys this node depends on
    keys, ok := g.reverse[nodeID]
    if !ok {
        return
    }

    // Remove from deps map
    for _, key := range keys {
        nodes := g.deps[key]
        for i, id := range nodes {
            if id == nodeID {
                g.deps[key] = append(nodes[:i], nodes[i+1:]...)
                break
            }
        }
    }

    // Remove from reverse map
    delete(g.reverse, nodeID)
}

// GetDependents returns all node IDs that depend on the given state key
func (g *DependencyGraph) GetDependents(stateKey string) []string {
    g.mu.RLock()
    defer g.mu.RUnlock()

    deps, ok := g.deps[stateKey]
    if !ok {
        return nil
    }

    result := make([]string, len(deps))
    copy(result, deps)
    return result
}

// GetDependencies returns all state keys that a node depends on
func (g *DependencyGraph) GetDependencies(nodeID string) []string {
    g.mu.RLock()
    defer g.mu.RUnlock()

    deps, ok := g.reverse[nodeID]
    if !ok {
        return nil
    }

    result := make([]string, len(deps))
    copy(result, deps)
    return result
}
```

**2. Update `ReactiveStore` in `tui/framework/binding/store.go`:**

```go
type ReactiveStore struct {
    mu        sync.RWMutex
    data      map[string]interface{}
    observers map[string][]Notifier
    global    []Notifier
    enabled   bool

    // NEW: Dependency graph
    depGraph  *DependencyGraph

    // ... existing batch fields ...
}

func NewReactiveStore() *ReactiveStore {
    return &ReactiveStore{
        // ... existing init ...
        depGraph: NewDependencyGraph(),
    }
}

// SetWithZone sets a value and marks dependent nodes dirty
func (s *ReactiveStore) SetWithZone(path string, value interface{}, zone StateZone) {
    s.mu.Lock()

    oldValue, exists := s.data[path]
    s.data[path] = value

    // Get dependents BEFORE releasing lock
    dependents := s.depGraph.GetDependents(path)

    s.mu.Unlock()

    changed := !exists || oldValue != value

    if changed && s.enabled {
        // Notify observers
        s.notify(path, oldValue, value)

        // NEW: Mark dependent nodes dirty via callback
        for _, nodeID := range dependents {
            if s.dirtyCallback != nil {
                s.dirtyCallback(nodeID, zone)
            }
        }
    }
}

// dirtyCallback is called when a dependent node should be marked dirty
type DirtyCallback func(nodeID string, zone StateZone)

func (s *ReactiveStore) SetDirtyCallback(cb DirtyCallback) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.dirtyCallback = cb
}

// RegisterDependency registers a node's dependency on a state key
func (s *ReactiveStore) RegisterDependency(nodeID, stateKey string) {
    s.mu.Lock()
    defer s.mu.Unlock()
    s.depGraph.Register(nodeID, stateKey)
}
```

**3. Update `Prop` in `tui/framework/binding/prop.go`:**

```go
// ResolveWithTracking resolves the property and tracks dependencies
func (p *Prop[T]) ResolveWithTracking(ctx Context, nodeID string, tracker *DependencyGraph) T {
    // First, register dependencies
    deps := p.GetDependencies()
    if deps != nil && tracker != nil {
        for _, dep := range deps {
            tracker.Register(nodeID, dep)
        }
    }

    // Then resolve normally
    return p.Resolve(ctx)
}
```

---

## Phase 3: Priority Scheduling System

### Goal
Process dirty nodes by priority level with time budget per frame, ensuring high-priority updates (input) are never blocked by low-priority updates (logs).

### Files to Create
- `tui/runtime/priority/types.go` - Priority and zone types
- `tui/runtime/priority/scheduler.go` - Time-sliced scheduler

### Files to Modify
- `tui/runtime/node.go` - Add priority field

### Implementation

**1. Create `tui/runtime/priority/types.go`:**

```go
package priority

// DirtyLevel represents the priority of a dirty node
type DirtyLevel int

const (
    DirtyLow DirtyLevel = iota    // Background logs
    DirtyNormal                    // Data tables
    DirtyHigh                      // Input focus
)

// StateZone represents the logical zone of a state change
type StateZone int

const (
    ZoneUI StateZone = iota     // Input, focus
    ZoneData                    // Tables, lists
    ZoneBackground              // Logs, monitoring
)

// ZoneToLevel maps StateZone to DirtyLevel
func ZoneToLevel(zone StateZone) DirtyLevel {
    switch zone {
    case ZoneUI:
        return DirtyHigh
    case ZoneData:
        return DirtyNormal
    case ZoneBackground:
        return DirtyLow
    default:
        return DirtyNormal
    }
}
```

**2. Update `LayoutNode` in `tui/runtime/node.go`:**

```go
type LayoutNode struct {
    // ... existing fields ...

    // Priority level for time-sliced rendering
    priority DirtyLevel

    // ... existing fields ...
}

func NewLayoutNode(id string, nodeType NodeType, style Style) *LayoutNode {
    return &LayoutNode{
        // ... existing init ...
        priority: DirtyNormal, // Default priority
    }
}

// SetPriority sets the node's priority level
func (n *LayoutNode) SetPriority(level DirtyLevel) {
    if n != nil {
        n.priority = level
    }
}

// GetPriority returns the node's priority level
func (n *LayoutNode) GetPriority() DirtyLevel {
    if n == nil {
        return DirtyNormal
    }
    return n.priority
}
```

**3. Create `tui/runtime/priority/scheduler.go`:**

```go
package priority

import (
    "time"
    "github.com/yaoapp/yao/tui/runtime"
)

// Scheduler manages priority-based rendering with time slicing
type Scheduler struct {
    defaultBudget time.Duration
}

// NewScheduler creates a new scheduler
func NewScheduler() *Scheduler {
    return &Scheduler{
        defaultBudget: 2 * time.Millisecond,
    }
}

// RenderFrame renders dirty nodes by priority with time budget
func (s *Scheduler) RenderFrame(root *runtime.LayoutNode, renderer Renderer) FrameResult {
    result := FrameResult{}

    // Process in priority order
    budgets := []struct {
        level  DirtyLevel
        budget time.Duration
    }{
        {DirtyHigh, s.defaultBudget},
        {DirtyNormal, s.defaultBudget},
        {DirtyLow, s.defaultBudget},
    }

    for _, b := range budgets {
        if s.hasDirtyAtLevel(root, b.level) {
            result = s.processDirty(root, b.level, b.budget, renderer)
            // If we ran out of budget, defer remaining to next frame
            if result.OutOfTime {
                break
            }
        }
    }

    return result
}

// processDirty processes all dirty nodes at a given level
func (s *Scheduler) processDirty(root *runtime.LayoutNode, level DirtyLevel, budget time.Duration, renderer Renderer) FrameResult {
    start := time.Now()
    result := FrameResult{}

    // Collect dirty nodes at this level
    dirtyNodes := s.collectDirtyByLevel(root, level)

    for _, node := range dirtyNodes {
        // Check time budget
        if time.Since(start) > budget {
            result.OutOfTime = true
            break
        }

        // Layout if needed
        if node.IsLayoutDirty() {
            renderer.Layout(node)
            node.ClearLayoutDirty()
        }

        // Paint if needed
        if node.IsPaintDirty() {
            renderer.Paint(node)
            node.ClearPaintDirty()
        }

        result.ProcessedCount++
    }

    return result
}

// collectDirtyByLevel collects all dirty nodes at a given priority level
func (s *Scheduler) collectDirtyByLevel(root *runtime.LayoutNode, level DirtyLevel) []*runtime.LayoutNode {
    var result []*runtime.LayoutNode

    var traverse func(*runtime.LayoutNode)
    traverse = func(node *runtime.LayoutNode) {
        if node == nil {
            return
        }

        // Check if node matches priority and is dirty
        if node.GetPriority() == level && (node.IsLayoutDirty() || node.IsPaintDirty()) {
            result = append(result, node)
        }

        // Traverse children
        for _, child := range node.Children {
            traverse(child)
        }
    }

    traverse(root)
    return result
}

// hasDirtyAtLevel checks if there are any dirty nodes at a given level
func (s *Scheduler) hasDirtyAtLevel(root *runtime.LayoutNode, level DirtyLevel) bool {
    if root == nil {
        return false
    }

    if root.GetPriority() == level && (root.IsLayoutDirty() || root.IsPaintDirty()) {
        return true
    }

    for _, child := range root.Children {
        if s.hasDirtyAtLevel(child, level) {
            return true
        }
    }

    return false
}

// Renderer is the interface for layout and rendering
type Renderer interface {
    Layout(node *runtime.LayoutNode)
    Paint(node *runtime.LayoutNode)
}

// FrameResult represents the result of a frame render
type FrameResult struct {
    ProcessedCount int
    OutOfTime      bool
}
```

---

## Phase 4: GPU-Style Command Batching

### Goal
Reduce terminal IO by batching draw commands and minimizing VT code changes.

### Files to Create
- `tui/runtime/paint/batch.go` - Command batching system
- `tui/runtime/paint/style_state.go` - Style state machine

### Files to Modify
- `tui/runtime/runtime.go` - Use command batching

### Implementation

**1. Create `tui/runtime/paint/batch.go`:**

```go
package paint

import (
    "bytes"
    "github.com/yaoapp/yao/tui/runtime"
)

// DrawCmd represents a single drawing command
type DrawCmd struct {
    X, Y  int
    Text  string
    Style runtime.CellStyle
}

// CommandBatch batches draw commands to minimize terminal IO
type CommandBatch struct {
    cmds    []DrawCmd
    styleVM *StyleStateMachine
}

// NewCommandBatch creates a new command batch
func NewCommandBatch() *CommandBatch {
    return &CommandBatch{
        cmds:    make([]DrawCmd, 0, 256),
        styleVM: NewStyleStateMachine(),
    }
}

// Add adds a draw command
func (b *CommandBatch) Add(x, y int, text string, style runtime.CellStyle) {
    b.cmds = append(b.cmds, DrawCmd{
        X:     x,
        Y:     y,
        Text:  text,
        Style: style,
    })
}

// AddCell adds a single cell command
func (b *CommandBatch) AddCell(x, y int, char rune, style runtime.CellStyle) {
    b.Add(x, y, string(char), style)
}

// Flush merges commands and generates the final output
func (b *CommandBatch) Flush() string {
    if len(b.cmds) == 0 {
        return ""
    }

    var buf bytes.Buffer
    b.styleVM.Reset()

    // Sort by Y then X for linear traversal
    b.sortCommands()

    // Merge adjacent commands with same style
    merged := b.mergeCommands()

    // Generate output with style state machine
    lastX, lastY := -1, -1
    for _, cmd := range merged {
        // Move cursor if needed
        if cmd.X != lastX || cmd.Y != lastY {
            buf.WriteString(b.moveCursor(cmd.X, cmd.Y))
            lastX, lastY = cmd.X, cmd.Y
        }

        // Apply style if changed
        if b.styleVM.NeedsUpdate(cmd.Style) {
            buf.WriteString(b.styleVM.Update(cmd.Style))
        }

        // Write text
        buf.WriteString(cmd.Text)
        lastX += len(cmd.Text)
    }

    // Reset style at end
    buf.WriteString("\x1b[0m")

    return buf.String()
}

// mergeCommands merges adjacent commands that can be combined
func (b *CommandBatch) mergeCommands() []DrawCmd {
    if len(b.cmds) == 0 {
        return nil
    }

    merged := make([]DrawCmd, 0, len(b.cmds))
    current := b.cmds[0]

    for i := 1; i < len(b.cmds); i++ {
        next := b.cmds[i]

        // Check if we can merge
        if b.canMerge(current, next) {
            current.Text += next.Text
        } else {
            merged = append(merged, current)
            current = next
        }
    }

    merged = append(merged, current)
    return merged
}

// canMerge checks if two commands can be merged
func (b *CommandBatch) canMerge(a, b DrawCmd) bool {
    // Must be on same line
    if a.Y != b.Y {
        return false
    }

    // Must be adjacent
    if a.X+len(a.Text) != b.X {
        return false
    }

    // Must have same style
    return a.Style == b.Style
}

// sortCommands sorts commands by Y then X
func (b *CommandBatch) sortCommands() {
    // Simple insertion sort for small batches
    for i := 1; i < len(b.cmds); i++ {
        for j := i; j > 0; j-- {
            if b.less(b.cmds[j], b.cmds[j-1]) {
                b.cmds[j], b.cmds[j-1] = b.cmds[j-1], b.cmds[j]
            } else {
                break
            }
        }
    }
}

// less compares two commands
func (b *CommandBatch) less(a, b DrawCmd) bool {
    if a.Y != b.Y {
        return a.Y < b.Y
    }
    return a.X < b.X
}

// moveCursor generates ANSI cursor movement
func (b *CommandBatch) moveCursor(x, y int) string {
    return "\x1b[" + itoa(y+1) + ";" + itoa(x+1) + "H"
}

// itoa converts int to string (faster than strconv)
func itoa(i int) string {
    if i == 0 {
        return "0"
    }
    var buf [20]byte
    var n int
    for i > 0 {
        buf[n] = byte('0' + i%10)
        i /= 10
        n++
    }
    // Reverse
    for i := 0; i < n/2; i++ {
        buf[i], buf[n-1-i] = buf[n-1-i], buf[i]
    }
    return string(buf[:n])
}

// Clear clears all commands
func (b *CommandBatch) Clear() {
    b.cmds = b.cmds[:0]
}

// Count returns the number of commands
func (b *CommandBatch) Count() int {
    return len(b.cmds)
}
```

**2. Create `tui/runtime/paint/style_state.go`:**

```go
package paint

import (
    "github.com/yaoapp/yao/tui/runtime"
)

// StyleStateMachine minimizes VT code changes by tracking current style
type StyleStateMachine struct {
    current runtime.CellStyle
}

// NewStyleStateMachine creates a new style state machine
func NewStyleStateMachine() *StyleStateMachine {
    return &StyleStateMachine{
        current: runtime.CellStyle{},
    }
}

// Reset resets the current style
func (s *StyleStateMachine) Reset() {
    s.current = runtime.CellStyle{}
}

// NeedsUpdate checks if the style needs to be updated
func (s *StyleStateMachine) NeedsUpdate(style runtime.CellStyle) bool {
    return s.current != style
}

// Update generates VT codes to transition to the new style
func (s *StyleStateMachine) Update(style runtime.CellStyle) string {
    codes := s.buildDiffCodes(s.current, style)
    s.current = style
    return codes
}

// buildDiffCodes builds only the codes that need to change
func (s *StyleStateMachine) buildDiffCodes(from, to runtime.CellStyle) string {
    if from == to {
        return ""
    }

    // If completely different, just emit full style
    if from == (runtime.CellStyle{}) {
        return s.fullStyle(to)
    }

    // Otherwise, emit only changes
    var codes []string

    // Check for reset needed
    if to == (runtime.CellStyle{}) {
        return "\x1b[0m"
    }

    // Bold
    if from.Bold != to.Bold {
        if to.Bold {
            codes = append(codes, "1")
        }
    }

    // Underline
    if from.Underline != to.Underline {
        if to.Underline {
            codes = append(codes, "4")
        }
    }

    // Italic
    if from.Italic != to.Italic {
        if to.Italic {
            codes = append(codes, "3")
        }
    }

    // Reverse
    if from.Reverse != to.Reverse {
        if to.Reverse {
            codes = append(codes, "7")
        }
    }

    // Colors
    if from.Foreground != to.Foreground && to.Foreground != "" {
        codes = append(codes, colorCode(to.Foreground, false))
    }
    if from.Background != to.Background && to.Background != "" {
        codes = append(codes, colorCode(to.Background, true))
    }

    if len(codes) == 0 {
        return ""
    }

    return "\x1b[" + joinCodes(codes) + "m"
}

// fullStyle generates full style codes
func (s *StyleStateMachine) fullStyle(style runtime.CellStyle) string {
    var codes []string

    if style.Bold {
        codes = append(codes, "1")
    }
    if style.Italic {
        codes = append(codes, "3")
    }
    if style.Underline {
        codes = append(codes, "4")
    }
    if style.Reverse {
        codes = append(codes, "7")
    }
    if style.Foreground != "" {
        codes = append(codes, colorCode(style.Foreground, false))
    }
    if style.Background != "" {
        codes = append(codes, colorCode(style.Background, true))
    }

    if len(codes) == 0 {
        return ""
    }

    return "\x1b[" + joinCodes(codes) + "m"
}

// colorCode converts a color to ANSI code
func colorCode(color string, isBackground bool) string {
    // ... existing colorToANSICode logic ...
    return ""
}

// joinCodes joins ANSI codes with semicolons
func joinCodes(codes []string) string {
    result := ""
    for i, c := range codes {
        if i > 0 {
            result += ";"
        }
        result += c
    }
    return result
}
```

---

## Phase 5: Independent Layer Buffers

### Goal
Create independent refresh regions for high-frequency updates (logs, streaming data) that don't affect static content (menus, borders).

### Files to Create
- `tui/runtime/paint/layer.go` - Layer system
- `tui/runtime/paint/compositor.go` - Layer composition

### Implementation

**1. Create `tui/runtime/paint/layer.go`:**

```go
package paint

import (
    "github.com/yaoapp/yao/tui/runtime"
)

// LayerType represents the type of layer
type LayerType int

const (
    LayerBackground LayerType = iota // Static UI (menus, borders)
    LayerContent                     // Forms, tables
    LayerStream                      // High-frequency logs
    LayerOverlay                     // Modals, popups
)

// Layer represents an independent rendering layer
type Layer struct {
    ID       string
    Type     LayerType
    ZIndex   int
    Buffer   *runtime.CellBuffer
    Dirty    bool
    Rect     runtime.Rect
    Enabled  bool
}

// NewLayer creates a new layer
func NewLayer(id string, layerType LayerType, zIndex int, width, height int) *Layer {
    return &Layer{
        ID:      id,
        Type:    layerType,
        ZIndex:  zIndex,
        Buffer:  runtime.NewCellBuffer(width, height),
        Dirty:   true,
        Enabled: true,
        Rect: runtime.Rect{
            X:      0,
            Y:      0,
            Width:  width,
            Height: height,
        },
    }
}

// MarkDirty marks the layer as dirty
func (l *Layer) MarkDirty() {
    l.Dirty = true
}

// ClearDirty clears the dirty flag
func (l *Layer) ClearDirty() {
    l.Dirty = false
}

// IsDirty returns true if the layer needs rendering
func (l *Layer) IsDirty() bool {
    return l.Enabled && l.Dirty
}

// SetRect sets the layer's rectangle
func (l *Layer) SetRect(rect runtime.Rect) {
    l.Rect = rect
    // Resize buffer if needed
    if rect.Width != l.Buffer.Width() || rect.Height != l.Buffer.Height() {
        l.Buffer = runtime.NewCellBuffer(rect.Width, rect.Height)
        l.MarkDirty()
    }
}
```

**2. Create `tui/runtime/paint/compositor.go`:**

```go
package paint

import (
    "bytes"
    "sort"

    "github.com/yaoapp/yao/tui/runtime"
)

// Compositor manages multiple layers and composites them
type Compositor struct {
    layers  []*Layer
    width   int
    height  int
}

// NewCompositor creates a new compositor
func NewCompositor(width, height int) *Compositor {
    return &Compositor{
        layers:  make([]*Layer, 0, 4),
        width:   width,
        height: height,
    }
}

// AddLayer adds a layer to the compositor
func (c *Compositor) AddLayer(layer *Layer) {
    c.layers = append(c.layers, layer)
    c.sortLayers()
}

// RemoveLayer removes a layer by ID
func (c *Compositor) RemoveLayer(id string) {
    for i, layer := range c.layers {
        if layer.ID == id {
            c.layers = append(c.layers[:i], c.layers[i+1:]...)
            return
        }
    }
}

// GetLayer returns a layer by ID
func (c *Compositor) GetLayer(id string) *Layer {
    for _, layer := range c.layers {
        if layer.ID == id {
            return layer
        }
    }
    return nil
}

// sortLayers sorts layers by Z-index
func (c *Compositor) sortLayers() {
    sort.Slice(c.layers, func(i, j int) bool {
        return c.layers[i].ZIndex < c.layers[j].ZIndex
    })
}

// RenderDirty renders only dirty layers and returns the composited output
func (c *Compositor) RenderDirty() string {
    var buf bytes.Buffer

    for _, layer := range c.layers {
        if !layer.IsDirty() {
            continue
        }

        // Output layer content
        buf.WriteString(c.renderLayer(layer))

        layer.ClearDirty()
    }

    return buf.String()
}

// renderLayer renders a single layer with region scrolling
func (c *Compositor) renderLayer(layer *Layer) string {
    // Set scroll region if applicable
    var output string

    // For stream layers, use scroll optimization
    if layer.Type == LayerStream {
        output = "\x1b[" + itoa(layer.Rect.Y+1) + ";" +
                 itoa(layer.Rect.Y+layer.Rect.Height) + "r"
    }

    // Output layer buffer content
    output += layer.Buffer.String()

    // Reset scroll region
    if layer.Type == LayerStream {
        output += "\x1b[r"
    }

    return output
}

// Composite creates a composite buffer from all layers
func (c *Compositor) Composite() *runtime.CellBuffer {
    buffer := runtime.NewCellBuffer(c.width, c.height)

    for _, layer := range c.layers {
        if !layer.Enabled {
            continue
        }

        c.blitLayer(buffer, layer)
    }

    return buffer
}

// blitLayer blits a layer onto the composite buffer
func (c *Compositor) blitLayer(dst *runtime.CellBuffer, src *Layer) {
    for y := 0; y < src.Rect.Height; y++ {
        for x := 0; x < src.Rect.Width; x++ {
            srcX := src.Rect.X + x
            srcY := src.Rect.Y + y

            if srcX >= dst.Width() || srcY >= dst.Height() {
                continue
            }

            cell := src.Buffer.GetCell(x, y)
            dst.SetContent(srcX, srcY, src.ZIndex, cell.Char, cell.Style, cell.NodeID)
        }
    }
}

// MarkAllDirty marks all layers as dirty
func (c *Compositor) MarkAllDirty() {
    for _, layer := range c.layers {
        layer.MarkDirty()
    }
}

// Resize handles window resize
func (c *Compositor) Resize(width, height int) {
    c.width = width
    c.height = height

    for _, layer := range c.layers {
        if layer.Rect.Width > width {
            layer.Rect.Width = width
        }
        if layer.Rect.Height > height {
            layer.Rect.Height = height
        }
        layer.SetRect(layer.Rect)
    }
}
```

---

## Phase 6: Remote Rendering Optimization

### Goal
Optimize for SSH/remote usage with frame buffering and delta encoding.

### Files to Create
- `tui/runtime/paint/remote.go` - Remote optimization

### Implementation

**1. Create `tui/runtime/paint/remote.go`:**

```go
package paint

import (
    "bytes"
    "time"
)

// RemoteOptimizer optimizes rendering for remote terminals
type RemoteOptimizer struct {
    frameBuffer    *bytes.Buffer
    lastFlush      time.Time
    frameInterval  time.Duration
    deltaEncoding  bool
}

// NewRemoteOptimizer creates a new remote optimizer
func NewRemoteOptimizer() *RemoteOptimizer {
    return &RemoteOptimizer{
        frameBuffer:   &bytes.Buffer{},
        frameInterval: 16 * time.Millisecond, // ~60 FPS
        deltaEncoding: true,
    }
}

// BufferFrame buffers a frame for later flushing
func (r *RemoteOptimizer) BufferFrame(data []byte) {
    r.frameBuffer.Write(data)
}

// ShouldFlush returns true if enough time has passed to flush
func (r *RemoteOptimizer) ShouldFlush() bool {
    return time.Since(r.lastFlush) >= r.frameInterval || r.frameBuffer.Len() > 4096
}

// Flush flushes the buffered frame data
func (r *RemoteOptimizer) Flush() []byte {
    data := r.frameBuffer.Bytes()
    r.frameBuffer.Reset()
    r.lastFlush = time.Now()
    return data
}

// EncodeDelta encodes the difference between two frames
func (r *RemoteOptimizer) EncodeDelta(prev, curr []byte) []byte {
    if len(prev) == 0 {
        return curr
    }

    // Simple delta encoding: find common prefix and suffix
    prefixLen := 0
    maxPrefix := min(len(prev), len(curr))
    for prefixLen < maxPrefix && prev[prefixLen] == curr[prefixLen] {
        prefixLen++
    }

    suffixLen := 0
    maxSuffix := min(len(prev)-prefixLen, len(curr)-prefixLen)
    for suffixLen < maxSuffix &&
         prev[len(prev)-1-suffixLen] == curr[len(curr)-1-suffixLen] {
        suffixLen++
    }

    // If most of the frame changed, send full frame
    changedLen := len(curr) - prefixLen - suffixLen
    if changedLen > len(curr)/2 {
        return curr
    }

    // Otherwise, send delta
    // Format: [prefix_len][changed_data][suffix_len]
    var delta bytes.Buffer
    delta.WriteByte(byte(prefixLen))
    delta.Write(curr[prefixLen : len(curr)-suffixLen])
    delta.WriteByte(byte(suffixLen))

    return delta.Bytes()
}

func min(a, b int) int {
    if a < b {
        return a
    }
    return b
}

// SetFrameInterval sets the minimum interval between frames
func (r *RemoteOptimizer) SetFrameInterval(interval time.Duration) {
    r.frameInterval = interval
}

// EnableDeltaEncoding enables or disables delta encoding
func (r *RemoteOptimizer) EnableDeltaEncoding(enabled bool) {
    r.deltaEncoding = enabled
}
```

---

## Summary of Files to Create/Modify

### New Files
| File | Purpose |
|------|---------|
| `tui/runtime/priority/types.go` | Priority and zone type definitions |
| `tui/runtime/priority/scheduler.go` | Time-sliced scheduler |
| `tui/framework/binding/dependency.go` | Dependency graph tracker |
| `tui/runtime/paint/batch.go` | Command batching |
| `tui/runtime/paint/style_state.go` | Style state machine |
| `tui/runtime/paint/layer.go` | Layer system |
| `tui/runtime/paint/compositor.go` | Layer compositor |
| `tui/runtime/paint/remote.go` | Remote optimization |

### Modified Files
| File | Changes |
|------|---------|
| `tui/runtime/node.go` | Add layoutDirty/paintDirty flags, priority field |
| `tui/framework/binding/store.go` | Integrate dependency graph, zone support |
| `tui/framework/binding/prop.go` | Add dependency tracking on resolve |

---

## Testing Strategy

1. **Unit tests** for each new component
2. **Integration tests** for the full render pipeline
3. **Performance benchmarks** to measure improvements
4. **Visual regression tests** to ensure correctness

### Test Commands
```bash
# Run all TUI tests
make unit-test-tui

# Run specific tests
go test ./tui/runtime/priority -v
go test ./tui/runtime/paint -v

# Benchmark
go test ./tui/runtime/... -bench=. -benchmem
```

---

## Verification

After implementation, verify:

1. **Layout vs Paint separation**: Components with content changes skip layout phase
2. **Automatic dirty marking**: State changes automatically mark dependent components
3. **Priority rendering**: High-priority updates complete within time budget
4. **Reduced IO**: Terminal output is batched and minimized
5. **Layer isolation**: High-frequency updates don't trigger full redraws
6. **SSH performance**: Remote usage remains responsive
