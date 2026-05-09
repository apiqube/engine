package template

import (
	"strconv"
	"time"

	"github.com/brianvoe/gofakeit/v7"
)

// Generator produces a fake value from optional positional arguments.
type Generator func(args []string) (any, bool)

var generators = map[string]Generator{
	"name":     func([]string) (any, bool) { return gofakeit.Name(), true },
	"email":    func([]string) (any, bool) { return gofakeit.Email(), true },
	"uuid":     func([]string) (any, bool) { return gofakeit.UUID(), true },
	"phone":    func([]string) (any, bool) { return gofakeit.Phone(), true },
	"url":      func([]string) (any, bool) { return gofakeit.URL(), true },
	"company":  func([]string) (any, bool) { return gofakeit.Company(), true },
	"city":     func([]string) (any, bool) { return gofakeit.City(), true },
	"country":  func([]string) (any, bool) { return gofakeit.Country(), true },
	"color":    func([]string) (any, bool) { return gofakeit.Color(), true },
	"word":     func([]string) (any, bool) { return gofakeit.Word(), true },
	"sentence": func(a []string) (any, bool) { return gofakeit.Sentence(intArg(a, 0, 6)), true },
	"address":  func([]string) (any, bool) { return gofakeit.Address().Address, true },
	"password": func(a []string) (any, bool) {
		n := intArg(a, 0, 12)
		return gofakeit.Password(true, true, true, true, false, n), true
	},
	"bool":  func([]string) (any, bool) { return gofakeit.Bool(), true },
	"int":   genInt,
	"uint":  genUint,
	"float": genFloat,
	"date":  genDate,
}

// applyGenerator looks up a generator by name and invokes it with args.
func applyGenerator(name string, args []string) (any, bool) {
	gen, ok := generators[name]
	if !ok {
		return nil, false
	}
	return gen(args)
}

// generateRegex produces a string matching the given regular expression.
func generateRegex(pattern string) (any, bool) {
	if pattern == "" {
		return nil, false
	}
	defer func() {
		_ = recover()
	}()
	return gofakeit.Regex(pattern), true
}

func genInt(args []string) (any, bool) {
	min, max := intArg(args, 0, 0), intArg(args, 1, 1<<31-1)
	if max < min {
		min, max = max, min
	}
	return gofakeit.IntRange(min, max), true
}

func genUint(args []string) (any, bool) {
	min, max := uintArg(args, 0, 0), uintArg(args, 1, 1<<31-1)
	if max < min {
		min, max = max, min
	}
	return gofakeit.UintRange(min, max), true
}

func genFloat(args []string) (any, bool) {
	mn, mx := floatArg(args, 0, 0), floatArg(args, 1, 1)
	if mx < mn {
		mn, mx = mx, mn
	}
	return gofakeit.Float64Range(mn, mx), true
}

func genDate(args []string) (any, bool) {
	if len(args) == 0 {
		return gofakeit.Date().Format(time.RFC3339), true
	}
	return gofakeit.Date().Format(args[0]), true
}

func intArg(args []string, i, def int) int {
	if i >= len(args) {
		return def
	}
	v, err := strconv.Atoi(args[i])
	if err != nil {
		return def
	}
	return v
}

func uintArg(args []string, i int, def uint) uint {
	if i >= len(args) {
		return def
	}
	v, err := strconv.ParseUint(args[i], 10, 64)
	if err != nil {
		return def
	}
	return uint(v)
}

func floatArg(args []string, i int, def float64) float64 {
	if i >= len(args) {
		return def
	}
	v, err := strconv.ParseFloat(args[i], 64)
	if err != nil {
		return def
	}
	return v
}
