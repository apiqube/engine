package engine

import "github.com/apiqube/engine/internal/wire"

// FieldSpec describes one field in a plugin contract — used both for manifest
// fields (what the plugin reads from test cases) and event fields (what the
// plugin writes into events it emits).
type FieldSpec = wire.FieldSpec

// FieldType enumerates the YAML/JSON types a plugin field may hold.
type FieldType = wire.FieldType

const (
	FieldString = wire.FieldString
	FieldNumber = wire.FieldNumber
	FieldBool   = wire.FieldBool
	FieldObject = wire.FieldObject
	FieldArray  = wire.FieldArray
	FieldMap    = wire.FieldMap
	FieldAny    = wire.FieldAny
)

// EventSpec describes a single event that a plugin can emit at runtime.
type EventSpec = wire.EventSpec

// PluginSchema is the subset of PluginInfo exposed to consumers of the engine.
// Frontends use it to introspect loaded plugins — their protocols, the host
// capabilities they require, the manifest fields they read, and the events
// they may emit.
type PluginSchema = wire.PluginSchema
