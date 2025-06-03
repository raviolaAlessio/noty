package ui

import (
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
	Alignment TableAlignment
}

func NewTableColumn(key string, title string, active bool) TableColumn {
	return TableColumn{
		Key:         key,
		Title:       title,
		Active:      active,
		Alignment:   TableLeft,
	}
}

func (c TableColumn) WithAlignment(a TableAlignment) TableColumn {
	c.Alignment = a
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

func (t Table) Render() string {
	aligments := []TableAlignment{}
	headers := []string{}
	for _, col := range t.columns {
		if !col.Active {
			continue
		}
		headers = append(headers, col.Title)
		aligments = append(aligments, col.Alignment)
	}

	rows := [][]string{}
	for _, rowEntry := range t.rows {
		row := []string{}
		for _, col := range t.columns {
			if !col.Active {
				continue
			}

			value := rowEntry[col.Key]
			if value == "" {
				value = t.emptyString
			}
			row = append(row, value)
		}
		rows = append(rows, row)
	}

	lt := table.New().
		Headers(headers...).
		Rows(rows...).
		Border(lipgloss.NormalBorder()).
		BorderLeft(false).BorderRight(false).BorderTop(false).BorderBottom(false).
		BorderColumn(false).BorderHeader(false).
		StyleFunc(func(row, col int) lipgloss.Style {
			var sty lipgloss.Style

			switch {
			case row == table.HeaderRow:
				sty = HeaderStyle
			default:
				sty = RowStyle
			}

			switch aligments[col] {
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
