package controllers

import (
	// "fmt"
	// "github.com/gin-gonic/gin"
	// "strconv"
	// "time"
)

func Fibonacci(n int) int {
	if n <= 0 {
		return 0
	} else if n == 1 {
		return 1
	} else {
		return Fibonacci(n-1) + Fibonacci(n-2)
	}
}

// type CustomError struct {
// 	Message string
// }

// func FibonacciHandler(c *gin.Context) {
// 	start := time.Now()

// 	nStr := c.Query("n")
// 	n, err := strconv.Atoi(nStr)
// 	if err != nil {
// 		c.AbortWithStatusJSON(500, CustomError{
// 			Message: "Server Error",
// 		})
// 		return
// 	}

// 	// Send the custom error as a response


// 	result := fibonacci(n)
// 	elapsed := time.Since(start)

// 	c.JSON(200, gin.H{
// 		"response": result,
// 	})

// 	fmt.Printf("Go Fibonacci took %s\n", elapsed)
// 	fmt.Printf("Fibonacci(%d) = %d", n, result)
// }
