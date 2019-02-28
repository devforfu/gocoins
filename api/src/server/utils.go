package server

import "fmt"

// CheckParameters verifies that dict dictionary has all expected keys.
// If any key from the keys list is missing, the error is returned.
func CheckParameters(dict map[string]string, keys ...string) error {
    for _, key := range keys {
        if _, ok := dict[key]; !ok {
            return fmt.Errorf("required key '%s' is missing", key)
        }
    }
    return nil
}
