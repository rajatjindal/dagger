package impl

import (
	"fmt"
)

// returns a device name and cidr to use; enables us to have unique devices+ip ranges for nested
// engine services to prevent conflicts
func GetUniqueNestedEngineNetwork(index int) (deviceName string, cidr string) {
	return fmt.Sprintf("dagger%d", index), fmt.Sprintf("10.89.%d.0/24", index)
}
