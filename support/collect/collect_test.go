package collect

import (
	"sort"
	"strconv"
	"testing"
)

// TestCount tests the Count function.
func TestCount(t *testing.T) {
	count := Count([]int{1, 5, 1})
	if count != 3 {
		t.Errorf("expected 3, got %d", count)
	}
}

// TestCountBy tests the CountBy function.
func TestCountBy(t *testing.T) {
	count := CountBy([]int{1, 5, 1}, func(i int) bool {
		return i < 4
	})
	if count != 2 {
		t.Errorf("expected 2, got %d", count)
	}
}

// TestEach tests the Each function.
func TestEach(t *testing.T) {
	Each([]string{"hello", "world"}, func(x string, i int) {
		if i == 0 {
			if x != "hello" {
				t.Errorf("expected hello, got %s", x)
			}
		} else {
			if x != "world" {
				t.Errorf("expected world, got %s", x)
			}
		}
	})
	Each([]int{0, 1, 2, 3}, func(x int, i int) {
		if x != i {
			t.Errorf("expected %d, got %d", i, x)
		}
	})
}

// TestFilter tests the Filter function.
func TestFilter(t *testing.T) {
	even := Filter([]int{1, 2, 3, 4}, func(x int, index int) bool {
		return x%2 == 0
	})
	if len(even) != 2 || even[0] != 2 || even[1] != 4 {
		t.Errorf("expected [2, 4], got %v", even)
	}
}

// TestGroupBy tests the GroupBy function.
func TestGroupBy(t *testing.T) {
	groups := GroupBy([]int{0, 1, 2, 3, 4, 5}, func(i int) int {
		return i % 3
	})
	expected := map[int][]int{0: {0, 3}, 1: {1, 4}, 2: {2, 5}}
	if len(groups) != len(expected) {
		t.Errorf("expected %v, got %v", expected, groups)
	}
}

// TestKeys tests the Keys function.
func TestKeys(t *testing.T) {
	keys1 := Keys[int, string](map[int]string{1: "foo", 2: "bar"})
	keys2 := Keys[string, int](map[string]int{"foo": 1, "bar": 2})
	sort.Ints(keys1)
	sort.Strings(keys2)
	if len(keys1) != 2 || keys1[0] != 1 || keys1[1] != 2 {
		t.Errorf("expected [1, 2], got %v", keys1)
	}
	if len(keys2) != 2 || keys2[0] != "bar" || keys2[1] != "foo" {
		t.Errorf("expected [bar, foo], got %v", keys2)
	}
}

// TestMap tests the Map function.
func TestMap(t *testing.T) {
	results1 := Map([]int64{1, 2, 3, 4}, func(x int64, _ int) string {
		return strconv.FormatInt(x, 10)
	})
	results2 := Map([]int64{1, 2, 3, 4}, func(x int64, _ int) int64 {
		return x + 1
	})
	if len(results1) != 4 || results1[0] != "1" {
		t.Errorf("expected [1, 2, 3, 4], got %v", results1)
	}
	if len(results2) != 4 || results2[0] != 2 {
		t.Errorf("expected [2, 3, 4, 5], got %v", results2)
	}
}

// TestMax tests the Max function.
func TestMax(t *testing.T) {
	max1 := Max([]int{1, 2, 3})
	max2 := Max([]int{})
	if max1 != 3 {
		t.Errorf("expected 3, got %d", max1)
	}
	if max2 != 0 {
		t.Errorf("expected 0, got %d", max2)
	}
}

// TestMerge tests the Merge function.
func TestMerge(t *testing.T) {
	mergedMaps1 := Merge[string, int](
		map[string]int{"a": 1, "b": 2},
		map[string]int{"b": 3, "c": 4},
	)
	mergedMaps2 := Merge[int, string](
		map[int]string{1: "a", 2: "b"},
		map[int]string{2: "b", 4: "c"},
	)
	if len(mergedMaps1) != 3 || mergedMaps1["a"] != 1 || mergedMaps1["b"] != 3 || mergedMaps1["c"] != 4 {
		t.Errorf("expected map with a:1, b:3, c:4, got %v", mergedMaps1)
	}
	if len(mergedMaps2) != 3 || mergedMaps2[1] != "a" || mergedMaps2[2] != "b" || mergedMaps2[4] != "c" {
		t.Errorf("expected map with 1:a, 2:b, 4:c, got %v", mergedMaps2)
	}
}

// TestMin tests the Min function.
func TestMin(t *testing.T) {
	min1 := Min([]int{1, 2, 3})
	min2 := Min([]int{})
	if min1 != 1 {
		t.Errorf("expected 1, got %d", min1)
	}
	if min2 != 0 {
		t.Errorf("expected 0, got %d", min2)
	}
}

// TestReverse tests the Reverse function.
func TestReverse(t *testing.T) {
	reverseOrder1 := Reverse([]int{0, 1, 2, 3, 4, 5})
	reverseOrder2 := Reverse([]string{"a", "b", "c", "d"})
	if len(reverseOrder1) != 6 || reverseOrder1[0] != 5 || reverseOrder1[5] != 0 {
		t.Errorf("expected [5, 4, 3, 2, 1, 0], got %v", reverseOrder1)
	}
	if len(reverseOrder2) != 4 || reverseOrder2[0] != "d" || reverseOrder2[3] != "a" {
		t.Errorf("expected [d, c, b, a], got %v", reverseOrder2)
	}
}

// TestSplit tests the Split function.
func TestSplit(t *testing.T) {
	result := Split([]int{0, 1, 2, 3, 4, 5}, 2)
	result1 := Split([]int{0, 1, 2, 3, 4, 5, 6}, 2)
	result2 := Split([]int{}, 2)
	result3 := Split([]int{0}, 2)
	result4 := Split([]string{"a", "b", "c", "d"}, 2)

	if len(result) != 3 || len(result[0]) != 2 || result[0][0] != 0 {
		t.Errorf("expected [[0,1],[2,3],[4,5]], got %v", result)
	}
	if len(result1) != 4 || len(result1[3]) != 1 {
		t.Errorf("expected [[0,1],[2,3],[4,5],[6]], got %v", result1)
	}
	if len(result2) != 0 {
		t.Errorf("expected [], got %v", result2)
	}
	if len(result3) != 1 || len(result3[0]) != 1 {
		t.Errorf("expected [[0]], got %v", result3)
	}
	if len(result4) != 2 || len(result4[0]) != 2 {
		t.Errorf("expected [[a,b],[c,d]], got %v", result4)
	}
}

// TestSum tests the Sum function.
func TestSum(t *testing.T) {
	list := []int{1, 2, 3, 4, 5}
	sum := Sum(list)
	if sum != 15 {
		t.Errorf("expected 15, got %d", sum)
	}
}

// TestUnique tests the Unique function.
func TestUnique(t *testing.T) {
	uniqValues1 := Unique([]int{1, 2, 2, 1})
	uniqValues2 := Unique([]string{"a", "b", "b", "a"})
	if len(uniqValues1) != 2 || uniqValues1[0] != 1 || uniqValues1[1] != 2 {
		t.Errorf("expected [1, 2], got %v", uniqValues1)
	}
	if len(uniqValues2) != 2 || uniqValues2[0] != "a" || uniqValues2[1] != "b" {
		t.Errorf("expected [a, b], got %v", uniqValues2)
	}
}

// TestValues tests the Values function.
func TestValues(t *testing.T) {
	values1 := Values[string, int](map[string]int{"foo": 1, "bar": 2})
	values2 := Values[int, string](map[int]string{1: "foo", 2: "bar"})
	sort.Ints(values1)
	sort.Strings(values2)
	if len(values1) != 2 || values1[0] != 1 || values1[1] != 2 {
		t.Errorf("expected [1, 2], got %v", values1)
	}
	if len(values2) != 2 || values2[0] != "bar" || values2[1] != "foo" {
		t.Errorf("expected [bar, foo], got %v", values2)
	}
}
