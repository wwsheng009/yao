package testing

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/yaoapp/yao/tui/runtime/action"
	"github.com/yaoapp/yao/tui/runtime/ai"
)

// =============================================================================
// Action Recorder (V3)
// =============================================================================
// Recorder 记录和回放操作
// 用于测试和调试

// Recorder 记录器
type Recorder struct {
	actions     []*action.Action
	timestamps  []time.Time
	recording   bool
	allowReplay bool
}

// NewRecorder 创建记录器
func NewRecorder() *Recorder {
	return &Recorder{
		actions:     make([]*action.Action, 0),
		timestamps:  make([]time.Time, 0),
		recording:   false,
		allowReplay: true,
	}
}

// =============================================================================
// Recording
// =============================================================================

// Record 开始记录
func (r *Recorder) Record() {
	r.recording = true
	r.actions = r.actions[:0]
	r.timestamps = r.timestamps[:0]
}

// Stop 停止记录
func (r *Recorder) Stop() {
	r.recording = false
}

// RecordAction 记录 Action
func (r *Recorder) RecordAction(a *action.Action) {
	if !r.recording {
		return
	}
	r.actions = append(r.actions, a.Clone())
	r.timestamps = append(r.timestamps, time.Now())
}

// RecordDispatch 创建一个记录分发的包装器
func (r *Recorder) RecordDispatch() action.Handler {
	return func(a *action.Action) bool {
		r.RecordAction(a)
		return false // 不实际处理，只是记录
	}
}

// GetActions 获取记录的所有 Action
func (r *Recorder) GetActions() []*action.Action {
	result := make([]*action.Action, len(r.actions))
	copy(result, r.actions)
	return result
}

// GetCount 获取记录的 Action 数量
func (r *Recorder) GetCount() int {
	return len(r.actions)
}

// Clear 清空记录
func (r *Recorder) Clear() {
	r.actions = r.actions[:0]
	r.timestamps = r.timestamps[:0]
}

// IsRecording 检查是否正在记录
func (r *Recorder) IsRecording() bool {
	return r.recording
}

// =============================================================================
// Replay
// =============================================================================

// Replay 回放记录的 Action
func (r *Recorder) Replay(ctrl *ai.RuntimeController, speed float64) error {
	return r.ReplayWithDelay(ctrl, 0, speed)
}

// ReplayWithDelay 带延迟的回放
func (r *Recorder) ReplayWithDelay(ctrl *ai.RuntimeController, initialDelay time.Duration, speed float64) error {
	if !r.allowReplay {
		return ErrReplayDisabled
	}

	if len(r.actions) == 0 {
		return ErrNoActionsToReplay
	}

	// 初始延迟
	if initialDelay > 0 {
		time.Sleep(initialDelay)
	}

	// 计算延迟间隔
	var lastTime time.Time
	for i, a := range r.actions {
		// 计算与上次 Action 的时间差
		delay := time.Duration(0)
		if i > 0 && !r.timestamps[i].IsZero() && !lastTime.IsZero() {
			originalDelay := r.timestamps[i].Sub(lastTime)
			if speed > 0 {
				delay = time.Duration(float64(originalDelay) / speed)
			} else {
				delay = originalDelay
			}
		}

		if delay > 0 {
			time.Sleep(delay)
		}

		if err := ctrl.Dispatch(a); err != nil {
			return &ReplayError{
				ActionIndex: i,
				Action:      a,
				Err:         err,
			}
		}

		lastTime = r.timestamps[i]
	}

	return nil
}

// ReplayFast 快速回放（无延迟）
func (r *Recorder) ReplayFast(ctrl *ai.RuntimeController) error {
	if !r.allowReplay {
		return ErrReplayDisabled
	}

	if len(r.actions) == 0 {
		return ErrNoActionsToReplay
	}

	for i, a := range r.actions {
		if err := ctrl.Dispatch(a); err != nil {
			return &ReplayError{
				ActionIndex: i,
				Action:      a,
				Err:         err,
			}
		}
	}

	return nil
}

// SetAllowReplay 设置是否允许回放
func (r *Recorder) SetAllowReplay(allow bool) {
	r.allowReplay = allow
}

// =============================================================================
// Import/Export
// =============================================================================

// ExportJSON 导出为 JSON
func (r *Recorder) ExportJSON(filename string) error {
	data := struct {
		Actions    []*action.Action `json:"actions"`
		Timestamps []time.Time      `json:"timestamps"`
	}{
		Actions:    r.actions,
		Timestamps: r.timestamps,
	}

	jsonData, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filename, jsonData, 0644)
}

// ImportJSON 从 JSON 导入
func (r *Recorder) ImportJSON(filename string) error {
	fileData, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	var data struct {
		Actions    []*action.Action `json:"actions"`
		Timestamps []time.Time      `json:"timestamps"`
	}

	if err := json.Unmarshal(fileData, &data); err != nil {
		return err
	}

	r.actions = data.Actions
	r.timestamps = data.Timestamps
	return nil
}

// =============================================================================
// Errors
// =============================================================================

// ErrReplayDisabled 回放被禁用
var ErrReplayDisabled = &RecorderError{Msg: "replay is disabled"}

// ErrNoActionsToReplay 没有可回放的 Action
var ErrNoActionsToReplay = &RecorderError{Msg: "no actions to replay"}

// RecorderError 记录器错误
type RecorderError struct {
	Msg string
}

func (e *RecorderError) Error() string {
	return e.Msg
}

// ReplayError 回放错误
type ReplayError struct {
	ActionIndex int
	Action      *action.Action
	Err         error
}

func (e *ReplayError) Error() string {
	return fmt.Sprintf("replay failed at action %d (%s): %v", e.ActionIndex, e.Action.Type, e.Err)
}

func (e *ReplayError) Unwrap() error {
	return e.Err
}

// =============================================================================
// Helper Types
// =============================================================================

// ActionLog Action 日志条目
type ActionLog struct {
	Action    *action.Action `json:"action"`
	Timestamp time.Time      `json:"timestamp"`
	Delay     time.Duration  `json:"delay,omitempty"` // 与上一个 Action 的延迟
}

// GetActionLog 获取 Action 日志（带延迟）
func (r *Recorder) GetActionLog() []ActionLog {
	if len(r.actions) == 0 {
		return nil
	}

	log := make([]ActionLog, len(r.actions))
	for i, a := range r.actions {
		var delay time.Duration
		if i > 0 && !r.timestamps[i].IsZero() && !r.timestamps[i-1].IsZero() {
			delay = r.timestamps[i].Sub(r.timestamps[i-1])
		}

		log[i] = ActionLog{
			Action:    a,
			Timestamp: r.timestamps[i],
			Delay:     delay,
		}
	}

	return log
}

// Summary 获取记录摘要
func (r *Recorder) Summary() map[string]interface{} {
	actionCounts := make(map[string]int)
	for _, a := range r.actions {
		actionCounts[string(a.Type)]++
	}

	var duration time.Duration
	if len(r.timestamps) >= 2 {
		duration = r.timestamps[len(r.timestamps)-1].Sub(r.timestamps[0])
	}

	return map[string]interface{}{
		"total_actions": len(r.actions),
		"duration_ms":   duration.Milliseconds(),
		"action_counts": actionCounts,
		"recording":     r.recording,
		"allow_replay":  r.allowReplay,
	}
}
