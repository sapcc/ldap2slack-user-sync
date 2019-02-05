/*******************************************************************************
*
* Copyright 2019 SAP SE
*
* Licensed under the Apache License, Version 2.0 (the "License");
* you may not use this file except in compliance with the License.
* You should have received a copy of the License along with this
* program. If not, you may obtain a copy of the License at
*
*     http://www.apache.org/licenses/LICENSE-2.0
*
* Unless required by applicable law or agreed to in writing, software
* distributed under the License is distributed on an "AS IS" BASIS,
* WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
* See the License for the specific language governing permissions and
* limitations under the License.
*
*******************************************************************************/
package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	linq "github.com/ahmetb/go-linq"
	"github.com/nlopes/slack"
	ldap "gopkg.in/ldap.v2"
)

var (
	region           = flag.String("LDAP_REGION", os.Getenv("LDAP_REGION"), "the domain")
	ldapBindPassword = flag.String("LDAP_BIND_PWD", os.Getenv("LDAP_BIND_PWD"), "")
	ldapBindUserName = flag.String("LDAP_BIND_USER_CN", os.Getenv("LDAP_BIND_USER_CN"), "e.g. CN=USER_CN,OU=Identities,DC=%s,DC=COMPANY,DC=com")
	ldapSearchCn     = flag.String("LDAP_SEARCH_CN", os.Getenv("LDAP_SEARCH_CN"), "e.g CN=USER_CN,OU=Identities,DC=%s,DC=COMPANY,DC=COM")
	ldapBaseCn       = flag.String("LDAP_BASE_CN", os.Getenv("LDAP_BASE_CN"), "e.g. DC=%s,DC=COMPANY,DC=COM")
	ldapHost         = flag.String("LDAP_HOST", os.Getenv("LDAP_HOST"), "ldap host, e.g. ldap.%s.com")
	ldapPort         = flag.Int("LDAP_PORT", 636, "ldap port - default is default ldaps port")

	slackAccessToken = flag.String("SLACK_TOKEN", os.Getenv("SLACK_TOKEN"), "SLACK security legacy token (starting with 'xoxp-')")
	slackGroupName   = flag.String("SLACK_LDAP_GROUP", os.Getenv("SLACK_LDAP_GROUP"), "SLACK group name (not ID) where u wanna add the LDAP users to")
	bWrite           = false
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
	err = l.Bind(*ldapBindUserName, *ldapBindPassword)
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

	if len(sr.Entries) == 0 {
		log.Print(fmt.Sprintf("Warning: no members in given LDAP Group %s", *ldapSearchCn))
	}

	/*
		for i, s := range sr.Entries() {
			fmt.Println(i, s.DN, "\t", s.GetAttributeValue("cn"), s.GetAttributeValue("displayName"))
			//for z, a := range s.Attributes { fmt.Println(z, a) }
		}*/
	return sr.Entries
}

func getSlackGroups(slackAPI slack.Client, bWithUser bool) []slack.UserGroup {

	defer func() {
		if r := recover(); r != nil {
			log.Fatal(fmt.Sprintf("SLACK>%s (Searched Group: %s)", r.(error), *slackGroupName))
		}
	}()

	//opt := []slack.GetUserGroupsOption{ slack.GetUserGroupsParams.IncludeUsers : bWithUser}
	slackGroups, err := slackAPI.GetUserGroups(slack.GetUserGroupsOptionIncludeUsers(bWithUser))
	if err != nil {
		log.Fatal("SLACK", ">", err)
		return nil
	}

	for _, group := range slackGroups {
		fmt.Printf("ID: %s, Name: %s, Count: %d (DateDeleted: %s) - %s\n", group.ID, group.Name, group.UserCount, group.DateDelete, group.Description)
	}

	return slackGroups
}

func getSlackUser(slackAPI slack.Client, ldapUsers []*ldap.Entry) []slack.User {

	defer func() {
		if r := recover(); r != nil {
			log.Fatal(fmt.Sprintf("SLACK>%s (findSlackUserFromLdap failed)", r.(error)))
		}
	}()

	// get all SLACK users, bcz. we need the SLACK user id
	slackUsers, err := slackAPI.GetUsers()
	if err != nil {
		log.Fatal("SLACK", ">", err)
		return nil
	}

	// if no LdapsUsers given, we don't need to filder
	if ldapUsers == nil {
		return slackUsers
	}

	// get all SLACK User Ids which are in our LDAP Group
	var ul []slack.User
	linq.From(slackUsers).WhereT(func(u slack.User) bool {
		return (linq.From(ldapUsers).WhereT(func(ldapU *ldap.Entry) bool {
			//fmt.Printf("SlackUser: %s - %s\n", ldapU.GetAttributeValue("cn"), u.Name)
			return (strings.Compare(strings.ToLower(ldapU.GetAttributeValue("cn")), strings.ToLower(u.Name)) == 0)
		}).Count() > 0)
	}).SelectT(func(u slack.User) slack.User {
		//fmt.Printf("SlackUser: %s - %s\n", u.ID, u.Name)
		return u
	}).ToSlice(&ul)

	log.Println(fmt.Printf("%d user in LDAP group | %d in SLACK at all | %d user will be in SLACK group %s\n", len(ldapUsers), len(slackUsers), len(ul), *slackGroupName))

	return ul
}

func setSlackGroupUser(slackAPI slack.Client, slackGroups []slack.UserGroup, slackUser []slack.User) {

	defer func() {
		if r := recover(); r != nil {
			log.Fatal(fmt.Sprintf("SLACK>%s (Searched Group: %s)", r.(error), *slackGroupName))
		}
	}()

	// get the group we are interested in
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

	if len(slackUser) == 0 {
		log.Fatal("SLACK", ">", *slackGroupName, "Given Users List was null, so no update done")
		return
	}

	fmt.Println(fmt.Sprintf("SLACK>TargetGroup.ID: %s [%s]", targetGroup.ID, targetGroup.Name))

	// we need a list of IDs
	var slackUserIds []string
	linq.From(slackUser).SelectT(func(u slack.User) string {
		return u.ID
	}).ToSlice(&slackUserIds)

	if bWrite {
		/*for _, user := range slackUserIds {
			fmt.Printf("ID: %s\n", user)
		}*/
		slackAPI.UpdateUserGroupMembers(targetGroup.ID, strings.Join(slackUserIds, ","))
		log.Println("SLACK>  changes were written!")
	} else {
		log.Println("SLACK> no changes were written, because flag 'bWrite' was set to 'false'")
	}
}

func diffSlackGroups(slackUsers []slack.User, slckGrps []slack.UserGroup, groupA string, groupB string) {

	defer func() {
		if r := recover(); r != nil {
			log.Fatal(fmt.Sprintf("SLACK>%s (Diff for Group: %s & %s failed)", r.(error), groupA, groupB))
		}
	}()

	// find the correct Group Object
	gA := linq.From(slckGrps).WhereT(func(group slack.UserGroup) bool {
		return (strings.Compare(group.Name, groupA) == 0)
	}).First().(slack.UserGroup)
	gB := linq.From(slckGrps).WhereT(func(group slack.UserGroup) bool {
		return (strings.Compare(group.Name, groupB) == 0)
	}).First().(slack.UserGroup)

	i := 0
	linq.From(gA.Users).WhereT(func(uA string) bool {
		return !(linq.From(gB.Users).Contains(uA))
	}).SelectT(func(u string) string {
		return linq.From(slackUsers).WhereT(func(uA slack.User) bool {
			return (strings.Compare(uA.ID, u) == 0)
		}).SelectT(func(uA slack.User) string {
			i++
			return fmt.Sprintf("%d. SlackId: %s SAP ID: %s DisplayName: %s \t is not in %s but in %s", i, uA.ID, uA.RealName, uA.Profile.DisplayName, gB.Name, gA.Name)
		}).First().(string)
	}).ForEach(func(userString interface{}) {
		fmt.Println(userString)
	})
}

func main() {

	flag.Parse()

	if len(*region) == 0 {
		log.Fatal("need flag or environment variable LDAP_REGION")
		return
	}
	if len(*ldapBindUserName) == 0 {
		log.Fatal("need flag or environment variable LDAP_BIND_USER_CN")
		os.Exit(-1)
	} else {
		*ldapBindUserName = fmt.Sprintf(os.Getenv("LDAP_BIND_USER_CN"), *region)
	}
	if len(*ldapSearchCn) == 0 {
		log.Fatal("need flag or environment variable LDAP_SEARCH_CN")
		os.Exit(-1)
	} else {
		*ldapSearchCn = fmt.Sprintf(os.Getenv("LDAP_SEARCH_CN"), *region)
	}
	if len(*ldapBaseCn) == 0 {
		log.Fatal("need flag or environment variable LDAP_BASE_CN")
		os.Exit(-1)
	} else {
		*ldapBaseCn = fmt.Sprintf(os.Getenv("LDAP_BASE_CN"), *region)
	}
	if len(*ldapHost) == 0 {
		log.Fatal("need flag or environment variable LDAP_HOST")
		os.Exit(-1)
	} else {
		*ldapHost = fmt.Sprintf(os.Getenv("LDAP_HOST"), *region)
	}
	if len(*slackAccessToken) == 0 {
		log.Fatal("need flag or environment variable SLACK_TOKEN")
		os.Exit(-1)
	}
	if len(*slackGroupName) == 0 {
		log.Fatal("need flag or environment variable SLACK_LDAP_GROUP")
		os.Exit(-1)
	}
	if len(*ldapBindPassword) == 0 {
		log.Fatal("need flag or environment variable LDAP_BIND_PWD")
		os.Exit(-1)
	}

	// check env settings
	for _, e := range os.Environ() {

		//fmt.Println(e)
		pair := strings.SplitN(e, "=", 2)
		if strings.Compare(pair[0], "LDAP_BIND_PWD") == 0 {
			log.Println(fmt.Sprintf("%s:\t\t *********", pair[0]))
			continue
		}
		if strings.HasPrefix(pair[0], "LDAP") {
			log.Println(fmt.Sprintf("%s:\t\t %s", pair[0], pair[1]))
		}
	}

	slackClient := slack.New(*slackAccessToken, slack.OptionDebug(false))

	// 1. get group by Name
	slckGrps := getSlackGroups(*slackClient, true)
	ldapUser := getLdapUser()

	// 2. get all SLACK users, bcz. we need the SLACK user id
	slckUserFilderedList := getSlackUser(*slackClient, ldapUser)

	slckUserList := getSlackUser(*slackClient, nil)
	diffSlackGroups(slckUserList, slckGrps, *slackGroupName, "CCloud_DevOps")
	diffSlackGroups(slckUserList, slckGrps, "CCloud_DevOps", *slackGroupName)
	diffSlackGroups(slckUserList, slckGrps, "Markus_Direct_Reports", *slackGroupName)

	// 3. add ldap users which are in slack to our slack group
	setSlackGroupUser(*slackClient, slckGrps, slckUserFilderedList)

}
