package main

import (
	"fmt"

	"github.com/feimumoke/labequipbms/api_idl/pbgenerator"
)

func main() {

	err := pbgenerator.Do()
	if err != nil {
		fmt.Println("Error: " + err.Error())
		return
	}
	fmt.Println("Success")
}
