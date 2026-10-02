package concise

import (
	"errors"
	"maps"
	"slices"
	"strconv"
	"testing"
)

func TestClosest(t *testing.T) {
	for _, tc := range []struct {
		arr  []int
		want int
	}{
		{[]int{-1, 1}, 0},
		{[]int{-1, -7, -5}, -2},
		{[]int{1, 2, 1, 6}, 3},
		{nil, 0},
		{[]int{0, 0}, 0},
		{[]int{0, 1, 2}, 3},
		{[]int{0, -1, -2}, -3},
	} {
		if got := Closest(tc.arr); got != tc.want {
			t.Errorf("Closest(%v) = %d, want %d", tc.arr, got, tc.want)
		}
	}
}

func TestJoinExcluding(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	for _, tc := range []struct {
		exclude []int
		nums    []int
		want    int
		err     error
	}{
		{[]int{0, 1, 0}, []int{1000, 1010}, 0, ErrNotFound},
		{[]int{0, 1, 0}, []int{1259, 2601}, 62952, nil},
		{nil, nil, 0, ErrNotFound},
		{nil, []int{0, 12}, 21, nil},
		{nil, []int{0}, 0, nil},
		{[]int{1, 2}, []int{120}, 0, nil},
		{[]int{-1, 10}, []int{12}, 21, nil},
		{nil, []int{-1}, 0, ErrNegativeNumber},
		{nil, []int{maxInt, 0, -1}, 0, ErrNegativeNumber},
		{nil, []int{maxInt, 0}, 0, ErrOverflow},
	} {
		got, err := JoinExcluding(tc.exclude, tc.nums...)
		if got != tc.want || !errors.Is(err, tc.err) {
			t.Errorf("JoinExcluding(%v, %v) = (%d, %v), want (%d, %v)", tc.exclude, tc.nums, got, err, tc.want, tc.err)
		}
		if errors.Is(err, ErrNotFound) && err.Error() != "not found" {
			t.Errorf("error message = %q, want %q", err.Error(), "not found")
		}
	}

	text := "1563847412"
	if strconv.IntSize == 64 {
		text = "8085774586302733229"
	}
	num, err := strconv.Atoi(text)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := JoinExcluding(nil, num); got != 0 || !errors.Is(err, ErrOverflow) {
		t.Errorf("reverse overflow: got (%d, %v), want (0, ErrOverflow)", got, err)
	}
}

func TestUniqueDescending(t *testing.T) {
	for _, tc := range []struct {
		a, b []int
		want []int
	}{
		{[]int{1, 2}, []int{1, 3}, []int{1}},
		{[]int{1, 2, 2}, []int{1, 2, 4}, []int{2, 1}},
		{nil, []int{1}, nil},
		{[]int{1}, []int{2}, nil},
		{[]int{-1, 0, 2, 2}, []int{2, 0, -1, 2}, []int{2, 0, -1}},
	} {
		if got := UniqueDescending(tc.a, tc.b); !slices.Equal(got, tc.want) || (tc.want == nil && got != nil) {
			t.Errorf("UniqueDescending(%v, %v) = %v, want %v", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestCount(t *testing.T) {
	for _, tc := range []struct {
		num  int
		want map[int]int
	}{
		{1223334, map[int]int{1: 1, 2: 2, 3: 3, 4: 1}},
		{-1223334, map[int]int{1: 1, 2: 2, 3: 3, 4: 1}},
		{0, map[int]int{0: 1}},
		{1002, map[int]int{0: 2, 1: 1, 2: 1}},
	} {
		if got := Count(tc.num); !maps.Equal(got, tc.want) {
			t.Errorf("Count(%d) = %v, want %v", tc.num, got, tc.want)
		}
	}
}

func TestSumProducts(t *testing.T) {
	for _, tc := range []struct {
		nums []int
		want int
	}{
		{[]int{1, 2, 3, 4}, 14},
		{[]int{1, 2, 3, 4, 5}, 39},
		{nil, 0},
		{[]int{5}, 25},
		{[]int{-1, 2, 3, -4}, -14},
	} {
		if got := SumProducts(tc.nums...); got != tc.want {
			t.Errorf("SumProducts(%v) = %d, want %d", tc.nums, got, tc.want)
		}
	}
}

func TestConcurrentCalls(t *testing.T) {
	for caller := range 8 {
		t.Run(strconv.Itoa(caller), func(t *testing.T) {
			t.Parallel()
			for range 16 {
				if got := SumProducts(1, 2, 3, 4, 5); got != 39 {
					t.Fatalf("got %d, want 39", got)
				}
			}
		})
	}
}
