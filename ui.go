package main

import (
	_ "embed"
	"fmt"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"fynance/stc"
)

// --- TOOL 1: Sell To Cover (Options) ---
func makeSTCTab(win fyne.Window) fyne.CanvasObject {
	// --- INPUT FIELDS ---
	exSharesEntry := newEntry("0", "Exercised Shares")
	exPriceEntry := newEntry("0.00", "Exercise Price")
	fmvEntry := newEntry("0.00", "Fair Market Value")

	fedTaxEntry := newEntry("0.22", "Federal (0.22)")
	medTaxEntry := newEntry("0.0145", "Medicare (0.0145)")
	ssTaxEntry := newEntry("0.062", "Social Security (0.062)")
	stateTaxEntry := newEntry("0.00", "State (e.g. 0.09)")
	localTaxEntry := newEntry("0.00", "Local/SDI (e.g. 0.01)")
	commRateEntry := newEntry("0.03", "Commission Rate (0.03)")
	minFeeEntry := newEntry("25.00", "Min Fee (25.00)")

	// --- OUTPUT LABELS ---
	lblNetShares := canvas.NewText("-", theme.PrimaryColor())
	lblNetShares.TextSize = 24
	lblNetShares.TextStyle = fyne.TextStyle{Bold: true}

	lblResidual := canvas.NewText("-", theme.SuccessColor())
	lblResidual.TextSize = 24
	lblResidual.TextStyle = fyne.TextStyle{Bold: true}

	lblSharesSold := widget.NewLabel("-")
	lblTotalCost := widget.NewLabel("-")
	lblGrossProceeds := widget.NewLabel("-")
	lblTaxes := widget.NewLabel("-")
	lblFees := widget.NewLabel("-")

	// --- LOGIC ---
	calculateFunc := func() {
		exPrice, err1 := parseFloat(exPriceEntry.Text)
		fmv, err3 := parseFloat(fmvEntry.Text)
		exShares, err2 := parseFloat(exSharesEntry.Text)

		fed, _ := parseFloat(fedTaxEntry.Text)
		med, _ := parseFloat(medTaxEntry.Text)
		ss, _ := parseFloat(ssTaxEntry.Text)
		state, _ := parseFloat(stateTaxEntry.Text)
		local, _ := parseFloat(localTaxEntry.Text)

		comm, _ := parseFloat(commRateEntry.Text)
		minFee, _ := parseFloat(minFeeEntry.Text)

		if err1 != nil || err2 != nil || err3 != nil {
			dialog.ShowError(fmt.Errorf("Please enter valid numbers for Price, Shares, and FMV"), win)
			return
		}

		if exPrice <= 0 || exShares <= 0 || fmv <= 0 {
			dialog.ShowError(fmt.Errorf("Price, Shares, and FMV must be greater than 0"), win)
			return
		}

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
		input := stc.Input{
			ExercisePrice:   exPrice,
			ExercisedShares: exShares,
			FMV:             fmv,
		}

		result := calculator.Calculate(input)

		lblNetShares.Text = fmt.Sprintf("%.0f", result.NetShares)
		lblNetShares.Refresh()
		lblResidual.Text = fmt.Sprintf("$%.2f", result.Residual)
		lblResidual.Refresh()
		lblSharesSold.SetText(fmt.Sprintf("%.0f", result.SharesToSell))
		lblTotalCost.SetText(fmt.Sprintf("$%.2f", result.TotalCosts))
		lblGrossProceeds.SetText(fmt.Sprintf("$%.2f", result.EstGrossProceeds))
		lblTaxes.SetText(fmt.Sprintf("$%.2f", result.TotalTax))
		lblFees.SetText(fmt.Sprintf("$%.2f", result.BrokerFees))
	}

	// --- LAYOUT ---
	calcBtn := widget.NewButtonWithIcon("CALCULATE", theme.ConfirmIcon(), calculateFunc)
	calcBtn.Importance = widget.HighImportance

	transForm := widget.NewForm(
		widget.NewFormItem("Exercise Price ($)", exPriceEntry),
		widget.NewFormItem("FMV ($)", fmvEntry),
		widget.NewFormItem("Exercised Shares", exSharesEntry),
	)

	taxForm := widget.NewForm(
		widget.NewFormItem("Federal", fedTaxEntry),
		widget.NewFormItem("Medicare", medTaxEntry),
		widget.NewFormItem("Social Sec", ssTaxEntry),
		widget.NewFormItem("State", stateTaxEntry),
		widget.NewFormItem("Local/SDI", localTaxEntry),
	)

	brokerForm := widget.NewForm(
		widget.NewFormItem("Commission Rate", commRateEntry),
		widget.NewFormItem("Minimum Fee ($)", minFeeEntry),
	)

	inputTabs := container.NewAppTabs(
		container.NewTabItem("Base", transForm),
		container.NewTabItem("Taxes", taxForm),
		container.NewTabItem("Service", brokerForm),
	)

	inputCard := widget.NewCard("Stock Options", "", container.NewVBox(
		inputTabs,
		layout.NewSpacer(),
		calcBtn,
	))

	// Result Layout using Grid
	// Summary in a 2-column top row
	summaryGrid := container.NewGridWithColumns(2,
		container.New(layout.NewFormLayout(), widget.NewLabel("Net Shares:"), lblNetShares),
		container.New(layout.NewFormLayout(), widget.NewLabel("Residual:"), lblResidual),
	)

	// Details in a 2-column grid
	detailsLeft := widget.NewForm(
		widget.NewFormItem("Shares Sold:", lblSharesSold),
		widget.NewFormItem("Gross Proceeds:", lblGrossProceeds),
	)

	detailsRight := widget.NewForm(
		widget.NewFormItem("Total Taxes:", lblTaxes),
		widget.NewFormItem("Broker Fees:", lblFees),
		widget.NewFormItem("Total Costs:", lblTotalCost),
	)

	detailsGrid := container.NewGridWithColumns(2, detailsLeft, detailsRight)

	resultCard := container.NewVBox(
		summaryGrid,
		widget.NewSeparator(),
		detailsGrid,
	)

	content := container.NewVBox(
		inputCard,
		layout.NewSpacer(),
		resultCard,
	)

	return container.NewPadded(content)
}

// --- TOOL 3: RSU Sell To Cover ---
func makeRSUTab(win fyne.Window) fyne.CanvasObject {
	// --- INPUT FIELDS ---
	// RSU Specific Inputs
	sharesReleasedEntry := newEntry("0", "Shares Released")
	vestPriceEntry := newEntry("0.00", "Release (FMV)")
	salePriceEntry := newEntry("0.00", "Est. Sale Price")

	// Tax Inputs (Defaults matching existing)
	fedTaxEntry := newEntry("0.22", "Federal (0.22)")
	medTaxEntry := newEntry("0.0145", "Medicare (0.0145)")
	ssTaxEntry := newEntry("0.062", "Social Security (0.062)")
	stateTaxEntry := newEntry("0.00", "State (e.g. 0.09)")
	localTaxEntry := newEntry("0.00", "Local/SDI (e.g. 0.01)")

	// Broker Inputs
	commRateEntry := newEntry("0.03", "Commission Rate (0.03)")
	minFeeEntry := newEntry("25.00", "Min Fee (25.00)")
	flatFeeEntry := newEntry("0.00", "Processing Fee ($)")

	// --- OUTPUT LABELS ---
	lblNetShares := canvas.NewText("-", theme.PrimaryColor())
	lblNetShares.TextSize = 24
	lblNetShares.TextStyle = fyne.TextStyle{Bold: true}

	lblResidual := canvas.NewText("-", theme.SuccessColor())
	lblResidual.TextSize = 24
	lblResidual.TextStyle = fyne.TextStyle{Bold: true}

	lblSharesSold := widget.NewLabel("-")
	lblTotalCost := widget.NewLabel("-")
	lblGrossProceeds := widget.NewLabel("-")
	lblTaxableGain := widget.NewLabel("-")
	lblTaxes := widget.NewLabel("-")
	lblFees := widget.NewLabel("-")

	// --- LOGIC ---
	calculateFunc := func() {
		sharesReleased, err1 := parseFloat(sharesReleasedEntry.Text)
		vestPrice, err2 := parseFloat(vestPriceEntry.Text)
		salePrice, err3 := parseFloat(salePriceEntry.Text)

		fed, _ := parseFloat(fedTaxEntry.Text)
		med, _ := parseFloat(medTaxEntry.Text)
		ss, _ := parseFloat(ssTaxEntry.Text)
		state, _ := parseFloat(stateTaxEntry.Text)
		local, _ := parseFloat(localTaxEntry.Text)

		comm, _ := parseFloat(commRateEntry.Text)
		minFee, _ := parseFloat(minFeeEntry.Text)
		flatFee, _ := parseFloat(flatFeeEntry.Text)

		if err1 != nil || err2 != nil || err3 != nil {
			dialog.ShowError(fmt.Errorf("Please enter valid numbers"), win)
			return
		}

		if sharesReleased <= 0 || vestPrice <= 0 || salePrice <= 0 {
			dialog.ShowError(fmt.Errorf("Shares and Prices must be greater than 0"), win)
			return
		}

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
				FlatFee:        flatFee,
			},
		}

		calculator := stc.NewCalculator(config)
		input := stc.RSUInput{
			SharesReleased: sharesReleased,
			VestPrice:      vestPrice,
			SalePrice:      salePrice,
		}

		result := calculator.CalculateRSU(input)

		lblNetShares.Text = fmt.Sprintf("%.0f", result.NetShares)
		lblNetShares.Refresh()
		lblResidual.Text = fmt.Sprintf("$%.2f", result.Residual)
		lblResidual.Refresh()

		lblSharesSold.SetText(fmt.Sprintf("%.0f", result.SharesToSell))
		lblTotalCost.SetText(fmt.Sprintf("$%.2f", result.TotalCosts))
		lblGrossProceeds.SetText(fmt.Sprintf("$%.2f", result.EstGrossProceeds))
		lblTaxableGain.SetText(fmt.Sprintf("$%.2f", result.TaxableGain))
		lblTaxes.SetText(fmt.Sprintf("$%.2f", result.TotalTax))
		lblFees.SetText(fmt.Sprintf("$%.2f", result.TotalFees))
	}

	// --- LAYOUT ---
	calcBtn := widget.NewButtonWithIcon("CALCULATE", theme.ConfirmIcon(), calculateFunc)
	calcBtn.Importance = widget.HighImportance

	rsuForm := widget.NewForm(
		widget.NewFormItem("Shares Released", sharesReleasedEntry),
		widget.NewFormItem("Vest Price (FMV) $", vestPriceEntry),
		widget.NewFormItem("Est. Sale Price $", salePriceEntry),
	)

	taxForm := widget.NewForm(
		widget.NewFormItem("Federal", fedTaxEntry),
		widget.NewFormItem("Medicare", medTaxEntry),
		widget.NewFormItem("Social Sec", ssTaxEntry),
		widget.NewFormItem("State", stateTaxEntry),
		widget.NewFormItem("Local/SDI", localTaxEntry),
	)

	brokerForm := widget.NewForm(
		widget.NewFormItem("Commission Rate", commRateEntry),
		widget.NewFormItem("Minimum Fee ($)", minFeeEntry),
		widget.NewFormItem("Processing Fee ($)", flatFeeEntry),
	)

	inputTabs := container.NewAppTabs(
		container.NewTabItem("Equity", rsuForm),
		container.NewTabItem("Taxes", taxForm),
		container.NewTabItem("Service", brokerForm),
	)

	inputCard := widget.NewCard("Restricted Stock", "", container.NewVBox(
		inputTabs,
		layout.NewSpacer(),
		calcBtn,
	))

	// Result Layout using Grid
	summaryGrid := container.NewGridWithColumns(2,
		container.New(layout.NewFormLayout(), widget.NewLabel("Net Shares:"), lblNetShares),
		container.New(layout.NewFormLayout(), widget.NewLabel("Residual:"), lblResidual),
	)

	detailsLeft := widget.NewForm(
		widget.NewFormItem("Taxable Gain:", lblTaxableGain),
		widget.NewFormItem("Shares Sold:", lblSharesSold),
		widget.NewFormItem("Gross Proceeds:", lblGrossProceeds),
	)

	detailsRight := widget.NewForm(
		widget.NewFormItem("Total Taxes:", lblTaxes),
		widget.NewFormItem("Total Fees:", lblFees),
		widget.NewFormItem("Total Costs:", lblTotalCost),
	)

	detailsGrid := container.NewGridWithColumns(2, detailsLeft, detailsRight)

	resultCard := container.NewVBox(
		summaryGrid,
		widget.NewSeparator(),
		detailsGrid,
	)

	content := container.NewVBox(
		inputCard,
		layout.NewSpacer(),
		resultCard,
	)

	return container.NewPadded(content)
}

// --- TOOL 2: Standard Calculator ---

func makeCalculatorTab() fyne.CanvasObject {
	// State
	var currentNum string = "0"
	var storedNum float64 = 0
	var currentOp string = ""
	var lastWasOp bool = false

	// Widgets
	display := widget.NewLabel("0")
	display.TextStyle = fyne.TextStyle{Monospace: true, Bold: true}
	display.Alignment = fyne.TextAlignTrailing

	history := widget.NewLabel("")
	history.TextStyle = fyne.TextStyle{Monospace: true}
	history.Alignment = fyne.TextAlignTrailing
	// Logic Helper
	updateDisplay := func() {
		display.SetText(currentNum)
	}

	calculate := func() {
		if currentOp == "" {
			return
		}
		val, _ := strconv.ParseFloat(currentNum, 64)
		var result float64

		switch currentOp {
		case "+":
			result = storedNum + val
		case "-":
			result = storedNum - val
		case "*":
			result = storedNum * val
		case "/":
			if val == 0 {
				currentNum = "Error"
				updateDisplay()
				return
			}
			result = storedNum / val
		}

		// Format result to avoid hanging decimals if whole number
		if result == float64(int64(result)) {
			currentNum = fmt.Sprintf("%.0f", result)
		} else {
			currentNum = fmt.Sprintf("%g", result)
		}
		currentOp = ""
		updateDisplay()
	}

	// Button Actions
	onNum := func(n string) {
		if lastWasOp {
			currentNum = "0"
			lastWasOp = false
		}
		if currentNum == "0" && n != "." {
			currentNum = n
		} else {
			if n == "." && strings.Contains(currentNum, ".") {
				return
			}
			currentNum += n
		}
		updateDisplay()
	}

	onOp := func(op string) {
		if currentOp != "" && !lastWasOp {
			calculate()
		}
		var err error
		storedNum, err = strconv.ParseFloat(currentNum, 64)
		if err != nil {
			currentNum = "0"
		}
		currentOp = op
		lastWasOp = true
		history.SetText(fmt.Sprintf("%g %s", storedNum, currentOp))
	}

	onClear := func() {
		currentNum = "0"
		storedNum = 0
		currentOp = ""
		lastWasOp = false
		history.SetText("")
		updateDisplay()
	}

	onEq := func() {
		history.SetText("")
		calculate()
		lastWasOp = true // Treat result as a starting point
	}

	// Layout Construction
	// Helper to make buttons uniform
	btn := func(label string, action func()) *widget.Button {
		b := widget.NewButton(label, action)
		return b
	}

	// Number pad
	grid := container.NewGridWithColumns(4,
		// Row 1
		btn("C", onClear), btn("", func() {}), btn("", func() {}), btn("/", func() { onOp("/") }),
		// Row 2
		btn("7", func() { onNum("7") }), btn("8", func() { onNum("8") }), btn("9", func() { onNum("9") }), btn("*", func() { onOp("*") }),
		// Row 3
		btn("4", func() { onNum("4") }), btn("5", func() { onNum("5") }), btn("6", func() { onNum("6") }), btn("-", func() { onOp("-") }),
		// Row 4
		btn("1", func() { onNum("1") }), btn("2", func() { onNum("2") }), btn("3", func() { onNum("3") }), btn("+", func() { onOp("+") }),
		// Row 5
		btn("0", func() { onNum("0") }), btn(".", func() { onNum(".") }), btn("", func() {}), widget.NewButtonWithIcon("=", theme.ConfirmIcon(), onEq),
	)

	// Make operators visually distinct (optional, depends on theme but HighImportance helps)
	for _, o := range grid.Objects {
		if b, ok := o.(*widget.Button); ok {
			if b.Text == "=" {
				b.Importance = widget.HighImportance
			}
		}
	}

	screen := container.NewBorder(
		container.NewVBox(
			widget.NewCard("", "", container.NewVBox(history, display)),
			widget.NewSeparator(),
		),
		nil, nil, nil,
		container.NewPadded(grid),
	)

	return screen
}

// --- SHARED HELPERS ---
func newEntry(placeholder, label string) *widget.Entry {
	e := widget.NewEntry()
	e.SetPlaceHolder(placeholder)
	e.Text = placeholder
	return e
}

func parseFloat(s string) (float64, error) {
	if s == "" {
		return 0, nil
	}
	return strconv.ParseFloat(s, 64)
}
