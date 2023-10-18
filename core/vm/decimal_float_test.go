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
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/params"
	"github.com/holiman/uint256"
)

// TODO normalize first?
func (d2 *Decimal) eq(d1 *Decimal) bool {
	return d1.c.Cmp(&d2.c) == 0 && d1.q.Cmp(&d2.q) == 0
}



func BenchmarkOpAdd(b *testing.B) {
	intArgs := []*uint256.Int{uint256.NewInt(987349875), uint256.NewInt(987349875), uint256.NewInt(987349875), uint256.NewInt(987349875)}
	benchmarkOpDec(b, intArgs, opAdd)
}

func BenchmarkOpDecAdd(b *testing.B) {
	intArgs := []*uint256.Int{uint256.NewInt(987349875), uint256.NewInt(987349875), uint256.NewInt(987349875), uint256.NewInt(987349875)}
	benchmarkOpDec(b, intArgs, opDecAdd)
}

func BenchmarkOpDecNeg(b *testing.B) {
	intArgs := []*uint256.Int{uint256.NewInt(987349875), uint256.NewInt(987349875)}
	benchmarkOpDec(b, intArgs, opDecNeg)
}

func BenchmarkOpDecMul(b *testing.B) {
	intArgs := []*uint256.Int{uint256.NewInt(987349875), uint256.NewInt(987349875), uint256.NewInt(987349875), uint256.NewInt(987349875)}
	benchmarkOpDec(b, intArgs, opDecMul)
}

func BenchmarkOpDecInv(b *testing.B) {
	// opDecInv benchmark does not depend on precision
	precision := uint256.NewInt(50)
	intArgs := []*uint256.Int{precision, uint256.NewInt(987349875), uint256.NewInt(987349875)}
	benchmarkOpDec(b, intArgs, opDecInv)
}

func BenchmarkOpDecExp0(b *testing.B) {
	// opDecExp benchmark depends on precision
	precision := uint256.NewInt(0);
	intArgs := []*uint256.Int{precision, uint256.NewInt(0), uint256.NewInt(1)}
	fmt.Println("BenchmarkOpDecExp precision=", precision)
	benchmarkOpDec(b, intArgs, opDecExp)
}

func BenchmarkOpDecExp1(b *testing.B) {
	// opDecExp benchmark depends on precision
	precision := uint256.NewInt(1);
	intArgs := []*uint256.Int{precision, uint256.NewInt(0), uint256.NewInt(1)}
	fmt.Println("BenchmarkOpDecExp precision=", precision)
	benchmarkOpDec(b, intArgs, opDecExp)
}

func BenchmarkOpDecExp10(b *testing.B) {
	// opDecExp benchmark depends on precision
	precision := uint256.NewInt(10);
	intArgs := []*uint256.Int{precision, uint256.NewInt(0), uint256.NewInt(1)}
	fmt.Println("BenchmarkOpDecExp precision=", precision)
	benchmarkOpDec(b, intArgs, opDecExp)
}

func BenchmarkOpDecLog20(b *testing.B) {
	// opDecExp benchmark depends on precision
	precision := uint256.NewInt(0);
	intArgs := []*uint256.Int{precision, new(uint256.Int).Neg(uint256.NewInt(1)), uint256.NewInt(15)}
	fmt.Println("BenchmarkOpDecLog2 precision=", precision)
	benchmarkOpDec(b, intArgs, opDecLog2)
}

func BenchmarkOpDecLog21(b *testing.B) {
	// opDecExp benchmark depends on precision
	precision := uint256.NewInt(1);
	intArgs := []*uint256.Int{precision, new(uint256.Int).Neg(uint256.NewInt(1)), uint256.NewInt(15)}
	fmt.Println("BenchmarkOpDecLog2 precision=", precision)
	benchmarkOpDec(b, intArgs, opDecLog2)
}

func BenchmarkOpDecLog210(b *testing.B) {
	// opDecExp benchmark depends on precision
	precision := uint256.NewInt(10);
	intArgs := []*uint256.Int{precision, new(uint256.Int).Neg(uint256.NewInt(1)), uint256.NewInt(15)}
	fmt.Println("BenchmarkOpDecLog2 precision=", precision)
	benchmarkOpDec(b, intArgs, opDecLog2)
}

func BenchmarkOpDecSin0(b *testing.B) {
	// opDecExp benchmark depends on precision
	precision := uint256.NewInt(0);
	intArgs := []*uint256.Int{precision, uint256.NewInt(0), uint256.NewInt(1)}
	fmt.Println("BenchmarkOpDecSin precision=", precision)
	benchmarkOpDec(b, intArgs, opDecSin)
}

func BenchmarkOpDecSin1(b *testing.B) {
	// opDecExp benchmark depends on precision
	precision := uint256.NewInt(1);
	intArgs := []*uint256.Int{precision, uint256.NewInt(0), uint256.NewInt(1)}
	fmt.Println("BenchmarkOpDecSin precision=", precision)
	benchmarkOpDec(b, intArgs, opDecSin)
}

func BenchmarkOpDecSin10(b *testing.B) {
	// opDecExp benchmark depends on precision
	precision := uint256.NewInt(10);
	intArgs := []*uint256.Int{precision, uint256.NewInt(0), uint256.NewInt(1)}
	fmt.Println("BenchmarkOpDecSin precision=", precision)
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

func TestDecUInt256IntToBigInt(t *testing.T) {
	tests := []struct {
		x uint256.Int
		y big.Int
	}{
		{*uint256.NewInt(5), *big.NewInt(5)},
		{*new(uint256.Int).Neg(uint256.NewInt(2)), *big.NewInt(-2)},
	}
	for _, tt := range tests {
		y := UInt256IntToBigInt(&tt.x)
		// fmt.Println(tt.x, y)
		// fmt.Println(y.String())

		if y.Cmp(&tt.y) != 0 {
			t.Fatal(tt.y, y)
		}
	}
}

func TestDecUInt256IntTupleToDecimal(t *testing.T) {
	tests := []struct {
		c uint256.Int
		q uint256.Int
		d Decimal
	}{
		{*uint256.NewInt(5), *uint256.NewInt(2), *createDecimal(big.NewInt(5), big.NewInt(2))},
		{*new(uint256.Int).Neg(uint256.NewInt(2)), *new(uint256.Int).Neg(uint256.NewInt(1)), *createDecimal(big.NewInt(-2), big.NewInt(-1))},
	}
	for _, tt := range tests {
		d := UInt256IntTupleToDecimal(&tt.c, &tt.q)
		// fmt.Println(d)
		// fmt.Println(d.c.String())
		// fmt.Println(d.q.String())
		if !d.eq(&tt.d) {
			t.Fatal(tt.c, tt.q, d, tt.d)
		}
	}
}

func TestDecBigIntToUInt256Int(t *testing.T) {
	tests := []struct {
		x big.Int
		y uint256.Int
	}{
		{*big.NewInt(5), *uint256.NewInt(5)},
		{*big.NewInt(-5), *new(uint256.Int).Neg(uint256.NewInt(5))},
	}
	for _, tt := range tests {
		y := BigIntToUInt256Int(&tt.x)

		if *y != tt.y {
			t.Fatal(tt.x, y, tt.y)
		}
	}
}

func TestDecAdd(t *testing.T) {
	tests := []struct {
		a Decimal
		b Decimal
		c Decimal
	}{
		{*createDecimal(big.NewInt(5), ZERO_BIG), *createDecimal(big.NewInt(121), MINUS_ONE_BIG), *createDecimal(big.NewInt(171), MINUS_ONE_BIG)},
		{*createDecimal(big.NewInt(5), ZERO_BIG), *createDecimal(big.NewInt(121), ZERO_BIG), *createDecimal(big.NewInt(126), ZERO_BIG)},
		{*createDecimal(big.NewInt(-2), MINUS_ONE_BIG), *createDecimal(big.NewInt(8), MINUS_ONE_BIG), *createDecimal(big.NewInt(6), MINUS_ONE_BIG)},
		{*createDecimal(big.NewInt(5), MINUS_ONE_BIG), *createDecimal(big.NewInt(-2), ZERO_BIG), *createDecimal(big.NewInt(-15), MINUS_ONE_BIG)},
	}
	for _, tt := range tests {
		var out, out2 Decimal
		out.Add(&tt.a, &tt.b)
		// fmt.Println("a", showDecimal(&tt.a), "b", showDecimal(&tt.b), "out", showDecimal(&out), "c", showDecimal(&tt.c))

		out2.normalize(&out)
		// fmt.Println("out2", showDecimal(&out2))

		if !out2.eq(&tt.c) {
			t.Fatal(tt.a, tt.b, out, out2, tt.c)
		}
	}
}

func TestDecNegate(t *testing.T) {
	tests := []struct {
		a Decimal
		b Decimal
	}{
		{*createDecimal(big.NewInt(2), ZERO_BIG), *createDecimal(big.NewInt(-2), ZERO_BIG)},
		{*createDecimal(big.NewInt(5), MINUS_ONE_BIG), *createDecimal(big.NewInt(-5), MINUS_ONE_BIG)},
	}
	for _, tt := range tests {
		var out, out2 Decimal
		out.Negate(&tt.a)
		// fmt.Println("a", showDecimal(&tt.a))
		// fmt.Println("b", showDecimal(&tt.b))
		// fmt.Println("out", showDecimal(&out))

		out2.normalize(&out)
		// fmt.Println("out2", showDecimal(&out2))

		if !out2.eq(&tt.b) {
			t.Fatal(tt.a, tt.b, out, out2)
		}
	}
}

func TestDecMultiply(t *testing.T) {
	tests := []struct {
		a Decimal
		b Decimal
		c Decimal
	}{
		{*createDecimal(big.NewInt(2), ZERO_BIG), *createDecimal(big.NewInt(2), ZERO_BIG), *createDecimal(big.NewInt(4), ZERO_BIG)},
		{*createDecimal(big.NewInt(2), ZERO_BIG), *createDecimal(big.NewInt(5), MINUS_ONE_BIG), *createDecimal(big.NewInt(1), ZERO_BIG)},
		{*createDecimal(big.NewInt(-2), ZERO_BIG), *createDecimal(big.NewInt(5), MINUS_ONE_BIG), *createDecimal(big.NewInt(-1), ZERO_BIG)},
		{*createDecimal(big.NewInt(-2), ZERO_BIG), *createDecimal(big.NewInt(-5), MINUS_ONE_BIG), *createDecimal(big.NewInt(1), ZERO_BIG)},
	}
	for _, tt := range tests {
		var out, out2 Decimal
		out.Multiply(&tt.a, &tt.b)

		out2.normalize(&out)
		// fmt.Println("a", showDecimal(&tt.a), "b", showDecimal(&tt.b), "out", showDecimal(&out), "c", showDecimal(&tt.c))

		if !out2.eq(&tt.c) {
			t.Fatal(tt.a, tt.b, out, out2, tt.c)
		}
	}
}

func TestDecInv(t *testing.T) {
	tests := []struct {
		a Decimal
		b Decimal
	}{
		{*copyDecimal(ONE), *copyDecimal(ONE)},
		{*createDecimal(big.NewInt(2), ZERO_BIG), *createDecimal(big.NewInt(5), MINUS_ONE_BIG)},
		{*createDecimal(big.NewInt(-20), MINUS_ONE_BIG), *createDecimal(big.NewInt(-5), MINUS_ONE_BIG)},
	}
	for _, tt := range tests {
		var out, out2 Decimal
		out.Inverse(&tt.a, *big.NewInt(5))
		// fmt.Println("a", showDecimal(&tt.a), "out", showDecimal(&out), "b", showDecimal(&tt.b))

		out2.normalize(&out)
		// fmt.Println("out2", showDecimal(&out2))

		if !out2.eq(&tt.b) {
			t.Fatal(tt.a, out, tt.b)
		}
	}
}

func TestDecNormalize(t *testing.T) {

	LARGE_TEN := big.NewInt(10)
	LARGE_TEN.Exp(LARGE_TEN, big.NewInt(75), ZERO_BIG)

	TEN_TEN := big.NewInt(10)
	TEN_TEN.Exp(TEN_TEN, big.NewInt(10), ZERO_BIG)

	NEG_45 := big.NewInt(-45)
	NEG_55 := big.NewInt(-55)
	// NEG_77 := big.NewInt(-77)
	NEG_75 := big.NewInt(-75)
	// NEG_76 := big.NewInt(-76)

	var TEN_48, FIVE_48, MINUS_FIVE_48 big.Int
	TEN_48.Exp(big.NewInt(10), big.NewInt(48), ZERO_BIG)
	FIVE_48.Mul(big.NewInt(5), &TEN_48)
	MINUS_FIVE_48.Neg(&FIVE_48)
	MINUS_49 := big.NewInt(-49)
	MINUS_5 := big.NewInt(-5)

	tests := []struct {
		a Decimal
		b Decimal
	}{
		{*copyDecimal(ONE), *copyDecimal(ONE)},
		{*createDecimal(big.NewInt(100), big.NewInt(-2)), *copyDecimal(ONE)},
		{*createDecimal(LARGE_TEN, NEG_75), *copyDecimal(ONE)},
		{*createDecimal(TEN_TEN, NEG_55), *createDecimal(ONE_BIG, NEG_45)},
		{*createDecimal(&MINUS_FIVE_48, MINUS_49), *createDecimal(MINUS_5, MINUS_ONE_BIG)},
	}
	for _, tt := range tests {
		var out Decimal
		out.normalize(&tt.a)
		// fmt.Println("a", showDecimal(&tt.a), "out", showDecimal(&out), "b", showDecimal(&tt.b))

		if !out.eq(&tt.b) {
			t.Fatal(tt.a, out, tt.b)
		}
	}
}

func TestDecExp(t *testing.T) {
	tests := []struct {
		a Decimal
		precision big.Int
		b Decimal
	}{
		{*createDecimal(big.NewInt(1), big.NewInt(-1)), *big.NewInt(10), *createDecimal(big.NewInt(27182815251), big.NewInt(-10))},
	}
	for _, tt := range tests {

		var out Decimal
		out.Exp(&tt.a, tt.precision)
		fmt.Println(out.String())

		// if !out.eq(&tt.b) {
		// 	t.Fatal(tt.a, out, tt.b)
		// }
	}
}

func TestDecLog2(t *testing.T) {
	tests := []struct {
		a Decimal
		precision big.Int
		b Decimal
	}{
		{*copyDecimal(HALF),  *big.NewInt(1), *copyDecimal(MINUS_ONE)},
		{*createDecimal(big.NewInt(15), big.NewInt(-1)), *big.NewInt(10), *createDecimal(big.NewInt(58496), big.NewInt(-5))},
	}
	for _, tt := range tests {
		var out Decimal
		// var out, out2 Decimal
		out.Log2(&tt.a, tt.precision)
		fmt.Println(out.String())
		// out2.normalize(&out)
		// if !out.eq(&tt.b) {
		// 	t.Fatal(tt.a, out, tt.b)
		// }
	}
}

func TestDecSin(t *testing.T) {
	tests := []struct {
		a Decimal
		precision big.Int
		b Decimal
	}{
		{*createDecimal(big.NewInt(11), big.NewInt(-1)), *big.NewInt(10), *createDecimal(big.NewInt(891207360), big.NewInt(-9))},
	}
	for _, tt := range tests {
		var out Decimal
		out.Sin(&tt.a, tt.precision)
		fmt.Println(out.String())
		// if !out.eq(&tt.b) {
		// 	t.Fatal(tt.a, out, tt.b)
		// }
	}
}

// func TestDecLnCF(t *testing.T) {
// 	// tests := []struct {
// 	// 	a Decimal
// 	// 	precision big.Int
// 	// 	b Decimal
// 	// }{
// 	// 	{*copyDecimal(ONE), *big.NewInt(10), *createDecimal(big.NewInt(8414709849), big.NewInt(-10))},
// 	// }
// 	// for _, tt := range tests {
// 	x := big.NewInt(1)
// 	y := big.NewInt(1)
// 	steps := big.NewInt(6)
// 	r := lnCF(x, y, *steps)
// 	// r := ln10CF(*big.NewInt(10))
// 	fmt.Println(r.String())
// 	// if !out.eq(&tt.b) {
// 	// 	t.Fatal(tt.a, out, tt.b)
// 	// }
// 	// }
// }

func TestDecLn(t *testing.T) {
	tests := []struct {
		a Decimal
		precision big.Int
		b Decimal
	}{
		// {*copyDecimal(TWO), *big.NewInt(20), *createDecimal(big.NewInt(8414709849), big.NewInt(-10))},
		{*createDecimal(big.NewInt(-1), big.NewInt(-1)), *big.NewInt(19), *createDecimal(big.NewInt(8414709849), big.NewInt(-10))},
	}
	for _, tt := range tests {
		var out Decimal
		out.Ln(&tt.a, tt.precision)
		fmt.Println(out.String())
		// if !out.eq(&tt.b) {
		// 	t.Fatal(tt.a, out, tt.b)
		// }
	}
}
