package dataflow

// Extract pulls a value at the given dot-separated path from a nested value.
// Supports gjson-style paths over any map[string]any / []any / primitive tree.
//
// Example paths:
//   "body.users.0.name"     → response.body["users"][0]["name"]
//   "body.items.length"     → len(response.body["items"])
//   "headers.Content-Type"  → response.headers["Content-Type"]
func Extract(source any, path string) (any, bool) {
	// TODO: implementation
	// 1. Split path by "."
	// 2. Walk segments:
	//    - if current is map: lookup key
	//    - if current is slice: parse segment as int index
	//    - special: "length" returns len() of slice/string/map
	// 3. Return final value or false if not found
	return nil, false
}
