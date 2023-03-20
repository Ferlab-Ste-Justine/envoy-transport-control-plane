package utils

import (
	"fmt"
	"os"
)

func AbortOnErr(err error) {
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}