// Package simple provides straightforward solutions to the five technical exercises.
package simple

import (
	"errors"
	"slices"
	"sort"
	"strconv"
	"strings"
)

// Closest (soal 1) finds the absent integer closest to zero on the input's side of zero.
// Mixed signs, empty input, and all zeros return zero.
func Closest(arr []int) int {
	seen := make(map[int]bool)
	positive, negative := false, false
	for _, value := range arr {
		seen[value] = true
		positive = positive || value > 0
		negative = negative || value < 0
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

var (
	ErrNotFound       = errors.New("not found")
	ErrOverflow       = errors.New("integer overflow")
	ErrNegativeNumber = errors.New("negative numbers are not supported")
)

// JoinExcluding (soal 2) joins nums into an int, then reverses its digits with exclusions.
// Empty input or no remaining digits returns ErrNotFound; a remaining zero returns (0, nil).
// Negative numbers return ErrNegativeNumber; integer overflow returns ErrOverflow.
func JoinExcluding(exclude []int, nums ...int) (int, error) {
	if len(nums) == 0 {
		return 0, ErrNotFound
	}
	var joinedText strings.Builder
	for _, num := range nums {
		if num < 0 {
			return 0, ErrNegativeNumber
		}
		joinedText.WriteString(strconv.Itoa(num))
	}
	joined, err := strconv.Atoi(joinedText.String())
	if err != nil {
		return 0, ErrOverflow
	}

	var excluded [10]bool
	for _, digit := range exclude {
		if digit >= 0 && digit < len(excluded) {
			excluded[digit] = true
		}
	}
	text := strconv.Itoa(joined)
	var filtered strings.Builder
	for i := len(text) - 1; i >= 0; i-- {
		if !excluded[text[i]-'0'] {
			filtered.WriteByte(text[i])
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

// UniqueDescending (soal 3) returns unique common values in descending order.
// An empty intersection returns nil. Neither input is modified.
func UniqueDescending(numsA, numsB []int) []int {
	if len(numsA) > len(numsB) {
		numsA, numsB = numsB, numsA
	}
	seen := make(map[int]bool)
	for _, value := range numsA {
		seen[value] = true
	}
	var result []int
	for _, value := range numsB {
		if seen[value] {
			result = append(result, value)
			delete(seen, value)
		}
	}
	sort.Sort(sort.Reverse(sort.IntSlice(result)))
	return result
}

// Count (soal 4) splits num into integer digits, then groups and counts them.
// The sign is ignored. Zero has one zero digit, and MinInt is supported.
func Count(num int) map[int]int {
	var digits []int
	for {
		digit := num % 10
		if digit < 0 {
			digit = -digit
		}
		digits = append(digits, digit)
		num /= 10
		if num == 0 {
			break
		}
	}
	slices.Reverse(digits)

	counts := make(map[int]int)
	for _, digit := range digits {
		counts[digit]++
	}
	return counts
}

// SumProducts (soal 5) multiplies each adjacent pair in its own goroutine.
// The caller sums the products; an unpaired final number is squared.
// Empty input returns zero. Arithmetic follows native int overflow semantics.
func SumProducts(nums ...int) int {
	pairs := len(nums)/2 + len(nums)%2
	products := make(chan int, pairs)
	for i := 0; i < len(nums); i += 2 {
		a, b := nums[i], nums[i]
		if i+1 < len(nums) {
			b = nums[i+1]
		}
		go func(a, b int) {
			products <- a * b
		}(a, b)
	}

	var sum int
	for range pairs {
		sum += <-products
	}
	return sum
}
