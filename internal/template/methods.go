package template

// Method transforms a value via a chainable operation on a template expression.
// Example: {{ fake.name.ToUpper() }} applies the ToUpper method to fake.name.
type Method func(input any, args []string) (any, error)

// methods is the registry of all chainable methods on template expressions.
var methods = map[string]Method{
	// TODO: implementation
	// "ToUpper":    methodToUpper,
	// "ToLower":    methodToLower,
	// "TrimSpace":  methodTrimSpace,
	// "Replace":    methodReplace,
	// "PadLeft":    methodPadLeft,
	// "PadRight":   methodPadRight,
	// "Substring":  methodSubstring,
	// "Capitalize": methodCapitalize,
	// "Reverse":    methodReverse,
	// "RandomCase": methodRandomCase,
	// "SnakeCase":  methodSnakeCase,
	// "CamelCase":  methodCamelCase,
	// "Split":      methodSplit,
	// "Join":       methodJoin,
	// "Index":      methodIndex,
	// "Cut":        methodCut,
	// "ToInt":      methodToInt,
	// "ToUint":     methodToUint,
	// "ToFloat":    methodToFloat,
	// "ToBool":     methodToBool,
	// "ToString":   methodToString,
	// "ToArray":    methodToArray,
}

// applyMethod invokes a named method on the current value.
// Lenient: if the method fails or doesn't exist, returns input unchanged.
func applyMethod(name string, input any, args []string) any {
	m, ok := methods[name]
	if !ok {
		return input
	}
	result, err := m(input, args)
	if err != nil {
		return input
	}
	return result
}
