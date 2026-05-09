# ApiQube Engine

> Declarative, plugin-driven API testing engine.
> One format. Any protocol. Zero boilerplate.

[![Go](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![CI](https://github.com/apiqube/engine/actions/workflows/ci.yml/badge.svg)](https://github.com/apiqube/engine/actions/workflows/ci.yml)
[![License](https://img.shields.io/badge/License-MIT-blue?style=flat-square)](LICENSE)
[![Status](https://img.shields.io/badge/Status-Active%20Development-brightgreen?style=flat-square)]()

## What is this?

`engine` is the core library behind [ApiQube](https://github.com/apiqube) — a declarative API testing platform. It parses test manifests, builds a dependency graph from template references, executes tests through WASM protocol plugins, and emits typed events for any frontend to consume.

**This is a library, not a CLI tool.** For the CLI, see [`apiqube/qube`](https://github.com/apiqube/qube).

## Key features

- **Protocol-agnostic** — HTTP, gRPC, GraphQL, WebSocket and more via WASM plugins.
- **Auto-dependency graph** — detects cross-test references in templates at build time.
- **Smart parallelism** — groups independent tests into waves, runs them concurrently.
- **Three-level data flow** — `prev` (implicit), `save` (named), `alias` (cross-file), plus streaming events.
- **Frontend-agnostic** — typed event interface for CLI, desktop, web, SDK.
- **Capability-based plugin host** — WASM plugins declare host capabilities they need; engine grants them per policy.

## Usage

```go
package main

import (
    "context"
    "fmt"

    "github.com/apiqube/engine"
)

func main() {
    eng := engine.New()
    results, err := eng.Run(context.Background(), engine.FromPaths("./tests/"))
    if err != nil {
        panic(err)
    }
    fmt.Printf("%d passed, %d failed\n", results.Passed, results.Failed)
}
```

## Architecture

```
engine/
├── engine.go          Engine constructor, Run(), Check()
├── input.go           Sealed Input source — FromPaths, FromBytes, FromReader
├── options.go         Engine-level functional options
├── runopts.go         Per-Run options (handler, signals, env, plugins)
├── checkopts.go       Per-Check options
├── events.go          Sealed Event hierarchy + EventHandler interface
├── dispatcher.go      Typed and plugin-event subscription routing
├── results.go         Results, TestResult, WaveResult, ValidationError
├── data.go            RequestData, ResponseData, AssertionResult
├── protocol.go        Open Protocol type
├── status.go          TestStatus, Signal
├── schema.go          PluginSchema, FieldSpec, EventSpec for introspection
├── introspect.go      Engine.Plugins(), EventSchema()
└── internal/
    ├── manifest/      TestFile / TestCase / Expect / RetryConfig / LoadConfig
    ├── config/        .qube.yaml schema and loader
    ├── parser/        YAML loader, three-syntax normalizer
    ├── graph/         Dependency graph builder, topological sort, save requirements
    ├── dataflow/      Runtime store, prev snapshot, channel-based wait, gjson extract
    ├── template/      Template DSL — fake.*, regex(), method chains
    ├── assertion/     Operator engine and type-coercion rules
    ├── plugin/        WASM host adapter (wazero), registry, capability layer
    │   └── capabilities/   log, time, events, http
    └── runner/        Wave execution, per-test pipeline
```

## Related repositories

| Repo | Description |
|---|---|
| [`apiqube/qube`](https://github.com/apiqube/qube) | CLI tool (imports this engine) |
| [`apiqube/plugin-http`](https://github.com/apiqube/plugin-http) | First-party HTTP plugin |
| [`apiqube/cli`](https://github.com/apiqube/cli) | V1 CLI (archived reference) |

## License

[MIT](LICENSE)
