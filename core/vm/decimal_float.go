package vm

import (
	"math/big"

	"github.com/holiman/uint256"
)

// Decimal struct and constructors

// c * 10^q
type Decimal struct {
	c big.Int // coefficient
	q big.Int // exponent
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

// Conversions

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
func (d *Decimal) SetUInt256IntTupleFromDecimal(c, q *uint256.Int) {
	c.Set(BigIntToUInt256Int(&d.c))
	q.Set(BigIntToUInt256Int(&d.q))
	// return c, q
}

// CONSTANTS

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

// OPCODE functions

// a + b
func (out *Decimal) Add(_a, _b *Decimal, precision big.Int) *Decimal {

	a := copyDecimal(_a)
	b := copyDecimal(_b)

	ca := add_helper(a, b)
	cb := add_helper(b, a)

	out.c.Add(&ca, &cb)
	out.q.Set(min(&a.q, &b.q))

	out.normalize(out, precision, false)

	return out
}

// -a
func (out *Decimal) Negate(_a *Decimal) *Decimal {
	a := copyDecimal(_a)

	out.c.Neg(&a.c)
	out.q.Set(&a.q)
	return out
}

// a * b
func (out *Decimal) Multiply(_a, _b *Decimal, precision big.Int) *Decimal {
	a := copyDecimal(_a)
	b := copyDecimal(_b)

	out.c.Mul(&a.c, &b.c)
	out.q.Add(&a.q, &b.q)
	out.normalize(out, precision, false)
	return out
}

// 1 / a
func (out *Decimal) Inverse(_a *Decimal, precision big.Int) *Decimal {
	a := copyDecimal(_a)

	var precision_m_aq big.Int
	precision_m_aq.Sub(&precision, &a.q)
	if precision_m_aq.Cmp(ZERO_BIG) == -1 {
		panic("precision_m_aq NEGATIVE")
	}

	precision_m_aq.Exp(TEN_BIG, &precision_m_aq, ZERO_BIG) // aq_m_precision not needed after
	out.c.Div(&precision_m_aq, &a.c)
	out.q.Neg(&precision)

	out.normalize(out, precision, false)

	return out
}

// e^a
// total decimal precision is where a^(taylor_steps+1)/(taylor_steps+1)! == 10^(-target_decimal_precision)
func (out *Decimal) Exp(_a *Decimal, precision big.Int) *Decimal {
	a := copyDecimal(_a)

	// out = 1
	out.c.Set(ONE_BIG)
	out.q.Set(ZERO_BIG)

	if a.isZero() {
		return out
	}

	var factorial_inv Decimal
	a_power := copyDecimal(ONE)
	factorial := copyDecimal(ONE)
	factorial_next := copyDecimal(ZERO)

	for i := big.NewInt(1); i.Cmp(&precision) == -1; i.Add(i, ONE_BIG) { // step 0 skipped as out set to 1
		a_power.Multiply(a_power, a, precision)                    // a^i
		factorial_next.Add(factorial_next, ONE, precision)         // i++
		factorial.Multiply(factorial, factorial_next, precision)   // i!
		factorial_inv.Inverse(factorial, precision)         // 1/i!
		factorial_inv.Multiply(&factorial_inv, a_power, precision) // store a^i/i! in factorial_inv as not needed anymore
		out.Add(out, &factorial_inv, precision)                    // out += a^i/i!
	}

	return out
}


// http://www.claysturner.com/dsp/BinaryLogarithm.pdf
// 0 < a
func (out *Decimal) Log2(_a *Decimal, precision big.Int) *Decimal {
	a := copyDecimal(_a)

	if a.c.Sign() != 1 {
		panic("Log2 needs 0 < a")
	}

	var a_norm Decimal
	a_norm.normalize(a, precision, false)

	// out = 0
	out.c.Set(ZERO_BIG)
	out.q.Set(ONE_BIG)

	if a_norm.isOne() {
		return out
	}

	// double a until 1 <= a
	for {
		if !a_norm.lessThan(ONE, precision) {
			break
		}

		a_norm.double()         // a *= 2
		out.Add(out, MINUS_ONE, precision) // out--
	}

	// half a until a < 2
	for {
		if a_norm.lessThan(TWO, precision) {
			break
		}

		a_norm.halve(precision)    // a /= 2
		out.Add(out, ONE, precision) // out++
	}

	// from here: 1 <= a < 2 <=> 0 < out < 1

	// compare a^2 to 2 to reveal out bit-by-bit
	steps_counter := big.NewInt(0) // for now, precision is naiive
	v := copyDecimal(HALF)
	for {
		if precision.Cmp(steps_counter) == 0 {
			break
		}

		a_norm.Multiply(&a_norm, &a_norm, precision) // THIS IS SLOW

		if !a_norm.lessThan(TWO, precision) {
			a_norm.halve(precision) // a /= 2
			out.Add(out, v, precision)
		}

		v.halve(precision)

		steps_counter.Add(steps_counter, ONE_BIG)
	}

	return out
}

// sin(a)
func (out *Decimal) Sin(_a *Decimal, precision big.Int) *Decimal {
	a := copyDecimal(_a)

	// out = a
	out.c.Set(&a.c)
	out.q.Set(&a.q)

	if a.isZero() || precision.Cmp(ONE_BIG) == 0 {
		return out
	}

	var a_squared, factorial_inv Decimal
	a_squared.Multiply(a, a, precision)
	a_power := copyDecimal(ONE)
	factorial := copyDecimal(ONE)
	factorial_next := copyDecimal(ONE)
	negate := true

	for i := big.NewInt(1); i.Cmp(&precision) == -1; i.Add(i, ONE_BIG) { // step 0 skipped as out set to a
		a_power.Multiply(a_power, &a_squared, precision) // a^(2i+1)

		factorial_next.Add(factorial_next, ONE, precision)       // i++
		factorial.Multiply(factorial, factorial_next, precision) // i!*2i
		factorial_next.Add(factorial_next, ONE, precision)       // i++
		factorial.Multiply(factorial, factorial_next, precision) // (2i+1)!

		factorial_inv.Inverse(factorial, precision)         // 1/(2i+1)!
		factorial_inv.Multiply(&factorial_inv, a_power, precision) // store a^(2i+1)/(2i+1)! in factorial_inv as not needed anymore
		if negate {
			factorial_inv.Negate(&factorial_inv) // (-1)^i*a^(2i+1)/(2i+1)!
		}
		negate = !negate

		out.Add(out, &factorial_inv, precision) // out += (-1)^i*a^(2i+1)/(2i+1)!
	}

	return out
}

// Helpers

// min(a, b)
func min(a, b *big.Int) (c *big.Int) {
	if a.Cmp(b) == -1 {
		return a
	} else {
		return b
	}
}

// a == 0
func (a *Decimal) isZero() bool {
	return a.c.Cmp(ZERO_BIG) == 0
}

// a should be normalized
// a == 1 ?
func (a *Decimal) isOne() bool {
	return a.c.Cmp(ONE_BIG) == 0 && a.q.Cmp(ZERO_BIG) == 0
}

// a < 0 ?
func (a *Decimal) isNegative() bool {
	return a.c.Sign() == -1
}

// a < b
func (a *Decimal) lessThan(b *Decimal, precision big.Int) bool {
	var diff Decimal
	diff.Add(a, diff.Negate(b), precision)
	return diff.c.Sign() == -1
}

// a *= 2
func (out *Decimal) double() {
	out.c.Lsh(&out.c, 1)
}

// a /= 2
func (out *Decimal) halve(precision big.Int) {
	out.Multiply(out, HALF, precision)
}

// c = (-1)^d1.s * d1.c * 10^max(d1.q - d2.q, 0)
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

// remove trailing zeros from coefficient
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

// remove trailing zeros in coefficient
func (out *Decimal) normalize(_a *Decimal, precision big.Int, rounded bool) *Decimal {
	a := copyDecimal(_a)

	p, ten_power := find_num_trailing_zeros_signed(&a.c)
	out.c.Div(&a.c, ten_power)

	a_neg := a.isNegative()
	if out.c.Cmp(ZERO_BIG) != 0 || a_neg {
		out.q.Add(&a.q, p)
	} else {
		out.q.Set(ZERO_BIG)
	}

	if rounded {
		return out
	}

	out.round(out, precision, true)
	return out
}

func (out *Decimal) round(_a *Decimal, precision big.Int, normal bool) *Decimal {
	a := copyDecimal(_a)

	var shift big.Int
	shift.Add(&precision, &a.q)

	if shift.Cmp(ZERO_BIG) == 1 || shift.Cmp(&a.q) == -1 {
		if normal {
			out.c.Set(&a.c)
			out.q.Set(&a.q)
			return out
		}
		out.normalize(a, precision, true)
		return out
	}

	shift.Neg(&shift)
	out.c.Exp(TEN_BIG, &shift, ZERO_BIG)
	out.c.Div(&a.c, &out.c)
	out.q.Add(&a.q, &shift)
	if normal {
		return out
	}
	out.normalize(out, precision, true)
	return out
}