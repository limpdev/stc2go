package stc

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
)

// BatchInput represents a batch of STC calculations
type BatchInput struct {
	Inputs []Input
}

// BatchResult represents results from a batch of STC calculations
type BatchResult struct {
	Results []Result
}

// CalculateBatch processes multiple STC calculations at once
func (c *Calculator) CalculateBatch(inputs []Input) BatchResult {
	results := make([]Result, len(inputs))
	for i, input := range inputs {
		results[i] = c.Calculate(input)
	}
	return BatchResult{Results: results}
}

// CSVRow represents a row in the CSV export
type CSVRow struct {
	ExercisePrice    string
	ExercisedShares  string
	FMV              string
	SharesToSell     string
	NetShares        string
	TotalCosts       string
	EstGrossProceeds string
	TaxableGain      string
	TotalTax         string
}

// ToCSV writes batch results to a CSV writer
func (br BatchResult) ToCSV(w io.Writer) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	// Write header
	header := []string{
		"Exercise Price",
		"Exercised Shares",
		"FMV",
		"Shares To Sell",
		"Net Shares",
		"Total Costs",
		"Est. Gross Proceeds",
		"Taxable Gain",
		"Total Tax",
		"Option Cost",
		"Broker Fees",
	}
	if err := writer.Write(header); err != nil {
		return fmt.Errorf("failed to write header: %w", err)
	}

	// Write data rows
	for _, result := range br.Results {
		row := []string{
			fmt.Sprintf("%.2f", result.ExercisePrice),
			fmt.Sprintf("%.2f", result.ExercisedShares),
			fmt.Sprintf("%.2f", result.FMV),
			fmt.Sprintf("%.4f", result.SharesToSell),
			fmt.Sprintf("%.4f", result.NetShares),
			fmt.Sprintf("%.2f", result.TotalCosts),
			fmt.Sprintf("%.2f", result.EstGrossProceeds),
			fmt.Sprintf("%.2f", result.TaxableGain),
			fmt.Sprintf("%.2f", result.TotalTax),
			fmt.Sprintf("%.2f", result.OptionCost),
			fmt.Sprintf("%.2f", result.BrokerFees),
		}
		if err := writer.Write(row); err != nil {
			return fmt.Errorf("failed to write row: %w", err)
		}
	}

	return nil
}

// FromCSV reads inputs from a CSV reader
func FromCSV(r io.Reader) ([]Input, error) {
	reader := csv.NewReader(r)

	// Skip header
	if _, err := reader.Read(); err != nil {
		return nil, fmt.Errorf("failed to read header: %w", err)
	}

	var inputs []Input

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to read row: %w", err)
		}

		if len(record) < 3 {
			continue // Skip invalid rows
		}

		exercisePrice, err := strconv.ParseFloat(record[0], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid exercise price: %w", err)
		}

		exercisedShares, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid exercised shares: %w", err)
		}

		fmv, err := strconv.ParseFloat(record[2], 64)
		if err != nil {
			return nil, fmt.Errorf("invalid FMV: %w", err)
		}

		inputs = append(inputs, Input{
			ExercisePrice:   exercisePrice,
			ExercisedShares: exercisedShares,
			FMV:             fmv,
		})
	}

	return inputs, nil
}

// Summary provides aggregate statistics for a batch of results
type Summary struct {
	TotalExercisedShares float64
	TotalSharesToSell    float64
	TotalNetShares       float64
	TotalCosts           float64
	TotalTaxes           float64
	TotalBrokerFees      float64
	AverageFMV           float64
	Count                int
}

// Summarize generates aggregate statistics from batch results
func (br BatchResult) Summarize() Summary {
	if len(br.Results) == 0 {
		return Summary{}
	}

	summary := Summary{
		Count: len(br.Results),
	}

	for _, result := range br.Results {
		summary.TotalExercisedShares += result.ExercisedShares
		summary.TotalSharesToSell += result.SharesToSell
		summary.TotalNetShares += result.NetShares
		summary.TotalCosts += result.TotalCosts
		summary.TotalTaxes += result.TotalTax
		summary.TotalBrokerFees += result.BrokerFees
		summary.AverageFMV += result.FMV
	}

	summary.AverageFMV /= float64(summary.Count)

	return summary
}

// String returns a formatted string representation of the summary
func (s Summary) String() string {
	return fmt.Sprintf(
		"Batch Summary (%d calculations):\n"+
			"  Total Exercised Shares: %.2f\n"+
			"  Total Shares To Sell:   %.2f\n"+
			"  Total Net Shares:       %.2f\n"+
			"  Total Costs:            $%.2f\n"+
			"  Total Taxes:            $%.2f\n"+
			"  Total Broker Fees:      $%.2f\n"+
			"  Average FMV:            $%.2f",
		s.Count,
		s.TotalExercisedShares,
		s.TotalSharesToSell,
		s.TotalNetShares,
		s.TotalCosts,
		s.TotalTaxes,
		s.TotalBrokerFees,
		s.AverageFMV,
	)
}
