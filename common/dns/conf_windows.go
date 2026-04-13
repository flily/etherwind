//go:build windows

package dns

const (
	DefaultSystemResolverConfigurePath   = `C:\Windows\System32\drivers\etc\resolv.conf`
	DefaultSystemResolverConfigureFormat = ConfigureTypeResolvConf
)
