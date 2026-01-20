package tui


// applyStateToProps processes component props and replaces {{}} expressions
// with actual values from the State.
func (m *Model) applyStateToProps(comp *Component) map[string]interface{} {
	if comp.Props == nil {
		return make(map[string]interface{})
	}

	result := make(map[string]interface{})

	// Process each prop
	for key, value := range comp.Props {
		// Check if value is a string with {{}} expression
		if str, ok := value.(string); ok {
			result[key] = m.applyState(str)
		} else {
			// For non-string values, try to evaluate them as expressions if they contain {{}}
			result[key] = m.evaluateValue(value)
		}
	}

	// Handle bind attribute - bind entire state object
	if comp.Bind != "" {
		m.StateMu.RLock()
		if bindValue, ok := m.State[comp.Bind]; ok {
			result["__bind_data"] = bindValue
		}
		m.StateMu.RUnlock()
	}

	return result
}

// resolveProps processes component props and resolves {{}} expressions to their actual types,
// without converting complex data types to strings.
func (m *Model) resolveProps(comp *Component) map[string]interface{} {
	if comp.Props == nil {
		return make(map[string]interface{})
	}

	// Get current state snapshot for cache comparison
	m.StateMu.RLock()
	currentState := make(map[string]interface{}, len(m.State))
	for k, v := range m.State {
		currentState[k] = v
	}
	m.StateMu.RUnlock()

	// Use props cache if available
	if m.propsCache != nil && comp.ID != "" {
		resolvedProps, err := m.propsCache.GetOrResolve(
			comp.ID,
			comp.Props,
			currentState,
			func() (map[string]interface{}, error) {
				result := make(map[string]interface{})

				// Process each prop
				for key, value := range comp.Props {
					// Use evaluateValue for consistent processing of all value types
					result[key] = m.evaluateValue(value)
				}

				// Handle bind attribute - bind entire state object
				if comp.Bind != "" {
					m.StateMu.RLock()
					if bindValue, ok := m.State[comp.Bind]; ok {
						result["__bind_data"] = bindValue
					}
					m.StateMu.RUnlock()
				}

				return result, nil
			},
		)
		if err == nil {
			return resolvedProps
		}
		// If caching fails, fall back to normal processing
	}

	// Normal processing without cache
	result := make(map[string]interface{})

	// Process each prop
	for key, value := range comp.Props {
		// Use evaluateValue for consistent processing of all value types
		result[key] = m.evaluateValue(value)
	}

	// Handle bind attribute - bind entire state object
	if comp.Bind != "" {
		m.StateMu.RLock()
		if bindValue, ok := m.State[comp.Bind]; ok {
			result["__bind_data"] = bindValue
		}
		m.StateMu.RUnlock()
	}

	return result
}

// getStringProp safely retrieves a string property with a default value.
func getStringProp(props map[string]interface{}, key, defaultValue string) string {
	if value, ok := props[key]; ok {
		if str, ok := value.(string); ok {
			return str
		}
	}
	return defaultValue
}

// getIntProp safely retrieves an int property with a default value.
func getIntProp(props map[string]interface{}, key string, defaultValue int) int {
	if value, ok := props[key]; ok {
		switch v := value.(type) {
		case int:
			return v
		case float64:
			return int(v)
		}
	}
	return defaultValue
}

// getBoolProp safely retrieves a bool property with a default value.
func getBoolProp(props map[string]interface{}, key string, defaultValue bool) bool {
	if value, ok := props[key]; ok {
		if b, ok := value.(bool); ok {
			return b
		}
	}
	return defaultValue
}
