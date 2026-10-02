package simple

import (
	"errors"
	"fmt"
	"maps"
	"math/rand"
	"slices"
	"strconv"
	"sync"
	"testing"

	"treasury.id/concise"
	"treasury.id/optimized"
)

func TestClosest(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	tests := []struct {
		name string
		arr  []int
		want int
	}{
		{"example_mixed", []int{-1, 1}, 0},
		{"example_negative", []int{-1, -7, -5}, -2},
		{"example_positive", []int{1, 2, 1, 6}, 3},
		{"empty", nil, 0},
		{"zeros", []int{0, 0}, 0},
		{"positive_gap", []int{2, 3}, 1},
		{"negative_gap", []int{-2, -3}, -1},
		{"positive_complete", []int{3, 1, 2}, 4},
		{"negative_complete", []int{-3, -1, -2}, -4},
		{"zero_positive", []int{0, 1, 2, 0}, 3},
		{"zero_negative", []int{0, -1, -3, 0}, -2},
		{"max_int", []int{1, maxInt}, 2},
		{"min_int", []int{-maxInt - 1, -1}, -2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := slices.Clone(tt.arr)
			if got := Closest(tt.arr); got != tt.want {
				t.Fatalf("Closest(%v) = %d, want %d", tt.arr, got, tt.want)
			}
			if !slices.Equal(tt.arr, before) {
				t.Fatal("input was modified")
			}
		})
	}
}

func TestJoinExcluding(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	allDigits := []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}
	tests := []struct {
		name    string
		exclude []int
		nums    []int
		want    int
		err     error
	}{
		{"example_not_found", []int{0, 1, 0}, []int{1000, 1010}, 0, ErrNotFound},
		{"example_result", []int{0, 1, 0}, []int{1259, 2601}, 62952, nil},
		{"no_exclusions", nil, []int{1259, 2601}, 10629521, nil},
		{"empty", nil, nil, 0, ErrNotFound},
		{"zero", nil, []int{0}, 0, nil},
		{"all_zeros", nil, []int{0, 0, 0}, 0, nil},
		{"zero_excluded", []int{0}, []int{0}, 0, ErrNotFound},
		{"leading_zeros", nil, []int{0, 0, 12}, 21, nil},
		{"trailing_zeros", nil, []int{120, 0}, 21, nil},
		{"zero_survives", []int{1, 2}, []int{120}, 0, nil},
		{"outside_digit_range", []int{-1, 10, maxInt}, []int{12}, 21, nil},
		{"duplicate_exclusion", []int{0, 1, 1}, []int{1012}, 2, nil},
		{"negative", nil, []int{-12}, 0, ErrNegativeNumber},
		{"negative_before_overflow", nil, []int{maxInt, 0, -1}, 0, ErrNegativeNumber},
		{"min_int", nil, []int{-maxInt - 1}, 0, ErrNegativeNumber},
		{"join_at_limit", allDigits, []int{maxInt / 10, maxInt % 10}, 0, ErrNotFound},
		{"join_past_limit", nil, []int{maxInt / 10, maxInt%10 + 1}, 0, ErrOverflow},
		{"join_overflow_even_filtered", allDigits, []int{maxInt, 0}, 0, ErrOverflow},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exclude, nums := slices.Clone(tt.exclude), slices.Clone(tt.nums)
			got, err := JoinExcluding(exclude, nums...)
			if got != tt.want || !errors.Is(err, tt.err) {
				t.Fatalf("got (%d, %v), want (%d, %v)", got, err, tt.want, tt.err)
			}
			if errors.Is(err, ErrNotFound) && err.Error() != "not found" {
				t.Fatalf("error message = %q, want %q", err.Error(), "not found")
			}
			if !slices.Equal(exclude, tt.exclude) || !slices.Equal(nums, tt.nums) {
				t.Fatal("inputs were modified")
			}
		})
	}
}

func TestJoinExcludingReverseOverflow(t *testing.T) {
	text := "1563847412" // Its reverse overflows a 32-bit int.
	if strconv.IntSize == 64 {
		text = "8085774586302733229"
	}
	num, err := strconv.Atoi(text)
	if err != nil {
		t.Fatal(err)
	}
	if got, err := JoinExcluding(nil, num); got != 0 || !errors.Is(err, ErrOverflow) {
		t.Fatalf("got (%d, %v), want (0, ErrOverflow)", got, err)
	}
}

func TestUniqueDescending(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	minInt := -maxInt - 1
	tests := []struct {
		name string
		a, b []int
		want []int
	}{
		{"example_single", []int{1, 2}, []int{1, 3}, []int{1}},
		{"example_unique", []int{1, 2, 2}, []int{1, 2, 4}, []int{2, 1}},
		{"empty", nil, []int{1}, nil},
		{"no_overlap", []int{1, 2}, []int{3, 4}, nil},
		{"duplicates_both", []int{2, 2, 1, 1}, []int{1, 2, 2, 1}, []int{2, 1}},
		{"mixed_signs", []int{-2, 0, 3, -1}, []int{-1, 3, 0, -2}, []int{3, 0, -1, -2}},
		{"shorter_second", []int{1, 2, 3, 4}, []int{4, 2}, []int{4, 2}},
		{"extremes", []int{minInt, maxInt, 0}, []int{0, minInt, maxInt}, []int{maxInt, 0, minInt}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, b := slices.Clone(tt.a), slices.Clone(tt.b)
			got := UniqueDescending(a, b)
			if !slices.Equal(got, tt.want) || (tt.want == nil && got != nil) {
				t.Fatalf("got %v, want %v", got, tt.want)
			}
			if !slices.Equal(a, tt.a) || !slices.Equal(b, tt.b) {
				t.Fatal("inputs were modified")
			}
		})
	}
}

func TestUniqueDescendingAliasedInputs(t *testing.T) {
	shared := []int{4, 1, 2, 4}
	if got := UniqueDescending(shared[:3], shared[1:]); !slices.Equal(got, []int{4, 2, 1}) {
		t.Fatalf("got %v, want [4 2 1]", got)
	}
	if !slices.Equal(shared, []int{4, 1, 2, 4}) {
		t.Fatal("shared input was modified")
	}
}

func TestCount(t *testing.T) {
	tests := []struct {
		name string
		num  int
		want map[int]int
	}{
		{"example", 1223334, map[int]int{1: 1, 2: 2, 3: 3, 4: 1}},
		{"zero", 0, map[int]int{0: 1}},
		{"single_digit", 7, map[int]int{7: 1}},
		{"negative", -1223334, map[int]int{1: 1, 2: 2, 3: 3, 4: 1}},
		{"internal_zeros", 1002, map[int]int{0: 2, 1: 1, 2: 1}},
		{"repeated", 99999, map[int]int{9: 5}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Count(tt.num); !maps.Equal(got, tt.want) {
				t.Fatalf("Count(%d) = %v, want %v", tt.num, got, tt.want)
			}
		})
	}
}

func TestCountIntegerLimits(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	for _, num := range []int{maxInt, -maxInt - 1} {
		want := make(map[int]int)
		for _, digit := range strconv.Itoa(num) {
			if digit != '-' {
				want[int(digit-'0')]++
			}
		}
		if got := Count(num); !maps.Equal(got, want) {
			t.Fatalf("Count(%d) = %v, want %v", num, got, want)
		}
	}
}

func TestSumProducts(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	tests := []struct {
		name string
		nums []int
		want int
	}{
		{"example_even", []int{1, 2, 3, 4}, 14},
		{"example_odd", []int{1, 2, 3, 4, 5}, 39},
		{"empty", nil, 0},
		{"single", []int{5}, 25},
		{"zero", []int{0}, 0},
		{"negative", []int{-1, 2, 3, -4}, -14},
		{"negative_odd", []int{-1, -2, -3}, 11},
		{"multiplication_overflow", []int{maxInt, 2}, maxInt * 2},
		{"sum_overflow", []int{maxInt, 1, 1, 1}, maxInt + 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			before := slices.Clone(tt.nums)
			if got := SumProducts(tt.nums...); got != tt.want {
				t.Fatalf("SumProducts(%v) = %d, want %d", tt.nums, got, tt.want)
			}
			if !slices.Equal(tt.nums, before) {
				t.Fatal("input was modified")
			}
		})
	}
}

func TestSumProductsConcurrentCalls(t *testing.T) {
	var callers sync.WaitGroup
	for range 8 {
		callers.Go(func() {
			for range 16 {
				if got := SumProducts(1, 2, 3, 4, 5); got != 39 {
					t.Errorf("got %d, want 39", got)
				}
			}
		})
	}
	callers.Wait()
}

func ExampleClosest() {
	fmt.Println(Closest([]int{-1, 1}))
	fmt.Println(Closest([]int{-1, -7, -5}))
	fmt.Println(Closest([]int{1, 2, 1, 6}))
	// Output:
	// 0
	// -2
	// 3
}

func ExampleJoinExcluding() {
	fmt.Println(JoinExcluding([]int{0, 1, 0}, 1000, 1010))
	fmt.Println(JoinExcluding([]int{0, 1, 0}, 1259, 2601))
	// Output:
	// 0 not found
	// 62952 <nil>
}

func ExampleUniqueDescending() {
	fmt.Println(UniqueDescending([]int{1, 2}, []int{1, 3}))
	fmt.Println(UniqueDescending([]int{1, 2, 2}, []int{1, 2, 4}))
	// Output:
	// [1]
	// [2 1]
}

func ExampleCount() {
	counts := Count(1223334)
	for _, digit := range []int{1, 2, 3, 4} {
		fmt.Printf("%d -> %d\n", digit, counts[digit])
	}
	// Output:
	// 1 -> 1
	// 2 -> 2
	// 3 -> 3
	// 4 -> 1
}

func ExampleSumProducts() {
	fmt.Println(SumProducts(1, 2, 3, 4))
	fmt.Println(SumProducts(1, 2, 3, 4, 5))
	// Output:
	// 14
	// 39
}

func TestLongSequences(t *testing.T) {
	random := rand.New(rand.NewSource(7))
	for _, sign := range []int{1, -1} {
		arr := make([]int, 4096)
		for i := range arr {
			arr[i] = sign * (i + 1)
		}
		random.Shuffle(len(arr), func(i, j int) { arr[i], arr[j] = arr[j], arr[i] })
		if got := Closest(arr); got != sign*4097 {
			t.Fatalf("Closest: got %d, want %d", got, sign*4097)
		}
	}

	a := make([]int, 4097)
	for i := range a {
		a[i] = i - 2048
	}
	b := slices.Clone(a)
	random.Shuffle(len(a), func(i, j int) { a[i], a[j] = a[j], a[i] })
	random.Shuffle(len(b), func(i, j int) { b[i], b[j] = b[j], b[i] })
	if got, want := UniqueDescending(a, b), optimized.UniqueDescending(a, b); !slices.Equal(got, want) {
		t.Fatal("large intersection differs from optimized version")
	}
	var want int
	for i := 0; i < len(a); i += 2 {
		partner := i + 1
		if partner == len(a) {
			partner = i
		}
		want += a[i] * a[partner]
	}
	if got := SumProducts(a...); got != want {
		t.Fatalf("SumProducts: got %d, want %d", got, want)
	}
}

// All simpler versions are checked against the independently implemented optimized one.
func FuzzAgreement(f *testing.F) {
	maxInt := int(^uint(0) >> 1)
	f.Add([]byte{1, 2, 1, 6}, uint8(1), 1223334)
	f.Add([]byte{1, 7, 5}, uint8(2), maxInt)
	f.Add([]byte{255, 1, 0}, uint8(0), -maxInt-1)
	f.Add([]byte{}, uint8(0), 0)
	f.Fuzz(func(t *testing.T, data []byte, mode uint8, num int) {
		if len(data) > 128 {
			t.Skip()
		}
		arr, nums := make([]int, len(data)), make([]int, len(data))
		for i, value := range data {
			nums[i] = int(value)
			switch mode % 3 {
			case 0:
				arr[i] = int(int8(value))
			case 1:
				arr[i] = int(value)
			case 2:
				arr[i] = -int(value)
			}
		}
		before := slices.Clone(arr)
		if got, want := Closest(arr), optimized.Closest(slices.Clone(arr)); got != want {
			t.Fatalf("Closest(%v): got %d, want %d", arr, got, want)
		}
		if got, want := concise.Closest(arr), optimized.Closest(slices.Clone(arr)); got != want {
			t.Fatalf("concise.Closest(%v): got %d, want %d", arr, got, want)
		}
		a, b := arr[:len(arr)/2], arr[len(arr)/2:]
		if got, want := UniqueDescending(a, b), optimized.UniqueDescending(a, b); !slices.Equal(got, want) {
			t.Fatalf("UniqueDescending(%v, %v): got %v, want %v", a, b, got, want)
		}
		if got, want := concise.UniqueDescending(a, b), optimized.UniqueDescending(a, b); !slices.Equal(got, want) {
			t.Fatalf("concise.UniqueDescending(%v, %v): got %v, want %v", a, b, got, want)
		}
		if got, want := Count(num), optimized.Count(num); !maps.Equal(got, want) {
			t.Fatalf("Count(%d): got %v, want %v", num, got, want)
		}
		if got, want := concise.Count(num), optimized.Count(num); !maps.Equal(got, want) {
			t.Fatalf("concise.Count(%d): got %v, want %v", num, got, want)
		}
		if got, want := SumProducts(arr...), optimized.SumProducts(arr...); got != want {
			t.Fatalf("SumProducts(%v): got %d, want %d", arr, got, want)
		}
		if got, want := concise.SumProducts(arr...), optimized.SumProducts(arr...); got != want {
			t.Fatalf("concise.SumProducts(%v): got %d, want %d", arr, got, want)
		}
		if !slices.Equal(arr, before) {
			t.Fatal("simple version modified its input")
		}
		exclude := []int{0, int(mode % 10), int(mode) - 128, int(mode % 10)}
		for _, input := range [][]int{nums, {num}} {
			got, err := JoinExcluding(exclude, input...)
			want, wantErr := optimized.JoinExcluding(exclude, input...)
			if got != want || errorText(err) != errorText(wantErr) {
				t.Fatalf("JoinExcluding(%v, %v): got (%d, %v), want (%d, %v)", exclude, input, got, err, want, wantErr)
			}
			got, err = concise.JoinExcluding(exclude, input...)
			if got != want || errorText(err) != errorText(wantErr) {
				t.Fatalf("concise.JoinExcluding(%v, %v): got (%d, %v), want (%d, %v)", exclude, input, got, err, want, wantErr)
			}
		}
	})
}

func errorText(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}
