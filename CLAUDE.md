# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

NautilusTrader is a high-performance algorithmic trading platform providing backtesting and live trading capabilities with identical strategy code. It's a hybrid Rust/Python system where Rust handles performance-critical components and Python provides the user-facing API.

**Key principle**: "Parity problem" solved - same strategy code works in backtest and live environments without modification.

## Build Commands

```bash
# Installation (uses uv package manager)
make install              # Release mode with all dependencies
make install-debug        # Debug mode for development (recommended)

# Building
make build                # Release mode
make build-debug          # Debug mode (faster iteration)

# Cleaning
make clean                # Remove all build artifacts
```

## Testing

```bash
# Python tests
make pytest               # Run all Python tests in parallel
pytest tests/unit_tests/path/to/test.py::test_function  # Single test
pytest -k "pattern"       # Tests matching pattern

# Rust tests
make cargo-test           # All Rust tests with nextest
make cargo-test FAIL_FAST=true  # Stop on first failure
make cargo-test-crate-nautilus-model  # Single crate
cargo nextest run -p nautilus-model test_name  # Single Rust test

# With optional features
make cargo-test EXTRA_FEATURES="capnp,hypersync"

# Performance tests
make test-performance

# Integration test services (Docker: Redis, Postgres)
make init-services        # Start containers and initialize database
make stop-services        # Stop containers (preserves data)
```

## Code Quality

```bash
make format               # Format Rust (nightly) and Python code
make pre-commit           # Run all pre-commit hooks
make check-code           # Run clippy and ruff --fix
make pre-flight           # Comprehensive checks before PR
```

## Architecture

### Hybrid Rust/Python Structure

- **`crates/`** - Rust crates (40+): core types, execution, networking, adapters
- **`nautilus_trader/`** - Python/Cython modules: user API, strategies, configuration
- **`tests/`** - Test suites: unit, integration, acceptance, performance, memory leak

### Core Components

- **NautilusKernel** - Central orchestration, initializes all system components
- **MessageBus** - Pub/Sub, Request/Response communication backbone
- **Cache** - High-performance in-memory storage for instruments, orders, positions
- **Adapters** - Modular venue/data provider integrations (Binance, Kraken, IB, etc.)

### Adapter Architecture

Adapters follow a layered pattern: **Rust core** (`crates/adapters/<name>/`) handles HTTP clients, WebSocket streaming, message parsing, and PyO3 bindings. **Python layer** (`nautilus_trader/adapters/<name>/`) integrates Rust clients with the platform's data and execution engines. Cython is being phased out; new code should use Rust.

### Environment Contexts

- **Backtest** - Historical data, simulated venues, nanosecond resolution
- **Sandbox** - Real-time data, simulated venues (paper trading)
- **Live** - Real-time data, live venues

### Design Patterns

- **Domain-Driven Design** - Rich trading domain model
- **Event-Driven** - Message-based inter-component communication
- **Ports and Adapters** - Pluggable venue adapters, core logic isolated
- **Crash-Only** - Unified recovery path, externalized state via Redis

## Coding Conventions

### Universal

- Spaces only, no tabs; lines under 100 characters
- American English spelling (`color`, `serialize`, `behavior`)
- Error variable naming: use `e`, not `err` or `error`
- Error messages: avoid ", got" - use ", was", ", received", or ", found"
- Single-line comments must not end with a period; multi-line comments end with a period
- All source files require standardized copyright headers (enforced by pre-commit)

### Commit Messages

- Limit subject to 60 characters, capitalize, no trailing period
- Use imperative voice ("Add feature" not "Added feature")

### Python

- Type hints required; use `msgspec` for serialization; use PEP 604 unions (`X | None` not `Optional[X]`)
- Single imports per line; trailing commas on multi-line arguments
- Prefer pytest-style functions over test classes
- NumPy docstring format; imperative mood ("Return a cached client")
- No docstrings on private methods (prefixed with `_`) unless non-trivially complex

### Rust

- Fully qualify `anyhow` macros (`anyhow::bail!`, `anyhow::Result<T>`)
- Fully qualify `tokio` types (`tokio::spawn`, `tokio::time::timeout`)
- Fully qualify logging macros (`log::info!`, `log::warn!`)
- Use inline format strings: `anyhow::bail!("Failed: {value}")` not positional args
- Use `#[rstest]` for all tests, not `#[test]`
- In adapters: use `get_runtime().spawn()` not `tokio::spawn()` (Python FFI compatibility)
- Doc comments use indicative mood ("Returns a cached client")
- Use `AHashMap`/`AHashSet` for hot paths; standard `HashMap`/`HashSet` otherwise
- Use `nautilus_core::correctness::FAILED` constant for `.expect()` messages in constructors
- No box-style banner/separator comments

### PyO3 Conventions

- Rust functions exposed to Python must use `py_` prefix: `pub fn py_do_something()`
- Use `#[pyo3(name = "do_something")]` to remove prefix in Python API
- For `PyObject` cloning, use `clone_py_object()` from `nautilus_core::python` (avoids `Arc<PyObject>` reference cycles)

### Constructor Pattern

```rust
// new_checked() returns Result for error handling
pub fn new_checked<T: AsRef<str>>(value: T) -> anyhow::Result<Self>

// new() panics on invalid input (uses new_checked internally)
pub fn new<T: AsRef<str>>(value: T) -> Self {
    Self::new_checked(value).expect(FAILED)
}
```

## Build Cache Alignment

Cargo rebuilds when features/profiles mismatch. Testing (`cargo-test`) and linting (`cargo-clippy`) both use features `ffi,python,high-precision,defi` with `nextest` profile to share build artifacts. Python extension builds (`build`/`build-debug`) use different features (`extension-module`) and will trigger rebuilds — this is expected.

## Key Files

- `Makefile` - Build automation (30+ targets, run `make help`)
- `pyproject.toml` - Python configuration (uv/poetry)
- `Cargo.toml` - Rust workspace manifest
- `.pre-commit-config.yaml` - Code quality gates (30+ hooks including custom convention checks)
- `docs/developer_guide/` - Comprehensive developer documentation

## Precision Modes

- **High-precision (default)**: 128-bit integers, up to 16 decimal places (Linux/macOS)
- **Standard-precision**: 64-bit integers, up to 9 decimal places (all platforms, Windows only)

Set via `HIGH_PRECISION=false make build` or Rust feature flag.

## Async Patterns in Adapters

```rust
use nautilus_common::live::get_runtime;

// Correct - works from Python threads
get_runtime().spawn(async move { /* ... */ });

// Incorrect - panics from Python threads (no Tokio context)
tokio::spawn(async move { /* ... */ });
```

## Testing Async Code

Use polling helpers instead of arbitrary sleeps:
- Python: `await eventually(...)` from `nautilus_trader.test_kit.functions`
- Rust: `wait_until_async(...)` from `nautilus_common::testing`
