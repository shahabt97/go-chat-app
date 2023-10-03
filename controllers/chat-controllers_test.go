package controllers

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
)

func TestHomePageHandler(t *testing.T) {
	// Create a new Gin router and a recorder for HTTP responses.
	router := gin.Default()
	recorder := httptest.NewRecorder()

	// Create a mock session.
	mockSession := &sessions.Session{Values: make(map[interface{}]interface{})}
	mockSession.Values["username"] = "shahab"
	// Define a handler function that sets the mock session in the Gin context.
	router.Use(func(c *gin.Context) {
		c.Set("session", mockSession)
	})

	// Register the route using your HomePageHandler.
	router.GET("/te", HomePageHandler) // Register the "/" route here

	// Create a test request.
	req, err := http.NewRequest("GET", "/te", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Perform the request.
	router.ServeHTTP(recorder, req)


	fmt.Println("this is the ressssss: ", recorder.Body.String())


	// Check the response status code.
	if recorder.Code != http.StatusFound {
		t.Errorf("Expected status code %d, but got %d", http.StatusOK, recorder.Code)
	}

	
		// Check if the response body contains the expected content (login.html in your case).
		expectedBody := "Your Expected HTML Content" // Replace with the expected content
		if body := recorder.Body.String(); body != expectedBody {
			t.Errorf("Expected response body to contain '%s', but got '%s'", expectedBody, body)
		}
	}
func TestLoginHandler(t *testing.T) {

	router := gin.Default()
	recorder := httptest.NewRecorder()

	// Register the route using your HomePageHandler.
	router.POST("/login", LoginHandler) // Register the "/" route here

	jsData := `{
		"username": "shahab",
		"password": "1234"
	  }`

	// Create a test request.
	req, err := http.NewRequest("POST", "/login", strings.NewReader(jsData))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}


	req.Header.Set("Content-Type", "application/json")

	// Perform the request.
	router.ServeHTTP(recorder, req)

	// Check the response status code.
	if recorder.Code != http.StatusCreated {
		t.Errorf("Expected status code %d, but got %d", http.StatusCreated, recorder.Code)
	}

	// t.Errorf()

	fmt.Println("this is the res: ", recorder.Body.String())

}
