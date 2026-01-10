package main

import (
	"fmt"
	"strconv"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"stc2go/stc" // Ensure this matches your module name (e.g., go.mod module name)
)

func main() {
	myApp := app.New()
	myWindow := myApp.NewWindow("EXERCISE")
	myWindow.Resize(fyne.NewSize(500, 700))
	myApp.Settings().SetTheme(newCustomTheme())

	// --- INPUT FIELDS ---

	// Transaction Inputs
	exSharesEntry := newEntry("0", "Exercised Shares")
	exPriceEntry := newEntry("0.00", "Exercise Price")
	fmvEntry := newEntry("0.00", "Fair Market Value")

	// Tax Inputs (Defaults pre-filled based on your previous config)
	fedTaxEntry := newEntry("0.22", "Federal Rate (0.22)")
	medTaxEntry := newEntry("0.0145", "Medicare Rate (0.0145)")
	ssTaxEntry := newEntry("0.062", "Social Security Rate (0.062)")
	stateTaxEntry := newEntry("0.00", "State Rate (e.g. 0.09)")
	localTaxEntry := newEntry("0.00", "Local/SDI Rate")

	// Broker Inputs
	commRateEntry := newEntry("0.03", "Commission Rate (0.03)")
	minFeeEntry := newEntry("25.00", "Min Fee (25.00)")

	// --- OUTPUT LABELS ---

	// Summary Labels (Big text)
	lblNetShares := canvas.NewText("-", theme.PrimaryColor())
	lblNetShares.TextSize = 24
	lblNetShares.TextStyle = fyne.TextStyle{Bold: true}

	lblResidual := canvas.NewText("-", theme.SuccessColor())
	lblResidual.TextSize = 24
	lblResidual.TextStyle = fyne.TextStyle{Bold: true}

	// Detailed Labels
	lblSharesSold := widget.NewLabel("-")
	lblTotalCost := widget.NewLabel("-")
	lblGrossProceeds := widget.NewLabel("-")
	lblTaxes := widget.NewLabel("-")
	lblFees := widget.NewLabel("-")

	// --- LOGIC ---

	calculateFunc := func() {
		// 1. Parse Inputs
		exPrice, err1 := parseFloat(exPriceEntry.Text)
		exShares, err2 := parseFloat(exSharesEntry.Text)
		fmv, err3 := parseFloat(fmvEntry.Text)

		// Parse Taxes
		fed, _ := parseFloat(fedTaxEntry.Text)
		med, _ := parseFloat(medTaxEntry.Text)
		ss, _ := parseFloat(ssTaxEntry.Text)
		state, _ := parseFloat(stateTaxEntry.Text)
		local, _ := parseFloat(localTaxEntry.Text)

		// Parse Fees
		comm, _ := parseFloat(commRateEntry.Text)
		minFee, _ := parseFloat(minFeeEntry.Text)

		if err1 != nil || err2 != nil || err3 != nil {
			dialog.ShowError(fmt.Errorf("Please enter valid numbers for Price, Shares, and FMV"), myWindow)
			return
		}

		if exPrice <= 0 || exShares <= 0 || fmv <= 0 {
			dialog.ShowError(fmt.Errorf("Price, Shares, and FMV must be greater than 0"), myWindow)
			return
		}

		// 2. Configure Calculator
		config := stc.Config{
			TaxRates: stc.TaxRates{
				Federal:   fed,
				Medicare:  med,
				SocialSec: ss,
				State:     state,
				LocalSDI:  local,
			},
			BrokerFees: stc.BrokerFees{
				CommissionRate: comm,
				MinimumFee:     minFee,
			},
		}

		calculator := stc.NewCalculator(config)

		// 3. Calculate
		input := stc.Input{
			ExercisePrice:   exPrice,
			ExercisedShares: exShares,
			FMV:             fmv,
		}

		result := calculator.Calculate(input)

		// 4. Update UI
		// Summary Update (Whole Numbers for Shares!)
		lblNetShares.Text = fmt.Sprintf("%.0f", result.NetShares)
		lblNetShares.Refresh()

		lblResidual.Text = fmt.Sprintf("$%.2f", result.Residual)
		lblResidual.Refresh()

		// Detail Update
		lblSharesSold.SetText(fmt.Sprintf("%.0f", result.SharesToSell))
		lblTotalCost.SetText(fmt.Sprintf("$%.2f", result.TotalCosts))
		lblGrossProceeds.SetText(fmt.Sprintf("$%.2f", result.EstGrossProceeds))
		lblTaxes.SetText(fmt.Sprintf("$%.2f", result.TotalTax))
		lblFees.SetText(fmt.Sprintf("$%.2f", result.BrokerFees))
	}

	// --- LAYOUT CONSTRUCTION ---

	// Button
	calcBtn := widget.NewButtonWithIcon("SEND IT", theme.ConfirmIcon(), calculateFunc)
	calcBtn.Importance = widget.HighImportance

	// Input Forms
	transForm := widget.NewForm(
		widget.NewFormItem("Exercise Price ($)", exPriceEntry),
		widget.NewFormItem("Exercised Shares", exSharesEntry),
		widget.NewFormItem("FMV ($)", fmvEntry),
	)

	taxForm := widget.NewForm(
		widget.NewFormItem("Federal Rate", fedTaxEntry),
		widget.NewFormItem("Medicare Rate", medTaxEntry),
		widget.NewFormItem("Social Sec Rate", ssTaxEntry),
		widget.NewFormItem("State Rate", stateTaxEntry),
		widget.NewFormItem("Local/SDI Rate", localTaxEntry),
	)

	brokerForm := widget.NewForm(
		widget.NewFormItem("Commission Rate", commRateEntry),
		widget.NewFormItem("Minimum Fee ($)", minFeeEntry),
	)

	// Input Card
	inputTabs := container.NewAppTabs(
		container.NewTabItem("Base", transForm),
		container.NewTabItem("Taxes", taxForm),
		container.NewTabItem("Service", brokerForm),
	)

	inputCard := widget.NewCard("Configuration", "", container.NewVBox(
		inputTabs,
		layout.NewSpacer(),
		calcBtn,
	))

	// Result Card Construction

	// Helper for Result Grid
	resultRow := func(label string, valueObj fyne.CanvasObject) *fyne.Container {
		return container.New(layout.NewFormLayout(), widget.NewLabel(label), valueObj)
	}

	summaryContainer := container.NewVBox(
		resultRow("Net Shares (To Keep):", lblNetShares),
		resultRow("Residual Cash (Check):", lblResidual),
	)

	detailsContainer := container.NewVBox(
		widget.NewSeparator(),
		resultRow("Shares Sold to Cover:", lblSharesSold),
		resultRow("Gross Proceeds:", lblGrossProceeds),
		widget.NewSeparator(),
		resultRow("Total Taxes:", lblTaxes),
		resultRow("Broker Fees:", lblFees),
		resultRow("Total Costs:", lblTotalCost),
	)

	resultCard := widget.NewCard("Calculation Results", "Sell-To-Cover Breakdown",
		container.NewVBox(
			summaryContainer,
			detailsContainer,
		),
	)

	// Main Scroll Container
	content := container.NewVBox(
		inputCard,
		layout.NewSpacer(),
		resultCard,
		layout.NewSpacer(),
		widget.NewRichTextFromMarkdown("*Generated by STC2GO*"),
	)

	myWindow.SetContent(container.NewPadded(content))
	myWindow.ShowAndRun()
}

// --- HELPERS ---

func newEntry(placeholder, label string) *widget.Entry {
	e := widget.NewEntry()
	e.SetPlaceHolder(placeholder)
	e.Text = placeholder // Pre-fill default
	return e
}

func parseFloat(s string) (float64, error) {
	if s == "" {
		return 0, nil
	}
	return strconv.ParseFloat(s, 64)
}
