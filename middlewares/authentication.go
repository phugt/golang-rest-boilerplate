package middlewares

import (
	"context"
	"net/http"
	"strings"

	"github.com/anyshare/anyshare-admin-api/enum"
	"github.com/anyshare/anyshare-common/mongodb"
	"github.com/anyshare/anyshare-common/schemas"
	"go.mongodb.org/mongo-driver/bson"
)

func Authentication(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") == "" {
			w.WriteHeader(401)
			return
		}
		var (
			user      = schemas.User{}
			userToken = schemas.UserToken{}
		)
		authHeader := r.Header.Get("Authorization")
		mongodb.GetCollection(userToken).FindOne(r.Context(), bson.M{
			"_id": strings.ReplaceAll(authHeader, "Bearer ", ""),
		}).Decode(&userToken)
		if userToken.ID == "" {
			w.WriteHeader(401)
			return
		}
		mongodb.GetCollection(user).FindOne(r.Context(), bson.M{"_id": userToken.UserID}).Decode(&user)
		if user.ID.IsZero() {
			w.WriteHeader(401)
			return
		}
		ctx := context.WithValue(r.Context(), enum.ContextKeyUser, user)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
