package stc

import (
	"math"
)

// CalculateRSU performs the STC calculation for Restricted Stock Units
func (c *Calculator) CalculateRSU(input RSUInput) RSUResult {
	result := RSUResult{
		SharesReleased: input.SharesReleased,
		VestPrice:      input.VestPrice,
		SalePrice:      input.SalePrice,
	}

	// 1. Calculate Taxable Gain (Basis is FMV at Vest)
	result.TaxableGain = roundMoney(input.SharesReleased * input.VestPrice)

	// 2. Calculate Taxes
	result.FederalTax = roundMoney(result.TaxableGain * c.config.TaxRates.Federal)
	result.MedicareTax = roundMoney(result.TaxableGain * c.config.TaxRates.Medicare)
	result.SocialSecTax = roundMoney(result.TaxableGain * c.config.TaxRates.SocialSec)
	result.StateTax = roundMoney(result.TaxableGain * c.config.TaxRates.State)
	result.LocalSDITax = roundMoney(result.TaxableGain * c.config.TaxRates.LocalSDI)

	result.TotalTax = result.FederalTax + result.MedicareTax + result.SocialSecTax +
		result.StateTax + result.LocalSDITax

	// 3. Base Liability (Taxes + Flat Fees)
	// Processing fees are added to the liability we must cover by selling shares.
	// baseLiability := result.TotalTax + c.config.BrokerFees.FlatFee
	// baseFee := c.config.BrokerFees.FlatFee
	baseBurden := result.TotalTax
	// 4. Iterative Solver for Shares to Sell
	// We need to cover: BaseLiability + Commission
	// Commission depends on Gross Proceeds (SharesSold * SalePrice)

	sharesToSell := 0.0

	// Initial guess
	if input.SalePrice > 0 {
		sharesToSell = math.Ceil(baseBurden / input.SalePrice)
	}

	const maxIterations = 100
	for i := 0; i < maxIterations; i++ {
		// Calculate Gross Proceeds for current guess
		sharesReleased := input.SharesReleased
		grossProceeds := (sharesReleased - sharesToSell) * input.SalePrice

		// Calculate Commission
		// Assumes CommissionRate is a percentage (e.g., 0.03 for 3%)
		commission := sharesToSell * c.config.BrokerFees.CommissionRate

		// Apply Minimum Fee Logic
		finalCommission := math.Max(commission, c.config.BrokerFees.MinimumFee)

		// Total Transaction Costs for this batch
		totalTransactionCosts := finalCommission + c.config.BrokerFees.FlatFee

		// Total Cash Required
		totalRequired := result.TotalTax + totalTransactionCosts

		// New Shares Needed (Round UP)
		newSharesToSell := math.Ceil(totalRequired / input.SalePrice)

		if newSharesToSell == sharesToSell {
			// Stabilized
			result.SharesToSell = sharesToSell
			result.BrokerCommission = finalCommission
			result.FlatFee = c.config.BrokerFees.FlatFee
			result.TotalFees = totalTransactionCosts
			result.TotalCosts = totalRequired
			result.EstGrossProceeds = grossProceeds
			break
		}

		sharesToSell = newSharesToSell
	}

	// 5. Finalize Results
	result.Residual = (sharesToSell * input.SalePrice) - result.TotalCosts
	result.NetShares = result.SharesReleased - result.SharesToSell

	return result
}
