package output

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/sxwebdev/tron-resource-calculator/internal/models"
)

// SaveJSON saves the monitoring report to a JSON file
func SaveJSON(report models.MonitorReport) (string, error) {
	filename := generateFilename(report.Metadata.Address, report.Metadata.StartTime)

	data, err := json.MarshalIndent(report, "", "  ")
	if err != nil {
		return "", fmt.Errorf("failed to marshal report: %w", err)
	}

	if err := os.WriteFile(filename, data, 0o644); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}

	return filename, nil
}

func generateFilename(address string, startTime time.Time) string {
	// Use first 4 and last 4 characters of address for short version
	shortAddr := address
	if len(address) > 8 {
		shortAddr = address[:4] + "..." + address[len(address)-4:]
	}

	timestamp := startTime.Format("20060102_150405")
	return fmt.Sprintf("tron_monitor_%s_%s.json", shortAddr, timestamp)
}

// BuildReport creates a MonitorReport from collected data
func BuildReport(
	address, node string,
	startTime, endTime time.Time,
	duration int,
	snapshots []models.ResourceSnapshot,
	analysis models.Analysis,
) models.MonitorReport {
	return models.MonitorReport{
		Metadata: models.Metadata{
			Address:         address,
			Node:            node,
			StartTime:       startTime,
			EndTime:         endTime,
			DurationSeconds: duration,
			SamplesCount:    len(snapshots),
		},
		Snapshots: snapshots,
		Analysis:  analysis,
	}
}
