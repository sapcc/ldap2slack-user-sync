package slack

import (
	"fmt"
	
	"strings"
	log "github.com/Sirupsen/logrus"
	linq "github.com/ahmetb/go-linq"
	slack "github.com/nlopes/slack"
	ldap "gopkg.in/ldap.v2"
)

// GetASlackClient provides a Slack Client Instance
func GetASlackClient(slackAccessToken string) *slack.Client {
	return slack.New(slackAccessToken, slack.OptionDebug(false))
}

// GetSlackGroups requests existing Groups
func GetSlackGroups(slackAPI slack.Client, slackGroupName string, bWithUser bool) ([]slack.UserGroup, error) {

	slackGroups, err := slackAPI.GetUserGroups(slack.GetUserGroupsOptionIncludeUsers(bWithUser))
	if err != nil {
		return slackGroups, fmt.Errorf(fmt.Sprintf("SLACK > %s", err))
	}

	for _, group := range slackGroups {
		log.Printf("SLACK Group: ID: %s, Name: %s, Count: %d (DateDeleted: %s) - %s\n", group.ID, group.Name, group.UserCount, group.DateDelete, group.Description)
	}
	return slackGroups, nil
}

// GetSlackUser delivers Slack User
func GetSlackUser(slackAPI slack.Client, ldapUsers []*ldap.Entry) []slack.User {

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
			// check on SAP ID and only not deleted Slack User
			return (strings.Compare(strings.ToLower(ldapU.GetAttributeValue("cn")), strings.ToLower(u.Name)) == 0 && !u.Deleted)
		}).Count() > 0)
	}).SelectT(func(u slack.User) slack.User {
		if (log.GetLevel() == log.DebugLevel) {
			fmt.Printf("SlackUser: %s - %s (%s) - %t\n", u.ID, u.Name, u.Profile.DisplayName, u.Deleted)
		}
		return u
	}).ToSlice(&ul)

	log.Println(fmt.Printf("%d user in LDAP group | %d in SLACK at all | %d user will be in SLACK group\n", len(ldapUsers), len(slackUsers), len(ul)))

	return ul
}

// SetSlackGroupUser sets an array of Slack User to an Slack Group (found by name)
func SetSlackGroupUser(slackAPI slack.Client, slackGroups []slack.UserGroup, slackGroupName string, slackUser []slack.User, bWrite bool) {

	defer func() {
		if r := recover(); r != nil {
			log.Fatal(fmt.Sprintf("SLACK>%s (Searched Group: %s)", r.(error), slackGroupName))
		}
	}()

	// get the group we are interested in
	q := linq.From(slackGroups).WhereT(func(group slack.UserGroup) bool {
		return (strings.Compare(group.Name, slackGroupName) == 0)
	}).First()

	var targetGroup slack.UserGroup
	if q != nil {
		targetGroup = q.(slack.UserGroup)
	} else {
		log.Fatal("SLACK", ">", slackGroupName, " wasn't there @SLACK - check config!")
		return
	}

	if len(slackUser) == 0 {
		log.Fatal("SLACK", ">", slackGroupName, "Given Users List was null, so no update done")
		return
	}

	fmt.Println(fmt.Sprintf("SLACK>TargetGroup.ID: %s [%s]", targetGroup.ID, targetGroup.Name))

	// we need a list of IDs
	var slackUserIds []string
	linq.From(slackUser).SelectT(func(u slack.User) string {
		return u.ID
	}).Distinct().ToSlice(&slackUserIds)

	if bWrite {
		for i, user := range slackUserIds {
			fmt.Printf("%d. ID: %s\n", i, user)
		}
		
		log.Debug(strings.Join(slackUserIds, ","))

		usergroup, err := slackAPI.UpdateUserGroupMembers(targetGroup.ID, strings.Join(slackUserIds, ","))
		if err != nil {
			log.Fatal("SLACK", "> SLACK error: ", err)
		} else {
			log.Println("SLACK> changes were written!", usergroup)
		}
	} else {
		log.Println("SLACK> no changes were written, because flag 'bWrite' was set to 'false'")
	}
}

// SendMessage sends a message 
func SendMessage(slackAPI slack.Client, channel string) {

	//slackAPI.SendMessage("")

	attachment := slack.Attachment{
		Pretext: "some pretext",
		Text:    "some text",
		// Uncomment the following part to send a field too
		Fields: []slack.AttachmentField{
			slack.AttachmentField{
				Title: "a",
				Value: "no",
			},
		},
		
	}

	channelID, timestamp, err := slackAPI.PostMessage("C6LUQKM5K", slack.MsgOptionText("Some text", false), slack.MsgOptionAttachments(attachment))
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	fmt.Printf("Message successfully sent to channel %s at %s", channelID, timestamp)

}


// DiffSlackGroups does a print out on who is in Slack Group A vs B 
func DiffSlackGroups(slackUsers []slack.User, slckGrps []slack.UserGroup, groupA string, groupB string) {

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

