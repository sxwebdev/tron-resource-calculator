package models

import "time"

// ResourceSnapshot represents a single measurement of TRON account resources
type ResourceSnapshot struct {
	Timestamp  time.Time `json:"timestamp"`
	ElapsedMs  int64     `json:"elapsed_ms"`

	// Energy
	EnergyLimit int64 `json:"energy_limit"`
	EnergyUsed  int64 `json:"energy_used"`

	// Bandwidth
	NetLimit     int64 `json:"net_limit"`
	NetUsed      int64 `json:"net_used"`
	FreeNetLimit int64 `json:"free_net_limit"`
	FreeNetUsed  int64 `json:"free_net_used"`

	// Computed
	EnergyAvailable    int64 `json:"energy_available"`
	BandwidthAvailable int64 `json:"bandwidth_available"`

	// Deltas from previous snapshot
	DeltaEnergy    int64 `json:"delta_energy"`
	DeltaBandwidth int64 `json:"delta_bandwidth"`
}

// TotalBandwidthLimit returns total bandwidth limit (staked + free)
func (s *ResourceSnapshot) TotalBandwidthLimit() int64 {
	return s.NetLimit + s.FreeNetLimit
}

// TotalBandwidthUsed returns total bandwidth used (staked + free)
func (s *ResourceSnapshot) TotalBandwidthUsed() int64 {
	return s.NetUsed + s.FreeNetUsed
}

// APIResponse represents the response from TRON getaccountresource API
type APIResponse struct {
	FreeNetLimit      int64 `json:"freeNetLimit"`
	FreeNetUsed       int64 `json:"freeNetUsed"`
	NetLimit          int64 `json:"NetLimit"`
	NetUsed           int64 `json:"NetUsed"`
	EnergyLimit       int64 `json:"EnergyLimit"`
	EnergyUsed        int64 `json:"EnergyUsed"`
	TotalNetLimit     int64 `json:"TotalNetLimit"`
	TotalNetWeight    int64 `json:"TotalNetWeight"`
	TotalEnergyLimit  int64 `json:"TotalEnergyLimit"`
	TotalEnergyWeight int64 `json:"TotalEnergyWeight"`
}

// Metadata contains information about the monitoring session
type Metadata struct {
	Address         string    `json:"address"`
	Node            string    `json:"node"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	DurationSeconds int       `json:"duration_seconds"`
	SamplesCount    int       `json:"samples_count"`
	IntervalMs      int       `json:"interval_ms"`
}

// TickAnalysis contains block tick detection results
type TickAnalysis struct {
	// Recovery ticks (positive deltas)
	RecoveryTicks       int       `json:"recovery_ticks"`
	AvgRecoveryInterval float64   `json:"avg_recovery_interval_sec"`
	EnergyPerTick       float64   `json:"energy_per_tick"`
	BandwidthPerTick    float64   `json:"bandwidth_per_tick"`
	RecoveryTicksPerHr  float64   `json:"recovery_ticks_per_hour"`
	RecoveryTicksPerDay float64   `json:"recovery_ticks_per_day"`

	// Consumption events (negative deltas)
	ConsumptionEvents     int     `json:"consumption_events"`
	TotalEnergyConsumed   int64   `json:"total_energy_consumed"`
	TotalBandwidthConsumed int64  `json:"total_bandwidth_consumed"`
	AvgEnergyPerConsume   float64 `json:"avg_energy_per_consumption"`
	AvgBandwidthPerConsume float64 `json:"avg_bandwidth_per_consumption"`

	// Raw data
	TickTimestampsMs    []int64 `json:"tick_timestamps_ms"`
	TickEnergyDeltas    []int64 `json:"tick_energy_deltas"`
	TickBandwidthDeltas []int64 `json:"tick_bandwidth_deltas"`
}

// UsedBasedAnalysis contains analysis based on resources used
type UsedBasedAnalysis struct {
	EnergyUsedRatio               float64 `json:"energy_used_ratio"`
	BandwidthUsedRatio            float64 `json:"bandwidth_used_ratio"`
	EnergyUsedAtStart             int64   `json:"energy_used_at_start"`
	BandwidthUsedAtStart          int64   `json:"bandwidth_used_at_start"`
	EstimatedFullRecoverySeconds  float64 `json:"estimated_full_recovery_seconds"`
	EstimatedFullRecoveryHours    float64 `json:"estimated_full_recovery_hours"`
	EnergyRecoveryMatchesUsedModel bool   `json:"energy_recovery_matches_used_model"`
	MeasuredRecoveryRate          float64 `json:"measured_recovery_rate_per_sec"`
	UsedBasedRecoveryRate         float64 `json:"used_based_recovery_rate_per_sec"`
}

// FormulaValidation contains model comparison
type FormulaValidation struct {
	TheoreticalModel string  `json:"theoretical_model"`
	MeasuredModel    string  `json:"measured_model"`
	BestFit          string  `json:"best_fit"`
	Confidence       float64 `json:"confidence"`
}

// PracticalEstimates contains transaction capacity estimates
type PracticalEstimates struct {
	TxPerDay65kWithBuffer  float64 `json:"tx_per_day_65k_with_buffer"`
	TxPerDay65kSustained   float64 `json:"tx_per_day_65k_sustained"`
	TxPerDay131kWithBuffer float64 `json:"tx_per_day_131k_with_buffer"`
	TxPerDay131kSustained  float64 `json:"tx_per_day_131k_sustained"`
	EnergyNeeded800Tx65k   int64   `json:"energy_needed_for_800_tx_65k"`
	EnergyNeeded800Tx131k  int64   `json:"energy_needed_for_800_tx_131k"`
	ImmediateCapacity65k   int64   `json:"immediate_capacity_65k"`
	ImmediateCapacity131k  int64   `json:"immediate_capacity_131k"`
}

// Analysis contains calculated statistics from the monitoring session
type Analysis struct {
	// Timing
	ActualDurationSec float64 `json:"actual_duration_seconds"`

	// Energy stats
	EnergyStart      int64 `json:"energy_start"`
	EnergyEnd        int64 `json:"energy_end"`
	EnergyTotalDelta int64 `json:"energy_total_delta"`

	// Separated rates for Energy
	EnergyRegenerated    int64   `json:"energy_regenerated"`
	EnergyConsumed       int64   `json:"energy_consumed"`
	EnergyRegenRatePerSec float64 `json:"energy_regen_rate_per_second"`
	EnergyRegenRatePerDay float64 `json:"energy_regen_rate_per_day"`
	EnergyConsumeRatePerSec float64 `json:"energy_consume_rate_per_second"`
	EnergyConsumeRatePerDay float64 `json:"energy_consume_rate_per_day"`
	EnergyNetRatePerSec   float64 `json:"energy_net_rate_per_second"`
	EnergyNetRatePerDay   float64 `json:"energy_net_rate_per_day"`

	// Bandwidth stats
	BandwidthStart      int64 `json:"bandwidth_start"`
	BandwidthEnd        int64 `json:"bandwidth_end"`
	BandwidthTotalDelta int64 `json:"bandwidth_total_delta"`

	// Separated rates for Bandwidth
	BandwidthRegenerated    int64   `json:"bandwidth_regenerated"`
	BandwidthConsumed       int64   `json:"bandwidth_consumed"`
	BandwidthRegenRatePerSec float64 `json:"bandwidth_regen_rate_per_second"`
	BandwidthRegenRatePerDay float64 `json:"bandwidth_regen_rate_per_day"`
	BandwidthConsumeRatePerSec float64 `json:"bandwidth_consume_rate_per_second"`
	BandwidthConsumeRatePerDay float64 `json:"bandwidth_consume_rate_per_day"`
	BandwidthNetRatePerSec   float64 `json:"bandwidth_net_rate_per_second"`
	BandwidthNetRatePerDay   float64 `json:"bandwidth_net_rate_per_day"`

	// Theoretical rates
	TheoreticalEnergyRatePerDay    float64 `json:"theoretical_energy_rate_per_day"`
	TheoreticalBandwidthRatePerDay float64 `json:"theoretical_bandwidth_rate_per_day"`
	EnergyRateMatchesTheory        bool    `json:"energy_rate_matches_theory"`
	BandwidthRateMatchesTheory     bool    `json:"bandwidth_rate_matches_theory"`

	// Transaction estimates (based on regen rate)
	TxPerDay65k  float64 `json:"tx_per_day_65k_energy"`
	TxPerDay131k float64 `json:"tx_per_day_131k_energy"`

	// Extended analysis
	TickAnalysis       TickAnalysis       `json:"tick_analysis"`
	UsedBasedAnalysis  UsedBasedAnalysis  `json:"used_based_analysis"`
	FormulaValidation  FormulaValidation  `json:"formula_validation"`
	PracticalEstimates PracticalEstimates `json:"practical_estimates"`
}

// MonitorReport is the complete output structure for JSON export
type MonitorReport struct {
	Metadata  Metadata           `json:"metadata"`
	Snapshots []ResourceSnapshot `json:"snapshots"`
	Analysis  Analysis           `json:"analysis"`
}

// SimulationResult contains transaction simulation output
type SimulationResult struct {
	TargetTx           int     `json:"target_tx"`
	TxCost             int64   `json:"tx_cost_energy"`
	CurrentAvailable   int64   `json:"current_available_energy"`
	ImmediateCapacity  int64   `json:"immediate_capacity"`
	RecoveryRatePerSec float64 `json:"recovery_rate_per_sec"`
	SecondsPerTx       float64 `json:"seconds_per_tx"`
	Total24hCapacity   int64   `json:"total_24h_capacity"`
	CanReachTarget     bool    `json:"can_reach_target"`
	RequiredEnergyLimit int64  `json:"required_energy_limit_for_target"`
	HourlyProjection   []int64 `json:"hourly_projection"`
}

// Config holds CLI configuration
type Config struct {
	Address     string
	Node        string
	Duration    int
	IntervalMs  int
	UntilFull   bool
	MaxDuration int
	CompareFile string
	Simulate    bool
	TxCost      int64
	TargetTx    int
}
