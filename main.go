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

