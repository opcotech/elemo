//go:build airgap

package deployment

// Current returns the compile-time deployment mode of this binary.
func Current() Mode {
	return ModeAirGap
}
