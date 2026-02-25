# Polymarket CLI

A full-featured command-line interface for [Polymarket](https://polymarket.com) prediction markets. Browse markets, place trades, and manage your portfolio — all from the terminal.

## Installation

```bash
git clone https://github.com/piyushgupta/polymarket-cli.git
cd polymarket-cli
make build
# Binary: ./polymarket
```

To use `polymarket` without the `./` prefix, move it into your `$PATH`:

```bash
sudo mv ./polymarket /usr/local/bin/
```

## Quick Start

```bash
# Check API connectivity
polymarket status

# Browse active markets
polymarket markets list --active --limit 10

# Search for a market
polymarket markets search "bitcoin"

# Get market details
polymarket markets get will-bitcoin-hit-100k
```

## Setup

Configure your wallet to enable trading:

```bash
# Interactive setup wizard
polymarket setup

# Non-interactive: pass key directly
polymarket setup --key 0xYOUR_PRIVATE_KEY

# Non-interactive: read key from environment variable
polymarket setup --key-env MY_SECRET_KEY

# View current config (key redacted)
polymarket setup --show

# Reset wallet configuration
polymarket setup --reset
```

Configuration is stored in `~/.config/polymarket/config.json`.

## Command Reference

### Markets

| Command | Description |
|---------|-------------|
| `markets list` | List markets with filters (--active, --closed, --limit, --order) |
| `markets get <id-or-slug>` | Get detailed market information |
| `markets search <query>` | Search markets by keyword |

### Events & Tags

| Command | Description |
|---------|-------------|
| `events list` | List events |
| `events get <id>` | Get event details |
| `tags list` | List available tags |

### CLOB (Order Book)

| Command | Description | Auth |
|---------|-------------|------|
| `clob ok` | Check CLOB API health | |
| `clob price --token <ID>` | Get token price | |
| `clob book --token <ID>` | Get order book | |
| `clob midpoint --token <ID>` | Get midpoint price | |
| `clob spread --token <ID>` | Get bid-ask spread | |
| `clob price-history --token <ID>` | Get price history | |
| `clob create-order` | Place a limit order | Yes |
| `clob cancel --order-id <ID>` | Cancel an order | Yes |
| `clob orders` | List open orders | Yes |
| `clob trades` | List trade history | Yes |
| `clob balance` | Check balance and allowance | Yes |

### Data (Portfolio & Analytics)

| Command | Description |
|---------|-------------|
| `data positions` | List open positions |
| `data closed` | List closed positions |
| `data trades` | List trade history |
| `data value` | Portfolio value history |
| `data activity` | Account activity |
| `data leaderboard` | Trading leaderboard |

### On-Chain

| Command | Description | Auth |
|---------|-------------|------|
| `approve check` | Check token approvals | Yes |
| `approve set` | Set token approvals | Yes |
| `ctf split` | Split position tokens | Yes |
| `ctf merge` | Merge position tokens | Yes |
| `ctf redeem` | Redeem resolved positions | Yes |

### Utilities

| Command | Description |
|---------|-------------|
| `status` | CLI and API connectivity status |
| `version` | Print version information |
| `setup` | Configure wallet for trading |
| `commands list` | List all commands (agent discovery) |
| `commands describe <path>` | Describe a command with flags |
| `tui` | Launch interactive TUI |
| `shell` | Start interactive REPL |
| `upgrade` | Self-update to latest version |

## Agent Integration Guide

This CLI is designed for use by AI agents and automation scripts. Every command supports machine-readable output.

### Output Formats

```bash
# JSON output (auto-detected when stdout is piped)
polymarket markets list -o json

# TSV output (tab-separated, for unix pipelines)
polymarket markets list -o tsv

# Table output (default in TTY)
polymarket markets list -o table
```

### Quiet Mode

Suppress non-data output (logs, decorations). Produces compact JSON (no indentation):

```bash
polymarket markets list -o json -q
```

### Non-Interactive Mode

Disable all interactive prompts. Auto-detected when stdin is not a TTY:

```bash
polymarket setup --key-env SECRET_KEY -n -o json -q
```

### Structured Errors

Errors are returned as JSON to stderr with machine-readable codes:

```json
{"error":{"code":"AUTH_REQUIRED","message":"authentication required","hint":"Run: polymarket setup --key <hex>"}}
```

Error codes: `AUTH_REQUIRED`, `AUTH_FAILED`, `NETWORK_ERROR`, `NOT_FOUND`, `INVALID_INPUT`, `INSUFFICIENT_FUNDS`, `RATE_LIMITED`, `CHAIN_ERROR`, `INTERNAL`.

### Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error |
| 2 | Invalid input |
| 3 | Authentication error |
| 4 | Network error |
| 5 | Not found |
| 6 | Insufficient funds |
| 7 | Chain/on-chain error |

### Command Discovery

Agents can discover all available commands and their flags programmatically:

```bash
# List all commands
polymarket commands list -o json -q

# Describe a specific command (flags, types, defaults, required markers)
polymarket commands describe "clob create-order" -o json -q
```

### Table Options

```bash
# Omit table headers (for scripting)
polymarket markets list --no-headers

# Disable color output
polymarket markets list --no-color
```

### CI/CD Authentication

```bash
export POLYMARKET_PRIVATE_KEY=0x...
polymarket setup --key-env POLYMARKET_PRIVATE_KEY -n -o json -q
polymarket clob balance -o json -q
```

## Environment Variables

| Variable | Description |
|----------|-------------|
| `POLYMARKET_PRIVATE_KEY` | Ethereum private key (hex) |
| `POLYMARKET_NON_INTERACTIVE` | Set to `1` to disable interactive prompts |
| `POLYMARKET_API_URL` | Override Gamma API base URL |
| `POLYMARKET_DATA_API_URL` | Override Data API base URL |
| `POLYMARKET_BRIDGE_API_URL` | Override Bridge API base URL |
| `POLYMARKET_RPC_URL` | Override Polygon RPC URL |
| `NO_COLOR` | Disable colored output (any value) |

## TUI

Launch the interactive terminal UI:

```bash
polymarket tui
```

Features: market browser with fuzzy search, order book visualization, order form, portfolio dashboard with positions/orders/trades tabs.

## License

MIT
