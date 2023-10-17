object "Plus" {
    code {
        datacopy(0, dataoffset("runtime"), datasize("runtime"))
        return(0, datasize("runtime"))
    }
    object "runtime" {
        code {
            // Dispatcher
            switch selector()
            case 0xb8e010de /* "set()" */ {
                let ac := calldataload(4)
                let aq := calldataload(36)
                let bc := calldataload(68)
                let bq := calldataload(100)
                let cc, cq := verbatim_4i_2o(hex"d0", ac, aq, bc, bq)
                sstore(0, cc)
                sstore(0, cq)
                let value := sload(0)
                mstore(0, value)
                return(0, 32)
            }
            default {
                revert(0, 0)
            }

            function selector() -> s {
                s := div(calldataload(0), 0x100000000000000000000000000000000000000000000000000000000)
            }
        }
    }
}
