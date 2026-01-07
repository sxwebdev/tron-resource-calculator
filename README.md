# TRON Resource Calculator

A CLI tool for real-time monitoring of TRON account Energy and Bandwidth resources with detailed analysis of regeneration and consumption rates.

## Features

- Real-time monitoring of Energy and Bandwidth
- Separate tracking of regeneration vs consumption rates
- Block tick detection and analysis
- Transaction capacity estimation
- JSON export for further analysis
- Compare multiple monitoring sessions
- Transaction simulation mode
- Graceful shutdown with Ctrl+C (saves collected data)

## Installation

### Using `go install`

```bash
go install github.com/sxwebdev/tron-resource-calculator/cmd/tron-resource-calculator@latest
```

### Build from source

```bash
git clone https://github.com/sxwebdev/tron-resource-calculator.git
cd tron-resource-calculator
go build -o tron-resource-calculator ./cmd/tron-resource-calculator
```

## Usage

### Basic Usage

```bash
# Monitor for 20 seconds (default)
tron-resource-calculator -a <TRON_ADDRESS>

# Monitor for 60 seconds
tron-resource-calculator -a <TRON_ADDRESS> -d 60

# Monitor with custom interval (3 seconds)
tron-resource-calculator -a <TRON_ADDRESS> -d 120 -i 3000
```

### CLI Flags

| Flag             | Short | Description                                 | Default                   |
| ---------------- | ----- | ------------------------------------------- | ------------------------- |
| `--address`      | `-a`  | TRON wallet address (required)              | -                         |
| `--node`         | `-n`  | TRON node URL                               | `https://api.trongrid.io` |
| `--duration`     | `-d`  | Monitoring duration in seconds              | `20`                      |
| `--interval`     | `-i`  | Sampling interval in milliseconds           | `1000`                    |
| `--until-full`   | -     | Monitor until resources are fully recovered | `false`                   |
| `--max-duration` | -     | Max duration for `--until-full` mode        | `86400`                   |
| `--compare`      | -     | Compare with previous JSON log file         | -                         |
| `--simulate`     | -     | Run transaction simulation                  | `false`                   |
| `--tx-cost`      | -     | Energy cost per transaction                 | `65000`                   |
| `--target-tx`    | -     | Target transactions per day                 | `800`                     |

### Examples

```bash
# Basic monitoring
tron-resource-calculator -a TYourAddressHere

# Extended monitoring with 3-second intervals
tron-resource-calculator -a TYourAddressHere --duration 3600 --interval 3000

# Monitor until resources fully recover
tron-resource-calculator -a TYourAddressHere --until-full --max-duration 86400

# Run with transaction simulation
tron-resource-calculator -a TYourAddressHere --simulate --tx-cost 65000 --target-tx 800

# Compare with previous run
tron-resource-calculator -a TYourAddressHere --compare ./previous_log.json
```

## Output

### Console Output

```text
TRON Resource Calculator
Address: TYourAddressHere
Node: https://api.trongrid.io
Duration: 20 seconds (interval: 1000ms)
Started: 2024-01-15 14:30:00 UTC
====================================================================================================

[T+000.3s] Energy: 508,257,857 / 1,302,229,514 (avail: 508,257,857) | BW: 12,763,677 / 16,191,537 (avail: 12,763,677)
[T+001.3s] Energy: 508,309,794 / 1,302,229,514 (avail: 508,309,794) | BW: 12,763,902 / 16,191,537 (avail: 12,763,902) | ΔE: +51,937 | ΔBW: +225
...

====================================================================================================
SUMMARY (20.5 seconds):

  Energy Rates:
    Regeneration: 15,073 /sec  (1,302,307,200 /day)
    Consumption:  26,211 /sec  (2,264,630,400 /day)
    Net:          -11,138 /sec  (-962,323,200 /day)

  Bandwidth Rates:
    Regeneration: 187 /sec  (16,156,800 /day)
    Consumption:  240 /sec  (20,736,000 /day)
    Net:          -53 /sec  (-4,579,200 /day)

  Resource Totals:
    Energy:    regenerated 1,506,101, consumed 2,866,343, net -1,360,242
    Bandwidth: regenerated 6,518, consumed 12,942, net -6,424

  Block Tick Analysis:
    Recovery ticks: 27 (avg interval: 4.6 sec, ~18865/day)
    Avg energy/tick: 55,799, bandwidth/tick: 241.3
    Consumption events: 14 (total: 2,866,343 energy, 12,942 bandwidth)

  Transaction Capacity (based on regen rate):
    Immediate (from buffer):
      At 65k Energy/tx:  7819 tx
      At 131k Energy/tx: 3879 tx
    Sustained (regen only):
      At 65k Energy/tx:  20035 tx/day
      At 131k Energy/tx: 9941 tx/day

Log saved to: tron_monitor_TYou...Here_20240115_143000.json
```

### JSON Output

The tool saves detailed JSON logs with all snapshots and analysis:

```json
{
  "metadata": {
    "address": "TYourAddressHere",
    "node": "https://api.trongrid.io",
    "start_time": "2024-01-15T14:30:00Z",
    "end_time": "2024-01-15T14:30:20Z",
    "duration_seconds": 20,
    "samples_count": 21,
    "interval_ms": 1000
  },
  "snapshots": [...],
  "analysis": {
    "energy_regenerated": 1506101,
    "energy_consumed": 2866343,
    "energy_regen_rate_per_second": 15073.5,
    "energy_consume_rate_per_second": 26211.2,
    "tick_analysis": {...},
    "practical_estimates": {...}
  }
}
```

## Understanding the Analysis

### Regeneration vs Consumption

- **Regeneration Rate**: How fast resources are being restored (positive deltas)
- **Consumption Rate**: How fast resources are being used by transactions (negative deltas)
- **Net Rate**: The difference (regen - consume), can be negative for active accounts

### Block Ticks

TRON resources regenerate in discrete "ticks" tied to block production (~3 seconds). The tool detects and analyzes these ticks to provide accurate regeneration metrics.

### Transaction Capacity

Based on measured regeneration rates, the tool estimates how many transactions per day are possible:

- **Immediate**: Using currently available energy buffer
- **Sustained**: Based only on regeneration rate (for continuous operation)
- **With Buffer**: Combining immediate capacity + daily regeneration

## API Reference

The tool uses the TRON HTTP API endpoint:

```text
POST /wallet/getaccountresource
Content-Type: application/json
Body: {"address": "<ADDRESS>", "visible": true}
```

## License

MIT
