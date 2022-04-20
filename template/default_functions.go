package template

import (
	"strconv"
	"strings"
	gotemplate "text/template"
	"time"
)

func GetDefaultFuncs() gotemplate.FuncMap {
	return map[string]interface{}{
		"Now":            Now,
		"NowUnix":        NowUnix,
		"NowUnixNano":    NowUnixNano,
		"TimeFormat":     TimeFormat,
		"ToLower":        ToLower,
		"Replace":        Replace,
		"Split":          Split,
		"ParseFloat":     ParseFloat,
		"MustParseFloat": MustParseFloat,
	}
}

func Now(vv ...interface{}) string {
	format := time.RFC3339
	if len(vv) > 0 {
		fmt, isString := vv[0].(string)
		if isString {
			format = fmt
		}
	}
	return time.Now().Format(format)
}

func NowUnix() int64 {
	return time.Now().Unix()
}

func NowUnixNano() int64 {
	return time.Now().UnixNano()
}

func TimeFormat(inLayout, outLayout, in string) string {
	t, err := time.Parse(inLayout, in)
	if err != nil {
		t = time.Unix(0, 0)
	}
	return t.Format(outLayout)
}

func MustTimeFormat(inLayout, outLayout, in string) (string, error) {
	t, err := time.Parse(inLayout, in)
	if err != nil {
		return "", err
	}
	return t.Format(outLayout), nil
}

func ToLower(str string) string {
	return strings.ToLower(str)
}

func Replace(s, old, new string, n int) string {
	return strings.Replace(s, old, new, n)
}

func Split(s, sep string) []string {
	return strings.Split(s, sep)
}

func ParseFloat(s string, bitSize int) float64 {
	f, _ := strconv.ParseFloat(s, bitSize)
	return f
}

func MustParseFloat(s string, bitSize int) (float64, error) {
	return strconv.ParseFloat(s, bitSize)
}
