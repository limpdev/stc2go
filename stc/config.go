package stc

import (
	"encoding/json"
	"fmt"
	"math"
)

// Config holds the static configuration for STC calculations
type Config struct {
	TaxRates   TaxRates   `json:"taxRates"`
	BrokerFees BrokerFees `json:"brokerFees"`
}

// TaxRates represents tax rate configuration
type TaxRates struct {
	Federal   float64 `json:"federal"`
	Medicare  float64 `json:"medicare"`
	SocialSec float64 `json:"socialSec"`
	State     float64 `json:"state"`
	LocalSDI  float64 `json:"localSdi"`
}

// BrokerFees represents broker fee configuration
type BrokerFees struct {
	CommissionRate float64 `json:"commissionRate"`
	MinimumFee     float64 `json:"minimumFee"`
}

// Input represents the user-provided inputs for STC calculation
type Input struct {
	ExercisePrice   float64 `json:"exercisePrice"`
	ExercisedShares float64 `json:"exercisedShares"`
	FMV             float64 `json:"fmv"`
}

// Result contains all calculated values from the STC calculation
type Result struct {
	// Input values
	ExercisePrice   float64 `json:"exercisePrice"`
	ExercisedShares float64 `json:"exercisedShares"`
	FMV             float64 `json:"fmv"`

	// Calculated costs
	OptionCost   float64 `json:"optionCost"`
	TaxableGain  float64 `json:"taxableGain"`
	FederalTax   float64 `json:"federalTax"`
	MedicareTax  float64 `json:"medicareTax"`
	SocialSecTax float64 `json:"socialSecTax"`
	StateTax     float64 `json:"stateTax"`
	LocalSDITax  float64 `json:"localSdiTax"`
	TotalTax     float64 `json:"totalTax"`

	// Broker fees
	BrokerCommission float64 `json:"brokerCommission"`
	BrokerFees       float64 `json:"brokerFees"`

	// Final calculations
	TotalCosts       float64 `json:"totalCosts"`
	SharesToSell     float64 `json:"sharesToSell"`
	EstGrossProceeds float64 `json:"estGrossProceeds"`
	Residual         float64 `json:"residual"`
	NetShares        float64 `json:"netShares"`
}

// Calculator handles STC calculations with a given configuration
type Calculator struct {
	config Config
}

// NewCalculator creates a new STC calculator with the given configuration
func NewCalculator(config Config) *Calculator {
	return &Calculator{config: config}
}

// NewDefaultCalculator creates a calculator with default tax rates and broker fees
func NewDefaultCalculator() *Calculator {
	return &Calculator{
		config: Config{
			TaxRates: TaxRates{
				Federal:   0.22,
				Medicare:  0.0145,
				SocialSec: 0.062,
				State:     0.0,
				LocalSDI:  0.0,
			},
			BrokerFees: BrokerFees{
				CommissionRate: 0.03,
				MinimumFee:     25.0,
			},
		},
	}
}

// Calculate performs the STC calculation
func (c *Calculator) Calculate(input Input) Result {
	result := Result{
		ExercisePrice:   input.ExercisePrice,
		ExercisedShares: input.ExercisedShares,
		FMV:             input.FMV,
	}

	// Calculate option cost and taxable gain
	result.OptionCost = roundMoney(input.ExercisedShares * input.ExercisePrice)
	result.TaxableGain = roundMoney((input.FMV - input.ExercisePrice) * input.ExercisedShares)

	// Calculate taxes
	// Note: We round these individually to simulate real-world line-item accounting
	result.FederalTax = roundMoney(result.TaxableGain * c.config.TaxRates.Federal)
	result.MedicareTax = roundMoney(result.TaxableGain * c.config.TaxRates.Medicare)
	result.SocialSecTax = roundMoney(result.TaxableGain * c.config.TaxRates.SocialSec)
	result.StateTax = roundMoney(result.TaxableGain * c.config.TaxRates.State)
	result.LocalSDITax = roundMoney(result.TaxableGain * c.config.TaxRates.LocalSDI)

	result.TotalTax = result.FederalTax + result.MedicareTax + result.SocialSecTax +
		result.StateTax + result.LocalSDITax

	// --- LOGIC FIX START ---

	// Base liability (Costs excluding broker fees)
	baseLiability := result.OptionCost + result.TotalTax

	// Initialize loop variables
	sharesToSell := 0.0

	// Initial guess: Cost / FMV, rounded UP to nearest whole number
	if input.FMV > 0 {
		sharesToSell = math.Ceil(baseLiability / input.FMV)
	}

	// Iteratively adjust for broker fees
	// We loop because adding a share to cover fees might increase fees enough
	// to require yet another share (edge case, but possible).
	const maxIterations = 100

	for i := 0; i < maxIterations; i++ {
		// 1. Calculate fees based on current WHOLE share count
		brokerCommission := sharesToSell * c.config.BrokerFees.CommissionRate
		brokerFeesApplied := math.Max(brokerCommission, c.config.BrokerFees.MinimumFee)

		// 2. Calculate Total Liability
		totalCosts := baseLiability + brokerFeesApplied

		// 3. Calculate new required shares (Rounded UP)
		newSharesToSell := math.Ceil(totalCosts / input.FMV)

		// 4. Check for stability
		if newSharesToSell == sharesToSell {
			// Logic has stabilized
			result.SharesToSell = sharesToSell
			result.BrokerCommission = brokerCommission
			result.BrokerFees = brokerFeesApplied
			result.TotalCosts = totalCosts
			break
		}

		sharesToSell = newSharesToSell
	}

	// Final Calculations
	// Gross Proceeds come from the WHOLE shares sold
	result.EstGrossProceeds = result.SharesToSell * input.FMV

	// Residual is the cash left over
	result.Residual = result.EstGrossProceeds - result.TotalCosts

	// Net shares is simple subtraction
	result.NetShares = input.ExercisedShares - result.SharesToSell

	// --- LOGIC FIX END ---

	return result
}

// UpdateTaxRates updates the tax rate configuration
func (c *Calculator) UpdateTaxRates(rates TaxRates) {
	c.config.TaxRates = rates
}

// UpdateBrokerFees updates the broker fee configuration
func (c *Calculator) UpdateBrokerFees(fees BrokerFees) {
	c.config.BrokerFees = fees
}

// GetConfig returns the current configuration
func (c *Calculator) GetConfig() Config {
	return c.config
}

// ToJSON converts the result to JSON string
func (r Result) ToJSON() (string, error) {
	bytes, err := json.MarshalIndent(r, "", "  ")
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// String returns a formatted string representation of the result
func (r Result) String() string {
	return fmt.Sprintf(
		"STC Result: %.4f shares to sell, $%.2f net proceeds, %.4f net shares remaining",
		r.SharesToSell,
		r.EstGrossProceeds-r.TotalCosts,
		r.NetShares,
	)
}

// roundMoney rounds a float64 to 2 decimal places for monetary values
func roundMoney(val float64) float64 {
	return math.Round(val*100) / 100
}
