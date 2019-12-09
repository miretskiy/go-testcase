package testcase

import (
	"reflect"
	"testing"
)

func TestGenerateTestCases(t *testing.T) {
	type testCase struct {
		Enabled bool `tc:"[true, false]"`
		Size    int  `tc:"[1,2]"`
	}

	expected := []testCase{{true, 1}, {true, 2}, {false, 1}, {false, 2}}
	var generated []testCase

	var tc testCase
	for gen := GenerateTestCases(t, &tc); gen.Next(); {
		generated = append(generated, tc)
	}

	if !reflect.DeepEqual(expected, generated) {
		t.Errorf("Expected\n%+v\nFound\n%+v\n", expected, generated)
	}
}

func TestGnerateMixedTestCases(t *testing.T) {
	type testCase struct {
		Enabled bool `tc:"[true, false]"`
		Size    int  `tc:"[1,2]"`
		i       int
	}
	expected := []testCase{
		{true, 1, 1}, {true, 2, 1}, {false, 1, 1}, {false, 2, 1},
		{true, 1, 2}, {true, 2, 2}, {false, 1, 2}, {false, 2, 2},
		{true, 1, 3}, {true, 2, 3}, {false, 1, 3}, {false, 2, 3},
	}

	var generated []testCase
	for _, tc := range []testCase{
		{i: 1}, {i: 2}, {i: 3},
	} {
		for gen := GenerateTestCases(t, &tc); gen.Next(); {
			generated = append(generated, tc)
		}
	}

	if !reflect.DeepEqual(expected, generated) {
		t.Errorf("Expected\n%+v\nFound\n%+v\n", expected, generated)
	}
}

func TestGenerateComplexTypes(t *testing.T) {
	type testCase struct {
		Slice []int          `tc:"[[1,2,3],[9,8],[0]]"`
		KV    map[string]int `tc:"[{\"one\":1, \"two\":2}, {\"zero\":0}]"`
	}

	a1 := []int{1, 2, 3}
	a2 := []int{9, 8}
	a3 := []int{0}
	m1 := map[string]int{"one": 1, "two": 2}
	m2 := map[string]int{"zero": 0}

	expected := []testCase{
		{a1, m1}, {a1, m2},
		{a2, m1}, {a2, m2},
		{a3, m1}, {a3, m2},
	}

	var generated []testCase
	var tc testCase
	for gen := GenerateTestCases(t, &tc); gen.Next(); {
		generated = append(generated, tc)
	}

	if !reflect.DeepEqual(expected, generated) {
		t.Errorf("Expected\n%+v\nFound\n%+v\n", expected, generated)
	}
}

func TestGenerateNextedComplexTypes(t *testing.T) {
	type testCase struct {
		Enabled bool             `tc:"[true, false]"`
		Slice   [][]int          `tc:"[[[1],[2,3]],[[9],[8]]]"`
		KV      map[string][]int `tc:"[{\"1-2\":[1,2], \"3-4\":[3, 4]}, {\"5-7\": [5,6,7]}]"`
	}

	a1 := [][]int{{1}, {2, 3}}
	a2 := [][]int{{9}, {8}}
	m1 := map[string][]int{"1-2": {1, 2}, "3-4": {3, 4}}
	m2 := map[string][]int{"5-7": {5, 6, 7}}

	expected := []testCase{
		{true, a1, m1}, {true, a1, m2}, {true, a2, m1}, {true, a2, m2},
		{false, a1, m1}, {false, a1, m2}, {false, a2, m1}, {false, a2, m2},
	}

	var generated []testCase
	var tc testCase
	for gen := GenerateTestCases(t, &tc); gen.Next(); {
		generated = append(generated, tc)
	}

	if !reflect.DeepEqual(expected, generated) {
		t.Errorf("Expected\n%+v\nFound\n%+v\n", expected, generated)
	}
}
