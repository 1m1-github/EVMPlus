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
                let Sc := calldataload(4)
                let Sq := calldataload(36)
                let Kc := calldataload(68)
                let Kq := calldataload(100)
                let rq := calldataload(132)
                let rc := calldataload(164)
                let sc := calldataload(196)
                let sq := calldataload(228)
                let Tc := calldataload(260)
                let Tq := calldataload(292)
                let precision := calldataload(224)
                let cc, cq := callprice(Sc, Sq, Kc, Kq, rc, rq, sc, sq, Tc, Tq, precision)
                sstore(0, cc)
                sstore(1, cq)
                return(0, 32)
            }
            default {
                revert(0, 0)
            }
            function selector() -> s {
                s := div(calldataload(0), 0x100000000000000000000000000000000000000000000000000000000)
            }

            // BlackScholes
            // S: underlying price
            // K: strike
            // r: interest
            // s: volatility
            // T: time
            function d_plus(Sc, Sq, Kc, Kq, rc, rq, sc, sq, Tc, Tq, precision) -> ac, aq, s_sqrt_T_c, s_sqrt_T_q {
                let s_sqr_c, s_sqr_q := dec_sqr(sc, sq)
                let sigma_sqr_half_c, sigma_sqr_half_q := dec_div(s_sqr_c, s_sqr_q, 2, 0)
                let r_p_s_c, r_p_s_q := dec_add(sigma_sqr_half_c, sigma_sqr_half_q, rc, rq)
                let right_side_c, right_side_q := dec_mul(r_p_s_c, r_p_s_q, Tc, Tq)
                let S_K_c, S_K_q := dec_div(Sc, Sq, Kc, Kq)
                let ln_S_K_c,ln_S_K_q := dec_ln(S_K_c, S_K_q, precision)
                ac, aq := dec_add(ln_S_K_c,ln_S_K_q, right_side_c, right_side_q)
                let sqrt_T_c, sqrt_T_q := dec_sqrt(Tc, Tq, precision)
                s_sqrt_T_c, s_sqrt_T_q := dec_mul(sc, sq, sqrt_T_c, sqrt_T_q)
                ac, aq := dec_add(s_sqrt_T_c, s_sqrt_T_q, ac, aq)
            }
            // approximation
            // 1/(1+dec_exp(-1.65451*a))
            function CDF(ac, aq, precision) -> bc, bq {
                let C := 0xfffffffffffffffffffffffffffffffffffffffffffffffffffffffffffd79b5 // -165451
                let MINUS_FIVE := 0xfffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffb // -5
                bc, bq := dec_mul(C, MINUS_FIVE, ac, aq)
                bc, bq := dec_exp(bc, bq, precision)
                bc, bq := dec_add(bc, bq, 1, 0)
                bc, bq := dec_div(1, 0, bc, bq)
            }
            function callprice(Sc, Sq, Kc, Kq, rc, rq, sc, sq, Tc, Tq, precision) -> ac, aq {
                let MINUS_ONE := 0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff // -1
                let dp_c, dp_q, s_sqrt_T_c, s_sqrt_T_q := d_plus(Sc, Sq, Kc, Kq, rc, rq, sc, sq, Tc, Tq, precision)
                let dm_c, dm_q := dec_sub(dp_c, dp_q, s_sqrt_T_c, s_sqrt_T_q)
                let bc, bq := CDF(dp_c, dp_q, precision)
                bc, bq := dec_mul(Sc, Sq, bc, bq)
                let cdf_dm_c, cdf_dm_q := CDF(dm_c, dm_q, precision)
                let cc, cq := dec_mul(MINUS_ONE, 0, rc, rq)
                cc, cq := dec_mul(cc, cq, Tc, Tq)
                cc, cq := dec_exp(cc, cq, precision)
                cc, cq := dec_mul(Kc, Kq, cc, cq)
                cc, cq := dec_mul(cdf_dm_c, cdf_dm_q, cc, cq)
                ac, aq := dec_sub(bc, bq, cc, cq)
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
            function dec_exp(ac, aq, precision) -> bc, bq {
                bc, bq := verbatim_3i_2o(hex"d4", ac, aq, precision)
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
                cc, cq := dec_exp(cc, cq, precision)
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
