package vm

import (
	"math/big"
	"fmt"
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

// a * b
func (out *Decimal) Multiply(a, b *Decimal) *Decimal {
	out.c.Mul(&a.c, &b.c)
	out.q.Add(&a.q, &b.q)
	// normalize?
	return out
}

// 1 / a
func (out *Decimal) Inverse(a *Decimal, precision big.Int) *Decimal {
	precision.Add(&precision, &a.q) // more than max decimal precision on 256 bits

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

// e^a
// total decimal precision is where a^(taylor_steps+1)/(taylor_steps+1)! == 10^(-target_decimal_precision)
func (out *Decimal) Exp(a *Decimal, steps big.Int) *Decimal {

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

	for i := big.NewInt(1); i.Cmp(&steps) == -1; i.Add(i, ONE_BIG) { // step 0 skipped as out set to 1
		a_power.Multiply(a_power, a)                    // a^i
		factorial_next.Add(factorial_next, ONE)         // i++
		factorial.Multiply(factorial, factorial_next)   // i!
		factorial_inv.Inverse(factorial, steps)         // 1/i!
		factorial_inv.Multiply(&factorial_inv, a_power) // store a^i/i! in factorial_inv as not needed anymore
		out.Add(out, &factorial_inv)                    // out += a^i/i!
	}

	return out
}


// http://www.claysturner.com/dsp/BinaryLogarithm.pdf
// 0 < a
func (out *Decimal) Log2(a *Decimal, steps big.Int) *Decimal {
	if a.c.Sign() != 1 {
		panic("Log2 needs 0 < a")
	}

	var a_norm Decimal
	a_norm.normalize(a)

	// out = 0
	out.c.Set(ZERO_BIG)
	out.q.Set(ONE_BIG)

	if a_norm.isOne() {
		return out
	}

	// double a until 1 <= a
	for {
		if !a_norm.lessThan(ONE) {
			break
		}

		a_norm.double()         // a *= 2
		out.Add(out, MINUS_ONE) // out--
	}

	// half a until a < 2
	for {
		if a_norm.lessThan(TWO) {
			break
		}

		a_norm.halve()    // a /= 2
		out.Add(out, ONE) // out++
	}

	// from here: 1 < a < 2 <=> 0 < out < 1

	// compare a^2 to 2 to reveal out bit-by-bit
	steps_counter := big.NewInt(0) // for now, precision is naiive
	v := copyDecimal(HALF)
	for {
		if steps.Cmp(steps_counter) == 0 {
			break
		}

		a_norm.Multiply(&a_norm, &a_norm) // THIS IS SLOW

		if !a_norm.lessThan(TWO) {
			a_norm.halve() // a /= 2
			out.Add(out, v)
		}

		v.halve()

		steps_counter.Add(steps_counter, ONE_BIG)
	}

	return out
}

// ln(1+x/y) using continued fractions: https://en.wikipedia.org/wiki/Natural_logarithm#Continued_fractions
func ln10CF(steps big.Int) *big.Rat {
	a := lnCF(ONE_BIG, big.NewInt(4), steps)
	b := lnCF(big.NewInt(3), big.NewInt(125), steps)
	a.Num().Mul(a.Num(), TEN_BIG)
	b.Num().Mul(b.Num(), big.NewInt(3))
	a.Add(a, b)
	return a
}
func lnCF(x, y *big.Int, steps big.Int) *big.Rat {
	var (
		two_y_plus_x big.Int
		two_x big.Rat
	)
	
	two_y_plus_x.Add(y, y)
	two_y_plus_x.Add(&two_y_plus_x, x)
	
	two_x.Num().Mul(TWO_BIG, x)

	step := big.NewInt(1)

	r := lnCF_recur(x, &two_y_plus_x, step, &steps)
	r.Inv(r)
	
	r.Mul(r, &two_x)

	return r
}
func lnCF_recur(x, two_y_plus_x, step, max_steps *big.Int) *big.Rat {
	var r big.Rat
	r.Num().Add(step, step)
	r.Num().Sub(r.Num(), ONE_BIG)
	r.Num().Mul(r.Num(), two_y_plus_x)

	if step.Cmp(max_steps) == 0 {	
		return &r
	}

	var nextStep big.Int ; nextStep.Add(step, ONE_BIG)
	r2 := lnCF_recur(x, two_y_plus_x, &nextStep, max_steps)
	r2.Inv(r2)

	var num big.Rat
	num.Num().Mul(step, x)
	num.Num().Mul(num.Num(), num.Num())

	r2.Mul(r2, &num)

	return r.Sub(&r, r2)
}

// ln(1+a), |a|<1
func (out *Decimal) Ln(a *Decimal, steps big.Int) *Decimal {

	// ln(a) not defined not a<=0
	var abs_a Decimal
	abs_a.c.Abs(&a.c)
	abs_a.q.Set(&a.q)
	if !abs_a.lessThan(ONE) {
		panic("|a|<1")
	}

	if a.isOne() {
		out.c.Set(ZERO_BIG)
		out.q.Set(ONE_BIG)
		return out
	}

	// out = a
	out.c.Set(&a.c)
	out.q.Set(&a.q)

	var factor Decimal
	a_power := copyDecimal(a)
	i := copyDecimal(TWO)
	max_i := createDecimal(&steps, ZERO_BIG)
	max_i.Add(max_i, ONE)
	negate := true

	for ; i.lessThan(max_i); i.Add(i, ONE) { // step 0 skipped as out set to a
		a_power.Multiply(a_power, a) // a^i

		factor.Inverse(i, steps)         // 1/i
		factor.Multiply(&factor, a_power) // store a^i/i in factor as not needed anymore
		if negate {
			factor.Negate(&factor) // (-1)^i*a^i/i
		}
		negate = !negate

		out.Add(out, &factor) // out += (-1)^i*a^i/i
	}

	return out
}

func (d *Decimal) String() string {
	return fmt.Sprintf("%v*10^%v", d.c.String(), d.q.String())
}

// sin(a)
func (out *Decimal) Sin(a *Decimal, steps big.Int) *Decimal {
	// out = a
	out.c.Set(&a.c)
	out.q.Set(&a.q)

	if a.isZero() || steps.Cmp(ONE_BIG) == 0 {
		return out
	}

	var a_squared, factorial_inv Decimal
	a_squared.Multiply(a, a)
	a_power := copyDecimal(a)
	factorial := copyDecimal(ONE)
	factorial_next := copyDecimal(ONE)
	negate := true

	for i := big.NewInt(1); i.Cmp(&steps) == -1; i.Add(i, ONE_BIG) { // step 0 skipped as out set to a
		a_power.Multiply(a_power, &a_squared) // a^(2i+1)

		factorial_next.Add(factorial_next, ONE)       // i++
		factorial.Multiply(factorial, factorial_next) // i!*2i
		factorial_next.Add(factorial_next, ONE)       // i++
		factorial.Multiply(factorial, factorial_next) // (2i+1)!

		factorial_inv.Inverse(factorial, steps)         // 1/(2i+1)!
		factorial_inv.Multiply(&factorial_inv, a_power) // store a^(2i+1)/(2i+1)! in factorial_inv as not needed anymore
		if negate {
			factorial_inv.Negate(&factorial_inv) // (-1)^i*a^(2i+1)/(2i+1)!
		}
		negate = !negate

		out.Add(out, &factorial_inv) // out += (-1)^i*a^(2i+1)/(2i+1)!
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
func (a *Decimal) lessThan(b *Decimal) bool {
	var diff Decimal
	diff.Add(a, diff.Negate(b))
	return diff.c.Sign() == -1
}

// a *= 2
func (out *Decimal) double() {
	out.c.Lsh(&out.c, 1)
}

// a /= 2
func (out *Decimal) halve() {
	out.Multiply(out, HALF)
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
func (out *Decimal) normalize(a *Decimal) *Decimal {
	p, ten_power := find_num_trailing_zeros_signed(&a.c)
	out.c.Div(&a.c, ten_power)

	a_neg := a.isNegative()
	if out.c.Cmp(ZERO_BIG) != 0 || a_neg {
		out.q.Add(&a.q, p)
	} else {
		out.q.Set(ZERO_BIG)
	}

	return out
}