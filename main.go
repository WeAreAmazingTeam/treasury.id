package main

import (
	"fmt"

	solution "treasury.id/simple" // Ganti simple dengan optimized atau concise untuk mencoba versi lain.
)

func main() {
	fmt.Println("Soal 1:")
	fmt.Println(solution.Closest([]int{-1, 1}))
	fmt.Println(solution.Closest([]int{-1, -7, -5}))
	fmt.Println(solution.Closest([]int{1, 2, 1, 6}))

	fmt.Println("\nSoal 2:")
	fmt.Println(solution.JoinExcluding([]int{0, 1, 0}, 1000, 1010))
	fmt.Println(solution.JoinExcluding([]int{0, 1, 0}, 1259, 2601))

	fmt.Println("\nSoal 3:")
	fmt.Println(solution.UniqueDescending([]int{1, 2}, []int{1, 3}))
	fmt.Println(solution.UniqueDescending([]int{1, 2, 2}, []int{1, 2, 4}))

	fmt.Println("\nSoal 4:")
	counts := solution.Count(1223334)
	for _, digit := range []int{1, 2, 3, 4} {
		fmt.Printf("%d -> %d\n", digit, counts[digit])
	}

	fmt.Println("\nSoal 5:")
	fmt.Println(solution.SumProducts(1, 2, 3, 4))
	fmt.Println(solution.SumProducts(1, 2, 3, 4, 5))
}
