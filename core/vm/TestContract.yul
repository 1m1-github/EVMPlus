object "TestContract" {
    code {
        datacopy(0, dataoffset("runtime"), datasize("runtime"))
        return(0, datasize("runtime"))
    }
    object "runtime" {
        code {
            // Dispatcher
            switch selector()
            case 0xb8e010de /* "set()" */ {
                sstore(0, add(sload(0), 1))
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
