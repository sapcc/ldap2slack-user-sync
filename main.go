package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ahmetb/go-linq"
	"github.com/nlopes/slack"
	"gopkg.in/ldap.v2"
)

var (
	region           = flag.String("region", fmt.Sprintf(os.Getenv("LDAP_REGION")), "the domain")
	ldapBindPassword = os.Getenv("LDAP_BIND_PWD") // not allowed via flag
	ldapBindUserName = flag.String("LDAP_BIND_USER_CN", fmt.Sprintf(os.Getenv("LDAP_BIND_USER_CN"), *region), "e.g. CN=USER_CN,OU=Identities,DC=%s,DC=COMPANY,DC=com")
	ldapSearchCn     = flag.String("LDAP_SEARCH_CN", fmt.Sprintf(os.Getenv("LDAP_SEARCH_CN"), *region), "e.g CN=USER_CN,OU=Identities,DC=%s,DC=COMPANY,DC=COM")
	ldapBaseCn       = flag.String("LDAP_BASE_CN", fmt.Sprintf(os.Getenv("LDAP_BASE_CN"), *region), "e.g. DC=%s,DC=COMPANY,DC=COM")
	ldapHost         = flag.String("LDAP_HOST", fmt.Sprintf(os.Getenv("LDAP_HOST"), *region), "ldap host, e.g. ldap.%s.com")
	ldapPort         = flag.Int("LDAP_PORT", 636, "ldap port - default is default ldaps port")

	slackAccessToken = flag.String("SLACK_TOKEN", os.Getenv("SLACK_TOKEN"), "SLACK security legacy token (starting with 'xoxp-')")
	slackGroupName   = flag.String("SLACK_LDAP_GROUP", os.Getenv("SLACK_LDAP_GROUP"), "SLACK group name (not ID) where u wanna add the LDAP users to")
)

func getLdapUser() []*ldap.Entry {
	fmt.Println(fmt.Sprintf("connect %s @ %s:%d", *ldapBindUserName, *ldapHost, *ldapPort))

	tlsConfig := &tls.Config{InsecureSkipVerify: true}
	l, err := ldap.DialTLS("tcp", fmt.Sprintf("%s:%d", *ldapHost, *ldapPort), tlsConfig)

	if err != nil {
		log.Fatal("LDAP", ">", err)
	}
	defer l.Close()

	// First bind read user
	err = l.Bind(*ldapBindUserName, ldapBindPassword)
	if err != nil {
		log.Fatal("LDAP", ">", err)
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
		log.Fatal("LDAP", ">", err)
	}

	return sr.Entries
}

func getSlackGroups(ldapUsers []*ldap.Entry) {

	defer func() {
		if r := recover(); r != nil {
		   log.Fatal(fmt.Sprintf("SLACK>%s (Searched Group: %s)",r.(error), *slackGroupName))
		}
	 }()

	api := slack.New(*slackAccessToken, slack.OptionDebug(false)) 

// 1. get group by Name
	slackGroups, err := api.GetUserGroups()
	if err != nil {
		log.Fatal("SLACK", ">", err)
		return
	}

	for _, group := range slackGroups {
		fmt.Printf("ID: %s, Name: %s, Count: %d (DateDeleted: %s) - %s\n", group.ID, group.Name, group.UserCount, group.DateDelete, group.Description)
	}

	q := linq.From(slackGroups).WhereT(func(group slack.UserGroup) bool {
		return (strings.Compare(group.Name, *slackGroupName) == 0)
	}).First()

	var targetGroup slack.UserGroup 
	if q != nil {
		targetGroup = q.(slack.UserGroup) 
	} else {
		log.Fatal("SLACK", ">", *slackGroupName, " wasn't there @SLACK - check config!")
		return
	}

	fmt.Println(fmt.Sprintf("SLACK>TargetGroup.ID: %s [%s]", targetGroup.ID, targetGroup.Name))

// 2. get all SLACK users, bcz. we need the SLACK user id
	slackUsers, err := api.GetUsers()
	
	/*for _, user := range slackUsers {
		fmt.Printf("ID: %s, Name: %s, RealName: %s (deleted: %t) - %s - %s\n", user.ID, user.Name, user.RealName, user.Deleted, user.Profile.DisplayName, user.Profile.Email)
	}*/

	// get all SLACK User Ids which are 
	var ul []string
	linq.From(slackUsers).WhereT(func(u slack.User) bool {
		return (linq.From(ldapUsers).WhereT(func(ldapU *ldap.Entry) bool {
			//fmt.Printf("SlackUser: %s - %s\n", ldapU.GetAttributeValue("cn"), u.Name)
			return (strings.Compare(strings.ToLower(ldapU.GetAttributeValue("cn")), strings.ToLower(u.Name)) == 0)
		}).Count() > 0)
	}).SelectT(func(u slack.User) string {
		//fmt.Printf("SlackUser: %s - %s\n", u.ID, u.Name)
		return u.ID
	}).ToSlice(&ul)
/*
	for _, user := range ul {
		fmt.Printf("ID: %s\n", user)
	}
*/
	// 3. add ldap users which are in slack to our group
	api.UpdateUserGroupMembers(targetGroup.ID,strings.Join(ul,","))
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

	/*
		for i, s := range getLdapUser() {
			fmt.Println(i, s.DN, "\t", s.GetAttributeValue("cn"), s.GetAttributeValue("displayName"))
			/*for z, a := range s.Attributes {
				fmt.Println(z, a)
			}*
		}
	*/

	getSlackGroups(getLdapUser())

}
