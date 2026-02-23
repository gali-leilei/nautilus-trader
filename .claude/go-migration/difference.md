# Comparison: nautilus_trader (Python) vs nautilus_go (Go)

## Context

Comparing the package structure and functionality between the Python reference implementation (`nautilus_trader/`) and the Go port (`nautilus_go/`) to identify what's been implemented, what's shared, and what's missing in Go.

---

## 1. What is the SAME between them?

Both implementations share these core architectural concepts:

| Area | Python (`nautilus_trader/`) | Go (`nautilus_go/pkg/`) | Notes |
|------|---------------------------|------------------------|-------|
| **Domain Model** | `model/` | `model/` | Both have: Bar/BarType/BarSpecification, Price/Quantity (fixed-point), Money/Currency, InstrumentId/Symbol/Venue, OrderSide, TimeInForce, enums |
| **Strategy Framework** | `trading/strategy.pyx` | `trading/strategy.go` | Both have: Strategy interface with OnStart/OnBar/OnStop lifecycle, StrategyContext as service provider |
| **Message Bus** | `common/messages.pyx` | `common/messagebus.go` | Both have: topic-based pub/sub for bar events |
| **Clock** | `common/` (TestClock/LiveClock) | `common/clock.go` | Both have: TestClock with manual time control for backtesting |
| **Cache** | `cache/cache.pyx` | `cache/cache.go` | Both have: in-memory storage for instruments and bars |
| **Portfolio** | `portfolio/portfolio.pyx` | `portfolio/portfolio.go` | Both have: position tracking per instrument, netting OMS, long/short/flat states |
| **Indicators** | `indicators/` | `indicators/` | Both have: Indicator interface, EMA implementation |
| **Data Engine** | `data/engine.pyx` | `data/engine.go` | Both have: bar validation, caching, publishing to message bus |
| **Backtest Engine** | `backtest/engine.pyx` | `backtest/engine.go` | Both have: main orchestration loop, venue management, strategy lifecycle |
| **Simulated Exchange** | `backtest/` | `backtest/exchange.go` | Both have: simulated venue that processes bars and fills market orders |
| **Order Factory** | `trading/` | `trading/order_factory.go` | Both have: auto-incrementing order ID generation |
| **Fixed-Point Arithmetic** | `model/objects.pyx` | `model/price.go` | Both use int64-based fixed-point for financial precision (Go: 9 digits, Python: up to 16 digits) |

**Shared design patterns:**
- Event-driven architecture with message bus
- Strategy context as single dependency injection point
- Lifecycle hooks: OnStart -> OnBar (repeated) -> OnStop
- Netting portfolio for position management
- Indicator auto-registration and bar feeding
- Simulated exchange for backtesting

---

## 2. What is DIFFERENT between them?

| Aspect | Python | Go |
|--------|--------|-----|
| **Precision** | 128-bit integers, up to 16 decimal places (high-precision mode) | 64-bit integers, 9 decimal places max |
| **Instrument types** | 17+ types (FuturesContract, CurrencyPair, Equity, CryptoPerpetual, Option, CFD, Betting, etc.) | 1 type (FuturesContract only) |
| **Order types** | 12 types (Market, Limit, Stop, StopLimit, TrailingStop, LimitIfTouched, MarketIfTouched, MarketToLimit, etc.) | 1 type (MarketOrder only) |
| **Order fill simulation** | Sophisticated models: slippage, partial fills, fee models, latency models, matching core | Simple: fills all pending orders at bar close price |
| **Message bus** | Full pub/sub + request/response, supports all message types (Data, Commands, Events) | Bar-only pub/sub |
| **Data types** | Bar, QuoteTick, TradeTick, OrderBookDelta, OrderBookDepth10, FundingRate, MarkPrice, etc. | Bar only |
| **Clock** | TestClock + LiveClock with timers and alerts | TestClock only (no timers/alerts) |
| **Actor model** | Full Actor base class with component lifecycle, heartbeats, state management | No actor model; strategies are direct interface implementations |
| **Cython/Rust core** | Performance-critical paths in Cython (.pyx) and Rust (PyO3) | Pure Go |
| **Config system** | Rich msgspec-based config classes with validation, serialization, defaults | Simple Go structs (EngineConfig, VenueConfig) |
| **Logging** | Structured logging with Rust backend | Go stdlib slog |
| **Enums** | Comprehensive (50+ enum types across the codebase) | 15 enum types covering basics |

---

## 3. What is MISSING in the Go version?

### Critical / Core Functionality

| Missing Package/Feature | Python Source | Importance | Description |
|--------------------------|--------------|------------|-------------|
| **Execution Engine** | `execution/engine.pyx` | **Critical** | Central execution management - order routing, state tracking, event generation. Go's exchange does basic fills but lacks the full execution pipeline |
| **Risk Engine** | `risk/engine.pyx` | **High** | Pre-trade risk checks, position limits, order rate throttling, max order size |
| **Accounting** | `accounting/` | **High** | Account management, balance tracking, margin calculation, exchange rate conversion. Go has no account/balance tracking at all |
| **Order Events & State Machine** | `model/events/order/` | **High** | Full order lifecycle events (OrderInitialized, OrderSubmitted, OrderAccepted, OrderFilled, OrderCanceled, OrderExpired, OrderRejected, etc.) with FSM transitions |
| **Position Events** | `model/events/position/` | **High** | PositionOpened, PositionChanged, PositionClosed events |
| **Account Events** | `model/events/account/` | **Medium** | AccountState events |

### Order & Execution Features

| Missing Feature | Python Source | Description |
|-----------------|--------------|-------------|
| **Limit Orders** | `model/orders/limit.pyx` | Price-limited orders |
| **Stop Orders** | `model/orders/stop_market.pyx`, `stop_limit.pyx` | Stop-loss and stop-limit orders |
| **Trailing Stops** | `model/orders/trailing_stop_*.pyx`, `execution/trailing.pyx` | Trailing stop logic |
| **Order Lists** | `model/orders/list.pyx` | Bracket orders (entry + SL + TP) |
| **Execution Algorithms** | `execution/algorithm.pyx` | TWAP, VWAP, iceberg execution |
| **Order Matching Core** | `execution/matching_core.pyx` | Realistic order book matching |
| **Order Emulator** | `execution/emulator.pyx` | Emulate complex order types from simple ones |
| **Fill Models** | `backtest/models/` | Slippage, partial fills, fee/commission models, latency simulation |

### Data & Market Data

| Missing Feature | Python Source | Description |
|-----------------|--------------|-------------|
| **Quote Ticks** | `model/data.pyx` | Bid/ask tick data |
| **Trade Ticks** | `model/data.pyx` | Trade-by-trade data |
| **Order Book** | `model/book.pyx` | L2/L3 order book, deltas, depth snapshots |
| **Bar Aggregation** | `data/aggregation.pyx` | Build bars from ticks (time, tick, volume, value, etc.) |
| **Multiple Data Types** | `data/engine.pyx` | Subscribe to and process ticks, order book updates, custom data |
| **Data Client** | `data/client.pyx` | Abstract data client for different providers |

### Instrument Types

| Missing Instrument | Python Source | Description |
|--------------------|--------------|-------------|
| **CurrencyPair** | `model/instruments/currency_pair.pyx` | Forex pairs |
| **Equity** | `model/instruments/equity.pyx` | Stocks |
| **CryptoPerpetual** | `model/instruments/crypto_perpetual.pyx` | Perpetual swaps |
| **OptionContract** | `model/instruments/option_contract.pyx` | Options |
| **CryptoFuture** | `model/instruments/crypto_future.pyx` | Crypto futures |
| **BettingInstrument** | `model/instruments/betting.pyx` | Betting markets |
| **Synthetic** | `model/instruments/synthetic.pyx` | Synthetic instruments |
| + 10 more types | | |

### Indicators

| Missing Indicator | Python Source | Description |
|-------------------|--------------|-------------|
| **SMA** | `indicators/averages.pyx` | Simple Moving Average |
| **DEMA, HMA, WMA, VIMA** | `indicators/averages.pyx` | Various moving averages (7 total, only EMA exists) |
| **RSI** | `indicators/momentum.pyx` | Relative Strength Index |
| **MACD** | `indicators/trend.pyx` | Moving Average Convergence Divergence |
| **ATR** | `indicators/volatility.pyx` | Average True Range |
| **Bollinger Bands** | `indicators/volatility.pyx` | Volatility bands |
| **Stochastics, CCI, etc.** | `indicators/momentum.pyx` | Momentum indicators |
| + 20 more indicators | | |

### Infrastructure & System

| Missing Feature | Python Source | Description |
|-----------------|--------------|-------------|
| **NautilusKernel** | `system/kernel.py` | Central orchestration / dependency injection |
| **Actor Model** | `common/actor.pyx` | Base class for all components with lifecycle, state machine, message handling |
| **Component Model** | `common/component.pyx` | Component lifecycle (PRE_INITIALIZED -> INITIALIZED -> RUNNING -> STOPPED -> etc.) |
| **Finite State Machine** | `core/fsm.pyx` | Generic FSM used by orders, components |
| **ID Generators** | `common/generators.pyx` | UUID and client order ID generators |
| **Timers & Alerts** | via Clock | Schedule callbacks at future times, set alerts |

### Live Trading (entire subsystem absent)

| Missing Feature | Python Source | Description |
|-----------------|--------------|-------------|
| **Live Node** | `live/node.py` | Live trading entry point |
| **Live Data Engine** | `live/data_engine.py` | Real-time data processing |
| **Live Execution Engine** | `live/execution_engine.py` | Real-time order execution |
| **Live Risk Engine** | `live/risk_engine.py` | Real-time risk management |
| **State Reconciliation** | `live/reconciliation.py` | Reconcile local state with exchange |
| **Retry Policies** | `live/retry.py` | Retry logic for network failures |

### Analysis & Persistence (entire subsystems absent)

| Missing Feature | Python Source | Description |
|-----------------|--------------|-------------|
| **Portfolio Analyzer** | `analysis/analyzer.py` | P&L, drawdown, Sharpe, Sortino, etc. |
| **Serialization** | `serialization/` | Arrow/Parquet serialization |
| **Data Catalog** | `persistence/catalog/` | Historical data management |
| **Data Wranglers** | `persistence/wranglers.pyx` | Transform raw data into NautilusTrader types |

### Adapters (entire subsystem absent)

The Python version has 16 exchange/data adapters (Binance, Kraken, Bybit, OKX, Interactive Brokers, Databento, etc.). The Go version has none.

---

## Summary: Priority Roadmap for Go Implementation

Based on what's needed to make the Go version a functional core:

**Tier 1 - Essential for a useful backtest engine:**
- More order types (Limit, Stop, StopLimit at minimum)
- Order event lifecycle / state machine
- Fill models (fees, slippage)
- Account/balance tracking
- Basic risk checks
- More indicators (SMA, RSI, MACD, ATR, Bollinger)
- Quote/Trade tick data support

**Tier 2 - Important for realistic backtesting:**
- Bar aggregation from ticks
- Order book support
- More instrument types (CurrencyPair, Equity, CryptoPerpetual)
- Execution engine with proper order routing
- Performance analysis / reporting
- Data persistence (Parquet read/write)

**Tier 3 - Required for live trading:**
- Live trading engine
- Adapter framework
- State reconciliation
- Retry/reconnection logic
- At least one exchange adapter

**Tier 4 - Nice to have:**
- Full indicator library
- All instrument types
- Execution algorithms
- Order emulator
- Data catalog
