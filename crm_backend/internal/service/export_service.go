package service

import (
	"context"
	"fmt"
	"io"

	"github.com/johnfercher/maroto/v2"
	"github.com/johnfercher/maroto/v2/pkg/components/row"
	"github.com/johnfercher/maroto/v2/pkg/components/text"
	"github.com/johnfercher/maroto/v2/pkg/config"
	"github.com/johnfercher/maroto/v2/pkg/consts/align"
	"github.com/johnfercher/maroto/v2/pkg/consts/fontstyle"
	"github.com/johnfercher/maroto/v2/pkg/consts/orientation"
	"github.com/johnfercher/maroto/v2/pkg/props"
	excel "github.com/xuri/excelize/v2"
)

type ExportService struct {
	leadRepo LeadRepository
}

func NewExportService(leadRepo LeadRepository) *ExportService {
	return &ExportService{leadRepo: leadRepo}
}

func (s *ExportService) GenerateLeadsExcel(ctx context.Context, w io.Writer) error {
	leads, _, err := s.leadRepo.List(ctx, 500, 0)
	if err != nil {
		return err
	}

	f := excel.NewFile()
	defer f.Close()

	sheet := "Leads"
	index, _ := f.NewSheet(sheet)
	f.DeleteSheet("Sheet1")

	headers := []string{"ID", "First Name", "Last Name", "Email", "Phone", "Program ID", "Created At"}
	for i, h := range headers {
		cell, _ := excel.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}

	for i, lead := range leads {
		row := i + 2
		f.SetCellValue(sheet, fmt.Sprintf("A%d", row), lead.ID)
		f.SetCellValue(sheet, fmt.Sprintf("B%d", row), lead.FirstName)
		f.SetCellValue(sheet, fmt.Sprintf("C%d", row), lead.LastName)
		f.SetCellValue(sheet, fmt.Sprintf("D%d", row), lead.Email)
		f.SetCellValue(sheet, fmt.Sprintf("E%d", row), lead.Phone)
		programID := ""
		if lead.ProgramID != nil {
			programID = fmt.Sprintf("%d", *lead.ProgramID)
		}
		f.SetCellValue(sheet, fmt.Sprintf("F%d", row), programID)
		f.SetCellValue(sheet, fmt.Sprintf("G%d", row), lead.CreatedAt.Format("2006-01-02 15:04"))
	}

	f.SetActiveSheet(index)
	return f.Write(w)
}

func (s *ExportService) GenerateLeadsPDF(ctx context.Context) ([]byte, error) {
	leads, _, err := s.leadRepo.List(ctx, 500, 0)
	if err != nil {
		return nil, err
	}

	cfg := config.NewBuilder().
		WithOrientation(orientation.Horizontal).
		Build()
	m := maroto.New(cfg)

	m.AddRows(row.New(15).Add(
		text.NewCol(12, "MDIS CRM - Detailed Leads Report", props.Text{Size: 18, Align: align.Center, Style: fontstyle.Bold}),
	))

	m.AddRows(row.New(10).Add(
		text.NewCol(1, "ID", props.Text{Size: 9, Style: fontstyle.Bold}),
		text.NewCol(3, "Name", props.Text{Size: 9, Style: fontstyle.Bold}),
		text.NewCol(3, "Email", props.Text{Size: 9, Style: fontstyle.Bold}),
		text.NewCol(2, "Phone", props.Text{Size: 9, Style: fontstyle.Bold}),
		text.NewCol(1, "Prog.", props.Text{Size: 9, Style: fontstyle.Bold}),
		text.NewCol(2, "Created At", props.Text{Size: 9, Style: fontstyle.Bold}),
	))

	for _, l := range leads {
		progStr := ""
		if l.ProgramID != nil {
			progStr = fmt.Sprintf("%d", *l.ProgramID)
		}
		m.AddRows(row.New(8).Add(
			text.NewCol(1, fmt.Sprintf("%d", l.ID), props.Text{Size: 8}),
			text.NewCol(3, l.FirstName+" "+l.LastName, props.Text{Size: 8}),
			text.NewCol(3, l.Email, props.Text{Size: 8}),
			text.NewCol(2, l.Phone, props.Text{Size: 8}),
			text.NewCol(1, progStr, props.Text{Size: 8}),
			text.NewCol(2, l.CreatedAt.Format("2006-01-02"), props.Text{Size: 8}),
		))
	}

	doc, err := m.Generate()
	if err != nil {
		return nil, err
	}
	return doc.GetBytes(), nil
}
