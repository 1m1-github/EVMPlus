// Copyright 2017 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.

package vm

import (
	"fmt"
	"testing"

	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
)

var PRECISION = uint256.NewInt(10)

// func (d *Decimal256) String() string {
// 	c := new(uint256.Int).Set(&d.c)
// 	q := new(uint256.Int).Set(&d.q)
// 	cs := ""
// 	if c.Sign() == -1 {
// 		cs = "-"
// 		c.Neg(c)
// 	}
// 	qs := ""
// 	if q.Sign() == -1 {
// 		qs = "-"
// 		q.Neg(q)
// 	}
// 	return fmt.Sprintf("%v%v*10^%v%v", cs, c.Dec(), qs, q.Dec())
// }

func BenchmarkOpAdd(b *testing.B) {
	intArgs := []*uint256.Int{uint256.NewInt(987349875), uint256.NewInt(987349875), uint256.NewInt(987349875), uint256.NewInt(987349875)}
	benchmarkOpDec(b, intArgs, opAdd)
}

func BenchmarkOpDecAdd(b *testing.B) {
	intArgs := []*uint256.Int{PRECISION, uint256.NewInt(987349875), uint256.NewInt(987349875), uint256.NewInt(987349875), uint256.NewInt(987349875)}
	benchmarkOpDec(b, intArgs, opDecAdd)
}

func BenchmarkOpDecNeg(b *testing.B) {
	intArgs := []*uint256.Int{uint256.NewInt(987349875), uint256.NewInt(987349875)}
	benchmarkOpDec(b, intArgs, opDecNeg)
}

func BenchmarkOpDecMul(b *testing.B) {
	intArgs := []*uint256.Int{PRECISION, uint256.NewInt(987349875), uint256.NewInt(987349875), uint256.NewInt(987349875), uint256.NewInt(987349875)}
	benchmarkOpDec(b, intArgs, opDecMul)
}

func BenchmarkOpDecInv(b *testing.B) {
	// opDecInv benchmark does not depend on precision
	intArgs := []*uint256.Int{PRECISION, MINUS_ONE_INT256, uint256.NewInt(1)}
	benchmarkOpDec(b, intArgs, opDecInv)
}

func BenchmarkOpDecExp0(b *testing.B) {
	// opDecExp benchmark depends on steps
	steps := uint256.NewInt(0)
	intArgs := []*uint256.Int{steps, PRECISION, uint256.NewInt(0), uint256.NewInt(1)}
	fmt.Println("BenchmarkOpDecExp steps=", steps)
	benchmarkOpDec(b, intArgs, opDecExp)
}

func BenchmarkOpDecExp1(b *testing.B) {
	// opDecExp benchmark depends on precision
	steps := uint256.NewInt(1)
	intArgs := []*uint256.Int{steps, PRECISION, uint256.NewInt(0), uint256.NewInt(1)}
	fmt.Println("BenchmarkOpDecExp steps=", steps)
	benchmarkOpDec(b, intArgs, opDecExp)
}

func BenchmarkOpDecExp10(b *testing.B) {
	// opDecExp benchmark depends on precision
	steps := uint256.NewInt(10)
	intArgs := []*uint256.Int{steps, PRECISION, uint256.NewInt(0), uint256.NewInt(1)}
	fmt.Println("BenchmarkOpDecExp steps=", steps)
	benchmarkOpDec(b, intArgs, opDecExp)
}

func BenchmarkOpDecLog20(b *testing.B) {
	// opDecExp benchmark depends on precision
	steps := uint256.NewInt(0)
	intArgs := []*uint256.Int{steps, PRECISION, new(uint256.Int).Neg(uint256.NewInt(1)), uint256.NewInt(15)}
	fmt.Println("BenchmarkOpDecLog2 steps=", steps)
	benchmarkOpDec(b, intArgs, opDecLog2)
}

func BenchmarkOpDecLog21(b *testing.B) {
	// opDecExp benchmark depends on precision
	steps := uint256.NewInt(1)
	intArgs := []*uint256.Int{steps, PRECISION, new(uint256.Int).Neg(uint256.NewInt(1)), uint256.NewInt(15)}
	fmt.Println("BenchmarkOpDecLog2 steps=", steps)
	benchmarkOpDec(b, intArgs, opDecLog2)
}

func BenchmarkOpDecLog210(b *testing.B) {
	// opDecExp benchmark depends on precision
	steps := uint256.NewInt(10)
	intArgs := []*uint256.Int{steps, PRECISION, new(uint256.Int).Neg(uint256.NewInt(1)), uint256.NewInt(15)}
	fmt.Println("BenchmarkOpDecLog2 steps=", steps)
	benchmarkOpDec(b, intArgs, opDecLog2)
}

func BenchmarkOpDecSin0(b *testing.B) {
	// opDecExp benchmark depends on precision
	steps := uint256.NewInt(0)
	intArgs := []*uint256.Int{steps, PRECISION, uint256.NewInt(0), uint256.NewInt(1)}
	fmt.Println("BenchmarkOpDecSin steps=", steps)
	benchmarkOpDec(b, intArgs, opDecSin)
}

func BenchmarkOpDecSin1(b *testing.B) {
	// opDecExp benchmark depends on precision
	steps := uint256.NewInt(1)
	intArgs := []*uint256.Int{steps, PRECISION, uint256.NewInt(0), uint256.NewInt(1)}
	fmt.Println("BenchmarkOpDecSin steps=", steps)
	benchmarkOpDec(b, intArgs, opDecSin)
}

func BenchmarkOpDecSin10(b *testing.B) {
	// opDecExp benchmark depends on precision
	steps := uint256.NewInt(10)
	intArgs := []*uint256.Int{steps, PRECISION, uint256.NewInt(0), uint256.NewInt(1)}
	fmt.Println("BenchmarkOpDecSin steps=", steps)
	benchmarkOpDec(b, intArgs, opDecSin)
}

func benchmarkOpDec(b *testing.B, intArgs []*uint256.Int, op executionFunc) {
	var (
		env            = NewEVM(BlockContext{}, TxContext{}, nil, params.TestChainConfig, Config{})
		stack          = newstack()
		scope          = &ScopeContext{nil, stack, nil}
		evmInterpreter = NewEVMInterpreter(env)
	)

	env.interpreter = evmInterpreter

	pc := uint64(0)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		for _, arg := range intArgs {
			stack.push(arg)
		}
		op(&pc, evmInterpreter, scope)
		stack.pop()
		stack.pop()
	}
	b.StopTimer()
}

func BenchmarkDirectLog2(b *testing.B) {
	a := createDecimal256(uint256.NewInt(15), MINUS_ONE_INT256)
	var out Decimal256
	steps := uint256.NewInt(10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out.Log2(a, PRECISION, steps)
	}
	b.StopTimer()
}
func BenchmarkDirectLn(b *testing.B) {
	a := createDecimal256(uint256.NewInt(15), MINUS_ONE_INT256)
	var out Decimal256
	steps := uint256.NewInt(10)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		out.Ln(a, PRECISION, steps)
	}
	b.StopTimer()
}

func TestSignedCmp(t *testing.T) {
	// b := uint256.NewInt(15)
	// a := uint256.NewInt(23)
	a := new(uint256.Int).Neg(uint256.NewInt(14))
	b := new(uint256.Int).Neg(uint256.NewInt(15))
	c := SignedCmp(a, b)
	fmt.Println(c)
}

func TestDecAdd(t *testing.T) {
	tests := []struct {
		a Decimal256
		b Decimal256
		c Decimal256
	}{
		{*createDecimal256(uint256.NewInt(5), ZERO_INT256), *createDecimal256(uint256.NewInt(121), MINUS_ONE_INT256), *createDecimal256(uint256.NewInt(171), MINUS_ONE_INT256)},
		{*createDecimal256(uint256.NewInt(5), ZERO_INT256), *createDecimal256(uint256.NewInt(121), ZERO_INT256), *createDecimal256(uint256.NewInt(126), ZERO_INT256)},
		{*createDecimal256(new(uint256.Int).Neg(TWO_INT256), MINUS_ONE_INT256), *createDecimal256(uint256.NewInt(8), MINUS_ONE_INT256), *createDecimal256(uint256.NewInt(6), MINUS_ONE_INT256)},
		{*createDecimal256(uint256.NewInt(5), MINUS_ONE_INT256), *createDecimal256(new(uint256.Int).Neg(TWO_INT256), ZERO_INT256), *createDecimal256(new(uint256.Int).Neg(uint256.NewInt(15)), MINUS_ONE_INT256)},
	}
	for _, tt := range tests {
		var out Decimal256
		out.Add(&tt.a, &tt.b, PRECISION)
		// fmt.Println("a", showDecimal(&tt.a), "b", showDecimal(&tt.b), "out", showDecimal(&out), "c", showDecimal(&tt.c))

		if !out.eq(&tt.c, PRECISION) {
			t.Fatal(tt.a, tt.b, out, tt.c)
		}
	}
}

func TestDecNegate(t *testing.T) {
	tests := []struct {
		a Decimal256
		b Decimal256
	}{
		{*createDecimal256(uint256.NewInt(2), ZERO_INT256), *createDecimal256(new(uint256.Int).Neg(TWO_INT256), ZERO_INT256)},
		{*createDecimal256(uint256.NewInt(5), MINUS_ONE_INT256), *createDecimal256(new(uint256.Int).Neg(FIVE_INT256), MINUS_ONE_INT256)},
	}
	for _, tt := range tests {
		var out Decimal256
		out.Negate(&tt.a)
		// fmt.Println("a", showDecimal(&tt.a))
		// fmt.Println("b", showDecimal(&tt.b))
		// fmt.Println("out", showDecimal(&out))

		if !out.eq(&tt.b, PRECISION) {
			t.Fatal(tt.a, tt.b, out)
		}
	}
}

func TestDecMultiply(t *testing.T) {
	tests := []struct {
		a Decimal256
		b Decimal256
		c Decimal256
	}{
		{*createDecimal256(uint256.NewInt(2), ZERO_INT256), *createDecimal256(uint256.NewInt(2), ZERO_INT256), *createDecimal256(uint256.NewInt(4), ZERO_INT256)},
		{*createDecimal256(uint256.NewInt(2), ZERO_INT256), *createDecimal256(uint256.NewInt(5), MINUS_ONE_INT256), *createDecimal256(uint256.NewInt(1), ZERO_INT256)},
		{*createDecimal256(new(uint256.Int).Neg(TWO_INT256), ZERO_INT256), *createDecimal256(uint256.NewInt(5), MINUS_ONE_INT256), *createDecimal256(MINUS_ONE_INT256, ZERO_INT256)},
		{*createDecimal256(new(uint256.Int).Neg(TWO_INT256), ZERO_INT256), *createDecimal256(new(uint256.Int).Neg(FIVE_INT256), MINUS_ONE_INT256), *createDecimal256(uint256.NewInt(1), ZERO_INT256)},
	}
	for _, tt := range tests {
		var out Decimal256
		out.Multiply(&tt.a, &tt.b, PRECISION)

		if !out.eq(&tt.c, PRECISION) {
			t.Fatal(tt.a, tt.b, out, tt.c)
		}
	}
}

func TestDecInv(t *testing.T) {
	tests := []struct {
		a Decimal256
		b Decimal256
	}{
		{*copyDecimal256(ONE_DECIMAL256), *copyDecimal256(ONE_DECIMAL256)},
		{*createDecimal256(uint256.NewInt(2), ZERO_INT256), *createDecimal256(uint256.NewInt(5), MINUS_ONE_INT256)},
		{*createDecimal256(new(uint256.Int).Neg(uint256.NewInt(20)), MINUS_ONE_INT256), *createDecimal256(new(uint256.Int).Neg(FIVE_INT256), MINUS_ONE_INT256)},
		{*createDecimal256(uint256.NewInt(2), ONE_INT256), *createDecimal256(uint256.NewInt(5), new(uint256.Int).Neg(TWO_INT256))},
		{*createDecimal256(uint256.NewInt(2), MINUS_ONE_INT256), *createDecimal256(uint256.NewInt(5), ZERO_INT256)},
	}
	for _, tt := range tests {
		var out Decimal256
		out.Inverse(&tt.a, PRECISION)
		// fmt.Println("a", showDecimal(&tt.a), "out", showDecimal(&out), "b", showDecimal(&tt.b))

		if !out.eq(&tt.b, PRECISION) {
			t.Fatal(tt.a, out, tt.b)
		}
	}
}

func TestDecNormalize(t *testing.T) {

	LARGE_TEN := uint256.NewInt(10)
	LARGE_TEN.Exp(LARGE_TEN, uint256.NewInt(75))

	TEN_TEN := uint256.NewInt(10)
	TEN_TEN.Exp(TEN_TEN, uint256.NewInt(10))

	NEG_45 := new(uint256.Int).Neg(uint256.NewInt(45))
	NEG_55 := new(uint256.Int).Neg(uint256.NewInt(55))
	// NEG_77 := new(uint256.Int).Neg(uint256.NewInt(77))
	NEG_75 := new(uint256.Int).Neg(uint256.NewInt(75))
	// NEG_76 := new(uint256.Int).Neg(uint256.NewInt(76))

	var TEN_48, FIVE_48, MINUS_FIVE_48 uint256.Int
	TEN_48.Exp(uint256.NewInt(10), uint256.NewInt(48))
	FIVE_48.Mul(uint256.NewInt(5), &TEN_48)
	MINUS_FIVE_48.Neg(&FIVE_48)
	MINUS_49 := new(uint256.Int).Neg(uint256.NewInt(49))
	MINUS_5 := new(uint256.Int).Neg(FIVE_INT256)

	tests := []struct {
		a Decimal256
		b Decimal256
		rounded bool
	}{
		{*createDecimal256(uint256.NewInt(15), MINUS_ONE_INT256), *createDecimal256(uint256.NewInt(15), MINUS_ONE_INT256), false},
		{*copyDecimal256(ONE_DECIMAL256), *copyDecimal256(ONE_DECIMAL256), false},
		{*createDecimal256(uint256.NewInt(100), new(uint256.Int).Neg(TWO_INT256)), *copyDecimal256(ONE_DECIMAL256), false},
		{*createDecimal256(LARGE_TEN, NEG_75), *copyDecimal256(ONE_DECIMAL256), false},
		{*createDecimal256(TEN_TEN, NEG_55), *createDecimal256(ONE_INT256, NEG_45), true},
		{*createDecimal256(&MINUS_FIVE_48, MINUS_49), *createDecimal256(MINUS_5, MINUS_ONE_INT256), false},
	}
	for _, tt := range tests {
		var out Decimal256
		out.normalize(&tt.a, PRECISION, tt.rounded)
		// fmt.Println("normalize", tt.a.String(), out.String(), tt.b.String())

		if !out.eq(&tt.b, PRECISION) {
			t.Fatal(tt.a, out, tt.b)
		}
	}
}

func TestDecExp(t *testing.T) {
	tests := []struct {
		a     Decimal256
		steps uint256.Int
		b     Decimal256
	}{
		{*copyDecimal256(ONE_DECIMAL256), *uint256.NewInt(10), *createDecimal256(uint256.NewInt(27182815251), new(uint256.Int).Neg(TEN_INT256))},
		{*createDecimal256(MINUS_ONE_INT256, uint256.NewInt(0)), *uint256.NewInt(10), *createDecimal256(uint256.NewInt(3678791887), new(uint256.Int).Neg(TEN_INT256))},
	}
	for _, tt := range tests {

		var out Decimal256
		out.Exp(&tt.a, PRECISION, &tt.steps)
		// fmt.Println(out.String())

		if !out.eq(&tt.b, PRECISION) {
			t.Fatal(tt.a, out, tt.b)
		}
	}
}

func TestDecLog2(t *testing.T) {
	tests := []struct {
		a     Decimal256
		steps uint256.Int
		b     Decimal256
	}{
		{*createDecimal256(uint256.NewInt(15), MINUS_ONE_INT256), *uint256.NewInt(0), *createDecimal256(uint256.NewInt(5849609375), new(uint256.Int).Neg(TEN_INT256))},
		{*copyDecimal256(HALF_DECIMAL256), *uint256.NewInt(1), *copyDecimal256(MINUS_ONE_DECIMAL256)},
		{*createDecimal256(uint256.NewInt(15), MINUS_ONE_INT256), *uint256.NewInt(10), *createDecimal256(uint256.NewInt(5849609375), new(uint256.Int).Neg(TEN_INT256))},
		{*createDecimal256(uint256.NewInt(100000), new(uint256.Int).Neg(FIVE_INT256)), *uint256.NewInt(5), *copyDecimal256(ZERO_DECIMAL256)},
	}
	for _, tt := range tests {
		var out Decimal256
		// var out, out2 Decimal
		out.Log2(&tt.a, PRECISION, &tt.steps)
		fmt.Println(out.String())
		// if !out.eq(&tt.b, PRECISION) {
		// 	t.Fatal(tt.a, out, tt.b)
		// }
	}
}

func TestDecSin(t *testing.T) {
	tests := []struct {
		a     Decimal256
		steps uint256.Int
		b     Decimal256
	}{
		{*copyDecimal256(ONE_DECIMAL256), *uint256.NewInt(10), *createDecimal256(uint256.NewInt(8414709849), new(uint256.Int).Neg(TEN_INT256))},
	}
	for _, tt := range tests {
		var out Decimal256
		out.Sin(&tt.a, PRECISION, &tt.steps)
		// fmt.Println(out.String())
		if !out.eq(&tt.b, PRECISION) {
			t.Fatal(tt.a, out, tt.b)
		}
	}
}

// func TestDecBS(t *testing.T) {
// 	tests := []struct {
// 		S Decimal
// 		K Decimal
// 		r Decimal
// 		s Decimal
// 		T Decimal
// 		precision uint256.Int
// 		y Decimal
// 	}{
// 		{
// 			*createDecimal(uint256.NewInt(11), uint256.NewInt(-1)),
// 			*copyDecimal(ONE),
// 			*copyDecimal(ZERO),
// 			*createDecimal(uint256.NewInt(1), uint256.NewInt(-1)),
// 			*copyDecimal(ONE),
// 			*uint256.NewInt(10),
// 			*copyDecimal(ZERO),
// 		},
// 	}
// 	for _, tt := range tests {
// 		var out Decimal
// 		out.callprice(&tt.S,&tt.K,&tt.r,&tt.s,&tt.T,tt.precision)
// 		fmt.Println("callprice", out.String())
// 		// if !out.eq(&tt.b) {
// 		// 	t.Fatal(tt.a, out, tt.b)
// 		// }
// 	}
// }
// var LN_2 = createDecimal(uint256.NewInt(6931471805), uint256.NewInt(-10))
// func (out *Decimal) callprice(S,K,r,s,T *Decimal, precision uint256.Int) {
// 	var dp,dm Decimal
// 	dp.d_plus(S,K,r,s,T,precision)
// 	fmt.Println("d_plus", dp.String())
// 	dm.d_minus(S,K,r,s,T,precision)
// 	fmt.Println("d_minus", dm.String())
// 	dp.CDF(&dp, precision)
// 	fmt.Println("dp.CDF", dp.String())
// 	dm.CDF(&dm, precision)
// 	fmt.Println("dm.CDF", dm.String())

// 	out.Multiply(&dp, S)

// 	var a Decimal
// 	a.Negate(r)
// 	a.Multiply(&a, T)
// 	a.Exp(&a, precision)
// 	a.Multiply(&a, K)
// 	a.Multiply(&a, &dm)
// 	a.Negate(&a)
// 	fmt.Println("right side", a.String())

// 	out.Add(out, &a)
// }
// func (out *Decimal) CDF(x *Decimal, precision uint256.Int) {
// 	C := createDecimal(uint256.NewInt(-165451), uint256.NewInt(-5))
// 	out.Multiply(C, x)
// 	out.Exp(out, precision)
// 	out.Add(out, ONE)
// 	out.Inverse(out, precision)
// }
// func (out *Decimal) d_minus(S,K,r,s,T *Decimal, precision uint256.Int) {
// 	out.d_plus(S,K,r,s,T,precision)

// 	var a Decimal
// 	a.sqrt(T, precision)
// 	a.Multiply(&a, s)
// 	a.Negate(&a)

// 	out.Add(out, &a)
// }
// func (out *Decimal) d_plus(S,K,r,s,T *Decimal, precision uint256.Int) {
// 	out.Inverse(K, precision)
// 	out.Multiply(S, out)
// 	out.ln(out, precision)

// 	var a Decimal
// 	a.Multiply(s, s)
// 	a.Multiply(&a, HALF)
// 	a.Add(r, &a)
// 	a.Multiply(&a, T)

// 	out.Add(out, &a)

// 	a.sqrt(T, precision)
// 	a.Multiply(&a, s)
// 	a.Inverse(&a, precision)

// 	out.Multiply(out, &a)
// }
// func (out *Decimal) sqrt(a *Decimal, precision uint256.Int) {
// 	out.pow(a, HALF, precision)
// }
// func (out *Decimal) ln(a *Decimal, precision uint256.Int) {
// 	out.Log2(a, precision)
// 	out.Multiply(out, LN_2)
// }
// func (out *Decimal) pow(a, b *Decimal, precision uint256.Int) {
// 	out.ln(a, precision)
// 	out.Multiply(out, b)
// 	out.Exp(out, precision)
// }

func TestDecLn(t *testing.T) {
	tests := []struct {
		a     Decimal256
		steps uint256.Int
		b     Decimal256
	}{
		{*ONE_DECIMAL256, *uint256.NewInt(10), *createDecimal256(uint256.NewInt(5849609375), new(uint256.Int).Neg(TEN_INT256))},
	}
	for _, tt := range tests {
		var out Decimal256
		// var out, out2 Decimal
		out.Ln(&tt.a, PRECISION, &tt.steps)
		fmt.Println(out.String())
		// if !out.eq(&tt.b, PRECISION) {
		// 	t.Fatal(tt.a, out, tt.b)
		// }
	}
}