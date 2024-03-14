package helpers

import (
	"context"

	"github.com/anyshare/anyshare-admin-api/enum"
	enTranslation "github.com/anyshare/anyshare-admin-api/translations/en"
	viTranslation "github.com/anyshare/anyshare-admin-api/translations/vi"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/vi"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
)

func ValidateStruct(ctx context.Context, data interface{}) validator.ValidationErrorsTranslations {
	validate := validator.New(validator.WithRequiredStructEnabled())
	utrans := ut.New(en.New(), en.New(), vi.New())
	utrans.Import(ut.FormatJSON, "translations")
	utrans.VerifyTranslations()

	locale := "en"
	if ctx.Value(enum.ContextKeyLocale) != nil {
		locale = ctx.Value(enum.ContextKeyLocale).(string)
	}
	trans, _ := utrans.GetTranslator(locale)

	if locale == "vi" {
		viTranslation.RegisterDefaultTranslations(validate, trans)
	} else {
		enTranslation.RegisterDefaultTranslations(validate, trans)
	}

	err := validate.Struct(data)
	if err != nil {
		errs := make(map[string]string)
		for key, value := range err.(validator.ValidationErrors).Translate(trans) {
			errs[LowerFirstChar(key)] = value
		}

		return errs
	}
	return nil
}
