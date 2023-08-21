// package main

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
	// "net/http"
)

func main() {
	r := gin.Default()

	r.Use(Middleware1)
	r.Use(Middleware2)

	r.GET("/hello", RouteHandler)

	r.Run(":8080")
}

func Middleware1(c *gin.Context) {
	time.Sleep(2 * time.Second)
	fmt.Println("Middleware 1 before Next()")
	c.Next() // Passing control to the next middleware or route handler

	
	errMeta := c.Errors[0].Err.Error()
	if errMeta != "403" {
		println("hiiiiiiiii allllll")
	}
	fmt.Println("c.Errors.Errors(): ", errMeta)
	time.Sleep(2 * time.Second)
	fmt.Println("Middleware 1 after Next()")
}

func Middleware2(c *gin.Context) {
	time.Sleep(1 * time.Second)
	fmt.Println("Middleware 2 before Next()")

	// Simulate an error

	c.Error(fmt.Errorf("an example error from Middleware2")).SetType(404)
	time.Sleep(1 * time.Second)
	fmt.Println("hiiiiiii")
	c.Next()
	time.Sleep(2 * time.Second)
	fmt.Println("Middleware 2 after Next()")
}

func RouteHandler(c *gin.Context) {
	time.Sleep(2 * time.Second)
	fmt.Println("Route Handler before Next()")
	c.Next()
	time.Sleep(2 * time.Second)
	fmt.Println("Route Handler after Next()")
}

// Output when accessing /hello:
// Middleware 1 before Next()
// Middleware 2 before Next()
// Route Handler before Next()
// Error occurred: an example error from Middleware2
// Middleware 1 after Next()
// Middleware 2 after Next()
// Route Handler after Next()
