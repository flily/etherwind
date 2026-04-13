//go:build darwin || linux

package dns

const (
	DefaultSystemResolverConfigurePath   = "/etc/resolv.conf"
	DefaultSystemResolverConfigureFormat = ConfigureTypeResolvConf
)
