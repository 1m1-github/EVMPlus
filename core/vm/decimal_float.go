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

func showDecimal(a *decimal) (string) {
	return fmt.Sprintf("%v %v", showInt(&a.s), showInt(&a.e))
}
func showInt(a *uint256.Int) (string) {
	return fmt.Sprintf("%v(%v)", a.Sign(), a.Dec())
}

func copyDecimal(a *decimal) (*decimal) {
	return &decimal{a.s, a.e}
}

var ZERO_uint256_Int = uint256.NewInt(0)
var ONE_uint256_Int = uint256.NewInt(1)
var TEN_uint256_Int = uint256.NewInt(10)
var MINUS_ONE_uint256_Int = new(uint256.Int).Neg(uint256.NewInt(1))

var ZERO = decimal{*ZERO_uint256_Int, *ZERO_uint256_Int}
var ONE = decimal{*ONE_uint256_Int, *ZERO_uint256_Int}

func add_helper(a, b *decimal) (*uint256.Int) {
	exponent_diff := new(uint256.Int).Sub(&a.e, &b.e)
	if exponent_diff.Sign() == -1 {
		exponent_diff = uint256.NewInt(0)
	}
	
	ten_power := *TEN_uint256_Int
	ten_power.Exp(&ten_power, exponent_diff)

	var ca uint256.Int
	ca.Mul(&a.s, &ten_power)
	return &ca
}

// c.s*10^c.e = a.s*10^a.e + b.s*10^b.e
func add(a, b, out *decimal, L bool) (*decimal) {
	if L {fmt.Println("add", "a", "b", showDecimal(a), showDecimal(b))}

	ca := add_helper(a, b)
	cb := add_helper(b, a)

	c := new(uint256.Int).Add(ca, cb)
	if L {fmt.Println("add", "c", showInt(c))}

	out.s.Abs(c)
	out.e = *signed_min(&a.e, &b.e, false)
	if L {fmt.Println("add", "out", showDecimal(out))}

	return out
}

// -a
func negate(a, out *decimal, L bool) (*decimal) {
	if L {fmt.Println("negate", showDecimal(a))}
	out.s.Neg(&a.s)
	out.e = a.e
	if L {fmt.Println("negate", showDecimal(out))}
	return out
}

// a - b
func subtract(a, b, out *decimal, L bool) (*decimal) {
	if L {fmt.Println("subtract", showDecimal(a), showDecimal(b))}
	negate(b, out, false)
	if L {fmt.Println("subtract 2", showDecimal(out))}
	add(a, out, out, true)
	if L {fmt.Println("subtract 3", showDecimal(out))}
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

func signed_min(a, b *uint256.Int, L bool) (*uint256.Int) {
	a_neg := a.Sign() == -1
	b_neg := b.Sign() == -1
	
	if a_neg && !b_neg {
		return a
	} else if b_neg && !a_neg {
		return b
	} else if !a_neg && !b_neg {
		if a.Lt(b) {
			return a
		} else {
			return b
		}
	} else { // both negative
		if a.Lt(b) {
			return b
		} else {
			return a
		}
	}
}

func signed_div(numerator, denominator, out *uint256.Int, L bool) (*uint256.Int) {
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
	signed_div(&ten_power, &a.s, &out.s, false)
	if L {fmt.Println("inverse after div", "out.s", out.s, out.s.Dec())}

	out.e.Sub(ZERO_uint256_Int, &precision)
	if L {fmt.Println("inverse after sub", "out.e", out.e, out.e.Dec())}

	return out
}

// a / b
func divide(a, b, out *decimal, L bool) (*decimal) {
	if L {fmt.Println("divide", "a", a, "b", b)}

	inverse(b, out, L)
	multiply(a, copyDecimal(out), out, L)
	return out
}

func iszero(a *decimal, L bool) (bool) {
	return a.s.IsZero()
}

// a should be normalized
func isone(a *decimal, L bool) (bool) {
	return a.s.Eq(ONE_uint256_Int) && a.e.Eq(ZERO_uint256_Int)
}

// a < b
func lessthan(a, b *decimal, L bool) (bool) {

	if iszero(a, false) && iszero(b, false) {
		return false
	}

	if a.e.Eq(&b.e) {
		var out uint256.Int
		return out.Abs(&a.s).Lt(out.Abs(&b.s))
	}

	var diff decimal
	subtract(b, a, &diff, false)

	farther_from_0 := diff.s.Gt(ZERO_uint256_Int) || (iszero(&diff, false) && diff.e.Gt(ZERO_uint256_Int))

    if diff.s.Sign() >= 1 {
		return !farther_from_0
	} else {
        return farther_from_0
	}
}

// a == b
// a,b should be both normalized
func equal(a, b *decimal) (bool) {
	return a.s.Eq(&b.s) && a.e.Eq(&b.e)
}

// e^a
// total decimal precision is where a^(taylor_steps+1)/(taylor_steps+1)! == 10^(-target_decimal_precision)
func exp(a, out *decimal, taylor_steps uint, L bool) (*decimal) {

	if L {fmt.Println("a", a, "taylor_precision", taylor_steps)}

	if iszero(a, false) {
		out.s = *ONE_uint256_Int
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


// http://www.claysturner.com/dsp/BinaryLogarithm.pdf
// 0 < a
// func log2(a, out *decimal, precision uint64, L bool) (*decimal) {

// 	b := copyDecimal(&ZERO)

// 	var a_norm decimal
// 	normalize(a, &a_norm, 0, false, false)

// 	if a_vs_zero := a_norm.s.Cmp(ZERO_uint256_Int); a_vs_zero <= 0 {
// 		out = nil
// 		return out
// 	}

// 	if isone(&a_norm) {
// 		return b
// 	}

// 	// double a until 1 <= a
// 	for {

// 		if a_vs_one := a.Cmp(ONE); a_vs_one != -1 {
// 			break
// 		}

// 		a.Num().Lsh(a.Num(), 1) // double
// 		b.Add(b, MINUS_ONE)
// 	}
// 	if L {
// 		fmt.Println("log2 doubled", a.FloatString(10), b.FloatString(10))
// 	}

// 	// half a until a < 2
// 	for {

// 		if a_vs_two := a.Cmp(TWO); a_vs_two == -1 {
// 			break
// 		}

// 		a.Denom().Lsh(a.Denom(), 1) // half
// 		b.Add(b, ONE)
// 	}
// 	if L {
// 		fmt.Println("log2 halved", a.FloatString(10), b.FloatString(10))
// 	}

// 	// from here: 1 <= a < 2 <=> 0 <= b < 1

// 	// compare a^2 to 2 to reveal b bit-by-bit
// 	precision_counter := 0 // for now, precision is naiive
// 	v := big.NewRat(1, 2)
// 	for {
// 		if precision == precision_counter {
// 			break
// 		}

// 		if L {
// 			fmt.Println("log2 precision_counter", precision_counter)
// 			fmt.Println("log2 v", v.FloatString(10))
// 			fmt.Println("log2 a", a.FloatString(10))
// 			fmt.Println("log2 b", b.FloatString(10))
// 		}

// 		a.Mul(a, a) // THIS IS SLOW
// 		// a = big.NewRat(a.Num().Int64()*a.Num().Int64(), a.Denom().Int64()*a.Denom().Int64())

// 		if L {
// 			fmt.Println("log2 a^2", a.FloatString(10))
// 		}

// 		if a2_vs_two := a.Cmp(TWO); a2_vs_two != -1 {

// 			if L {
// 				fmt.Println("log2 2 <= a^2", a.FloatString(10))
// 			}

// 			a.Denom().Lsh(a.Denom(), 1) // half
// 			b.Add(b, v)
// 		} else {
// 			if L {
// 				fmt.Println("log2 a^2 < 2")
// 			}
// 		}

// 		v.Denom().Lsh(v.Denom(), 1) // half

// 		precision_counter++
// 	}

// 	if L {
// 		fmt.Println("log2 b", b.FloatString((10)))
// 	}

// 	return b
// }

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
	signed_div(&out.s, &ten_power, &out.s, false)
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

	signed_div(&ten_power, TEN_uint256_Int, &ten_power, false) // 10^p
	if L {fmt.Println("normalize", "p", p.Dec())}
	if L {fmt.Println("normalize", "ten_power", ten_power.Dec(), ten_power.Sign())}
	signed_div(&a.s, &ten_power, &out.s, false) // out.s = a.s / 10^p
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