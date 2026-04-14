package engine

// Plugins returns the metadata of every plugin currently loaded by this engine.
//
// Frontends use this to introspect plugin capabilities at startup: list of
// supported protocols, available manifest fields, and the full catalog of
// events a plugin can emit. A CLI might print this as a table, a web UI
// might render it as a browsable panel, and an SDK might use it to
// auto-generate typed helper methods.
//
// Returns an empty slice if no plugins have been loaded yet.
func (e *Engine) Plugins() []PluginSchema {
	// TODO: implementation — when plugin loading is wired up,
	// return snapshots from the plugin registry.
	return nil
}

// EventSchema looks up the schema for a single plugin event by its
// fully-qualified name ("<plugin>.<kind>").
//
// Returns nil if no loaded plugin declares an event with that name.
func (e *Engine) EventSchema(fullyQualifiedName string) *EventSpec {
	// TODO: implementation
	_ = fullyQualifiedName
	return nil
}

// EventSchemas returns the full catalog of plugin events across every loaded
// plugin, keyed by fully-qualified name ("<plugin>.<kind>").
//
// Useful for documentation generation, UI event pickers, and schema validation
// at subscribe time.
func (e *Engine) EventSchemas() map[string]EventSpec {
	// TODO: implementation
	return nil
}
