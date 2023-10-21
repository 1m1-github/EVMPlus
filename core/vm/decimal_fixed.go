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

var MINUS_ONE_DECIMAL = createDecimal(MINUS_ONE_INT256, ZERO_INT256)
var HALF_DECIMAL = createDecimal(FIVE_INT256, MINUS_ONE_INT256)
var ZERO_DECIMAL = createDecimal(ZERO_INT256, ONE_INT256)
var ONE_DECIMAL = createDecimal(ONE_INT256, ZERO_INT256)
var TWO_DECIMAL = createDecimal(TWO_INT256, ZERO_INT256)
var TEN_DECIMAL = createDecimal(TEN_INT256, ZERO_INT256)

// OPCODE functions

// a + b
func (out *Decimal) Add(a, b *Decimal, precision *int256) *Decimal {
	// ok even if out == a || out == b

	ca := add_helper(a, b)
	cb := add_helper(b, a)

	out.c.Add(&ca, &cb)
	out.q.Set(min_DECIMAL256(&a.q, &b.q))

	out.normalize(out, precision, false)

	return out
}

// -a
func (out *Decimal) Negate(a *Decimal) *Decimal {
	// ok even if out == a
	out.c.Neg(&a.c)
	out.q.Set(&a.q)
	// no need to normalize
	return out
}

// a * b
func (out *Decimal) Multiply(a, b *Decimal, precision *int256) *Decimal {
	// ok even if out == a || out == b
	out.c.Mul(&a.c, &b.c)
	out.q.Add(&a.q, &b.q)
	out.normalize(out, precision, false)
	return out
}

// 1 / a
func (out *Decimal) Inverse(a *Decimal, precision *int256) *Decimal {
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
func (out *Decimal) Exp(_a *Decimal, precision, steps *int256) *Decimal {
	a := copyDecimal(_a) // in case out == _a

	// out = 1
	out.c.Set(ONE_INT256)
	out.q.Set(ZERO_INT256)

	if a.isZero() {
		return out
	}

	var factorial_inv Decimal
	a_power := copyDecimal(ONE_DECIMAL)
	factorial := copyDecimal(ONE_DECIMAL)
	factorial_next := copyDecimal(ZERO_DECIMAL)

	for i := uint256.NewInt(1); i.Cmp(steps) == -1; i.Add(i, ONE_INT256) { // step 0 skipped as out set to 1
		a_power.Multiply(a_power, a, precision)                    // a^i
		factorial_next.Add(factorial_next, ONE_DECIMAL, precision) // i++
		factorial.Multiply(factorial, factorial_next, precision)   // i!
		factorial_inv.Inverse(factorial, precision)                // 1/i!
		factorial_inv.Multiply(&factorial_inv, a_power, precision) // store a^i/i! in factorial_inv as not needed anymore
		out.Add(out, &factorial_inv, precision)                    // out += a^i/i!
	}

	return out
}


// 0 < _a
func (out *Decimal) Ln(_a *Decimal, precision, steps *int256) *Decimal {
	a := copyDecimal(_a)

	if a.isNegative() {
		panic("Ln: need 0 < x")
	}

	// ln(1) = 0
	if a.isOne() {
		out.c.Set(ZERO_INT256)
		out.q.Set(ONE_INT256)
		return out
	}

	// adjust x
	// divide x by 10 until x in [0,2]
	adjust := uint256.NewInt(0)
	for {
		if a.lessThan(TWO_DECIMAL, precision) {
			break
		}

		// x /= 10
		a.q.Add(&a.q, MINUS_ONE_INT256)
		adjust.Add(adjust, ONE_INT256)
	}

	// ln works with 1+x
	a.Add(a, MINUS_ONE_DECIMAL, precision)

	// main
	out.ln(a, precision, steps)

	// readjust back: ln(a*10^n) = ln(a)+n*ln(10)
	var LN10 Decimal
	LN10.ln10(precision, steps)
	adjustDec := createDecimal(adjust, ZERO_INT256)
	LN10.Multiply(adjustDec, &LN10, precision)
	out.Add(out, &LN10, precision)

	return out
}

// sin(a)
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
	a_power := copyDecimal(ONE_DECIMAL)
	factorial := copyDecimal(ONE_DECIMAL)
	factorial_next := copyDecimal(ONE_DECIMAL)
	negate := true

	for i := uint256.NewInt(1); i.Cmp(steps) == -1; i.Add(i, ONE_INT256) { // step 0 skipped as out set to a
		a_power.Multiply(a_power, &a_squared, precision) // a^(2i+1)

		factorial_next.Add(factorial_next, ONE_DECIMAL, precision) // i++
		factorial.Multiply(factorial, factorial_next, precision)   // i!*2i
		factorial_next.Add(factorial_next, ONE_DECIMAL, precision) // i++
		factorial.Multiply(factorial, factorial_next, precision)   // (2i+1)!

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

// convenience methods

func DecAdd(ac, aq, bc, bq, precision *int256) (cc, cq *int256) {
	a := createDecimal(ac, aq)
	b := createDecimal(bc, bq)
	a.Add(a, b, precision)
	cc = &a.c
	cq = &a.q
	return
}
func DecNegate(ac, aq *int256) (bc, bq *int256) {
	a := createDecimal(ac, aq)
	a.Negate(a)
	bc = &a.c
	bq = &a.q
	return
}
func DecMultiply(ac, aq, bc, bq, precision *int256) (cc, cq *int256) {
	a := createDecimal(ac, aq)
	b := createDecimal(bc, bq)
	a.Multiply(a, b, precision)
	cc = &a.c
	cq = &a.q
	return
}
func DecInverse(ac, aq, precision *int256) (bc, bq *int256) {
	a := createDecimal(ac, aq)
	a.Inverse(a, precision)
	bc = &a.c
	bq = &a.q
	return
}
func DecExp(ac, aq, precision, steps *int256) (bc, bq *int256) {
	a := createDecimal(ac, aq)
	a.Exp(a, precision, steps)
	bc = &a.c
	bq = &a.q
	return
}
func DecLn(ac, aq, precision, steps *int256) (bc, bq *int256) {
	a := createDecimal(ac, aq)
	a.Ln(a, precision, steps)
	bc = &a.c
	bq = &a.q
	return
}
func DecSin(ac, aq, precision, steps *int256) (bc, bq *int256) {
	a := createDecimal(ac, aq)
	a.Sin(a, precision, steps)
	bc = &a.c
	bq = &a.q
	return
}

// helpers

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
func (a *Decimal) isZero() bool {
	return a.c.IsZero()
}

// a should be normalized
// a == 1 ?
func (a *Decimal) isOne() bool {
	return a.c.Cmp(ONE_INT256) == 0 && a.q.IsZero() // Cmp ok vs SignedCmp when comparing to zero
}

// a < 0 ?
func (a *Decimal) isNegative() bool {
	return a.c.Sign() == -1
}

func (d2 *Decimal) eq(d1 *Decimal, precision *int256) bool {
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
func (a *Decimal) lessThan(b *Decimal, precision *int256) bool {
	var diff Decimal
	diff.Add(a, diff.Negate(b), precision)
	return diff.c.Sign() == -1
}

// a *= 2
func (out *Decimal) double() {
	out.c.Lsh(&out.c, 1)
}

// a /= 2
func (out *Decimal) halve(precision *int256) {
	out.Multiply(out, HALF_DECIMAL, precision)
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
func add_helper(d1, d2 *Decimal) (c int256) {
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
func (out *Decimal) normalize(a *Decimal, precision *int256, rounded bool) *Decimal {
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

func (out *Decimal) round(a *Decimal, precision *int256, normal bool) *Decimal {
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

// ln helpers

// https://en.wikipedia.org/wiki/Natural_logarithm#Continued_fractions
// using CF (continued fractions) for ln(1+x/y). we set y=1
// ln(1+a), a in [-1,1]
func (out *Decimal) ln(a *Decimal, precision, steps *int256) *Decimal {
	var two_y_plus_x Decimal
	two_y_plus_x.Add(a, TWO_DECIMAL, precision)

	step := uint256.NewInt(1)

	// recursion of continued fraction
	out2 := ln_recur(a, &two_y_plus_x, precision, steps, step)
	out.c.Set(&out2.c)
	out.q.Set(&out2.q)
	out.Inverse(out, precision)

	// 2x / out
	var two_x Decimal
	two_x.Multiply(a, TWO_DECIMAL, precision)
	out.Multiply(out, &two_x, precision)

	return out
}

// ln10 needed for scaling
func (out *Decimal) ln10(precision, steps *uint256.Int) *Decimal {
	THREE_INT256 := uint256.NewInt(3)
	THREE_DECIMAL256 := createDecimal(THREE_INT256, ZERO_INT256)
	ONE_OVER_FOUR := createDecimal(uint256.NewInt(25), new(uint256.Int).Neg(TWO_INT256))
	THREE_OVER_125 := createDecimal(uint256.NewInt(24), new(uint256.Int).Neg(THREE_INT256))
	var a, b Decimal
	a.ln(ONE_OVER_FOUR, precision, steps)
	b.ln(THREE_OVER_125, precision, steps)
	a.Multiply(&a, TEN_DECIMAL, precision)
	b.Multiply(&b, THREE_DECIMAL256, precision)
	out.Add(&a, &b, precision)
	return out
}

// out !== a
func ln_recur(a, two_y_plus_x *Decimal, precision, max_steps, step *int256) *Decimal {
	var out Decimal

	// (2*step-1)*(2+x)
	stepDec := createDecimal(step, ZERO_INT256)
	stepDec.Multiply(stepDec, TWO_DECIMAL, precision)
	stepDec.Add(stepDec, MINUS_ONE_DECIMAL, precision)
	out.Multiply(stepDec, two_y_plus_x, precision)

	// end recursion?
	if step.Cmp(max_steps) == 0 {
		return &out
	}

	// recursion
	step.Add(step, ONE_INT256)
	r := ln_recur(a, two_y_plus_x, precision, max_steps, step)
	step.Sub(step, ONE_INT256)
	r.Inverse(r, precision)

	// (step*x)^2
	stepDec2 := createDecimal(step, ZERO_INT256)
	stepDec2.Multiply(stepDec2, a, precision)
	stepDec2.Multiply(stepDec2, stepDec2, precision)

	r.Multiply(stepDec2, r, precision)
	r.Negate(r)

	out.Add(&out, r, precision)

	return &out
}
