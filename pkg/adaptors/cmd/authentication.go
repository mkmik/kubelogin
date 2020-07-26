package cmd

import (
	"fmt"
	"strings"

	"github.com/int128/kubelogin/pkg/usecases/authentication"
	"github.com/spf13/pflag"
	"golang.org/x/xerrors"
)

type authenticationOptions struct {
	GrantType              string
	ListenAddress          []string
	ListenPort             []int // deprecated
	SkipOpenBrowser        bool
	RedirectURLHostname    string
	AuthRequestExtraParams map[string]string
	Username               string
	Password               string
}

// determineListenAddress returns the addresses from the flags.
// Note that --listen-address is always given due to the default value.
// If --listen-port is not set, it returns --listen-address.
// If --listen-port is set, it returns the strings of --listen-port.
func (o *authenticationOptions) determineListenAddress() []string {
	if len(o.ListenPort) == 0 {
		return o.ListenAddress
	}
	var a []string
	for _, p := range o.ListenPort {
		a = append(a, fmt.Sprintf("127.0.0.1:%d", p))
	}
	return a
}

var allGrantType = strings.Join([]string{
	"auto",
	"authcode",
	"authcode-keyboard",
	"password",
}, "|")

func (o *authenticationOptions) addFlags(f *pflag.FlagSet) {
	f.StringVar(&o.GrantType, "grant-type", "auto", fmt.Sprintf("Authorization grant type to use. One of (%s)", allGrantType))
	f.StringSliceVar(&o.ListenAddress, "listen-address", defaultListenAddress, "[authcode] Address to bind to the local server. If multiple addresses are set, it will try binding in order")
	//TODO: remove the deprecated flag
	f.IntSliceVar(&o.ListenPort, "listen-port", nil, "[authcode] deprecated: port to bind to the local server")
	if err := f.MarkDeprecated("listen-port", "use --listen-address instead"); err != nil {
		panic(err)
	}
	f.BoolVar(&o.SkipOpenBrowser, "skip-open-browser", false, "[authcode] Do not open the browser automatically")
	f.StringVar(&o.RedirectURLHostname, "oidc-redirect-url-hostname", "localhost", "[authcode] Hostname of the redirect URL")
	f.StringToStringVar(&o.AuthRequestExtraParams, "oidc-auth-request-extra-params", nil, "[authcode, authcode-keyboard] Extra query parameters to send with an authentication request")
	f.StringVar(&o.Username, "username", "", "[password] Username for resource owner password credentials grant")
	f.StringVar(&o.Password, "password", "", "[password] Password for resource owner password credentials grant")
}

func (o *authenticationOptions) grantOptionSet() (s authentication.GrantOptionSet, err error) {
	switch {
	case o.GrantType == "authcode" || (o.GrantType == "auto" && o.Username == ""):
		s.AuthCodeOption = &authentication.AuthCodeOption{
			BindAddress:            o.determineListenAddress(),
			SkipOpenBrowser:        o.SkipOpenBrowser,
			RedirectURLHostname:    o.RedirectURLHostname,
			AuthRequestExtraParams: o.AuthRequestExtraParams,
		}
	case o.GrantType == "authcode-keyboard":
		s.AuthCodeKeyboardOption = &authentication.AuthCodeKeyboardOption{
			AuthRequestExtraParams: o.AuthRequestExtraParams,
		}
	case o.GrantType == "password" || (o.GrantType == "auto" && o.Username != ""):
		s.ROPCOption = &authentication.ROPCOption{
			Username: o.Username,
			Password: o.Password,
		}
	default:
		err = xerrors.Errorf("grant-type must be one of (%s)", allGrantType)
	}
	return
}
