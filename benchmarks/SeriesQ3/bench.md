`go test -run=^$ -bench=SeriesQ3 -benchmem -benchtime=20s -timeout=1h`

# Original (old median)
goos: linux
goarch: amd64
pkg: github.com/jpoly1219/gambas
cpu: Intel(R) Core(TM) i7-7700K CPU @ 4.20GHz
BenchmarkSeriesQ3/1_Points-8            125634000              189.2 ns/op           104 B/op          5 allocs/op
BenchmarkSeriesQ3/10_Points-8           63209424               371.7 ns/op           296 B/op          7 allocs/op
BenchmarkSeriesQ3/100_Points-8           6174436              3819 ns/op            2088 B/op         10 allocs/op
BenchmarkSeriesQ3/1000_Points-8           270987             89003 ns/op           25256 B/op         14 allocs/op
BenchmarkSeriesQ3/10000_Points-8           20316           1177924 ns/op          357672 B/op         21 allocs/op
BenchmarkSeriesQ3/100000_Points-8           1648          14476772 ns/op         4101416 B/op         30 allocs/op
BenchmarkSeriesQ3/1000000_Points-8           138         173771437 ns/op        41678120 B/op         40 allocs/op
BenchmarkSeriesQ3/10000000_Points-8           10        2042518852 ns/op        492000553 B/op        51 allocs/op
PASS
ok      github.com/jpoly1219/gambas     255.070s

# New (new median)
goos: linux
goarch: amd64
pkg: github.com/jpoly1219/gambas
cpu: Intel(R) Core(TM) i7-7700K CPU @ 4.20GHz
BenchmarkSeriesQ3/1_Points-8            127729239              185.9 ns/op           104 B/op          5 allocs/op
BenchmarkSeriesQ3/10_Points-8           60920182               369.0 ns/op           296 B/op          7 allocs/op
BenchmarkSeriesQ3/100_Points-8           6181755              3830 ns/op            2088 B/op         10 allocs/op
BenchmarkSeriesQ3/1000_Points-8           261720             88848 ns/op           25256 B/op         14 allocs/op
BenchmarkSeriesQ3/10000_Points-8           20217           1177276 ns/op          357672 B/op         21 allocs/op
BenchmarkSeriesQ3/100000_Points-8           1609          14558430 ns/op         4101416 B/op         30 allocs/op
BenchmarkSeriesQ3/1000000_Points-8           138         172481374 ns/op        41678120 B/op         40 allocs/op
BenchmarkSeriesQ3/10000000_Points-8           10        2046607879 ns/op        492000552 B/op        51 allocs/op
PASS
ok      github.com/jpoly1219/gambas     252.342s

# New (new median + quickselect)


# Fit
