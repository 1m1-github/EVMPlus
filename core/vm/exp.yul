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

            function selector() -> s {
                s := div(calldataload(0), 0x100000000000000000000000000000000000000000000000000000000)
            }
        }
    }
}
