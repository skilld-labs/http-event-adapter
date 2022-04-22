package template

import (
	"strconv"
	"strings"
	gotemplate "text/template"
	"time"
)

func GetDefaultFuncs() gotemplate.FuncMap {
	return map[string]interface{}{
		"Now":                Now,
		"NowUnix":            NowUnix,
		"NowUnixNano":        NowUnixNano,
		"TimeFormat":         TimeFormat,
		"TimeChangeTimeZone": TimeChangeTimeZone,
		"TimeParse":          TimeParse,
		"MustTimeParse":      MustTimeParse,
		"ToLower":            ToLower,
		"Replace":            Replace,
		"Split":              Split,
		"ParseFloat":         ParseFloat,
		"MustParseFloat":     MustParseFloat,
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

func TimeChangeTimeZone(timeZone string, t time.Time) (time.Time, error) {
	tz, err := time.LoadLocation(timeZone)
	if err != nil {
		return t, err
	}
	return t.In(tz), nil
}

func TimeParse(layout, timeZone, value string) time.Time {
	t, err := MustTimeParse(layout, timeZone, value)
	if err != nil {
		t = time.Unix(0, 0)
	}
	return t
}

func MustTimeParse(layout, timeZone, value string) (time.Time, error) {
	tz, err := time.LoadLocation(timeZone)
	if err != nil {
		return time.Time{}, err
	}
	return time.ParseInLocation(layout, value, tz)
}

func TimeFormat(layout string, t time.Time) string {
	return t.Format(layout)
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
