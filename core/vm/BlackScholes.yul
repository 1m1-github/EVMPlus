// solc --strict-assembly BlackScholes.yul >> BlackScholes.txt

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
                // Sc, Sq, Kc, Kq, rc, rq, sc, sq, Tc, Tq, precision
                // 0, 32, 64, 96, 128, 160, 192, 224, 256, 288, 320
                
                // calldatacopy(0, 4, 352)
                // hardcoded to test - TODO as input
                let Sc := 1
                mstore(0, Sc)
                let Sq := 0
                mstore(32, Sq)
                let Kc := 1
                mstore(64, Kc)
                let Kq := 0
                mstore(96, Kq)
                let rc := 0
                mstore(128, rc)
                let rq := 1
                mstore(160, rq)
                let sc := 1
                mstore(192, sc)
                let sq := 0xffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff
                mstore(224, sq)
                let Tc := 1
                mstore(256, Tc)
                let Tq := 0
                mstore(288, Tq)
                let precision := 5
                mstore(320, precision)

                let ac, aq := callprice()
                sstore(0, ac)
                sstore(1, aq)
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

            function r_s2_T() {
                let sc := mload(192)
                let sq := mload(224)
                let s_sqr_c, s_sqr_q := dec_sqr(sc, sq)

                let precision := mload(320)
                let sigma_sqr_half_c, sigma_sqr_half_q := dec_div(s_sqr_c, s_sqr_q, 2, 0, precision)
                
                let rc := mload(128)
                let rq := mload(160)
                let r_p_s_c, r_p_s_q := dec_add(sigma_sqr_half_c, sigma_sqr_half_q, rc, rq)

                let Tc := mload(256)
                let Tq := mload(288)
                let r_s2_T_c, r_s2_T_q := dec_mul(r_p_s_c, r_p_s_q, Tc, Tq)

                mstore(352, r_s2_T_c)
                mstore(384, r_s2_T_q)
            }

            function ln_S_K() {
                let Sc := mload(0)
                let Sq := mload(32)
                let Kc := mload(64)
                let Kq := mload(96)
                let precision := mload(320)
                let S_K_c, S_K_q := dec_div(Sc, Sq, Kc, Kq, precision)
                let ln_S_K_c, ln_S_K_q := dec_ln(S_K_c, S_K_q, precision)
                mstore(416, ln_S_K_c)
                mstore(448, ln_S_K_q)
            }

            function d_plus() {
                r_s2_T()
                let r_s2_T_c := mload(352)
                let r_s2_T_q := mload(384)
                ln_S_K()
                let ln_S_K_c := mload(416)
                let ln_S_K_q := mload(448)
                let ln_S_K_p_r_s2_T_c, ln_S_K_p_r_s2_T_q := dec_add(ln_S_K_c,ln_S_K_q, r_s2_T_c, r_s2_T_q)
                
                let sc := mload(192)
                let sq := mload(224)
                let Tc := mload(256)
                let Tq := mload(288)
                let precision := mload(320)
                let sqrt_T_c, sqrt_T_q := dec_sqrt(Tc, Tq, precision)
                let s_sqrt_T_c, s_sqrt_T_q := dec_mul(sc, sq, sqrt_T_c, sqrt_T_q)
                mstore(352, s_sqrt_T_c)
                mstore(384, s_sqrt_T_q)

                let d_plus_c, d_plus_q := dec_div(ln_S_K_p_r_s2_T_c, ln_S_K_p_r_s2_T_q, s_sqrt_T_c, s_sqrt_T_q, precision)
                mstore(416, d_plus_c)
                mstore(448, d_plus_q)
            }
            function d_minus() {
                let d_plus_c := mload(416)
                let d_plus_q := mload(448)
                let s_sqrt_T_c := mload(352)
                let s_sqrt_T_q := mload(384)
                let d_minus_c, d_minus_q := dec_sub(d_plus_c, d_plus_q, s_sqrt_T_c, s_sqrt_T_q)
                mstore(352, d_minus_c)
                mstore(384, d_minus_q)
            }
            // approximation
            // 1/(1+dec_exp(-1.65451*a))
            function CDF(ac, aq) -> bc, bq {
                let C := 0xfffffffffffffffffffffffffffffffffffffffffffffffffffffffffffd79b5 // -165451
                let MINUS_FIVE := 0xfffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffb // -5
                let precision := mload(320)
                let b1_c, b1_q := dec_mul(C, MINUS_FIVE, ac, aq)
                let b2_c, b2_q := dec_exp(b1_c, b1_q, precision)
                let b3_c, b3_q := dec_add(b2_c, b2_q, 1, 0)
                bc, bq := dec_inv(b3_c, b3_q, precision)
            }
            function cdf_dp_S() {
                let d_plus_c := mload(416)
                let d_plus_q := mload(448)
                let cdf_dp_c, cdf_dp_q := CDF(d_plus_c, d_plus_q)
                
                let Sc := mload(0)
                let Sq := mload(32)
                let cdf_dp_S_c, cdf_dp_S_q := dec_mul(Sc, Sq, cdf_dp_c, cdf_dp_q)

                mstore(416, cdf_dp_S_c)
                mstore(448, cdf_dp_S_q)
            }
            function cdf_dm_K() {
                let rc := mload(128)
                let rq := mload(160)
                let r_n_c, r_n_q := dec_neg(rc, rq)
                let Tc := mload(256)
                let Tq := mload(288)
                let r_T_c, r_T_q := dec_mul(r_n_c, r_n_q, Tc, Tq)
                let precision := mload(320)
                let exp_r_T_c, exp_r_T_q := dec_exp(r_T_c, r_T_q, precision)
                let Kc := mload(64)
                let Kq := mload(96)
                let K_exp_r_T_c, K_exp_r_T_q := dec_mul(Kc, Kq, exp_r_T_c, exp_r_T_q)

                let d_minus_c := mload(352)
                let d_minus_q := mload(384)
                let cdf_dm_c, cdf_dm_q := CDF(d_minus_c, d_minus_q)
                let cdf_dm_K_c, cdf_dm_K_q := dec_mul(cdf_dm_c, cdf_dm_q, K_exp_r_T_c, K_exp_r_T_q)
                
                mstore(352, cdf_dm_K_c)
                mstore(384, cdf_dm_K_q)
            }
            function callprice() -> ac, aq {
                d_plus()
                d_minus()
                cdf_dp_S()
                cdf_dm_K()

                let cdf_dm_K_c := mload(352)
                let cdf_dm_K_q := mload(384)
                let cdf_dp_S_c := mload(416)
                let cdf_dp_S_q := mload(448)

                ac, aq := dec_sub(cdf_dp_S_c, cdf_dp_S_q, cdf_dm_K_c, cdf_dm_K_q)
            }

            // OPCODE -> function

            // a + b = c
            function dec_add(ac, aq, bc, bq) -> cc, cq {
                cc, cq := verbatim_4i_2o(hex"d0", ac, aq, bc, bq)
            }

            // -a = b
            function dec_neg(ac, aq) -> bc, bq {
                bc, bq := verbatim_2i_2o(hex"d1", ac, aq)
            }

            // a * b = c
            function dec_mul(ac, aq, bc, bq) -> cc, cq {
                cc, cq := verbatim_4i_2o(hex"d2", ac, aq, bc, bq)
            }

            // 1 / a = b
            function dec_inv(ac, aq, precision) -> bc, bq {
                bc, bq := verbatim_3i_2o(hex"d3", ac, aq, precision)
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

            // a - b = c
            function dec_sub(ac, aq, bc, bq) -> cc, cq {
                cc, cq := dec_neg(bc, bq)
                cc, cq := dec_add(ac, aq, cc, cq)
            }

            // a / b = c
            function dec_div(ac, aq, bc, bq, precision) -> cc, cq {
                cc, cq := dec_inv(bc, bq, precision)
                cc, cq := dec_mul(ac, aq, cc, cq)
            }

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
