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

func (d *Decimal) String() string {
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
	return createDecimal(&d.c, &d.q)
}
func createDecimal(_c, _q *big.Int) (*Decimal) {
	var c, q big.Int
	c.Set(_c)
	q.Set(_q)
	return &Decimal{c, q}
}

// TODO all needed?
var MINUS_ONE_BIG = big.NewInt(-1)
var ZERO_BIG = big.NewInt(0)
var ONE_BIG = big.NewInt(1)
var TEN_BIG = big.NewInt(10)

var ZERO = createDecimal(ZERO_BIG, ZERO_BIG)
var ONE = createDecimal(ONE_BIG, ZERO_BIG)

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
func (a *Decimal) Add(b *Decimal) *Decimal {
	ca := add_helper(a, b)
	cb := add_helper(b, a)

	a.c.Add(&ca, &cb)
	a.q.Set(min(&a.q, &b.q))

	return a
}

// -a
func (a *Decimal) Negate() *Decimal {
	a.c.Neg(&a.c)
	return a
}

// a - b
func (a *Decimal) Subtract(b *Decimal) *Decimal {
	// out.Negate(b)
	b.Negate()
	a.Add(b)
	// out.Add(a, out)
	return a
}

// a * b
func (a *Decimal) Multiply(b *Decimal) *Decimal {
	a.c.Mul(&a.c, &b.c)
	a.q.Add(&a.q, &b.q)
	// return normalize(copy(out), out, precision, false, L)
	return a
}

// 1 / a
func (a *Decimal) Inverse() *Decimal {
	max_precision := big.NewInt(50) // TODO choose correct max_precision
	var precision big.Int
	precision.Add(max_precision, &a.q) // more than max decimal precision on 256 bits

	var aq_m_precision big.Int
	aq_m_precision.Sub(&precision, &a.q)
	if aq_m_precision.Cmp(ZERO_BIG) == -1 {
		panic("ae_m_precision NEGATIVE")
	}

	aq_m_precision.Exp(TEN_BIG, &aq_m_precision, ZERO_BIG) // aq_m_precision not needed after
	a.c.Div(&aq_m_precision, &a.c)
	a.q.Neg(&precision)

	return a
}

// a / b
func (a *Decimal) Divide(b *Decimal) *Decimal {
	b_inv := b.copyDecimal()
	b_inv.Inverse()
	a.Multiply(b_inv)
	return a
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
	diff := a.copyDecimal()
	diff.Subtract(b)
	return diff.c.Sign() == -1
}

// e^a
// total decimal precision is where a^(taylor_steps+1)/(taylor_steps+1)! == 10^(-target_decimal_precision)
func (a *Decimal) Exp(taylor_steps uint) *Decimal {
	if a.IsZero() {
		a = ONE.copyDecimal()
		return a
	}
	
	a = ONE.copyDecimal()
	a_power := ONE.copyDecimal()
	factorial := ONE.copyDecimal()
	factorial_next := ZERO.copyDecimal()
	// factorial_inv := ONE.copyDecimal()
	var factorial_inv Decimal

	for i := uint(0); i < taylor_steps; i++ {
		a_power.Multiply(a) // a^i
		factorial_next.Add(ONE) // i + 1
		factorial.Multiply(factorial_next) // i!
		factorial_inv = *factorial.copyDecimal()
		factorial_inv.Inverse() // 1 / i!
		factorial_inv.Multiply(a_power) // store in factorial_inv as not needed anymore
		a.Add(&factorial_inv)
	}

	return a
}

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

func (a *Decimal) Normalize(precision uint64, rounded bool) *Decimal {
	// remove trailing zeros in significand
	p, ten_power := find_num_trailing_zeros_signed(&a.c)
	a.c.Div(&a.c, ten_power)

	// a.q = *ZERO_BIG
	a_neg := a.IsNegative()
	if a.c.Cmp(ZERO_BIG) != 0 || a_neg {
		a.q.Add(&a.q, p)
	} else {
		a.q.Set(ZERO_BIG)
	}

	// if rounded {
	return a
	// }

	// return round(copyDecimal(out), out, precision, true, L)
}
