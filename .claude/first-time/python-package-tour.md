## NautilusTrader Subpackage Overview for First-Time Contributors

Here's a guide to the `nautilus_trader/` Python package structure:

## Quick Mental Model

Here's a **Quick Mental Model** for each package in `nautilus_trader/`:

| Package | Question |
|---------|----------|
| **accounting** | "What happened in THIS account?" (balances, margins, P&L) |
| **portfolio** | "What's my overall position across EVERYTHING?" |
| **adapters** | "How do I talk to THIS specific exchange/broker?" |
| **analysis** | "How well did my strategy perform?" (Sharpe, drawdown, tearsheets) |
| **backtest** | "What would have happened if I traded this historically?" |
| **cache** | "What's the current state of everything RIGHT NOW?" |
| **common** | "What time is it? Who should receive this message?" (clocks, message bus, logging) |
| **config** | "How should this component be configured?" |
| **core** | "What's the base type for this thing?" (events, commands, messages) |
| **data** | "What's happening in the market?" (quotes, trades, bars) |
| **execution** | "Where is my order and what's its status?" |
| **indicators** | "What are the technicals telling me?" (RSI, MACD, Bollinger) |
| **live** | "How do I run this in production with real money?" |
| **model** | "What IS this thing?" (Price, Order, Position, Instrument) |
| **persistence** | "Where do I store/load historical data?" (Parquet catalog) |
| **risk** | "Should I be allowed to place this order?" |
| **serialization** | "How do I convert this object to/from bytes?" |
| **system** | "Who wires everything together and runs the show?" (NautilusKernel) |
| **test_kit** | "How do I test my strategy?" (mocks, fixtures, helpers) |
| **trading** | "What should I buy/sell and when?" (Strategy, Trader) |

---

## Grouped by Concern

**Data Flow:** "What's the market doing?"
- `data` → "What prices/trades are coming in?"
- `indicators` → "What patterns do I see?"
- `cache` → "What's the latest snapshot?"

**Order Flow:** "What about my orders?"
- `execution` → "Did my order fill?"
- `risk` → "Am I allowed to trade this?"
- `accounting` → "What's my account balance?"
- `portfolio` → "What's my net exposure?"

**Infrastructure:** "How does the system work?"
- `core` → "What are the building blocks?"
- `common` → "What utilities do I need?"
- `config` → "How is it configured?"
- `system` → "Who orchestrates everything?"

**Environment:** "Where am I running?"
- `backtest` → "Simulating the past"
- `live` → "Trading for real"
- `adapters` → "Connected to which venue?"

### Core Infrastructure

| Package | Description |
|---------|-------------|
| **core** | Foundation layer with constants, message types (Command, Event, Request, Response), UUID generation, and a generic FiniteStateMachine. Start here to understand the building blocks. |
| **common** | Shared infrastructure: clocks (Test/Live), timers, ID generators, logging, and high-performance queues. Used across backtest and live environments. |
| **config** | Configuration classes inheriting from `NautilusConfig`. Uses `msgspec` for serialization. Every component has a corresponding config class here. |
| **system** | Contains `NautilusKernel` - the central orchestrator that initializes all components and connects them via the message bus. |

### Trading Domain Model

| Package | Description |
|---------|-------------|
| **model** | Rich trading domain model: orders, positions, instruments, bars, ticks, order books. Domain-agnostic design that any trading system could build upon. |
| **trading** | User-facing API where you inherit from `Strategy` to implement custom trading strategies. **Most users start here.** |

### Data & Execution Stacks

| Package | Description |
|---------|-------------|
| **data** | Data engine, clients, and handlers for market data (ticks, bars, order books). Layered architecture reused between backtest and live. |
| **execution** | Execution engine for order management. Mirrors the data stack design for consistent logic across environments. |
| **cache** | High-performance in-memory storage for instruments, orders, positions, and account state. Central access point for all components. |

### Environment Contexts

| Package | Description |
|---------|-------------|
| **backtest** | Backtesting framework with simulated venues at nanosecond resolution. Uses historical data to test strategies. |
| **live** | Live trading implementation using `uvloop`. Common event loop across all live components for performance. |
| **adapters** | Exchange/broker integrations (Binance, Kraken, Interactive Brokers, etc.). Implements HTTP REST and WebSocket clients. |

### Portfolio & Risk

| Package | Description |
|---------|-------------|
| **accounting** | Account management, P&L calculations, and FX/crypto exchange rate calculations for multi-currency portfolios. |
| **portfolio** | Position tracking, portfolio-level P&L, and multi-asset management via the `Portfolio` class. |
| **risk** | Risk management components including `PositionSizer` for strategies to use in sizing positions. |

### Analysis & Persistence

| Package | Description |
|---------|-------------|
| **analysis** | Performance statistics, visualization (tearsheets, charts, heatmaps) for evaluating backtest and live results. |
| **indicators** | Technical indicators (moving averages, momentum, volatility, etc.). Use as-is or as templates for custom indicators. |
| **persistence** | Data storage/retrieval for backtesting. Handles persisting and loading market data. |
| **serialization** | Base classes for serialization (msgspec-based). Enables saving/loading trading objects. |

### Development Support

| Package | Description |
|---------|-------------|
| **test_kit** | Test doubles, mocks, and fixtures for the test suite. Also useful for downstream projects testing their strategies. |
| **examples** | Example strategies and configurations to learn from. |

---

### Architecture Quick Reference

```
┌─────────────────────────────────────────────────────────┐
│                    NautilusKernel                       │
│                      (system)                           │
├─────────────────────────────────────────────────────────┤
│                     MessageBus                          │
│              (pub/sub communication)                    │
├──────────────┬──────────────┬──────────────┬────────────┤
│  DataEngine  │ ExecEngine   │  Portfolio   │   Risk     │
│    (data)    │ (execution)  │ (portfolio)  │  (risk)    │
├──────────────┴──────────────┴──────────────┴────────────┤
│                       Cache                             │
│              (in-memory state store)                    │
├─────────────────────────────────────────────────────────┤
│     Adapters (adapters) ←→ External Venues/Data         │
└─────────────────────────────────────────────────────────┘
```

### Key Concepts for Contributors

1. **Parity Problem Solved**: Same strategy code works identically in backtest and live - this is the core design principle
2. **Hybrid Rust/Python**: Performance-critical code lives in `crates/`, Python provides the user API
3. **Event-Driven**: Components communicate via the MessageBus using events
4. **Ports and Adapters**: Core logic is isolated; adapters are pluggable