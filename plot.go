package gambas

import (
	"bytes"
	"fmt"
	"math/rand"
	"os/exec"
	"path/filepath"
	"time"
)

// Plotting functionality uses gnuplot as its backend.

// PlotData holds the data required for plotting a dataset.
// Df is the DataFrame object you would like to plot.
// Columns are the columns in Df that you want to plot. Usually, it's a pair of columns [xcol, ycol].
// If you want to create a bar graph or a histogram, you can add more columns to Columns.
// Function is an arbitrary function such as sin(x) or an equation of the line of best fit.
// If you want to graph an arbitrary function, leave Df and Columns as nil.
// Otherwise, leave Function as "".
// Opts is whatever gnuplot option you would like to set.
type PlotData struct {
	Df       *DataFrame
	Columns  []string
	Function string
	Opts     []GnuplotOpt
}

// Plot plots the DataFrame object.
// Choose two columns to use for the x axis and y axis.
// Then pass in any options you need. Refer to the gnuplot documentation for options.
// For example, `set xrange [-10:10]; set xlabel "myX"; set ylabel "myY"; plot "myDf.dat" using 0:1 lc 0 w lines`
// Plot(<xcol>, <ycol>, SetXrange("[-10:10]"), SetXlabel("myX"), SetYlabel("myY"), Using("0:1 lc 0 w lines"))
func Plot(pd PlotData, setOpts ...GnuplotOpt) error {
	path := ""

	if pd.Function != "" && pd.Df == nil && pd.Columns == nil {
		path = pd.Function
	} else {
		rand.Seed(time.Now().UnixNano())
		newDf, err := pd.Df.LocCols(pd.Columns...)
		if err != nil {
			return err
		}

		path = filepath.Join("/", "tmp", fmt.Sprintf("%x.csv", rand.Intn(100000000)))
		_, err = WriteCsv(newDf, path, true)
		if err != nil {
			return err
		}
	}

	var setBuf bytes.Buffer
	for _, setOpt := range setOpts {
		str := setOpt.createCmdString()
		setBuf.WriteString(str)
		setBuf.WriteString("; ")
	}

	var usingBuf bytes.Buffer
	var withBuf bytes.Buffer
	for _, opt := range pd.Opts {
		str := opt.createCmdString()
		switch opt.getOption() {
		case "using":
			usingBuf.WriteString(str)
		case "with":
			withBuf.WriteString(str)
		default:
		}
	}

	cmdString := fmt.Sprintf(`%s %s "%s" %s %s`, setBuf.String(), "plot", path, usingBuf.String(), withBuf.String())
	fmt.Println(cmdString)
	cmd := exec.Command("gnuplot", "-persist", "-e", cmdString)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf(fmt.Sprint(err, cmd.Stderr))
	}

	return nil
}

// PlotN plots n number of datasets side by side.
// This is useful when you want to compare two different datasets,
// or a dataset with a line of best fit.
// Set options should be passed in separately as a parameter, not inside PlotData.
func PlotN(plotdata []PlotData, setOpts ...GnuplotOpt) error {
	rand.Seed(time.Now().UnixNano())

	var setBuf bytes.Buffer
	for _, setOpt := range setOpts {
		str := setOpt.createCmdString()
		setBuf.WriteString(str)
		setBuf.WriteString("; ")
	}

	cmdString := fmt.Sprintf(`%s %s `, setBuf.String(), "plot")

	for _, pd := range plotdata {
		path := ``

		if pd.Function != "" && pd.Df == nil && pd.Columns == nil {
			path = pd.Function
		} else {
			newDf, err := pd.Df.LocCols(pd.Columns...)
			if err != nil {
				return err
			}

			path = filepath.Join("/", "tmp", fmt.Sprintf("%x.csv", rand.Intn(100000000)))
			_, err = WriteCsv(newDf, path, true)
			if err != nil {
				return err
			}
			path = fmt.Sprintf(`"%s"`, path)
		}

		var usingBuf bytes.Buffer
		var withBuf bytes.Buffer
		for _, opt := range pd.Opts {
			str := opt.createCmdString()
			switch opt.getOption() {
			case "using":
				usingBuf.WriteString(str)
			case "with":
				withBuf.WriteString(str)
			default:
			}
		}

		cmdStringPiece := fmt.Sprintf(`%s %s %s,`, path, usingBuf.String(), withBuf.String())
		cmdString += cmdStringPiece
	}
	fmt.Println(cmdString)
	cmd := exec.Command("gnuplot", "-persist", "-e", cmdString)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf(fmt.Sprint(err, cmd.Stderr))
	}

	return nil
}

// Fit calculates the line of best fit.
// ff is the fitting function.
// pd is the PlotData you would like to fit. Only the data pd.Df and pd.Columns will be used.
// Pass options such as `using` or `via` in opts.
func Fit(ff string, pd PlotData, opts ...GnuplotOpt) error {
	rand.Seed(time.Now().UnixNano())
	newDf, err := pd.Df.LocCols(pd.Columns...)
	if err != nil {
		return err
	}

	path := filepath.Join("/", "tmp", fmt.Sprintf("%x.csv", rand.Intn(100000000)))
	_, err = WriteCsv(newDf, path, true)
	if err != nil {
		return err
	}

	var usingBuf, viaBuf bytes.Buffer
	for _, opt := range opts {
		str := opt.createCmdString()
		if opt.getOption() == "using" {
			usingBuf.WriteString(str)
		}
		if opt.getOption() == "via" {
			viaBuf.WriteString(str)
		}
	}

	cmdString := fmt.Sprintf(`%s %s "%s" %s %s`, "fit", ff, path, usingBuf.String(), viaBuf.String())
	cmd := exec.Command("gnuplot", "-persist", "-e", cmdString)
	combOutput, err := cmd.CombinedOutput()
	if err != nil {
		return err
	}

	fmt.Printf("%s", combOutput)

	return nil
}
