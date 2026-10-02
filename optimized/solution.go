// Package optimized contains the optimized solutions to the five exercises.
package optimized

import (
	"errors"
	"slices"
	"strconv"
	"sync"
)

// Closest (soal 1) finds the absent integer closest to zero on the input's side of zero.
// Mixed signs, empty input, and all zeros return zero.
func Closest(arr []int) int {
	var sign, count int
	for _, value := range arr {
		if value == 0 {
			continue
		}
		if sign == 0 {
			sign = 1
			if value < 0 {
				sign = -1
			}
		} else if (value > 0) != (sign > 0) {
			return 0
		}
		count++
	}
	if count == 0 {
		return 0
	}

	for i := range arr {
		for {
			value := arr[i]
			var target int
			if sign > 0 {
				if value < 1 || value > count {
					break
				}
				target = value - 1
			} else {
				// Check bounds before negation so MinInt cannot overflow.
				if value > -1 || value < -count {
					break
				}
				target = -value - 1
			}
			if arr[target] == value {
				break
			}
			// Every successful swap places a value in its final slot.
			arr[i], arr[target] = arr[target], value
		}
	}

	for i := 0; i < count; i++ {
		want := sign * (i + 1)
		if arr[i] != want {
			return want
		}
	}
	return sign * (count + 1)
}

var (
	// ErrNotFound means no digit remains after exclusion.
	ErrNotFound = errors.New("not found")
	// ErrOverflow means the joined integer or filtered result cannot fit in int.
	ErrOverflow = errors.New("integer overflow")
	// ErrNegativeNumber means an argument cannot be joined as an unsigned numeral.
	ErrNegativeNumber = errors.New("negative numbers are not supported")
)

const maxInt = int(^uint(0) >> 1)

// JoinExcluding (soal 2) joins nums into an int, then reverses its digits with exclusions.
// Empty input or no remaining digits returns ErrNotFound; a remaining zero returns (0, nil).
// Negative numbers return ErrNegativeNumber; integer overflow returns ErrOverflow.
func JoinExcluding(exclude []int, nums ...int) (int, error) {
	if len(nums) == 0 {
		return 0, ErrNotFound
	}
	for _, num := range nums {
		if num < 0 {
			return 0, ErrNegativeNumber
		}
	}
	var excluded [10]bool
	for _, digit := range exclude {
		if digit >= 0 && digit < len(excluded) {
			excluded[digit] = true
		}
	}

	var joined int
	for _, num := range nums {
		if joined == 0 {
			// Leading zero arguments vanish in the integer representation.
			joined = num
			continue
		}
		place := 1
		for place <= num/10 {
			place *= 10
		}
		for place > 0 {
			digit := num / place
			if joined > (maxInt-digit)/10 {
				return 0, ErrOverflow
			}
			joined = joined*10 + digit
			num %= place
			place /= 10
		}
	}

	var result int
	found := false
	for {
		digit := joined % 10
		if !excluded[digit] {
			if result > (maxInt-digit)/10 {
				return 0, ErrOverflow
			}
			result = result*10 + digit
			found = true
		}
		joined /= 10
		if joined == 0 {
			break
		}
	}
	if !found {
		return 0, ErrNotFound
	}
	return result, nil
}

// Small results avoid radix's fixed bucket initialization cost.
const radixThreshold = 1024

// UniqueDescending (soal 3) returns unique common values in descending order.
// An empty intersection returns nil. Neither input is modified.
func UniqueDescending(numsA, numsB []int) []int {
	if len(numsA) == 0 || len(numsB) == 0 {
		return nil
	}
	if len(numsA) > len(numsB) {
		numsA, numsB = numsB, numsA
	}
	members := make(map[int]struct{})
	for _, value := range numsA {
		members[value] = struct{}{}
	}
	var result []int
	for _, value := range numsB {
		if _, found := members[value]; !found {
			continue
		}
		result = append(result, value)
		delete(members, value)
		if len(members) == 0 {
			break
		}
	}
	sortDescending(result)
	return result
}

// sortDescending (soal 3) chooses the descending sort for the result.
func sortDescending(nums []int) {
	if len(nums) < 2 {
		return
	}
	ascending, descending := true, true
	for i := 1; i < len(nums); i++ {
		ascending = ascending && nums[i-1] <= nums[i]
		descending = descending && nums[i-1] >= nums[i]
		if !ascending && !descending {
			break
		}
	}
	if descending {
		return
	}
	if ascending {
		slices.Reverse(nums)
		return
	}
	if len(nums) <= radixThreshold {
		slices.Sort(nums)
		slices.Reverse(nums)
		return
	}
	radixSortDescending(nums)
}

// radixSortDescending (soal 3) sorts signed integers by their bytes.
func radixSortDescending(nums []int) {
	if len(nums) < 2 {
		return
	}
	// Equivalent to ^(uint(value) ^ signBit): ascending keys order signed
	// values descending, without subtracting or negating signed integers.
	const keyMask = ^(uint(1) << (strconv.IntSize - 1))
	src, dst := nums, make([]int, len(nums))
	for shift := uint(0); shift < strconv.IntSize; shift += 8 {
		var offsets [256]int
		for _, value := range src {
			digit := byte((uint(value) ^ keyMask) >> shift)
			offsets[digit]++
		}
		start := 0
		for digit, count := range offsets {
			offsets[digit] = start
			start += count
		}
		for _, value := range src {
			digit := byte((uint(value) ^ keyMask) >> shift)
			dst[offsets[digit]] = value
			offsets[digit]++
		}
		src, dst = dst, src
	}
	// Both supported int sizes require an even number of passes (4 or 8),
	// so the final pass writes into nums without an additional copy.
}

// Count (soal 4) splits num into integer digits, then groups and counts them.
// The sign is ignored. Zero has one zero digit, and MinInt is supported.
func Count(num int) map[int]int {
	// A 64-bit int has at most 19 decimal digits, including MinInt.
	var digits [19]int
	start := len(digits)
	for {
		digit := num % 10
		if digit < 0 {
			digit = -digit
		}
		start--
		digits[start] = digit
		num /= 10
		if num == 0 {
			break
		}
	}

	var counts [10]int
	distinct := 0
	for _, digit := range digits[start:] {
		if counts[digit] == 0 {
			distinct++
		}
		counts[digit]++
	}
	result := make(map[int]int, distinct)
	for digit, count := range counts {
		if count != 0 {
			result[digit] = count
		}
	}
	return result
}

// SumProducts (soal 5) multiplies each adjacent pair in its own goroutine.
// The caller sums the products; an unpaired final number is squared.
// Empty input returns zero. Arithmetic follows native int overflow semantics.
func SumProducts(nums ...int) int {
	pairs := len(nums)/2 + len(nums)%2
	if pairs == 0 {
		return 0
	}
	products := make([]int, pairs)
	var wg sync.WaitGroup
	wg.Add(pairs)
	for pair := range pairs {
		i := pair * 2
		// Each goroutine captures its own operands and result slot.
		index, a, b := pair, nums[i], nums[min(i+1, len(nums)-1)]
		go func() {
			defer wg.Done()
			products[index] = a * b
		}()
	}
	wg.Wait()
	var sum int
	for _, product := range products {
		sum += product
	}
	return sum
}
