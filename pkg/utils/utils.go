package utils

import (
	"fmt"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/tw"
	"os"
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

func PrintListWithCellConfig(headers []string, data *[][]string, cell tw.CellConfig) {
	if data == nil || len(*data) <= 0 {
		return
	}
	table := tablewriter.NewTable(os.Stdout,
		tablewriter.WithConfig(tablewriter.Config{
			Row: cell,
		}),
	)

	table.Header(headers)
	err := table.Bulk(*data)
	if err != nil {
		fmt.Printf("PrintListWithCellConfig error: %v\n", err)
		return
	}
	err = table.Render()
	if err != nil {
		fmt.Printf("PrintListWithCellConfig error: %v\n", err)
		return
	}
}

func PrintList(headers []string, data *[][]string) {
	c := tw.CellConfig{
		Formatting: tw.CellFormatting{
			MaxWidth:  50,
			AutoWrap:  tw.WrapTruncate,
			Alignment: tw.AlignNone,
		},
	}
	PrintListWithCellConfig(headers, data, c)
}

const (
	GB   string = "GB"
	MB   string = "MB"
	KB   string = "KB"
	BYTE string = "B"
)

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
