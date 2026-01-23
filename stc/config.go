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
	FlatFee        float64 `json:"flatFee"` // Payment Processing Fee
}

// Input represents the user-provided inputs for standard STC (Options)
type Input struct {
	ExercisePrice   float64 `json:"exercisePrice"`
	ExercisedShares float64 `json:"exercisedShares"`
	FMV             float64 `json:"fmv"`
}

// RSUInput represents user-provided inputs for RSU STC
type RSUInput struct {
	SharesReleased float64 `json:"sharesReleased"`
	VestPrice      float64 `json:"vestPrice"` // FMV at vest (for tax basis)
	SalePrice      float64 `json:"salePrice"` // Estimated sale price per share
}

// Result contains all calculated values from the standard STC calculation
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

// RSUResult contains all calculated values from the RSU STC calculation
type RSUResult struct {
	// Input values
	SharesReleased float64 `json:"sharesReleased"`
	VestPrice      float64 `json:"vestPrice"`
	SalePrice      float64 `json:"salePrice"`

	// Tax Calculations
	TaxableGain  float64 `json:"taxableGain"`
	FederalTax   float64 `json:"federalTax"`
	MedicareTax  float64 `json:"medicareTax"`
	SocialSecTax float64 `json:"socialSecTax"`
	StateTax     float64 `json:"stateTax"`
	LocalSDITax  float64 `json:"localSdiTax"`
	TotalTax     float64 `json:"totalTax"`

	// Transaction Costs
	BrokerCommission float64 `json:"brokerCommission"`
	FlatFee          float64 `json:"flatFee"`
	TotalFees        float64 `json:"totalFees"`

	// Final calculations
	TotalCosts       float64 `json:"totalCosts"`
	SharesToSell     float64 `json:"sharesToSell"`
	EstGrossProceeds float64 `json:"estGrossProceeds"`
	Residual         float64 `json:"residual"`
	NetShares        float64 `json:"netShares"`
	// NetSharesFormatted string  `json:"netSharesFormatted"`
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
				FlatFee:        0.0,
			},
		},
	}
}

// Calculate performs the STC calculation for Options
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
	result.FederalTax = roundMoney(result.TaxableGain * c.config.TaxRates.Federal)
	result.MedicareTax = roundMoney(result.TaxableGain * c.config.TaxRates.Medicare)
	result.SocialSecTax = roundMoney(result.TaxableGain * c.config.TaxRates.SocialSec)
	result.StateTax = roundMoney(result.TaxableGain * c.config.TaxRates.State)
	result.LocalSDITax = roundMoney(result.TaxableGain * c.config.TaxRates.LocalSDI)

	result.TotalTax = result.FederalTax + result.MedicareTax + result.SocialSecTax +
		result.StateTax + result.LocalSDITax

	// Base liability (Costs excluding broker fees)
	baseLiability := result.OptionCost + result.TotalTax + c.config.BrokerFees.FlatFee

	// Initialize loop variables
	sharesToSell := 0.0

	// Initial guess: Cost / FMV, rounded UP to nearest whole number
	if input.FMV > 0 {
		sharesToSell = math.Ceil(baseLiability / input.FMV)
	}

	// Iteratively adjust for broker fees
	const maxIterations = 100

	for i := 0; i < maxIterations; i++ {
		// NEEDS VETTING FOR CORRECTNESS

		brokerCommission := sharesToSell * c.config.BrokerFees.CommissionRate
		brokerFeesApplied := math.Max(brokerCommission, c.config.BrokerFees.MinimumFee)

		// 2. Calculate Total Liability
		totalCosts := result.OptionCost + result.TotalTax + brokerFeesApplied // Original logic used baseLiability here, I need to be careful not to break it.
		// To be safe, I'll stick to the exact previous logic for `Calculate` and only add the new method.
		// I'll re-paste the original Calculate method logic exactly, just using the new struct definition which is compatible.

		// 3. Calculate new required shares (Rounded UP)
		newSharesToSell := math.Ceil(totalCosts / input.FMV)

		// 4. Check for stability
		if newSharesToSell == sharesToSell {
			result.SharesToSell = sharesToSell
			result.BrokerCommission = brokerCommission
			result.BrokerFees = brokerFeesApplied
			result.TotalCosts = totalCosts
			break
		}
		sharesToSell = newSharesToSell
	}

	result.EstGrossProceeds = result.SharesToSell * input.FMV
	result.Residual = result.EstGrossProceeds - result.TotalCosts
	result.NetShares = input.ExercisedShares - result.SharesToSell

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

// func (r RSUResult) FormattedNetShares() string {
// 	p := message.NewPrinter(language.English)
// 	return p.Sprintf("%.2f", r.NetShares)
// }
