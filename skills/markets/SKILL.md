---
name: markets
description: Browse and search active Polymarket prediction markets.
argument-hint: "[search query]"
---

# Markets Skill

Browse and search active Polymarket prediction markets.

## Instructions

1. If `$ARGUMENTS` is provided, search for markets matching the query:
   ```bash
   polymarket markets search "$ARGUMENTS" -o json -q
   ```

2. If no arguments, list active markets:
   ```bash
   polymarket markets list --active --limit 15 -o json -q
   ```

3. Parse the JSON results and present a clean summary for each market:
   - Market question
   - Current prices (Yes/No percentages)
   - Volume and liquidity if available
   - Whether the market is active or closed

4. If the user asks about a specific market from the results, fetch full details:
   ```bash
   polymarket markets get <slug-or-id> -o json -q
   ```

## Formatting

- Show prices as percentages (e.g., 0.72 → 72%)
- Sort by relevance for search, or by volume for browse
- Keep the output scannable — one line per market with key stats
