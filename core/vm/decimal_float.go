package vm

import (
	// "fmt" // TODO remove
	"fmt"
	"math/big"

	"github.com/holiman/uint256"
)

// c * 10^q
type Decimal struct {
	c big.Int // coefficient
	q big.Int // exponent
}

// TODO normalize first
func (d2 *Decimal) Eq(d1 *Decimal) bool {
	return d1.c.Cmp(&d2.c) == 0 && d1.q.Cmp(&d2.q) == 0
}

func UInt256IntToBigInt(x *uint256.Int) (y *big.Int) {
	if x.Sign() == -1 {
		var xn uint256.Int
		y = xn.Neg(x).ToBig()
		y.Neg(y)
	} else {
		y = x.ToBig()
	}
	return y
}
func UInt256IntTupleToDecimal(_c, _q *uint256.Int) *Decimal {
	c := UInt256IntToBigInt(_c)
	q := UInt256IntToBigInt(_q)
	return &Decimal{*c, *q}
}

func BigIntToUInt256Int(x *big.Int) (y *uint256.Int) {
	y, overflow := uint256.FromBig(x)
	if overflow { // x more than 256 bits
		panic("overflow")
	}
	if y.Sign() != x.Sign() {
		panic("overflow")
	}
	return y
}
func (d *Decimal) DecimalToUInt256IntTuple() (c, q *uint256.Int) {
	c = BigIntToUInt256Int(&d.c)
	q = BigIntToUInt256Int(&d.q)
	return c, q
}

func String(d *Decimal) string {
	sc := '+'
	if d.c.Sign() == -1 {
		sc = '-'
	}
	sq := '+'
	if d.q.Sign() == -1 {
		sq = '-'
	}
	return fmt.Sprintf("%c%v*10^%c%v", sc, d.c.String(), sq, d.q.String())
}

func (d *Decimal) copyDecimal() *Decimal {
	return &Decimal{d.c, d.q}
}

// TODO all needed?
var MINUS_ONE_BIG = big.NewInt(-1)
var ZERO_BIG = big.NewInt(0)
var ONE_BIG = big.NewInt(1)
var TEN_BIG = big.NewInt(10)

var ZERO = Decimal{*ZERO_BIG, *ZERO_BIG}
var ONE = Decimal{*ONE_BIG, *ZERO_BIG}

func min(a, b *big.Int) (c *big.Int) {
	if a.Cmp(b) == -1 {
		return a
	} else {
		return b
	}
}

func add_helper(d1, d2 *Decimal) (c big.Int) {
	var exponent_diff big.Int
	exponent_diff.Sub(&d1.q, &d2.q)
	if exponent_diff.Sign() == -1 {
		exponent_diff = *ZERO_BIG
	}

	c.Exp(TEN_BIG, &exponent_diff, ZERO_BIG)
	c.Mul(&d1.c, &c)

	return c
}

// a + b
func (out *Decimal) Add(d1, d2 *Decimal) *Decimal {
	ca := add_helper(d1, d2)
	cb := add_helper(d2, d1)

	out.c.Add(&ca, &cb)
	out.q = *min(&d1.q, &d2.q)

	return out
}

// -a
func (out *Decimal) Negate(a *Decimal) *Decimal {
	out.c.Neg(&a.c)
	out.q = a.q
	return out
}

// a - b
func (out *Decimal) Subtract(a, b *Decimal) *Decimal {
	out.Negate(b)
	out.Add(a, out)
	return out
}

// a * b
func (out *Decimal) Multiply(a, b *Decimal) *Decimal {
	out.c.Mul(&a.c, &b.c)
	out.q.Add(&a.q, &b.q)
	// return normalize(copy(out), out, precision, false, L)
	return out
}

// 1 / a
func (out *Decimal) Inverse(a *Decimal) *Decimal {
	max_precision := big.NewInt(50) // TODO choose correct max_precision
	var precision big.Int
	precision.Add(max_precision, &a.q) // more than max decimal precision on 256 bits

	var aq_m_precision big.Int
	aq_m_precision.Sub(&precision, &a.q)
	if aq_m_precision.Cmp(ZERO_BIG) == -1 {
		panic("ae_m_precision NEGATIVE")
	}

	out.c.Exp(TEN_BIG, &aq_m_precision, ZERO_BIG)
	out.c.Div(&out.c, &a.c)

	out.q.Neg(&precision)

	return out
}

// // a / b
// func divide(a, b, out *Decimal, L bool) *Decimal {
// 	if L {
// 		fmt.Println("divide", "a", a, "b", b)
// 	}

// 	inverse(b, out, L)
// 	multiply(a, copyDecimal(out), out, L)
// 	return out
// }

// func iszero(a *Decimal, L bool) bool {
// 	return a.c.IsZero()
// }

// // a should be normalized
// func isone(a *Decimal, L bool) bool {
// 	return a.c.Eq(ONE_uint256_Int) && a.q.Eq(ZERO_uint256_Int)
// }

// // a < b
// func lessthan(a, b *Decimal, L bool) bool {
// 	if L {
// 		fmt.Println("lessthan", String(a), String(b))
// 	}
// 	var diff Decimal
// 	subtract(a, b, &diff, false)
// 	if L {
// 		fmt.Println("lessthan diff", String(&diff))
// 	}
// 	return diff.c.Sign() == -1
// }

// // a == b
// // a,b should be both normalized
// func equal(a, b *Decimal) bool {
// 	return a.c.Eq(&b.c) && a.q.Eq(&b.q)
// }

// // e^a
// // total decimal precision is where a^(taylor_steps+1)/(taylor_steps+1)! == 10^(-target_decimal_precision)
// func exp(a, out *Decimal, taylor_steps uint, L bool) *Decimal {

// 	if L {
// 		fmt.Println("a", a, "taylor_precision", taylor_steps)
// 	}

// 	if iszero(a, false) {
// 		out.c = *ONE_uint256_Int
// 		out.q = *ZERO_uint256_Int
// 		return out
// 	}

// 	ONE := Decimal{*ONE_uint256_Int, *ZERO_uint256_Int}             // 1
// 	a_power := Decimal{*ONE_uint256_Int, *ZERO_uint256_Int}         // 1
// 	factorial := Decimal{*ONE_uint256_Int, *ZERO_uint256_Int}       // 1
// 	factorial_next := Decimal{*ZERO_uint256_Int, *ZERO_uint256_Int} // 0
// 	factorial_inv := Decimal{*ONE_uint256_Int, *ZERO_uint256_Int}   // 1

// 	// out = 1
// 	out.c = *ONE_uint256_Int
// 	out.q = *ZERO_uint256_Int

// 	if L {
// 		fmt.Println("out", out)
// 	}

// 	for i := uint(0); i < taylor_steps; i++ {
// 		if L {
// 			fmt.Println("i", i)
// 		}

// 		if L {
// 			fmt.Println("a", a)
// 		}
// 		if L {
// 			fmt.Println("a_power", a_power)
// 		}
// 		multiply(copyDecimal(&a_power), a, &a_power, false) // a^i
// 		if L {
// 			fmt.Println("a_power", a_power)
// 		}

// 		if L {
// 			fmt.Println("ONE", ONE_uint256_Int)
// 		}
// 		if L {
// 			fmt.Println("factorial_next", factorial_next)
// 		}
// 		add(copyDecimal(&factorial_next), &ONE, &factorial_next, false) // i + 1
// 		if L {
// 			fmt.Println("factorial_next", factorial_next)
// 		}

// 		if L {
// 			fmt.Println("factorial", factorial)
// 		}
// 		multiply(copyDecimal(&factorial), &factorial_next, &factorial, false) // i!
// 		if L {
// 			fmt.Println("factorial", factorial)
// 		}

// 		if L {
// 			fmt.Println("factorial_inv", factorial_inv)
// 		}
// 		inverse(&factorial, &factorial_inv, false) // 1 / i!
// 		if L {
// 			fmt.Println("factorial_inv", factorial_inv)
// 		}

// 		multiply(&a_power, copyDecimal(&factorial_inv), &factorial_inv, false) // store in factorial_inv as not needed anymore
// 		if L {
// 			fmt.Println("factorial_inv", factorial_inv)
// 		}

// 		if L {
// 			fmt.Println("out", out)
// 		}
// 		add(copyDecimal(out), &factorial_inv, out, false)
// 		if L {
// 			fmt.Println("out", out)
// 		}
// 	}

// 	if L {
// 		fmt.Println("out", out)
// 	}

// 	return out
// }

// // http://www.claysturner.com/dsp/BinaryLogarithm.pdf
// // 0 < a
// // func log2(a, out *decimal, precision uint64, L bool) (*decimal) {

// // 	b := copyDecimal(&ZERO)

// // 	var a_norm decimal
// // 	normalize(a, &a_norm, 0, false, false)

// // 	if a_vs_zero := a_norm.s.Cmp(ZERO_uint256_Int); a_vs_zero <= 0 {
// // 		out = nil
// // 		return out
// // 	}

// // 	if isone(&a_norm) {
// // 		return b
// // 	}

// // 	// double a until 1 <= a
// // 	for {

// // 		if a_vs_one := a.Cmp(ONE); a_vs_one != -1 {
// // 			break
// // 		}

// // 		a.Num().Lsh(a.Num(), 1) // double
// // 		b.Add(b, MINUS_ONE)
// // 	}
// // 	if L {
// // 		fmt.Println("log2 doubled", a.FloatString(10), b.FloatString(10))
// // 	}

// // 	// half a until a < 2
// // 	for {

// // 		if a_vs_two := a.Cmp(TWO); a_vs_two == -1 {
// // 			break
// // 		}

// // 		a.Denom().Lsh(a.Denom(), 1) // half
// // 		b.Add(b, ONE)
// // 	}
// // 	if L {
// // 		fmt.Println("log2 halved", a.FloatString(10), b.FloatString(10))
// // 	}

// // 	// from here: 1 <= a < 2 <=> 0 <= b < 1

// // 	// compare a^2 to 2 to reveal b bit-by-bit
// // 	precision_counter := 0 // for now, precision is naiive
// // 	v := big.NewRat(1, 2)
// // 	for {
// // 		if precision == precision_counter {
// // 			break
// // 		}

// // 		if L {
// // 			fmt.Println("log2 precision_counter", precision_counter)
// // 			fmt.Println("log2 v", v.FloatString(10))
// // 			fmt.Println("log2 a", a.FloatString(10))
// // 			fmt.Println("log2 b", b.FloatString(10))
// // 		}

// // 		a.Mul(a, a) // THIS IS SLOW
// // 		// a = big.NewRat(a.Num().Int64()*a.Num().Int64(), a.Denom().Int64()*a.Denom().Int64())

// // 		if L {
// // 			fmt.Println("log2 a^2", a.FloatString(10))
// // 		}

// // 		if a2_vs_two := a.Cmp(TWO); a2_vs_two != -1 {

// // 			if L {
// // 				fmt.Println("log2 2 <= a^2", a.FloatString(10))
// // 			}

// // 			a.Denom().Lsh(a.Denom(), 1) // half
// // 			b.Add(b, v)
// // 		} else {
// // 			if L {
// // 				fmt.Println("log2 a^2 < 2")
// // 			}
// // 		}

// // 		v.Denom().Lsh(v.Denom(), 1) // half

// // 		precision_counter++
// // 	}

// // 	if L {
// // 		fmt.Println("log2 b", b.FloatString((10)))
// // 	}

// // 	return b
// // }

// // func round(a, out *decimal, precision uint64, normal bool, L bool) *decimal {

// // 	var shift uint256.Int
// // 	shift.Add(uint256.NewInt(precision), &a.q)

// // 	out.c = a.c
// // 	out.q = a.q

// // 	if shift.Cmp(ZERO_uint256_Int) == 1 || shift.Cmp(&a.q) == -1 {
// // 		if normal {
// // 			return out
// // 		}
// // 		return normalize(out, out, precision, true, L)
// // 	}

// // 	if L {fmt.Println(showInt(&shift))}
// // 	shift.Neg(&shift) // shift *= -1
// // 	if L {fmt.Println(showInt(&shift))}
// // 	var ten_power uint256.Int
// // 	ten_power.Exp(TEN_uint256_Int, &shift) // 10^shift // TODO if shift<0, problem
// // 	// out.s.Div(&out.s, &ten_power)
// // 	signed_div(&out.c, &ten_power, &out.c, false)

// // 	out.q.Add(&out.q, &shift)

// // 	if normal {
// // 		return out
// // 	}

// // 	return normalize(copyDecimal(out), out, precision, true, L)
// // }

func find_num_trailing_zeros_signed(a *big.Int) (p, ten_power *big.Int) {
	b := *a
	if b.Sign() == -1 {
		b.Neg(&b)
	}

	p = big.NewInt(0)
	ten_power = big.NewInt(10)
	if b.Cmp(ZERO_BIG) != 0 { // if b != 0
		for {
			var m big.Int
			m.Mod(&b, ten_power)
			if m.Cmp(ZERO_BIG) != 0 { // if b % 10^(p+1) != 0
				break
			}
			p.Add(p, ONE_BIG)
			ten_power.Mul(ten_power, TEN_BIG) // 10^(p+1)
		}
	}
	ten_power.Div(ten_power, TEN_BIG)

	return p, ten_power
}

func (out *Decimal) Normalize(a *Decimal, precision uint64, rounded bool) *Decimal {
	// remove trailing zeros in significand
	p, ten_power := find_num_trailing_zeros_signed(&a.c)
	out.c.Div(&a.c, ten_power)

	out.q = *ZERO_BIG
	a_neg := a.c.Sign() == -1
	if !(out.c.Cmp(ZERO_BIG) == 0 && !a_neg) { // if out.c == 0
		out.q.Add(&a.q, p)
	}

	// if rounded {
	return out
	// }

	// return round(copyDecimal(out), out, precision, true, L)
}
