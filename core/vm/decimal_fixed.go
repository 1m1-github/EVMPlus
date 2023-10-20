// math based on https://github.com/JuliaMath/Decimals.jl

package vm

import (
	"github.com/holiman/uint256"
)

type int256 = uint256.Int

// Decimal struct and constructors

// c * 10^q
type Decimal struct {
	c int256 // coefficient, interpreted as int256
	q int256 // exponent, interpreted as int256
}

func copyDecimal(d *Decimal) *Decimal {
	return createDecimal(&d.c, &d.q)
}
func createDecimal(_c, _q *int256) *Decimal {
	var c, q int256
	c.Set(_c)
	q.Set(_q)
	return &Decimal{c, q}
}

// CONSTANTS

var MINUS_ONE_INT256 = new(uint256.Int).Neg(ONE_INT256)
var ZERO_INT256 = uint256.NewInt(0)
var ONE_INT256 = uint256.NewInt(1)
var TWO_INT256 = uint256.NewInt(2)
var FIVE_INT256 = uint256.NewInt(5)
var TEN_INT256 = uint256.NewInt(10)

var MINUS_ONE = createDecimal(MINUS_ONE_INT256, ZERO_INT256)
var HALF = createDecimal(FIVE_INT256, MINUS_ONE_INT256)
var ZERO = createDecimal(ZERO_INT256, ONE_INT256)
var ONE = createDecimal(ONE_INT256, ZERO_INT256)
var TWO = createDecimal(TWO_INT256, ZERO_INT256)

// OPCODE functions

// a + b
func AddDec(ac, aq, bc, bq, precision *int256) (cc, cq *int256) {
	a := createDecimal(ac, aq)
	b := createDecimal(ac, aq)
	a.Add(a, b, precision)
	cc = &a.c
	cq = &a.q
	return
}
func (out *Decimal) Add(a, b *Decimal, precision *int256) *Decimal {
	// ok even if out == a || out == b

	ca := add_helper(a, b)
	cb := add_helper(b, a)

	out.c.Add(&ca, &cb)
	out.q.Set(min(&a.q, &b.q))

	out.normalize(out, precision, false)

	return out
}

// -a
func NegateDec(ac, aq *int256) (bc, bq *int256) {
	a := createDecimal(ac, aq)
	a.Negate(a)
	bc = &a.c
	bq = &a.q
	return
}
func (out *Decimal) Negate(a *Decimal) *Decimal {
	// ok even if out == a
	out.c.Neg(&a.c)
	out.q.Set(&a.q)
	// no need to normalize
	return out
}

// a * b
func MultiplyDec(ac, aq, bc, bq, precision *int256) (cc, cq *int256) {
	a := createDecimal(ac, aq)
	b := createDecimal(ac, aq)
	a.Multiply(a, b, precision)
	cc = &a.c
	cq = &a.q
	return
}
func (out *Decimal) Multiply(a, b *Decimal, precision *int256) *Decimal {
	// ok even if out == a || out == b
	out.c.Mul(&a.c, &b.c)
	out.q.Add(&a.q, &b.q)
	out.normalize(out, precision, false)
	return out
}

// 1 / a
func InverseDec(ac, aq, precision *int256) (bc, bq *int256) {
	a := createDecimal(ac, aq)
	a.Inverse(a, precision)
	bc = &a.c
	bq = &a.q
	return
}
func (out *Decimal) Inverse(a *Decimal, precision *int256) *Decimal {
	// ok even if out == a

	var precision_m_aq int256
	precision_m_aq.Sub(precision, &a.q)
	if precision_m_aq.Cmp(ZERO_INT256) == -1 {
		panic("precision_m_aq NEGATIVE")
	}

	precision_m_aq.Exp(TEN_INT256, &precision_m_aq) // save space: precision_m_aq not needed after
	out.c.Div(&precision_m_aq, &a.c)
	out.q.Neg(precision)

	out.normalize(out, precision, false)

	return out
}

// e^a
// total decimal precision is where a^(taylor_steps+1)/(taylor_steps+1)! == 10^(-target_decimal_precision)
func ExpDec(ac, aq, precision, steps *int256) (bc, bq *int256) {
	a := createDecimal(ac, aq)
	a.Exp(a, precision, steps)
	bc = &a.c
	bq = &a.q
	return
}
func (out *Decimal) Exp(_a *Decimal, precision, steps *int256) *Decimal {
	a := copyDecimal(_a) // in case out == _a

	// out = 1
	out.c.Set(ONE_INT256)
	out.q.Set(ZERO_INT256)

	if a.isZero() {
		return out
	}

	var factorial_inv Decimal
	a_power := copyDecimal(ONE)
	factorial := copyDecimal(ONE)
	factorial_next := copyDecimal(ZERO)

	for i := uint256.NewInt(1); i.Cmp(steps) == -1; i.Add(i, ONE_INT256) { // step 0 skipped as out set to 1
		a_power.Multiply(a_power, a, precision)                    // a^i
		factorial_next.Add(factorial_next, ONE, precision)         // i++
		factorial.Multiply(factorial, factorial_next, precision)   // i!
		factorial_inv.Inverse(factorial, precision)                // 1/i!
		factorial_inv.Multiply(&factorial_inv, a_power, precision) // store a^i/i! in factorial_inv as not needed anymore
		out.Add(out, &factorial_inv, precision)                    // out += a^i/i!
	}

	return out
}

// http://www.claysturner.com/dsp/BinaryLogarithm.pdf
// 0 < a
func Log2Dec(ac, aq, precision, steps *int256) (bc, bq *int256) {
	a := createDecimal(ac, aq)
	a.Log2(a, precision, steps)
	bc = &a.c
	bq = &a.q
	return
}
func (out *Decimal) Log2(a *Decimal, precision, steps *int256) *Decimal {
	// ok if out == a

	if a.c.Sign() != 1 {
		panic("Log2 needs 0 < a")
	}

	// out = 0
	out.c.Set(ZERO_INT256)
	out.q.Set(ONE_INT256)

	// isOne needs a normalized
	var a_norm Decimal
	a_norm.normalize(a, precision, false)
	if a_norm.isOne() {
		return out
	}

	// double a until 1 <= a
	for {
		if !a_norm.lessThan(ONE, precision) {
			break
		}

		a_norm.double()                    // a *= 2
		out.Add(out, MINUS_ONE, precision) // out--
	}

	// half a until a < 2
	for {
		if a_norm.lessThan(TWO, precision) {
			break
		}

		a_norm.halve(precision)      // a /= 2
		out.Add(out, ONE, precision) // out++
	}

	// from here: 1 <= a < 2 <=> 0 < out < 1

	// compare a^2 to 2 to reveal out bit-by-bit
	steps_counter := uint256.NewInt(0) // for now, precision is naiive
	v := copyDecimal(HALF)
	for {
		if steps.Cmp(steps_counter) == 0 {
			break
		}

		a_norm.Multiply(&a_norm, &a_norm, precision) // THIS IS SLOW

		if !a_norm.lessThan(TWO, precision) {
			a_norm.halve(precision) // a /= 2
			out.Add(out, v, precision)
		}

		v.halve(precision)

		steps_counter.Add(steps_counter, ONE_INT256)
	}

	return out
}

// sin(a)
func SinDec(ac, aq, precision, steps *int256) (bc, bq *int256) {
	a := createDecimal(ac, aq)
	a.Sin(a, precision, steps)
	bc = &a.c
	bq = &a.q
	return
}
func (out *Decimal) Sin(_a *Decimal, precision, steps *int256) *Decimal {
	a := copyDecimal(_a) // in case out == _a

	// out = a
	out.c.Set(&a.c)
	out.q.Set(&a.q)

	if a.isZero() || precision.Cmp(ONE_INT256) == 0 {
		return out
	}

	var a_squared, factorial_inv Decimal
	a_squared.Multiply(a, a, precision)
	a_power := copyDecimal(ONE)
	factorial := copyDecimal(ONE)
	factorial_next := copyDecimal(ONE)
	negate := true

	for i := uint256.NewInt(1); i.Cmp(steps) == -1; i.Add(i, ONE_INT256) { // step 0 skipped as out set to a
		a_power.Multiply(a_power, &a_squared, precision) // a^(2i+1)

		factorial_next.Add(factorial_next, ONE, precision)       // i++
		factorial.Multiply(factorial, factorial_next, precision) // i!*2i
		factorial_next.Add(factorial_next, ONE, precision)       // i++
		factorial.Multiply(factorial, factorial_next, precision) // (2i+1)!

		factorial_inv.Inverse(factorial, precision)                // 1/(2i+1)!
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
func min(a, b *int256) (c *int256) {
	if a.Cmp(b) == -1 {
		return a
	} else {
		return b
	}
}

// a == 0
func (a *Decimal) isZero() bool {
	return a.c.IsZero()
}

// a should be normalized
// a == 1 ?
func (a *Decimal) isOne() bool {
	return a.c.Cmp(ONE_INT256) == 0 && a.q.IsZero()
}

// a < 0 ?
func (a *Decimal) isNegative() bool {
	return a.c.Sign() == -1
}

func (d2 *Decimal) eq(d1 *Decimal, precision int256) bool {
	d1_zero := d1.isZero()
	d2_zero := d2.isZero()
	if d1_zero || d2_zero {
		return d1_zero == d2_zero
	}

	d1.normalize(d1, precision, false)
	d2.normalize(d2, precision, false)
	return d1.c.Cmp(&d2.c) == 0 && d1.q.Cmp(&d2.q) == 0
}

// a < b
func (a *Decimal) lessThan(b *Decimal, precision int256) bool {
	var diff Decimal
	diff.Add(a, diff.Negate(b), precision)
	return diff.c.Sign() == -1
}

// a *= 2
func (out *Decimal) double() {
	out.c.Lsh(&out.c, 1)
}

// a /= 2
func (out *Decimal) halve(precision int256) {
	out.Multiply(out, HALF, precision)
}

// c = (-1)^d1.s * d1.c * 10^max(d1.q - d2.q, 0)
func add_helper(d1, d2 *Decimal) (c int256) {
	var exponent_diff int256
	exponent_diff.Sub(&d1.q, &d2.q)
	if exponent_diff.Sign() == -1 {
		exponent_diff = *ZERO_INT256 // shallow copy ok
	}

	c.Exp(TEN_INT256, &exponent_diff)
	c.Mul(&d1.c, &c)

	return c
}

// remove trailing zeros from coefficient
func find_num_trailing_zeros_signed(a *int256) (p, ten_power *int256) {
	var b int256
	b.Set(a)
	if b.Sign() == -1 {
		b.Neg(&b)
	}

	p = uint256.NewInt(0)
	ten_power = uint256.NewInt(10)
	if b.Cmp(ZERO_INT256) != 0 { // if b != 0
		for {
			var m int256
			m.Mod(&b, ten_power)
			if m.Cmp(ZERO_INT256) != 0 { // if b % 10^(p+1) != 0
				break
			}
			p.Add(p, ONE_INT256)
			ten_power.Mul(ten_power, TEN_INT256) // 10^(p+1)
		}
	}
	ten_power.Div(ten_power, TEN_INT256)

	return p, ten_power
}

// remove trailing zeros in coefficient
func (out *Decimal) normalize(a *Decimal, precision int256, rounded bool) *Decimal {
	// ok even if out == a

	p, ten_power := find_num_trailing_zeros_signed(&a.c)
	out.c.Div(&a.c, ten_power) // does not change polarity [in case out == a]

	a_neg := a.isNegative()
	if out.c.Cmp(ZERO_INT256) != 0 || a_neg {
		out.q.Add(&a.q, p)
	} else {
		out.q.Set(ZERO_INT256)
	}

	if rounded {
		return out
	}

	out.round(out, precision, true)
	return out
}

func (out *Decimal) round(a *Decimal, precision int256, normal bool) *Decimal {
	// ok if out == a

	var shift, ten_power int256
	shift.Add(&precision, &a.q)

	if shift.Cmp(ZERO_INT256) == 1 || shift.Cmp(&a.q) == -1 {
		if normal {
			out.c.Set(&a.c)
			out.q.Set(&a.q)
			return out
		}
		out.normalize(a, precision, true)
		return out
	}

	shift.Neg(&shift)
	ten_power.Exp(TEN_INT256, &shift)
	out.c.Div(&a.c, &ten_power)
	out.q.Add(&a.q, &shift)
	if normal {
		return out
	}
	out.normalize(out, precision, true)
	return out
}
