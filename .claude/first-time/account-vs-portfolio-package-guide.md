## Portfolio vs Account: A First-Time Contributor's Guide

Think of it like a **bank** analogy:

| Module | Analogy | Responsibility |
|--------|---------|----------------|
| **Account** | Your individual bank account | Tracks balances, calculates interest/fees, records transactions |
| **Portfolio** | Your overall financial dashboard | Aggregates all your accounts, shows net worth, risk exposure |

---

### **Account Module** (`nautilus_trader/accounting/`)

Handles **low-level, single-account operations**:

- **Balance tracking**: total, free (available), locked (reserved for orders)
- **PnL calculation**: computes profit/loss when a trade fills
- **Margin/locked amounts**: how much capital is tied up
- **Account types**: `CashAccount` (no leverage) vs `MarginAccount` (leverage)

```
Account knows: "I have $10,000 total, $8,000 free, $2,000 locked for pending orders"
```

---

### **Portfolio Module** (`nautilus_trader/portfolio/`)

Handles **high-level, cross-account aggregation**:

- **Aggregate metrics**: total PnL across all accounts/venues
- **Net position tracking**: "Am I net long or short on BTC overall?"
- **Risk exposure**: combined exposure across accounts
- **Multi-venue support**: you might trade on Binance AND Kraken

```
Portfolio knows: "Across all my accounts, I'm net long 5 BTC, my total unrealized PnL is +$15,000"
```

---

### **Why Two Modules Instead of One?**

1. **Different scopes**: One account vs. the entire trading operation
2. **Multi-venue trading**: You can have accounts at Binance, Kraken, Interactive Brokers simultaneously - Portfolio aggregates them
3. **Separation of concerns**: Account does math, Portfolio does aggregation
4. **Testability**: Account logic can be tested without Portfolio complexity
5. **Performance**: Portfolio caches aggregated results instead of recalculating every query

---

### **How They Work Together**

```
Order fills on Binance
       ↓
Portfolio.update_order()
       ↓
AccountsManager.update_balances()  ← bridge layer
       ↓
CashAccount.calculate_pnls()       ← account does the math
       ↓
Portfolio caches the result        ← portfolio stores aggregates
```

---

### **Quick Mental Model**

- **Account** = "What happened in THIS account?"
- **Portfolio** = "What's my overall position across EVERYTHING?"

If you only traded on one venue with one account, they'd seem redundant. But NautilusTrader supports complex setups with multiple venues/accounts, making this separation essential.