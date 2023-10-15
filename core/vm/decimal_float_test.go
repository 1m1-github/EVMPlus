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


func BenchmarkOpDecAdd(b *testing.B) {
	intArgs := []*uint256.Int{uint256.NewInt(987349875), uint256.NewInt(987349875), uint256.NewInt(987349875), uint256.NewInt(987349875)}
	benchmarkOpDec(b, intArgs, opDecAdd)
}

func BenchmarkOpDecSub(b *testing.B) {
	intArgs := []*uint256.Int{uint256.NewInt(987349875), uint256.NewInt(987349875), uint256.NewInt(987349875), uint256.NewInt(987349875)}
	benchmarkOpDec(b, intArgs, opDecSub)
}

func BenchmarkOpDecMul(b *testing.B) {
	intArgs := []*uint256.Int{uint256.NewInt(987349875), uint256.NewInt(987349875), uint256.NewInt(987349875), uint256.NewInt(987349875)}
	benchmarkOpDec(b, intArgs, opDecMul)
}

func BenchmarkOpDecDiv(b *testing.B) {
	intArgs := []*uint256.Int{uint256.NewInt(987349875), uint256.NewInt(987349875), uint256.NewInt(987349875), uint256.NewInt(987349875)}
	benchmarkOpDec(b, intArgs, opDecDiv)
}

func BenchmarkOpDecExp(b *testing.B) {
	intArgs := []*uint256.Int{uint256.NewInt(0), uint256.NewInt(1)}
	benchmarkOpDec(b, intArgs, opDecExp)
}

func BenchmarkOpDecLog2(b *testing.B) {
	intArgs := []*uint256.Int{uint256.NewInt(0), uint256.NewInt(1)}
	benchmarkOpDec(b, intArgs, opDecLog2)
}

func BenchmarkOpDecNorm(b *testing.B) {
	intArgs := []*uint256.Int{uint256.NewInt(0), uint256.NewInt(10000)}
	benchmarkOpDec(b, intArgs, opDecNorm)
}

func benchmarkOpDec(b *testing.B, intArgs[]*uint256.Int, op executionFunc) {
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

func TestEq(t *testing.T) {
	tests := []struct {
		d1 Decimal
		d2 Decimal
		x  bool
	}{
		{*createDecimal(big.NewInt(5), big.NewInt(2)), *createDecimal(big.NewInt(5), big.NewInt(2)), true},
		// {*createDecimal(big.NewInt(10), *big.NewInt(2)}, *createDecimal(big.NewInt(100), *big.NewInt(1)}, true},
	}
	for _, tt := range tests {
		x := tt.d1.Eq(&tt.d2)
		fmt.Println(tt.d1.String())

		if x != tt.x {
			t.Fatal(tt.d1, tt.d2, x, tt.x)
		}
	}
}

func TestUInt256IntToBigInt(t *testing.T) {
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

func TestUInt256IntTupleToDecimal(t *testing.T) {
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
		if !d.Eq(&tt.d) {
			t.Fatal(tt.c, tt.q, d, tt.d)
		}
	}
}

func TestBigIntToUInt256Int(t *testing.T) {
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

func TestAdd(t *testing.T) {
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

		out2.Normalize(&out, 0, true)
		// fmt.Println("out2", showDecimal(&out2))

		if !out2.Eq(&tt.c) {
			t.Fatal(tt.a, tt.b, out, out2, tt.c)
		}
	}
}

func TestSubtract(t *testing.T) {
	tests := []struct {
		a Decimal
		b Decimal
		c Decimal
	}{
		{*createDecimal(big.NewInt(2), ZERO_BIG), *createDecimal(big.NewInt(5), MINUS_ONE_BIG), *createDecimal(big.NewInt(15), MINUS_ONE_BIG)},
		{*createDecimal(big.NewInt(5), MINUS_ONE_BIG), *createDecimal(big.NewInt(2), ZERO_BIG), *createDecimal(big.NewInt(-15), MINUS_ONE_BIG)},
	}
	for _, tt := range tests {
		var out, out2 Decimal
		out.Subtract(&tt.a, &tt.b)
		// fmt.Println("a", showDecimal(&tt.a))
		// fmt.Println("b", showDecimal(&tt.b))
		// fmt.Println("out", showDecimal(&out))

		out2.Normalize(&out, 0, true)
		// fmt.Println("out2", showDecimal(&out2))

		if !out2.Eq(&tt.c) {
			t.Fatal(tt.a, tt.b, out, out2, tt.c)
		}
	}
}

func TestMultiply(t *testing.T) {
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

		out2.Normalize(&out, 0, true)
		// fmt.Println("a", showDecimal(&tt.a), "b", showDecimal(&tt.b), "out", showDecimal(&out), "c", showDecimal(&tt.c))

		if !out2.Eq(&tt.c) {
			t.Fatal(tt.a, tt.b, out, out2, tt.c)
		}
	}
}

func TestInv(t *testing.T) {
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
		out.Inverse(&tt.a)
		// fmt.Println("a", showDecimal(&tt.a), "out", showDecimal(&out), "b", showDecimal(&tt.b))

		out2.Normalize(&out, 0, true)
		// fmt.Println("out2", showDecimal(&out2))

		if !out2.Eq(&tt.b) {
			t.Fatal(tt.a, out, tt.b)
		}
	}
}

func TestDiv(t *testing.T) {
	tests := []struct {
		a Decimal
		b Decimal
		c Decimal
	}{
		{*createDecimal(ONE_BIG, TEN_BIG), *copyDecimal(ONE), *createDecimal(ONE_BIG, TEN_BIG)},
	}
	for _, tt := range tests {
		var out, out2 Decimal
		out.Divide(&tt.a, &tt.b)

		out2.Normalize(&out, 0, true)

		// fmt.Println("a", showDecimal(&tt.a), "b", showDecimal(&tt.b), "out", showDecimal(&out), "c", showDecimal(&tt.c), "out2", showDecimal(&out2))

		if !out2.Eq(&tt.c) {
			t.Fatal(tt.a, tt.b, out, out2, tt.c)
		}
	}
}

func TestNormalize(t *testing.T) {

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
		out.Normalize(&tt.a, 0, true)
		// fmt.Println("a", showDecimal(&tt.a), "out", showDecimal(&out), "b", showDecimal(&tt.b))

		if !out.Eq(&tt.b) {
			t.Fatal(tt.a, out, tt.b)
		}
	}
}

func TestLt(t *testing.T) {
	tests := []struct {
		a Decimal
		b Decimal
		c bool
	}{
		{*createDecimal(big.NewInt(5), ZERO_BIG), *createDecimal(big.NewInt(2), ZERO_BIG), false},
		{*createDecimal(big.NewInt(5), MINUS_ONE_BIG), *createDecimal(big.NewInt(2), ZERO_BIG), true},
	}
	for _, tt := range tests {
		lt := tt.a.LessThan(&tt.b)
		if lt != tt.c {
			t.Fatal(tt.a, tt.b, tt.c)
		}
	}
}

// // func TestRound(t *testing.T) {
// // 	tests := []struct {
// // 		a decimal
// // 		b decimal
// // 	}{
// // 		{*createDecimal(big.NewInt(31415926), *new(big.Int).Neg(big.NewInt(1))}, *createDecimal(big.NewInt(2718281), *new(big.Int).Neg(big.NewInt(6))}},
// // 	}
// // 	for _, tt := range tests {
// // 		var out decimal
// // 		precision := uint64(4)
// // 		round(&tt.a, &out, precision, true, false)
// // 		fmt.Println("a", showDecimal(&tt.a), "out", showDecimal(&out), "b", showDecimal(&tt.b))
// // 		if out != tt.b {
// // 			t.Fatal(tt.a, out, tt.b)
// // 		}
// // 	}
// // }

func TestExp(t *testing.T) {
	STEPS := uint(10)
	tests := []struct {
		a Decimal
		b Decimal
	}{
		{*copyDecimal(ONE), *createDecimal(big.NewInt(2718281), big.NewInt(-6))},
	}
	for _, tt := range tests {

		var out Decimal
		out.Exp(&tt.a, STEPS)
		fmt.Println(out.String())

		// if out != tt.b {
		// 	t.Fatal(tt.a, out, tt.b)
		// }
	}
}

func TestLog(t *testing.T) {
	STEPS := uint(10)
	tests := []struct {
		a Decimal
		b Decimal
	}{
		{*copyDecimal(HALF), *copyDecimal(MINUS_ONE)},
	}
	for _, tt := range tests {
		var out, out2 Decimal
		out.Log2(&tt.a, STEPS)
		
		out2.Normalize(&out, 0, true)
		if !out2.Eq(&tt.b) {
			t.Fatal(tt.a, out, tt.b)
		}
	}
}