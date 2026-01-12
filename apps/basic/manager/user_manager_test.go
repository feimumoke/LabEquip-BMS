package manager

import (
	"fmt"
	"testing"
)

func TestHashPassword(t *testing.T) {

	password, bmsError := HashPassword("123456", "random_salt")
	fmt.Println(bmsError)
	fmt.Println(password)
}
