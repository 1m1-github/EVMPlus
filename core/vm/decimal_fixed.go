// math based on https://github.com/JuliaMath/Decimals.jl

package vm

import (
	"fmt"

	"github.com/holiman/uint256"
)

func (d *Decimal256) String() string {
	c := new(uint256.Int).Set(&d.c)
	q := new(uint256.Int).Set(&d.q)
	cs := ""
	if c.Sign() == -1 {
		cs = "-"
		c.Neg(c)
	}
	qs := ""
	if q.Sign() == -1 {
		qs = "-"
		q.Neg(q)
	}
	return fmt.Sprintf("%v%v*10^%v%v", cs, c.Dec(), qs, q.Dec())
}

type int256 = uint256.Int

// Decimal struct and constructors

// c * 10^q
type Decimal256 struct {
	c int256 // coefficient, interpreted as int256
	q int256 // exponent, interpreted as int256
}

func copyDecimal256(d *Decimal256) *Decimal256 {
	return createDecimal256(&d.c, &d.q)
}
func createDecimal256(_c, _q *int256) *Decimal256 {
	var c, q int256
	c.Set(_c)
	q.Set(_q)
	return &Decimal256{c, q}
}

// CONSTANTS

var MINUS_ONE_INT256 = new(uint256.Int).Neg(ONE_INT256)
var ZERO_INT256 = uint256.NewInt(0)
var ONE_INT256 = uint256.NewInt(1)
var TWO_INT256 = uint256.NewInt(2)
var FIVE_INT256 = uint256.NewInt(5)
var TEN_INT256 = uint256.NewInt(10)

var MINUS_ONE_DECIMAL256 = createDecimal256(MINUS_ONE_INT256, ZERO_INT256)
var HALF_DECIMAL256 = createDecimal256(FIVE_INT256, MINUS_ONE_INT256)
var ZERO_DECIMAL256 = createDecimal256(ZERO_INT256, ONE_INT256)
var ONE_DECIMAL256 = createDecimal256(ONE_INT256, ZERO_INT256)
var TWO_DECIMAL256 = createDecimal256(TWO_INT256, ZERO_INT256)
var TEN_DECIMAL256 = createDecimal256(TEN_INT256, ZERO_INT256)

// OPCODE functions

// a + b
func DecAdd(ac, aq, bc, bq, precision *int256) (cc, cq *int256) {
	a := createDecimal256(ac, aq)
	b := createDecimal256(bc, bq)
	a.Add(a, b, precision)
	cc = &a.c
	cq = &a.q
	return
}
func (out *Decimal256) Add(a, b *Decimal256, precision *int256) *Decimal256 {
	// ok even if out == a || out == b

	ca := add_helper_DECIMAL256(a, b)
	cb := add_helper_DECIMAL256(b, a)

	out.c.Add(&ca, &cb)
	out.q.Set(min_DECIMAL256(&a.q, &b.q))

	out.normalize(out, precision, false)

	return out
}

// -a
func DecNegate(ac, aq *int256) (bc, bq *int256) {
	a := createDecimal256(ac, aq)
	a.Negate(a)
	bc = &a.c
	bq = &a.q
	return
}
func (out *Decimal256) Negate(a *Decimal256) *Decimal256 {
	// ok even if out == a
	out.c.Neg(&a.c)
	out.q.Set(&a.q)
	// no need to normalize
	return out
}

// a * b
func DecMultiply(ac, aq, bc, bq, precision *int256) (cc, cq *int256) {
	a := createDecimal256(ac, aq)
	b := createDecimal256(bc, bq)
	a.Multiply(a, b, precision)
	cc = &a.c
	cq = &a.q
	return
}
func (out *Decimal256) Multiply(a, b *Decimal256, precision *int256) *Decimal256 {
	// ok even if out == a || out == b
	out.c.Mul(&a.c, &b.c)
	out.q.Add(&a.q, &b.q)
	out.normalize(out, precision, false)
	return out
}

// 1 / a
func DecInverse(ac, aq, precision *int256) (bc, bq *int256) {
	a := createDecimal256(ac, aq)
	a.Inverse(a, precision)
	bc = &a.c
	bq = &a.q
	return
}
func (out *Decimal256) Inverse(a *Decimal256, precision *int256) *Decimal256 {
	// ok even if out == a

	var precision_m_aq int256
	precision_m_aq.Sub(precision, &a.q)
	if SignedCmp(&precision_m_aq, ZERO_INT256) == -1 {
		panic("precision_m_aq NEGATIVE")
	}

	precision_m_aq.Exp(TEN_INT256, &precision_m_aq) // save space: precision_m_aq not needed after
	signedDiv(&precision_m_aq, &a.c, &out.c)
	out.q.Neg(precision)

	out.normalize(out, precision, false)

	return out
}

// e^a
// total decimal precision is where a^(taylor_steps+1)/(taylor_steps+1)! == 10^(-target_decimal_precision)
func DecExp(ac, aq, precision, steps *int256) (bc, bq *int256) {
	a := createDecimal256(ac, aq)
	a.Exp(a, precision, steps)
	bc = &a.c
	bq = &a.q
	return
}
func (out *Decimal256) Exp(_a *Decimal256, precision, steps *int256) *Decimal256 {
	a := copyDecimal256(_a) // in case out == _a

	// out = 1
	out.c.Set(ONE_INT256)
	out.q.Set(ZERO_INT256)

	if a.isZero() {
		return out
	}

	var factorial_inv Decimal256
	a_power := copyDecimal256(ONE_DECIMAL256)
	factorial := copyDecimal256(ONE_DECIMAL256)
	factorial_next := copyDecimal256(ZERO_DECIMAL256)

	for i := uint256.NewInt(1); i.Cmp(steps) == -1; i.Add(i, ONE_INT256) { // step 0 skipped as out set to 1
		a_power.Multiply(a_power, a, precision)                       // a^i
		factorial_next.Add(factorial_next, ONE_DECIMAL256, precision) // i++
		factorial.Multiply(factorial, factorial_next, precision)      // i!
		factorial_inv.Inverse(factorial, precision)                   // 1/i!
		factorial_inv.Multiply(&factorial_inv, a_power, precision)    // store a^i/i! in factorial_inv as not needed anymore
		out.Add(out, &factorial_inv, precision)                       // out += a^i/i!
	}

	return out
}

// http://www.claysturner.com/dsp/BinaryLogarithm.pdf
// 0 < a
func DecLog2(ac, aq, precision, steps *int256) (bc, bq *int256) {
	a := createDecimal256(ac, aq)
	a.Log2(a, precision, steps)
	bc = &a.c
	bq = &a.q
	return
}
func (out *Decimal256) Log2(a *Decimal256, precision, steps *int256) *Decimal256 {
	// ok if out == a

	if a.c.Sign() != 1 {
		panic("Log2 needs 0 < a")
	}

	// isOne needs a normalized
	var a_norm Decimal256
	a_norm.normalize(a, precision, false)

	// after a_norm.normalize, in case out == a
	// out = 0
	out.c.Set(ZERO_INT256)
	out.q.Set(ONE_INT256)

	if a_norm.isOne() {
		return out
	}

	// double a until 1 <= a
	for {
		if !a_norm.lessThan(ONE_DECIMAL256, precision) {
			break
		}

		a_norm.double() // a *= 2
		// fmt.Println("log2 after double", a_norm.String())
		out.Add(out, MINUS_ONE_DECIMAL256, precision) // out--
	}

	// half a until a < 2
	for {
		if a_norm.lessThan(TWO_DECIMAL256, precision) {
			break
		}

		a_norm.halve(precision)                 // a /= 2
		out.Add(out, ONE_DECIMAL256, precision) // out++
	}

	// from here: 1 <= a < 2 <=> 0 < out < 1

	// compare a^2 to 2 to reveal out bit-by-bit
	steps_counter := uint256.NewInt(0) // for now, precision is naiive
	v := copyDecimal256(HALF_DECIMAL256)
	for {
		if steps.Cmp(steps_counter) == 0 {
			break
		}

		a_norm.Multiply(&a_norm, &a_norm, precision) // THIS IS SLOW

		if !a_norm.lessThan(TWO_DECIMAL256, precision) {
			a_norm.halve(precision) // a /= 2
			out.Add(out, v, precision)
		}

		v.halve(precision)

		steps_counter.Add(steps_counter, ONE_INT256)
	}

	return out
}

// sin(a)
func DecSin(ac, aq, precision, steps *int256) (bc, bq *int256) {
	a := createDecimal256(ac, aq)
	a.Sin(a, precision, steps)
	bc = &a.c
	bq = &a.q
	return
}
func (out *Decimal256) Sin(_a *Decimal256, precision, steps *int256) *Decimal256 {
	a := copyDecimal256(_a) // in case out == _a

	// out = a
	out.c.Set(&a.c)
	out.q.Set(&a.q)

	if a.isZero() || precision.Cmp(ONE_INT256) == 0 {
		return out
	}

	var a_squared, factorial_inv Decimal256
	a_squared.Multiply(a, a, precision)
	a_power := copyDecimal256(ONE_DECIMAL256)
	factorial := copyDecimal256(ONE_DECIMAL256)
	factorial_next := copyDecimal256(ONE_DECIMAL256)
	negate := true

	for i := uint256.NewInt(1); i.Cmp(steps) == -1; i.Add(i, ONE_INT256) { // step 0 skipped as out set to a
		a_power.Multiply(a_power, &a_squared, precision) // a^(2i+1)

		factorial_next.Add(factorial_next, ONE_DECIMAL256, precision) // i++
		factorial.Multiply(factorial, factorial_next, precision)      // i!*2i
		factorial_next.Add(factorial_next, ONE_DECIMAL256, precision) // i++
		factorial.Multiply(factorial, factorial_next, precision)      // (2i+1)!

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

// -1 if a <  b
//
//	0 if a == b
//	1 if b <  a
func SignedCmp(a, b *int256) int {
	c := a.Cmp(b)

	if c == 0 { // a == b
		return 0
	}

	as := a.Sign()
	bs := b.Sign()

	if as == 0 {
		return -bs
	}
	if bs == 0 {
		return as
	}

	if c == -1 { // a < b
		if a.Sign() == b.Sign() {
			return -1 // a < b
		} else {
			return 1 // b < a
		}
	}

	// c == 1 <=> b < a
	if a.Sign() == b.Sign() {
		return 1 // b < a
	} else {
		return -1 // a < b
	}
}

// min_DECIMAL256(a, b)
func min_DECIMAL256(a, b *int256) (c *int256) {
	if SignedCmp(a, b) == -1 {
		return a
	} else {
		return b
	}
}

// a == 0
func (a *Decimal256) isZero() bool {
	return a.c.IsZero()
}

// a should be normalized
// a == 1 ?
func (a *Decimal256) isOne() bool {
	return a.c.Cmp(ONE_INT256) == 0 && a.q.IsZero() // Cmp ok vs SignedCmp when comparing to zero
}

// a < 0 ?
func (a *Decimal256) isNegative() bool {
	return a.c.Sign() == -1
}

func (d2 *Decimal256) eq(d1 *Decimal256, precision *int256) bool {
	d1_zero := d1.isZero()
	d2_zero := d2.isZero()
	if d1_zero || d2_zero {
		return d1_zero == d2_zero
	}

	d1.normalize(d1, precision, false)
	d2.normalize(d2, precision, false)
	return d1.c.Cmp(&d2.c) == 0 && d1.q.Cmp(&d2.q) == 0 // Cmp ok vs SignedCmp when comparing to zero
}

// a < b
func (a *Decimal256) lessThan(b *Decimal256, precision *int256) bool {
	var diff Decimal256
	diff.Add(a, diff.Negate(b), precision)
	// fmt.Println("lessThan", a.String(), diff.String())
	return diff.c.Sign() == -1
}

// a *= 2
func (out *Decimal256) double() {
	out.c.Lsh(&out.c, 1)
}

// a /= 2
func (out *Decimal256) halve(precision *int256) {
	out.Multiply(out, HALF_DECIMAL256, precision)
}

func signedDiv(numerator, denominator, out *uint256.Int) *uint256.Int {
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

// c = (-1)^d1.s * d1.c * 10^max(d1.q - d2.q, 0)
func add_helper_DECIMAL256(d1, d2 *Decimal256) (c int256) {
	var exponent_diff int256
	exponent_diff.Sub(&d1.q, &d2.q) // GasFastestStep
	if exponent_diff.Sign() == -1 {
		exponent_diff = *ZERO_INT256 // shallow copy ok
	}

	c.Exp(TEN_INT256, &exponent_diff)
	c.Mul(&d1.c, &c) // GasFastStep

	return c
}

// remove trailing zeros from coefficient
func find_num_trailing_zeros_signed_DECIMAL256(a *int256) (p, ten_power *int256) {
	var b int256
	b.Set(a)
	if b.Sign() == -1 {
		b.Neg(&b)
	}

	p = uint256.NewInt(0)
	ten_power = uint256.NewInt(10)
	if b.Cmp(ZERO_INT256) != 0 { // if b != 0  // Cmp ok vs SignedCmp when comparing to zero
		for {
			var m int256
			m.Mod(&b, ten_power)
			if m.Cmp(ZERO_INT256) != 0 { // if b % 10^(p+1) != 0  // Cmp ok vs SignedCmp when comparing to zero
				break
			}
			p.Add(p, ONE_INT256)
			ten_power.Mul(ten_power, TEN_INT256) // 10^(p+1)
		}
	}
	ten_power.Div(ten_power, TEN_INT256) // all positive

	return p, ten_power
}

// remove trailing zeros in coefficient
func (out *Decimal256) normalize(a *Decimal256, precision *int256, rounded bool) *Decimal256 {
	// ok even if out == a

	p, ten_power := find_num_trailing_zeros_signed_DECIMAL256(&a.c)
	signedDiv(&a.c, ten_power, &out.c) // does not change polarity [in case out == a]

	a_neg := a.isNegative()
	if out.c.Cmp(ZERO_INT256) != 0 || a_neg { // Cmp ok vs SignedCmp when comparing to zero
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

func (out *Decimal256) round(a *Decimal256, precision *int256, normal bool) *Decimal256 {
	// ok if out == a

	var shift, ten_power int256
	shift.Add(precision, &a.q)

	if SignedCmp(&shift, ZERO_INT256) == 1 || SignedCmp(&shift, &a.q) == -1 {
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
	signedDiv(&a.c, &ten_power, &out.c)
	out.q.Add(&a.q, &shift)
	if normal {
		return out
	}
	out.normalize(out, precision, true)
	return out
}

// LN using CF
// ln(1+x/y) using continued fractions: https://en.wikipedia.org/wiki/Natural_logarithm#Continued_fractions
func (out *Decimal256) ln10(precision, steps *uint256.Int) *Decimal256 {
	THREE_INT256 := uint256.NewInt(3)
	THREE_DECIMAL256 := createDecimal256(THREE_INT256, ZERO_INT256)
	ONE_OVER_FOUR := createDecimal256(uint256.NewInt(25), new(uint256.Int).Neg(TWO_INT256))
	THREE_OVER_125 := createDecimal256(uint256.NewInt(24), new(uint256.Int).Neg(THREE_INT256))
	var a, b Decimal256
	a.ln(ONE_OVER_FOUR, precision, steps)
	b.ln(THREE_OVER_125, precision, steps)
	a.Multiply(&a, TEN_DECIMAL256, precision)
	b.Multiply(&b, THREE_DECIMAL256, precision)
	out.Add(&a, &b, precision)
	return out
}

func (out *Decimal256) Ln(_x *Decimal256, precision, steps *int256) *Decimal256 {
	x := copyDecimal256(_x)

	// fmt.Println("1", x.String())

	if x.isNegative() {
		panic("Ln: need 0 < x")
	}

	// ln(1) = 0
	if x.isOne() {
		out.c.Set(ZERO_INT256)
		out.q.Set(ONE_INT256)
		return out
	}

	// adjust x
	// divide x by 10 until x in [0,2]
	adjust := uint256.NewInt(0)
	for {
		if x.lessThan(TWO_DECIMAL256, precision) {
			break
		}

		// x /= 10
		x.q.Add(&x.q, MINUS_ONE_INT256)
		adjust.Add(adjust, ONE_INT256)
	}

	// fmt.Println("2", x.String(), adjust.Dec())

	// ln works with 1+x
	x.Add(x, MINUS_ONE_DECIMAL256, precision)

	// fmt.Println("3", x.String())

	// main
	out.ln(x, precision, steps)
	// fmt.Println("4", x.String(), out.String())

	// readjust back
	var LN10 Decimal256
	LN10.ln10(precision, steps)
	adjustDec := createDecimal256(adjust, ZERO_INT256)
	LN10.Multiply(adjustDec, &LN10, precision)
	out.Add(out, &LN10, precision)
	// fmt.Println("5", out.String())

	return out
}
// ln(1+x)
// _x in [-1,1]
func (out *Decimal256) ln(x *Decimal256, precision, steps *int256) *Decimal256 {
	var two_y_plus_x Decimal256
	two_y_plus_x.Add(x, TWO_DECIMAL256, precision)

	step := uint256.NewInt(1)

	out2 := ln_recur(x, &two_y_plus_x, precision, steps, step)
	out.c.Set(&out2.c)
	out.q.Set(&out2.q)
	out.Inverse(out, precision)

	var two_x Decimal256
	two_x.Multiply(x, TWO_DECIMAL256, precision)
	out.Multiply(out, &two_x, precision)

	return out
}

// out !== x
func ln_recur(x, two_y_plus_x *Decimal256, precision, max_steps, step *int256) *Decimal256 {
	var out Decimal256

	stepDec := createDecimal256(step, ZERO_INT256)
	stepDec.Multiply(stepDec, TWO_DECIMAL256, precision)
	stepDec.Add(stepDec, MINUS_ONE_DECIMAL256, precision)
	out.Multiply(stepDec, two_y_plus_x, precision)

	if step.Cmp(max_steps) == 0 {
		return &out
	}

	step.Add(step, ONE_INT256)
	r := ln_recur(x, two_y_plus_x, precision, max_steps, step)
	step.Sub(step, ONE_INT256)
	r.Inverse(r, precision)

	stepDec2 := createDecimal256(step, ZERO_INT256)
	stepDec2.Multiply(stepDec2, x, precision)
	stepDec2.Multiply(stepDec2, stepDec2, precision)

	r.Multiply(stepDec2, r, precision)
	r.Negate(r)

	out.Add(&out, r, precision)

	return &out
}
