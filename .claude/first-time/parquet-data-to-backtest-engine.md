#nautilus-trader #trading 
Now I have the complete picture. Here's the data flow from **ParquetDataCatalog** → **BacktestEngine** → **Strategy**:

---

## Data Flow: Catalog → BacktestEngine → Strategy

```
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              CONFIGURATION LAYER                                │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│   BacktestRunConfig                                                             │
│   ├── venues: list[BacktestVenueConfig]     # Simulated exchange config         │
│   ├── data: list[BacktestDataConfig]        # What data to load from catalog    │
│   │       ├── catalog_path: str                                                 │
│   │       ├── data_cls: str                 # "QuoteTick", "TradeTick", "Bar"   │
│   │       ├── instrument_id: InstrumentId                                       │
│   │       ├── start_time / end_time                                             │
│   │       └── bar_spec: str (for bars)                                          │
│   └── engine: BacktestEngineConfig          # Strategy/Actor configuration      │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
                                       │
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              1. DATA LOADING LAYER                              │
│                                  BacktestNode                                   │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│   ParquetDataCatalog.query()                                                    │
│   ├── Uses Rust backend (DataBackendSession) for core types:                    │
│   │   OrderBookDelta, QuoteTick, TradeTick, Bar, OrderBookDepth10               │
│   │                                                                             │
│   └── Returns: list[Data]  (QuoteTick, TradeTick, Bar, etc.)                    │
│       └── Data objects are Nautilus domain objects, fully deserialized          │
│                                                                                 │
│   Two modes:                                                                    │
│   ├── Oneshot: Load all data → engine.add_data() → engine.run()                 │
│   └── Streaming: chunk_size → iterate session → add_data/run/clear per chunk    │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
                                       │
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              2. ENGINE DATA INTERFACE                           │
│                                  BacktestEngine                                 │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│   engine.add_data(data: list[Data])                                             │
│   ├── Validates data types & instruments exist in cache                         │
│   ├── Tracks which instruments have data (self._has_data)                       │
│   ├── Sorts data by ts_init (chronological order)                               │
│   └── Stores in internal list: self._data                                       │
│                                                                                 │
│   engine.add_instrument(instrument: Instrument)                                 │
│   └── Must be called BEFORE add_data for that instrument                        │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
                                       │
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              3. BACKTEST RUN LOOP                               │
│                              engine._run() main loop                            │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│   for data in self._data_iterator:                                              │
│       │                                                                         │
│       ├── 1. Advance simulated clock to data.ts_init                            │
│       │                                                                         │
│       ├── 2. Process through SimulatedExchange (venue simulation):              │
│       │      exchange.process_quote_tick(data)   # Updates order book           │
│       │      exchange.process_trade_tick(data)   # Triggers order matching      │
│       │      exchange.process_bar(data)          # Bar-based execution          │
│       │      exchange.process_order_book_delta() # L2/L3 book updates           │
│       │                                                                         │
│       ├── 3. Process through DataEngine:                                        │
│       │      data_engine.process(data)                                          │
│       │      └── Publishes to MessageBus topics                                 │
│       │                                                                         │
│       └── 4. Process exchange messages (fills, etc.)                            │
│              exchange.process(data.ts_init)                                     │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
                                       │
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              4. DATA ENGINE ROUTING                             │
│                                  DataEngine                                     │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│   _handle_data(data):                                                           │
│   ├── QuoteTick  → cache.add_quote_tick() → publish to topic                    │
│   ├── TradeTick  → cache.add_trade_tick() → publish to topic                    │
│   ├── Bar        → cache.add_bar()        → publish to topic                    │
│   ├── OrderBook* → (updates internal book) → publish to topic                   │
│   └── CustomData → publish to custom topic                                      │
│                                                                                 │
│   Topic format examples:                                                        │
│   ├── "data.quotes.BTCUSDT.BINANCE"                                             │
│   ├── "data.trades.BTCUSDT.BINANCE"                                             │
│   └── "data.bars.BTCUSDT.BINANCE-1-MINUTE-LAST-EXTERNAL"                        │
│                                                                                 │
│   msgbus.publish_c(topic=..., msg=data)                                         │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
                                       │
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              5. MESSAGE BUS (Pub/Sub)                           │
│                                  MessageBus                                     │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│   Subscriptions (registered by Strategy/Actor):                                 │
│                                                                                 │
│   strategy.subscribe_quote_ticks(instrument_id)                                 │
│   └── msgbus.subscribe(topic="data.quotes.{id}", handler=handle_quote_tick)     │
│                                                                                 │
│   strategy.subscribe_trade_ticks(instrument_id)                                 │
│   └── msgbus.subscribe(topic="data.trades.{id}", handler=handle_trade_tick)     │
│                                                                                 │
│   strategy.subscribe_bars(bar_type)                                             │
│   └── msgbus.subscribe(topic="data.bars.{bar_type}", handler=handle_bar)        │
│                                                                                 │
│   When data published → all matching handlers invoked synchronously             │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
                                       │
                                       ▼
┌─────────────────────────────────────────────────────────────────────────────────┐
│                              6. STRATEGY HANDLERS                               │
│                              Strategy (extends Actor)                           │
├─────────────────────────────────────────────────────────────────────────────────┤
│                                                                                 │
│   # User implements these callback methods:                                     │
│                                                                                 │
│   def on_quote_tick(self, tick: QuoteTick) -> None:                             │
│       # React to quote updates                                                  │
│       pass                                                                      │
│                                                                                 │
│   def on_trade_tick(self, tick: TradeTick) -> None:                             │
│       # React to trade updates                                                  │
│       pass                                                                      │
│                                                                                 │
│   def on_bar(self, bar: Bar) -> None:                                           │
│       # React to bar close                                                      │
│       if bar.bar_type == self.bar_type:                                         │
│           self.submit_order(...)                                                │
│                                                                                 │
│   def on_order_book_deltas(self, deltas: OrderBookDeltas) -> None:              │
│       # React to order book changes                                             │
│       pass                                                                      │
│                                                                                 │
└─────────────────────────────────────────────────────────────────────────────────┘
```

---

## Key Interfaces Summary

### 1. BacktestEngine Data Ingestion Interface

```python
# Primary methods for adding data
engine.add_instrument(instrument: Instrument)     # Must add instruments first
engine.add_data(data: list[Data], sort=True)      # Add market data
engine.add_venue(venue, ...)                      # Add simulated exchange

# Run control
engine.run(start=..., end=..., streaming=False)
engine.clear_data()                               # For streaming mode
```

### 2. Transformations Applied

| Stage | Transformation |
|-------|----------------|
| **Parquet → Python** | Arrow bytes → Nautilus domain objects (via wranglers) |
| **add_data()** | Sort by `ts_init`, validate instruments exist |
| **Run loop** | Chronological iteration, clock advancement |
| **Exchange** | Updates simulated order book, triggers order matching |
| **DataEngine** | Caches data, publishes to message bus topics |
| **Strategy** | Receives via handler callbacks (already domain objects) |

### 3. Data Types Supported for Backtest

| Data Type | Exchange Processing | Strategy Handler |
|-----------|---------------------|------------------|
| `QuoteTick` | Updates L1 book | `on_quote_tick()` |
| `TradeTick` | Order matching trigger | `on_trade_tick()` |
| `Bar` | OHLCV execution | `on_bar()` |
| `OrderBookDelta` | L2/L3 book update | `on_order_book_deltas()` |
| `OrderBookDepth10` | Full depth snapshot | `on_order_book_depth()` |
| `InstrumentStatus` | Market status | `on_instrument_status()` |
| `CustomData` | Pass-through | `on_data()` |

The data arrives at your strategy **as native Nautilus objects** - no transformation needed in strategy code. The Parquet deserialization happens once during catalog load.