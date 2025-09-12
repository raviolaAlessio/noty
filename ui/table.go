package ui

import (
	"encoding/csv"
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
)

// Style definitions
var (
	HeaderStyle = lipgloss.NewStyle().Foreground(Primary).Bold(true).Padding(0, 1)
	RowStyle    = lipgloss.NewStyle().Padding(0, 1)
)

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

func NewTableColumn(key string, title string, active bool) TableColumn {
	return TableColumn{
		Key:       key,
		Title:     title,
		Active:    active,
		MaxWidth: -1,
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
}

func NewTable(columns []TableColumn) Table {
	return Table{
		columns:     columns,
		rows:        []TableRow{},
		emptyString: "-",
	}
}

func (t Table) WithEmptyString(s string) Table {
	t.emptyString = s
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
				value = fmt.Sprintf("%s...", value[0:col.MaxWidth - 3])
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
		Border(lipgloss.NormalBorder()).
		BorderLeft(false).BorderRight(false).BorderTop(false).BorderBottom(false).
		BorderColumn(false).BorderHeader(false).
		StyleFunc(func(row int, col int) lipgloss.Style {
			var sty lipgloss.Style
			column := t.columns[col + columnOffsets[col]]

			if row == table.HeaderRow {
				sty = HeaderStyle
			} else {
				sty = column.StyleFunc(RowStyle, rows[row][col])
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
