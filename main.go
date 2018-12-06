package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/ldap.v2"
)

var (
	region       = flag.String("region", fmt.Sprintf(os.Getenv("LDAP_REGION")), "the domain")
	bindpassword = os.Getenv("LDAP_BIND_PWD")
	bindusername = flag.String("LDAP_BIND_USER_CN", fmt.Sprintf(os.Getenv("LDAP_BIND_USER_CN"), *region), "e.g. CN=USER_CN,OU=Identities,DC=%s,DC=COMPANY,DC=com")
	searchCn     = flag.String("LDAP_SEARCH_CN", fmt.Sprintf(os.Getenv("LDAP_SEARCH_CN"), *region), "e.g CN=USER_CN,OU=Identities,DC=%s,DC=COMPANY,DC=COM")
	baseCn       = flag.String("LDAP_BASE_CN", fmt.Sprintf(os.Getenv("LDAP_BASE_CN"), *region), "e.g. DC=%s,DC=COMPANY,DC=COM")
	ldapHost     = flag.String("LDAP_HOST", fmt.Sprintf(os.Getenv("LDAP_HOST"), *region), "ldap host, e.g. ldap.%s.com")
	ldapPort     = flag.Int("LDAP_PORT", 636, "ldap port - default is default ldaps port")
)

func main() {
	
	// check env settings
	for _, e := range os.Environ() {
		pair := strings.SplitN(e, "=", 2)
		if strings.Compare(pair[0], "LDAP_BIND_PWD") == 0 {
			fmt.Println(fmt.Sprintf("%s:\t\t *********", pair[0]))
			continue
		}
		if strings.HasPrefix(pair[0], "LDAP") {
			fmt.Println(fmt.Sprintf("%s:\t\t %s", pair[0], pair[1]))
			//fmt.Println(pair)
		}
	}

	fmt.Println(fmt.Sprintf("connect %s @ %s:%d", *bindusername, *ldapHost, *ldapPort))

	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	l, err := ldap.DialTLS("tcp", fmt.Sprintf("%s:%d", *ldapHost, *ldapPort), tlsConfig)

	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	// First bind read user
	err = l.Bind(*bindusername, bindpassword)
	if err != nil {
		log.Fatal(err)
	}

	// Search Request Definition
	searchRequest := ldap.NewSearchRequest(
		*baseCn,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf("(&(objectClass=organizationalPerson)(memberof=%s))", *searchCn),
		[]string{"dn", "sn", "givenName", "cn", "displayName"},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		log.Fatal(err)
	}

	for i, s := range sr.Entries {
		fmt.Println(i, s.DN, "\t", s.GetAttributeValue("cn"), s.GetAttributeValue("displayName"))
		/*for z, a := range s.Attributes {
			fmt.Println(z, a)
		}*/
	}
}
