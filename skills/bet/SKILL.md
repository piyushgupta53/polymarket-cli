---
name: bet
description: Place a bet on a Polymarket prediction market. Looks up the market, shows the price, confirms, then submits the order.
argument-hint: "<amount> <yes|no> on <market query>"
---

# Bet Skill

Place a bet on a Polymarket prediction market. Walks through lookup, price check, confirmation, and order submission.

## Instructions

1. Parse `$ARGUMENTS` to extract:
   - **amount** (dollar size of the bet)
   - **side** (yes/no → maps to BUY on Yes token or BUY on No token)
   - **market query** (the rest of the text)

   Example: `10 yes on bitcoin above 100k` → amount=10, side=yes, query="bitcoin above 100k"

2. Search for the market:
   ```bash
   polymarket markets search "<query>" -o json -q
   ```
3. If multiple markets match, present them and ask the user to pick one.

4. Get the current price for the chosen outcome token:
   ```bash
   polymarket clob price --token <TOKEN_ID> --side buy -o json -q
   ```

5. Show the user a confirmation summary:
   - Market question
   - Outcome (Yes/No)
   - Current price
   - Size (number of shares = amount / price)
   - Total cost

6. **Ask the user to confirm** before placing the order. Never place an order without explicit confirmation.

7. On confirmation, place the order:
   ```bash
   polymarket clob create-order --token <TOKEN_ID> --side BUY --price <PRICE> --size <SIZE> -o json -q
   ```

8. Report the result (order ID, status, or error).

## Important

- **Always confirm before placing an order.** This involves real money.
- If the wallet is not configured, tell the user to run `polymarket setup` first.
- Use `--side BUY` for both Yes and No bets — the token ID determines the outcome.
- The price must be between 0.01 and 0.99.
