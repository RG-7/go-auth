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
func GetUsers() gin.HandlerFunc {
	return func(c *gin.Context) {
		// retrive claims for context
		claims, extists := c.Get("claims")

		if extists {
			c.JSON(http.StatusUnauthorized, gin.H{"error:": "Unauthorized"})
			return
		}

		tokenClaims, ok := claims.(*helpers.Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error:": "Invalid claims"})
		}

		if tokenClaims.Role != "ADMIN" {
			c.JSON(http.StatusUnauthorized, gin.H{"message:": "Unauthorized"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		cursor, err := userCollection.Find(ctx, bson.M{})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error:": err.Error()})
			return
		}

		defer cursor.Close(ctx)

		var users []models.User

		if err := cursor.All(ctx, &users); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error:": err.Error()})
		}

		c.JSON(http.StatusOK, users)
	}
}

// func to get each user by user id
func GetUser() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestUserId := c.Param("id")

		// get claims from the context
		claims, exists := c.Get("claims")
		if exists {
			c.JSON(http.StatusUnauthorized, gin.H{"error:": "Unauthorized"})
			return
		}

		// type assertion to get the claims object
		tokenClaims, ok := claims.(*helpers.Claims)
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error:": "Invalid claims"})
			return
		}

		userType := tokenClaims.Role
		tokeUserId := tokenClaims.UserID

		if userType != "ADMIN" && tokeUserId != requestUserId {
			c.JSON(http.StatusUnauthorized, gin.H{"error:": "Unauthorized"})
			return
		}

		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		var user models.User
		err := userCollection.FindOne(ctx, bson.M{"user_id": requestUserId}).Decode(&user)
		if err != nil {

			c.JSON(http.StatusNotFound, gin.H{"error:": "user not found"})
			return
		}

		c.JSON(http.StatusOK, user)

	}
}
