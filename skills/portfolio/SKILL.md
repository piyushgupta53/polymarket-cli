---
name: portfolio
description: Show your Polymarket portfolio — positions, open orders, recent trades, and balance.
argument-hint: "[positions|orders|trades|balance]"
---

# Portfolio Skill

Show the user's Polymarket portfolio: positions, open orders, recent trades, and/or balance.

## Instructions

1. If `$ARGUMENTS` specifies a section (positions, orders, trades, balance), show only that section. Otherwise show all.

2. **Positions** — open market positions:
   ```bash
   polymarket data positions -o json -q
   ```

3. **Open Orders**:
   ```bash
   polymarket clob orders -o json -q
   ```

4. **Recent Trades**:
   ```bash
   polymarket data trades --limit 10 -o json -q
   ```

5. **Balance**:
   ```bash
   polymarket clob balance -o json -q
   ```

6. Format the results in a clear, readable summary for the user. Include:
   - Position: market name, outcome, size, avg price, current value, P&L if calculable
   - Orders: market, side, price, size, status
   - Trades: market, side, price, size, timestamp
   - Balance: available balance and allowance

## Error Handling

- If auth fails, tell the user to run `polymarket setup` first.
- If a section returns empty results, say so (e.g., "No open positions").
