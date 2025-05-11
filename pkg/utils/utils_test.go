package utils

import (
	"fmt"
	"regexp"
	"testing"
)

func TestTruncateString(t *testing.T) {
	println(TruncateString("中文abc", 0, 2))
	//println(len("中文"))
	//println(len([]rune("中文")))
	//fs := make(map[string]string, 2)
	//fmt.Println(len(fs))
	//for i := range fs {
	//	fmt.Println(fs[i] + "sssss")
	//}
	//fmt.Println(filepath.Ext("ab.txt"))
	//fmt.Println(len(strings.Split("ab.txt", filepath.Ext("ab.txt"))))
	//fmt.Println(strings.Split("a", "/"))

	//var JPCodeRegex = regexp.MustCompile("([a-zA-Z]{2,5}-[0-9]{3,5}|FC2-PPV-\\d{5,})")
	//var JPCodeRegex = regexp.MustCompile("\\d+[-_]([123])")
	//matches := JPCodeRegex.FindStringSubmatch("4k2.com@ipvr00301_1_hq.mp4")
	//fmt.Println(matches[1])

	var JP4KRegex = regexp.MustCompile(`([-\[])(4[kK])`)
	fmt.Println(JP4KRegex.FindStringSubmatch("-4k]"))

	// 4k2.com@ipvr00301_1_hq.mp4
}

func TestFormatFileSize(t *testing.T) {
	println(FormatFileSizeAuto(129789017287, 1))
	println(FormatFileSizeAuto(12978901, 1))
	println(FormatFileSizeAuto(200000, 0))
	println(FormatFileSizeAuto(1024, 0))
	println(FormatFileSizeAuto(124, 0))
}
