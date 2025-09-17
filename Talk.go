package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("Chat started.")

	if len(os.Args) > 1 {
		for i := 1; i < len(os.Args); i++ {
			fmt.Println(os.Args[i])
		}
	}
}
