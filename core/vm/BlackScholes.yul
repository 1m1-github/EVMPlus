object "BlackScholes" {
    code {
        datacopy(0, dataoffset("runtime"), datasize("runtime"))
        return(0, datasize("runtime"))
    }
    object "runtime" {
        code {
            // a = ac*10^aq is a decimal
            // ac, aq int256 as 2's complement

            // CONSTANTS
            let MINUS_ONE := ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff // -1
            let MINUS_FIVE := fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffb // -5
            let LN_2_C := 69314
            let LN_2_Q := MINUS_FIVE

            // Dispatcher
            switch selector()
            case 0xb8e010de /* "set()" */ {
                let ac := calldataload(4)
                let aq := calldataload(36)
                let steps := calldataload(68)
                let cc, cq := verbatim_3i_2o(hex"d4", ac, aq, steps)
                sstore(0, cc)
                sstore(1, cq)
                let value := sload(0)
                mstore(0, value)
                return(0, 32)
            }
            default {
                revert(0, 0)
            }

            // OPCODE -> function

            // a + b = c
            function add(ac, aq, bc, bq) -> cc, cq {
                cc, cq := verbatim_3i_2o(hex"d0", ac, aq, bc, bq)
            }

            // a - b = c
            function sub(ac, aq, bc, bq) -> cc, cq {
                cc, cq := verbatim_3i_2o(hex"d1", ac, aq, bc, bq)
            }

            // a * b = c
            function mul(ac, aq, bc, bq) -> cc, cq {
                cc, cq := verbatim_3i_2o(hex"d2", ac, aq, bc, bq)
            }

            // a / b = c
            function div(ac, aq, bc, bq) -> cc, cq {
                cc, cq := verbatim_3i_2o(hex"d3", ac, aq, bc, bq)
            }

            // exp(a) = b
            function exp(ac, aq) -> bc, bq {
                bc, bq := verbatim_3i_2o(hex"d4", ac, aq, 5)
            }

            // log2(a) = b
            function log2(ac, aq) -> bc, bq {
                bc, bq := verbatim_3i_2o(hex"d5", ac, aq, 5)
            }

            // sin(a) = b
            function sin(ac, aq) -> bc, bq {
                bc, bq := verbatim_3i_2o(hex"d6", ac, aq, 5)
            }

            // derived functions

            // ln(a) = ln(2) * log2(a)
            function ln(ac, aq) -> bc, bq {
                bc, bq := log2(ac, aq)
                bc, bq := mul(LN_C, LN_Q, bc, bq)
            }
            
            // a^b = exp(b * ln(a))
            function pow(ac, aq, bc, bq) -> cc, cq {
                cc, cq := ln(ac, aq)
                cc, cq := mul(bc, bq, cc, cq)
                cc, cq := exp(cc, cq)
            }

            // sqrt(a) = a^(1/2)
            function sqrt(ac, aq) -> bc, bq {
                bc, bq := pow(ac, aq, 5, MINUS_ONE)
            }

            // sqr(a) = a^2
            function sqr(ac, aq) -> bc, bq {
                bc, bq := mul(ac, aq, ac, aq)
            }

            // BlackScholes
            // S: underlying price
            // K: strike
            // r: interest
            // s: volatility
            // T: time
            function d_plus(Sc, Sq, Kc, Kq, rc, rq, sc, sq, Tc, Tq) -> dc, dq, vol_timec, vol_timeq {
                let sigma_sqr_halfc, sigma_sqr_halfq := div(sqr(sc, sq), 2, 0)
                let right_sidec, right_sideq := mul(add(sigma_sqr_halfc, sigma_sqr_halfq, rc, rq), Tc, Tq)
                let left_sidec,left_sideq := ln(div(Sc, Sq, Kc, Kq))
                let dc, dq := add(left_sidec,left_sideq, right_sidec, right_sideq)
                let vol_timec, vol_timeq := mul(sc, sq, sqrt(Tc, Tq))
                let dc, dq := add(vol_timec, vol_timeq, dc, dq)
            }
            function d_minus(Sc, Sq, Kc, Kq, rc, rq, sc, sq, Tc, Tq) -> dc, dq {
                let dpc, dpq, vol_timec, vol_timeq := d_plus(Sc, Sq, Kc, Kq, rc, rq, sc, sq, Tc, Tq)
                sub(dpc, dpq, vol_timec, vol_timeq)
            }

            // approximation
            // 1/(1+exp(-1.65451*a))
            function CDF(ac, aq) -> bc, bq {
                let C := fffffffffffffffffffffffffffffffffffffffffffffffffffffffffffd79b5 // -165451
                let ec, eq := add(exp(mul(C, MINUS_FIVE, ac, aq)), 1, 0)
                bc, bq := div(1, 0, ec, eq)
            }
            function call(Sc, Sq, Kc, Kq, rc, rq, sc, sq, Tc, Tq) -> cc, cq {
                let dpc, dpq, vol_timec, vol_timeq := d_plus(Sc, Sq, Kc, Kq, rc, rq, sc, sq, Tc, Tq)
                let dmc, dmq := d_plus(Sc, Sq, Kc, Kq, rc, rq, sc, sq, Tc, Tq)
                let ac, aq := mul(Sc, Sq, CDF(dpc, dpq))
                let ec, eq := exp(mul(MINUS_ONE, 0, rc, rq), Tc, Tq)
                let bc, bq := mul(mul(Kc, Kq, CDF(dmc, dmq)), ec, eq)
                let cc, cq := sub(ac, aq, bc, bq)
            }

            function selector() -> s {
                s := div(calldataload(0), 0x100000000000000000000000000000000000000000000000000000000)
            }
        }
    }
}
