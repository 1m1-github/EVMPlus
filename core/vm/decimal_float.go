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

var ZERO = uint256.NewInt(0)

func isNegativeIfInterpretedAsInt256(value *uint256.Int) bool {
	msb := new(uint256.Int).Lsh(uint256.NewInt(1), 255) // 2^255
	return new(uint256.Int).And(value, msb).Sign() != 0
}

// c.s*10^c.e = a.s*10^a.e + b.s*10^b.e
func add(a, b, out *decimal, L bool) (*decimal) {
	if L {fmt.Println("add", "a", "b", a, b)}

	aqmbq := new(uint256.Int).Sub(&a.e, &b.e)
	if L {fmt.Println("add", "aqmbq", aqmbq)}
	
	aqmbq_abs := new(uint256.Int).Abs(aqmbq)
	if L {fmt.Println("add", "aqmbq_abs", aqmbq_abs)}

	ten_power := uint256.NewInt(10)
	ten_power.Exp(ten_power, aqmbq_abs) // todo faster way should exist
	if L {fmt.Println("add", "ten_power", ten_power, ten_power.String())}

	ca := new(uint256.Int).Abs(&a.s)
	if L {fmt.Println("add", "ca", ca, ca.String())}

	cb := new(uint256.Int).Abs(&b.s)
	if L {fmt.Println("add", "cb", cb, cb.String())}

	aqmbq_neg := isNegativeIfInterpretedAsInt256(aqmbq)
	if !aqmbq_neg {
		ca.Mul(ca, ten_power)
	} else if aqmbq.Cmp(ZERO) != 0 {
		cb.Mul(cb, ten_power)
	}
	if L {fmt.Println("add", "ca", ca, ca.String())}
	if L {fmt.Println("add", "cb", cb, cb.String())}

	c := new(uint256.Int).Add(ca, cb)
	if L {fmt.Println("add", "c", c, c.String())}
	
	// min(a.e, b.e)
	ae_neg := isNegativeIfInterpretedAsInt256(&a.e)
	if L {fmt.Println("add", "b.e", b.e)}
	be_neg := isNegativeIfInterpretedAsInt256(&b.e)
	e := a.e
	if L {fmt.Println("add", "a.e", a.e)}
	if L {fmt.Println("add", "b.e", b.e)}
	if L {fmt.Println("add", "ae_neg", ae_neg)}
	if L {fmt.Println("add", "be_neg", be_neg)}
	if L {fmt.Println("add", "e", e)}
	if ae_neg && !be_neg {
		if L {fmt.Println("1")}
	} else if !ae_neg && be_neg {
		if L {fmt.Println("2")}
		e = b.e
	} else if a.e.Cmp(&b.e) == 1 {
		if L {fmt.Println("3")}
		e = b.e
	}
	if L {fmt.Println("add", "e", e)}
	// min(a.e, b.e)

	out.s = *c
	out.e = e
	if L {fmt.Println("add", "out", out)}

	return out
}

// -a
func negate(a, out *decimal, L bool) (*decimal) {
	out.s.Neg(&a.s)
	out.e = a.e
	return out
}

