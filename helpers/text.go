package helpers

import (
	"context"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/anyshare/anyshare-admin-api/enum"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/vi"
	ut "github.com/go-playground/universal-translator"
)

func Translate(ctx context.Context, tag string) string {
	utrans := ut.New(en.New(), en.New(), vi.New())
	utrans.Import(ut.FormatJSON, "translations")
	utrans.VerifyTranslations()

	locale := "en"
	if ctx.Value(enum.ContextKeyLocale) != nil {
		locale = ctx.Value(enum.ContextKeyLocale).(string)
	}

	trans, _ := utrans.GetTranslator(locale)
	traslated, err := trans.T(tag)
	if err != nil {
		return tag
	}
	return traslated
}

func LowerFirstChar(s string) string {
	r, size := utf8.DecodeRuneInString(s)
	if r == utf8.RuneError && size <= 1 {
		return s
	}
	lc := unicode.ToLower(r)
	if r == lc {
		return s
	}
	return string(lc) + s[size:]
}

func ToSnakeCase(str string) string {
	matchFirstCap := regexp.MustCompile("(.)([A-Z][a-z]+)")
	matchAllCap := regexp.MustCompile("([a-z0-9])([A-Z])")
	snake := matchFirstCap.ReplaceAllString(str, "${1}_${2}")
	snake = matchAllCap.ReplaceAllString(snake, "${1}_${2}")
	return strings.ToLower(snake)
}

func StringToInt64(str string, df int64) int64 {
	if str == "" {
		return df
	}
	rs, err := strconv.ParseInt(str, 10, 0)
	if err != nil {
		rs = df
	}
	return rs
}
