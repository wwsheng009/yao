package animation

import (
	"math"
)

// EasingFunction defines the interpolation for animations.
type EasingFunction func(float64) float64

// Easing functions for smooth animations

// Linear - constant speed
func Linear(t float64) float64 {
	return t
}

// EaseInQuad - quadratic easing in
func EaseInQuad(t float64) float64 {
	return t * t
}

// EaseOutQuad - quadratic easing out
func EaseOutQuad(t float64) float64 {
	return t * (2 - t)
}

// EaseInOutQuad - quadratic easing in and out
func EaseInOutQuad(t float64) float64 {
	if t < 0.5 {
		return 2 * t * t
	}
	return -1 + (4-2*t)*t
}

// EaseInCubic - cubic easing in
func EaseInCubic(t float64) float64 {
	return t * t * t
}

// EaseOutCubic - cubic easing out
func EaseOutCubic(t float64) float64 {
	t--
	return t*t*t + 1
}

// EaseInOutCubic - cubic easing in and out
func EaseInOutCubic(t float64) float64 {
	if t < 0.5 {
		return 4 * t * t * t
	}
	t--
	return 1 + 4*t*t*t
}

// EaseInQuart - quartic easing in
func EaseInQuart(t float64) float64 {
	return t * t * t * t
}

// EaseOutQuart - quartic easing out
func EaseOutQuart(t float64) float64 {
	t--
	return 1 - t*t*t*t
}

// EaseInOutQuart - quartic easing in and out
func EaseInOutQuart(t float64) float64 {
	if t < 0.5 {
		return 8 * t * t * t * t
	}
	t--
	return 1 - 8*t*t*t*t
}

// EaseInQuint - quintic easing in
func EaseInQuint(t float64) float64 {
	return t * t * t * t * t
}

// EaseOutQuint - quintic easing out
func EaseOutQuint(t float64) float64 {
	t--
	return 1 + t*t*t*t*t
}

// EaseInOutQuint - quintic easing in and out
func EaseInOutQuint(t float64) float64 {
	if t < 0.5 {
		return 16 * t * t * t * t * t
	}
	t--
	return 1 + 16*t*t*t*t*t
}

// EaseInSine - sinusoidal easing in
func EaseInSine(t float64) float64 {
	return 1 - math.Cos(t*math.Pi/2)
}

// EaseOutSine - sinusoidal easing out
func EaseOutSine(t float64) float64 {
	return math.Sin(t * math.Pi / 2)
}

// EaseInOutSine - sinusoidal easing in and out
func EaseInOutSine(t float64) float64 {
	return 0.5 * (1 - math.Cos(t*math.Pi))
}

// EaseInExpo - exponential easing in
func EaseInExpo(t float64) float64 {
	if t == 0 {
		return 0
	}
	return math.Pow(2, 10*(t-1))
}

// EaseOutExpo - exponential easing out
func EaseOutExpo(t float64) float64 {
	if t == 1 {
		return 1
	}
	return 1 - math.Pow(2, -10*t)
}

// EaseInOutExpo - exponential easing in and out
func EaseInOutExpo(t float64) float64 {
	if t == 0 {
		return 0
	}
	if t == 1 {
		return 1
	}
	if t < 0.5 {
		return math.Pow(2, 20*t-10) / 2
	}
	return (2 - math.Pow(2, -20*t+10)) / 2
}

// EaseInCirc - circular easing in
func EaseInCirc(t float64) float64 {
	return 1 - math.Sqrt(1-t*t)
}

// EaseOutCirc - circular easing out
func EaseOutCirc(t float64) float64 {
	t--
	return math.Sqrt(1 - t*t)
}

// EaseInOutCirc - circular easing in and out
func EaseInOutCirc(t float64) float64 {
	if t < 0.5 {
		return (1 - math.Sqrt(1-4*t*t)) / 2
	}
	t--
	return (1 + math.Sqrt(1-4*t*t)) / 2
}

// EaseInElastic - elastic easing in
func EaseInElastic(t float64) float64 {
	if t == 0 || t == 1 {
		return t
	}
	return -math.Pow(2, 10*t-10) * math.Sin((t*10-10.75)*((2*math.Pi)/3))
}

// EaseOutElastic - elastic easing out
func EaseOutElastic(t float64) float64 {
	if t == 0 || t == 1 {
		return t
	}
	return math.Pow(2, -10*t) * math.Sin((t*10-0.75)*((2*math.Pi)/3)) + 1
}

// EaseInOutElastic - elastic easing in and out
func EaseInOutElastic(t float64) float64 {
	if t == 0 || t == 1 {
		return t
	}
	if t < 0.5 {
		return -(math.Pow(2, 20*t-10) * math.Sin((20*t-11.125)*((2*math.Pi)/4.5))) / 2
	}
	return (math.Pow(2, -20*t+10)*math.Sin((20*t-11.125)*((2*math.Pi)/4.5)))/2 + 1
}

// EaseInBounce - bounce easing in
func EaseInBounce(t float64) float64 {
	return 1 - EaseOutBounce(1-t)
}

// EaseOutBounce - bounce easing out
func EaseOutBounce(t float64) float64 {
	const n1 = 7.5625
	const d1 = 2.75

	if t < 1/d1 {
		return n1 * t * t
	} else if t < 2/d1 {
		t -= 1.5 / d1
		return n1*t*t + 0.75
	} else if t < 2.5/d1 {
		t -= 2.25 / d1
		return n1*t*t + 0.9375
	}
	t -= 2.625 / d1
	return n1*t*t + 0.984375
}

// EaseInOutBounce - bounce easing in and out
func EaseInOutBounce(t float64) float64 {
	if t < 0.5 {
		return (1 - EaseOutBounce(1-2*t)) / 2
	}
	return (1 + EaseOutBounce(2*t-1)) / 2
}

// GetEasingFunction returns an easing function by name.
func GetEasingFunction(name string) EasingFunction {
	switch name {
	case "linear":
		return Linear
	case "ease-in-quad", "easeInQuad":
		return EaseInQuad
	case "ease-out-quad", "easeOutQuad":
		return EaseOutQuad
	case "ease-in-out-quad", "easeInOutQuad":
		return EaseInOutQuad
	case "ease-in-cubic", "easeInCubic":
		return EaseInCubic
	case "ease-out-cubic", "easeOutCubic":
		return EaseOutCubic
	case "ease-in-out-cubic", "easeInOutCubic":
		return EaseInOutCubic
	case "ease-in-quart", "easeInQuart":
		return EaseInQuart
	case "ease-out-quart", "easeOutQuart":
		return EaseOutQuart
	case "ease-in-out-quart", "easeInOutQuart":
		return EaseInOutQuart
	case "ease-in-quint", "easeInQuint":
		return EaseInQuint
	case "ease-out-quint", "easeOutQuint":
		return EaseOutQuint
	case "ease-in-out-quint", "easeInOutQuint":
		return EaseInOutQuint
	case "ease-in-sine", "easeInSine":
		return EaseInSine
	case "ease-out-sine", "easeOutSine":
		return EaseOutSine
	case "ease-in-out-sine", "easeInOutSine":
		return EaseInOutSine
	case "ease-in-expo", "easeInExpo":
		return EaseInExpo
	case "ease-out-expo", "easeOutExpo":
		return EaseOutExpo
	case "ease-in-out-expo", "easeInOutExpo":
		return EaseInOutExpo
	case "ease-in-circ", "easeInCirc":
		return EaseInCirc
	case "ease-out-circ", "easeOutCirc":
		return EaseOutCirc
	case "ease-in-out-circ", "easeInOutCirc":
		return EaseInOutCirc
	case "ease-in-elastic", "easeInElastic":
		return EaseInElastic
	case "ease-out-elastic", "easeOutElastic":
		return EaseOutElastic
	case "ease-in-out-elastic", "easeInOutElastic":
		return EaseInOutElastic
	case "ease-in-bounce", "easeInBounce":
		return EaseInBounce
	case "ease-out-bounce", "easeOutBounce":
		return EaseOutBounce
	case "ease-in-out-bounce", "easeInOutBounce":
		return EaseInOutBounce
	default:
		return Linear
	}
}
