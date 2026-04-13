# ApiQube Engine

> Declarative, plugin-driven API testing engine.
> One format. Any protocol. Zero boilerplate.

[![Go](https://img.shields.io/badge/Go-1.24+-00ADD8?style=flat-square&logo=go&logoColor=white)](https://go.dev)
[![License](https://img.shields.io/badge/License-MIT-blue?style=flat-square)](LICENSE)
[![Status](https://img.shields.io/badge/Status-Active%20Development-brightgreen?style=flat-square)]()

## What is this?

`engine` is the core library behind [ApiQube](https://github.com/apiqube) — a declarative API testing platform. It handles parsing, validation, dependency analysis, execution, cross-test data flow, and reporting for any protocol through a unified WASM plugin system.

**This is a library, not a CLI tool.** For the CLI, see [`apiqube/qube`](https://github.com/apiqube/qube).

## Key Features

- **Protocol-agnostic** — HTTP, gRPC, GraphQL, WebSocket, SQL, Kafka via plugins
- **Auto-dependency graph** — detects cross-test references from templates at build time
- **Smart parallelism** — groups independent tests into waves, runs them concurrently
- **Three-level data flow** — `prev` (implicit), `save` (named), `alias` (cross-file)
- **WASM plugin system** — extend with plugins written in any language (Go, Rust, etc.)
- **Frontend-agnostic** — event-based Output interface for CLI, desktop, web, SDK

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
    results, err := eng.Run(context.Background(), "./tests/")
    if err != nil {
        panic(err)
    }
    fmt.Printf("%d passed, %d failed\n", results.Passed, results.Failed)
}
```

## Architecture

```
engine/
├── manifest/     Parsing, normalization, validation
├── graph/        Dependency analysis, topological sort, wave grouping
├── runner/       Execution engine, parallel wave runner
├── dataflow/     PassManager — prev, save, alias resolution
├── template/     Template DSL — Fake.*, chainable methods
├── assert/       Assertion engine — operators, type checks
├── plugin/       WASM plugin host (wazero), registry, contract
├── builtin/      Built-in plugins (HTTP executor)
└── config/       .qube.yaml parser
```

## Related Repositories

| Repo | Description |
|---|---|
| [`apiqube/qube`](https://github.com/apiqube/qube) | CLI tool (imports this engine) |
| [`apiqube/cli`](https://github.com/apiqube/cli) | V1 CLI (archived reference) |

## License

[MIT](LICENSE)
