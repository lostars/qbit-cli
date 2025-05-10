package utils

import (
	"github.com/olekukonko/tablewriter"
	"os"
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
	table := tablewriter.NewWriter(os.Stdout)
	table.SetHeader(headers)
	table.SetAutoWrapText(false)
	for _, v := range *data {
		table.Append(v)
	}
	table.Render()
}
