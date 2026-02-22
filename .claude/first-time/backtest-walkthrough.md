# Backtest Walkthrough for First-Time Contributors

This guide walks through a complete NautilusTrader backtest from setup to execution,
showing exactly what happens at each step and which source files are involved.

## What is a Backtest?

A backtest replays historical market data through a simulated exchange, letting you
test trading strategies against past prices without risking real money. NautilusTrader's
key design principle is **parity**: the exact same strategy code runs in both backtest
and live environments, so behavior you observe in a backtest carries over to production.

## The Simplest Example

We'll start with the simplest backtest in the repository:
`examples/backtest/example_01_load_bars_from_custom_csv/`. It loads 1-minute bars
from a CSV file and runs a strategy that simply counts them.

### Step 1: Configure and create the engine

```python
# examples/backtest/example_01_load_bars_from_custom_csv/run_example.py

engine_config = BacktestEngineConfig(
    trader_id=TraderId("BACKTEST_TRADER-001"),
    logging=LoggingConfig(log_level="DEBUG"),
)
engine = BacktestEngine(config=engine_config)
```

The `BacktestEngine` constructor creates a `NautilusKernel` — the central orchestrator
that wires together every component:

```
BacktestEngine.__init__()                          # backtest/engine.pyx:228
  └─ NautilusKernel.__init__()                     # system/kernel.py:132
       ├─ TestClock()            # deterministic simulated clock
       ├─ MessageBus()           # pub/sub + point-to-point messaging
       ├─ Cache()                # in-memory store for instruments, orders, positions
       ├─ Portfolio()            # tracks account balances and positions
       ├─ DataEngine()           # routes market data to subscribers
       ├─ RiskEngine()           # validates orders before execution
       ├─ ExecutionEngine()      # routes orders to execution clients
       ├─ OrderEmulator()        # handles emulated order types
       └─ Trader()               # manages strategies and actors
```

### Step 2: Add a venue (simulated exchange)

```python
XCME = Venue("XCME")
engine.add_venue(
    venue=XCME,
    oms_type=OmsType.NETTING,
    account_type=AccountType.MARGIN,
    starting_balances=[Money(1_000_000, USD)],
    base_currency=USD,
    default_leverage=Decimal(1),
)
```

Under the hood (`engine.pyx:497`), this creates a `SimulatedExchange` with an order
matching engine, plus a `BacktestExecClient` that bridges the execution engine to
the simulated exchange.

### Step 3: Add an instrument

```python
EURUSD_FUTURES_INSTRUMENT = TestInstrumentProvider.eurusd_future(
    expiry_year=2024, expiry_month=3, venue_name="XCME",
)
engine.add_instrument(EURUSD_FUTURES_INSTRUMENT)
```

This registers the instrument in both the `Cache` (via `DataEngine.process()`) and
the `SimulatedExchange`.

### Step 4: Load and add data

```python
# Load CSV into pandas DataFrame
df = pd.read_csv("6EH4.XCME_1min_bars.csv", sep=";", header=0, index_col=False)

# Restructure: columns must be open, high, low, close, volume with timestamp index
df = df.reindex(columns=["timestamp_utc", "open", "high", "low", "close", "volume"])
df["timestamp_utc"] = pd.to_datetime(df["timestamp_utc"], format="%Y-%m-%d %H:%M:%S")
df = df.rename(columns={"timestamp_utc": "timestamp"}).set_index("timestamp")

# Define bar type and wrangle into Bar objects
bar_type = BarType.from_str(f"{EURUSD_FUTURES_INSTRUMENT.id}-1-MINUTE-LAST-EXTERNAL")
wrangler = BarDataWrangler(bar_type, EURUSD_FUTURES_INSTRUMENT)
bars: list[Bar] = wrangler.process(df)

# Add to engine
engine.add_data(bars)
```

`add_data()` (`engine.pyx:778`) stores the bars in an internal list and feeds them
into a `BacktestDataIterator` that will replay them in timestamp order during the run.

### Step 5: Add a strategy

```python
strategy = DemoStrategy(primary_bar_type=bar_type)
engine.add_strategy(strategy)
```

This delegates to `Trader.add_strategy()`, which registers the strategy with the
kernel's component infrastructure.

### Step 6: Run the backtest

```python
engine.run()
engine.dispose()
```

This is where all the action happens. See the next section for details.

## The DemoStrategy

Here's the complete strategy from
`examples/backtest/example_01_load_bars_from_custom_csv/strategy.py`:

```python
class DemoStrategy(Strategy):
    def __init__(self, primary_bar_type: BarType):
        super().__init__()
        self.primary_bar_type = primary_bar_type
        self.bars_processed = 0

    def on_start(self):
        self.subscribe_bars(self.primary_bar_type)

    def on_bar(self, bar: Bar):
        self.bars_processed += 1
        self.log.info(f"Processed bars: {self.bars_processed}", color=LogColor.YELLOW)

    def on_stop(self):
        self.log.info(f"Total bars processed: {self.bars_processed}")
```

Three lifecycle methods define the strategy:

| Method | Called when | Typical use |
|--------|-----------|-------------|
| `on_start()` | Engine starts the strategy | Subscribe to data feeds |
| `on_bar()` | A new bar arrives | Analyze data, submit orders |
| `on_stop()` | Engine stops the strategy | Log results, cancel orders |

## Execution Flow: What Happens When You Call `engine.run()`

### Phase 1: Initialization

```
BacktestEngine.run()                               # engine.pyx:1302
  └─ BacktestEngine._run()                         # engine.pyx:1446
       ├─ Set all component clocks to data start time
       ├─ Initialize exchange accounts
       └─ NautilusKernel.start()                   # kernel.py:987
            ├─ DataEngine.start()
            ├─ RiskEngine.start()
            ├─ ExecutionEngine.start()
            ├─ OrderEmulator.start()
            └─ Trader.start()
                 └─ Strategy._start()              # strategy.pyx:356
                      └─ Strategy.on_start()       # YOUR CODE
                           └─ subscribe_bars()     # registers with MessageBus
```

When `on_start()` calls `self.subscribe_bars(bar_type)`, the strategy registers its
`handle_bar` method as a subscriber on the MessageBus for the topic
`"data.bars.<BarType>"`. From this point on, any bar published to that topic will
be delivered to the strategy.

### Phase 2: Main Simulation Loop

This is the core of the backtest (`engine.pyx:1566`). For each data point in
timestamp order:

```
┌─────────────────────────────────────────────────────────────────────┐
│                     MAIN LOOP (for each data point)                 │
│                                                                     │
│  1. ADVANCE TIME                                                    │
│     Clock moves to data timestamp; pending timers fire              │
│                                                                     │
│  2. UPDATE SIMULATED EXCHANGE                                       │
│     Data goes to SimulatedExchange first                            │
│     └─ exchange.process_bar(bar)                                    │
│          └─ OrderMatchingEngine updates bid/ask prices              │
│          └─ Open orders checked for fills                           │
│                                                                     │
│  3. PUBLISH DATA TO STRATEGIES                                      │
│     DataEngine.process(bar)                                         │
│     └─ Validates and caches the bar                                 │
│     └─ MessageBus.publish("data.bars.<BarType>", bar)               │
│          └─ Strategy.handle_bar(bar)                                │
│               └─ Updates registered indicators                      │
│               └─ Strategy.on_bar(bar)    ← YOUR CODE RUNS HERE      │
│                    └─ (may call submit_order, etc.)                 │
│                                                                     │
│  4. PROCESS EXCHANGE COMMANDS                                       │
│     _process_and_settle_venues()                                    │
│     └─ Drain queued commands (orders submitted in step 3)           │
│     └─ Match and fill orders                                        │
│     └─ Publish fill events back to strategies                       │
│                                                                     │
│  5. NEXT DATA POINT → go to step 1                                  │
└─────────────────────────────────────────────────────────────────────┘
```

**Key insight**: The exchange sees data *before* strategies do (step 2 before step 3).
This ensures market state is updated before strategy logic runs. Orders submitted by
strategies in step 3 are queued and processed in step 4 — they don't fill instantly
within the same `on_bar()` call.

### Phase 3: Shutdown

```
BacktestEngine.end()                               # engine.pyx:1362
  ├─ Trader.stop()
  │    └─ Strategy._stop()
  │         └─ Strategy.on_stop()                  # YOUR CODE
  ├─ DataEngine.stop()
  ├─ RiskEngine.stop()
  ├─ ExecutionEngine.stop()
  └─ Final _process_and_settle_venues()            # drain remaining commands
```

## Adding Trading Logic: The EMA Cross Strategy

The `DemoStrategy` above doesn't trade. Let's look at `EMACross`
(`nautilus_trader/examples/strategies/ema_cross.py`) to see the full order flow.

### Strategy setup

```python
class EMACross(Strategy):
    def __init__(self, config: EMACrossConfig) -> None:
        super().__init__(config)
        self.fast_ema = ExponentialMovingAverage(config.fast_ema_period)
        self.slow_ema = ExponentialMovingAverage(config.slow_ema_period)

    def on_start(self) -> None:
        self.instrument = self.cache.instrument(self.config.instrument_id)
        # Register indicators so they auto-update on each bar
        self.register_indicator_for_bars(self.config.bar_type, self.fast_ema)
        self.register_indicator_for_bars(self.config.bar_type, self.slow_ema)
        self.subscribe_bars(self.config.bar_type)
```

`register_indicator_for_bars()` is the key addition — it tells the framework to
automatically feed each bar into the indicator's `handle_bar()` *before* calling
`on_bar()`. By the time your `on_bar()` runs, `self.fast_ema.value` is already
up to date.

### Trading logic in `on_bar()`

```python
def on_bar(self, bar: Bar) -> None:
    if not self.indicators_initialized():
        return  # wait for indicators to warm up

    # BUY: fast EMA crosses above slow EMA
    if self.fast_ema.value >= self.slow_ema.value:
        if self.portfolio.is_flat(self.config.instrument_id):
            self.buy()
        elif self.portfolio.is_net_short(self.config.instrument_id):
            self.close_all_positions(self.config.instrument_id)
            self.buy()

    # SELL: fast EMA crosses below slow EMA
    elif self.fast_ema.value < self.slow_ema.value:
        if self.portfolio.is_flat(self.config.instrument_id):
            self.sell()
        elif self.portfolio.is_net_long(self.config.instrument_id):
            self.close_all_positions(self.config.instrument_id)
            self.sell()
```

### Order submission

```python
def buy(self) -> None:
    order: MarketOrder = self.order_factory.market(
        instrument_id=self.config.instrument_id,
        order_side=OrderSide.BUY,
        quantity=self.instrument.make_qty(self.config.trade_size),
        time_in_force=TimeInForce.GTC,
    )
    self.submit_order(order)
```

## Order Flow: From Strategy to Fill

When `self.submit_order(order)` is called, here's the complete path through the system:

```
Strategy.submit_order(order)                       # strategy.pyx:805
  ├─ Publish OrderInitialized event
  ├─ Cache.add_order(order)
  ├─ Create SubmitOrder command
  └─ OrderManager.send_risk_command(command)
       └─ MessageBus.send("RiskEngine.execute")
            │
            ▼
RiskEngine._handle_submit_order()                  # risk/engine.pyx:415
  ├─ Validate reduce-only constraints
  ├─ Look up instrument from Cache
  ├─ Check order price/quantity precision
  ├─ Check order quantity within instrument limits
  ├─ Check account has sufficient balance
  ├─ Check trading state (not HALTED/REDUCING)
  ├─ Apply order submit rate throttling
  └─ MessageBus.send("ExecEngine.execute")
       │
       ▼
ExecutionEngine._handle_submit_order()             # execution/engine.pyx:1107
  ├─ Find execution client for venue
  └─ BacktestExecClient.submit_order(command)      # execution_client.pyx:104
       ├─ Generate OrderSubmitted event
       └─ SimulatedExchange.send(command)           # engine.pyx:3227
            └─ Queue command in _message_queue
                 │
                 │  ... main loop continues ...
                 │  ... next: _process_and_settle_venues() ...
                 │
                 ▼
SimulatedExchange._drain_commands()                # engine.pyx:3507
  └─ _process_trading_command(SubmitOrder)          # engine.pyx:3595
       └─ OrderMatchingEngine.process_order()       # engine.pyx:4726
            ├─ Validate precision, expiry, reduce-only
            └─ _process_market_order()
                 └─ fill_market_order()             # engine.pyx:5745
                      └─ fill_order()               # engine.pyx:7030
                           ├─ Calculate commission
                           └─ Generate OrderFilled event
                                │
                                ▼
                           MessageBus.send("ExecEngine.process")
                                │
                                ▼
ExecutionEngine.process(OrderFilled)
  └─ Update Cache (order status, position)
  └─ MessageBus.publish("events.order.*")
       │
       ▼
Strategy.handle_event(OrderFilled)                 # strategy.pyx:1886
  └─ Strategy.on_order_filled()                    # YOUR CODE (if overridden)
```

## The MessageBus: How Components Communicate

All inter-component communication flows through the `MessageBus`
(`common/component.pyx:2178`). It supports three patterns:

| Pattern | API | Used for |
|---------|-----|----------|
| **Point-to-point** | `register(endpoint, handler)` / `send(endpoint, msg)` | Commands: `"RiskEngine.execute"`, `"ExecEngine.execute"` |
| **Pub/Sub** | `subscribe(topic, handler)` / `publish(topic, msg)` | Market data: `"data.bars.*"`, events: `"events.order.*"` |
| **Request/Response** | `request(endpoint, req)` / `response(resp)` | Historical data requests |

When a strategy calls `subscribe_bars(bar_type)`, it registers on the topic
`"data.bars.<BarType>"`. When the DataEngine processes a bar, it publishes to that
same topic, and the MessageBus delivers it to all matching subscribers.

## Key Components Quick Reference

| Component | File | Purpose |
|-----------|------|---------|
| `BacktestEngine` | `nautilus_trader/backtest/engine.pyx:212` | Top-level orchestrator; owns the main simulation loop |
| `NautilusKernel` | `nautilus_trader/system/kernel.py:101` | Wires together all components; shared by backtest and live |
| `MessageBus` | `nautilus_trader/common/component.pyx:2178` | Pub/sub and point-to-point messaging backbone |
| `Cache` | `nautilus_trader/cache/cache.pyx` | In-memory store for instruments, orders, positions, bars |
| `DataEngine` | `nautilus_trader/data/engine.pyx` | Routes market data to subscribers via MessageBus |
| `RiskEngine` | `nautilus_trader/risk/engine.pyx` | Pre-trade checks (balance, limits, throttling) |
| `ExecutionEngine` | `nautilus_trader/execution/engine.pyx` | Routes orders to execution clients |
| `SimulatedExchange` | `nautilus_trader/backtest/engine.pyx:2579` | Simulates a venue with order book and matching |
| `OrderMatchingEngine` | `nautilus_trader/backtest/engine.pyx:3737` | Matches and fills orders against simulated book |
| `Strategy` | `nautilus_trader/trading/strategy.pyx` | Base class for all trading strategies |
| `BacktestExecClient` | `nautilus_trader/backtest/execution_client.pyx` | Bridges ExecutionEngine to SimulatedExchange |
| `Portfolio` | `nautilus_trader/portfolio/portfolio.pyx` | Tracks positions, balances, and P&L |

## Where to Go Next

- **More examples**: Browse `examples/backtest/` for progressively complex setups
  (multiple venues, order book data, custom indicators)
- **Write your own strategy**: Subclass `Strategy`, implement `on_start()` and
  `on_bar()`, and use `self.submit_order()` to trade
- **Developer guide**: See `docs/developer_guide/` for architecture deep-dives
- **Live trading**: The same strategy code works live — only the engine configuration
  changes (use `TradingNode` instead of `BacktestEngine`)
