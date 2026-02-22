Here's a guide for first-time contributors on integrating a new exchange adapter.

---

## What Adapters Do

Adapters are **pluggable modules** that connect NautilusTrader to external exchanges/venues. They translate between:
- **Exchange-specific APIs** (REST, WebSocket, authentication, message formats)
- **NautilusTrader's unified domain model** (orders, fills, instruments, market data)

This abstraction lets you write strategy code once and run it against any supported venue.

---

## Directory Structure

Create your adapter at `nautilus_trader/adapters/your_exchange/`:

```
your_exchange/
├── __init__.py           # Public exports
├── core.py               # Constants (venue identifier)
├── config.py             # Configuration classes
├── data.py               # Market data client
├── execution.py          # Order execution client
├── factories.py          # Factory functions (required for integration)
├── providers.py          # Instrument provider
├── http/                 # REST API client (optional subpackage)
│   └── client.py
├── websocket/            # WebSocket client (optional subpackage)
│   └── client.py
└── schemas/              # API response schemas (msgspec)
```

---

## Core Interfaces to Implement

### 1. Data Client (`LiveMarketDataClient`)

**File:** `nautilus_trader/live/data_client.py`

**Required methods:**
| Method | Purpose |
|--------|---------|
| `_connect()` | Establish connection to exchange |
| `_disconnect()` | Clean up connections |

**Optional methods (implement what your exchange supports):**
| Method | Purpose |
|--------|---------|
| `_subscribe_trade_ticks()` | Real-time trade feed |
| `_subscribe_quote_ticks()` | Real-time bid/ask quotes |
| `_subscribe_order_book_deltas()` | Order book updates |
| `_subscribe_bars()` | OHLCV candle data |
| `_request_instrument()` | Fetch instrument info |
| `_request_bars()` | Historical bar data |

### 2. Execution Client (`LiveExecutionClient`)

**File:** `nautilus_trader/live/execution_client.py`

**Required methods:**
| Method | Purpose |
|--------|---------|
| `_connect()` / `_disconnect()` | Connection lifecycle |
| `_submit_order()` | Send order to exchange |
| `_cancel_order()` | Cancel a single order |
| `_cancel_all_orders()` | Cancel all open orders |
| `generate_order_status_report()` | Query single order status |
| `generate_order_status_reports()` | Query all order statuses |
| `generate_fill_reports()` | Query trade/fill history |
| `generate_position_status_reports()` | Query position state |
| `generate_mass_status()` | Full reconciliation report |

**Optional methods:**
| Method | Purpose |
|--------|---------|
| `_modify_order()` | Amend order price/quantity |
| `_submit_order_list()` | Submit bracket/OCO orders |
| `_batch_cancel_orders()` | Cancel multiple orders at once |

### 3. Configuration Classes

Extend from base configs in `nautilus_trader/live/config.py`:

```python
from nautilus_trader.live.config import LiveDataClientConfig
from nautilus_trader.live.config import LiveExecClientConfig

class MyExchangeDataClientConfig(LiveDataClientConfig):
    api_key: str
    api_secret: str
    testnet: bool = False

class MyExchangeExecClientConfig(LiveExecClientConfig):
    api_key: str
    api_secret: str
    testnet: bool = False
```

### 4. Factory Classes

**Critical for integration.** The trading node uses factories to instantiate your clients:

```python
from nautilus_trader.live.factories import LiveDataClientFactory
from nautilus_trader.live.factories import LiveExecClientFactory

class MyExchangeLiveDataClientFactory(LiveDataClientFactory):
    @staticmethod
    def create(loop, name, config, msgbus, cache, clock) -> LiveDataClient:
        return MyExchangeDataClient(...)

class MyExchangeLiveExecClientFactory(LiveExecClientFactory):
    @staticmethod
    def create(loop, name, config, msgbus, cache, clock) -> LiveExecutionClient:
        return MyExchangeExecClient(...)
```

---

## Key Patterns

### Publishing Data to the System

Use inherited methods to publish data to the message bus:

```python
# In your data client
self._handle_trade_tick(trade_tick)      # Publish trade
self._handle_quote_tick(quote_tick)      # Publish quote  
self._handle_order_book_delta(delta)     # Publish book update
self._handle_bar(bar)                    # Publish OHLCV bar
```

### Generating Execution Events

```python
# In your execution client
self.generate_order_accepted(...)    # Order acknowledged
self.generate_order_filled(...)      # Trade executed
self.generate_order_canceled(...)    # Order cancelled
self.generate_order_rejected(...)    # Order rejected
```

### Symbol Mapping

Convert between exchange symbols and Nautilus `InstrumentId`:

```python
# Exchange: "BTCUSDT" -> Nautilus: "BTCUSDT.MYEXCHANGE"
instrument_id = InstrumentId(Symbol("BTCUSDT"), MYEXCHANGE_VENUE)
```

---

## Getting Started Checklist

1. **Copy the template** from `nautilus_trader/adapters/_template/`

2. **Define your venue** in `core.py`:
   ```python
   MYEXCHANGE_VENUE = Venue("MYEXCHANGE")
   ```

3. **Implement data client** - Start with `_connect()`, `_subscribe_trade_ticks()`

4. **Implement execution client** - Start with `_connect()`, `_submit_order()`, `_cancel_order()`

5. **Create config classes** with API credentials and settings

6. **Create factory classes** to wire everything together

7. **Write tests** - See `tests/integration_tests/adapters/` for examples

---

## Reference Implementations

Study these existing adapters (ordered by complexity):

| Adapter | Path | Notes |
|---------|------|-------|
| **Template** | `adapters/_template/` | Minimal skeleton |
| **Kraken** | `adapters/kraken/` | Single product type, simpler |
| **Binance** | `adapters/binance/` | Multi-product (spot/futures), complex |

The Kraken adapter is a good middle-ground reference - full-featured but not overly complex.