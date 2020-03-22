package orm

// InterfaceSliceContains checks if interface based slice contains value.
func InterfaceSliceContains(slice []interface{}, value interface{}) bool {
	for _, entry := range slice {
		if entry == value {
			return true
		}
	}
	return false
}

// StringSliceContains checks if a string slice contains value.
func StringSliceContains(slice []string, value string) bool {
	interfaceSlice := make([]interface{}, len(slice))
	for _, item := range slice {
		var interfaceItem interface{} = item
		interfaceSlice = append(interfaceSlice, interfaceItem)
	}
	var interfaceValue interface{} = value
	return InterfaceSliceContains(interfaceSlice, interfaceValue)
}

// InterfaceSliceReversed reverse an interface slice.
func InterfaceSliceReversed(slice []interface{}) []interface{} {
	reversed := make([]interface{}, len(slice))
	copy(reversed, slice)
	for i := len(reversed)/2 - 1; i >= 0; i-- {
		opposite := len(reversed) - 1 - i
		reversed[i], reversed[opposite] = reversed[opposite], reversed[i]
	}
	return reversed
}

func StringSlicePrepend(slice []string, entry string) []string {
	slice = append(slice, entry)
	copy(slice[1:], slice)
	slice[0] = entry
	return slice
}
