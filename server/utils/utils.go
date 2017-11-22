package utils

import (
	"strconv"
	"fmt"
)

func StringInSlice(a string, list []string) bool {
    for _, b := range list {
        if b == a {
            return true
        }
    }
    return false
}

func LastFloat(arr []float64) float64 {
	return arr[len(arr) - 1]
}

func Str2flo(arg string) float64 {
	r, err := strconv.ParseFloat(arg, 64)
	HandleError(err)
	return r
}


func HandleError(err error) {
	if err != nil {
		fmt.Println("Trading error: ", err)
		panic(err)
	}
}

