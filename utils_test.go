package gambas

import (
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestCheckTypeIntegrity(t *testing.T) {
	type checkTypeIntegrityTest struct {
		arg1     []interface{}
		expected bool
	}
	checkTypeIntegrityTests := []checkTypeIntegrityTest{
		{
			[]interface{}{0, 1, 2, 3, 4, 5, 6},
			true,
		},
		{
			[]interface{}{3.4, 5.6, 2.4, 6.5, 7},
			true,
		},
		{
			[]interface{}{1, 2, "3", "4", 5},
			false,
		},
		{
			[]interface{}{"a", "b", "c"},
			true,
		},
		{
			[]interface{}{true, false, true},
			false,
		},
	}

	for _, test := range checkTypeIntegrityTests {
		output, err := checkTypeIntegrity(test.arg1)
		if !cmp.Equal(output, test.expected) || (output != test.expected && err == nil) {
			t.Fatalf("expected %v, got %v, error %v", test.expected, output, err)
		}
	}
}

func TestI2f(t *testing.T) {
	type i2fTest struct {
		arg1          interface{}
		expected      float64
		expectedError error
	}
	i2fTests := []i2fTest{
		{1, 1.0, nil},
		{-5, -5.0, nil},
		{"45", 0.0, fmt.Errorf("%v is not a number", "45")},
		{2.0, 2.0, nil},
		{-5.6, -5.6, nil},
		{0.0, 0.0, nil},
		{0.000, 0.0, nil},
	}

	for _, test := range i2fTests {
		output, err := i2f(test.arg1)
		if !cmp.Equal(output, test.expected) || !cmp.Equal(fmt.Sprint(err), fmt.Sprint(test.expectedError)) {
			t.Fatalf("expected %v, got %v, error %v", test.expected, output, err)
		}
	}
}

// func TestCheckType(t *testing.T) {

// }

func TestInterface2F64Data(t *testing.T) {
	type interface2F64DataTest struct {
		arg1     []interface{}
		expected F64Data
	}
	interface2F64DataTests := []interface2F64DataTest{
		{
			[]interface{}{0.0, 1.0, 2.0, 3.0, 4.0},
			F64Data{0.0, 1.0, 2.0, 3.0, 4.0},
		},
		{
			[]interface{}{0, 1, 2, 3, 4},
			nil,
		},
		{
			[]interface{}{"0", "1", "2", "3", "4"},
			nil,
		},
		{
			[]interface{}{"a", "b", "c", "d", "e"},
			nil,
		},
		{
			[]interface{}{0.5, 1, 1.5, 2, 2.5},
			nil,
		},
		{
			[]interface{}{"a", 1, "b", 2, "c", 3},
			nil,
		},
	}
	for _, test := range interface2F64DataTests {
		output, err := interface2F64Data(test.arg1)
		if !cmp.Equal(output, test.expected) || (output != nil && err != nil) {
			t.Fatalf("expected %v, got %v, error %v", test.expected, output, err)
		}
	}
}

func TestInterface2StringData(t *testing.T) {
	type interface2StringDataTest struct {
		arg1     []interface{}
		expected StringData
	}
	interface2StringDataTests := []interface2StringDataTest{
		{
			[]interface{}{0.0, 1.0, 2.0, 3.0, 4.0},
			nil,
		},
		{
			[]interface{}{0, 1, 2, 3, 4},
			nil,
		},
		{
			[]interface{}{"0", "1", "2", "3", "4"},
			StringData{"0", "1", "2", "3", "4"},
		},
		{
			[]interface{}{"a", "b", "c", "d", "e"},
			StringData{"a", "b", "c", "d", "e"},
		},
		{
			[]interface{}{0.5, 1, 1.5, 2, 2.5},
			nil,
		},
		{
			[]interface{}{"a", 1, "b", 2, "c", 3},
			nil,
		},
	}
	for _, test := range interface2StringDataTests {
		output, err := interface2StringData(test.arg1)
		if !cmp.Equal(output, test.expected) || (output != nil && err != nil) {
			t.Fatalf("expected %v, got %v, error %v", test.expected, output, err)
		}
	}
}
