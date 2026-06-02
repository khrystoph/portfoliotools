package pkg

import (
	"fmt"
	"sort"
	"time"

	"github.com/xuri/excelize/v2"
)

func GenerateStockReportXLSX(data map[string]CondensedRangesJSON, outputPath string) error {
	// --- Sort tickers ---
	tickers := make([]string, 0, len(data))
	for t := range data {
		tickers = append(tickers, t)
	}
	sort.Strings(tickers)

	// --- Create workbook ---
	f := excelize.NewFile()
	sheet := "Stock Report"
	index, err := f.NewSheet(sheet)
	if err != nil {
		fmt.Println(err, "Excelizer.go L22. Could not create new excel sheet")
	}
	f.SetActiveSheet(index)

	// --- Title and subtitle ---
	title := "📊 Stock Range Report"
	subtitle := fmt.Sprintf("Generated on %s", time.Now().Format("January 2, 2006"))

	f.MergeCell(sheet, "A1", "M1")
	f.MergeCell(sheet, "A2", "M2")

	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Size:  18,
			Color: "#004B87",
		},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	subtitleStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Size:  11,
			Color: "#333333",
		},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})

	f.SetCellValue(sheet, "A1", title)
	f.SetCellStyle(sheet, "A1", "A1", titleStyle)
	f.SetCellValue(sheet, "A2", subtitle)
	f.SetCellStyle(sheet, "A2", "A2", subtitleStyle)

	// --- Table header ---
	headers := []string{
		"Ticker", "Close", "Avg Vol Ratio", "RVol %",
		"VAPR Low", "VAPR High", "Trade Slope", "Trend Slope", "Tail Slope",
		"Trade Direction", "Trend Direction", "Tail Direction", "Timestamp",
	}
	headerRow := 4

	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:  true,
			Color: "#FFFFFF",
		},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#1F4E79"}, Pattern: 1},
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "808080", Style: 1},
			{Type: "top", Color: "808080", Style: 1},
			{Type: "right", Color: "808080", Style: 1},
			{Type: "bottom", Color: "808080", Style: 1},
		},
	})

	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, headerRow)
		f.SetCellValue(sheet, cell, h)
		f.SetCellStyle(sheet, cell, cell, headerStyle)
	}

	// --- Table body ---
	rowStart := headerRow + 1
	altColor := "#F2F2F2"

	bodyStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Border: []excelize.Border{
			{Type: "left", Color: "808080", Style: 1},
			{Type: "top", Color: "808080", Style: 1},
			{Type: "right", Color: "808080", Style: 1},
			{Type: "bottom", Color: "808080", Style: 1},
		},
	})
	altBodyStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{altColor}, Pattern: 1},
		Border: []excelize.Border{
			{Type: "left", Color: "808080", Style: 1},
			{Type: "top", Color: "808080", Style: 1},
			{Type: "right", Color: "808080", Style: 1},
			{Type: "bottom", Color: "808080", Style: 1},
		},
	})

	rowIndex := rowStart
	for _, ticker := range tickers {
		s := data[ticker]
		row := []interface{}{
			ticker,
			fmt.Sprintf("%.2f", s.Close),
			fmt.Sprintf("%.2f", s.AvgVolRatio),
			fmt.Sprintf("%.2f", s.RVolPercent),
			fmt.Sprintf("%.2f", s.RiskRangeLow),
			fmt.Sprintf("%.2f", s.RiskRangeHigh),
			fmt.Sprintf("%.2f", s.TradeSlope),
			fmt.Sprintf("%.2f", s.TrendSlope),
			fmt.Sprintf("%.2f", s.TailSlope),
			s.TradeDirection,
			s.TrendDirection,
			s.TailDirection,
			s.Timestamp.Format("2006-01-02"),
		}

		for i, val := range row {
			cell, _ := excelize.CoordinatesToCellName(i+1, rowIndex)
			f.SetCellValue(sheet, cell, val)
			if (rowIndex-rowStart)%2 == 0 {
				f.SetCellStyle(sheet, cell, cell, altBodyStyle)
			} else {
				f.SetCellStyle(sheet, cell, cell, bodyStyle)
			}
		}
		rowIndex++
	}

	// --- Freeze header row ---
	f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		YSplit:      headerRow,
		TopLeftCell: "A5",
		ActivePane:  "bottomLeft",
	})

	// --- Footer ---
	footerRow := rowIndex + 1
	f.MergeCell(sheet, fmt.Sprintf("A%d", footerRow), fmt.Sprintf("M%d", footerRow))
	f.SetCellValue(sheet, fmt.Sprintf("A%d", footerRow), "Report generated automatically from stock range data.")
	footerStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Font:      &excelize.Font{Italic: true, Size: 10},
	})
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", footerRow), fmt.Sprintf("M%d", footerRow), footerStyle)

	// --- Column widths ---
	widths := []float64{12, 10, 12, 14, 10, 10, 12, 12, 10, 12, 18}
	for i, w := range widths {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheet, col, col, w)
	}

	// --- Save file ---
	if err := f.SaveAs(outputPath); err != nil {
		return fmt.Errorf("failed to save XLSX: %v", err)
	}

	fmt.Printf("✅ XLSX generated successfully at: %s\n", outputPath)
	return nil
}
