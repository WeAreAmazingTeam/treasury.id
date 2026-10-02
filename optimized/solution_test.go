package optimized

import (
	"encoding/binary"
	"errors"
	"fmt"
	"maps"
	"math/rand"
	"slices"
	"strconv"
	"strings"
	"sync"
	"testing"
)

func TestClosest(t *testing.T) {
	const maxInt = int(^uint(0) >> 1)
	const minInt = -maxInt - 1
	tests := []struct {
		name string
		arr  []int
		want int
	}{
		{"example_mixed", []int{-1, 1}, 0},
		{"example_negative", []int{-1, -7, -5}, -2},
		{"example_positive", []int{1, 2, 1, 6}, 3},
		{"nil", nil, 0},
		{"empty", []int{}, 0},
		{"zeros", []int{0, 0}, 0},
		{"positive_gap_first", []int{2, 3, 4}, 1},
		{"negative_gap_first", []int{-2, -3, -4}, -1},
		{"positive_complete", []int{3, 1, 2}, 4},
		{"negative_complete", []int{-3, -1, -2}, -4},
		{"duplicate", []int{1, 1, 1}, 2},
		{"negative_duplicate", []int{-1, -1}, -2},
		{"zero_positive", []int{0, 1, 2, 0}, 3},
		{"zero_negative", []int{0, -1, -3, 0}, -2},
		{"mixed_late", []int{1, 2, 3, 0, -1}, 0},
		{"sparse", []int{1, maxInt}, 2},
		{"min_int", []int{minInt, -1}, -2},
		{"extremes_mixed", []int{minInt, maxInt}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := slices.Clone(tt.arr)
			if got := Closest(input); got != tt.want {
				t.Fatalf("Closest(%v) = %d, want %d", tt.arr, got, tt.want)
			}
			if !maps.Equal(histogram(input), histogram(tt.arr)) {
				t.Fatalf("input values changed: before %v, after %v", tt.arr, input)
			}
		})
	}
}

func TestClosestLongSequence(t *testing.T) {
	for _, sign := range []int{1, -1} {
		arr := make([]int, 4096)
		for i := range arr {
			arr[i] = sign * (i + 1)
		}
		arr = append(arr, 0, sign, sign)
		rand.New(rand.NewSource(7)).Shuffle(len(arr), func(i, j int) {
			arr[i], arr[j] = arr[j], arr[i]
		})
		if got := Closest(arr); got != sign*4097 {
			t.Fatalf("sign %d: got %d, want %d", sign, got, sign*4097)
		}
	}
}

func FuzzClosest(f *testing.F) {
	f.Add([]byte{1, 2, 1, 6}, uint8(1))
	f.Add([]byte{1, 7, 5}, uint8(2))
	f.Add([]byte{255, 1}, uint8(0))
	f.Add([]byte{0, 0}, uint8(1))
	f.Fuzz(func(t *testing.T, data []byte, mode uint8) {
		if len(data) > 4096 {
			t.Skip()
		}
		arr := make([]int, len(data))
		for i, value := range data {
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
		want := missingOracle(before)
		if got := Closest(arr); got != want {
			t.Fatalf("Closest(%v) = %d, want %d", before, got, want)
		}
		if !maps.Equal(histogram(arr), histogram(before)) {
			t.Fatalf("lost input values: before %v, after %v", before, arr)
		}
		if got := Closest(arr); got != want {
			t.Fatalf("second call = %d, want %d", got, want)
		}
	})
}

func missingOracle(arr []int) int {
	positive, negative := false, false
	seen := make(map[int]bool)
	for _, value := range arr {
		positive = positive || value > 0
		negative = negative || value < 0
		seen[value] = true
	}
	if positive == negative {
		return 0
	}
	step := 1
	if negative {
		step = -1
	}
	for candidate := step; ; candidate += step {
		if !seen[candidate] {
			return candidate
		}
	}
}

func histogram(arr []int) map[int]int {
	counts := make(map[int]int)
	for _, value := range arr {
		counts[value]++
	}
	return counts
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

func TestJoinExcluding(t *testing.T) {
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
		{"no_arguments", nil, nil, 0, ErrNotFound},
		{"zero", nil, []int{0}, 0, nil},
		{"zero_excluded", []int{0}, []int{0}, 0, ErrNotFound},
		{"all_zeros", nil, []int{0, 0, 0}, 0, nil},
		{"leading_zeros", nil, []int{0, 0, 12}, 21, nil},
		{"trailing_zeros", nil, []int{120, 0}, 21, nil},
		{"zero_survives", []int{1, 2}, []int{120}, 0, nil},
		{"outside_digit_range", []int{-1, 10, maxInt}, []int{12}, 21, nil},
		{"duplicate_exclusion", []int{0, 1, 1}, []int{1012}, 2, nil},
		{"negative", nil, []int{-12}, 0, ErrNegativeNumber},
		{"negative_later", nil, []int{1, -2}, 0, ErrNegativeNumber},
		{"negative_before_calculation", nil, []int{maxInt, 0, -1}, 0, ErrNegativeNumber},
		{"min_int", nil, []int{-maxInt - 1}, 0, ErrNegativeNumber},
		{"max_int_filtered", allDigits, []int{maxInt}, 0, ErrNotFound},
		{"join_at_limit", allDigits, []int{maxInt / 10, maxInt % 10}, 0, ErrNotFound},
		{"join_past_limit", nil, []int{maxInt / 10, maxInt%10 + 1}, 0, ErrOverflow},
		{"join_overflow_even_filtered", allDigits, []int{maxInt, 0}, 0, ErrOverflow},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exclude, nums := slices.Clone(tt.exclude), slices.Clone(tt.nums)
			got, err := JoinExcluding(exclude, nums...)
			if got != tt.want || !errors.Is(err, tt.err) {
				t.Fatalf("JoinExcluding(%v, %v) = (%d, %v), want (%d, %v)", tt.exclude, tt.nums, got, err, tt.want, tt.err)
			}
			if tt.err == ErrNotFound && err.Error() != "not found" {
				t.Fatalf("error message = %q, want %q", err.Error(), "not found")
			}
			if !slices.Equal(exclude, tt.exclude) || !slices.Equal(nums, tt.nums) {
				t.Fatal("inputs were modified")
			}
		})
	}
}

func TestJoinExcludingReverseOverflow(t *testing.T) {
	num := 1563847412 // Fits int32; its reverse, 2147483651, does not.
	if strconv.IntSize == 64 {
		var err error
		num, err = strconv.Atoi("8085774586302733229")
		if err != nil {
			t.Fatal(err)
		}
	}
	if got, err := JoinExcluding(nil, num); got != 0 || !errors.Is(err, ErrOverflow) {
		t.Fatalf("JoinExcluding(nil, %d) = (%d, %v), want (0, ErrOverflow)", num, got, err)
	}
}

func TestJoinExcludingAliasedInputs(t *testing.T) {
	shared := []int{1, 2, 3}
	if got, err := JoinExcluding(shared[:1], shared[1:]...); got != 32 || err != nil {
		t.Fatalf("got (%d, %v), want (32, nil)", got, err)
	}
	if !slices.Equal(shared, []int{1, 2, 3}) {
		t.Fatalf("shared input modified: %v", shared)
	}
}

func FuzzJoinExcluding(f *testing.F) {
	f.Add([]byte{0, 1, 0}, []byte{12, 59, 26, 1})
	f.Add([]byte{}, []byte{0, 12})
	f.Add([]byte{0}, []byte{0})
	f.Fuzz(func(t *testing.T, excluded, data []byte) {
		if len(excluded) > 256 || len(data) > 128 {
			t.Skip()
		}
		exclude := make([]int, len(excluded))
		for i, value := range excluded {
			exclude[i] = int(int8(value))
		}
		nums := make([]int, len(data))
		for i, value := range data {
			nums[i] = int(value)
		}
		want, wantErr := reverseOracle(exclude, nums)
		got, err := JoinExcluding(exclude, nums...)
		if got != want || !errors.Is(err, wantErr) {
			t.Fatalf("JoinExcluding(%v, %v) = (%d, %v), oracle = (%d, %v)", exclude, nums, got, err, want, wantErr)
		}
	})
}

func reverseOracle(exclude, nums []int) (int, error) {
	if len(nums) == 0 {
		return 0, ErrNotFound
	}
	var text strings.Builder
	for _, num := range nums {
		if num < 0 {
			return 0, ErrNegativeNumber
		}
		text.WriteString(strconv.Itoa(num))
	}
	joined, err := strconv.Atoi(text.String())
	if err != nil {
		return 0, ErrOverflow
	}
	digits := strconv.Itoa(joined)
	var filtered strings.Builder
	for i := len(digits) - 1; i >= 0; i-- {
		if !slices.Contains(exclude, int(digits[i]-'0')) {
			filtered.WriteByte(digits[i])
		}
	}
	if filtered.Len() == 0 {
		return 0, ErrNotFound
	}
	result, err := strconv.Atoi(filtered.String())
	if err != nil {
		return 0, ErrOverflow
	}
	return result, nil
}

func FuzzSingleNumber(f *testing.F) {
	for _, num := range []int{0, 120, 1223334, maxInt, -maxInt - 1} {
		f.Add(num, uint16(0))
		f.Add(num, uint16(1023))
	}
	f.Fuzz(func(t *testing.T, num int, mask uint16) {
		var exclude []int
		for digit := range 10 {
			if mask&(1<<digit) != 0 {
				exclude = append(exclude, digit)
			}
		}
		want, wantErr := reverseOracle(exclude, []int{num})
		got, err := JoinExcluding(exclude, num)
		if got != want || !errors.Is(err, wantErr) {
			t.Fatalf("JoinExcluding(%v, %d) = (%d, %v), oracle = (%d, %v)", exclude, num, got, err, want, wantErr)
		}
	})
}

func ExampleJoinExcluding() {
	fmt.Println(JoinExcluding([]int{0, 1, 0}, 1000, 1010))
	fmt.Println(JoinExcluding([]int{0, 1, 0}, 1259, 2601))
	// Output:
	// 0 not found
	// 62952 <nil>
}

func TestUniqueDescending(t *testing.T) {
	const maxInt = int(^uint(0) >> 1)
	const minInt = -maxInt - 1
	tests := []struct {
		name string
		a, b []int
		want []int
	}{
		{"example_single", []int{1, 2}, []int{1, 3}, []int{1}},
		{"example_unique", []int{1, 2, 2}, []int{1, 2, 4}, []int{2, 1}},
		{"nil", nil, nil, nil},
		{"empty_a", nil, []int{1}, nil},
		{"empty_b", []int{1}, nil, nil},
		{"disjoint", []int{1, 2}, []int{3, 4}, nil},
		{"duplicates_both", []int{1, 2, 2, 1}, []int{2, 1, 1, 2}, []int{2, 1}},
		{"negative_zero", []int{-2, 0, 3, -1}, []int{0, -1, 3}, []int{3, 0, -1}},
		{"shorter_b", []int{5, 3, 1, 3, 7, 0}, []int{3, 5}, []int{5, 3}},
		{"extremes", []int{minInt, 0, maxInt}, []int{maxInt, minInt, 0}, []int{maxInt, 0, minInt}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a, b := slices.Clone(tt.a), slices.Clone(tt.b)
			got := UniqueDescending(a, b)
			if !slices.Equal(got, tt.want) || (got == nil) != (tt.want == nil) {
				t.Fatalf("UniqueDescending(%v, %v) = %v, want %v", tt.a, tt.b, got, tt.want)
			}
			if !slices.Equal(a, tt.a) || !slices.Equal(b, tt.b) {
				t.Fatal("inputs were modified")
			}
		})
	}
}

func TestUniqueDescendingAliasedInputs(t *testing.T) {
	shared := []int{4, 2, 2, -1, 9, 7}
	before := slices.Clone(shared)
	if got := UniqueDescending(shared[:5], shared[2:]); !slices.Equal(got, []int{9, 2, -1}) {
		t.Fatalf("got %v, want [9 2 -1]", got)
	}
	if !slices.Equal(shared, before) {
		t.Fatal("shared backing array modified")
	}
}

func TestUniqueDescendingLargeIntersection(t *testing.T) {
	a := make([]int, 2*radixThreshold+1)
	for i := range a {
		a[i] = (i - len(a)/2) * 7919
	}
	b := append(slices.Clone(a), a[0], a[len(a)-1])
	rng := rand.New(rand.NewSource(11))
	rng.Shuffle(len(a), func(i, j int) { a[i], a[j] = a[j], a[i] })
	rng.Shuffle(len(b), func(i, j int) { b[i], b[j] = b[j], b[i] })
	want := intersectionOracle(a, b)
	if got := UniqueDescending(a, b); !slices.Equal(got, want) {
		t.Fatalf("large intersection differs from oracle (size %d)", len(want))
	}
}

func TestSortDescending(t *testing.T) {
	const maxInt = int(^uint(0) >> 1)
	tests := [][]int{
		nil, {}, {1}, {1, 1, 1}, {0, -1, maxInt, -maxInt - 1, 1},
		{-4, -2, 0, 2, 4}, {4, 2, 0, -2, -4},
	}
	rng := rand.New(rand.NewSource(13))
	for _, size := range []int{radixThreshold - 1, radixThreshold, radixThreshold + 1, 4 * radixThreshold} {
		input := make([]int, size)
		for i := range input {
			input[i] = int(rng.Uint64())
		}
		input[0], input[1] = maxInt, -maxInt-1
		tests = append(tests, input)
	}
	for i, input := range tests {
		t.Run(fmt.Sprint(i), func(t *testing.T) {
			want := slices.Clone(input)
			slices.Sort(want)
			slices.Reverse(want)
			for name, sort := range map[string]func([]int){"adaptive": sortDescending, "radix": radixSortDescending} {
				got := slices.Clone(input)
				sort(got)
				if !slices.Equal(got, want) {
					t.Fatalf("%s differs from standard sort, size %d", name, len(input))
				}
			}
		})
	}
}

func FuzzUniqueDescending(f *testing.F) {
	f.Add([]byte{1, 2, 2}, []byte{1, 2, 4})
	f.Add([]byte{0, 255, 128}, []byte{128, 0})
	f.Fuzz(func(t *testing.T, dataA, dataB []byte) {
		if len(dataA) > 512 || len(dataB) > 512 {
			t.Skip()
		}
		a, b := make([]int, len(dataA)), make([]int, len(dataB))
		for i, value := range dataA {
			a[i] = int(int8(value))
		}
		for i, value := range dataB {
			b[i] = int(int8(value))
		}
		beforeA, beforeB := slices.Clone(a), slices.Clone(b)
		if got, want := UniqueDescending(a, b), intersectionOracle(a, b); !slices.Equal(got, want) {
			t.Fatalf("UniqueDescending(%v, %v) = %v, want %v", a, b, got, want)
		}
		if !slices.Equal(a, beforeA) || !slices.Equal(b, beforeB) {
			t.Fatal("inputs were modified")
		}
	})
}

func FuzzSortDescending(f *testing.F) {
	seed := make([]byte, (radixThreshold+1)*8)
	rng := rand.New(rand.NewSource(17))
	for i := 0; i < len(seed); i += 8 {
		binary.LittleEndian.PutUint64(seed[i:], rng.Uint64())
	}
	f.Add(seed)
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 65536 {
			t.Skip()
		}
		input := make([]int, len(data)/8)
		for i := range input {
			input[i] = int(binary.LittleEndian.Uint64(data[i*8:]))
		}
		want := slices.Clone(input)
		slices.Sort(want)
		slices.Reverse(want)
		sortDescending(input)
		if !slices.Equal(input, want) {
			t.Fatalf("sort differs from oracle, size %d", len(input))
		}
	})
}

func intersectionOracle(a, b []int) []int {
	a, b = slices.Clone(a), slices.Clone(b)
	slices.Sort(a)
	slices.Sort(b)
	var result []int
	for i, j := 0, 0; i < len(a) && j < len(b); {
		switch {
		case a[i] < b[j]:
			i++
		case a[i] > b[j]:
			j++
		default:
			if len(result) == 0 || result[len(result)-1] != a[i] {
				result = append(result, a[i])
			}
			i++
			j++
		}
	}
	slices.Reverse(result)
	return result
}

func ExampleUniqueDescending() {
	fmt.Println(UniqueDescending([]int{1, 2}, []int{1, 3}))
	fmt.Println(UniqueDescending([]int{1, 2, 2}, []int{1, 2, 4}))
	// Output:
	// [1]
	// [2 1]
}

func TestCount(t *testing.T) {
	const maxInt = int(^uint(0) >> 1)
	tests := []struct {
		name string
		num  int
		want map[int]int
	}{
		{"example", 1223334, map[int]int{1: 1, 2: 2, 3: 3, 4: 1}},
		{"zero", 0, map[int]int{0: 1}},
		{"single_digit", 7, map[int]int{7: 1}},
		{"negative_single_digit", -7, map[int]int{7: 1}},
		{"repeated", 99999, map[int]int{9: 5}},
		{"zeros_inside", 100200, map[int]int{0: 4, 1: 1, 2: 1}},
		{"negative", -1223334, map[int]int{1: 1, 2: 2, 3: 3, 4: 1}},
		{"min_int", -maxInt - 1, digitsOracle(-maxInt - 1)},
		{"max_int", maxInt, digitsOracle(maxInt)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Count(tt.num)
			if !maps.Equal(got, tt.want) {
				t.Fatalf("Count(%d) = %v, want %v", tt.num, got, tt.want)
			}
			for digit, count := range got {
				if digit < 0 || digit > 9 || count <= 0 {
					t.Fatalf("invalid histogram entry %d: %d", digit, count)
				}
			}
		})
	}
}

func TestCountIndependentResults(t *testing.T) {
	first := Count(111)
	first[1] = 99
	if got := Count(111); !maps.Equal(got, map[int]int{1: 3}) {
		t.Fatalf("result shared with a previous call: %v", got)
	}
}

func FuzzCount(f *testing.F) {
	const maxInt = int(^uint(0) >> 1)
	for _, num := range []int{0, 1223334, -1223334, maxInt, -maxInt - 1} {
		f.Add(num)
	}
	f.Fuzz(func(t *testing.T, num int) {
		if got, want := Count(num), digitsOracle(num); !maps.Equal(got, want) {
			t.Fatalf("Count(%d) = %v, oracle = %v", num, got, want)
		}
	})
}

func digitsOracle(num int) map[int]int {
	result := make(map[int]int)
	for _, digit := range strconv.Itoa(num) {
		if digit != '-' {
			result[int(digit-'0')]++
		}
	}
	return result
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

func TestSumProducts(t *testing.T) {
	maxInt := int(^uint(0) >> 1)
	tests := []struct {
		name string
		nums []int
		want int
	}{
		{"example_even", []int{1, 2, 3, 4}, 14},
		{"example_odd", []int{1, 2, 3, 4, 5}, 39},
		{"nil", nil, 0},
		{"empty", []int{}, 0},
		{"one_positive", []int{5}, 25},
		{"one_negative", []int{-5}, 25},
		{"one_zero", []int{0}, 0},
		{"zero_pair", []int{0, 123, 4}, 16},
		{"negative_products", []int{-1, 2, 3, -4}, -14},
		{"negative_odd", []int{-1, -2, -3}, 11},
		{"multiplication_overflow", []int{maxInt, 2}, maxInt * 2},
		{"sum_overflow", []int{maxInt, 1, 1, 1}, maxInt + 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nums := slices.Clone(tt.nums)
			if got := SumProducts(nums...); got != tt.want {
				t.Fatalf("SumProducts(%v) = %d, want %d", tt.nums, got, tt.want)
			}
			if !slices.Equal(nums, tt.nums) {
				t.Fatal("input was modified")
			}
		})
	}
}

func TestSumProductsLargeInputs(t *testing.T) {
	for _, size := range []int{63, 64, 65, 1024, 1025, 10001} {
		nums := make([]int, size)
		for i := range nums {
			nums[i] = i%31 - 15
		}
		if got, want := SumProducts(nums...), pairOracle(nums); got != want {
			t.Fatalf("size %d: got %d, want %d", size, got, want)
		}
	}
}

func TestSumProductsConcurrentCalls(t *testing.T) {
	var callers sync.WaitGroup
	for caller := range 8 {
		callers.Add(1)
		go func(id int) {
			defer callers.Done()
			nums := make([]int, 129+id)
			for i := range nums {
				nums[i] = i%17 - 8
			}
			want := pairOracle(nums)
			for range 8 {
				if got := SumProducts(nums...); got != want {
					t.Errorf("caller %d: got %d, want %d", id, got, want)
					return
				}
			}
		}(caller)
	}
	callers.Wait()
}

func pairOracle(nums []int) int {
	var result int
	for i := 0; i < len(nums); i += 2 {
		if i+1 < len(nums) {
			result += nums[i] * nums[i+1]
		} else {
			result += nums[i] * nums[i]
		}
	}
	return result
}

func ExampleSumProducts() {
	fmt.Println(SumProducts(1, 2, 3, 4))
	fmt.Println(SumProducts(1, 2, 3, 4, 5))
	// Output:
	// 14
	// 39
}
