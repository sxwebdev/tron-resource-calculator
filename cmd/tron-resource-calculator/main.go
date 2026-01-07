package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sxwebdev/tron-resource-calculator/internal/client"
	"github.com/sxwebdev/tron-resource-calculator/internal/models"
	"github.com/sxwebdev/tron-resource-calculator/internal/monitor"
	"github.com/sxwebdev/tron-resource-calculator/internal/output"
)

const (
	defaultNode        = "https://api.trongrid.io"
	defaultDuration    = 20
	defaultInterval    = 1000
	defaultMaxDuration = 86400
)

func main() {
	// Parse command line flags
	address := flag.String("address", "", "TRON wallet address (required)")
	addressShort := flag.String("a", "", "TRON wallet address (shorthand)")
	node := flag.String("node", defaultNode, "TRON node URL")
	nodeShort := flag.String("n", "", "TRON node URL (shorthand)")
	duration := flag.Int("duration", defaultDuration, "Monitoring duration in seconds")
	durationShort := flag.Int("d", 0, "Monitoring duration in seconds (shorthand)")

	// New flags
	interval := flag.Int("interval", defaultInterval, "Sampling interval in milliseconds")
	intervalShort := flag.Int("i", 0, "Sampling interval in ms (shorthand)")
	untilFull := flag.Bool("until-full", false, "Monitor until resources are fully recovered")
	maxDuration := flag.Int("max-duration", defaultMaxDuration, "Max duration when using --until-full (seconds)")
	compareFile := flag.String("compare", "", "Compare with previous log file (JSON)")

	// Simulation flags
	simulate := flag.Bool("simulate", false, "Run transaction simulation")
	txCost := flag.Int64("tx-cost", 65000, "Energy cost per transaction for simulation")
	targetTx := flag.Int("target-tx", 800, "Target transactions per day for simulation")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s --address <TRON_ADDRESS> [options]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Monitor TRON account Energy and Bandwidth resources in real-time.\n\n")
		fmt.Fprintf(os.Stderr, "Basic Flags:\n")
		fmt.Fprintf(os.Stderr, "  -a, --address      TRON wallet address (required, format: T...)\n")
		fmt.Fprintf(os.Stderr, "  -n, --node         TRON node URL (default: %s)\n", defaultNode)
		fmt.Fprintf(os.Stderr, "  -d, --duration     Monitoring duration in seconds (default: %d)\n", defaultDuration)
		fmt.Fprintf(os.Stderr, "  -i, --interval     Sampling interval in ms (default: %d)\n", defaultInterval)
		fmt.Fprintf(os.Stderr, "\nAdvanced Flags:\n")
		fmt.Fprintf(os.Stderr, "      --until-full   Monitor until resources are fully recovered\n")
		fmt.Fprintf(os.Stderr, "      --max-duration Max duration for --until-full (default: %d)\n", defaultMaxDuration)
		fmt.Fprintf(os.Stderr, "      --compare      Compare with previous log file\n")
		fmt.Fprintf(os.Stderr, "\nSimulation Flags:\n")
		fmt.Fprintf(os.Stderr, "      --simulate     Run transaction simulation after monitoring\n")
		fmt.Fprintf(os.Stderr, "      --tx-cost      Energy cost per transaction (default: 65000)\n")
		fmt.Fprintf(os.Stderr, "      --target-tx    Target transactions per day (default: 800)\n")
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  %s -a TXxx -d 60\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -a TXxx --duration 3600 --interval 3000\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -a TXxx --until-full --max-duration 86400\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -a TXxx --simulate --tx-cost 65000 --target-tx 800\n", os.Args[0])
	}

	flag.Parse()

	// Build config
	cfg := models.Config{
		Address:     *address,
		Node:        *node,
		Duration:    *duration,
		IntervalMs:  *interval,
		UntilFull:   *untilFull,
		MaxDuration: *maxDuration,
		CompareFile: *compareFile,
		Simulate:    *simulate,
		TxCost:      *txCost,
		TargetTx:    *targetTx,
	}

	// Handle shorthand flags
	if cfg.Address == "" {
		cfg.Address = *addressShort
	}
	if *nodeShort != "" {
		cfg.Node = *nodeShort
	}
	if *durationShort > 0 {
		cfg.Duration = *durationShort
	}
	if *intervalShort > 0 {
		cfg.IntervalMs = *intervalShort
	}

	// Validate address
	if cfg.Address == "" {
		fmt.Fprintln(os.Stderr, "Error: address is required")
		flag.Usage()
		os.Exit(1)
	}

	if err := client.ValidateAddress(cfg.Address); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	// Validate duration
	if cfg.Duration <= 0 {
		fmt.Fprintln(os.Stderr, "Error: duration must be positive")
		os.Exit(1)
	}

	// Validate interval
	if cfg.IntervalMs < 100 {
		fmt.Fprintln(os.Stderr, "Error: interval must be at least 100ms")
		os.Exit(1)
	}

	// Run the monitor
	if err := run(cfg); err != nil {
		output.PrintError(err)
		os.Exit(1)
	}
}

func run(cfg models.Config) error {
	// Setup context with cancellation for graceful shutdown
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle Ctrl+C
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Create client and monitor
	c := client.New(cfg.Node)
	m := monitor.NewWithInterval(c, cfg.Address, cfg.Duration, cfg.IntervalMs)

	startTime := time.Now()
	output.PrintHeader(cfg.Address, cfg.Node, cfg.Duration, cfg.IntervalMs, startTime)

	// Channel to collect snapshots
	var snapshots []models.ResourceSnapshot
	done := make(chan error, 1)

	go func() {
		var err error
		if cfg.UntilFull {
			snapshots, err = m.RunUntilFull(ctx, cfg.MaxDuration, func(snapshot models.ResourceSnapshot, index int) {
				output.PrintSnapshot(snapshot, index)
			})
		} else {
			snapshots, err = m.Run(ctx, func(snapshot models.ResourceSnapshot, index int) {
				output.PrintSnapshot(snapshot, index)
			})
		}
		done <- err
	}()

	// Wait for completion or interrupt
	var runErr error
	select {
	case <-sigChan:
		output.PrintInterrupted()
		cancel()
		select {
		case runErr = <-done:
		case <-time.After(2 * time.Second):
		}
	case runErr = <-done:
	}

	endTime := time.Now()

	// Even if interrupted, save what we have
	if len(snapshots) > 0 {
		analysis := monitor.Analyze(snapshots, cfg.Duration)

		// Build and save report - use actual duration from analysis
		actualDurationInt := int(analysis.ActualDurationSec)
		if actualDurationInt < 1 {
			actualDurationInt = 1
		}
		report := output.BuildReport(cfg.Address, cfg.Node, startTime, endTime, actualDurationInt, snapshots, analysis)
		report.Metadata.IntervalMs = cfg.IntervalMs

		filename, saveErr := output.SaveJSON(report)
		if saveErr != nil {
			fmt.Fprintf(os.Stderr, "\nWarning: failed to save JSON: %v\n", saveErr)
		} else {
			output.PrintSummary(analysis, filename)
		}

		// Run simulation if requested
		if cfg.Simulate && len(snapshots) > 0 {
			sim := monitor.Simulate(snapshots[len(snapshots)-1], analysis, cfg.TxCost, cfg.TargetTx)
			output.PrintSimulation(sim)
		}

		// Compare with previous file if requested
		if cfg.CompareFile != "" {
			if err := compareWithPrevious(cfg.CompareFile, analysis); err != nil {
				fmt.Fprintf(os.Stderr, "\nWarning: failed to compare: %v\n", err)
			}
		}
	}

	return runErr
}

func compareWithPrevious(filename string, current models.Analysis) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return fmt.Errorf("failed to read file: %w", err)
	}

	var prev models.MonitorReport
	if err := json.Unmarshal(data, &prev); err != nil {
		return fmt.Errorf("failed to parse JSON: %w", err)
	}

	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Comparison with: %s\n", filename)
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	prevAnalysis := prev.Analysis

	fmt.Printf("Energy Regen Rate:    %.1f -> %.1f /sec (delta: %+.1f)\n",
		prevAnalysis.EnergyRegenRatePerSec,
		current.EnergyRegenRatePerSec,
		current.EnergyRegenRatePerSec-prevAnalysis.EnergyRegenRatePerSec)

	fmt.Printf("Energy Consume Rate:  %.1f -> %.1f /sec (delta: %+.1f)\n",
		prevAnalysis.EnergyConsumeRatePerSec,
		current.EnergyConsumeRatePerSec,
		current.EnergyConsumeRatePerSec-prevAnalysis.EnergyConsumeRatePerSec)

	fmt.Printf("Bandwidth Regen Rate: %.1f -> %.1f /sec (delta: %+.1f)\n",
		prevAnalysis.BandwidthRegenRatePerSec,
		current.BandwidthRegenRatePerSec,
		current.BandwidthRegenRatePerSec-prevAnalysis.BandwidthRegenRatePerSec)

	fmt.Printf("Tx/day (65k):         %.0f -> %.0f (delta: %+.0f)\n",
		prevAnalysis.TxPerDay65k,
		current.TxPerDay65k,
		current.TxPerDay65k-prevAnalysis.TxPerDay65k)

	return nil
}
