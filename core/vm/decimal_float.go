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

func copyDecimal(a *decimal) (*decimal) {
	return &decimal{a.s, a.e}
}

var ZERO_uint256_Int = uint256.NewInt(0)
var ONE_uint256_Int = uint256.NewInt(1)
var TEN_uint256_Int = uint256.NewInt(10)

// c.s*10^c.e = a.s*10^a.e + b.s*10^b.e
func add(a, b, out *decimal, L bool) (*decimal) {
	if L {fmt.Println("add", "a", "b", a, b)}

	aqmbq := new(uint256.Int).Sub(&a.e, &b.e)
	if L {fmt.Println("add", "aqmbq", aqmbq)}
	
	aqmbq_abs := new(uint256.Int).Abs(aqmbq)
	if L {fmt.Println("add", "aqmbq_abs", aqmbq_abs)}

	ten_power := *TEN_uint256_Int
	ten_power.Exp(&ten_power, aqmbq_abs) // todo faster way should exist
	if L {fmt.Println("add", "ten_power", ten_power, ten_power.String())}

	ca := new(uint256.Int).Abs(&a.s)
	if L {fmt.Println("add", "ca", ca, ca.String())}

	cb := new(uint256.Int).Abs(&b.s)
	if L {fmt.Println("add", "cb", cb, cb.String())}

	aqmbq_neg := aqmbq.Sign() == -1
	if !aqmbq_neg {
		ca.Mul(ca, &ten_power)
	} else if aqmbq.Cmp(ZERO_uint256_Int) != 0 {
		cb.Mul(cb, &ten_power)
	}
	if L {fmt.Println("add", "ca", ca, ca.String())}
	if L {fmt.Println("add", "cb", cb, cb.String())}

	c := new(uint256.Int).Add(ca, cb)
	if L {fmt.Println("add", "c", c, c.String())}
	
	// min(a.e, b.e)
	ae_neg := a.e.Sign() == -1
	if L {fmt.Println("add", "b.e", b.e)}
	be_neg := b.e.Sign() == -1
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
func subtract(a, b, out *decimal, L bool) (*decimal) {
	negate(b, out, L)
	add(a, out, out, L)
	return out
}

// a * b
func multiply(a, b, out *decimal, L bool) (*decimal) {
	if L {fmt.Println("multiply", "a", a, "b", b)}
	// if L {fmt.Println("multiply", "a", String(a), "b", String(b), "precision", precision)}
	// if L {fmt.Println("multiply", "a", a, "b", b)}
	out.s.Mul(&a.s, &b.s)
	// if L {fmt.Println("multiply", "out.c", out.c)}
	out.e.Add(&a.e, &b.e)
	if L {fmt.Println("multiply", "out", out)}
	// return normalize(copy(out), out, precision, false, L)
	return out
}

func signed_div(numerator, denominator, out *uint256.Int) (*uint256.Int) {
	sn := numerator.Sign()
	sd := denominator.Sign()
	if sn == 0 && sd == 0 {
		out = nil
		return nil
	}
	if sn == 0 { 
		out = uint256.NewInt(0)
		return out
	}

	n := *numerator
	if sn == -1 {
		n.Neg(numerator)
	}

	d := *denominator
	if sd == -1 {
		d.Neg(denominator)
	}

	out.Div(&n, &d)

	if (sn == -1) != (sd == -1) {
		out.Neg(out)
	}

	return out
}

// 1 / a
func inverse(a, out *decimal, L bool) (*decimal) {
	if L {fmt.Println("inverse", "a", a, a.s.Sign(), a.s.Dec(), a.e.Sign(), a.e.Dec())}

	// out.n = a.n
	// if L {fmt.Println("inverse", "out.n", out.n)}
	
	max_precision := uint256.NewInt(76)
	var precision uint256.Int
	precision.Add(max_precision, &a.e) // more than max decimal precision on 256 bits
	if L {fmt.Println("inverse", "precision", precision, precision.Dec())}

	ten_power := *TEN_uint256_Int
	ae_m_precision := new(uint256.Int).Neg(&a.e)
	ae_m_precision.Add(ae_m_precision, &precision)
	if L {fmt.Println("inverse", "ae_m_precision", ae_m_precision, ae_m_precision.Dec())}
	if ae_m_precision.Cmp(ZERO_uint256_Int) == -1 {
		fmt.Println("ae_m_precision NEGATIVE")
		return nil
	}
	ten_power.Exp(&ten_power, ae_m_precision)
	if L {fmt.Println("inverse", "ten_power", ten_power, ten_power.Dec())}
	// out.s.Div(&ten_power, &a.s)
	signed_div(&ten_power, &a.s, &out.s)
	if L {fmt.Println("inverse after div", "out.s", out.s, out.s.Dec())}

	out.e.Sub(ZERO_uint256_Int, &precision)
	if L {fmt.Println("inverse after sub", "out.e", out.e, out.e.Dec())}
	
	// out.s.Sub(&out.s, &precision)

	// if L {fmt.Println("inverse", "out", out, String(out))}
	
	// norm := normalize(copy(out), out, precision, false, L)
	// if L {fmt.Println("inverse", "norm", norm, String(norm))}
	// return norm
	
	// c = round(BigInt(10)^(-x.q + DIGITS) / x.c) # the decimal point of 1/x.c is shifted by -x.q so that the integer part of the result is correct and then it is shifted further by DIGITS to also cover some digits from the fractional part.
    // q = -DIGITS # we only need to remember that there are these digits after the decimal point
    // normalize(Decimal(x.s, c, q))

	return out
}

// a / b
func divide(a, b, out *decimal, L bool) (*decimal) {
	if L {fmt.Println("divide", "a", a, "b", b)}

	inverse(b, out, L)
	multiply(a, copyDecimal(out), out, L)
	return out
}

func iszero(a *decimal) (bool) {
	return a.s.IsZero()
}

// e^a
// total decimal precision is where a^(taylor_steps+1)/(taylor_steps+1)! == 10^(-target_decimal_precision)
func exp(a, out *decimal, taylor_steps uint, L bool) (*decimal) {

	if L {fmt.Println("a", a, "taylor_precision", taylor_steps)}

	if iszero(a) {
		out.s = *ONE_uint256_Int // possible problem
		out.e = *ZERO_uint256_Int
		return out
	}

	ONE := decimal{*ONE_uint256_Int, *ZERO_uint256_Int} // 1
	a_power := decimal{*ONE_uint256_Int, *ZERO_uint256_Int} // 1
	factorial := decimal{*ONE_uint256_Int, *ZERO_uint256_Int} // 1
	factorial_next := decimal{*ZERO_uint256_Int, *ZERO_uint256_Int} // 0
	factorial_inv := decimal{*ONE_uint256_Int, *ZERO_uint256_Int} // 1
	
	// out = 1
	out.s = *ONE_uint256_Int
	out.e = *ZERO_uint256_Int

	if L {fmt.Println("out", out)}

	for i := uint(0) ; i < taylor_steps ; i++ {
		if L {fmt.Println("i", i)}

		if L {fmt.Println("a", a)}
		if L {fmt.Println("a_power", a_power)}
		multiply(copyDecimal(&a_power), a, &a_power, false) // a^i
		if L {fmt.Println("a_power", a_power)}

		if L {fmt.Println("ONE", ONE_uint256_Int)}
		if L {fmt.Println("factorial_next", factorial_next)}
		add(copyDecimal(&factorial_next), &ONE, &factorial_next, false) // i + 1
		if L {fmt.Println("factorial_next", factorial_next)}
		
		if L {fmt.Println("factorial", factorial)}
		multiply(copyDecimal(&factorial), &factorial_next, &factorial, false) // i!
		if L {fmt.Println("factorial", factorial)}
		
		if L {fmt.Println("factorial_inv", factorial_inv)}
		inverse(&factorial, &factorial_inv, false) // 1 / i!
		if L {fmt.Println("factorial_inv", factorial_inv)}

		multiply(&a_power, copyDecimal(&factorial_inv), &factorial_inv, false) // store in factorial_inv as not needed anymore
		if L {fmt.Println("factorial_inv", factorial_inv)}

		if L {fmt.Println("out", out)}
		add(copyDecimal(out), &factorial_inv, out, false)
		if L {fmt.Println("out", out)}
	}

	if L {fmt.Println("out", out)}

	return out
}

func round(a, out *decimal, precision uint64, normal bool, L bool) (*decimal) {

	var shift uint256.Int
	shift.Add(uint256.NewInt(precision), &a.e)

	out.s = a.s
	out.e = a.e

	if shift.Cmp(ZERO_uint256_Int) == 1 || shift.Cmp(&a.e) == -1 {
		if normal {
			return out
		}
		return normalize(out, out, precision, true, L)
	}

	shift.Neg(&shift) // shift *= -1
	ten_power := *TEN_uint256_Int
	ten_power.Exp(&ten_power, &shift) // 10^shift // TODO if shift<0, problem
	// out.s.Div(&out.s, &ten_power)
	signed_div(&out.s, &ten_power, &out.s)
	out.e.Add(&out.e, &shift)

	if normal {
		return out
	}

	return normalize(copyDecimal(out), out, precision, true, L)
}

func normalize(a, out *decimal, precision uint64, rounded bool, L bool) (*decimal) {
	if L {fmt.Println("normalize", "a", a)}

	// remove trailing zeros in significand
	p := *ZERO_uint256_Int
	ten_power := *TEN_uint256_Int // 10^(p+1)
	if a.s.Cmp(ZERO_uint256_Int) != 0 { // if a.c != 0
		for {
			t := uint256.NewInt(0)
			tt := t.Mod(&a.s, &ten_power)
			ttt := tt.Cmp(ZERO_uint256_Int)
			if ttt != 0 { // if a.s % 10^(p+1) != 0
				break
			}
			p.Add(&p, ONE_uint256_Int)
			ten_power.Mul(&ten_power, TEN_uint256_Int) // 10^(p+1)
		}
	}
	// ten_power.Div(&ten_power, TEN_uint256_Int) // 10^p
	signed_div(&ten_power, TEN_uint256_Int, &ten_power) // 10^p
	if L {fmt.Println("normalize", "p", p.Dec())}
	if L {fmt.Println("normalize", "ten_power", ten_power.Dec(), ten_power.Sign())}
	// out.s.Div(&a.s, &ten_power) // out.s = a.s / 10^p
	signed_div(&a.s, &ten_power, &out.s) // out.s = a.s / 10^p
	if L {fmt.Println("normalize", "out.s", out.s.Dec(), out.s.Sign())}
	out.s.Abs(&out.s) // out.s = abs(out.s)
	if L {fmt.Println("normalize", "out.s", out.s.Dec(), out.s.Sign())}

	out.e = *ZERO_uint256_Int
	a_pos := a.s.Sign() == 1
	if !(out.s.Cmp(ZERO_uint256_Int) == 0 && a_pos) { // if out.c == 0
		out.e.Add(&a.e, &p)
	}
	if L {fmt.Println("normalize", "out.e", out.e.Dec(), out.e.Sign())}
	
	if rounded {
		return out
	}

	return round(copyDecimal(out), out, precision, true, L)
}