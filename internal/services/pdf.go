package services

import (
	"bytes"
	"fmt"
	"time"

	"github.com/go-pdf/fpdf"

	"carhistory/internal/models"
)

// VehicleReportData holds everything needed to build a vehicle service history PDF.
type VehicleReportData struct {
	Vehicle            models.Vehicle
	ServiceRecords     []models.ServiceRecord
	MaintenanceRecords []models.MaintenanceRecord
	InsuranceRecords   []models.InsuranceRecord
	PUCRecords         []models.PUCRecord
	Expenses           []models.Expense
	FuelLogs           []models.FuelLog
	GeneratedAt        time.Time
}

// GenerateVehicleReportPDF creates a formatted PDF report and returns its bytes.
func GenerateVehicleReportPDF(data *VehicleReportData) ([]byte, error) {
	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.SetAutoPageBreak(true, 20)
	pdf.SetMargins(15, 15, 15)

	pdf.SetHeaderFunc(func() {
		header(pdf)
	})
	pdf.SetFooterFunc(func() {
		footer(pdf, data.GeneratedAt)
	})

	pdf.AddPage()

	// Watermark behind content.
	watermark(pdf)

	// Main content.
	vehicleInfoBox(pdf, data.Vehicle)
	summaryBoxes(pdf, data)

	serviceSection(pdf, data.ServiceRecords)
	maintenanceSection(pdf, data.MaintenanceRecords)
	insuranceSection(pdf, data.InsuranceRecords)
	pucSection(pdf, data.PUCRecords)
	expenseSection(pdf, data.Expenses)

	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func header(pdf *fpdf.Fpdf) {
	pdf.SetFillColor(25, 55, 109)
	pdf.Rect(0, 0, 210, 30, "F")

	pdf.SetTextColor(255, 255, 255)
	pdf.SetFont("Arial", "B", 20)
	pdf.SetXY(15, 10)
	pdf.Cell(0, 8, "CarHistory")

	pdf.SetFont("Arial", "", 11)
	pdf.SetXY(15, 19)
	pdf.Cell(0, 6, "Vehicle Service History Report")

	pdf.SetTextColor(0, 0, 0)
	pdf.SetY(32)
}

func footer(pdf *fpdf.Fpdf, generatedAt time.Time) {
	pdf.SetY(-15)
	pdf.SetFont("Arial", "I", 9)
	pdf.SetTextColor(128, 128, 128)
	pdf.Cell(0, 10, fmt.Sprintf("Generated on %s | Page %d", generatedAt.Format("2006-01-02"), pdf.PageNo()))
}

func watermark(pdf *fpdf.Fpdf) {
	pdf.SetTextColor(230, 230, 230)
	pdf.SetFont("Arial", "B", 55)
	pdf.TransformBegin()
	pdf.TransformRotate(45, 105, 148)
	pdf.Text(25, 150, "CARHISTORY")
	pdf.TransformEnd()
	pdf.SetTextColor(0, 0, 0)
}

func vehicleInfoBox(pdf *fpdf.Fpdf, v models.Vehicle) {
	pdf.SetFont("Arial", "B", 13)
	pdf.SetTextColor(25, 55, 109)
	pdf.Cell(0, 8, "Vehicle Details")
	pdf.Ln(8)

	pdf.SetDrawColor(200, 200, 200)
	pdf.SetFillColor(245, 247, 250)
	pdf.Rect(15, pdf.GetY(), 180, 28, "FD")

	pdf.SetFont("Arial", "B", 11)
	pdf.SetTextColor(33, 37, 41)
	pdf.SetXY(18, pdf.GetY()+3)

	brand := "-"
	if v.Brand != nil {
		brand = *v.Brand
	}
	model := "-"
	if v.Model != nil {
		model = *v.Model
	}
	pdf.Cell(0, 6, fmt.Sprintf("%s %s", brand, model))
	pdf.Ln(6)

	pdf.SetFont("Arial", "", 10)
	pdf.SetTextColor(80, 80, 80)
	pdf.SetX(18)

	reg := "-"
	if v.RegistrationNumber != nil {
		reg = *v.RegistrationNumber
	}
	fuel := "-"
	if v.FuelType != nil {
		fuel = *v.FuelType
	}
	year := "-"
	if v.ManufacturingYear != nil {
		year = fmt.Sprintf("%d", *v.ManufacturingYear)
	}

	pdf.Cell(55, 6, fmt.Sprintf("Registration: %s", reg))
	pdf.Cell(55, 6, fmt.Sprintf("Fuel: %s", fuel))
	pdf.Cell(0, 6, fmt.Sprintf("Year: %s", year))
	pdf.Ln(6)

	pdf.SetX(18)
	pdf.Cell(55, 6, fmt.Sprintf("Odometer: %.0f km", v.CurrentOdometer))
	pdf.Cell(55, 6, fmt.Sprintf("Type: %s", v.VehicleTypeName))
	pdf.Ln(12)
}

func summaryBoxes(pdf *fpdf.Fpdf, data *VehicleReportData) {
	serviceCost := 0.0
	for _, s := range data.ServiceRecords {
		serviceCost += s.TotalCost
	}
	maintCost := 0.0
	for _, m := range data.MaintenanceRecords {
		if m.Cost != nil {
			maintCost += *m.Cost
		}
	}
	expenseCost := 0.0
	for _, e := range data.Expenses {
		expenseCost += e.Amount
	}

	boxes := []struct {
		label string
		value string
	}{
		{"Service Records", fmt.Sprintf("%d", len(data.ServiceRecords))},
		{"Maintenance Records", fmt.Sprintf("%d", len(data.MaintenanceRecords))},
		{"Total Service Cost", fmt.Sprintf("%.2f", serviceCost)},
		{"Total Maintenance Cost", fmt.Sprintf("%.2f", maintCost)},
		{"Other Expenses", fmt.Sprintf("%.2f", expenseCost)},
	}

	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(25, 55, 109)
	pdf.Cell(0, 8, "Summary")
	pdf.Ln(8)

	boxW := 34.0
	boxH := 18.0
	gap := 3.5
	startX := 15.0

	for i, b := range boxes {
		x := startX + float64(i)*(boxW+gap)
		pdf.SetFillColor(232, 240, 254)
		pdf.SetDrawColor(200, 216, 240)
		pdf.Rect(x, pdf.GetY(), boxW, boxH, "FD")

		pdf.SetFont("Arial", "B", 11)
		pdf.SetTextColor(25, 55, 109)
		pdf.SetXY(x+2, pdf.GetY()+2)
		pdf.Cell(boxW-4, 6, b.value)

		pdf.SetFont("Arial", "", 8)
		pdf.SetTextColor(80, 80, 80)
		pdf.SetXY(x+2, pdf.GetY()+7)
		pdf.Cell(boxW-4, 5, b.label)
	}
	pdf.Ln(boxH + 6)
}

func serviceSection(pdf *fpdf.Fpdf, records []models.ServiceRecord) {
	sectionTitle(pdf, "Service History")

	if len(records) == 0 {
		noDataRow(pdf, "No service records found.")
		return
	}

	headers := []string{"Date", "Centre", "Odometer", "Cost"}
	widths := []float64{35, 80, 30, 35}
	tableHeader(pdf, headers, widths)

	for _, r := range records {
		odometer := "-"
		if r.Odometer != nil {
			odometer = fmt.Sprintf("%.0f", *r.Odometer)
		}
		centre := "-"
		if r.ServiceCentre != nil {
			centre = *r.ServiceCentre
		}
		cells := []string{
			r.ServiceDate.Format("2006-01-02"),
			centre,
			odometer,
			fmt.Sprintf("%.2f", r.TotalCost),
		}
		tableRow(pdf, cells, widths, false)
	}
	pdf.Ln(4)
}

func maintenanceSection(pdf *fpdf.Fpdf, records []models.MaintenanceRecord) {
	sectionTitle(pdf, "Maintenance History")

	if len(records) == 0 {
		noDataRow(pdf, "No maintenance records found.")
		return
	}

	headers := []string{"Date", "Type", "Title", "Cost"}
	widths := []float64{35, 45, 65, 35}
	tableHeader(pdf, headers, widths)

	for _, r := range records {
		cost := "-"
		if r.Cost != nil {
			cost = fmt.Sprintf("%.2f", *r.Cost)
		}
		cells := []string{
			r.DatePerformed.Format("2006-01-02"),
			r.MaintenanceTypeName,
			r.Title,
			cost,
		}
		tableRow(pdf, cells, widths, false)
	}
	pdf.Ln(4)
}

func insuranceSection(pdf *fpdf.Fpdf, records []models.InsuranceRecord) {
	sectionTitle(pdf, "Insurance History")

	if len(records) == 0 {
		noDataRow(pdf, "No insurance records found.")
		return
	}

	headers := []string{"Company", "Policy", "Start", "Expiry", "Premium"}
	widths := []float64{45, 45, 30, 30, 30}
	tableHeader(pdf, headers, widths)

	for _, r := range records {
		policy := "-"
		if r.PolicyNumber != nil {
			policy = *r.PolicyNumber
		}
		premium := "-"
		if r.Premium != nil {
			premium = fmt.Sprintf("%.2f", *r.Premium)
		}
		cells := []string{
			r.InsuranceCompany,
			policy,
			r.StartDate.Format("2006-01-02"),
			r.ExpiryDate.Format("2006-01-02"),
			premium,
		}
		tableRow(pdf, cells, widths, false)
	}
	pdf.Ln(4)
}

func pucSection(pdf *fpdf.Fpdf, records []models.PUCRecord) {
	sectionTitle(pdf, "PUC History")

	if len(records) == 0 {
		noDataRow(pdf, "No PUC records found.")
		return
	}

	headers := []string{"Certificate", "Test Date", "Expiry", "Centre"}
	widths := []float64{50, 35, 35, 60}
	tableHeader(pdf, headers, widths)

	for _, r := range records {
		cert := "-"
		if r.CertificateNumber != nil {
			cert = *r.CertificateNumber
		}
		centre := "-"
		if r.TestingCentre != nil {
			centre = *r.TestingCentre
		}
		cells := []string{
			cert,
			r.TestDate.Format("2006-01-02"),
			r.ExpiryDate.Format("2006-01-02"),
			centre,
		}
		tableRow(pdf, cells, widths, false)
	}
	pdf.Ln(4)
}

func expenseSection(pdf *fpdf.Fpdf, records []models.Expense) {
	sectionTitle(pdf, "Recent Expenses")

	if len(records) == 0 {
		noDataRow(pdf, "No expenses found.")
		return
	}

	headers := []string{"Date", "Category", "Description", "Amount"}
	widths := []float64{35, 35, 75, 35}
	tableHeader(pdf, headers, widths)

	for _, e := range records {
		desc := "-"
		if e.Description != nil {
			desc = *e.Description
		}
		cells := []string{
			e.ExpenseDate.Format("2006-01-02"),
			e.Category,
			desc,
			fmt.Sprintf("%.2f", e.Amount),
		}
		tableRow(pdf, cells, widths, false)
	}
	pdf.Ln(4)
}

func sectionTitle(pdf *fpdf.Fpdf, title string) {
	pdf.SetFont("Arial", "B", 12)
	pdf.SetTextColor(25, 55, 109)
	pdf.Cell(0, 8, title)
	pdf.Ln(6)

	pdf.SetDrawColor(25, 55, 109)
	pdf.Line(15, pdf.GetY(), 195, pdf.GetY())
	pdf.Ln(3)
}

func noDataRow(pdf *fpdf.Fpdf, message string) {
	pdf.SetFont("Arial", "I", 10)
	pdf.SetTextColor(128, 128, 128)
	pdf.Cell(0, 6, message)
	pdf.Ln(10)
}

func tableHeader(pdf *fpdf.Fpdf, headers []string, widths []float64) {
	pdf.SetFont("Arial", "B", 9)
	pdf.SetFillColor(25, 55, 109)
	pdf.SetTextColor(255, 255, 255)

	for i, h := range headers {
		pdf.CellFormat(widths[i], 7, h, "1", 0, "C", true, 0, "")
	}
	pdf.Ln(7)
}

func tableRow(pdf *fpdf.Fpdf, cells []string, widths []float64, fill bool) {
	pdf.SetFont("Arial", "", 9)
	pdf.SetTextColor(33, 37, 41)
	pdf.SetDrawColor(100, 100, 100)
	if fill {
		pdf.SetFillColor(245, 247, 250)
	} else {
		pdf.SetFillColor(255, 255, 255)
	}

	lineHeight := 6.0
	padding := 1.0
	startX := pdf.GetX()
	startY := pdf.GetY()

	// Calculate row height from the tallest wrapped cell.
	maxLines := 1
	for i, cell := range cells {
		lines := pdf.SplitLines([]byte(cell), widths[i]-2*padding)
		if len(lines) > maxLines {
			maxLines = len(lines)
		}
	}
	rowH := float64(maxLines) * lineHeight

	// Draw each cell with its own border and background.
	for i, cell := range cells {
		x := startX
		for j := 0; j < i; j++ {
			x += widths[j]
		}

		pdf.Rect(x, startY, widths[i], rowH, "FD")
		pdf.SetXY(x+padding, startY+padding)
		pdf.MultiCell(widths[i]-2*padding, lineHeight, cell, "0", "L", false)
	}

	pdf.SetXY(startX, startY+rowH)
}
