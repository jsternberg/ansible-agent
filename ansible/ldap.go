package ansible

import (
	"crypto/tls"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/auth"
	"github.com/mavricknz/ldap"
)

type LdapOptions struct {
	Enabled    bool
	Host       string
	Port       uint16
	SSL        bool
	BaseDN     string `toml:"base_dn"`
	UserFilter string `toml:"user_filter"`
}

func LdapAuthenticator(options *LdapOptions) martini.Handler {
	if strings.HasPrefix(options.Host, "ldaps://") {
		options.Host = options.Host[8:]
		options.SSL = true
	} else if strings.HasPrefix(options.Host, "ldap://") {
		options.Host = options.Host[7:]
		options.SSL = false
	}

	if options.Port == 0 {
		if options.SSL {
			options.Port = 636
		} else {
			options.Port = 389
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
			if options.SSL {
				tlsConfig := tls.Config{
					ServerName: options.Host,
				}
				conn = ldap.NewLDAPSSLConnection(options.Host, options.Port, &tlsConfig)
			} else {
				conn = ldap.NewLDAPConnection(options.Host, options.Port)
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
	}
}
