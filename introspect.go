package engine

// Plugins returns the metadata of every plugin currently loaded by this engine.
//
// Frontends use this to introspect plugin capabilities at startup: the list of
// supported protocols, the host capabilities each plugin requires, the
// available manifest fields, and the catalog of events a plugin can emit.
//
// Returns an empty slice if no plugins have been loaded yet (i.e. before the
// first Run or Check).
func (e *Engine) Plugins() []PluginSchema {
	if e.loader == nil {
		return nil
	}
	reg := e.loader.Current()
	if reg == nil {
		return nil
	}
	return reg.Snapshot()
}

// EventSchema looks up the schema for a single plugin event by its
// fully-qualified name ("<plugin>.<kind>").
//
// Returns nil if no loaded plugin declares an event with that name, or if no
// plugins have been loaded yet.
func (e *Engine) EventSchema(fullyQualifiedName string) *EventSpec {
	if e.loader == nil {
		return nil
	}
	reg := e.loader.Current()
	if reg == nil {
		return nil
	}
	return reg.EventSchema(fullyQualifiedName)
}
