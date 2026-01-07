package output

import (
	"fmt"
	"strings"
	"time"

	"github.com/sxwebdev/tron-resource-calculator/internal/models"
)

// PrintHeader prints the monitoring session header
func PrintHeader(address, node string, duration int, intervalMs int, startTime time.Time) {
	fmt.Println("TRON Resource Monitor")
	fmt.Printf("Address: %s\n", address)
	fmt.Printf("Node: %s\n", node)
	fmt.Printf("Duration: %d seconds (interval: %dms)\n", duration, intervalMs)
	fmt.Printf("Started: %s\n", startTime.UTC().Format("2006-01-02 15:04:05 UTC"))
	fmt.Println(strings.Repeat("=", 100))
	fmt.Println()
}

// PrintSnapshot prints a single snapshot line
func PrintSnapshot(snapshot models.ResourceSnapshot, index int) {
	// Use actual elapsed time from snapshot
	elapsedSec := float64(snapshot.ElapsedMs) / 1000.0

	if index == 0 {
		fmt.Printf("[T+%05.1fs] Energy: %s / %s (avail: %s) | BW: %s / %s (avail: %s)\n",
			elapsedSec,
			formatNumber(snapshot.EnergyAvailable),
			formatNumber(snapshot.EnergyLimit),
			formatNumber(snapshot.EnergyAvailable),
			formatNumber(snapshot.BandwidthAvailable),
			formatNumber(snapshot.TotalBandwidthLimit()),
			formatNumber(snapshot.BandwidthAvailable),
		)
	} else {
		fmt.Printf("[T+%05.1fs] Energy: %s / %s (avail: %s) | BW: %s / %s (avail: %s) | ΔE: %s | ΔBW: %s\n",
			elapsedSec,
			formatNumber(snapshot.EnergyAvailable),
			formatNumber(snapshot.EnergyLimit),
			formatNumber(snapshot.EnergyAvailable),
			formatNumber(snapshot.BandwidthAvailable),
			formatNumber(snapshot.TotalBandwidthLimit()),
			formatNumber(snapshot.BandwidthAvailable),
			formatDelta(snapshot.DeltaEnergy),
			formatDelta(snapshot.DeltaBandwidth),
		)
	}
}

// PrintSummary prints the analysis summary
func PrintSummary(analysis models.Analysis, filename string) {
	fmt.Println()
	fmt.Println(strings.Repeat("=", 100))
	fmt.Printf("SUMMARY (%.1f seconds):\n", analysis.ActualDurationSec)

	// Separated rates
	fmt.Println()
	fmt.Println("  Energy Rates:")
	fmt.Printf("    Regeneration: %s /sec  (%s /day)\n",
		formatFloat(analysis.EnergyRegenRatePerSec),
		formatNumber(int64(analysis.EnergyRegenRatePerDay)),
	)
	fmt.Printf("    Consumption:  %s /sec  (%s /day)\n",
		formatFloat(analysis.EnergyConsumeRatePerSec),
		formatNumber(int64(analysis.EnergyConsumeRatePerDay)),
	)
	fmt.Printf("    Net:          %s /sec  (%s /day)\n",
		formatFloat(analysis.EnergyNetRatePerSec),
		formatDelta(int64(analysis.EnergyNetRatePerDay)),
	)

	fmt.Println()
	fmt.Println("  Bandwidth Rates:")
	fmt.Printf("    Regeneration: %s /sec  (%s /day)\n",
		formatFloat(analysis.BandwidthRegenRatePerSec),
		formatNumber(int64(analysis.BandwidthRegenRatePerDay)),
	)
	fmt.Printf("    Consumption:  %s /sec  (%s /day)\n",
		formatFloat(analysis.BandwidthConsumeRatePerSec),
		formatNumber(int64(analysis.BandwidthConsumeRatePerDay)),
	)
	fmt.Printf("    Net:          %s /sec  (%s /day)\n",
		formatFloat(analysis.BandwidthNetRatePerSec),
		formatDelta(int64(analysis.BandwidthNetRatePerDay)),
	)

	// Resource totals
	fmt.Println()
	fmt.Println("  Resource Totals:")
	fmt.Printf("    Energy:    regenerated %s, consumed %s, net %s\n",
		formatNumber(analysis.EnergyRegenerated),
		formatNumber(analysis.EnergyConsumed),
		formatDelta(analysis.EnergyTotalDelta),
	)
	fmt.Printf("    Bandwidth: regenerated %s, consumed %s, net %s\n",
		formatNumber(analysis.BandwidthRegenerated),
		formatNumber(analysis.BandwidthConsumed),
		formatDelta(analysis.BandwidthTotalDelta),
	)

	// Tick analysis
	tick := analysis.TickAnalysis
	if tick.RecoveryTicks > 0 || tick.ConsumptionEvents > 0 {
		fmt.Println()
		fmt.Println("  Block Tick Analysis:")
		fmt.Printf("    Recovery ticks: %d (avg interval: %.1f sec, ~%.0f/day)\n",
			tick.RecoveryTicks, tick.AvgRecoveryInterval, tick.RecoveryTicksPerDay)
		fmt.Printf("    Avg energy/tick: %s, bandwidth/tick: %.1f\n",
			formatNumber(int64(tick.EnergyPerTick)), tick.BandwidthPerTick)

		if tick.ConsumptionEvents > 0 {
			fmt.Printf("    Consumption events: %d (total: %s energy, %s bandwidth)\n",
				tick.ConsumptionEvents,
				formatNumber(tick.TotalEnergyConsumed),
				formatNumber(tick.TotalBandwidthConsumed))
			fmt.Printf("    Avg per consumption: %s energy, %.0f bandwidth\n",
				formatNumber(int64(tick.AvgEnergyPerConsume)),
				tick.AvgBandwidthPerConsume)
		}
	}

	// Used-based analysis
	used := analysis.UsedBasedAnalysis
	if used.EnergyUsedAtStart > 0 {
		fmt.Println()
		fmt.Println("  Recovery Analysis:")
		fmt.Printf("    Energy used ratio: %.1f%% (%s / %s)\n",
			used.EnergyUsedRatio*100,
			formatNumber(used.EnergyUsedAtStart),
			formatNumber(analysis.EnergyStart+used.EnergyUsedAtStart))
		fmt.Printf("    Bandwidth used ratio: %.1f%%\n", used.BandwidthUsedRatio*100)
		if used.EstimatedFullRecoveryHours > 0 {
			fmt.Printf("    Estimated full recovery: %.1f hours\n", used.EstimatedFullRecoveryHours)
		}
		fmt.Printf("    Measured regen: %.1f/sec, Theoretical (used-based): %.1f/sec\n",
			used.MeasuredRecoveryRate, used.UsedBasedRecoveryRate)

		matchStr := "NO"
		if used.EnergyRecoveryMatchesUsedModel {
			matchStr = "YES"
		}
		fmt.Printf("    Matches used-based model: %s\n", matchStr)
	}

	// Formula validation
	fv := analysis.FormulaValidation
	if fv.BestFit != "" {
		fmt.Println()
		fmt.Println("  Formula Validation:")
		fmt.Printf("    Best fit model: %s (confidence: %.1f%%)\n", fv.BestFit, fv.Confidence*100)
	}

	// Theoretical comparison
	fmt.Println()
	fmt.Println("  Theoretical vs Measured (Regen Rate):")
	energyMatch := "NO"
	if analysis.EnergyRateMatchesTheory {
		energyMatch = "YES"
	}
	bwMatch := "NO"
	if analysis.BandwidthRateMatchesTheory {
		bwMatch = "YES"
	}
	fmt.Printf("    Energy:    theoretical %s/day, measured %s/day, match: %s\n",
		formatNumber(int64(analysis.TheoreticalEnergyRatePerDay)),
		formatNumber(int64(analysis.EnergyRegenRatePerDay)),
		energyMatch,
	)
	fmt.Printf("    Bandwidth: theoretical %s/day, measured %s/day, match: %s\n",
		formatNumber(int64(analysis.TheoreticalBandwidthRatePerDay)),
		formatNumber(int64(analysis.BandwidthRegenRatePerDay)),
		bwMatch,
	)

	// Practical estimates
	est := analysis.PracticalEstimates
	fmt.Println()
	fmt.Println("  Transaction Capacity (based on regen rate):")
	fmt.Printf("    Immediate (from buffer):\n")
	fmt.Printf("      At 65k Energy/tx:  %d tx\n", est.ImmediateCapacity65k)
	fmt.Printf("      At 131k Energy/tx: %d tx\n", est.ImmediateCapacity131k)
	fmt.Printf("    Sustained (regen only):\n")
	fmt.Printf("      At 65k Energy/tx:  %.0f tx/day\n", est.TxPerDay65kSustained)
	fmt.Printf("      At 131k Energy/tx: %.0f tx/day\n", est.TxPerDay131kSustained)
	fmt.Printf("    With buffer (immediate + regen):\n")
	fmt.Printf("      At 65k Energy/tx:  %.0f tx/day\n", est.TxPerDay65kWithBuffer)
	fmt.Printf("      At 131k Energy/tx: %.0f tx/day\n", est.TxPerDay131kWithBuffer)

	fmt.Println()
	fmt.Printf("Log saved to: %s\n", filename)
}

func formatFloat(f float64) string {
	if f >= 1000 {
		return formatNumber(int64(f))
	}
	return fmt.Sprintf("%.1f", f)
}

// PrintSimulation prints simulation results
func PrintSimulation(sim models.SimulationResult) {
	fmt.Println()
	fmt.Println(strings.Repeat("━", 60))
	fmt.Printf("Transaction Simulation (target: %d tx @ %s energy each)\n",
		sim.TargetTx, formatNumber(sim.TxCost))
	fmt.Println(strings.Repeat("━", 60))

	fmt.Printf("Current available: %s energy\n", formatNumber(sim.CurrentAvailable))
	fmt.Printf("Immediate capacity: %d tx\n", sim.ImmediateCapacity)
	fmt.Println()

	fmt.Printf("Recovery rate: %.1f energy/sec = 1 tx every %.1f sec\n",
		sim.RecoveryRatePerSec, sim.SecondsPerTx)
	fmt.Println()

	fmt.Println("Projection for next 24 hours:")
	for hour := 0; hour < 24; hour++ {
		if hour < 6 || hour >= 22 {
			fmt.Printf("  Hour %2d: %4d tx\n", hour, sim.HourlyProjection[hour])
		} else if hour == 6 {
			fmt.Println("  ...")
		}
	}
	fmt.Println()

	fmt.Printf("Total 24h: %d tx\n", sim.Total24hCapacity)
	fmt.Println()

	if sim.CanReachTarget {
		fmt.Printf("✓ Can reach target of %d tx/day\n", sim.TargetTx)
	} else {
		fmt.Printf("✗ Cannot reach %d tx/day with current resources\n", sim.TargetTx)
		fmt.Println()
		fmt.Printf("Required energy_limit for %d tx/day: %s\n",
			sim.TargetTx, formatNumber(sim.RequiredEnergyLimit))
	}
}

// PrintError prints an error in a formatted way
func PrintError(err error) {
	fmt.Printf("\nError: %v\n", err)
}

// PrintInterrupted prints a message when monitoring is interrupted
func PrintInterrupted() {
	fmt.Println("\n\nMonitoring interrupted by user.")
}

func formatNumber(n int64) string {
	if n == 0 {
		return "0"
	}

	sign := ""
	if n < 0 {
		sign = "-"
		n = -n
	}

	str := fmt.Sprintf("%d", n)
	result := ""

	for i, c := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result += ","
		}
		result += string(c)
	}

	return sign + result
}

func formatDelta(n int64) string {
	if n >= 0 {
		return "+" + formatNumber(n)
	}
	return formatNumber(n)
}
