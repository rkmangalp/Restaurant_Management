package controllers

import (
	"context"
	"fmt"
	"log"
	"math"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"               // Gin framework for handling HTTP requests
	"github.com/go-playground/validator/v10" // Input validation package
	"github.com/rkmangalp/Restaurant_Management/database"
	"github.com/rkmangalp/Restaurant_Management/models"
	"go.mongodb.org/mongo-driver/bson" // BSON format for MongoDB interactions
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var foodCollection *mongo.Collection = database.OpenCollection(database.Client, "food")
var validate = validator.New()

// GetFoods returns a function (gin.HandlerFunc) to be used as a route handler in Gin.
// c *gin.Context provides the HTTP request context, allowing access to query params,
// response writing, etc.

func GetFoods() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Creates a context with a timeout of 100 seconds to prevent MongoDB operations from
		// running indefinitely.
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// Get recordPerPage from query, default to 10
		recordPerPage, err := strconv.Atoi(c.Query("recordPerPage"))
		if err != nil || recordPerPage < 1 {
			recordPerPage = 10
		}

		// Get page number from query, default to 1
		page, err := strconv.Atoi(c.Query("page"))
		if err != nil || page < 1 {
			page = 1
		}

		// Calculate startIndex
		startIndex := (page - 1) * recordPerPage

		// Get startIndex from query if provided
		queryStartIndex := c.Query("startIndex")
		if queryStartIndex != "" {
			startIndex, _ = strconv.Atoi(queryStartIndex)
		}

		// Aggregation pipeline
		matchStage := bson.D{{Key: "$match", Value: bson.D{}}}

		groupStage := bson.D{
			{Key: "$group", Value: bson.D{
				{Key: "_id", Value: "null"},
				{Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}},
				{Key: "data", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}},
			}},
		}

		projectStage := bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "total_count", Value: 1},
				{Key: "food_items", Value: bson.D{
					{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}},
				}},
			}},
		}

		// Execute the aggregation pipeline
		result, err := foodCollection.Aggregate(ctx, mongo.Pipeline{matchStage, groupStage, projectStage})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error occurred while listing items"})
			return
		}

		// Decode results into a slice
		var allFood []bson.M
		if err = result.All(ctx, &allFood); err != nil {
			log.Fatal(err)
		}

		// Return response
		if len(allFood) > 0 {
			c.JSON(http.StatusOK, allFood[0])
		} else {
			c.JSON(http.StatusOK, gin.H{"message": "No food items found"})
		}
	}
}

// returns food

func GetFood() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Creates a context.Context with a 100-second timeout.
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		// Retrieves food_id from the URL path
		foodId := c.Param("food_id")

		var food models.Food

		// mongo query which searches for the document where food_id matches the given foodId
		err := foodCollection.FindOne(ctx, bson.M{"food_id": foodId}).Decode(&food)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error while fetching the food item"})
		}

		// Sending the Food Item as JSON Response
		c.JSON(http.StatusOK, food)
	}
}

func CreateFood() gin.HandlerFunc {
	return func(c *gin.Context) {

		// Creates a context.Context with a 100-second timeout.
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		var menu models.Menu
		var food models.Food

		if err := c.BindJSON(&food); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(food)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		err := menuCollections.FindOne(ctx, bson.M{"menu_id": food.Menu_id}).Decode(&menu)
		defer cancel()
		if err != nil {
			msg := fmt.Sprintf("menu was not found")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		food.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		food.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		food.ID = primitive.NewObjectID()
		food.Food_id = food.ID.Hex()
		var num = toFixed(*food.Price, 2)
		food.Price = &num

		result, insertErr := foodCollection.InsertOne(ctx, food)
		if insertErr != nil {
			msg := fmt.Sprintf("food item was not created")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, result)

	}
}

func round(num float64) int {
	return int(num + math.Copysign(0.5, num))
}

func toFixed(num float64, precision int) float64 {
	output := math.Pow(10, float64(precision))
	return float64(round(num*output)) / output

}

func UpdateFood() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		var menu models.Menu
		var food models.Food

		foodId := c.Param("food_id")

		if err := c.BindJSON(&food); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		var updateObj primitive.D

		if food.Name != nil {
			updateObj = append(updateObj, bson.E{Key: "name", Value: food.Name})
		}

		if food.Price != nil {
			updateObj = append(updateObj, bson.E{Key: "price", Value: food.Price})
		}

		if food.Food_image != nil {
			updateObj = append(updateObj, bson.E{Key: "food_image", Value: food.Food_image})
		}

		if food.Menu_id != nil {
			err := menuCollections.FindOne(ctx, bson.M{"menu_id": food.Menu_id}).Decode(&menu)
			defer cancel()
			if err != nil {
				msg := fmt.Sprintf("message: Menu not found")
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				return
			}
			updateObj = append(updateObj, bson.E{Key: "menu", Value: food.Price})
		}

		food.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		updateObj = append(updateObj, bson.E{Key: "updated_at", Value: food.Updated_at})

		upsert := true
		filter := bson.M{"food_id": foodId}

		opt := options.UpdateOptions{
			Upsert: &upsert,
		}

		result, err := foodCollection.UpdateOne(
			ctx,
			filter,
			bson.D{
				{Key: "$set", Value: updateObj},
			},
			&opt,
		)

		if err != nil {
			msg := fmt.Sprintf("error while updating food")
			c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}
		defer cancel()
		c.JSON(http.StatusOK, result)

	}
}
