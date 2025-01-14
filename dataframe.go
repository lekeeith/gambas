package gambas

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"text/tabwriter"
)

// DataFrame type represents a 2D tabular dataset.
// A DataFrame object is comprised of multiple Series objects.
type DataFrame struct {
	series  []Series
	index   IndexData
	columns []string
}

func (df DataFrame) Series() []Series {
	return df.series
}

func (df DataFrame) Index() IndexData {
	return df.index
}
func (df DataFrame) Columns() []string {
	return df.columns
}

// MarshalJSON is used to implement the json.Marshaler interface{}.
func (df *DataFrame) MarshalJSON() ([]byte, error) {
	type serJson struct {
		Data  []interface{} `json:"data"`
		Name  string        `json:"name"`
		Dtype string        `json:"dtype"`
	}
	type dfJson struct {
		Series  []serJson `json:"series"`
		Columns []string  `json:"columns"`
	}

	serjs := make([]serJson, 0)
	for _, ser := range df.series {
		serj := new(serJson)
		for _, data := range ser.data {
			switch v := data.(type) {
			case float64:
				if math.IsNaN(v) {
					serj.Data = append(serj.Data, nil)
				} else {
					serj.Data = append(serj.Data, data)
				}
			default:
				serj.Data = append(serj.Data, data)
			}
		}
		serj.Name = ser.name
		serj.Dtype = ser.dtype

		serjs = append(serjs, *serj)
	}

	dfj := new(dfJson)
	dfj.Series = append(dfj.Series, serjs...)
	dfj.Columns = append(dfj.Columns, df.columns...)
	return json.Marshal(dfj)
}

// Print prints all data in a DataFrame object.
func (df *DataFrame) Print() {
	w := new(tabwriter.Writer)

	w.Init(os.Stdout, 5, 0, 4, ' ', 0)

	for i := range df.index.names {
		fmt.Fprint(w, df.index.names[i], "\t")
	}

	fmt.Fprint(w, "|", "\t")

	for i := range df.columns {
		fmt.Fprint(w, df.columns[i], "\t")
	}
	fmt.Fprintln(w)

	for i := 0; i < len(df.series[0].data); i++ {
		if len(df.index.index[i].value) > 1 {
			for j := range df.index.index[i].value {
				fmt.Fprint(w, df.index.index[i].value[j], "\t")
			}
		} else {
			fmt.Fprint(w, df.index.index[i].value[0], "\t")
		}

		fmt.Fprint(w, "|", "\t")

		for j := range df.columns {
			fmt.Fprint(w, df.series[j].data[i], "\t")
		}
		fmt.Fprintln(w)
	}
	w.Flush()
}

// PrintRange prints data in a DataFrame object at a given range.
// Index starts at 0.
func (df *DataFrame) PrintRange(start, end int) {
	w := new(tabwriter.Writer)

	w.Init(os.Stdout, 5, 0, 4, ' ', 0)

	for i := range df.index.names {
		fmt.Fprint(w, df.index.names[i], "\t")
	}

	fmt.Fprint(w, "|", "\t")

	for i := range df.columns {
		fmt.Fprint(w, df.columns[i], "\t")
	}
	fmt.Fprintln(w)

	for i := start; i < end; i++ {
		if len(df.index.index[i].value) > 1 {
			for j := range df.index.index[i].value {
				fmt.Fprint(w, df.index.index[i].value[j], "\t")
			}
		} else {
			fmt.Fprint(w, df.index.index[i].value[0], "\t")
		}

		fmt.Fprint(w, "|", "\t")

		for j := range df.columns {
			fmt.Fprint(w, df.series[j].data[i], "\t")
		}
		fmt.Fprintln(w)
	}
	w.Flush()
}

// Head prints the first howMany items in a DataFrame object.
func (df *DataFrame) Head(howMany int) {
	df.PrintRange(0, howMany)
}

// Tail prints the last howMany items in a DataFrame object.
func (df *DataFrame) Tail(howMany int) {
	df.PrintRange(len(df.series[0].data)-howMany, len(df.series[0].data))
}

// LocRows returns a set of rows as a new DataFrame object, given a list of labels.
// You are only allowed to pass in the indices of the DataFrame as rows.
func (df *DataFrame) LocRows(rows ...[]interface{}) (DataFrame, error) {
	filteredData := make([][]interface{}, 0)
	filteredColname := make([]string, 0)
	filteredIndex := IndexData{}
	for _, series := range df.series {
		located, err := series.Loc(rows...)
		if err != nil {
			return DataFrame{}, err
		}
		filteredData = append(filteredData, located.data)
		filteredColname = append(filteredColname, located.name)
		filteredIndex = located.index
	}

	dataframe, err := NewDataFrame(filteredData, filteredColname, filteredIndex.names)
	if err != nil {
		return DataFrame{}, err
	}

	// When NewDataFrame is called, the resulting dataframe may have empty index values.
	// This is because NewDataFrame searches for index values in filtered2D,
	// but if the columns in the dataframe does not match filteredIndex.names,
	// there would be no matching columns, thus returning empty indexes.
	dataframe.index = filteredIndex
	for i := range dataframe.series {
		dataframe.series[i].index = filteredIndex
	}

	return dataframe, nil
}

// LocRowsItems will return a slice of rows.
// You are only allowed to pass in indices of the DataFrame as rows.
// Use this over LocRows if you want to extract the items directly
// instead of getting a DataFrame object.
func (df *DataFrame) LocRowsItems(rows ...[]interface{}) ([][]interface{}, error) {
	filteredRows := make([][]interface{}, len(rows))
	for i := 0; i < len(rows); i++ {
		filteredRows[i] = make([]interface{}, 0)
	}

	for _, series := range df.series {
		located, err := series.LocItems(rows...)
		if err != nil {
			return nil, err
		}
		for i := range located {
			filteredRows[i] = append(filteredRows[i], located[i])
		}
	}

	return filteredRows, nil
}

// LocCol returns a column as a new Series object.
func (df *DataFrame) LocCol(col string) (Series, error) {
	for _, series := range df.series {
		if series.name == col {
			ser, err := NewSeries(series.data, series.name, &series.index)
			if err != nil {
				return Series{}, err
			}

			return ser, nil
		}
	}

	return Series{}, fmt.Errorf("column '%v' does not exist", col)
}

// LocCols returns a set of columns as a new DataFrame object, given a list of labels.
func (df *DataFrame) LocCols(cols ...string) (DataFrame, error) {
	filtered2D := make([][]interface{}, 0)
	for _, column := range cols {
		for _, series := range df.series {
			if series.name == column {
				filtered2D = append(filtered2D, series.data)
			}
		}
	}

	dataframe, err := NewDataFrame(filtered2D, cols, df.index.names)
	if err != nil {
		return DataFrame{}, err
	}

	// When NewDataFrame is called, the resulting dataframe may have empty index values.
	// This is because NewDataFrame searches for index values in filtered2D,
	// but if the index column name is different from the column the user is trying to LocCols,
	// there would be no matching columns.

	copy(dataframe.index.index, df.index.index)

	for _, ser := range dataframe.series {
		copy(ser.index.index, df.index.index)
	}

	return dataframe, nil
}

// LocColsItems will return a slice of columns.
// Use this over LocCols if you want to extract the items directly
// instead of getting a DataFrame object.
func (df *DataFrame) LocColsItems(cols ...string) ([][]interface{}, error) {
	filtered2D := make([][]interface{}, 0)
	for _, column := range cols {
		for _, series := range df.series {
			if series.name == column {
				filtered2D = append(filtered2D, series.data)
			}
		}
	}

	if len(filtered2D) == 0 {
		return nil, fmt.Errorf("no columns found")
	}

	return filtered2D, nil
}

// Loc indexes the DataFrame object given a slice of row and column labels, and returns the result as a new DataFrame object.
// You are only allowed to pass in indices of the DataFrame as rows.
func (df *DataFrame) Loc(cols []string, rows ...[]interface{}) (DataFrame, error) {
	df1, err := df.LocCols(cols...)
	if err != nil {
		return DataFrame{}, err
	}

	df2, err := df1.LocRows(rows...)
	if err != nil {
		return DataFrame{}, err
	}

	return df2, nil
}

/* Basic arithmetic operations for columns. */

// ColAdd adds the given value to each element in the specified column.
func (df *DataFrame) ColAdd(colname string, value float64) (DataFrame, error) {
	newDf := copyDf(df)

	for _, series := range newDf.series {
		if series.name == colname {
			for i, data := range series.data {
				switch v := data.(type) {
				case float64:
					v += value
					series.data[i] = v
				default:
					return DataFrame{}, fmt.Errorf("cannot add, column data type is not float64")
				}
			}
			return newDf, nil
		}
	}
	return DataFrame{}, fmt.Errorf("colname does not match any of the existing column names")
}

// ColSub subtracts the given value from each element in the specified column.
func (df *DataFrame) ColSub(colname string, value float64) (DataFrame, error) {
	newDf := copyDf(df)
	for _, series := range newDf.series {
		if series.name == colname {
			for i, data := range series.data {
				switch v := data.(type) {
				case float64:
					v -= value
					series.data[i] = v
				default:
					return DataFrame{}, fmt.Errorf("cannot subtract, column data type is not float64")
				}
			}
			return newDf, nil
		}
	}
	return DataFrame{}, fmt.Errorf("colname does not match any of the existing column names")
}

// ColMul multiplies each element in the specified column by the given value.
func (df *DataFrame) ColMul(colname string, value float64) (DataFrame, error) {
	newDf := copyDf(df)
	for _, series := range newDf.series {
		if series.name == colname {
			for i, data := range series.data {
				switch v := data.(type) {
				case float64:
					v *= value
					series.data[i] = v
				default:
					return DataFrame{}, fmt.Errorf("cannot multiply, column data type is not float64")
				}
			}
			return newDf, nil
		}
	}
	return DataFrame{}, fmt.Errorf("colname does not match any of the existing column names")
}

// ColDiv divides each element in the specified column by the given value.
func (df *DataFrame) ColDiv(colname string, value float64) (DataFrame, error) {
	newDf := copyDf(df)
	for _, series := range newDf.series {
		if series.name == colname {
			for i, data := range series.data {
				switch v := data.(type) {
				case float64:
					v /= value
					series.data[i] = v
				default:
					return DataFrame{}, fmt.Errorf("cannot divide, column data type is not float64")
				}
			}
			return newDf, nil
		}
	}
	return DataFrame{}, fmt.Errorf("colname does not match any of the existing column names")
}

// ColMod applies modulus calculations on each element in the specified column, returning the remainder.
func (df *DataFrame) ColMod(colname string, value float64) (DataFrame, error) {
	newDf := copyDf(df)
	for _, series := range newDf.series {
		if series.name == colname {
			for i, data := range series.data {
				switch v := data.(type) {
				case float64:
					series.data[i] = math.Mod(v, value)
				default:
					return DataFrame{}, fmt.Errorf("cannot use modulus, column data type is not float64")
				}
			}
			return newDf, nil
		}
	}
	return DataFrame{}, fmt.Errorf("colname does not match any of the existing column names")
}

// Basic boolean operators for columns.

// ColGt checks if each element in the specified column is greater than the given value.
func (df *DataFrame) ColGt(colname string, value float64) (DataFrame, error) {
	newDf := copyDf(df)
	for i, series := range newDf.series {
		if series.name == colname {
			newDf.series[i].dtype = "bool"
			for i, data := range series.data {
				switch v := data.(type) {
				case float64:
					isGt := (v > value)
					series.data[i] = isGt
				default:
					return DataFrame{}, fmt.Errorf("cannot compare, column data type is not float64")
				}
			}
			return newDf, nil
		}
	}
	return DataFrame{}, fmt.Errorf("colname does not match any of the existing column names")
}

// ColLt checks if each element in the specified column is less than the given value.
func (df *DataFrame) ColLt(colname string, value float64) (DataFrame, error) {
	newDf := copyDf(df)
	for i, series := range newDf.series {
		if series.name == colname {
			newDf.series[i].dtype = "bool"
			for i, data := range series.data {
				switch v := data.(type) {
				case float64:
					isLt := (v < value)
					series.data[i] = isLt
				default:
					return DataFrame{}, fmt.Errorf("cannot compare, column data type is not float64")
				}
			}
			return newDf, nil
		}
	}
	return DataFrame{}, fmt.Errorf("colname does not match any of the existing column names")
}

// ColEq checks if each element in the specified column is equal to the given value.
func (df *DataFrame) ColEq(colname string, value float64) (DataFrame, error) {
	newDf := copyDf(df)
	for i, series := range newDf.series {
		if series.name == colname {
			newDf.series[i].dtype = "bool"
			for i, data := range series.data {
				switch v := data.(type) {
				case float64:
					isEq := (v == value)
					series.data[i] = isEq
				default:
					return DataFrame{}, fmt.Errorf("cannot compare, column data type is not float64")
				}
			}
			return newDf, nil
		}
	}
	return DataFrame{}, fmt.Errorf("colname does not match any of the existing column names")
}

/* Editing Properties */

// NewCol creates a new column with the given data and column name.
// To create a blank column, pass in nil.
func (df *DataFrame) NewCol(colname string, data []interface{}) (DataFrame, error) {
	newDf := copyDf(df)
	if data == nil {
		for i := 0; i < len(df.series[0].data); i++ {
			data = append(data, math.NaN())
		}
	}
	newSeries, err := NewSeries(data, colname, &newDf.index)
	if err != nil {
		return DataFrame{}, err
	}

	newDf.series = append(newDf.series, newSeries)
	newDf.columns = append(newDf.columns, colname)

	return newDf, nil
}

// NewDerivedCol creates a new column derived from an existing column.
// It copies over the data from srcCol into a new column.
func (df *DataFrame) NewDerivedCol(colname, srcCol string) (DataFrame, error) {
	newDf := copyDf(df)
	for i := range newDf.series {
		if newDf.series[i].name == srcCol {
			dataframe, err := newDf.NewCol(colname, newDf.series[i].data)
			if err != nil {
				return DataFrame{}, err
			}
			return dataframe, nil
		}
	}
	return DataFrame{}, fmt.Errorf("the column doesn't exist: %s", srcCol)
}

// RenameCol renames columns in a DataFrame.
func (df *DataFrame) RenameCol(colnames map[string]string) error {
	for oldName, newName := range colnames {
		exists := false
		for i, col := range df.columns {
			if col == oldName {
				df.columns[i] = newName
				exists = true
			}
		}
		if !exists {
			return fmt.Errorf("column does not exist: %v", oldName)
		}

		for i, name := range df.index.names {
			if name == oldName {
				df.index.names[i] = newName
			}
		}

		for i, series := range df.series {
			if series.name == oldName {
				df.series[i].RenameCol(newName)
			}
			err := df.series[i].RenameIndex(colnames)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

// DropNaN drops rows or columns with NaN values.
// Specify axis to choose whether to remove rows with NaN or columns with NaN.
// axis=0 is row, axis=1 is column.
func (df *DataFrame) DropNaN(axis int) (DataFrame, error) {
	if axis > 1 || axis < 0 {
		return DataFrame{}, fmt.Errorf("axis can only be either 0 or 1")
	}

	newDf := df

	// for each series, iterate through the series until NaN is found
	// if NaN, save NaN index
	// sort the NaNindex slice

	indexSlice := make([]int, 0)
	seriesHasNaNSlice := make([]bool, len(newDf.series))
	for i, ser := range newDf.series {
		for j, data := range ser.data {
			switch v := data.(type) {
			case string:
				if v == "NaN" {
					indexSlice = append(indexSlice, j)
					seriesHasNaNSlice[i] = true
				}
			case float64:
				if math.IsNaN(v) {
					indexSlice = append(indexSlice, j)
					seriesHasNaNSlice[i] = true
				}
			}
		}
	}

	sort.Ints(indexSlice)

	// deleting rows containing NaN
	// for each series, remove data at the index. length of newDf.series.data will decrease by 1.
	// subtract 1 from each index so that it matches the new length newDf.series.data.
	if axis == 0 {
		for i := range newDf.series {
			iSlice := make([]int, len(indexSlice))
			copy(iSlice, indexSlice)

			for j, index := range iSlice {
				newDf.series[i].index.index = append(newDf.series[i].index.index[:index], newDf.series[i].index.index[index+1:]...)
				newDf.series[i].data = append(newDf.series[i].data[:index], newDf.series[i].data[index+1:]...)

				for k := j + 1; k < len(iSlice); k++ {
					iSlice[k] -= 1
				}
			}
		}
		newDf.index.index = newDf.series[0].index.index
	}

	// deleting columns containing NaN
	// for each series, remove data at the index. length of newDf.series.data will decrease by 1.
	// subtract 1 from each index so that it matches the new length newDf.series.data.
	if axis == 1 {
		for i, hasNaN := range seriesHasNaNSlice {
			if hasNaN {
				newDf.columns = append(newDf.columns[:i], newDf.columns[i+1:]...)
				newDf.series = append(newDf.series[:i], newDf.series[i+1:]...)
				seriesHasNaNSlice = append(seriesHasNaNSlice[:i], seriesHasNaNSlice[i+1:]...)
			}
		}
	}

	return *newDf, nil
}

/* Merging */

// MergeDfsHorizontally merges two DataFrame objects side by side.
// The target DataFrame will always be appended to the right of the source DataFrame.
// Index will reset and become a RangeIndex.
func (df *DataFrame) MergeDfsHorizontally(target DataFrame) (DataFrame, error) {
	newDf := copyDf(df)
	if len(newDf.series[0].data) >= len(target.series[0].data) {
		newDf.index = CreateRangeIndex(len(newDf.series[0].data))
		lenDiff := len(newDf.series[0].data) - len(target.series[0].data)

		// fill missing data in target with NaN
		for i, ser := range target.series {
			for j := 0; j < lenDiff; j++ {
				target.series[i].data = append(target.series[i].data, math.NaN())
			}

			if ser.dtype == "int" {
				target.series[i].data = consolidateToFloat64(target.series[i].data)
				target.series[i].dtype = "float64"
			}
		}
	} else {
		newDf.index = CreateRangeIndex(len(target.series[0].data))
		lenDiff := len(target.series[0].data) - len(newDf.series[0].data)

		// fill missing data in source with NaN
		for i, ser := range newDf.series {
			for j := 0; j < lenDiff; j++ {
				newDf.series[i].data = append(newDf.series[i].data, math.NaN())
			}

			if ser.dtype == "int" {
				newDf.series[i].data = consolidateToFloat64(newDf.series[i].data)
				newDf.series[i].dtype = "float64"
			}
		}
	}

	newDf.columns = append(newDf.columns, target.columns...)
	newDf.series = append(newDf.series, target.series...)

	for i, ser := range newDf.series {
		newDf.series[i].index = CreateRangeIndex(len(ser.data))
	}

	return newDf, nil
}

// MergeDfsVertically stacks two DataFrame objects vertically.
func (df *DataFrame) MergeDfsVertically(target DataFrame) (DataFrame, error) {
	if len(target.columns) != len(df.columns) {
		return DataFrame{}, fmt.Errorf("number of columns is different")
	}
	for i, col := range df.columns {
		if col != target.columns[i] {
			return DataFrame{}, fmt.Errorf("column names do not match")
		}
	}

	newDf := copyDf(df)
	for _, index := range target.index.index {
		newDf.index.index = append(newDf.index.index, Index{len(df.index.index) + index.id, index.value})
	}

	for i := range newDf.series {
		if newDf.series[i].dtype != target.series[i].dtype {
			return DataFrame{}, fmt.Errorf("column dtypes do not match")
		}
		if newDf.series[i].name != target.series[i].name {
			return DataFrame{}, fmt.Errorf("column names do not match")
		}

		newDf.series[i].data = append(newDf.series[i].data, target.series[i].data...)

		for _, index := range target.series[i].index.index {
			newDf.series[i].index.index = append(newDf.series[i].index.index, Index{len(df.series[i].index.index) + index.id, index.value})
		}
	}

	return newDf, nil
}

/* Sorting Functions */

// SortByIndex sorts the items by index.
func (df *DataFrame) SortByIndex(ascending bool) error {
	if len(df.series) > 0 {
		for i := range df.series {
			df.series[i].SortByIndex(ascending)
		}
	}
	df.index = df.series[0].index
	return nil
}

// SortByValues sorts the items by values in a selected Series.
func (df *DataFrame) SortByValues(by string, ascending bool) error {
	var index IndexData
	for i := range df.series {
		if df.series[i].name == by {
			df.series[i].SortByValues(ascending)
			index = df.series[i].index
			break
		}
	}

	for i := range df.series {
		df.series[i].SortByGivenIndex(index, true)
	}

	df.index = index
	return nil
}

// SortByColumns sorts the columns of the DataFrame object.
func (df *DataFrame) SortByColumns() {
	sort.Slice(df.series, func(i, j int) bool {
		return df.series[i].name < df.series[j].name
	})
	sort.Strings(df.columns)
}

// SortIndexColFirst puts the index column at the front.
func (df *DataFrame) SortIndexColFirst() {
	counter := 0
	for _, indexName := range df.index.names {
		for j, ser := range df.series {
			if ser.name == indexName {
				df.series[counter], df.series[j] = df.series[j], df.series[counter]
				df.columns[counter], df.columns[j] = df.columns[j], df.columns[counter]
				counter++
			}
		}
	}
}

/* Reshaping Fuctions */

// Pivot returns an organized Dataframe that has values corresponding to the index and the given column.
func (df *DataFrame) Pivot(column, value string) (DataFrame, error) {
	// check if index contains duplicate entires.
	// for the same index, if column has a value that is repeated, then raise an error.

	// loc each individual values, then concat them.
	filteredDf, err := df.LocCols(column, value)
	if err != nil {
		return DataFrame{}, err
	}

	type dataMap struct {
		column        string
		indexValueMap map[string]interface{}
	}
	dataMaps := make([]dataMap, 0)

	newDfData := make([][]interface{}, 0)
	newDfColumns := make([]string, 0)
	newDfIndexIndex := make([]Index, 0)
	newDfIndexNames := filteredDf.index.names

	for i, data := range filteredDf.series[0].data {
		if !containsString(newDfColumns, fmt.Sprint(data)) {
			newDfColumns = append(newDfColumns, fmt.Sprint(data))
			dm := dataMap{fmt.Sprint(data), map[string]interface{}{}}
			dataMaps = append(dataMaps, dm)
		}
		if !containsIndexWithoutId(newDfIndexIndex, filteredDf.index.index[i]) {
			newIndex := Index{len(newDfIndexIndex), filteredDf.index.index[i].value}
			newDfIndexIndex = append(newDfIndexIndex, newIndex)
		}
	}

	for i, index := range filteredDf.index.index {
		colname := fmt.Sprint(filteredDf.series[0].data[i])
		for _, dm := range dataMaps {
			if dm.column == colname {
				innerKey, err := index.hashKeyValueOnly()
				if err != nil {
					return DataFrame{}, err
				}

				dm.indexValueMap[*innerKey] = filteredDf.series[1].data[i]
			}
		}
	}

	for _, col := range newDfColumns {
		eachColData := make([]interface{}, 0)
		for _, dm := range dataMaps {
			if dm.column == col {
				for _, index := range newDfIndexIndex {

					innerKey, err := index.hashKeyValueOnly()
					if err != nil {
						return DataFrame{}, err
					}

					val, exists := dm.indexValueMap[*innerKey]
					if !exists {
						switch filteredDf.series[1].data[0].(type) {
						case string:
							eachColData = append(eachColData, "")
						case float64:
							eachColData = append(eachColData, math.NaN())
						case int:
							// should make a null for integer later
							eachColData = append(eachColData, math.NaN())
						}
					} else {
						eachColData = append(eachColData, val)
					}
				}
			}
		}

		newDfData = append(newDfData, eachColData)
	}

	newDf, err := NewDataFrame(newDfData, newDfColumns, nil)
	if err != nil {
		return DataFrame{}, err
	}

	newDfIndex := IndexData{newDfIndexIndex, newDfIndexNames}
	newDf.index = newDfIndex
	for i := range newDf.series {
		newDf.series[i].index = newDf.index
	}

	return newDf, nil
}

// PivotTable rearranges the data by a given index and column.
// Each value will be aggregated via an aggregation function.
// Pick three columns from the DataFrame, each to serve as the index, column, and value.
// PivotTable ignores NaN values.
func (df *DataFrame) PivotTable(index, column, value string, aggFunc StatsFunc) (DataFrame, error) {
	filteredData, err := df.LocColsItems(index, column, value)
	if err != nil {
		return DataFrame{}, err
	}

	// iterate through filteredData
	// check for unique combinations of index and column
	// for each unique combination, store the value in a valueSlice

	// how to check for unique combination
	// create and store a hash for index+column
	// iterate through filteredData[0]
	// if index+column hash doesnt exist, create and store it
	// if exists, skip
	// either way, store the vale in a valueSlice

	dataMap := make(map[string][]interface{}, 0)
	uniqueHashSlice := make([]string, 0)
	uniqueIndexSlice := make([]string, 0)
	uniqueColSlice := make([]string, 0)
	for i, col := range filteredData[1] {
		idx := filteredData[0][i]
		val := filteredData[2][i]

		if !containsString(uniqueIndexSlice, fmt.Sprint(idx)) {
			uniqueIndexSlice = append(uniqueIndexSlice, fmt.Sprint(idx))
		}
		if !containsString(uniqueColSlice, fmt.Sprint(col)) {
			uniqueColSlice = append(uniqueColSlice, fmt.Sprint(col))
		}

		index := Index{i, []interface{}{idx, col}}
		key, err := index.hashKeyValueOnly()
		if err != nil {
			return DataFrame{}, err
		}

		if !containsString(uniqueHashSlice, *key) {
			uniqueHashSlice = append(uniqueHashSlice, *key)
		}
		dataMap[*key] = append(dataMap[*key], val)
	}

	sort.Strings(uniqueColSlice)
	sort.Strings(uniqueIndexSlice)

	valSlice := make([][]interface{}, 0)
	for i, col := range uniqueColSlice {
		val := make([]interface{}, 0)
		for _, idx := range uniqueIndexSlice {
			index := Index{i, []interface{}{idx, col}}
			key, err := index.hashKeyValueOnly()
			if err != nil {
				return DataFrame{}, err
			}

			result := aggFunc(dataMap[*key])
			if result.Err != nil {
				if math.IsNaN(result.Result) {

				} else {
					return DataFrame{}, result.Err
				}
			}

			val = append(val, result.Result)
		}
		valSlice = append(valSlice, val)
	}

	newDf, err := NewDataFrame(valSlice, uniqueColSlice, nil)
	if err != nil {
		return DataFrame{}, err
	}

	newDfIndex := IndexData{[]Index{}, []string{index}}
	for i, uniqueIndex := range uniqueIndexSlice {
		idx := Index{i, []interface{}{uniqueIndex}}
		newDfIndex.index = append(newDfIndex.index, idx)
	}
	newDf.index = newDfIndex
	for i := range newDf.series {
		newDf.series[i].index = newDf.index
	}

	return newDf, nil
}

// Melt returns the table from wide to long format.
// Use Melt to revert to pre-Pivot format.
func (df *DataFrame) Melt(colName, valueName string) (DataFrame, error) {
	newDfIndexSlice := make([]interface{}, 0)
	newDfColumnSlice := make([]interface{}, 0)
	newDfValueSlice := make([]interface{}, 0)

	for i, idx := range df.index.index {
		for _, col := range df.series {
			newDfIndexSlice = append(newDfIndexSlice, idx.value...)
			newDfColumnSlice = append(newDfColumnSlice, col.name)
			newDfValueSlice = append(newDfValueSlice, col.data[i])
		}
	}

	newDfSlice := make([][]interface{}, 0)
	newDfSlice = append(newDfSlice, newDfIndexSlice, newDfColumnSlice, newDfValueSlice)

	colNameSlice := make([]string, 0)
	colNameSlice = append(colNameSlice, df.index.names...)
	colNameSlice = append(colNameSlice, colName, valueName)
	newDf, err := NewDataFrame(newDfSlice, colNameSlice, df.index.names)
	if err != nil {
		return DataFrame{}, err
	}

	return newDf, nil
}

// GroupBy groups selected columns in a DataFrame object and returns a GroupBy object.
func (df *DataFrame) GroupBy(by ...string) (GroupBy, error) {
	filtered, err := df.LocCols(by...)
	if err != nil {
		return GroupBy{}, err
	}

	colIndMap := make(map[string][]interface{})
	colTuples := make([][]interface{}, 0)

	for i, row := range filtered.index.index {
		colTuple := make([]interface{}, 0)
		for _, ser := range filtered.series {
			colTuple = append(colTuple, ser.data[i])
		}

		index := Index{i, colTuple}
		key, err := index.hashKeyValueOnly()
		if err != nil {
			return GroupBy{}, err
		}

		colIndMap[*key] = append(colIndMap[*key], row.id)
		if !containsSlice(colTuples, colTuple) {
			colTuples = append(colTuples, colTuple)
		}
	}

	gb := new(GroupBy)
	gb.dataFrame = df
	gb.colIndMap = colIndMap
	gb.colTuples = colTuples
	gb.colTuplesLabels = filtered.columns

	return *gb, nil
}

func (df DataFrame) Shape() (shape [2]int) {
	shape[0] = df.index.Len()
	shape[1] = len(df.columns)
	return shape
}

func (df DataFrame) GetRecords() (resMapList []map[string]interface{}) {
	df.Print()
	fmt.Println(df.Shape())
	resMapList = make([]map[string]interface{}, df.Shape()[0])
	for cindex, col := range df.columns {
		for sindex := range resMapList {
			resMap := resMapList[sindex]
			if resMap == nil {
				resMap = make(map[string]interface{})
			}
			s_data := df.series[cindex].Data()
			resMap[col] = s_data[sindex]
			resMapList[sindex] = resMap
		}
	}
	return
}
