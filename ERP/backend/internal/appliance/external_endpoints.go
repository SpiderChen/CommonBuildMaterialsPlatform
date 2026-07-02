package appliance

import "fmt"

func validateNoMockEndpoint(endpoint string, label string) error {
	if hasMockEndpoint(endpoint) {
		return fmt.Errorf("%s不能使用模拟端点", label)
	}
	return nil
}
