// Package template implements the ApiQube template DSL.
//
// Templates are enclosed in {{ ... }} and can be:
//
//   - Fake data generators: {{ fake.name }}, {{ fake.email }}
//   - Chainable methods:    {{ fake.email.ToLower() }}
//   - Regex patterns:       {{ regex('[A-Z]{3}') }}
//   - Environment vars:     {{ env.API_HOST }}
//   - Data references:      {{ prev.body.id }}, {{ token }}, {{ auth.response.body.user }}
//
// The template engine is lenient by design — if a method cannot be applied,
// the original value is returned instead of panicking. Tests should fail on
// assertions, not on template typos.
package template
