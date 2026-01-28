package priority

// DirtyLevel represents the priority of a dirty node
type DirtyLevel int

const (
	DirtyLow DirtyLevel = iota // Background logs
	DirtyNormal                 // Data tables
	DirtyHigh                   // Input focus
)

// StateZone represents the logical zone of a state change
type StateZone int

const (
	ZoneUI StateZone = iota // Input, focus
	ZoneData                // Tables, lists
	ZoneBackground          // Logs, monitoring
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

// String returns the string representation of DirtyLevel
func (d DirtyLevel) String() string {
	switch d {
	case DirtyLow:
		return "Low"
	case DirtyNormal:
		return "Normal"
	case DirtyHigh:
		return "High"
	default:
		return "Unknown"
	}
}

// String returns the string representation of StateZone
func (z StateZone) String() string {
	switch z {
	case ZoneUI:
		return "UI"
	case ZoneData:
		return "Data"
	case ZoneBackground:
		return "Background"
	default:
		return "Unknown"
	}
}
