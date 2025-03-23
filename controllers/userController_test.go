package controllers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rkmangalp/Restaurant_Management/database"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

// ✅ Helper function to create a test router
func setupRouter() *gin.Engine {
	router := gin.Default()
	router.POST("/users/signup", SignUp())
	return router
}

func TestSignupSuccess(t *testing.T) {
	router := setupRouter()

	// ✅ Generate unique email & phone to avoid duplicate entry error
	randomEmail := fmt.Sprintf("testuser%d@example.com", time.Now().UnixNano())
	randomPhone := fmt.Sprintf("98765%d", time.Now().UnixNano()%100000)

	user := map[string]string{
		"first_name": "John",
		"last_name":  "Doe",
		"email":      randomEmail, // Unique email
		"phone":      randomPhone, // Unique phone
		"password":   "SecurePass123",
		"user_type":  "USER",
	}

	// Convert struct to JSON
	jsonValue, _ := json.Marshal(user)
	req, _ := http.NewRequest("POST", "/users/signup", bytes.NewBuffer(jsonValue))
	req.Header.Set("Content-Type", "application/json")

	// Simulate HTTP request
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// ✅ Check response status code
	assert.Equal(t, http.StatusOK, w.Code)

	// ✅ Check response contains success message
	assert.Contains(t, w.Body.String(), "User created sucessfully")

	// ✅ Cleanup: Delete the test user from MongoDB
	db := database.OpenCollection(database.Client, "users")
	_, err := db.DeleteOne(context.TODO(), bson.M{"email": randomEmail})
	if err != nil {
		t.Fatalf("Failed to delete test user: %v", err)
	}
}
