package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/sapcc/ldap2slack-user-sync/pkg/ldap"
	"github.com/sapcc/ldap2slack-user-sync/pkg/slack"
)

const (
	diff string = "diff"
	sync string = "sync"
)

var (
	mode             = flag.String("MODE", "sync", fmt.Sprintf("%s|%s", diff, sync))
	region           = flag.String("LDAP_REGION", os.Getenv("LDAP_REGION"), "the domain")
	ldapBindPassword = flag.String("LDAP_BIND_PWD", os.Getenv("LDAP_BIND_PWD"), "")
	ldapBindUserName = flag.String("LDAP_BIND_USER_CN", os.Getenv("LDAP_BIND_USER_CN"), "e.g. CN=USER_CN,OU=Identities,DC=%s,DC=COMPANY,DC=com")
	ldapSearchCn     = flag.String("LDAP_SEARCH_CN", os.Getenv("LDAP_SEARCH_CN"), "e.g CN=USER_CN,OU=Identities,DC=%s,DC=COMPANY,DC=COM")
	ldapBaseCn       = flag.String("LDAP_BASE_CN", os.Getenv("LDAP_BASE_CN"), "e.g. DC=%s,DC=COMPANY,DC=COM")
	ldapHost         = flag.String("LDAP_HOST", os.Getenv("LDAP_HOST"), "ldap host, e.g. ldap.%s.com")
	ldapPort         = flag.Int("LDAP_PORT", 636, "ldap port - default is default ldaps port")
	ldapCerts        = os.Getenv("LDAP_CERTIFICATES")
	slackAccessToken = flag.String("SLACK_TOKEN", os.Getenv("SLACK_TOKEN"), "SLACK security legacy token (starting with 'xoxp-')")
	slackGroupName   = flag.String("SLACK_LDAP_GROUP", os.Getenv("SLACK_LDAP_GROUP"), "SLACK group name (not ID) where u wanna add the LDAP users to")
	bWrite           = false
)

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

	slackClient := slack.GetASlackClient(*slackAccessToken)
	// get all slack groups
	slckGrps := slack.GetSlackGroups(*slackClient, *slackGroupName, true)

	switch *mode {
	case sync:

		// find members of given group
		ldapUser := ldap.GetLdapUser(*ldapBaseCn, *ldapBindUserName, *ldapHost, *ldapBindPassword, *ldapSearchCn, *ldapPort)

		// get all SLACK users, bcz. we need the SLACK user id and match them with the ldap users
		slckUserFilderedList := slack.GetSlackUser(*slackClient, ldapUser)

		// put ldap users which also have a slack account to our slack group (who's not in the ldap group is out)
		slack.SetSlackGroupUser(*slackClient, slckGrps, *slackGroupName, slckUserFilderedList, bWrite)

	case diff:
		slckUserList := slack.GetSlackUser(*slackClient, nil)
		slack.DiffSlackGroups(slckUserList, slckGrps, *slackGroupName, "CCloud_DevOps")
		slack.DiffSlackGroups(slckUserList, slckGrps, "CCloud_DevOps", *slackGroupName)
		slack.DiffSlackGroups(slckUserList, slckGrps, "Markus_Direct_Reports", *slackGroupName)
	default:

	}

}
