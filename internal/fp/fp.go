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
func ParseJSONFloatPrefix(data []byte) (f float64, n int, err error) {
	var mantissa uint64
	var exp int
	var neg, trunc, ok bool
	mantissa, exp, neg, trunc, n, ok = readFloat(data)
	if !ok {
		return 0, 0, errSyntax
	}

	// to match json behavior: a '.' at the end of the parsed data is an error.
	if n > 0 && data[n-1] == '.' {
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
	if !d.set(data[:n]) {
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
func readFloat(data []byte) (mantissa uint64, exp int, neg, trunc bool, p int, ok bool) {
	pe := len(data)
	if pe == 0 {
		return 0, 0, false, false, p, false
	}

	// optional sign
	if data[0] == '-' {
		neg = true
		p++
	}

	if p == pe {
		return 0, 0, false, false, p, false
	}

	const maxMantDigits = 19 // 10^19 fits in uint64
	sawdot := false
	sawdigits := false
	nd := 0
	ndMant := 0
	dp := 0

	// first digit

	switch data[p] {
	case '.':
		return mantissa, 0, neg, trunc, p, false
	case '0':
		sawdigits = true
		nd++
		ndMant++
	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		sawdigits = true
		mantissa += uint64(data[p] - '0')
		nd++
		ndMant++
	default:
		goto finishUp
	}
	p++

	if p == pe {
		goto finishUp
	}

	// second digit

	switch data[p] {
	case '.':
		sawdot = true
		dp = nd
	case '0':
		if mantissa == 0 {
			goto finishUp
		}
		sawdigits = true
		nd++
		ndMant++
		mantissa *= 10
	case '1', '2', '3', '4', '5', '6', '7', '8', '9':
		if mantissa == 0 {
			goto finishUp
		}
		sawdigits = true
		nd++
		ndMant++
		mantissa *= 10
		mantissa += uint64(data[p] - '0')
	default:
		goto finishUp
	}
	p++

	for ; p < len(data); p++ {
		if digits[data[p]] {
			nd++
			if ndMant >= maxMantDigits {
				trunc = true
				continue
			}
			mantissa *= 10
			mantissa += uint64(data[p] - '0')
			ndMant++
			continue
		}
		if data[p] == '.' {
			if sawdot {
				goto finishUp
			}
			sawdot = true
			dp = nd
			continue
		}
		break
	}

finishUp:
	if !sawdigits {
		return mantissa, 0, neg, trunc, p, false
	}
	if !sawdot {
		dp = nd
	}

	// optional exponent moves decimal point.
	// if we read a very large, very long number,
	// just be sure to move the decimal point by
	// a lot (say, 100000).  it doesn't matter if it's
	// not the exact number.
	if p < len(data) && (data[p] == 'e' || data[p] == 'E') {
		if data[p-1] == '.' {
			p = 0
			return mantissa, 0, neg, trunc, p, false
		}
		p++
		if p >= len(data) {
			return mantissa, 0, neg, trunc, p, false
		}
		esign := 1
		if data[p] == '+' {
			p++
		} else if data[p] == '-' {
			p++
			esign = -1
		}
		if p >= len(data) || data[p] < '0' || data[p] > '9' {
			return mantissa, 0, neg, trunc, p, false
		}
		e := 0
		for ; p < len(data) && (data[p] >= '0' && data[p] <= '9'); p++ {
			if e < 10000 {
				e = e*10 + int(data[p]) - '0'
			}
		}
		dp += e * esign
	}

	if mantissa != 0 {
		exp = dp - ndMant
	}

	return mantissa, exp, neg, trunc, p, true
}

var digits = [256]bool{
	'0': true,
	'1': true,
	'2': true,
	'3': true,
	'4': true,
	'5': true,
	'6': true,
	'7': true,
	'8': true,
	'9': true,
}

func (a *decimal) set(data []byte) (ok bool) {
	if len(data) == 0 {
		return false
	}
	a.neg = false
	a.trunc = false

	i := 0
	// optional sign
	if data[0] == '-' {
		a.neg = true
		i = 1
	}

	// digits
	sawdot := false
	sawdigits := false
	for ; i < len(data); i++ {
		switch {
		case data[i] == '.':
			if sawdot {
				return false
			}
			sawdot = true
			a.dp = a.nd
			continue

		case data[i] >= '0' && data[i] <= '9':
			sawdigits = true
			if data[i] == '0' && a.nd == 0 { // ignore leading zeros
				a.dp--
				continue
			}
			if a.nd < len(a.d) {
				a.d[a.nd] = data[i]
				a.nd++
			} else if data[i] != '0' {
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
	if i < len(data) && (data[i] == 'e' || data[i] == 'E') {
		i++
		if i >= len(data) {
			return false
		}
		esign := 1
		if data[i] == '+' {
			i++
		} else if data[i] == '-' {
			i++
			esign = -1
		}
		if i >= len(data) || data[i] < '0' || data[i] > '9' {
			return false
		}
		e := 0
		for ; i < len(data); i++ {
			if data[i] < '0' || data[i] > '9' {
				break
			}
			if e < 10000 {
				e = e*10 + int(data[i]) - '0'
			}
		}
		a.dp += e * esign
	}

	if i != len(data) {
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
