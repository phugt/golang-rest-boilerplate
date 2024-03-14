package api

import (
	"context"
	"encoding/json"
	"math"
	"net/http"
	"strings"
	"time"

	"github.com/anyshare/anyshare-admin-api/helpers"
	"github.com/anyshare/anyshare-common/mongodb"
	"github.com/anyshare/anyshare-common/schemas"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/render"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

func ListAdmin(w http.ResponseWriter, r *http.Request) {
	page := helpers.StringToInt64(r.URL.Query().Get("page"), 1)
	pageSize := helpers.StringToInt64(r.URL.Query().Get("pageSize"), 50)
	skip := (page - 1) * pageSize

	filter := bson.M{}
	keyword := strings.TrimSpace(r.URL.Query().Get("keyword"))
	if keyword != "" {
		userIds, err := findUsersByKeyword(r.Context(), keyword)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		filter["userId"] = bson.M{"$in": userIds}
	}
	status := strings.TrimSpace(r.URL.Query().Get("status"))
	switch status {
	case "deleted":
		filter["deleteTime"] = bson.M{"$gt": 0}
	default:
		filter["deleteTime"] = nil
	}

	items := []schemas.Admin{}
	db := mongodb.GetCollection(items)

	count, err := db.CountDocuments(r.Context(), filter)
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	pageCount := math.Ceil(float64(count) / float64(pageSize))

	cursor, err := db.Aggregate(r.Context(), bson.A{
		bson.M{"$match": filter},
		bson.M{"$sort": bson.M{"joinTime": -1}},
		bson.M{"$skip": skip},
		bson.M{"$limit": pageSize},
		bson.M{"$lookup": bson.M{
			"from":         "users",
			"localField":   "userId",
			"foreignField": "_id",
			"as":           "user",
		}},
		bson.M{"$project": bson.M{
			"user":  bson.M{"$arrayElemAt": []interface{}{"$user", 0}},
			"roles": "$roles",
		}},
	})
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	for cursor.Next(r.Context()) {
		item := schemas.Admin{}
		err := cursor.Decode(&item)
		if err != nil {
			w.WriteHeader(500)
			w.Write([]byte(err.Error()))
			return
		}
		items = append(items, item)
	}

	render.JSON(w, r, render.M{
		"items":     items,
		"page":      page,
		"itemCount": count,
		"pageCount": pageCount,
	})
}

func GetAdmin(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}
	item := schemas.Admin{}
	mongodb.GetCollection(item).FindOne(r.Context(), bson.M{"_id": id}).Decode(&item)
	if item.ID.IsZero() {
		w.WriteHeader(404)
		return
	}
	render.JSON(w, r, item)
}

func CreateAdmin(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var form struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=7,max=50"`
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
	hash, _ := bcrypt.GenerateFromPassword([]byte(form.Password), 10)
	result, err := mongodb.GetCollection(user).InsertOne(r.Context(), schemas.User{
		Email:    form.Email,
		Password: string(hash),
		FullName: form.FullName,
		Address:  form.Address,
		Desc:     form.Desc,
		JoinTime: time.Now().Unix(),
	})
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	render.JSON(w, r, result)
}

func UpdateAdmin(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var form struct {
		ID       string `json:"id" validate:"required"`
		Password string `json:"password" validate:"omitempty,omitnil,min=7,max=50"`
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
	id, err := primitive.ObjectIDFromHex(form.ID)
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}
	user := schemas.User{}
	db := mongodb.GetCollection(user)
	db.FindOne(r.Context(), bson.M{"_id": id}).Decode(&user)
	if user.ID.IsZero() {
		w.WriteHeader(404)
		return
	}
	updateData := bson.M{
		"fullName": form.FullName,
		"address":  form.Address,
		"desc":     form.Desc,
	}
	if form.Password != "" {
		hash, _ := bcrypt.GenerateFromPassword([]byte(form.Password), 10)
		updateData["password"] = string(hash)
	}
	result, err := db.UpdateByID(r.Context(), id, bson.M{"$set": updateData})
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	render.JSON(w, r, result)
}

func DeleteAdmin(w http.ResponseWriter, r *http.Request) {
	id, err := primitive.ObjectIDFromHex(chi.URLParam(r, "id"))
	if err != nil {
		w.WriteHeader(400)
		w.Write([]byte(err.Error()))
		return
	}
	user := schemas.User{}
	db := mongodb.GetCollection(user)
	db.FindOne(r.Context(), bson.M{"_id": id}).Decode(&user)
	if user.ID.IsZero() {
		w.WriteHeader(404)
		return
	}
	result, err := db.UpdateByID(r.Context(), id, bson.M{"$set": bson.M{"deleteTime": time.Now().Unix()}})
	if err != nil {
		w.WriteHeader(500)
		w.Write([]byte(err.Error()))
		return
	}
	render.JSON(w, r, result)
}

func findUsersByKeyword(ctx context.Context, keyword string) (userIds []primitive.ObjectID, err error) {
	var pageSize int64 = 30
	cursor, err := mongodb.GetCollection(schemas.User{}).
		Find(ctx,
			bson.M{"$regex": keyword},
			&options.FindOptions{Limit: &pageSize, Sort: bson.D{{Key: "joinTime", Value: -1}}},
		)
	if err != nil {
		return nil, err
	}
	for cursor.Next(ctx) {
		user := schemas.User{}
		err := cursor.Decode(&user)
		if err != nil {
			return nil, err
		}
		userIds = append(userIds, user.ID)
	}
	return userIds, nil
}
