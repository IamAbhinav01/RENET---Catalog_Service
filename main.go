package main

import (
	"fmt"
	"renet-catalog/app"
)

func main() {

	application:= app.NewApplication()
	err := application.Run()

	if err != nil{
		fmt.Printf("Error occured %v: ",err)
	}
	

}

/* BASIC ROUTER CONNECTION USING GIN

// 	fmt.Printf("Hello")
// 	router := gin.Default()
// 	router.GET("/ping", func(c *gin.Context) {
//     c.JSON(200, gin.H{
//       "message": "pong",
//     })
//   })
//   router.Run()


*/