package middlewares

import (
	"context"
	"net/http"

	"github.com/anyshare/anyshare-admin-api/enum"
)

func LocaleHeader(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.WithValue(r.Context(), enum.ContextKeyLocale, r.Header.Get("Accept-Language"))
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
