package controllers

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rkmangalp/Restaurant_Management/database"
	"github.com/rkmangalp/Restaurant_Management/models"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var menuCollections *mongo.Collection = database.OpenCollection(database.Client, "menu")

func GetMenus() gin.HandlerFunc {
	return func(c *gin.Context) {

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)

		result, err := menuCollections.Find(context.TODO(), bson.M{})
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error while listing the menu"})
			return
		}

		var allMenu []bson.M

		if err := result.All(ctx, &allMenu); err != nil {
			log.Fatal(err)
		}

		c.JSON(http.StatusOK, allMenu)
	}
}

func GetMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		menuId := c.Param("menu_id")

		var menu models.Menu

		err := foodCollection.FindOne(ctx, bson.M{"menu_id": menuId}).Decode(&menu)
		defer cancel()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "error while fetching the menu"})
		}
		c.JSON(http.StatusOK, menu)
	}
}

func CreateMenu() gin.HandlerFunc {
	return func(c *gin.Context) {

		var menus []models.Menu
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		// ✅ Bind JSON request body
		if err := c.BindJSON(&menus); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		var insertMenus []gin.H

		// ✅ Iterate through each menu item in the request
		for i := range menus {

			//validate menustruct
			validationErr := validate.Struct(menus[i])
			if validationErr != nil {
				c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
				return
			}

			// ✅ Check if a menu with the same category already exists
			var existingMenu models.Menu

			err := menuCollections.FindOne(ctx, gin.H{"menu_id": menus[i].Menu_id}).Decode(&existingMenu)
			if err == nil {

				// menu exists, assign the existing "menu_id"
				menus[i].Menu_id = existingMenu.Menu_id
			} else {
				// Menu does not exists, create a new one
				menus[i].ID = primitive.NewObjectID()
				menus[i].Menu_id = menus[i].ID.Hex()
				menus[i].Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
				menus[i].Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

				// Append assigned menu details
				_, insertErr := menuCollections.InsertOne(ctx, menus[i])
				if insertErr != nil {
					c.JSON(http.StatusInternalServerError, gin.H{"error": "error while instering"})
					return
				}
			}

			// Append assigned menu details
			insertMenus = append(insertMenus, gin.H{
				"category": menus[i].Category,
				"menu_id":  menus[i].Menu_id,
			})

		}

		// Return response with assigned `menu_id`s
		c.JSON(http.StatusOK, gin.H{
			"message": "Menu processed successfully",
			"menus":   insertMenus,
		})
	}
}

func inTimeSpan(start, end, check time.Time) bool {
	return start.After(time.Now()) && end.After(time.Now())
}

func UpdateMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var menu models.Menu

		if err := c.BindJSON(&menu); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		menuId := c.Param("menu_id")
		filter := bson.M{"menu_id": menuId}

		var updateObj primitive.D

		if menu.Start_date != nil && menu.End_date != nil {
			if !inTimeSpan(*menu.Start_date, *menu.End_date, time.Now()) {
				msg := "kindly retype the time"
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
				defer cancel()
				return
			}

			updateObj = append(updateObj, bson.E{Key: "start_date", Value: menu.Start_date})
			updateObj = append(updateObj, bson.E{Key: "end_date", Value: menu.End_date})

			if menu.Name != "" {
				updateObj = append(updateObj, bson.E{Key: "name", Value: menu.Name})
			}

			if menu.Category != "" {
				updateObj = append(updateObj, bson.E{Key: "category", Value: menu.Category})
			}

			menu.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
			updateObj = append(updateObj, bson.E{Key: "updated_at", Value: menu.Updated_at})

			upsert := true

			opt := options.UpdateOptions{
				Upsert: &upsert,
			}

			result, err := menuCollections.UpdateOne(
				ctx,
				filter,
				bson.D{
					{Key: "$set", Value: updateObj},
				},
				&opt,
			)
			if err != nil {
				msg := "error while updating menu"
				c.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			}

			defer cancel()
			c.JSON(http.StatusOK, result)
		}

	}
}

func DeleteMenu() gin.HandlerFunc {
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		menuID := c.Param("menu_id")

		// ✅ Check if menu exists before deleting
		var menu models.Menu
		err := menuCollections.FindOne(ctx, bson.M{"menu_id": menuID}).Decode(&menu)
		if err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "menu not found"})
			return
		}

		// ✅ Delete associated food items
		_, err = menuCollections.DeleteOne(ctx, bson.M{"menu_id": menuID})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete menu"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"message": "Menu deleted successfully"})

	}
}
