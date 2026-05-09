package plugin

import (
	"context"
	"errors"
	"fmt"

	"github.com/tetratelabs/wazero/api"
)

// allocateExportName is the WASM export plugins must provide so the host can
// place bytes into the plugin's linear memory. See the plugin contract docs.
const allocateExportName = "allocate"

// errNoAllocate indicates the loaded WASM module did not export `allocate`.
var errNoAllocate = errors.New("plugin: WASM module missing 'allocate' export")

// readPacked reads bytes from the module's memory at the location encoded in
// the packed (ptr<<32)|len uint64.
func readPacked(mem api.Memory, packed uint64) ([]byte, error) {
	ptr := uint32(packed >> 32)
	length := uint32(packed)
	if length == 0 {
		return nil, nil
	}
	data, ok := mem.Read(ptr, length)
	if !ok {
		return nil, fmt.Errorf("read %d bytes at 0x%x failed", length, ptr)
	}
	out := make([]byte, len(data))
	copy(out, data)
	return out, nil
}

// writePacked allocates `len(data)` bytes inside the plugin (via its allocate
// export), writes data there, and returns the packed (ptr<<32)|len pointer.
func writePacked(ctx context.Context, mod api.Module, data []byte) (uint64, error) {
	if len(data) == 0 {
		return 0, nil
	}
	alloc := mod.ExportedFunction(allocateExportName)
	if alloc == nil {
		return 0, errNoAllocate
	}
	res, err := alloc.Call(ctx, uint64(len(data)))
	if err != nil {
		return 0, fmt.Errorf("allocate(%d): %w", len(data), err)
	}
	if len(res) == 0 {
		return 0, errors.New("allocate returned no value")
	}
	packed := res[0]
	ptr := uint32(packed >> 32)
	length := uint32(packed)
	if length != uint32(len(data)) {
		return 0, fmt.Errorf("allocate gave %d bytes, wanted %d", length, len(data))
	}
	if !mod.Memory().Write(ptr, data) {
		return 0, fmt.Errorf("write %d bytes at 0x%x failed", length, ptr)
	}
	return packed, nil
}
