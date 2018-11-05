package nfsv3driver

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"time"

	"fmt"

	"code.cloudfoundry.org/dockerdriver"
	"code.cloudfoundry.org/goshims/ldapshim"
	"gopkg.in/ldap.v2"
)

//go:generate counterfeiter -o nfsdriverfakes/fake_id_resolver.go . IdResolver
type IdResolver interface {
	Resolve(env dockerdriver.Env, username string, password string) (uid string, gid string, err error)
}

type ldapIdResolver struct {
	svcUser     string
	svcPass     string
	ldapHost    string
	ldapPort    int
	ldapProto   string
	ldapFqdn    string // ldap domain to search for users .in, e.g. "cn=Users,dc=corp,dc=persi,dc=cf-app,dc=com"
	ldapCACert  string
	ldap        ldapshim.Ldap
	ldapTimeout time.Duration
}

func NewLdapIdResolver(
	svcUser string,
	svcPass string,
	ldapHost string,
	ldapPort int,
	ldapProto string,
	ldapFqdn string,
	ldapCACert string,
	ldap ldapshim.Ldap,
	ldapTimeout time.Duration,
) IdResolver {
	return &ldapIdResolver{
		svcUser:     svcUser,
		svcPass:     svcPass,
		ldapHost:    ldapHost,
		ldapPort:    ldapPort,
		ldapProto:   ldapProto,
		ldapFqdn:    ldapFqdn,
		ldapCACert:  ldapCACert,
		ldap:        ldap,
		ldapTimeout: ldapTimeout,
	}
}

func (d *ldapIdResolver) Resolve(env dockerdriver.Env, username string, password string) (uid string, gid string, err error) {
	addr := fmt.Sprintf("%s:%d", d.ldapHost, d.ldapPort)

	var l ldapshim.LdapConnection
	if d.ldapCACert != "" {
		roots := x509.NewCertPool()
		ok := roots.AppendCertsFromPEM([]byte(d.ldapCACert))
		if !ok {
			return "", "", errors.New("Failed to load CA certificate")
		}

		l, err = d.ldap.DialTLS(d.ldapProto, addr, &tls.Config{
			ServerName: d.ldapHost,
			RootCAs:    roots,
		})
	} else {
		l, err = d.ldap.Dial(d.ldapProto, addr)
	}
	if err != nil {
		return "", "", err
	}

	l.SetTimeout(d.ldapTimeout)
	defer l.Close()

	// First bind with a read only user
	err = l.Bind(d.svcUser, d.svcPass)
	if err != nil {
		return "", "", err
	}

	// Search for the given username
	searchRequest := d.ldap.NewSearchRequest(
		d.ldapFqdn,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		fmt.Sprintf("(&(objectClass=User)(cn=%s))", username),
		[]string{"dn", "uidNumber", "gidNumber"},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		return "", "", err
	}

	if len(sr.Entries) == 0 {
		return "", "", errors.New("User does not exist")
	}
	if len(sr.Entries) > 1 {
		return "", "", errors.New("Ambiguous search--too many results")
	}

	userdn := sr.Entries[0].DN

	uid = sr.Entries[0].GetAttributeValue("uidNumber")
	gid = sr.Entries[0].GetAttributeValue("gidNumber")
	if gid == "" {
		gid = uid
	}

	// Bind as the user to verify their password
	err = l.Bind(userdn, password)
	if err != nil {
		return "", "", err
	}

	return uid, gid, nil
}
