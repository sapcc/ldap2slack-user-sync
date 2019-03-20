package main

import (
	"flag"
	"fmt"

	"github.com/sapcc/ldap2slack-user-sync/pkg/ldap"
	"github.com/sapcc/ldap2slack-user-sync/pkg/slack"
	"github.com/sapcc/ldap2slack-user-sync/pkg/config"
	log "github.com/Sirupsen/logrus"
)

var opts config.Config

const (
	slackdiff string = "slackdiff"
	sync      string = "sync"
)

func main() {
	log.SetFormatter(&log.JSONFormatter{})
    log.SetLevel(log.DebugLevel)

	flag.StringVar(&opts.ConfigFilePath, "config", "./config.yml", "config file path including filen name")
	flag.StringVar(&opts.Default.Mode, "mode", sync, fmt.Sprintf("%s|%s", slackdiff, sync))
	flag.BoolVar(&opts.Default.Write, "write", false, "(true|false) write changes?")
	flag.Parse()

	cfg, err := config.NewConfig(opts.ConfigFilePath)

	if err != nil {
		log.Panic(err)
	}


	slackClient := slack.GetASlackClient(cfg.Slack.SecurityToken)

	slack.SendMessage(*slackClient, "")

	// get all slack groups
	slckGrps, err := slack.GetSlackGroups(*slackClient, cfg.Slack.TargetGroup, true)
	if err != nil {
		log.Panic(err)
	}
	
	switch cfg.Default.Mode {
	case sync:

		// find members of given group
		ldapUser := ldap.GetLdapUser(&cfg.Ldap)

		// get all SLACK users, bcz. we need the SLACK user id and match them with the ldap users
		slckUserFilderedList := slack.GetSlackUser(*slackClient, ldapUser)

		// put ldap users which also have a slack account to our slack group (who's not in the ldap group is out)
		slack.SetSlackGroupUser(*slackClient, slckGrps, cfg.Slack.TargetGroup, slckUserFilderedList, cfg.Default.Write)

	case slackdiff:
		slckUserList := slack.GetSlackUser(*slackClient, nil)
		for _, g := range cfg.Slack.DiffGroups {
			slack.DiffSlackGroups(slckUserList, slckGrps, g[0], g[1])
		}

	default:
		// show usage
	}

}
