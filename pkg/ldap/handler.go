package ldap

import (
	"crypto/tls"
	"fmt"
	"log"

	ldap "gopkg.in/ldap.v2"
)

// GetLdapUser gets LDAP object by given cn, it's just a wrapper func
func GetLdapUser(ldapBaseCn, ldapBindUserName, ldapHost, ldapBindPassword, ldapSearchCn string, ldapPort int) []*ldap.Entry {

	fmt.Println(fmt.Sprintf("connect %s @ %s:%d", ldapBindUserName, ldapHost, ldapPort))



	tlsConfig := &tls.Config{
		InsecureSkipVerify: true, 
	//	Certificates: []Certificates{ }
	}
	//tls.Certificate
	l, err := ldap.DialTLS("tcp", fmt.Sprintf("%s:%d", ldapHost, ldapPort), tlsConfig)

	if err != nil {
		log.Fatal("LDAP", ">", err)
	}
	defer l.Close()

	// First bind read user
	err = l.Bind(ldapBindUserName, ldapBindPassword)
	if err != nil {
		log.Fatal("LDAP", ">", err)
	}

	// Search Request Definition
	searchRequest := ldap.NewSearchRequest(
		ldapBaseCn,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf("(&(objectClass=organizationalPerson)(memberof=%s))", ldapSearchCn),
		[]string{"dn", "sn", "givenName", "cn", "displayName"},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		log.Fatal("LDAP", ">", err)
	}

	if len(sr.Entries) == 0 {
		log.Print(fmt.Sprintf("Warning: no members in given LDAP Group %s", ldapSearchCn))
	}

	/*
		for i, s := range sr.Entries() {
			fmt.Println(i, s.DN, "\t", s.GetAttributeValue("cn"), s.GetAttributeValue("displayName"))
			//for z, a := range s.Attributes { fmt.Println(z, a) }
		}*/
	return sr.Entries
}
