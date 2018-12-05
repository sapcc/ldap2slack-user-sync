package main

import (
	"flag"
	"os"
	"gopkg.in/ldap.v2"
	"crypto/tls"
	"fmt"
	"log"
	"strings"
)

var (
	region   	  = flag.String("region", "staging", "The domain for the uniform distribution.")

	user		  = os.Getenv("LDAP_BIND_USER")
	bindpassword  = os.Getenv("LDAP_BIND_PWD")
	bindusername  = fmt.Sprintf("CN=%s,OU=Identities,DC=ad,DC=%s,DC=cloud,DC=sap", user, *region)
	searchCn 	  = fmt.Sprintf("CN=CP_CONV_SLACK_ACK,OU=Permissions,OU=CCloud,DC=ad,DC=%s,DC=cloud,DC=sap", *region)
	baseCn 	 	  = fmt.Sprintf("DC=ad,DC=%s,DC=cloud,DC=sap", *region)
	ldapHost 	  = fmt.Sprintf("ldap.%s.cloud.sap", *region)
	ldapPort 	  = 636
)

func main() {

    for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		fmt.Println(pair)
        //fmt.Println(pair[0],pair[1],pair)
    }

	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	l, err := ldap.DialTLS("tcp", fmt.Sprintf("%s:%d", ldapHost, ldapPort), tlsConfig)

	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()

	// First bind read user
	err = l.Bind(bindusername, bindpassword)
	if err != nil {
		log.Fatal(err)
	}

	// Search Request Definition
	searchRequest := ldap.NewSearchRequest(
		baseCn,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf("(&(objectClass=organizationalPerson)(memberof=%s))", searchCn),
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
