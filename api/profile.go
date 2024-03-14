package api

import (
	"encoding/json"
	"net/http"

	"github.com/anyshare/anyshare-admin-api/enum"
	"github.com/anyshare/anyshare-admin-api/helpers"
	"github.com/anyshare/anyshare-common/mongodb"
	"github.com/anyshare/anyshare-common/schemas"
	"github.com/go-chi/render"
	"go.mongodb.org/mongo-driver/bson"
	"golang.org/x/crypto/bcrypt"
)

func GetProfile(w http.ResponseWriter, r *http.Request) {
	render.JSON(w, r, r.Context().Value(enum.ContextKeyUser))
}

func UpdateProfile(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var form struct {
		Email    string `json:"email"`
		FullName string `json:"fullName" validate:"required,max=50"`
		Address  string `json:"address" validate:"required,max=250"`
		Desc     string `json:"desc" validate:"max=1000"`
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
	db := mongodb.GetCollection(user)
	userId := r.Context().Value(enum.ContextKeyUser).(schemas.User).ID
	db.FindOne(r.Context(), bson.M{"_id": userId}).Decode(&user)
	if user.ID.IsZero() {
		w.WriteHeader(422)
		render.JSON(w, r, render.M{"email": helpers.Translate(r.Context(), "accountNotExist")})
		return
	}
	result, err := db.UpdateByID(r.Context(), userId, bson.M{"$set": bson.M{
		"full_name": form.FullName,
		"address":   form.Address,
		"desc":      form.Desc,
	}})
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	render.JSON(w, r, result)
}

func ChangePassword(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var form struct {
		OldPassword string `json:"oldPassword" validate:"required"`
		NewPassword string `json:"newPassword" validate:"required,min=7,max=50"`
		RePassword  string `json:"rePassword" validate:"required"`
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
	db := mongodb.GetCollection(user)
	userId := r.Context().Value(enum.ContextKeyUser).(schemas.User).ID
	db.FindOne(r.Context(), bson.M{"_id": userId}).Decode(&user)
	if user.ID.IsZero() {
		w.WriteHeader(422)
		render.JSON(w, r, render.M{"oldPassword": helpers.Translate(r.Context(), "wrongPassword")})
		return
	}
	if bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(form.OldPassword)) != nil {
		w.WriteHeader(422)
		render.JSON(w, r, render.M{"oldPassword": helpers.Translate(r.Context(), "wrongPassword")})
		return
	}
	if form.NewPassword != form.RePassword {
		w.WriteHeader(422)
		render.JSON(w, r, render.M{"rePassword": helpers.Translate(r.Context(), "passwordNotMatch")})
		return
	}
	hash, _ := bcrypt.GenerateFromPassword([]byte(form.NewPassword), 10)
	result, err := db.UpdateByID(r.Context(), userId, bson.M{"$set": bson.M{"password": string(hash)}})
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	render.JSON(w, r, result)
}
