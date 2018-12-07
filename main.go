package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"gopkg.in/ldap.v2"
	"github.com/nlopes/slack"
)

var (
	region       		= flag.String("region", fmt.Sprintf(os.Getenv("LDAP_REGION")), "the domain")
	ldapBindPassword 	= os.Getenv("LDAP_BIND_PWD") // not allowed via flag
	ldapBindUserName 	= flag.String("LDAP_BIND_USER_CN", fmt.Sprintf(os.Getenv("LDAP_BIND_USER_CN"), *region), "e.g. CN=USER_CN,OU=Identities,DC=%s,DC=COMPANY,DC=com")
	ldapSearchCn     	= flag.String("LDAP_SEARCH_CN", fmt.Sprintf(os.Getenv("LDAP_SEARCH_CN"), *region), "e.g CN=USER_CN,OU=Identities,DC=%s,DC=COMPANY,DC=COM")
	ldapBaseCn       	= flag.String("LDAP_BASE_CN", fmt.Sprintf(os.Getenv("LDAP_BASE_CN"), *region), "e.g. DC=%s,DC=COMPANY,DC=COM")
	ldapHost     		= flag.String("LDAP_HOST", fmt.Sprintf(os.Getenv("LDAP_HOST"), *region), "ldap host, e.g. ldap.%s.com")
	ldapPort     		= flag.Int("LDAP_PORT", 636, "ldap port - default is default ldaps port")

	slackAccessToken     		= flag.String("SLACK_TOKEN", os.Getenv("SLACK_TOKEN"), "SLACK security token")
)

func getLdapUser() []*ldap.Entry {
	fmt.Println(fmt.Sprintf("connect %s @ %s:%d", *ldapBindUserName, *ldapHost, *ldapPort))

	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	l, err := ldap.DialTLS("tcp", fmt.Sprintf("%s:%d", *ldapHost, *ldapPort), tlsConfig)

	if err != nil {
		log.Fatal("LDAP",">",err)
	}
	defer l.Close()

	// First bind read user
	err = l.Bind(*ldapBindUserName, ldapBindPassword)
	if err != nil {
		log.Fatal("LDAP",">",err)
	}

	// Search Request Definition
	searchRequest := ldap.NewSearchRequest(
		*ldapBaseCn,
		ldap.ScopeWholeSubtree,
		ldap.NeverDerefAliases,
		0,
		0,
		false,
		fmt.Sprintf("(&(objectClass=organizationalPerson)(memberof=%s))", *ldapSearchCn),
		[]string{"dn", "sn", "givenName", "cn", "displayName"},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		log.Fatal("LDAP",">",err)
	}

	return sr.Entries
}

func getSlackGroups(){
	api := slack.New(*slackAccessToken)
	// If you set debugging, it will log all requests to the console
	// Useful when encountering issues
	// api.SetDebug(true)
	groups, err := api.GetGroups(false)
	if err != nil {
		log.Fatal("SLACK",">",err)
		return
	}
	for _, group := range groups {
		fmt.Printf("ID: %s, Name: %s\n", group.ID, group.Name)
	}
}

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
		}
	}


	for i, s := range getLdapUser() {
		fmt.Println(i, s.DN, "\t", s.GetAttributeValue("cn"), s.GetAttributeValue("displayName"))
		/*for z, a := range s.Attributes {
			fmt.Println(z, a)
		}*/
	}

	getSlackGroups()

}
