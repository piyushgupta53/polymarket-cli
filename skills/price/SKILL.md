---
name: price
description: Look up the current price for a Polymarket prediction market.
argument-hint: "<market slug or search query>"
---

# Price Skill

Look up the current price for a Polymarket prediction market.

## Instructions

1. The user provides a market slug, name, or search query via `$ARGUMENTS`.
2. Search for the market to find its token IDs:
   ```bash
   polymarket markets search "$ARGUMENTS" -o json -q
   ```
3. Parse the JSON output. Each market has `tokens` with `token_id` and `outcome` fields.
4. For each outcome token, get the current price:
   ```bash
   polymarket clob price --token <TOKEN_ID> --side buy -o json -q
   ```
5. Present results to the user in a clear format:
   - Market question
   - Each outcome with its current price (as a percentage, e.g. 0.65 = 65%)
   - If there are exactly 2 outcomes (Yes/No), show both

## Error Handling

- If no markets match the search, tell the user and suggest refining their query.
- If the CLOB price endpoint fails, fall back to the `outcomePrices` field from the Gamma API search results.

## Example Output

```
Bitcoin above $100k by June 2025?
  Yes: 72% ($0.72)
  No:  28% ($0.28)
```
