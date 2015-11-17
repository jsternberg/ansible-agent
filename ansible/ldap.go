package ansible

import (
	"fmt"

	"github.com/go-martini/martini"
	"github.com/martini-contrib/auth"
	"github.com/mavricknz/ldap"
)

func LdapAuthenticator() martini.Handler {
	return auth.BasicFunc(func(username, password string) bool {
		// connect to the ldap server
		conn := ldap.NewLDAPConnection("rodc.knewton.net", 389)
		if err := conn.Connect(); err != nil {
			fmt.Println(err)
			return false
		}

		// perform an anonymous search for the user's dn so we can attempt to bind as them
		req := ldap.SearchRequest{
			BaseDN: "dc=knewton,dc=net",
			Filter: fmt.Sprintf("(uid=%s)", ldap.EscapeFilterValue(username)),
			Scope:  ldap.ScopeWholeSubtree,
		}
		res, err := conn.Search(&req)
		if err != nil {
			fmt.Println(err)
			return false
		}

		// Return false if the number of entries isn't exactly 1.  If multiple
		// results were returned, there is an ambiguity so return false instead
		// of proceeding. If no entries were returned, we have no idea who this
		// is and cannot authenticate.
		if len(res.Entries) != 1 {
			if len(res.Entries) > 1 {
				fmt.Printf("User '%s' attempted to authenticate but multiple entries exists", username)
			} else {
				fmt.Printf("User '%s' attempted to authenticate but does not exist", username)
			}
			return false
		}

		dn := res.Entries[0].DN
		if err := conn.Bind(dn, password); err != nil {
			fmt.Printf("User '%s' attempted to authenticate but provided an invalid password", username)
			return false
		}
		return true
	})
}
