package controllers

import (
	"context"
	"net/http"
	"time"

	"github.com/RG-7/go-auth/config"
	"github.com/RG-7/go-auth/helpers"
	"github.com/RG-7/go-auth/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var userCollection = config.OpenCollection("users")

var validate = validator.New()

// function for signup
func Signup() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		var user models.User

		// get user input
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		// validate user input
		if validationErr := validate.Struct(user); validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		//
		count, err := userCollection.CountDocuments(ctx, bson.M{
			"$or": []bson.M{
				{"email": user.Email},
				{"phone": user.Phone},
			},
		})

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		if count > 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Email or Phone already exists."})
		}

		user.Password = helpers.HashPassword(user.Password)
		user.Created_at = time.Now()
		user.Updated_at = time.Now()
		user.ID = primitive.NewObjectID()
		user.User_id = user.ID.Hex()
		accessToken, refreshToken := helpers.GenerateToken(*user.Email, *user.Role, user.User_id)
		user.Token = &accessToken
		user.Refresh_token = &refreshToken

		_, insertErr := userCollection.InsertOne(ctx, user)
		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error: ": insertErr.Error()})
		}

		c.JSON(http.StatusOK, gin.H{"message:": "user created successfully"})
	}
}

// fucntion for login
func Login() gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		var user models.User
		var foundUser models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := userCollection.FindOne(ctx, bson.M{"email": user.Email, "phone": user.Phone}).Decode(&foundUser)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error:": "Invalid email or password"})
		}
		passwordIsValid, msg := helpers.VerifyPassword(*foundUser.Password, *user.Password)
		if !passwordIsValid {
			c.JSON(http.StatusUnauthorized, gin.H{"error:": msg})
		}

		token, refreshToken := helpers.GenerateToken(*foundUser.Email, foundUser.User_id, *foundUser.Role)
		helpers.UpdateAllToken(token, refreshToken, foundUser.User_id)

		c.JSON(http.StatusOK, gin.H{"user": foundUser, "token": token, "refresh_token": refreshToken})

	}
}

// func to get users
func GetUsers() {}

// func to get each user by user id
func GetUser() {}
