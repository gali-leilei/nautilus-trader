# NautilusTrader Data Persistence Protocols

A language-agnostic reference for reading NautilusTrader persisted data from
any language (JavaScript, Go, Python, Rust, etc.) **without** depending on the
NautilusTrader libraries.

---

## Table of Contents

1. [Historical Tick Data (Parquet Catalog)](#1-historical-tick-data-parquet-catalog)
2. [User Account & Portfolio State (Redis)](#2-user-account--portfolio-state-redis)
3. [Order & Fill Information](#3-order--fill-information)
4. [Backtest Results & Strategy Performance](#4-backtest-results--strategy-performance)
5. [Cross-Cutting Gotchas](#5-cross-cutting-gotchas)

---

## 1. Historical Tick Data (Parquet Catalog)

### File Format

Apache Parquet, written by PyArrow's `pq.write_table()` with **default
compression** (Snappy). Any Parquet reader that supports Snappy will work.

### Directory Layout

```
{catalog_path}/
└── data/
    ├── quote_tick/
    │   └── {instrument_id}/
    │       └── {start_ts}_{end_ts}.parquet
    ├── trade_tick/
    │   └── {instrument_id}/
    │       └── {start_ts}_{end_ts}.parquet
    ├── bar/
    │   └── {bar_type}/
    │       └── {start_ts}_{end_ts}.parquet
    ├── order_book_deltas/
    │   └── {instrument_id}/
    │       └── {start_ts}_{end_ts}.parquet
    ├── order_book_depths/
    │   └── {instrument_id}/
    │       └── {start_ts}_{end_ts}.parquet
    ├── mark_price_update/
    │   └── {instrument_id}/
    │       └── {start_ts}_{end_ts}.parquet
    ├── funding_rate_update/
    │   └── {instrument_id}/
    │       └── {start_ts}_{end_ts}.parquet
    ├── instrument_close/
    │   └── {start_ts}_{end_ts}.parquet
    ├── instrument_status/
    │   └── {start_ts}_{end_ts}.parquet
    ├── instrument/
    │   └── {start_ts}_{end_ts}.parquet
    └── custom_{snake_case_name}/
        └── {identifier}/
            └── {start_ts}_{end_ts}.parquet
```

**Subdirectory names** are the snake_case data type (see
[Class-to-Filename Mapping](#class-to-filename-mapping)).

**Instrument ID directories** use a URI-safe encoding: forward slashes are
simply removed.

| Original ID | Directory Name |
|---|---|
| `BTC/USDT.BINANCE` | `BTCUSDT.BINANCE` |
| `EUR/USD.SIM` | `EURUSD.SIM` |
| `AAPL.XNAS` | `AAPL.XNAS` (unchanged) |

### Filename Convention

```
{start_ts}_{end_ts}.parquet
```

Each timestamp is an ISO 8601 string in UTC with nanosecond precision, but
with colons and dots replaced by hyphens to be filesystem-safe:

| Unix nanos | ISO 8601 | Filename form |
|---|---|---|
| `1698307850123456789` | `2023-10-26T07:30:50.123456789Z` | `2023-10-26T07-30-50-123456789Z` |

Intervals within a directory are **disjoint** (non-overlapping).

### Class-to-Filename Mapping

| Python Class | Directory Name |
|---|---|
| `QuoteTick` | `quote_tick` |
| `TradeTick` | `trade_tick` |
| `Bar` | `bar` |
| `OrderBookDelta` / `OrderBookDeltas` | `order_book_deltas` |
| `OrderBookDepth10` | `order_book_depths` |
| `MarkPriceUpdate` | `mark_price_update` |
| `FundingRateUpdate` | `funding_rate_update` |
| `InstrumentClose` | `instrument_close` |
| `InstrumentStatus` | `instrument_status` |
| Custom data class `Foo` | `custom_foo` |

### Arrow Schemas (Rust-Written Parquet)

All schemas carry **metadata** in the schema header. Common metadata keys:

| Key | Description | Example |
|---|---|---|
| `instrument_id` | Full instrument identifier | `BTCUSDT-PERP.BINANCE` |
| `price_precision` | Decimal places for prices | `2` |
| `size_precision` | Decimal places for quantities | `4` |
| `bar_type` | Bar specification (bars only) | `BTCUSDT.BINANCE-1-MINUTE-LAST-EXTERNAL` |

#### QuoteTick

| Column | Arrow Type | Signed? | Description |
|---|---|---|---|
| `bid_price` | `FixedSizeBinary(N)` | signed | Best bid price (fixed-point) |
| `ask_price` | `FixedSizeBinary(N)` | signed | Best ask price (fixed-point) |
| `bid_size` | `FixedSizeBinary(N)` | unsigned | Bid quantity (fixed-point) |
| `ask_size` | `FixedSizeBinary(N)` | unsigned | Ask quantity (fixed-point) |
| `ts_event` | `UInt64` | — | Exchange timestamp (Unix nanos) |
| `ts_init` | `UInt64` | — | System init timestamp (Unix nanos) |

`N` = 16 bytes in high-precision mode, 8 bytes in standard-precision mode.
See [Fixed-Point Decoding](#fixed-point-decoding) below.

#### TradeTick

| Column | Arrow Type | Signed? | Description |
|---|---|---|---|
| `price` | `FixedSizeBinary(N)` | signed | Trade price |
| `size` | `FixedSizeBinary(N)` | unsigned | Trade size |
| `aggressor_side` | `UInt8` | — | Enum (see below) |
| `trade_id` | `Utf8` | — | Exchange trade ID |
| `ts_event` | `UInt64` | — | Exchange timestamp |
| `ts_init` | `UInt64` | — | System init timestamp |

**AggressorSide enum:**

| Value | Meaning |
|---|---|
| `0` | NoAggressor |
| `1` | Buyer |
| `2` | Seller |

#### Bar

| Column | Arrow Type | Signed? | Description |
|---|---|---|---|
| `open` | `FixedSizeBinary(N)` | signed | Open price |
| `high` | `FixedSizeBinary(N)` | signed | High price |
| `low` | `FixedSizeBinary(N)` | signed | Low price |
| `close` | `FixedSizeBinary(N)` | signed | Close price |
| `volume` | `FixedSizeBinary(N)` | unsigned | Bar volume |
| `ts_event` | `UInt64` | — | Bar close timestamp |
| `ts_init` | `UInt64` | — | System init timestamp |

Metadata uses `bar_type` instead of `instrument_id`.

#### OrderBookDelta

| Column | Arrow Type | Signed? | Description |
|---|---|---|---|
| `action` | `UInt8` | — | BookAction enum |
| `side` | `UInt8` | — | OrderSide enum |
| `price` | `FixedSizeBinary(N)` | signed | Order price |
| `size` | `FixedSizeBinary(N)` | unsigned | Order size |
| `order_id` | `UInt64` | — | Order book entry ID |
| `flags` | `UInt8` | — | Bit flags |
| `sequence` | `UInt64` | — | Sequence number |
| `ts_event` | `UInt64` | — | Exchange timestamp |
| `ts_init` | `UInt64` | — | System init timestamp |

**BookAction enum:**

| Value | Meaning |
|---|---|
| `1` | Add |
| `2` | Update |
| `3` | Delete |
| `4` | Clear |

**OrderSide enum:**

| Value | Meaning |
|---|---|
| `0` | NoOrderSide |
| `1` | Buy |
| `2` | Sell |

#### OrderBookDepth10

64 columns total — 10 levels of bid/ask price/size/count:

| Columns | Arrow Type | Count | Description |
|---|---|---|---|
| `bid_price_0` .. `bid_price_9` | `FixedSizeBinary(N)` (signed) | 10 | Bid prices |
| `ask_price_0` .. `ask_price_9` | `FixedSizeBinary(N)` (signed) | 10 | Ask prices |
| `bid_size_0` .. `bid_size_9` | `FixedSizeBinary(N)` (unsigned) | 10 | Bid sizes |
| `ask_size_0` .. `ask_size_9` | `FixedSizeBinary(N)` (unsigned) | 10 | Ask sizes |
| `bid_count_0` .. `bid_count_9` | `UInt32` | 10 | Bid order counts |
| `ask_count_0` .. `ask_count_9` | `UInt32` | 10 | Ask order counts |
| `flags` | `UInt8` | 1 | Bit flags |
| `sequence` | `UInt64` | 1 | Sequence number |
| `ts_event` | `UInt64` | 1 | Exchange timestamp |
| `ts_init` | `UInt64` | 1 | System init timestamp |

#### InstrumentClose

| Column | Arrow Type | Description |
|---|---|---|
| `close_price` | `FixedSizeBinary(N)` (signed) | Close price |
| `close_type` | `UInt8` | InstrumentCloseType enum |
| `ts_event` | `UInt64` | Exchange timestamp |
| `ts_init` | `UInt64` | System init timestamp |

#### MarkPriceUpdate / IndexPriceUpdate

| Column | Arrow Type | Description |
|---|---|---|
| `value` | `FixedSizeBinary(N)` (signed) | Price value |
| `ts_event` | `UInt64` | Exchange timestamp |
| `ts_init` | `UInt64` | System init timestamp |

### Fixed-Point Decoding

This is the **most critical detail** for any external consumer. Prices and
quantities are stored as fixed-point integers in `FixedSizeBinary` columns,
**not** floating-point numbers.

#### Precision Modes

| Mode | Feature Flag | Byte Width | Integer Type (Price) | Integer Type (Quantity) | Scalar |
|---|---|---|---|---|---|
| **High-precision** (default) | `high-precision` | 16 | `i128` (signed) | `u128` (unsigned) | `10^16` |
| **Standard-precision** | _(none)_ | 8 | `i64` (signed) | `u64` (unsigned) | `10^9` |

#### Detecting the Mode

Read the byte width of any `FixedSizeBinary` column from the Parquet schema:

- **16 bytes** → high-precision mode, divide by `10^16`
- **8 bytes** → standard-precision mode, divide by `10^9`

#### Decoding Formula

```
float_value = raw_integer / SCALAR
```

Where:
- `raw_integer` = little-endian signed (prices) or unsigned (quantities)
  integer read from the `FixedSizeBinary` bytes
- `SCALAR` = `10_000_000_000_000_000` (10^16) for high-precision, or
  `1_000_000_000` (10^9) for standard-precision

#### Byte Order

**Little-endian** throughout. Rust writes via `.to_le_bytes()` and reads via
`::from_le_bytes()`.

#### Worked Example (High-Precision)

A price of `101.25` with `price_precision=2`:

```
Encoding:
  raw = round(101.25 * 10^2) * 10^(16-2) = 10125 * 10^14
      = 10125_00000000000000  (i128)
  bytes = little-endian encoding of 10125_00000000000000 as 16 bytes

Decoding (what you do):
  1. Read 16 bytes from the FixedSizeBinary column
  2. Interpret as little-endian signed 128-bit integer:
     raw = 10125_00000000000000
  3. Divide by 10^16:
     price = 10125_00000000000000 / 10_000_000_000_000_000 = 101.25
```

#### Worked Example (Standard-Precision)

A price of `101.25` with `price_precision=2`:

```
Encoding:
  raw = round(101.25 * 10^2) * 10^(9-2) = 10125 * 10^7
      = 101250000000  (i64)
  bytes = little-endian encoding of 101250000000 as 8 bytes

Decoding:
  1. Read 8 bytes from the FixedSizeBinary column
  2. Interpret as little-endian signed 64-bit integer:
     raw = 101250000000
  3. Divide by 10^9:
     price = 101250000000 / 1_000_000_000 = 101.25
```

#### Sentinel Values

| Sentinel | Meaning | Value |
|---|---|---|
| `PRICE_UNDEF` | Price not set | `i128::MAX` / `i64::MAX` |
| `PRICE_ERROR` | Error sentinel | `i128::MIN` / `i64::MIN` |
| `QUANTITY_UNDEF` | Quantity not set | `u128::MAX` / `u64::MAX` |

Check for these before dividing by the scalar.

#### Precision from Schema Metadata

The `price_precision` and `size_precision` metadata values tell you the
*logical* decimal precision. You do **not** need them to decode — just divide
by the scalar. They are useful for rounding the result to the correct number
of decimal places:

```
price_display = round(raw / SCALAR, price_precision)
```

---

## 2. User Account & Portfolio State (Redis)

### Backend

Primary: **Redis** (>= 6.2.0, required for streams).
Alternative: PostgreSQL (partial support, not covered here).

### Wire Format

Each value stored in Redis is serialized as either **MessagePack** (default)
or **JSON**, controlled by the `encoding` config setting.

**Serialization pipeline:**

```
Object → Serde JSON Value (intermediate) → Timestamp Conversion → Final Encoding
```

**Timestamp handling:** Fields matching `ts_*` (e.g., `ts_event`, `ts_init`)
or `expire_time_ns` are converted between Unix nanoseconds and RFC 3339
ISO 8601 strings during serialization. The `timestamps_as_iso8601` config
flag controls this behavior. When `false` (default), timestamps remain as
nanosecond integers.

### Connection Architecture

NautilusTrader uses two Redis connections:

- **READ** connection: synchronous queries for loading state
- **WRITE** connection: background async task, receives commands via an
  unbounded MPSC channel

### Key Pattern

```
{trader_prefix}:{collection}:{entity_id}
```

Where `trader_prefix` depends on configuration:

| Config | Prefix Format | Example |
|---|---|---|
| `use_trader_prefix=true, use_instance_id=false` | `trader-{trader_id}` | `trader-TESTER-001` |
| `use_trader_prefix=true, use_instance_id=true` | `trader-{trader_id}:{instance_id}` | `trader-TESTER-001:abc-uuid-123` |
| `use_trader_prefix=false` | _(no prefix)_ | `orders:O-123456` |

### Collections and Redis Data Types

| Collection | Redis Type | Key Example | Description |
|---|---|---|---|
| `currencies` | STRING | `...:currencies:USD` | Currency definitions |
| `instruments` | STRING | `...:instruments:BTCUSDT.BINANCE` | Instrument specs |
| `synthetics` | STRING | `...:synthetics:SPREAD-1` | Synthetic instruments |
| `accounts` | **LIST** | `...:accounts:BINANCE-001` | Account event stream |
| `orders` | **LIST** | `...:orders:O-20231026-001` | Order event stream |
| `positions` | **LIST** | `...:positions:P-001` | Position event stream |
| `actors` | STRING | `...:actors:{actor_id}:{key}` | Actor persisted state |
| `strategies` | STRING | `...:strategies:{strategy_id}:{key}` | Strategy persisted state |
| `general` | STRING | `...:general:{key}` | General key-value |
| `snapshots` | LIST | `...:snapshots:orders:{id}` | Optional snapshots |
| `health` | STRING | `...:health:{key}` | Health/monitoring |

### Index Keys (Sets and Hashes)

Index keys live under the same trader prefix:

| Key | Redis Type | Description |
|---|---|---|
| `index:order_ids` | SET | All client order IDs |
| `index:order_position` | HASH | Maps `client_order_id` → `position_id` |
| `index:order_client` | HASH | Maps `client_order_id` → `client_id` |
| `index:orders` | SET | All order IDs |
| `index:orders_open` | SET | Currently open order IDs |
| `index:orders_closed` | SET | Closed order IDs |
| `index:orders_emulated` | SET | Emulated order IDs |
| `index:orders_inflight` | SET | In-flight order IDs |
| `index:positions` | SET | All position IDs |
| `index:positions_open` | SET | Currently open position IDs |
| `index:positions_closed` | SET | Closed position IDs |

### Event-Sourcing Model

**Orders, positions, and accounts** are stored as **event lists** (Redis
LISTs). Each list entry is a serialized event (MessagePack or JSON). Objects
are reconstructed by replaying all events in the list from index 0:

```
LRANGE trader-TESTER:orders:O-20231026-001 0 -1
→ [event_0_bytes, event_1_bytes, event_2_bytes, ...]
```

- `event_0` = `OrderInitialized` (always first)
- `event_1` = `OrderSubmitted`
- `event_2` = `OrderAccepted`
- ...
- `event_N` = `OrderFilled` or `OrderCanceled` (terminal)

New events are appended via `RPUSH`. The `Update` command uses
`rpush_exists()` — it only appends if the key already exists.

**Instruments, currencies, synthetics** are stored as simple STRING values
— the entire serialized object, overwritten on each update.

### Message Bus Streams

Redis Streams are used for the event bus (separate from cache):

```
{trader_prefix}:stream:{topic}
```

Each stream entry has two fields:

| Field | Description |
|---|---|
| `topic` | Message topic string |
| `payload` | Serialized message bytes |

Auto-trimming via `XTRIM MINID` is optional, configured by `autotrim_mins`.

### Configuration Reference

```python
# CacheConfig (controls Redis cache behavior)
encoding: str = "msgpack"              # "msgpack" or "json"
timestamps_as_iso8601: bool = False    # True = RFC 3339 strings
buffer_interval_ms: int | None = None  # Write batching interval
bulk_read_batch_size: int | None = None
use_trader_prefix: bool = True
use_instance_id: bool = False
flush_on_start: bool = False

# MessageBusConfig (controls Redis streams)
autotrim_mins: int | None = None
stream_per_topic: bool = True
streams_prefix: str = "stream"
```

### Reading Redis Data from Another Language

1. Connect to Redis (default `127.0.0.1:6379`)
2. `SCAN 0 MATCH trader-*:orders:* COUNT 5000` to discover order keys
3. `LRANGE <key> 0 -1` to get all events for an order
4. Deserialize each entry as MessagePack (or JSON) — the result is a
   dictionary/map with string keys
5. Timestamps in `ts_*` fields are RFC 3339 strings if
   `timestamps_as_iso8601` was set, otherwise integer nanoseconds

---

## 3. Order & Fill Information

Order and fill data lives in two places: **Redis** (live state) and **Feather
files** (streaming output during backtests/live runs).

### In Redis

See [Section 2](#2-user-account--portfolio-state-redis). Key patterns:

```
{trader_prefix}:orders:{client_order_id}    → LIST of order events
{trader_prefix}:positions:{position_id}     → LIST of position events
{trader_prefix}:accounts:{account_id}       → LIST of account events
```

Index lookups:

```
SMEMBERS {trader_prefix}:index:orders_open
SMEMBERS {trader_prefix}:index:orders_closed
HGET {trader_prefix}:index:order_position {client_order_id}
```

### In Feather Files (Arrow IPC Streaming)

During backtest and live runs, a `StreamingFeatherWriter` writes events to
Feather (Arrow IPC streaming format) files:

```
{catalog_path}/{backtest|live}/{instance_id}/
├── quote_tick/
│   └── {instrument_id}/
│       └── {instrument_id}_{timestamp_ns}.feather
├── trade_tick/
│   └── {instrument_id}/
│       └── {instrument_id}_{timestamp_ns}.feather
├── bar/
│   └── {bar_type}/
│       └── {bar_type}_{timestamp_ns}.feather
├── order_book_deltas/
│   └── {instrument_id}/
│       └── {instrument_id}_{timestamp_ns}.feather
├── order_initialized_{timestamp_ns}.feather
├── order_denied_{timestamp_ns}.feather
├── order_submitted_{timestamp_ns}.feather
├── order_accepted_{timestamp_ns}.feather
├── order_rejected_{timestamp_ns}.feather
├── order_canceled_{timestamp_ns}.feather
├── order_expired_{timestamp_ns}.feather
├── order_triggered_{timestamp_ns}.feather
├── order_updated_{timestamp_ns}.feather
├── order_filled_{timestamp_ns}.feather
├── order_pending_cancel_{timestamp_ns}.feather
├── order_pending_update_{timestamp_ns}.feather
├── order_cancel_rejected_{timestamp_ns}.feather
├── order_modify_rejected_{timestamp_ns}.feather
├── order_emulated_{timestamp_ns}.feather
├── order_released_{timestamp_ns}.feather
├── position_opened_{timestamp_ns}.feather
├── position_changed_{timestamp_ns}.feather
├── position_closed_{timestamp_ns}.feather
├── account_state_{timestamp_ns}.feather
└── ...
```

Per-instrument data types (`quote_tick`, `trade_tick`, `bar`,
`order_book_deltas`, `order_book_depths`, `funding_rate_update`) get
subdirectories. All other event types are flat files at the run root.

The `{timestamp_ns}` in feather filenames is the system clock nanosecond
timestamp at file creation time.

### Order Event Arrow Schemas (Python Layer)

These schemas apply to the Feather files written by the Python
`StreamingFeatherWriter`. All string-like fields use
`dictionary(int8|int16|int64, string)` for interning.

#### OrderInitialized

| Column | Arrow Type |
|---|---|
| `trader_id` | `dictionary(int16, string)` |
| `strategy_id` | `dictionary(int16, string)` |
| `instrument_id` | `dictionary(int64, string)` |
| `client_order_id` | `string` |
| `order_side` | `dictionary(int8, string)` |
| `order_type` | `dictionary(int8, string)` |
| `quantity` | `string` |
| `time_in_force` | `dictionary(int8, string)` |
| `post_only` | `bool` |
| `reduce_only` | `bool` |
| `price` | `string` |
| `trigger_price` | `string` |
| `trigger_type` | `dictionary(int8, string)` |
| `limit_offset` | `string` |
| `trailing_offset` | `string` |
| `trailing_offset_type` | `dictionary(int8, string)` |
| `expire_time_ns` | `uint64` |
| `display_qty` | `string` |
| `quote_quantity` | `bool` |
| `options` | `binary` |
| `emulation_trigger` | `string` |
| `trigger_instrument_id` | `string` |
| `contingency_type` | `string` |
| `order_list_id` | `string` |
| `linked_order_ids` | `binary` |
| `parent_order_id` | `string` |
| `exec_algorithm_id` | `string` |
| `exec_algorithm_params` | `binary` |
| `exec_spawn_id` | `string` |
| `tags` | `binary` |
| `event_id` | `string` |
| `ts_init` | `uint64` |
| `reconciliation` | `bool` |

#### OrderFilled

| Column | Arrow Type |
|---|---|
| `trader_id` | `dictionary(int16, string)` |
| `strategy_id` | `dictionary(int16, string)` |
| `account_id` | `dictionary(int16, string)` |
| `instrument_id` | `dictionary(int64, string)` |
| `client_order_id` | `string` |
| `venue_order_id` | `string` |
| `trade_id` | `string` |
| `position_id` | `string` |
| `order_side` | `dictionary(int8, string)` |
| `order_type` | `dictionary(int8, string)` |
| `last_qty` | `string` |
| `last_px` | `string` |
| `currency` | `string` |
| `commission` | `string` |
| `liquidity_side` | `string` |
| `event_id` | `string` |
| `ts_event` | `uint64` |
| `ts_init` | `uint64` |
| `info` | `binary` |
| `reconciliation` | `bool` |

#### OrderDenied

| Column | Arrow Type |
|---|---|
| `trader_id` | `dictionary(int16, string)` |
| `strategy_id` | `dictionary(int16, string)` |
| `instrument_id` | `dictionary(int64, string)` |
| `client_order_id` | `string` |
| `reason` | `dictionary(int16, string)` |
| `event_id` | `string` |
| `ts_init` | `uint64` |

#### OrderSubmitted / OrderAccepted / OrderRejected / OrderCanceled / OrderExpired / OrderTriggered / OrderPendingCancel / OrderPendingUpdate

These share a common schema structure:

| Column | Arrow Type |
|---|---|
| `trader_id` | `dictionary(int16, string)` |
| `strategy_id` | `dictionary(int16, string)` |
| `account_id` | `dictionary(int16, string)` |
| `instrument_id` | `dictionary(int64, string)` |
| `client_order_id` | `string` |
| `venue_order_id` | `string` |
| `event_id` | `string` |
| `ts_event` | `uint64` |
| `ts_init` | `uint64` |
| `reconciliation` | `bool` |

Notes:
- `OrderSubmitted` and `OrderEmulated` omit `account_id`
- `OrderRejected` adds a `reason` field (`dictionary(int16, string)`)
- `OrderCancelRejected` and `OrderModifyRejected` add a `reason` field

#### OrderUpdated

| Column | Arrow Type |
|---|---|
| `trader_id` | `dictionary(int16, string)` |
| `strategy_id` | `dictionary(int16, string)` |
| `account_id` | `dictionary(int16, string)` |
| `instrument_id` | `dictionary(int64, string)` |
| `client_order_id` | `string` |
| `venue_order_id` | `string` |
| `price` | `string` |
| `quantity` | `string` |
| `trigger_price` | `string` |
| `event_id` | `string` |
| `ts_event` | `uint64` |
| `ts_init` | `uint64` |
| `reconciliation` | `bool` |

#### OrderReleased

| Column | Arrow Type |
|---|---|
| `trader_id` | `dictionary(int16, string)` |
| `strategy_id` | `dictionary(int16, string)` |
| `instrument_id` | `dictionary(int64, string)` |
| `client_order_id` | `string` |
| `released_price` | `string` |
| `event_id` | `string` |
| `ts_event` | `uint64` |
| `ts_init` | `uint64` |

### Key Differences: Rust vs Python Schemas

| Aspect | Rust (Parquet catalog) | Python (Feather events) |
|---|---|---|
| Price/Qty encoding | `FixedSizeBinary` (fixed-point integers) | `string` (human-readable decimals) |
| Enum encoding | `UInt8` integer codes | `dictionary(int8, string)` with labels |
| String interning | Plain `Utf8` | `dictionary(intN, string)` |
| Timestamps | `UInt64` nanoseconds | `uint64` nanoseconds |

---

## 4. Backtest Results & Strategy Performance

### Streaming Output (Feather)

During a backtest run, *all* data flows through the `StreamingFeatherWriter`
to Feather files:

```
{catalog_path}/backtest/{instance_id}/
├── quote_tick/{instrument_id}/{instrument_id}_{ns}.feather
├── trade_tick/{instrument_id}/{instrument_id}_{ns}.feather
├── bar/{bar_type}/{bar_type}_{ns}.feather
├── order_initialized_{ns}.feather
├── order_filled_{ns}.feather
├── position_opened_{ns}.feather
├── position_changed_{ns}.feather
├── position_closed_{ns}.feather
├── account_state_{ns}.feather
└── ...
```

These files use Arrow IPC streaming format and can be read with any Arrow
library (PyArrow, arrow-rs, Arrow.js, Apache Arrow Go, etc.).

### Writer Configuration

```python
class StreamingConfig:
    catalog_path: str                # Root output directory
    flush_interval_ms: int = 1000    # Auto-flush interval (ms)
    rotation_mode: RotationMode      # NO_ROTATION, SIZE, INTERVAL, SCHEDULED_DATES
    max_file_size: int = 1_073_741_824  # 1 GB threshold for SIZE rotation
    replace_existing: bool = False
    include_types: list[type] | None # Filter which event types to write
```

### BacktestResult Dataclass

After a backtest completes, a `BacktestResult` captures the summary:

```python
@dataclass
class BacktestResult:
    trader_id: str                # Trader identifier
    machine_id: str               # Machine that ran the backtest
    run_config_id: str | None     # Configuration ID (if applicable)
    instance_id: str              # Unique run instance UUID
    run_id: str                   # Run identifier
    run_started: int | None       # Unix nanoseconds
    run_finished: int | None      # Unix nanoseconds
    backtest_start: int | None    # Simulated start time (Unix nanos)
    backtest_end: int | None      # Simulated end time (Unix nanos)
    elapsed_time: float           # Wall-clock duration in seconds
    iterations: int               # Number of iterations
    total_events: int             # Total events processed
    total_orders: int             # Total orders created
    total_positions: int          # Total positions opened
    stats_pnls: dict[str, dict[str, float]]   # PnL stats by currency
    stats_returns: dict[str, float]            # Return statistics
```

### Conversion Path: Feather → Parquet

After a backtest, streaming Feather files can be converted to the Parquet
catalog format for long-term storage and efficient querying:

```
Feather files (streaming, sequential writes)
    ↓  convert_stream_to_data() / _convert_feather_table_to_parquet()
Parquet files (columnar, query-optimized, disjoint intervals)
```

During conversion:
- `ts_init` values are verified to be monotonically non-decreasing
- Internal bar types are converted to external format
- Disjoint interval constraints are enforced
- Row groups are written with configurable size (`max_rows_per_group`,
  default 5000)

### Reading Backtest Output

To read backtest output from another language:

1. List directories under `{catalog_path}/backtest/` — each is an
   `instance_id`
2. For market data: navigate into per-instrument subdirectories and read
   `.feather` files as Arrow IPC streams
3. For order/position events: read flat `.feather` files at the run root
4. Schema metadata on each file tells you the data type, instrument ID,
   and precision values
5. Use the [Python Arrow schemas](#order-event-arrow-schemas-python-layer)
   from Section 3 to interpret order event columns

### Reading from Live Runs

Identical structure, under `{catalog_path}/live/{instance_id}/`.

---

## 5. Cross-Cutting Gotchas

### Fixed-Point Binary Decoding

The single most important detail: **price and quantity columns in Parquet
files are NOT floats**. They are fixed-point integers stored as
`FixedSizeBinary`. You must:

1. Read N bytes (8 or 16)
2. Interpret as little-endian signed (prices) or unsigned (quantities) integer
3. Divide by `10^9` (standard) or `10^16` (high-precision)
4. Check for sentinel values before dividing

### High-Precision vs Standard-Precision Detection

Read the `FixedSizeBinary` width from the Arrow/Parquet schema:

| Width | Mode | Scalar | Price Type | Quantity Type |
|---|---|---|---|---|
| 16 bytes | High-precision | `10^16` | `i128` (signed) | `u128` (unsigned) |
| 8 bytes | Standard-precision | `10^9` | `i64` (signed) | `u64` (unsigned) |

If you read a file created with a different precision mode than expected,
NautilusTrader itself will error with a `PrecisionMismatch` message. Your
external reader should handle both modes gracefully.

### Legacy Data Correction

Catalog data created before December 16, 2025 may contain floating-point
rounding errors in raw values. NautilusTrader corrects these on read using a
round-to-nearest-valid-multiple algorithm. External readers processing legacy
data should apply the same correction:

```
scale = 10^(FIXED_PRECISION - price_precision)
remainder = raw % scale
if remainder == 0: return raw        # Already valid
half_scale = scale / 2
if remainder >= half_scale:
    return raw + (scale - remainder)  # Round up
else:
    return raw - remainder            # Round down
```

### Instrument ID URI-Safe Encoding

Forward slashes are removed when creating directory names:

```
BTC/USDT.BINANCE → BTCUSDT.BINANCE
EUR/USD.SIM      → EURUSD.SIM
```

This is a one-way transform for filesystem paths. The original ID with
slashes is preserved in schema metadata (`instrument_id` key).

### Byte Order

**Little-endian throughout.** All `FixedSizeBinary` price/quantity columns,
all integer fields. This matches x86/ARM conventions.

### Timestamp Precision

All timestamps (`ts_event`, `ts_init`) are **nanoseconds since Unix epoch**
stored as `UInt64`. Maximum representable time: ~2554 AD.

In Redis, timestamps may be stored as RFC 3339 strings (if
`timestamps_as_iso8601=true`) to avoid precision loss in some serialization
formats. Default is nanosecond integers.

### Feather File Format

Feather files are **Arrow IPC streaming format** (not the legacy Feather v1).
Read them with any Arrow IPC stream reader. The file extension is `.feather`
but the bytes are standard Arrow IPC.

### Parquet Row Group Size

Default `max_rows_per_group = 5000`. This affects predicate pushdown
granularity if you are using column statistics for filtering.

---

## Source File Reference

| Topic | File |
|---|---|
| Fixed-point constants | `crates/model/src/types/fixed.rs` |
| Arrow serialization (Rust) | `crates/serialization/src/arrow/*.rs` |
| Arrow schemas (Python) | `nautilus_trader/serialization/arrow/schema.py` |
| Redis cache database | `crates/infrastructure/src/redis/cache.rs` |
| Redis queries/serialization | `crates/infrastructure/src/redis/queries.rs` |
| Redis message bus streams | `crates/infrastructure/src/redis/msgbus.rs` |
| Parquet catalog | `nautilus_trader/persistence/catalog/parquet.py` |
| Streaming Feather writer | `nautilus_trader/persistence/writer.py` |
| Filename/URI helpers | `nautilus_trader/persistence/funcs.py` |
| BacktestResult | `nautilus_trader/backtest/results.py` |
| Enum definitions | `crates/model/src/enums.rs` |
