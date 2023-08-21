// package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Person represents a person's details
type Person struct {
	Name    string `json:"name"`
	Age     int    `json:"age"`
	Address string `json:"address,omitempty"` // The "omitempty" tag omits the field if it's empty
}

var Array [][]byte

func main() {
	person1 := Person{
		Name:    "John Doe",
		Age:     30,
		Address: "123 Main St",
	}
	person2 := Person{
		Name:    "John wick",
		Age:     40,
		Address: "1000 Main St",
	}
	file, _ := os.Create("data.json")
	// Marshal the Person struct into JSON
	jsonData1, err := json.Marshal(person1)
	jsonData2, err := json.Marshal(person2)

	Array = append(Array, jsonData1)
	Array = append(Array, jsonData2)

	fmt.Println(Array)
	file.Write(jsonData1)

	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	fmt.Println(string(Array[0]),"and: ",string(Array[1]))
}
