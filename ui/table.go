package ui

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

type TableStyle struct {
	HeaderStyle  lipgloss.Style
	RowStyle     lipgloss.Style
	BorderStyle  lipgloss.Border
	BorderHeader bool
	BorderColumn bool
	BorderTop    bool
	BorderLeft   bool
	BorderBottom bool
	BorderRight  bool
}

var TableStyleDefault = TableStyle{
	HeaderStyle:  lipgloss.NewStyle().Foreground(Primary).Bold(true).Padding(0, 1),
	RowStyle:     lipgloss.NewStyle().Padding(0, 1),
	BorderStyle:  lipgloss.HiddenBorder(),
	BorderHeader: false,
	BorderColumn: false,
	BorderTop:    false,
	BorderLeft:   false,
	BorderBottom: false,
	BorderRight:  false,
}

var TableStyleMarkdown = TableStyle{
	HeaderStyle: lipgloss.NewStyle().Bold(true).Padding(0, 1),
	RowStyle:    lipgloss.NewStyle().Padding(0, 1),
	BorderStyle: lipgloss.Border{
		Left:  "|",
		Right: "|",

		Top:      "-",
		TopLeft:  "|",
		TopRight: "|",

		Bottom:      "-",
		BottomLeft:  "|",
		BottomRight: "|",

		Middle:      "|",
		MiddleLeft:  "|",
		MiddleRight: "|",

		MiddleTop:    "|",
		MiddleBottom: "|",
	},
	BorderHeader: true,
	BorderColumn: true,
	BorderTop:    false,
	BorderLeft:   true,
	BorderBottom: false,
	BorderRight:  true,
}

type TableRow = map[string]string

type TableAlignment int

const (
	TableLeft TableAlignment = iota
	TableRight
	TableCenter
)

type TableColumn struct {
	Key       string
	Title     string
	Active    bool
	MaxWidth  int
	Alignment TableAlignment
	ValueFunc func(value string) string
	StyleFunc func(style lipgloss.Style, value string) lipgloss.Style
}

func NewTableColumn(key string, title string) TableColumn {
	return TableColumn{
		Key:       key,
		Title:     title,
		Active:    true,
		MaxWidth:  -1,
		Alignment: TableLeft,
		StyleFunc: func(style lipgloss.Style, value string) lipgloss.Style {
			return style
		},
		ValueFunc: func(value string) string {
			return value
		},
	}
}

func (c TableColumn) WithMaxWidth(w int) TableColumn {
	c.MaxWidth = w
	return c
}

func (c TableColumn) WithAlignment(a TableAlignment) TableColumn {
	c.Alignment = a
	return c
}

func (c TableColumn) WithActive(a bool) TableColumn {
	c.Active = a
	return c
}

func (c TableColumn) WithValueFunc(
	valueFunc func(value string) string,
) TableColumn {
	c.ValueFunc = valueFunc
	return c
}

func (c TableColumn) WithStyleFunc(
	styleFunc func(style lipgloss.Style, value string) lipgloss.Style,
) TableColumn {
	c.StyleFunc = styleFunc
	return c
}

type Table struct {
	columns     []TableColumn
	rows        []TableRow
	emptyString string
	style       TableStyle
}

func NewTable(columns []TableColumn) Table {
	return Table{
		columns:     columns,
		rows:        []TableRow{},
		emptyString: "-",
		style:       TableStyleDefault,
	}
}

func (t Table) WithEmptyString(s string) Table {
	t.emptyString = s
	return t
}

func (t Table) WithStyle(s TableStyle) Table {
	t.style = s
	return t
}

func (t Table) WithRows(rows []TableRow) Table {
	t.rows = rows
	return t
}

func (t *Table) getRowMatrix() [][]string {
	rows := make([][]string, 0)
	for _, rowEntry := range t.rows {
		row := []string{}
		for _, col := range t.columns {
			if !col.Active {
				continue
			}

			value := col.ValueFunc(rowEntry[col.Key])
			if value == "" {
				value = t.emptyString
			}
			if col.MaxWidth > 0 && col.MaxWidth < len(value) {
				value = fmt.Sprintf("%.*s...", col.MaxWidth-3, value)
			}
			row = append(row, value)
		}
		rows = append(rows, row)
	}
	return rows
}

func (t *Table) Render() string {
	headers := make([]string, 0)

	columnOffset := 0
	columnOffsets := make([]int, 0)
	for _, col := range t.columns {
		if !col.Active {
			columnOffset += 1
			continue
		}

		columnOffsets = append(columnOffsets, columnOffset)
		headers = append(headers, col.Title)
	}

	rows := t.getRowMatrix()

	lt := table.New().
		Headers(headers...).
		Rows(rows...).
		Border(t.style.BorderStyle).
		BorderLeft(t.style.BorderLeft).BorderRight(t.style.BorderRight).
		BorderTop(t.style.BorderTop).BorderBottom(t.style.BorderBottom).
		BorderHeader(t.style.BorderHeader).BorderColumn(t.style.BorderColumn).
		StyleFunc(func(row int, col int) lipgloss.Style {
			var sty lipgloss.Style
			column := t.columns[col+columnOffsets[col]]

			if row == table.HeaderRow {
				sty = t.style.HeaderStyle
			} else {
				sty = column.StyleFunc(t.style.RowStyle, rows[row][col])
			}

			switch column.Alignment {
			case TableLeft:
				sty = sty.Align(lipgloss.Left)
			case TableCenter:
				sty = sty.Align(lipgloss.Center)
			case TableRight:
				sty = sty.Align(lipgloss.Right)
			}

			return sty
		})

	return lt.Render()
}

func (t *Table) ExportCSV(path string) error {
	fd, err := os.Create(path)
	if err != nil {
		return err
	}
	w := csv.NewWriter(fd)

	header := make([]string, 0)
	for _, col := range t.columns {
		if col.Active {
			header = append(header, col.Title)
		}
	}

	err = w.Write(header)
	if err != nil {
		return err
	}
	err = w.WriteAll(t.getRowMatrix())
	if err != nil {
		return err
	}

	return nil
}
