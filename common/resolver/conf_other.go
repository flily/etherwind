//go:build darwin || linux

package resolver

const (
	DefaultSystemResolverConfigurePath   = "/etc/resolv.conf"
	DefaultSystemResolverConfigureFormat = ConfigureTypeResolvConf
)
