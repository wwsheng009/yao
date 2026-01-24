package animation

import (
	"time"
)

// AnimationType defines the type of animation.
type AnimationType string

const (
	AnimationFade     AnimationType = "fade"
	AnimationSlide    AnimationType = "slide"
	AnimationScale    AnimationType = "scale"
	AnimationRotate   AnimationType = "rotate"
	AnimationProgress AnimationType = "progress"
	AnimationCustom   AnimationType = "custom"
)

// SlideDirection defines the direction for slide animations.
type SlideDirection string

const (
	SlideUp    SlideDirection = "up"
	SlideDown  SlideDirection = "down"
	SlideLeft  SlideDirection = "left"
	SlideRight SlideDirection = "right"
)

// AnimationState represents the current state of an animation.
type AnimationState string

const (
	AnimationStateIdle     AnimationState = "idle"
	AnimationStateRunning  AnimationState = "running"
	AnimationStatePaused   AnimationState = "paused"
	AnimationStateCompleted AnimationState = "completed"
	AnimationStateCancelled AnimationState = "cancelled"
)

// Animation represents a single animation.
type Animation struct {
	ID          string        // Unique identifier
	Type        AnimationType // Type of animation
	Duration    time.Duration // Total duration
	Elapsed     time.Duration // Time elapsed
	Easing      EasingFunction
	State       AnimationState
	Delay       time.Duration // Start delay
	Repeat      int           // Number of times to repeat (-1 for infinite)
	RepeatDelay time.Duration // Delay between repeats
	Alternate   bool          // Alternate direction on repeat (ping-pong)

	// Progress callback - called on each frame
	OnProgress func(progress float64)

	// Complete callback - called when animation completes
	OnComplete func()

	// Direction (for slide animations)
	SlideDir SlideDirection

	// Start and end values for interpolation
	From interface{}
	To   interface{}

	// Current value (updated during animation)
	Current interface{}

	// Whether animation is reversible
	Reversible bool
}

// NewAnimation creates a new animation.
func NewAnimation(id string, animType AnimationType, duration time.Duration) *Animation {
	return &Animation{
		ID:       id,
		Type:     animType,
		Duration: duration,
		Easing:   Linear,
		State:    AnimationStateIdle,
		Repeat:   0, // No repeat by default
	}
}

// WithEasing sets the easing function.
func (a *Animation) WithEasing(easing EasingFunction) *Animation {
	a.Easing = easing
	return a
}

// WithEasingName sets the easing function by name.
func (a *Animation) WithEasingName(name string) *Animation {
	a.Easing = GetEasingFunction(name)
	return a
}

// WithDelay sets the start delay.
func (a *Animation) WithDelay(delay time.Duration) *Animation {
	a.Delay = delay
	return a
}

// WithRepeat sets the repeat count (-1 for infinite).
func (a *Animation) WithRepeat(repeat int) *Animation {
	a.Repeat = repeat
	return a
}

// WithRepeatDelay sets the delay between repeats.
func (a *Animation) WithRepeatDelay(delay time.Duration) *Animation {
	a.RepeatDelay = delay
	return a
}

// WithAlternate enables alternating direction on repeat.
func (a *Animation) WithAlternate(alternate bool) *Animation {
	a.Alternate = alternate
	return a
}

// WithOnProgress sets the progress callback.
func (a *Animation) WithOnProgress(callback func(float64)) *Animation {
	a.OnProgress = callback
	return a
}

// WithOnComplete sets the complete callback.
func (a *Animation) WithOnComplete(callback func()) *Animation {
	a.OnComplete = callback
	return a
}

// WithSlideDirection sets the slide direction.
func (a *Animation) WithSlideDirection(dir SlideDirection) *Animation {
	a.SlideDir = dir
	return a
}

// WithFromTo sets the start and end values.
func (a *Animation) WithFromTo(from, to interface{}) *Animation {
	a.From = from
	a.To = to
	a.Current = from
	return a
}

// WithReversible marks the animation as reversible.
func (a *Animation) WithReversible(reversible bool) *Animation {
	a.Reversible = reversible
	return a
}

// GetProgress returns the current progress (0.0 to 1.0).
func (a *Animation) GetProgress() float64 {
	if a.Duration == 0 {
		return 1.0
	}
	progress := float64(a.Elapsed) / float64(a.Duration)
	if progress < 0 {
		return 0
	}
	if progress > 1 {
		return 1
	}
	return progress
}

// GetEasedProgress returns the eased progress.
func (a *Animation) GetEasedProgress() float64 {
	progress := a.GetProgress()
	if a.Easing != nil {
		return a.Easing(progress)
	}
	return progress
}

// IsFinished returns true if the animation has completed.
func (a *Animation) IsFinished() bool {
	return a.Elapsed >= a.Duration
}

// Reset resets the animation to its initial state.
func (a *Animation) Reset() {
	a.Elapsed = 0
	a.State = AnimationStateIdle
	a.Current = a.From
}

// Clone creates a copy of the animation.
func (a *Animation) Clone() *Animation {
	clone := &Animation{
		ID:          a.ID + "-clone",
		Type:        a.Type,
		Duration:    a.Duration,
		Elapsed:     0,
		Easing:      a.Easing,
		State:       AnimationStateIdle,
		Delay:       a.Delay,
		Repeat:      a.Repeat,
		RepeatDelay: a.RepeatDelay,
		Alternate:   a.Alternate,
		OnProgress:  a.OnProgress,
		OnComplete:  a.OnComplete,
		SlideDir:    a.SlideDir,
		From:        a.From,
		To:          a.To,
		Current:     a.From,
		Reversible:  a.Reversible,
	}
	return clone
}

// Transition represents a transition between two states.
type Transition struct {
	Animation *Animation
	From      interface{}
	To        interface{}
}

// NewTransition creates a new transition.
func NewTransition(id string, from, to interface{}, duration time.Duration) *Transition {
	anim := NewAnimation(id, AnimationCustom, duration).
		WithFromTo(from, to)

	return &Transition{
		Animation: anim,
		From:      from,
		To:        to,
	}
}

// WithEasing sets the easing function for the transition.
func (t *Transition) WithEasing(easing EasingFunction) *Transition {
	t.Animation.WithEasing(easing)
	return t
}

// WithEasingName sets the easing function by name.
func (t *Transition) WithEasingName(name string) *Transition {
	t.Animation.WithEasingName(name)
	return t
}

// FadeTransition creates a fade in/out transition.
func FadeTransition(id string, fromOpacity, toOpacity float64, duration time.Duration) *Transition {
	return NewTransition(id, fromOpacity, toOpacity, duration).
		WithEasingName("ease-out-quad")
}

// SlideTransition creates a slide transition.
func SlideTransition(id string, dir SlideDirection, duration time.Duration) *Transition {
	anim := NewAnimation(id, AnimationSlide, duration).
		WithSlideDirection(dir).
		WithEasingName("ease-out-quad")

	return &Transition{
		Animation: anim,
	}
}

// ScaleTransition creates a scale transition.
func ScaleTransition(id string, fromScale, toScale float64, duration time.Duration) *Transition {
	return NewTransition(id, fromScale, toScale, duration).
		WithEasingName("ease-out-back")
}
