object "BlackScholes" {
    code {
        datacopy(0, dataoffset("runtime"), datasize("runtime"))
        return(0, datasize("runtime"))
    }
    object "runtime" {
        code {
            // a = ac*10^aq is a decimal
            // ac, aq int256 as 2's complement

            // Dispatcher
            switch selector()
            case 0xc4df80c7 /* "callprice(int256,int256,int256,int256,int256,int256,int256,int256,int256,int256,int256)" */ {
                calldatacopy(0, 4, 96)
                let ac := mload(0)
                let aq := mload(32)
                let precision := mload(64)
                let cc := dec_exp(ac, aq, precision)
                // let cc, cq := dec_exp(ac, aq, precision)
                pop(aq)
                sstore(0, cc)
                // sstore(1, cq)
                sstore(2, ac)
                return(0, 32)
            }
            default {
                revert(0, 0)
            }
            function selector() -> s {
                s := div(calldataload(0), 0x100000000000000000000000000000000000000000000000000000000)
            }

            // OPCODE -> function

            // a + b = c
            function dec_add(ac, aq, bc, bq) -> cc, cq {
                cc, cq := verbatim_4i_2o(hex"d0", ac, aq, bc, bq)
            }

            // a - b = c
            function dec_sub(ac, aq, bc, bq) -> cc, cq {
                cc, cq := verbatim_4i_2o(hex"d1", ac, aq, bc, bq)
            }

            // a * b = c
            function dec_mul(ac, aq, bc, bq) -> cc, cq {
                cc, cq := verbatim_4i_2o(hex"d2", ac, aq, bc, bq)
            }

            // a / b = c
            function dec_div(ac, aq, bc, bq) -> cc, cq {
                cc, cq := verbatim_4i_2o(hex"d3", ac, aq, bc, bq)
            }

            // dec_exp(a) = b
            function dec_exp(ac, aq, precision) -> bc {
            // function dec_exp(ac, aq, precision) -> bc, bq {
                // bc, bq := verbatim_3i_2o(hex"d4", ac, aq, precision)
                // bc := verbatim_2i_1o(hex"01", ac, aq)
                let a := 43
                bc := add(ac, aq)
            }

            // dec_log2(a) = b
            function dec_log2(ac, aq, precision) -> bc, bq {
                bc, bq := verbatim_3i_2o(hex"d5", ac, aq, precision)
            }

            // dec_sin(a) = b
            function dec_sin(ac, aq, precision) -> bc, bq {
                bc, bq := verbatim_3i_2o(hex"d6", ac, aq, precision)
            }

            // derived functions

            // dec_ln(a) = dec_ln(2) * dec_log2(a)
            function dec_ln(ac, aq, precision) -> bc, bq {
                let LN_2_C := 6931471805
                let LN_2_Q := 0xfffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff6 // -10
                bc, bq := dec_log2(ac, aq, precision)
                bc, bq := dec_mul(LN_2_C, LN_2_Q, bc, bq)
            }
            
            // a^b = dec_exp(b * dec_ln(a))
            function pow(ac, aq, bc, bq, precision) -> cc, cq {
                cc, cq := dec_ln(ac, aq, precision)
                cc, cq := dec_mul(bc, bq, cc, cq)
                // cc, cq := dec_exp(cc, cq, precision)
            }

            // dec_sqrt(a) = a^(1/2)
            function dec_sqrt(ac, aq, precision) -> bc, bq {
                let MINUS_ONE := 0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff // -1
                bc, bq := pow(ac, aq, 5, MINUS_ONE, precision)
            }

            // dec_sqr(a) = a^2
            function dec_sqr(ac, aq) -> bc, bq {
                bc, bq := dec_mul(ac, aq, ac, aq)
            }
        }
    }
}
