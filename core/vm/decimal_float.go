package vm

import (
	"fmt"

	"github.com/holiman/uint256"
)

// s * 10^e
// consider s and e as int256 using 2's complement
type decimal struct {
	s uint256.Int // significand
	e uint256.Int // exponent
}

var ZERO = uint256.NewInt(0)

// c.s*10^c.e = a.s*10^a.e + b.s*10^b.e
func add(a, b, out *decimal, precision int64, L bool) (*decimal) {
	if L {fmt.Println("add", "a", "b", a, b)}

	aqmbq := new(uint256.Int).Sub(&a.e, &b.e)
	if L {fmt.Println("add", "aqmbq", aqmbq)}
	
	aqmbq_abs := new(uint256.Int).Abs(aqmbq)

	ten_power := uint256.NewInt(10)
	ten_power.Exp(ten_power, aqmbq_abs) // todo faster way should exist
	if L {fmt.Println("add", "ten_power", ten_power, ten_power.String())}

	ca := new(uint256.Int).Abs(&a.s)
	if L {fmt.Println("add", "ca", ca, ca.String())}

	cb := new(uint256.Int).Abs(&b.s)
	if L {fmt.Println("add", "cb", cb, cb.String())}

	if aqmbq.Cmp(ZERO) == 1 {
		ca.Mul(ca, ten_power)
	} else if aqmbq.Cmp(ZERO) == -1 {
		cb.Mul(cb, ten_power)
	}
	if L {fmt.Println("add", "ca", ca, ca.String())}
	if L {fmt.Println("add", "cb", cb, cb.String())}

	// s = (abs(cx) > abs(cy)) ? x.s : y.s
	// var n bool
	// switch ca.CmpAbs(&cb) {
	// case 1: n = a.n
	// default: n = b.n
	// }
	// if L {fmt.Println("add", "n", n)}
	// s = (abs(cx) > abs(cy)) ? x.s : y.s

	// c = BigInt(cx) + BigInt(cy)
	c := new(uint256.Int).Add(ca, cb)
	if L {fmt.Println("add", "c", c, c.String())}
	// c = BigInt(cx) + BigInt(cy)
	
	// min(x.e, y.e)
	e := a.e
	if (b.e.Cmp(&a.e) == -1) {
		e = b.e
	}
	if L {fmt.Println("add", "q", e)}
	// min(x.q, y.q)

	out.s = *c
	out.e = e

	return out
}