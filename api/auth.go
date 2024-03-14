package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/anyshare/anyshare-admin-api/helpers"
	"github.com/anyshare/anyshare-common/mongodb"
	"github.com/anyshare/anyshare-common/schemas"
	"github.com/go-chi/render"
	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

func Login(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var form struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}
	err := decoder.Decode(&form)
	if err != nil {
		w.WriteHeader(400)
		return
	}
	formErrors := helpers.ValidateStruct(r.Context(), form)
	if formErrors != nil {
		w.WriteHeader(422)
		render.JSON(w, r, formErrors)
		return
	}
	user := schemas.User{}
	mongodb.GetCollection(user).FindOne(r.Context(), bson.M{"email": strings.ToLower(form.Email)}).Decode(&user)
	if user.ID.IsZero() {
		w.WriteHeader(422)
		render.JSON(w, r, render.M{"email": helpers.Translate(r.Context(), "accountNotExist")})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(form.Password)) != nil {
		w.WriteHeader(422)
		render.JSON(w, r, render.M{"password": helpers.Translate(r.Context(), "wrongPassword")})
		return
	}
	userToken := schemas.UserToken{
		ID:         uuid.NewString(),
		UserID:     user.ID,
		UserAgent:  r.UserAgent(),
		CreateTime: time.Now().Unix(),
	}
	result, err := mongodb.GetCollection(userToken).InsertOne(r.Context(), userToken)
	if result == nil || err != nil {
		w.WriteHeader(500)
		return
	}
	render.JSON(w, r, render.M{
		"token": result.InsertedID,
		"user":  user,
	})
}
