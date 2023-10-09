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
	"github.com/holiman/uint256"
	"testing"
)

func TestAdd(t *testing.T) {
	tests := []struct {
		a decimal
		b decimal
	}{
		{decimal{*uint256.NewInt(5), *uint256.NewInt(0)}, decimal{*uint256.NewInt(121), *new(uint256.Int).Neg(uint256.NewInt(1))}},
		{decimal{*uint256.NewInt(5), *uint256.NewInt(0)}, decimal{*uint256.NewInt(121), *uint256.NewInt(0)}},
	}
	for _, tt := range tests {
		var c decimal
		add(&tt.a, &tt.b, &c, true)
		fmt.Println("a", tt.a, "b", tt.b, "c", c)
	}
}
