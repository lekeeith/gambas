package gambas

import (
	"encoding/csv"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
)

func checkTypeIntegrity(data []interface{}) (string, error) {
	determinant := 0
	isBool := 0
	isInt := 0
	isFloat64 := 0
	isString := 0
	dtype := ""

	emptyValLocations := make([]int, 0)
	for i, d := range data {
		if d == "" {
			emptyValLocations = append(emptyValLocations, i)
			continue
		}
		switch d.(type) {
		case bool:
			isBool = 1
		case int:
			isInt = 2
		case float64:
			isFloat64 = 4
		case string:
			isString = 8
		}
	}

	determinant = isBool + isInt + isFloat64 + isString

	switch determinant {
	case 1:
		dtype = "bool"
	case 2:
		if len(emptyValLocations) > 0 {
			dtype = "float64"
		} else {
			dtype = "int"
		}
	case 4:
		dtype = "float64"
	case 8:
		dtype = "string"
	case 6:
		dtype = "float64"
	case 0:
		if len(emptyValLocations) == len(data) {
			dtype = "string"
		} else {
			return "", fmt.Errorf("invalid data type; data type should be either bool, int, float64, or string")
		}
	default:
		dtype = "string"
	}

	return dtype, nil

	// isFloat64 := 0
	// isString := 0
	// isNil := 0
	// for _, v := range data {
	// 	switch t := v.(type) {
	// 	case float64:
	// 		if math.IsNaN(t) {
	// 			continue
	// 		}
	// 		isFloat64 = 1
	// 	case string:
	// 		isString = 1
	// 	case nil:
	// 		isNil = 1
	// 	default:
	// 		_, err := i2f(v)
	// 		if err != nil {
	// 			return false, fmt.Errorf("invalid type: %T", t)
	// 		} else {
	// 			isFloat64 = 1
	// 		}
	// 	}

	// 	if isFloat64+isString+isNil > 1 {
	// 		return false, nil
	// 	} else if isFloat64+isString+isNil == 0 {
	// 		panic("type not detected")
	// 	}
	// }

	// return true, nil
}

func i2f(data interface{}) (float64, error) {
	var x float64
	switch v := data.(type) {
	case int:
		x = float64(v)
	case int8:
		x = float64(v)
	case int16:
		x = float64(v)
	case int32:
		x = float64(v)
	case int64:
		x = float64(v)
	case uint:
		x = float64(v)
	case uint8:
		x = float64(v)
	case uint16:
		x = float64(v)
	case uint32:
		x = float64(v)
	case uint64:
		x = float64(v)
	case float32:
		x = float64(v)
	case float64:
		x = v
	default:
		return 0.0, fmt.Errorf("%v is not a number", data)
	}

	return x, nil
}

// tryBool checks if a string can be converted into bool.
// tryBool only accepts "TRUE", "True", "true", and "FALSE", "False", "false".
func tryBool(data string) (bool, error) {
	ignored := []string{"1", "t", "T", "0", "f", "F"}
	if containsString(ignored, data) {
		return false, fmt.Errorf("ignored string")
	}

	b, err := strconv.ParseBool(data)
	if err != nil {
		return false, err
	}
	return b, nil
}

// tryInt checks if a string can be converted into int.
func tryInt(data string) (int, error) {
	s, err := strconv.Atoi(data)
	if err != nil {
		return 0, err
	}
	return s, nil
}

// tryFloat64 checks if a string can be converted into float64.
func tryFloat64(data string) (float64, error) {
	if data == "" || data == "NaN" {
		return math.NaN(), nil
	}
	f, err := strconv.ParseFloat(data, 64)
	if err != nil {
		return 0, err
	}
	return f, nil
}

// tryDataType accepts a string and tries to convert it to the correct data type.
// It will try to convert the data into a bool, then int, then float64, and finally string.
func tryDataType(data string) interface{} {
	b, err := tryBool(data)
	if err != nil {
		i, err := tryInt(data)
		if err != nil {
			f, err := tryFloat64(data)
			if err != nil {
				return data
			}
			return f
		}
		return i
	}
	return b
}

// checkType checks to see if the data can be represented as a float64.
// Because CSV is read as an array of strings, there has to be a way to check the type.
func checkCSVDataType(data string) interface{} {
	if data == "" {
		return math.NaN()
	}
	v, ok := strconv.ParseFloat(data, 64)
	switch ok {
	case nil:
		return v
	default:
		return data
	}
}

// checkJsonDataType checks for JSON null data, and converts it into math.NaN().
func checkJsonDataType(data interface{}) interface{} {
	if data == nil {
		return math.NaN()
	}
	return data
}

// consolidateToFloat64 consolidates all data in an []interface{} to float64.
// This is necessary to convert empty string values into math.NaN().
// In order to stay compatible with Series.data,
// the data type of the slice is still an empty interface.
func consolidateToFloat64(data []interface{}) []interface{} {
	result := make([]interface{}, len(data))
	for i, d := range data {
		switch dd := d.(type) {
		case float64:
			result[i] = dd
		case int:
			result[i] = float64(dd)
		case string:
			if dd == "" || dd == "NaN" {
				result[i] = math.NaN()
			} else {
				f, err := strconv.ParseFloat(fmt.Sprint(d), 64)
				if err != nil {
					panic("consolidateToFloat64 should be called only when data is in the form of float64")
				}
				result[i] = f
			}
		}
	}

	return result
}

// consolidateToString consolidates all data in an []interface{} to string.
// In order to stay compatible with Series.data,
// the data type of the slice is still an empty interface.
func consolidateToString(data []interface{}) []interface{} {
	result := make([]interface{}, len(data))
	for i, d := range data {
		if d == "" || d == "NaN" {
			result[i] = math.NaN()
		} else if conv, ok := d.(float64); ok && math.IsNaN(conv) {
			result[i] = math.NaN()
		} else {
			result[i] = fmt.Sprint(d)
		}
	}

	return result
}

// interface2F64Data() converts a slice of interface{} into F64Data.
func interface2F64Slice(data []interface{}) ([]float64, error) {
	fd := make([]float64, 0)
	for _, v := range data {
		switch converted := v.(type) {
		case float64:
			if math.IsNaN(converted) {
				continue
			}
			fd = append(fd, converted)
		// case int:
		// 	fd = append(fd, float64(converted))
		default:
			return nil, fmt.Errorf("data is not a float64: %v", v)
		}
	}

	return fd, nil
}

// interface2StringData() converts a slice of interface{} into StringData.
func interface2StringSlice(data []interface{}) ([]string, error) {
	sd := make([]string, 0)
	for _, v := range data {
		switch converted := v.(type) {
		case string:
			sd = append(sd, converted)
		default:
			return nil, fmt.Errorf("data is not a string: %v", v)
		}
	}

	return sd, nil
}

// slicesAreEqual checks whether two slices are equal.
func slicesAreEqual(slice1, slice2 []interface{}) bool {
	if len(slice1) != len(slice2) {
		return false
	}
	for i, v := range slice1 {
		if v != slice2[i] {
			return false
		}
	}
	return true
}

// containsString checks whether a string exists in a slice of strings.
func containsString(strSlice []string, str string) bool {
	for _, data := range strSlice {
		if data == str {
			return true
		}
	}
	return false
}

// containsIndex checks whether an Index object exists in a slice of Index objects.
// For a lenient comaprison that checks for values only, use containsIndexWithoutId.
func containsIndex(indexSlice []Index, index Index) bool {
	for _, data := range indexSlice {
		if slicesAreEqual(data.value, index.value) && (data.id == index.id) {
			return true
		}
	}
	return false
}

// containsIndex checks whether an Index object exists in a slice of Index objects.
// For a strict comaprison that checks for id as well, use containsIndex.
func containsIndexWithoutId(indexSlice []Index, index Index) bool {
	for _, data := range indexSlice {
		if slicesAreEqual(data.value, index.value) {
			return true
		}
	}
	return false
}

// containsSlice checks whether a slice of interface{} exists in a slice of []interface{}.
func containsSlice(s1 [][]interface{}, s2 []interface{}) bool {
	for _, data := range s1 {
		if slicesAreEqual(data, s2) {
			return true
		}
	}
	return false
}

// Summary statistics functions (internal use only)

// median() returns the median of the elements in an array.
func median(data []float64) (float64, error) {
	median := 0.0
	sort.Float64s(data)

	total := len(data)
	if total == 0 {
		return math.NaN(), fmt.Errorf("no elements in this column")
	}
	if total%2 == 0 {
		lower := data[total/2-1]
		upper := data[total/2]

		median = (lower + upper) / 2
	} else {
		median := data[(total+1)/2-1]
		return median, nil
	}

	return median, nil
}

// copyDf takes a source DataFrame and returns a copy of it with different memory address.
func copyDf(src *DataFrame) DataFrame {
	newDf := new(DataFrame)
	newDf.series = make([]Series, len(src.series))
	for i, ser := range src.series {
		newDf.series[i].data = append(newDf.series[i].data, ser.data...)
		newDf.series[i].index.index = append(newDf.series[i].index.index, ser.index.index...)
		newDf.series[i].index.names = append(newDf.series[i].index.names, ser.index.names...)
		newDf.series[i].name = ser.name
		newDf.series[i].dtype = ser.dtype
	}
	newDf.index.index = append(newDf.index.index, src.index.index...)
	newDf.index.names = append(newDf.index.names, src.index.names...)
	newDf.columns = append(newDf.columns, src.columns...)

	return *newDf
}

// readCsvColIntoData extracts a column in a CSV file to a [][]interface{}.
func readCsvColIntoData(filepath string, col string) ([][]interface{}, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	csvr := csv.NewReader(f)
	rows, err := csvr.ReadAll()
	if err != nil {
		return nil, err
	}

	res := make([][]interface{}, 0)
	colIndex := 0
	for i, colLabel := range rows[0] {
		if colLabel == col {
			colIndex = i
			break
		}
	}
	for i, row := range rows {
		if i == 0 {
			continue
		}
		res = append(res, []interface{}{row[colIndex]})
	}

	return res, nil
}

// generateAlphabets will generate alphabets based on i.
// This is mostly used to dynamically create column labels for excel sheets.
// For example, if i=1 then it will return A.
// i=25, Z. i=26, AA. And vice versa...
func generateAlphabets(i int) string {
	i--
	result := ""
	if first := i / 26; first > 0 {
		result += generateAlphabets(first)
		result += string(rune('A' + i%26))
	} else {
		result += string(rune('A' + i))
	}

	return result
}

func quickSelect(arr []float64, lo int, hi int, k int) float64 {
	if lo == hi {
		return arr[k]
	}

	pivotIndex := hoarePartition(arr, lo, hi)

	if k < pivotIndex {
		return quickSelect(arr, lo, pivotIndex, k)
	} else if k > pivotIndex {
		return quickSelect(arr, pivotIndex+1, hi, k)
	} else {
		return arr[k]
	}
}

func hoarePartition(arr []float64, lo int, hi int) int {
	pivot := arr[(lo+hi)/2]
	i := lo - 1
	j := hi + 1

	for {
		for ok := true; ok; ok = arr[i] < pivot {
			i++
		}
		for ok := true; ok; ok = arr[j] > pivot {
			j--
		}

		if i >= j {
			return j
		}

		arr[i], arr[j] = arr[j], arr[i]
	}
}
