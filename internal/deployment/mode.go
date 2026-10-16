package deployment

const (
	ModeSelfHosted Mode = iota + 1 // self_hosted
	ModeAirGap                     // airgap
)

// Mode identifies the compiled commercial artifact. It is selected at build
// time and cannot be changed by configuration or environment variables.
//
//go:generate go tool enumer -type=Mode -text -transform=noop -linecomment -output=mode_gen.go
type Mode uint8

// IsAirGap reports whether this process was compiled as the AirGap artifact.
func IsAirGap() bool {
	return Current() == ModeAirGap
}
