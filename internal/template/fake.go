package template

// Generator is a function that produces fake data given optional arguments.
// Used for {{ fake.name }}, {{ fake.int.1.100 }}, etc.
type Generator func(args ...string) (any, error)

// generators is the registry of all fake.* generators.
var generators = map[string]Generator{
	// TODO: implement using brianvoe/gofakeit/v7
	// "name":     generateName,
	// "email":    generateEmail,
	// "uuid":     generateUUID,
	// "int":      generateInt,
	// "uint":     generateUint,
	// "float":    generateFloat,
	// "bool":     generateBool,
	// "phone":    generatePhone,
	// "date":     generateDate,
	// "sentence": generateSentence,
	// "word":     generateWord,
	// "url":      generateURL,
	// "company":  generateCompany,
	// "address":  generateAddress,
	// "country":  generateCountry,
	// "city":     generateCity,
	// "color":    generateColor,
	// "password": generatePassword,
}

// applyGenerator looks up a generator by name and invokes it with args.
// Returns (value, true) if generator exists, (nil, false) otherwise.
func applyGenerator(name string, args []string) (any, bool) {
	gen, ok := generators[name]
	if !ok {
		return nil, false
	}
	val, err := gen(args...)
	if err != nil {
		return nil, false
	}
	return val, true
}
