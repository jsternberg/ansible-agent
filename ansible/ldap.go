package ansible

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"regexp"
	"strconv"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/auth"
	"github.com/mavricknz/ldap"
)

var HostExpr = regexp.MustCompile(`^(ldaps?)://([\w-.]+)(:(\d+))?$`)

type LdapOptions struct {
	Enabled    bool
	Host       string
	Port       uint16
	SSL        bool
	BaseDN     string `toml:"base_dn"`
	UserFilter string `toml:"user_filter"`
}

type ldapConfig struct {
	Host string
	Port uint16
	SSL  bool
}

func LdapAuthenticator(options *LdapOptions) (martini.Handler, error) {
	hostInfo := HostExpr.FindStringSubmatch(options.Host)

	config := &ldapConfig{}
	switch hostInfo[1] {
	case "ldap":
		config.SSL = false
	case "ldaps":
		config.SSL = true
	default:
		return nil, fmt.Errorf("invalid ldap protocol: %s", hostInfo[1])
	}
	config.Host = hostInfo[2]

	if hostInfo[4] != "" {
		port, err := strconv.ParseUint(hostInfo[4], 10, 16)
		if err != nil {
			return nil, fmt.Errorf("unable to parse ldap port: %s", err)
		}
		config.Port = uint16(port)
	} else {
		if config.SSL {
			config.Port = 636
		} else {
			config.Port = 389
		}
	}

	return func(res http.ResponseWriter, req *http.Request, c martini.Context, log *log.Logger) {
		// HACK TODO: do not put routing logic in the auth handler
		// The /ping endpoint does not have auth, so explicitly exclude it here
		if req.URL.Path == "/ping" {
			return
		}

		authHandler := auth.BasicFunc(func(username, password string) bool {
			// create the ldap server connection
			var conn *ldap.LDAPConnection
			if config.SSL {
				tlsConfig := tls.Config{
					ServerName: config.Host,
				}
				conn = ldap.NewLDAPSSLConnection(config.Host, config.Port, &tlsConfig)
			} else {
				conn = ldap.NewLDAPConnection(config.Host, config.Port)
			}

			// attempt to connect to the ldap server
			if err := conn.Connect(); err != nil {
				log.Printf("Unable to connect to LDAP: %s", err)
				return false
			}

			// perform an anonymous search for the user's dn so we can attempt to bind as them
			req := ldap.SearchRequest{
				BaseDN: options.BaseDN,
				Filter: fmt.Sprintf(options.UserFilter, ldap.EscapeFilterValue(username)),
				Scope:  ldap.ScopeWholeSubtree,
			}
			res, err := conn.Search(&req)
			if err != nil {
				log.Printf("Error performing LDAP search: %s", err)
				return false
			}

			// Return false if the number of entries isn't exactly 1.  If multiple
			// results were returned, there is an ambiguity so return false instead
			// of proceeding. If no entries were returned, we have no idea who this
			// is and cannot authenticate.
			if len(res.Entries) != 1 {
				if len(res.Entries) > 1 {
					log.Printf("User '%s' attempted to authenticate but multiple entries exists", username)
				} else {
					log.Printf("User '%s' attempted to authenticate but does not exist", username)
				}
				return false
			}

			dn := res.Entries[0].DN
			if err := conn.Bind(dn, password); err != nil {
				log.Printf("User '%s' attempted to authenticate but provided an invalid password", username)
				return false
			}
			log.Printf("Authenticated successfully as %s", username)
			return true
		})
		authenticate := authHandler.(func(http.ResponseWriter, *http.Request, martini.Context))
		authenticate(res, req, c)
	}, nil
}
