package quadratic_sieve

import (
	"math"
	"math/big"
)

func factorBase(n *big.Int) []*big.Int {
	/* calculate 'S' upper bound for the primes to collect */
	lnn := float64(n.BitLen()) * math.Log(2)

	if lnn < 1.0 {
		/* if this is not done. if n is 1 everything explodes sqrt(<0) = NaN */
		lnn = 1.0
	}

	lnlnn := math.Log(lnn)
	exp := math.Sqrt(lnn*lnlnn) * 0.5

	if exp >= 43 {
		/* this is reached when trying to Factorize about 2^1500 or larger */
		panic("factorBase(): exponent too large ... reimplement this using big ints")
	}

	S := int64(math.Ceil(math.Pow(math.E, exp))) // magic parameter (wikipedia)

	primes := make([]*big.Int, 1)
	primes[0] = MinusOne

	for p := int64(2); p <= S; p += 1 {
		if IsPrimeBruteForceSmallInt(p) {

			P := big.NewInt(int64(p))

			nModP := big.NewInt(0)
			nModP.Mod(n, P)

			if p == 2 || nModP.Cmp(Zero) == 0 {
				/* n is always a square rest (mod 2), and 0 is always a squarerest of 0 */
				primes = append(primes, P)
			} else {
				/* euler criterium. given ggt(a,p)=1: n is square rest mod p, iff n**((p-1)/2) \equiv 1 (mod p) */
				result := big.NewInt(0)
				result.Exp(n, big.NewInt((p-1)/2), P)
				result.Mod(result, P)

				if result.Cmp(One) == 0 {
					primes = append(primes, P)
				}
			}
		}
	}

	return primes
}

func sieveInterval(n *big.Int) (min, max *big.Int) {
	lnn := float64(n.BitLen()) * math.Log(2)

	if lnn < 1.0 {
		/* if this is not done. if n is 1 everything explodes sqrt(<0) = NaN */
		lnn = 1.0
	}

	lnlnn := math.Log(lnn)
	exp := int64(math.Ceil(math.Sqrt(lnn*lnlnn) * math.Log2(math.E)))

	L := big.NewInt(0)
	L.Exp(Two, big.NewInt(exp), nil) // magic parameter (wikipedia)

	sqrtN := SquareRootCeil(n)

	sieveMin := big.NewInt(0)
	sieveMin.Sub(sqrtN, L)
	sieveMax := big.NewInt(0)
	sieveMax.Add(sqrtN, L)

	return sieveMin, sieveMax
}

func sieve(n *big.Int, factorBase []*big.Int, cMin, cMax *big.Int) (retCis, retDis []*big.Int, retExponents [][]int) {
	intervalBig := big.NewInt(0)
	intervalBig.Sub(cMax, cMin)
	intervalBig.Add(intervalBig, One)

	if intervalBig.BitLen() > 31 {
		panic("fufufu sieve interval too large. code newly.")
	}

	retCis = make([]*big.Int, 0)
	retDis = make([]*big.Int, 0)
	retExponents = make([][]int, 0)

	ci := big.NewInt(0)
	ci.Set(cMin)

	di := big.NewInt(0)                       /* to be factorized */
	diCopy := big.NewInt(0)                   /* to be factorized */
	exponents := make([]int, len(factorBase)) /* exponents for the factors in factorbase for di */
	rest := big.NewInt(0)
	quotient := big.NewInt(0)

	/* foreach c(i) in [cMin, cMax] */
	for ; ci.Cmp(cMax) <= 0; ci.Add(ci, One) {

		/* d(i) = c(i)^2 - n */
		di.Mul(ci, ci)
		di.Sub(di, n)

		diCopy.Set(di)

		/* i = 0 (p = -1) needs special handling */
		if di.Sign() == -1 {
			exponents[0] = 1
			di.Mul(di, MinusOne)
		} else {
			exponents[0] = 0
		}

		for i := 1; i < len(factorBase); i += 1 {
			p := factorBase[i]
			exponents[i] = 0

			/* repeat as long as di % p == 0 -> add 1 to the exponent for each division by p */
			for {
				quotient.DivMod(di, p, rest)

				if rest.Cmp(Zero) == 0 {
					exponents[i] += 1
					di.Set(quotient)
				} else {
					break
				}

				if quotient.Cmp(Zero) == 0 {
					break
				}
			}
		}

		/* if d(i) is 1, d(i) has been successfully broken down and can be represented through
		the factor base -> save c(i) and the exponents for the prime factors in factorbase */
		if di.Cmp(One) == 0 {
			ciCopy := big.NewInt(0)
			ciCopy.Set(ci)
			diCopy2 := big.NewInt(0)
			diCopy2.Set(diCopy)
			exponentsCopy := make([]int, len(exponents))
			copy(exponentsCopy, exponents)

			retCis = append(retCis, ciCopy)
			retDis = append(retDis, diCopy2)
			retExponents = append(retExponents, exponentsCopy)
		}
	}

	return retCis, retDis, retExponents
}

func linearSystemFromExponents(exponents [][]int) *LinearSystem {
	if len(exponents) == 0 || len(exponents[0]) == 0 {
		panic("exponents is not supposed to be empty")
	}

	rows := len(exponents[0])
	columns := len(exponents)

	ret := NewLinearSystem(rows, columns)

	for i, column := range exponents {
		for j, value := range column {
			ret.Row(j).SetColumn(i, Bit(value%2))
		}
	}

	return ret
}

type abSquared struct {
	a, b *big.Int
}

// createPowerSetRecursively returns true if cancelled through the cancelChannel -> cascade collapse
func createPowerSetRecursively(currentIndex int, currentValues abSquared, set []abSquared,
	retChan chan<- abSquared, cancelChannel <-chan bool, doneChannel chan<- bool) bool {

	select {
	case <-cancelChannel:
		return true
	default:
	}

	if currentIndex == len(set) {
		return false
	}

	copyCurrentValues := abSquared{big.NewInt(0), big.NewInt(0)}
	copyCurrentValues.a.Set(currentValues.a)
	copyCurrentValues.b.Set(currentValues.b)

	if createPowerSetRecursively(currentIndex+1, copyCurrentValues, set, retChan, cancelChannel, doneChannel) == true {
		return true
	}

	currentValues.a.Mul(currentValues.a, set[currentIndex].a)
	currentValues.b.Mul(currentValues.b, set[currentIndex].b)

	copyCurrentValues = abSquared{big.NewInt(0), big.NewInt(0)}
	copyCurrentValues.a.Set(currentValues.a)
	copyCurrentValues.b.Set(currentValues.b)
	retChan <- copyCurrentValues

	if createPowerSetRecursively(currentIndex+1, currentValues, set, retChan, cancelChannel, doneChannel) == true {
		return true
	}

	if currentIndex == 0 {
		doneChannel <- true
	}

	return false
}

func findXY(n *big.Int, cis, dis []*big.Int, exponents [][]int) (*big.Int, *big.Int) {

	ls := linearSystemFromExponents(exponents)
	ls.GaussianElimination(ls)
	ls = ls.EliminateEmptyRows()
	ls = ls.Transpose()
	usedCombinations := ls.MakeEmptyRows()

	if len(usedCombinations) == 0 {
		return nil, nil
	}

	var aAndBSquaredList []abSquared

	for _, indexSet := range usedCombinations {

		a := big.NewInt(1)
		bb := big.NewInt(1)

		for _, i := range indexSet {
			a.Mul(a, cis[i])
			bb.Mul(bb, dis[i])
		}

		aAndBSquaredList = append(aAndBSquaredList, abSquared{a, bb})
	}

	x := big.NewInt(0)
	y := big.NewInt(0)

	xTimesY := big.NewInt(0)
	multiplicity := big.NewInt(0)
	testMod := big.NewInt(0)
	gcd := big.NewInt(0)

	abbChannel := make(chan abSquared, 100000)
	doneChannel := make(chan bool)
	cancelChannel := make(chan bool, 1)

	var abb abSquared

	go createPowerSetRecursively(0, abSquared{big.NewInt(1), big.NewInt(1)},
		aAndBSquaredList, abbChannel, cancelChannel, doneChannel)

	numberOfIndexSetsToGo := -1

	for attempts := 0; attempts < 100; attempts += 1 {
		if numberOfIndexSetsToGo == 0 {
			break
		}

		select {
		case <-doneChannel:
			numberOfIndexSetsToGo = len(abbChannel)
			continue
		case abb = <-abbChannel:
			if numberOfIndexSetsToGo > -1 {
				numberOfIndexSetsToGo -= 1
			}
		}

		x.SetInt64(0)
		y.SetInt64(0)

		abb.b = SquareRootCeil(abb.b)

		x.Add(abb.a, abb.b)
		x.Mod(x, n)

		y.Sub(abb.a, abb.b)
		y.Mod(y, n)

		if x.Cmp(Zero) == 0 || x.Cmp(One) == 0 || y.Cmp(Zero) == 0 || y.Cmp(One) == 0 {
			continue
		}

		xTimesY.Mul(x, y)
		multiplicity.DivMod(xTimesY, n, testMod)

		if testMod.Cmp(Zero) != 0 {
			continue
		}

		if multiplicity.Cmp(One) == 1 {

			gcd.GCD(nil, nil, x, multiplicity)
			if x.Cmp(gcd) != 0 {
				x.Div(x, gcd)
			}

			multiplicity.Div(multiplicity, gcd)

			gcd.GCD(nil, nil, y, multiplicity)
			if y.Cmp(gcd) != 0 {
				y.Div(y, gcd)
			}
		}

		cancelChannel <- true
		return x, y
	}

	cancelChannel <- true
	return nil, nil
}

// Factorize returns nil, nil if n cannot be factorized
func Factorize(n *big.Int) (*big.Int, *big.Int) {
	factorBase := factorBase(n)

	min, max := sieveInterval(n)

	cis, dis, exponents := sieve(n, factorBase, min, max)

	if len(cis) > 0 {
		x, y := findXY(n, cis, dis, exponents)

		if x != nil && y != nil && x.Cmp(y) == 1 {
			x, y = y, x
		}

		return x, y
	}

	return nil, nil
}
