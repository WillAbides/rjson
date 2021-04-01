package fp

import (
	"fmt"
	"math"
)

var (
	errSyntax = fmt.Errorf("syntax error")
	errRange  = fmt.Errorf("number out of float64 range")
)

// ParseJSONFloatPrefix is a bit like strconv.ParseFloat but it deals in byte slices instead of strings and
// has everything not needed for json stripped out.
func ParseJSONFloatPrefix(s []byte) (f float64, n int, err error) {
	var mantissa uint64
	var exp int
	var neg, trunc, ok bool
	mantissa, exp, neg, trunc, n, ok = readFloat(s)
	if !ok {
		return 0, n, errSyntax
	}

	// to match json behavior: a '.' at the end of the parsed data is an error.
	if n > 0 && s[n-1] == '.' {
		return 0, 0, errSyntax
	}

	// Try pure floating-point arithmetic conversion, and if that fails,
	// the Eisel-Lemire algorithm.
	if !trunc {
		if f2, ok := atof64exact(mantissa, exp, neg); ok {
			return f2, n, nil
		}
	}

	if f2, ok := eiselLemire64(mantissa, exp, neg); ok {
		if !trunc {
			return f2, n, nil
		}
		// Even if the mantissa was truncated, we may
		// have found the correct result. Confirm by
		// converting the upper mantissa bound.
		fUp, ok := eiselLemire64(mantissa+1, exp, neg)
		if ok && f2 == fUp {
			return f2, n, nil
		}
	}

	// Slow fallback.
	var d decimal
	if !d.set(s[:n]) {
		return 0, n, errSyntax
	}
	b, ovf := d.floatBits()
	f = math.Float64frombits(b)
	if ovf {
		err = errRange
	}
	if err != nil {
		return 0, n, err
	}
	return f, n, err
}

// readFloat reads a decimal mantissa and exponent from a float
// string representation in s; the number may be followed by other characters.
// readFloat reports the number of bytes consumed (i), and whether the number
// is valid (ok).
//
//nolint:gocyclo // yep...it's complex
func readFloat(data []byte) (mantissa uint64, exp int, neg, trunc bool, i int, ok bool) {
	if len(data) == 0 {
		return 0, 0, false, false, 0, false
	}

	// optional sign
	if data[0] == '-' {
		neg = true
		i++
	}

	const maxMantDigits = 19 // 10^19 fits in uint64
	sawdot := false
	sawdigits := false
	sawNonZero := false
	nd := 0
	ndMant := 0
	dp := 0
	startI := i
loop:
	for ; i < len(data); i++ {
		switch data[i] {
		case '.':
			if sawdot || !sawdigits {
				break loop
			}
			sawdot = true
			sawNonZero = true
			dp = nd
			continue
		case '0':
			// if it starts with 00 or -00, stop after the first 0 to comply with json's no leading zero rule
			// and match the behavior of json.Decoder's Token() method.
			if !sawNonZero && i > startI {
				break loop
			}

			sawdigits = true

			// This prevents leading zeros
			if nd == 0 &&
				len(data) > i+1 &&
				data[i+1] >= '0' && data[i+1] <= '9' {
				i++
				break loop
			}
			nd++
			if ndMant < maxMantDigits {
				mantissa *= 10
				ndMant++
			}
			continue
		case '1', '2', '3', '4', '5', '6', '7', '8', '9':
			sawdigits = true
			sawNonZero = true
			nd++
			if ndMant >= maxMantDigits {
				trunc = true
				continue
			}
			mantissa *= 10
			mantissa += uint64(data[i] - '0')
			ndMant++
			continue
		}
		break
	}
	if !sawdigits {
		return mantissa, 0, neg, trunc, i, false
	}
	if !sawdot {
		dp = nd
	}

	// optional exponent moves decimal point.
	// if we read a very large, very long number,
	// just be sure to move the decimal point by
	// a lot (say, 100000).  it doesn't matter if it's
	// not the exact number.
	if i < len(data) && (data[i] == 'e' || data[i] == 'E') {
		if data[i-1] == '.' {
			i = 0
			return mantissa, 0, neg, trunc, i, false
		}
		i++
		if i >= len(data) {
			return mantissa, 0, neg, trunc, i, false
		}
		esign := 1
		if data[i] == '+' {
			i++
		} else if data[i] == '-' {
			i++
			esign = -1
		}
		if i >= len(data) || data[i] < '0' || data[i] > '9' {
			return mantissa, 0, neg, trunc, i, false
		}
		e := 0
		for ; i < len(data) && (data[i] >= '0' && data[i] <= '9'); i++ {
			if e < 10000 {
				e = e*10 + int(data[i]) - '0'
			}
		}
		dp += e * esign
	}

	if mantissa != 0 {
		exp = dp - ndMant
	}

	return mantissa, exp, neg, trunc, i, true
}

func (a *decimal) set(s []byte) (ok bool) {
	if len(s) == 0 {
		return false
	}
	a.neg = false
	a.trunc = false

	i := 0
	// optional sign
	if s[0] == '-' {
		a.neg = true
		i = 1
	}

	// digits
	sawdot := false
	sawdigits := false
	for ; i < len(s); i++ {
		switch {
		case s[i] == '.':
			if sawdot {
				return false
			}
			sawdot = true
			a.dp = a.nd
			continue

		case s[i] >= '0' && s[i] <= '9':
			sawdigits = true
			if s[i] == '0' && a.nd == 0 { // ignore leading zeros
				a.dp--
				continue
			}
			if a.nd < len(a.d) {
				a.d[a.nd] = s[i]
				a.nd++
			} else if s[i] != '0' {
				a.trunc = true
			}
			continue
		}
		break
	}
	if !sawdigits {
		return false
	}
	if !sawdot {
		a.dp = a.nd
	}

	// optional exponent moves decimal point.
	// if we read a very large, very long number,
	// just be sure to move the decimal point by
	// a lot (say, 100000).  it doesn't matter if it's
	// not the exact number.
	if i < len(s) && (s[i] == 'e' || s[i] == 'E') {
		i++
		if i >= len(s) {
			return false
		}
		esign := 1
		if s[i] == '+' {
			i++
		} else if s[i] == '-' {
			i++
			esign = -1
		}
		if i >= len(s) || s[i] < '0' || s[i] > '9' {
			return false
		}
		e := 0
		for ; i < len(s); i++ {
			if s[i] < '0' || s[i] > '9' {
				break
			}
			if e < 10000 {
				e = e*10 + int(s[i]) - '0'
			}
		}
		a.dp += e * esign
	}

	if i != len(s) {
		return false
	}

	return true
}

// decimal power of ten to binary power of two.
var powtab = []int{1, 3, 6, 9, 13, 16, 19, 23, 26}

func (a *decimal) floatBits() (b uint64, overflow bool) {
	var exp int
	var mant uint64

	// Zero is always a special case.
	if a.nd == 0 {
		mant = 0
		exp = bias
		goto out
	}

	// Obvious overflow/underflow.
	// These bounds are for 64-bit floats.
	// Will have to change if we want to support 80-bit floats in the future.
	if a.dp > 310 {
		goto overflow
	}
	if a.dp < -330 {
		// zero
		mant = 0
		exp = bias
		goto out
	}

	// Scale by powers of two until in range [0.5, 1.0)
	exp = 0
	for a.dp > 0 {
		var n int
		if a.dp >= len(powtab) {
			n = 27
		} else {
			n = powtab[a.dp]
		}
		a.Shift(-n)
		exp += n
	}
	for a.dp < 0 || a.dp == 0 && a.d[0] < '5' {
		var n int
		if -a.dp >= len(powtab) {
			n = 27
		} else {
			n = powtab[-a.dp]
		}
		a.Shift(n)
		exp -= n
	}

	// Our range is [0.5,1) but floating point range is [1,2).
	exp--

	// Minimum representable exponent is bias+1.
	// If the exponent is smaller, move it up and
	// adjust a accordingly.
	if exp < bias+1 {
		n := bias + 1 - exp
		a.Shift(-n)
		exp += n
	}

	if exp-bias >= 1<<expbits-1 {
		goto overflow
	}

	// Extract 1+flt.mantbits bits.
	a.Shift(int(1 + mantbits))
	mant = a.RoundedInteger()

	// Rounding might have added a bit; shift down.
	if mant == 2<<mantbits {
		mant >>= 1
		exp++
		if exp-bias >= 1<<expbits-1 {
			goto overflow
		}
	}

	// Denormalized?
	if mant&(1<<mantbits) == 0 {
		exp = bias
	}
	goto out

overflow:
	// ±Inf
	mant = 0
	exp = 1<<expbits - 1 + bias
	overflow = true

out:
	// Assemble bits.
	bits := mant & (uint64(1)<<mantbits - 1)
	bits |= uint64((exp-bias)&(1<<expbits-1)) << mantbits
	if a.neg {
		bits |= 1 << mantbits << expbits
	}
	return bits, overflow
}

const (
	mantbits uint = 52
	expbits  uint = 11
	bias     int  = -1023
)

// Exact powers of 10.
var float64pow10 = []float64{
	1e0, 1e1, 1e2, 1e3, 1e4, 1e5, 1e6, 1e7, 1e8, 1e9,
	1e10, 1e11, 1e12, 1e13, 1e14, 1e15, 1e16, 1e17, 1e18, 1e19,
	1e20, 1e21, 1e22,
}

// If possible to convert decimal representation to 64-bit float f exactly,
// entirely in floating-point math, do so, avoiding the expense of decimalToFloatBits.
// Three common cases:
//	value is exact integer
//	value is exact integer * exact power of ten
//	value is exact integer / exact power of ten
// These all produce potentially inexact but correctly rounded answers.
func atof64exact(mantissa uint64, exp int, neg bool) (f float64, ok bool) {
	if mantissa>>mantbits != 0 {
		return
	}
	f = float64(mantissa)
	if neg {
		f = -f
	}
	switch {
	case exp == 0:
		// an integer.
		return f, true
	// Exact integers are <= 10^15.
	// Exact powers of ten are <= 10^22.
	case exp > 0 && exp <= 15+22: // int * 10^k
		// If exponent is big but number of digits is not,
		// can move a few zeros into the integer part.
		if exp > 22 {
			f *= float64pow10[exp-22]
			exp = 22
		}
		if f > 1e15 || f < -1e15 {
			// the exponent was really too large.
			return
		}
		return f * float64pow10[exp], true
	case exp < 0 && exp >= -22: // int / 10^k
		return f / float64pow10[-exp], true
	}
	return
}
