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
	return createDecimal(c, q)
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
func (d *Decimal) SetUInt256IntTupleFromDecimal(c, q *uint256.Int) {
	c.Set(BigIntToUInt256Int(&d.c))
	q.Set(BigIntToUInt256Int(&d.q))
	// return c, q
}

func (d *Decimal) String() string {
	return fmt.Sprintf("%v*10^%v", d.c.String(), d.q.String())
}

func copyDecimal(d *Decimal) *Decimal {
	return createDecimal(&d.c, &d.q)
}
func createDecimal(_c, _q *big.Int) *Decimal {
	var c, q big.Int
	c.Set(_c)
	q.Set(_q)
	return &Decimal{c, q}
}

// TODO all needed?
var MINUS_ONE_BIG = big.NewInt(-1)
var ZERO_BIG = big.NewInt(0)
var ONE_BIG = big.NewInt(1)
var TWO_BIG = big.NewInt(2)
var TEN_BIG = big.NewInt(10)

var ZERO = createDecimal(ZERO_BIG, ONE_BIG)
var HALF = createDecimal(big.NewInt(5), MINUS_ONE_BIG)
var ONE = createDecimal(ONE_BIG, ZERO_BIG)
var TWO = createDecimal(TWO_BIG, ZERO_BIG)
var MINUS_ONE = createDecimal(MINUS_ONE_BIG, ZERO_BIG)

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
		exponent_diff = *ZERO_BIG // shallow copy ok
	}

	c.Exp(TEN_BIG, &exponent_diff, ZERO_BIG)
	c.Mul(&d1.c, &c)

	return c
}

// a + b
func (out *Decimal) Add(a, b *Decimal) *Decimal {
	ca := add_helper(a, b)
	cb := add_helper(b, a)

	out.c.Add(&ca, &cb)
	out.q.Set(min(&a.q, &b.q))

	return out
}

// -a
func (out *Decimal) Negate(a *Decimal) *Decimal {
	out.c.Neg(&a.c)
	out.q.Set(&a.q)
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
	// normalize?
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

	aq_m_precision.Exp(TEN_BIG, &aq_m_precision, ZERO_BIG) // aq_m_precision not needed after
	out.c.Div(&aq_m_precision, &a.c)
	out.q.Neg(&precision)

	// normalize?

	return out
}

// a / b
func (out *Decimal) Divide(a, b *Decimal) *Decimal {
	out.Inverse(b)
	out.Multiply(a, out)
	return out
}

func (a *Decimal) IsZero() bool {
	return a.c.Cmp(ZERO_BIG) == 0
}

// a should be normalized
func (a *Decimal) IsOne() bool {
	return a.c.Cmp(ONE_BIG) == 0 && a.q.Cmp(ZERO_BIG) == 0
}

func (a *Decimal) IsNegative() bool {
	return a.c.Sign() == -1
}

// a < b
func (a *Decimal) LessThan(b *Decimal) bool {
	var diff Decimal
	diff.Subtract(a, b)
	return diff.c.Sign() == -1
}

// e^a
// total decimal precision is where a^(taylor_steps+1)/(taylor_steps+1)! == 10^(-target_decimal_precision)
func (out *Decimal) Exp(a *Decimal, taylor_steps uint) *Decimal {
	// out = 1
	out.c.Set(ONE_BIG)
	out.q.Set(ZERO_BIG)

	if a.IsZero() {
		return out
	}

	var factorial_inv Decimal
	a_power := copyDecimal(ONE)
	factorial := copyDecimal(ONE)
	factorial_next := copyDecimal(ZERO)

	for i := uint(1); i <= taylor_steps; i++ { // step 0 skipped as a set to 1
		a_power.Multiply(a_power, a)                    // a^i
		factorial_next.Add(factorial_next, ONE)         // i++
		factorial.Multiply(factorial, factorial_next)   // i!
		factorial_inv.Inverse(factorial)                // 1/i!
		factorial_inv.Multiply(&factorial_inv, a_power) // store a^i/i! in factorial_inv as not needed anymore
		out.Add(out, &factorial_inv)                    // out += a^i/i!
	}

	return out
}

func (out *Decimal) double() {
	out.c.Lsh(&out.c, 1)
}
func (out *Decimal) halve() {
	out.Multiply(out, HALF)
}

// http://www.claysturner.com/dsp/BinaryLogarithm.pdf
// 0 < a
func (out *Decimal) Log2(a *Decimal, precision uint) *Decimal {
	if a.c.Sign() != 1 {
		panic("Log2 needs 0 < a")
	}

	var a_norm Decimal
	a_norm.Normalize(a, 0, false)

	// out = 0
	out.c.Set(ZERO_BIG)
	out.q.Set(ONE_BIG)

	if a_norm.IsOne() {
		return out
	}

	// double a until 1 <= a
	for {
		if !a.LessThan(ONE) {
			break
		}

		a.double()              // a *= 2
		out.Add(out, MINUS_ONE) // out--
	}

	// half a until a < 2
	for {
		if a.LessThan(TWO) {
			break
		}

		a.halve()         // a /= 2
		out.Add(out, ONE) // out++
	}

	// from here: 1 <= a < 2 <=> 0 <= out < 1

	// compare a^2 to 2 to reveal out bit-by-bit
	precision_counter := uint(0) // for now, precision is naiive
	v := copyDecimal(HALF)
	for {
		if precision == precision_counter {
			break
		}

		a.Multiply(a, a) // THIS IS SLOW

		if !a.LessThan(TWO) {
			a.halve() // a /= 2
			out.Add(out, v)
		}

		v.halve()

		precision_counter++
	}

	return out
}

// // func round(a, out *decimal, precision uint, normal bool, L bool) *decimal {

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
	var b big.Int
	b.Set(a)
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

func (out *Decimal) Normalize(a *Decimal, precision uint, rounded bool) *Decimal {
	// remove trailing zeros in significand
	p, ten_power := find_num_trailing_zeros_signed(&a.c)
	out.c.Div(&a.c, ten_power)

	a_neg := a.IsNegative()
	if out.c.Cmp(ZERO_BIG) != 0 || a_neg {
		out.q.Add(&a.q, p)
	} else {
		out.q.Set(ZERO_BIG)
	}

	// if rounded {
	return out
	// }

	// return round(copyDecimal(out), out, precision, true, L)
}

// sin(a)
// TODO
func (out *Decimal) Sin(a *Decimal, taylor_steps uint) *Decimal {
	// out = 1
	out.c.Set(ONE_BIG)
	out.q.Set(ZERO_BIG)

	if a.IsZero() {
		return out
	}

	var factorial_inv Decimal
	a_power := copyDecimal(ONE)
	factorial := copyDecimal(ONE)
	factorial_next := copyDecimal(ZERO)

	for i := uint(1); i <= taylor_steps; i++ { // step 0 skipped as a set to 1
		a_power.Multiply(a_power, a)                    // a^i
		factorial_next.Add(factorial_next, ONE)         // i++
		factorial.Multiply(factorial, factorial_next)   // i!
		factorial_inv.Inverse(factorial)                // 1/i!
		factorial_inv.Multiply(&factorial_inv, a_power) // store a^i/i! in factorial_inv as not needed anymore
		out.Add(out, &factorial_inv)                    // out += a^i/i!
	}

	return out
}