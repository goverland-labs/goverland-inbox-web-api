package dao

import (
	"math"
)

type delegatesForFraction struct {
	address string
	percent int
}

// Function to calculate the total weight and proportions
func calculateProportions(weights []int) (int, []float64) {
	total := 0
	for _, w := range weights {
		total += w
	}

	proportions := make([]float64, len(weights))
	for i, w := range weights {
		proportions[i] = float64(w) / float64(total) * 100 // Convert to percentage
	}

	return total, proportions
}

// Function to find approximate proportions with constraints
func approximateProportions(weights []int) []int {
	totalWeight, proportions := calculateProportions(weights)
	minWeight := math.MaxInt
	var bestWeights []int

	// Try different total weights from the sum of the original weights down to a reasonable limit
	for total := 1; total <= totalWeight; total++ {
		newWeights := make([]int, len(weights))
		for i := range weights {
			newWeights[i] = int(math.Round(proportions[i] / 100 * float64(total)))
		}

		// Check if the new weights meet the 1% difference requirement
		valid := true
		for i := range weights {
			if newWeights[i] > 0 {
				originalProportion := float64(weights[i]) / float64(totalWeight) * 100
				newProportion := float64(newWeights[i]) / float64(total) * 100
				if math.Abs(originalProportion-newProportion) > 1 {
					valid = false
					break
				}
			} else {
				valid = false
				break
			}
		}

		if valid {
			// Check if the current new weights are lower than the previously found minimum
			currentSum := sum(newWeights)
			if currentSum < minWeight {
				minWeight = currentSum
				bestWeights = newWeights
			}

			if currentSum < len(weights)*10 {
				break
			}
		}
	}

	return bestWeights
}

// Function to calculate the sum of elements in an array
func sum(arr []int) int {
	total := 0
	for _, v := range arr {
		total += v
	}
	return total
}

func calculateRatio(dPercents []delegatesForFraction) map[string]int {
	// Create an array of percentages
	percentages := make([]int, len(dPercents))
	for i, d := range dPercents {
		percentages[i] = d.percent
	}

	// Calculate proportional weights
	weights := approximateProportions(percentages)

	// Create a map of addresses to weights
	result := make(map[string]int)
	for i, d := range dPercents {
		result[d.address] = weights[i]
	}

	return result
}
