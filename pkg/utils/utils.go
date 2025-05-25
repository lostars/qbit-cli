package utils

import (
	"fmt"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/lipgloss/table"
	"strconv"
	"strings"
	"time"
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

func PrintListWithColWidth(headers []string, data *[][]string, widthMap map[int]int, warp bool) {
	PrintListWithStyleFunc(headers, data, func(row, col int) lipgloss.Style {
		if row == table.HeaderRow {
			return DefaultHeaderStyle()
		}
		if width := widthMap[col]; width > 0 {
			return DefaultCellStyle().Width(width)
		}
		return DefaultCellStyle()
	}, warp)
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
	return lipgloss.NewStyle().PaddingLeft(1).PaddingRight(1).Align(lipgloss.Left)
}

type InteractiveModel struct {
	Rows          *[][]string
	Header        *[]string
	cursor        int
	width, height int
	startIndex    int
	WidthMap      map[int]int
	Delegate      InteractiveKeyMsgDelegate
	DataDelegate  DynamicDataDelegate
	notifyMsg     *NotifyMsg
	clickedMap    map[int]bool
}

type NotifyMsg struct {
	Msg      string
	Duration time.Duration
}

type ClearNotifyMsg struct{}

type DateUpdateMsg struct{}

type InteractiveKeyMsgDelegate interface {
	Operation(msg tea.KeyMsg, cursor int) tea.Cmd
}

type DynamicDataDelegate interface {
	Update() *[][]string
	Frequency() time.Duration
}

func (m *InteractiveModel) Init() tea.Cmd {
	if m.Rows == nil {
		m.Rows = &[][]string{}
	}
	if m.Header == nil {
		m.Header = &[]string{}
	}
	if m.WidthMap == nil {
		m.WidthMap = map[int]int{}
	}
	m.clickedMap = make(map[int]bool, len(*m.Rows))
	if m.DataDelegate != nil {
		return tea.Tick(m.DataDelegate.Frequency(), func(t time.Time) tea.Msg {
			return DateUpdateMsg{}
		})
	}
	return nil
}

func (m *InteractiveModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case DateUpdateMsg:
		data := m.DataDelegate.Update()
		if data == nil {
			data = &[][]string{}
		}
		m.Rows = data
		return m, tea.Tick(m.DataDelegate.Frequency(), func(t time.Time) tea.Msg {
			return DateUpdateMsg{}
		})
	case ClearNotifyMsg:
		m.notifyMsg = nil
	case NotifyMsg:
		m.notifyMsg = &msg
		if m.notifyMsg != nil && m.notifyMsg.Duration != 0 {
			return m, tea.Tick(m.notifyMsg.Duration, func(t time.Time) tea.Msg {
				return ClearNotifyMsg{}
			})
		}
	case tea.KeyMsg:
		switch msg.String() {
		default:
			if m.Delegate != nil {
				m.clickedMap[m.cursor] = true
				return m, m.Delegate.Operation(msg, m.cursor)
			}
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(*m.Rows)-1 {
				m.cursor++
			}
		case "ctrl+c", "q":
			return m, tea.Quit
		}
		maxVisible := m.height - 4
		if maxVisible < 1 {
			maxVisible = 1
		}
		if m.cursor < m.startIndex {
			m.startIndex = m.cursor
		} else if m.cursor >= m.startIndex+maxVisible {
			m.startIndex = m.cursor - maxVisible + 1
		}
	case tea.WindowSizeMsg:
		m.width = msg.Width - 1
		m.height = msg.Height - 2
	}
	return m, nil
}

func (m *InteractiveModel) View() string {
	wrap := len(m.WidthMap) <= 0
	ta := table.New().
		Border(lipgloss.ASCIIBorder()).
		Width(m.width).
		Headers(*m.Header...).
		StyleFunc(func(row, col int) lipgloss.Style {
			if row == table.HeaderRow {
				return DefaultHeaderStyle()
			}
			style := DefaultCellStyle()
			if !wrap {
				if width := m.WidthMap[col]; width > 0 {
					style = style.Width(width)
				}
			}

			dataIndex := m.startIndex + row
			if dataIndex == m.cursor {
				return style.Foreground(lipgloss.Color("#000000")).Background(lipgloss.Color("#FFFFFF"))
			}
			if m.clickedMap[dataIndex] {
				return style.Foreground(lipgloss.Color("#383535"))
			}

			return style
		}).Wrap(wrap)

	maxVisible := m.height - 4
	if maxVisible < 1 {
		maxVisible = 1
	}
	end := m.startIndex + maxVisible
	if end > len(*m.Rows) {
		end = len(*m.Rows)
	}
	for i := m.startIndex; i < end; i++ {
		d := *m.Rows
		ta.Rows(d[i])
	}

	if m.notifyMsg != nil {
		return fmt.Sprintf("%s\n%s\n", ta.String(), m.notifyMsg.Msg)
	} else {
		return fmt.Sprintf("%s\n\n", ta.String())
	}
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
	str := ""
	unit := ""
	if bytes < 1024 {
		str = strconv.FormatUint(bytes, 10)
		unit = BYTE
	} else if bytes < 1024*1024 {
		str = strconv.FormatFloat(float64(bytes)/1024, 'f', decimal, 64)
		unit = KB
	} else if bytes < 1024*1024*1024 {
		str = strconv.FormatFloat(float64(bytes)/1024/1024, 'f', decimal, 64)
		unit = MB
	} else {
		str = strconv.FormatFloat(float64(bytes)/1024/1024/1024, 'f', decimal, 64)
		unit = GB
	}
	if decimal > 0 {
		zero := make([]string, 0, decimal+1)
		zero = append(zero, ".")
		for i := 0; i < decimal; i++ {
			zero = append(zero, "0")
		}
		return strings.TrimSuffix(str, strings.Join(zero, "")) + unit
	} else {
		return str + unit
	}
}
