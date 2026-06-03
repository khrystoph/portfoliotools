package pkg

import (
	"fmt"
	"sort"
	"time"

	"github.com/xuri/excelize/v2"
)

// GenerateStockReportXLSX writes the stock report to an Excel file.
// showTail adds Tail Slope and Tail Dir columns (13 cols total vs default 11).
func GenerateStockReportXLSX(data map[string]CondensedRangesJSON, outputPath string, showTail bool) error {
	tickers := make([]string, 0, len(data))
	for t := range data {
		tickers = append(tickers, t)
	}
	sort.Strings(tickers)

	f := excelize.NewFile()
	sheet := "Stock Report"
	index, err := f.NewSheet(sheet)
	if err != nil {
		fmt.Println(err, "Excelizer.go: Could not create new excel sheet")
	}
	f.SetActiveSheet(index)

	headers := []string{
		"Ticker", "Close", "Avg Vol Ratio", "RVol %",
		"VAPR Low", "VAPR High", "Trade Slope", "Trend Slope",
	}
	if showTail {
		headers = append(headers, "Tail Slope")
	}
	headers = append(headers, "Trade Dir", "Trend Dir")
	if showTail {
		headers = append(headers, "Tail Dir")
	}
	headers = append(headers, "Timestamp")

	lastCol, _ := excelize.ColumnNumberToName(len(headers))

	title := "Stock Range Report"
	subtitle := fmt.Sprintf("Generated on %s", time.Now().Format("January 2, 2006"))
	f.MergeCell(sheet, "A1", lastCol+"1")
	f.MergeCell(sheet, "A2", lastCol+"2")

	titleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Size: 18, Color: "#004B87"},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	subtitleStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Size: 11, Color: "#333333"},
		Alignment: &excelize.Alignment{Horizontal: "center"},
	})
	f.SetCellValue(sheet, "A1", title)
	f.SetCellStyle(sheet, "A1", "A1", titleStyle)
	f.SetCellValue(sheet, "A2", subtitle)
	f.SetCellStyle(sheet, "A2", "A2", subtitleStyle)

	headerRow := 4
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font:      &excelize.Font{Bold: true, Color: "#FFFFFF"},
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

	border := []excelize.Border{
		{Type: "left", Color: "808080", Style: 1},
		{Type: "top", Color: "808080", Style: 1},
		{Type: "right", Color: "808080", Style: 1},
		{Type: "bottom", Color: "808080", Style: 1},
	}

	bullishStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#70AD47"}},
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Border:    border,
	})
	bearishStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#FF0000"}},
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Border:    border,
	})
	neutralStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"#A5A5A5"}},
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Border:    border,
	})
	indeterminateStyle, _ := f.NewStyle(&excelize.Style{
		Fill:      excelize.Fill{Type: "pattern", Pattern: 7, Color: []string{"#FF0000", "#FFFFFF"}},
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Border:    border,
	})

	bodyStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Border:    border,
	})
	altBodyStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Fill:      excelize.Fill{Type: "pattern", Color: []string{"#F2F2F2"}, Pattern: 1},
		Border:    border,
	})

	// Direction column indices (0-based).
	// Default (no tail): 0:Ticker 1:Close 2:AvgVolRatio 3:RVol% 4:VAPRLow 5:VAPRHigh
	//   6:TradeSlope 7:TrendSlope 8:TradeDir 9:TrendDir 10:Timestamp
	// With tail: same through 7:TrendSlope, then 8:TailSlope 9:TradeDir 10:TrendDir 11:TailDir 12:Timestamp
	dirColTrade := 8
	dirColTrend := 9
	dirColTail := -1
	if showTail {
		dirColTrade = 9
		dirColTrend = 10
		dirColTail = 11
	}

	rowStart := headerRow + 1
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
			fmt.Sprintf("%.4f", s.TradeSlope),
			fmt.Sprintf("%.4f", s.TrendSlope),
		}
		if showTail {
			row = append(row, fmt.Sprintf("%.4f", s.TailSlope))
		}
		row = append(row, s.TradeDirection, s.TrendDirection)
		if showTail {
			row = append(row, s.TailDirection)
		}
		row = append(row, s.Timestamp.Format("2006-01-02"))

		for i, val := range row {
			cell, _ := excelize.CoordinatesToCellName(i+1, rowIndex)
			f.SetCellValue(sheet, cell, val)

			var style int
			switch {
			case i == dirColTrade:
				style = directionCellStyle(s.TradeDirection, bullishStyle, bearishStyle, neutralStyle, indeterminateStyle)
			case i == dirColTrend:
				style = directionCellStyle(s.TrendDirection, bullishStyle, bearishStyle, neutralStyle, indeterminateStyle)
			case showTail && i == dirColTail:
				style = directionCellStyle(s.TailDirection, bullishStyle, bearishStyle, neutralStyle, indeterminateStyle)
			case (rowIndex-rowStart)%2 == 0:
				style = altBodyStyle
			default:
				style = bodyStyle
			}
			f.SetCellStyle(sheet, cell, cell, style)
		}
		rowIndex++
	}

	f.SetPanes(sheet, &excelize.Panes{
		Freeze:      true,
		YSplit:      headerRow,
		TopLeftCell: "A5",
		ActivePane:  "bottomLeft",
	})

	footerRow := rowIndex + 1
	f.MergeCell(sheet, fmt.Sprintf("A%d", footerRow), fmt.Sprintf("%s%d", lastCol, footerRow))
	f.SetCellValue(sheet, fmt.Sprintf("A%d", footerRow), "Report generated automatically from stock range data.")
	footerStyle, _ := f.NewStyle(&excelize.Style{
		Alignment: &excelize.Alignment{Horizontal: "center"},
		Font:      &excelize.Font{Italic: true, Size: 10},
	})
	f.SetCellStyle(sheet, fmt.Sprintf("A%d", footerRow), fmt.Sprintf("%s%d", lastCol, footerRow), footerStyle)

	var widths []float64
	if showTail {
		widths = []float64{12, 10, 14, 12, 12, 12, 12, 12, 12, 12, 12, 12, 18}
	} else {
		widths = []float64{12, 10, 14, 12, 12, 12, 12, 12, 12, 12, 18}
	}
	for i, w := range widths {
		col, _ := excelize.ColumnNumberToName(i + 1)
		f.SetColWidth(sheet, col, col, w)
	}

	if err := f.SaveAs(outputPath); err != nil {
		return fmt.Errorf("failed to save XLSX: %v", err)
	}
	fmt.Printf("XLSX generated successfully at: %s\n", outputPath)
	return nil
}

// directionCellStyle returns the excelize style ID for a given direction label.
func directionCellStyle(direction string, bullish, bearish, neutral, indeterminate int) int {
	switch direction {
	case "Bullish":
		return bullish
	case "Bearish":
		return bearish
	case "Neutral":
		return neutral
	default:
		return indeterminate
	}
}
