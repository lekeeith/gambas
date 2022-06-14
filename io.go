package gambas

import (
	"bufio"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
)

// ReadCsv reads a CSV file and returns a new DataFrame object.
// It is recommended to generate pathToFile using `filepath.Join`.
// TODO: users should be able to define custom indices.
func ReadCsv(pathToFile string, indexCols []string) (*DataFrame, error) {
	// read line by line
	f, err := os.Open(pathToFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	csvr := csv.NewReader(f)

	rowNum := 0
	columnArray := make([]string, 0)
	data2DArray := make([][]interface{}, 0)
	for {
		row, err := csvr.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Fatal(err)
		}
		// first line is column name
		if rowNum == 0 {
			// add to columnArray
			columnArray = append(columnArray, row...)
			rowNum++
			continue
		}
		// second line onwards is the actual data
		for i, v := range row {
			// add to data2DArray
			if len(data2DArray) < len(row) {
				data2DArray = append(data2DArray, make([]interface{}, 0))
			}
			// each data should be checked to see what type it is
			vChecked := checkCSVDataType(v)
			data2DArray[i] = append(data2DArray[i], vChecked)
		}
		rowNum++
	}
	// create new DataFrame object and return it
	if indexCols != nil {
		df, err := NewDataFrame(data2DArray, columnArray, indexCols)
		if err != nil {
			return nil, err
		}

		return df, nil
	}
	df, err := NewDataFrame(data2DArray, columnArray, []string{columnArray[0]})
	if err != nil {
		return nil, err
	}

	return df, nil
}

// WriteCsv writes a DataFrame object to CSV file.
// It is recommended to generate pathToFile using `filepath.Join`.
func WriteCsv(df DataFrame, pathToFile string) (os.FileInfo, error) {
	f, err := os.Create(pathToFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	// write column names in the first row
	for i, col := range df.columns {
		_, err := w.WriteString(col)
		if err != nil {
			return nil, err
		}

		if i+1 != len(df.columns) {
			_, err := w.WriteString(",")
			if err != nil {
				return nil, err
			}
		}
	}
	w.WriteString("\n")

	// write the data in the following rows
	for i := range df.series[0].data {
		for j, ser := range df.series {
			_, err := w.WriteString(fmt.Sprint(ser.data[i]))
			if err != nil {
				return nil, err
			}

			if j+1 != len(df.series) {
				_, err := w.WriteString(",")
				if err != nil {
					return nil, err
				}
			}
		}

		w.WriteString("\n")
	}

	w.Flush()

	info, err := os.Stat(pathToFile)
	if err != nil {
		return nil, err
	}
	return info, nil
}

// ReadJson reads a JSON file and returns a new DataFrame object.
// It is recommended to generate pathToFile using `filepath.Join`.
// The JSON file should be in this format:
// {"col1":[val1, val2, ...], "col2":[val1, val2, ...], ...}
// You can either set a column to be the index, or set it as nil.
// If nil, a new RangeIndex will be created.
// Your index column should not have any missing values.
func ReadJsonByColumns(pathToFile string, indexCols []string) (*DataFrame, error) {
	f, err := os.Open(pathToFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fbyte, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var decoded map[string]interface{}
	err = json.Unmarshal(fbyte, &decoded)
	if err != nil {
		return nil, err
	}

	newDfData := make([][]interface{}, 0)
	newDfCols := make([]string, 0)
	newDfIndexNames := make([]string, 0)

	for col, colData := range decoded {
		newDfCols = append(newDfCols, col)
		colDataAsserted := colData.([]interface{})
		for i, cda := range colDataAsserted {
			checked := checkJsonDataType(cda)
			colDataAsserted[i] = checked
		}
		newDfData = append(newDfData, colDataAsserted)

		if indexCols != nil && containsString(indexCols, col) {
			newDfIndexNames = append(newDfIndexNames, col)
		}
	}

	if indexCols == nil {
		newDfIndexNames = nil
	}

	newDf, err := NewDataFrame(newDfData, newDfCols, newDfIndexNames)
	if err != nil {
		return nil, err
	}
	newDf.SortByIndex(true)
	return newDf, nil
}

// ReadJsonByRows reads a JSON file and returns a new DataFrame object.
// The JSON file should be in this format:
// {"index1":{"col1":val1, "col2":val2, ...}, "index2":{}...}
func ReadJsonByRows(pathToFile string) (*DataFrame, error) {
	f, err := os.Open(pathToFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	fbyte, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	var decoded map[string]interface{}
	err = json.Unmarshal(fbyte, &decoded)
	if err != nil {
		return nil, err
	}

	newDfData := make([][]interface{}, 0)
	newDfCols := make([]string, 0)
	newDfCols = append(newDfCols, "index")

	for index, rowData := range decoded {
		rowDataAsserted := rowData.(map[string]interface{})
		d := make([][]interface{}, len(rowDataAsserted)+1)
		d[0] = append(d[0], index)

		count := 1
		for col, val := range rowDataAsserted {
			if !containsString(newDfCols, col) {
				newDfCols = append(newDfCols, col)
			}
			d[count] = append(d[count], val)
		}
		newDfData = d
	}

	newDf, err := NewDataFrame(newDfData, newDfCols, []string{"index"})
	if err != nil {
		return nil, err
	}
	return newDf, nil
}

// ReadJsonStream reads a JSON stream and returns a new DataFrame object.
// The JSON file should be in this format:
// {"col1":val1, "col2":val2, ...}{"col1":val1, "col2":val2, ...}
func ReadJsonStream(pathToFile string, indexCols []string) (*DataFrame, error) {
	f, err := os.Open(pathToFile)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dec := json.NewDecoder(f)
	newDfCols := make([]string, 0)
	colData := make(map[string][]interface{}, 0)

	for {
		var row map[string]interface{}
		err := dec.Decode(&row)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, err
		}

		for k, v := range row {
			colData[k] = append(colData[k], v)
		}
	}

	newDfData := make([][]interface{}, 0)

	for k, v := range colData {
		newDfCols = append(newDfCols, k)
		newDfData = append(newDfData, v)
	}

	newDf, err := NewDataFrame(newDfData, newDfCols, indexCols)
	if err != nil {
		return nil, err
	}
	return newDf, nil
}
