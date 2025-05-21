package utils

import (
	"fmt"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"strconv"
	"unicode/utf8"
)

func TruncateStringFromStart(str string, l int) string {
	return TruncateString(str, 0, l)
}

func TruncateString(str string, start int, end int) string {
	if len(str) == utf8.RuneCountInString(str) {
		l := len(str)
		if start < 0 {
			start = 0
		}
		if end > l {
			end = l
		}
		return str[start:end]
	}

	runes := []rune(str)
	l := len(runes)
	if start < 0 {
		start = 0
	}
	if end > l {
		end = l
	}
	return string(runes[start:end])
}

func PrintList(headers []string, data *[][]string) {
	PrintListWithStyleFunc(headers, data, func(row, col int) lipgloss.Style {
		if row == table.HeaderRow {
			return DefaultHeaderStyle()
		}
		return DefaultCellStyle()
	}, true)
}

func PrintListWithStyleFunc(headers []string, data *[][]string, styleFunc table.StyleFunc, wrap bool) {
	t := table.New().
		Border(lipgloss.ASCIIBorder()).
		Headers(headers...).
		StyleFunc(styleFunc).
		Wrap(wrap)
	for _, row := range *data {
		t.Rows(row)
	}
	fmt.Println(t)
}

func DefaultHeaderStyle() lipgloss.Style {
	return lipgloss.NewStyle().Align(lipgloss.Center)
}

func DefaultCellStyle() lipgloss.Style {
	return lipgloss.NewStyle().MarginLeft(1).MarginRight(1).Align(lipgloss.Left)
}

const (
	GB   string = "GB"
	MB   string = "MB"
	KB   string = "KB"
	BYTE string = "B"
)

func FormatPercent(decimal float64) string {
	return strconv.FormatInt(int64(decimal*100), 10) + "%"
}

func FormatFileSizeAuto(bytes uint64, decimal int) string {
	if bytes < 1024 {
		return strconv.FormatUint(bytes, 10) + BYTE
	} else if bytes < 1024*1024 {
		return strconv.FormatFloat(float64(bytes)/1024, 'f', decimal, 64) + KB
	} else if bytes < 1024*1024*1024 {
		return strconv.FormatFloat(float64(bytes)/1024/1024, 'f', decimal, 64) + MB
	} else {
		return strconv.FormatFloat(float64(bytes)/1024/1024/1024, 'f', decimal, 64) + GB
	}
}
