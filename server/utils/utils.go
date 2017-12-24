package utils

import (
	"encoding/json"
	"fmt"
	"math"
	"strconv"
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
	return arr[len(arr)-1]
}

func Str2flo(arg string) float64 {
	r, err := strconv.ParseFloat(arg, 64)
	HandleError(err)
	return r
}

func Flo2str(flo float64) string {
	return strconv.FormatFloat(flo, 'f', 6, 64)
}

func HandleError(err error) {
	if err != nil {
		fmt.Println("Trading error: ", err)
		panic(err)
	}
}

func ARange(start, stop, step float64) []int64 {
	N := int(math.Ceil((stop - start + step) / step))
	rnge := make([]int64, N, N)
	i := 0
	for x := start; x <= stop; x += step {
		rnge[i] = int64(x)
		i += 1
	}
	return rnge
}

type BySuperTestResult []map[string]string

func (a BySuperTestResult) Len() int      { return len(a) }
func (a BySuperTestResult) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a BySuperTestResult) Less(i, j int) bool {
	return a[i]["superTestResult"] < a[j]["superTestResult"]
}

func CopyMapString(originalMap map[string]string) map[string]string {
	targetMap := make(map[string]string)
	for key, value := range originalMap {
		targetMap[key] = value
	}
	return targetMap
}

func PrependStringToArray(prepend string, arr []string) []string {
	targetArr := []string{prepend}
	for _, value := range arr {
		targetArr = append(targetArr, value)
	}
	return targetArr
}

func CopyMapFloat(originalMap map[string]float64) map[string]float64 {
	targetMap := make(map[string]float64)
	for key, value := range originalMap {
		targetMap[key] = value
	}
	return targetMap
}

func MapStringToJson(originalMap map[string]string) []byte {
	jsonResponse, _ := json.Marshal(originalMap)
	return jsonResponse
}
