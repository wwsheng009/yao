package event

// EventPhase represents the phase of event propagation.
// This follows the DOM event model with three distinct phases.
//
// The propagation order is:
//   1. Capture (PhaseCapture): Root → Target
//   2. Target (PhaseTarget): At the target component
//   3. Bubble (PhaseBubble): Target → Root
type EventPhase int

const (
	// PhaseNone indicates no phase (initial state)
	PhaseNone EventPhase = iota

	// PhaseCapture is the capturing phase where events propagate
	// from the root down to the target component.
	// This is the first phase and allows ancestors to intercept
	// events before they reach the target.
	PhaseCapture

	// PhaseTarget is when the event has reached the target component.
	// This is the second phase where the target component can
	// handle the event.
	PhaseTarget

	// PhaseBubble is the bubbling phase where events propagate
	// from the target back up to the root.
	// This is the third phase and allows ancestors to handle
	// events that weren't stopped by the target.
	PhaseBubble
)

// String returns the string representation of the event phase.
func (p EventPhase) String() string {
	switch p {
	case PhaseNone:
		return "None"
	case PhaseCapture:
		return "Capture"
	case PhaseTarget:
		return "Target"
	case PhaseBubble:
		return "Bubble"
	default:
		return "Unknown"
	}
}

// IsCapture returns true if the phase is PhaseCapture.
func (p EventPhase) IsCapture() bool {
	return p == PhaseCapture
}

// IsTarget returns true if the phase is PhaseTarget.
func (p EventPhase) IsTarget() bool {
	return p == PhaseTarget
}

// IsBubble returns true if the phase is PhaseBubble.
func (p EventPhase) IsBubble() bool {
	return p == PhaseBubble
}

// IsValid returns true if the phase is one of the valid propagation phases.
func (p EventPhase) IsValid() bool {
	return p >= PhaseCapture && p <= PhaseBubble
}
