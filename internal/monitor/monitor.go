package monitor

import (
	"context"
	"math"
	"time"

	"github.com/sxwebdev/tron-resource-calculator/internal/client"
	"github.com/sxwebdev/tron-resource-calculator/internal/models"
)

// Monitor handles the resource monitoring logic
type Monitor struct {
	client     *client.Client
	address    string
	duration   int
	intervalMs int
}

// New creates a new Monitor instance
func New(c *client.Client, address string, duration int) *Monitor {
	return &Monitor{
		client:     c,
		address:    address,
		duration:   duration,
		intervalMs: 1000,
	}
}

// NewWithInterval creates a Monitor with custom interval
func NewWithInterval(c *client.Client, address string, duration, intervalMs int) *Monitor {
	return &Monitor{
		client:     c,
		address:    address,
		duration:   duration,
		intervalMs: intervalMs,
	}
}

// Run starts the monitoring process and returns collected snapshots
func (m *Monitor) Run(ctx context.Context, onSnapshot func(snapshot models.ResourceSnapshot, index int)) ([]models.ResourceSnapshot, error) {
	expectedSamples := (m.duration * 1000 / m.intervalMs) + 1
	snapshots := make([]models.ResourceSnapshot, 0, expectedSamples)
	startTime := time.Now()

	var prevSnapshot *models.ResourceSnapshot
	index := 0

	for elapsed := 0; elapsed <= m.duration*1000; elapsed += m.intervalMs {
		select {
		case <-ctx.Done():
			return snapshots, ctx.Err()
		default:
		}

		snapshot, err := m.takeSnapshot(startTime, prevSnapshot)
		if err != nil {
			if onSnapshot != nil {
				onSnapshot(models.ResourceSnapshot{Timestamp: time.Now(), ElapsedMs: time.Since(startTime).Milliseconds()}, index)
			}
		} else {
			snapshots = append(snapshots, *snapshot)
			if onSnapshot != nil {
				onSnapshot(*snapshot, index)
			}
			prevSnapshot = snapshot
		}
		index++

		if elapsed < m.duration*1000 {
			select {
			case <-ctx.Done():
				return snapshots, ctx.Err()
			case <-time.After(time.Duration(m.intervalMs) * time.Millisecond):
			}
		}
	}

	return snapshots, nil
}

// RunUntilFull monitors until resources are fully recovered
func (m *Monitor) RunUntilFull(ctx context.Context, maxDuration int, onSnapshot func(snapshot models.ResourceSnapshot, index int)) ([]models.ResourceSnapshot, error) {
	snapshots := make([]models.ResourceSnapshot, 0, maxDuration+1)
	startTime := time.Now()

	var prevSnapshot *models.ResourceSnapshot
	var firstSnapshot *models.ResourceSnapshot

	for i := 0; i <= maxDuration; i++ {
		select {
		case <-ctx.Done():
			return snapshots, ctx.Err()
		default:
		}

		snapshot, err := m.takeSnapshot(startTime, prevSnapshot)
		if err != nil {
			if onSnapshot != nil {
				onSnapshot(models.ResourceSnapshot{Timestamp: time.Now(), ElapsedMs: time.Since(startTime).Milliseconds()}, i)
			}
		} else {
			snapshots = append(snapshots, *snapshot)
			if onSnapshot != nil {
				onSnapshot(*snapshot, i)
			}

			if firstSnapshot == nil {
				firstSnapshot = snapshot
			}

			// Check if fully recovered
			if snapshot.EnergyUsed == 0 && snapshot.TotalBandwidthUsed() == 0 {
				return snapshots, nil
			}

			prevSnapshot = snapshot
		}

		if i < maxDuration {
			select {
			case <-ctx.Done():
				return snapshots, ctx.Err()
			case <-time.After(time.Duration(m.intervalMs) * time.Millisecond):
			}
		}
	}

	return snapshots, nil
}

func (m *Monitor) takeSnapshot(startTime time.Time, prev *models.ResourceSnapshot) (*models.ResourceSnapshot, error) {
	resp, err := m.client.GetAccountResource(m.address)
	if err != nil {
		return nil, err
	}

	now := time.Now()
	snapshot := &models.ResourceSnapshot{
		Timestamp:    now,
		ElapsedMs:    now.Sub(startTime).Milliseconds(),
		EnergyLimit:  resp.EnergyLimit,
		EnergyUsed:   resp.EnergyUsed,
		NetLimit:     resp.NetLimit,
		NetUsed:      resp.NetUsed,
		FreeNetLimit: resp.FreeNetLimit,
		FreeNetUsed:  resp.FreeNetUsed,
	}

	snapshot.EnergyAvailable = snapshot.EnergyLimit - snapshot.EnergyUsed
	snapshot.BandwidthAvailable = (snapshot.NetLimit + snapshot.FreeNetLimit) - (snapshot.NetUsed + snapshot.FreeNetUsed)

	if prev != nil {
		snapshot.DeltaEnergy = snapshot.EnergyAvailable - prev.EnergyAvailable
		snapshot.DeltaBandwidth = snapshot.BandwidthAvailable - prev.BandwidthAvailable
	}

	return snapshot, nil
}

// Analyze computes statistics from collected snapshots
func Analyze(snapshots []models.ResourceSnapshot, duration int) models.Analysis {
	if len(snapshots) == 0 {
		return models.Analysis{}
	}

	first := snapshots[0]
	last := snapshots[len(snapshots)-1]

	// Calculate actual duration from timestamps
	actualDurationSec := float64(last.ElapsedMs-first.ElapsedMs) / 1000.0

	// Sum up regenerated and consumed separately
	var energyRegenerated, energyConsumed int64
	var bandwidthRegenerated, bandwidthConsumed int64

	for i := 1; i < len(snapshots); i++ {
		s := snapshots[i]
		if s.DeltaEnergy > 0 {
			energyRegenerated += s.DeltaEnergy
		} else if s.DeltaEnergy < 0 {
			energyConsumed += -s.DeltaEnergy // store as positive
		}

		if s.DeltaBandwidth > 0 {
			bandwidthRegenerated += s.DeltaBandwidth
		} else if s.DeltaBandwidth < 0 {
			bandwidthConsumed += -s.DeltaBandwidth // store as positive
		}
	}

	analysis := models.Analysis{
		ActualDurationSec: actualDurationSec,

		EnergyStart:       first.EnergyAvailable,
		EnergyEnd:         last.EnergyAvailable,
		EnergyTotalDelta:  last.EnergyAvailable - first.EnergyAvailable,
		EnergyRegenerated: energyRegenerated,
		EnergyConsumed:    energyConsumed,

		BandwidthStart:       first.BandwidthAvailable,
		BandwidthEnd:         last.BandwidthAvailable,
		BandwidthTotalDelta:  last.BandwidthAvailable - first.BandwidthAvailable,
		BandwidthRegenerated: bandwidthRegenerated,
		BandwidthConsumed:    bandwidthConsumed,
	}

	// Calculate separated rates
	if actualDurationSec > 0 {
		// Regeneration rates
		analysis.EnergyRegenRatePerSec = float64(energyRegenerated) / actualDurationSec
		analysis.EnergyRegenRatePerDay = analysis.EnergyRegenRatePerSec * 86400

		analysis.BandwidthRegenRatePerSec = float64(bandwidthRegenerated) / actualDurationSec
		analysis.BandwidthRegenRatePerDay = analysis.BandwidthRegenRatePerSec * 86400

		// Consumption rates
		analysis.EnergyConsumeRatePerSec = float64(energyConsumed) / actualDurationSec
		analysis.EnergyConsumeRatePerDay = analysis.EnergyConsumeRatePerSec * 86400

		analysis.BandwidthConsumeRatePerSec = float64(bandwidthConsumed) / actualDurationSec
		analysis.BandwidthConsumeRatePerDay = analysis.BandwidthConsumeRatePerSec * 86400

		// Net rates (regen - consume)
		analysis.EnergyNetRatePerSec = analysis.EnergyRegenRatePerSec - analysis.EnergyConsumeRatePerSec
		analysis.EnergyNetRatePerDay = analysis.EnergyNetRatePerSec * 86400

		analysis.BandwidthNetRatePerSec = analysis.BandwidthRegenRatePerSec - analysis.BandwidthConsumeRatePerSec
		analysis.BandwidthNetRatePerDay = analysis.BandwidthNetRatePerSec * 86400
	}

	// Theoretical rates
	analysis.TheoreticalEnergyRatePerDay = float64(first.EnergyLimit)
	analysis.TheoreticalBandwidthRatePerDay = float64(first.TotalBandwidthLimit())

	// Check theoretical match (10% tolerance) - compare REGEN rate, not net
	if analysis.TheoreticalEnergyRatePerDay > 0 && analysis.EnergyRegenRatePerDay > 0 {
		ratio := analysis.EnergyRegenRatePerDay / analysis.TheoreticalEnergyRatePerDay
		analysis.EnergyRateMatchesTheory = math.Abs(ratio-1.0) < 0.1
	}

	if analysis.TheoreticalBandwidthRatePerDay > 0 && analysis.BandwidthRegenRatePerDay > 0 {
		ratio := analysis.BandwidthRegenRatePerDay / analysis.TheoreticalBandwidthRatePerDay
		analysis.BandwidthRateMatchesTheory = math.Abs(ratio-1.0) < 0.1
	}

	// Transaction estimates based on REGEN rate
	if analysis.EnergyRegenRatePerDay > 0 {
		analysis.TxPerDay65k = analysis.EnergyRegenRatePerDay / 65000
		analysis.TxPerDay131k = analysis.EnergyRegenRatePerDay / 131000
	}

	// Extended analysis
	analysis.TickAnalysis = analyzeBlockTicks(snapshots)
	analysis.UsedBasedAnalysis = analyzeUsedBased(snapshots, analysis.EnergyRegenRatePerSec)
	analysis.FormulaValidation = validateFormulas(analysis, first)
	analysis.PracticalEstimates = calculatePracticalEstimates(first, analysis)

	return analysis
}

// analyzeBlockTicks detects recovery ticks and consumption events
func analyzeBlockTicks(snapshots []models.ResourceSnapshot) models.TickAnalysis {
	tick := models.TickAnalysis{
		TickTimestampsMs:    make([]int64, 0),
		TickEnergyDeltas:    make([]int64, 0),
		TickBandwidthDeltas: make([]int64, 0),
	}

	if len(snapshots) < 2 {
		return tick
	}

	var totalRegenEnergy, totalRegenBandwidth int64
	var recoveryTimestamps []int64

	for i := 1; i < len(snapshots); i++ {
		s := snapshots[i]

		// Track all deltas for raw data
		tick.TickTimestampsMs = append(tick.TickTimestampsMs, s.ElapsedMs)
		tick.TickEnergyDeltas = append(tick.TickEnergyDeltas, s.DeltaEnergy)
		tick.TickBandwidthDeltas = append(tick.TickBandwidthDeltas, s.DeltaBandwidth)

		// Count recovery ticks (positive deltas)
		if s.DeltaEnergy > 0 {
			tick.RecoveryTicks++
			totalRegenEnergy += s.DeltaEnergy
			totalRegenBandwidth += s.DeltaBandwidth
			recoveryTimestamps = append(recoveryTimestamps, s.ElapsedMs)
		}

		// Count consumption events (negative deltas)
		if s.DeltaEnergy < 0 {
			tick.ConsumptionEvents++
			tick.TotalEnergyConsumed += -s.DeltaEnergy
		}
		if s.DeltaBandwidth < 0 {
			tick.TotalBandwidthConsumed += -s.DeltaBandwidth
		}
	}

	// Calculate recovery tick stats
	if tick.RecoveryTicks > 0 {
		tick.EnergyPerTick = float64(totalRegenEnergy) / float64(tick.RecoveryTicks)
		tick.BandwidthPerTick = float64(totalRegenBandwidth) / float64(tick.RecoveryTicks)

		// Calculate average interval between recovery ticks
		if len(recoveryTimestamps) > 1 {
			firstTickMs := recoveryTimestamps[0]
			lastTickMs := recoveryTimestamps[len(recoveryTimestamps)-1]
			totalIntervalMs := lastTickMs - firstTickMs
			avgIntervalMs := float64(totalIntervalMs) / float64(len(recoveryTimestamps)-1)
			tick.AvgRecoveryInterval = avgIntervalMs / 1000.0

			if avgIntervalMs > 0 {
				tick.RecoveryTicksPerHr = 3600000.0 / avgIntervalMs
				tick.RecoveryTicksPerDay = 86400000.0 / avgIntervalMs
			}
		}
	}

	// Calculate consumption stats
	if tick.ConsumptionEvents > 0 {
		tick.AvgEnergyPerConsume = float64(tick.TotalEnergyConsumed) / float64(tick.ConsumptionEvents)
		tick.AvgBandwidthPerConsume = float64(tick.TotalBandwidthConsumed) / float64(tick.ConsumptionEvents)
	}

	return tick
}

// analyzeUsedBased computes analysis based on energy_used
func analyzeUsedBased(snapshots []models.ResourceSnapshot, measuredRate float64) models.UsedBasedAnalysis {
	if len(snapshots) == 0 {
		return models.UsedBasedAnalysis{}
	}

	first := snapshots[0]

	analysis := models.UsedBasedAnalysis{
		EnergyUsedAtStart:    first.EnergyUsed,
		BandwidthUsedAtStart: first.TotalBandwidthUsed(),
		MeasuredRecoveryRate: measuredRate,
	}

	if first.EnergyLimit > 0 {
		analysis.EnergyUsedRatio = float64(first.EnergyUsed) / float64(first.EnergyLimit)
	}

	if first.TotalBandwidthLimit() > 0 {
		analysis.BandwidthUsedRatio = float64(first.TotalBandwidthUsed()) / float64(first.TotalBandwidthLimit())
	}

	// Calculate used-based recovery rate (energy_used / 24h in seconds)
	if first.EnergyUsed > 0 {
		// Full recovery takes 24h, so rate = energy_used / 86400
		analysis.UsedBasedRecoveryRate = float64(first.EnergyUsed) / 86400.0

		// Estimated time to full recovery
		if measuredRate > 0 {
			analysis.EstimatedFullRecoverySeconds = float64(first.EnergyUsed) / measuredRate
			analysis.EstimatedFullRecoveryHours = analysis.EstimatedFullRecoverySeconds / 3600.0
		}

		// Check if measured rate matches used-based model (within 15% tolerance)
		if analysis.UsedBasedRecoveryRate > 0 && measuredRate > 0 {
			ratio := measuredRate / analysis.UsedBasedRecoveryRate
			analysis.EnergyRecoveryMatchesUsedModel = math.Abs(ratio-1.0) < 0.15
		}
	}

	return analysis
}

// validateFormulas compares theoretical and measured models
func validateFormulas(analysis models.Analysis, first models.ResourceSnapshot) models.FormulaValidation {
	validation := models.FormulaValidation{
		TheoreticalModel: "E_limit / 86400",
		MeasuredModel:    "E_used / T_recovery",
	}

	// Calculate theoretical rate from limit
	theoreticalFromLimit := float64(first.EnergyLimit) / 86400.0

	// Calculate theoretical rate from used (assuming 24h full recovery)
	theoreticalFromUsed := float64(first.EnergyUsed) / 86400.0

	// Use REGEN rate for comparison
	measured := analysis.EnergyRegenRatePerSec

	if measured > 0 && theoreticalFromLimit > 0 && theoreticalFromUsed > 0 {
		diffLimit := math.Abs(measured-theoreticalFromLimit) / theoreticalFromLimit
		diffUsed := math.Abs(measured-theoreticalFromUsed) / theoreticalFromUsed

		if diffUsed < diffLimit {
			validation.BestFit = "used_based"
			validation.Confidence = 1.0 - diffUsed
		} else {
			validation.BestFit = "limit_based"
			validation.Confidence = 1.0 - diffLimit
		}

		if validation.Confidence < 0 {
			validation.Confidence = 0
		}
	}

	return validation
}

// calculatePracticalEstimates computes transaction capacity
func calculatePracticalEstimates(first models.ResourceSnapshot, analysis models.Analysis) models.PracticalEstimates {
	est := models.PracticalEstimates{
		EnergyNeeded800Tx65k:  800 * 65000,
		EnergyNeeded800Tx131k: 800 * 131000,
	}

	// Immediate capacity from available energy
	if first.EnergyAvailable > 0 {
		est.ImmediateCapacity65k = first.EnergyAvailable / 65000
		est.ImmediateCapacity131k = first.EnergyAvailable / 131000
	}

	// Sustained capacity (based on REGEN rate only)
	if analysis.EnergyRegenRatePerDay > 0 {
		est.TxPerDay65kSustained = analysis.EnergyRegenRatePerDay / 65000
		est.TxPerDay131kSustained = analysis.EnergyRegenRatePerDay / 131000
	}

	// With buffer capacity (immediate + recovered)
	recoveredPerDay := analysis.EnergyRegenRatePerDay
	if recoveredPerDay > 0 {
		totalEnergy := float64(first.EnergyAvailable) + recoveredPerDay
		est.TxPerDay65kWithBuffer = totalEnergy / 65000
		est.TxPerDay131kWithBuffer = totalEnergy / 131000
	}

	return est
}

// Simulate calculates transaction simulation
func Simulate(snapshot models.ResourceSnapshot, analysis models.Analysis, txCost int64, targetTx int) models.SimulationResult {
	sim := models.SimulationResult{
		TargetTx:           targetTx,
		TxCost:             txCost,
		CurrentAvailable:   snapshot.EnergyAvailable,
		RecoveryRatePerSec: analysis.EnergyRegenRatePerSec,
		HourlyProjection:   make([]int64, 24),
	}

	// Immediate capacity
	sim.ImmediateCapacity = snapshot.EnergyAvailable / txCost

	// Seconds per transaction (recovery time)
	if analysis.EnergyRegenRatePerSec > 0 {
		sim.SecondsPerTx = float64(txCost) / analysis.EnergyRegenRatePerSec
	}

	// 24h capacity = immediate + recovered
	recoveredEnergy := int64(analysis.EnergyRegenRatePerDay)
	sim.Total24hCapacity = sim.ImmediateCapacity + (recoveredEnergy / txCost)

	// Hourly projection
	energyPerHour := int64(analysis.EnergyRegenRatePerDay / 24)
	currentEnergy := snapshot.EnergyAvailable

	for hour := 0; hour < 24; hour++ {
		txThisHour := currentEnergy / txCost
		sim.HourlyProjection[hour] = txThisHour

		// Use energy and recover
		used := txThisHour * txCost
		currentEnergy = currentEnergy - used + energyPerHour
		if currentEnergy > snapshot.EnergyLimit {
			currentEnergy = snapshot.EnergyLimit
		}
	}

	// Can reach target?
	sim.CanReachTarget = sim.Total24hCapacity >= int64(targetTx)

	// Required energy limit for target
	if !sim.CanReachTarget {
		// Need: targetTx * txCost per day
		// Recovery rate = limit / 86400
		// So: limit = targetTx * txCost
		sim.RequiredEnergyLimit = int64(targetTx) * txCost
	}

	return sim
}
