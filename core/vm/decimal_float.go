package vm

import (
	"fmt" // TODO remove

	"github.com/holiman/uint256"
)

// s * 10^e
// consider s and e as int256 using 2's complement
type decimal struct {
	s uint256.Int // significand
	e uint256.Int // exponent
}

var ZERO = uint256.NewInt(0)
var TEN = uint256.NewInt(10)

func isNegativeIfInterpretedAsInt256(value *uint256.Int) bool {
	msb := new(uint256.Int).Lsh(uint256.NewInt(1), 255) // 2^255
	return new(uint256.Int).And(value, msb).Sign() != 0
}

// c.s*10^c.e = a.s*10^a.e + b.s*10^b.e
func add(a, b, out *decimal, L bool) (*decimal) {
	if L {fmt.Println("add", "a", "b", a, b)}

	aqmbq := new(uint256.Int).Sub(&a.e, &b.e)
	if L {fmt.Println("add", "aqmbq", aqmbq)}
	
	aqmbq_abs := new(uint256.Int).Abs(aqmbq)
	if L {fmt.Println("add", "aqmbq_abs", aqmbq_abs)}

	ten_power := uint256.NewInt(10)
	ten_power.Exp(ten_power, aqmbq_abs) // todo faster way should exist
	if L {fmt.Println("add", "ten_power", ten_power, ten_power.String())}

	ca := new(uint256.Int).Abs(&a.s)
	if L {fmt.Println("add", "ca", ca, ca.String())}

	cb := new(uint256.Int).Abs(&b.s)
	if L {fmt.Println("add", "cb", cb, cb.String())}

	aqmbq_neg := isNegativeIfInterpretedAsInt256(aqmbq)
	if !aqmbq_neg {
		ca.Mul(ca, ten_power)
	} else if aqmbq.Cmp(ZERO) != 0 {
		cb.Mul(cb, ten_power)
	}
	if L {fmt.Println("add", "ca", ca, ca.String())}
	if L {fmt.Println("add", "cb", cb, cb.String())}

	c := new(uint256.Int).Add(ca, cb)
	if L {fmt.Println("add", "c", c, c.String())}
	
	// min(a.e, b.e)
	ae_neg := isNegativeIfInterpretedAsInt256(&a.e)
	if L {fmt.Println("add", "b.e", b.e)}
	be_neg := isNegativeIfInterpretedAsInt256(&b.e)
	e := a.e
	if L {fmt.Println("add", "a.e", a.e)}
	if L {fmt.Println("add", "b.e", b.e)}
	if L {fmt.Println("add", "ae_neg", ae_neg)}
	if L {fmt.Println("add", "be_neg", be_neg)}
	if L {fmt.Println("add", "e", e)}
	if ae_neg && !be_neg {
		if L {fmt.Println("1")}
	} else if !ae_neg && be_neg {
		if L {fmt.Println("2")}
		e = b.e
	} else if a.e.Cmp(&b.e) == 1 {
		if L {fmt.Println("3")}
		e = b.e
	}
	if L {fmt.Println("add", "e", e)}
	// min(a.e, b.e)

	out.s = *c
	out.e = e
	if L {fmt.Println("add", "out", out)}

	return out
}

// -a
func negate(a, out *decimal, L bool) (*decimal) {
	out.s.Neg(&a.s)
	out.e = a.e
	return out
}

// a - b
func subtract(a, b, out *decimal, precision int64, L bool) (*decimal) {
	negate(b, out, L)
	add(a, out, out, L)
	return out
}

// a * b
func multiply(a, b, out *decimal, L bool) (*decimal) {
	// if L {fmt.Println("multiply", "a", String(a), "b", String(b), "precision", precision)}
	// if L {fmt.Println("multiply", "a", a, "b", b)}
	out.s.Mul(&a.s, &b.s)
	// if L {fmt.Println("multiply", "out.c", out.c)}
	out.e.Add(&a.e, &b.e)
	// if L {fmt.Println("multiply", "out.q", out.q)}
	// return normalize(copy(out), out, precision, false, L)
	return out
}

// 1 / a
func inverse(a, out *decimal, precision uint256.Int, L bool) (*decimal) {
	// if L {fmt.Println("inverse", "a", String(a), "precision", precision)}

	// out.n = a.n
	// if L {fmt.Println("inverse", "out.n", out.n)}

	ten_power := TEN
	ae_m_precision := new(uint256.Int).Neg(&a.e)
	ae_m_precision.Add(ae_m_precision, &precision)
	ten_power.Exp(ten_power, ae_m_precision)
	out.s.Div(ten_power, &a.s)

	if L {fmt.Println("inverse", "out.s", out.s)}
	
	out.s.Sub(&out.s, &precision)

	if L {fmt.Println("inverse", "out.s", out.s)}
	// if L {fmt.Println("inverse", "out", out, String(out))}
	
	// norm := normalize(copy(out), out, precision, false, L)
	// if L {fmt.Println("inverse", "norm", norm, String(norm))}
	// return norm
	
	// c = round(BigInt(10)^(-x.q + DIGITS) / x.c) # the decimal point of 1/x.c is shifted by -x.q so that the integer part of the result is correct and then it is shifted further by DIGITS to also cover some digits from the fractional part.
    // q = -DIGITS # we only need to remember that there are these digits after the decimal point
    // normalize(Decimal(x.s, c, q))

	return out
}