# NautilusTrader Parquet Data Catalog Specification

## Overview

NautilusTrader stores market data in Apache Parquet format organized in a hierarchical directory structure. This document describes how to read this data in any programming language.

## Directory Structure

```shell
{catalog_root}/
└── data/
    └── {data_type}/
        └── {identifier}/
            └── {start_timestamp}_{end_timestamp}.parquet
```

### Data Type Directory Names

| Python Class      | Directory Name       |
|-------------------|----------------------|
| `QuoteTick`       | `quote_tick`         |
| `TradeTick`       | `trade_tick`         |
| `Bar`             | `bar`                |
| `OrderBookDelta`  | `order_book_deltas`  |
| `OrderBookDepth10`| `order_book_depths`  |
| `MarkPriceUpdate` | `mark_price_update`  |
| Custom data       | `custom_{class_name}`|

### Identifier Format

The `{identifier}` subdirectory is the instrument ID or bar type with `/` characters removed:
- `BTCUSDT.BINANCE` → `BTCUSDT.BINANCE`
- `AAPL.XNAS/USD` → `AAPL.XNASUSD` (forward slash removed)
- For bars: `BTCUSDT.BINANCE-1-MINUTE-LAST-EXTERNAL` → `BTCUSDT.BINANCE-1-MINUTE-LAST-EXTERNAL`

### Filename Format

Files are named with ISO 8601 timestamps (colons/periods replaced with hyphens):

```shell
{start_ts}_{end_ts}.parquet
```

Example: `2023-10-26T07-30-50-123456789Z_2023-10-26T08-30-50-987654321Z.parquet`

The timestamps are nanosecond-precision Unix timestamps converted to ISO 8601 format.

---

## Arrow/Parquet Schemas

### Precision Modes

NautilusTrader supports two precision modes that affect numeric field storage:

| Mode               | Integer Backing | Binary Size | Max Decimals | Scalar        |
|--------------------|-----------------|-------------|--------------|---------------|
| Standard (default) | 64-bit          | 8 bytes     | 9            | 10^9          |
| High-precision     | 128-bit         | 16 bytes    | 16           | 10^16         |

**How to detect:** Check the `FixedSizeBinary` field size in the schema - 8 bytes = standard, 16 bytes = high-precision.

### Fixed-Point Number Encoding

Prices and quantities are stored as **fixed-point integers** in **little-endian** byte order:

```
actual_value = raw_integer / FIXED_SCALAR
```

Where:
- Standard mode: `FIXED_SCALAR = 1,000,000,000` (10^9)
- High-precision: `FIXED_SCALAR = 10,000,000,000,000,000` (10^16)

**Example (standard precision):**
```python
# To decode a price of 123.456789:
raw_bytes = b'\x00\x80\xc7\x67\x00\x00\x00\x00'  # 8 bytes, little-endian
raw_int = int.from_bytes(raw_bytes, 'little', signed=True)  # for prices (signed)
actual_price = raw_int / 1_000_000_000  # = 123.456789
```

### Schema Metadata

Parquet file metadata contains essential decoding info:

| Key                | Description                    |
|--------------------|--------------------------------|
| `instrument_id`    | Full instrument identifier     |
| `price_precision`  | Decimal places for prices      |
| `size_precision`   | Decimal places for quantities  |
| `bar_type`         | (Bars only) Full bar type spec |

---

## Data Type Schemas

### QuoteTick

| Column      | Arrow Type                     | Description                        |
|-------------|--------------------------------|------------------------------------|
| `bid_price` | `FixedSizeBinary(8)` or `(16)` | Best bid price (signed int, LE)    |
| `ask_price` | `FixedSizeBinary(8)` or `(16)` | Best ask price (signed int, LE)    |
| `bid_size`  | `FixedSizeBinary(8)` or `(16)` | Bid quantity (unsigned int, LE)    |
| `ask_size`  | `FixedSizeBinary(8)` or `(16)` | Ask quantity (unsigned int, LE)    |
| `ts_event`  | `UInt64`                       | Event timestamp (nanoseconds)      |
| `ts_init`   | `UInt64`                       | Initialization timestamp (nanos)   |

### TradeTick

| Column           | Arrow Type                     | Description                        |
|------------------|--------------------------------|------------------------------------|
| `price`          | `FixedSizeBinary(8)` or `(16)` | Trade price (signed int, LE)       |
| `size`           | `FixedSizeBinary(8)` or `(16)` | Trade quantity (unsigned int, LE)  |
| `aggressor_side` | `UInt8`                        | 0=NO_AGGRESSOR, 1=BUYER, 2=SELLER  |
| `trade_id`       | `Utf8` (String)                | Exchange trade ID                  |
| `ts_event`       | `UInt64`                       | Event timestamp (nanoseconds)      |
| `ts_init`        | `UInt64`                       | Initialization timestamp (nanos)   |

### Bar (OHLCV)

| Column    | Arrow Type                     | Description                        |
|-----------|--------------------------------|------------------------------------|
| `open`    | `FixedSizeBinary(8)` or `(16)` | Open price (signed int, LE)        |
| `high`    | `FixedSizeBinary(8)` or `(16)` | High price (signed int, LE)        |
| `low`     | `FixedSizeBinary(8)` or `(16)` | Low price (signed int, LE)         |
| `close`   | `FixedSizeBinary(8)` or `(16)` | Close price (signed int, LE)       |
| `volume`  | `FixedSizeBinary(8)` or `(16)` | Volume (unsigned int, LE)          |
| `ts_event`| `UInt64`                       | Bar close timestamp (nanos)        |
| `ts_init` | `UInt64`                       | Initialization timestamp (nanos)   |

**Metadata:** `bar_type` contains the full bar specification (e.g., `BTCUSDT.BINANCE-1-MINUTE-LAST-EXTERNAL`)

### OrderBookDelta

| Column     | Arrow Type                     | Description                        |
|------------|--------------------------------|------------------------------------|
| `action`   | `UInt8`                        | 1=ADD, 2=UPDATE, 3=DELETE, 4=CLEAR |
| `side`     | `UInt8`                        | 1=BID, 2=ASK, 0=NO_SIDE            |
| `price`    | `FixedSizeBinary(8)` or `(16)` | Price level (signed int, LE)       |
| `size`     | `FixedSizeBinary(8)` or `(16)` | Size at level (unsigned int, LE)   |
| `order_id` | `UInt64`                       | Order ID (0 if not provided)       |
| `flags`    | `UInt8`                        | Flags (F_LAST=128 marks batch end) |
| `sequence` | `UInt64`                       | Sequence number                    |
| `ts_event` | `UInt64`                       | Event timestamp (nanos)            |
| `ts_init`  | `UInt64`                       | Initialization timestamp (nanos)   |

### OrderBookDepth10

| Column           | Arrow Type                     | Description                   |
|------------------|--------------------------------|-------------------------------|
| `bid_price_0..9` | `FixedSizeBinary(8)` or `(16)` | Top 10 bid prices (signed)    |
| `ask_price_0..9` | `FixedSizeBinary(8)` or `(16)` | Top 10 ask prices (signed)    |
| `bid_size_0..9`  | `FixedSizeBinary(8)` or `(16)` | Top 10 bid sizes (unsigned)   |
| `ask_size_0..9`  | `FixedSizeBinary(8)` or `(16)` | Top 10 ask sizes (unsigned)   |
| `bid_count_0..9` | `UInt32`                       | Order count at each bid level |
| `ask_count_0..9` | `UInt32`                       | Order count at each ask level |
| `flags`          | `UInt8`                        | Flags (F_LAST=128)            |
| `sequence`       | `UInt64`                       | Sequence number               |
| `ts_event`       | `UInt64`                       | Event timestamp (nanos)       |
| `ts_init`        | `UInt64`                       | Initialization timestamp      |

---

## Reading Example (Python without Nautilus)

```python
import pyarrow.parquet as pq
import struct

def decode_fixed_point(binary_value: bytes, signed: bool = True) -> float:
    """Decode a NautilusTrader fixed-point binary value."""
    byte_size = len(binary_value)
    
    if byte_size == 8:
        scalar = 1_000_000_000  # 10^9
        fmt = '<q' if signed else '<Q'  # little-endian int64/uint64
    elif byte_size == 16:
        scalar = 10_000_000_000_000_000  # 10^16
        raw = int.from_bytes(binary_value, 'little', signed=signed)
        return raw / scalar
    else:
        raise ValueError(f"Unexpected binary size: {byte_size}")
    
    raw = struct.unpack(fmt, binary_value)[0]
    return raw / scalar

# Read a QuoteTick file
table = pq.read_table("catalog/data/quote_tick/BTCUSDT.BINANCE/2024-01-01T00-00-00-000000000Z_2024-01-02T00-00-00-000000000Z.parquet")

# Get metadata
metadata = table.schema.metadata
instrument_id = metadata[b'instrument_id'].decode()
price_precision = int(metadata[b'price_precision'].decode())

# Decode first row
row = table.to_pydict()
bid_price = decode_fixed_point(row['bid_price'][0], signed=True)
ask_price = decode_fixed_point(row['ask_price'][0], signed=True)
bid_size = decode_fixed_point(row['bid_size'][0], signed=False)
ask_size = decode_fixed_point(row['ask_size'][0], signed=False)
ts_event = row['ts_event'][0]  # nanoseconds since Unix epoch

print(f"Instrument: {instrument_id}")
print(f"Bid: {bid_price:.{price_precision}f} x {bid_size}")
print(f"Ask: {ask_price:.{price_precision}f} x {ask_size}")
```

---

## Reading Example (Rust)

```rust
use arrow::array::{FixedSizeBinaryArray, UInt64Array};
use parquet::arrow::arrow_reader::ParquetRecordBatchReaderBuilder;
use std::fs::File;

fn decode_price_i64(bytes: &[u8]) -> f64 {
    let raw = i64::from_le_bytes(bytes.try_into().unwrap());
    raw as f64 / 1_000_000_000.0
}

fn decode_quantity_u64(bytes: &[u8]) -> f64 {
    let raw = u64::from_le_bytes(bytes.try_into().unwrap());
    raw as f64 / 1_000_000_000.0
}

fn main() -> Result<(), Box<dyn std::error::Error>> {
    let file = File::open("catalog/data/quote_tick/BTCUSDT.BINANCE/file.parquet")?;
    let builder = ParquetRecordBatchReaderBuilder::try_new(file)?;
    
    // Get metadata
    let metadata = builder.schema().metadata();
    let instrument_id = metadata.get("instrument_id").unwrap();
    
    let reader = builder.build()?;
    for batch in reader {
        let batch = batch?;
        let bid_price = batch.column(0).as_any()
            .downcast_ref::<FixedSizeBinaryArray>().unwrap();
        
        for i in 0..batch.num_rows() {
            let price = decode_price_i64(bid_price.value(i));
            println!("Bid price: {}", price);
        }
    }
    Ok(())
}
```

---

## Key Points to Remember

1. **Byte order:** All fixed-point values use **little-endian** encoding
2. **Signed vs unsigned:** Prices are **signed** (can be negative), quantities are **unsigned**
3. **Precision:** Check `FixedSizeBinary` size (8 or 16 bytes) to determine the scalar
4. **Metadata is essential:** Always read schema metadata for `instrument_id` and precision values
5. **Timestamps:** All timestamps are **nanoseconds since Unix epoch** as `UInt64`
6. **File filtering:** Parse filename timestamps to filter files by time range without opening them