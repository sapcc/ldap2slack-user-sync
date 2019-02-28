package config

import (
	"fmt"
	"io/ioutil"
	"time"
	"gopkg.in/yaml.v2"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

// Config we need
type Config struct {

	Slack   slackConfig   `yaml:"slack"`
	Ldap    ldapConfig    `yaml:"ldap"`

	ListenPort  int
	ExternalURL string

	ConfigFilePath string
	SecretFilePath string
}

type slackConfig struct {
	AuthorizedGroups []string `yaml:"authorized_groups"`

	// AccessToken to authenticate the stargate to messenger.
	AccessToken string `yaml:"access_token"`

	// BotUserAccessToken is the access token used by the bot.
	BotUserAccessToken string `yaml:"bot_user_access_token"`

	// SigningSecret to verify slack messages.
	SigningSecret string `yaml:"signing_secret"`

	// VerificationToken to verify slack messages.
	VerificationToken string `yaml:"verification_token"`

	// UserName for slack messages.
	UserName string `yaml:"user_name"`

	// UserIcon for slack messages.
	UserIcon string `yaml:"user_icon"`

	// Command to trigger actions.
	Command string `yaml:"command"`

	// RecheckInterval for user group memberships.
	RecheckInterval time.Duration `yaml:"recheck_interval"`

	// IsDisableRTM allows disabeling the slack RTM (real time messaging).
	IsDisableRTM bool `yaml:"-"`
}
type ldapConfig struct {
	// 
	AuthToken string `yaml:"auth_token"`

	// 
	DefaultUserEmail string `yaml:"default_user_email"`
}

// NewConfig reads the configuration from the given filePath.
func NewConfig(opts Options) (cfg Config, err error) {
	if opts.ConfigFilePath == "" {
		return cfg,  errors.New("path to configuration file not provided")
	}

	cfgBytes, err := ioutil.ReadFile(opts.ConfigFilePath)
	if err != nil {
		return cfg, fmt.Errorf("read configuration file: %s", err.Error())
	}
	err = yaml.Unmarshal(cfgBytes, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("parse configuration: %s", err.Error())
	}

	if opts.ExternalURL != "" {
		cfg.ExternalURL = opts.ExternalURL
	}
	if opts.ListenPort != 0 {
		cfg.ListenPort = opts.ListenPort
	}
	cfg.Slack.IsDisableRTM = opts.IsDisableSlackRTM

	if err := cfg.Slack.validate(); err != nil {
		log.Logger.Log. .LogFatal("invalid slack configuration", "err", err)
	}

	if err := cfg.AlertManager.validate(); err != nil {
		logger.LogFatal("invalid alertmanager configuration", "err", err)
	}

	return cfg, nil
}