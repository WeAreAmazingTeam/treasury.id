package concise

import (
	"errors"
	"slices"
	"strconv"
	"strings"

	"github.com/samber/lo"
	"github.com/samber/lo/mutable"
)

var (
	ErrNotFound       = errors.New("not found")
	ErrOverflow       = errors.New("integer overflow")
	ErrNegativeNumber = errors.New("negative numbers are not supported")
)

// Closest (soal 1) finds the absent integer closest to zero on the input's side of zero.
// Mixed signs, empty input, and all zeros return zero.
func Closest(arr []int) int {
	low, high := lo.Min(arr), lo.Max(arr)
	if low < 0 && high > 0 || low == 0 && high == 0 {
		return 0
	}
	step := 1
	if low < 0 {
		step = -1
	}
	value := step
	for lo.Contains(arr, value) {
		value += step
	}
	return value
}

// JoinExcluding (soal 2) joins nums into an int, then reverses its digits with exclusions.
// Empty input or no remaining digits returns ErrNotFound; a remaining zero returns (0, nil).
// Negative numbers return ErrNegativeNumber; integer overflow returns ErrOverflow.
func JoinExcluding(exclude []int, nums ...int) (int, error) {
	if len(nums) == 0 {
		return 0, ErrNotFound
	}
	if lo.Min(nums) < 0 {
		return 0, ErrNegativeNumber
	}
	parts := lo.Map(nums, func(num int, _ int) string { return strconv.Itoa(num) })
	joined, err := strconv.Atoi(strings.Join(parts, ""))
	if err != nil {
		return 0, ErrOverflow
	}
	digits := []byte(strconv.Itoa(joined))
	mutable.Reverse(digits)
	digits = lo.Filter(digits, func(digit byte, _ int) bool {
		return !lo.Contains(exclude, int(digit-'0'))
	})
	if len(digits) == 0 {
		return 0, ErrNotFound
	}
	result, err := strconv.Atoi(string(digits))
	if err != nil {
		return 0, ErrOverflow
	}
	return result, nil
}

// UniqueDescending (soal 3) returns unique common values in descending order.
// An empty intersection returns nil. Neither input is modified.
func UniqueDescending(numsA, numsB []int) []int {
	result := lo.Intersect(numsA, numsB)
	if len(result) == 0 {
		return nil
	}
	slices.Sort(result)
	mutable.Reverse(result)
	return result
}

// Count (soal 4) splits num into integer digits, then groups and counts them.
// The sign is ignored. Zero has one zero digit, and MinInt is supported.
func Count(num int) map[int]int {
	text := strings.TrimPrefix(strconv.Itoa(num), "-")
	digits := lo.Map([]byte(text), func(digit byte, _ int) int { return int(digit - '0') })
	return lo.CountValues(digits)
}

// SumProducts (soal 5) multiplies each adjacent pair in its own goroutine.
// The caller sums the products; an unpaired final number is squared.
// Empty input returns zero. Arithmetic follows native int overflow semantics.
func SumProducts(nums ...int) int {
	jobs := lo.Map(lo.Chunk(nums, 2), func(pair []int, _ int) <-chan int {
		a, b := pair[0], pair[len(pair)-1]
		return lo.Async(func() int { return a * b })
	})
	return lo.Reduce(jobs, func(total int, job <-chan int, _ int) int {
		return total + <-job
	}, 0)
}
