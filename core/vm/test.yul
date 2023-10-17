object "Counter" {
    code {
        // Deploy the contract
        datacopy(0, dataoffset("Runtime"), datasize("Runtime"))
        return(0, datasize("Runtime"))
    }
    object "Runtime" {
        code {
            // The storage slot where the counter is stored
            let slot := 0
            // The function selector
            let selector := calldataload(0)
            // Check if the call data matches the increment function
            if eq(selector, 0xb8e010de) {
                // Increment the counter by 1
                sstore(slot, add(sload(slot), 1))
                // Return the new value
                let value := sload(slot)
                mstore(0, value)
                return(0, 32)
            }
            // Check if the call data matches the get function
            if eq(selector, 0x6d4ce63c) {
                // Return the current value
                let value := sload(slot)
                mstore(0, value)
                return(0, 32)
            }
            // Invalid function selector
            revert(0, 0)
        }
    }
}
