package dao

type fraction struct {
	numerator, denominator int
}

type delegatesForFraction struct {
	address string
	percent int
}

func (f fraction) Reduce() fraction {
	gcd := gcdEuclidean(f.numerator, f.denominator)
	f.numerator /= gcd
	f.denominator /= gcd

	return f
}

func gcdEuclidean(a, b int) int {
	for a != b {
		if a > b {
			a -= b
		} else {
			b -= a
		}
	}

	return a
}

func lcm(a, b int) int {
	return a * b / gcdEuclidean(a, b)
}

func calculateRatio(dPercents []delegatesForFraction) map[string]int {
	total := 0
	for _, d := range dPercents {
		total += d.percent
	}

	addressFractions := make(map[string]fraction)
	for _, d := range dPercents {
		addressFractions[d.address] = fraction{d.percent, total}
	}

	reducedFractions := make(map[string]fraction)
	for address, f := range addressFractions {
		reducedFractions[address] = f.Reduce()
	}

	lcmDenominator := 1
	for _, f := range reducedFractions {
		lcmDenominator = lcm(lcmDenominator, f.denominator)
	}

	resultFractions := make(map[string]fraction)
	for address, f := range reducedFractions {
		resultFractions[address] = fraction{f.numerator * (lcmDenominator / f.denominator), lcmDenominator}
	}

	result := make(map[string]int)
	for address, f := range resultFractions {
		result[address] = f.numerator
	}

	return result
}
