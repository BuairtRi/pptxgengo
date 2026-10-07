//go:build !darwin && !linux && !windows

package noreplacedir

import "fmt"

func Publish(from, to string) error {
	return fmt.Errorf("exclusive directory publication is unsupported on this platform")
}
