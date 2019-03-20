package config

import (
	"errors"
	"fmt"
	"io/ioutil"

	"time"
	log "github.com/Sirupsen/logrus"
	yaml "gopkg.in/yaml.v2"
)

// Config we need
type Config struct {
	Slack          SlackConfig   `yaml:"slack"`
	Ldap           LdapConfig    `yaml:"ldap"`
	Default        defaultConfig `yaml:"default"`
	ConfigFilePath string
}

// Options passed via cmd line
type defaultConfig struct {
	// diff or sync
	Mode string `yaml:"mode"`

	// write
	Write bool `yaml:"write"`

	RecheckInterval time.Duration
}

// SlackConfig Struct
type SlackConfig struct {
	// Token to authenticate
	SecurityToken string `yaml:"securityToken"`

	//TargetGroup
	TargetGroup string `yaml:"targetGroup"`

	// Array of groups to extract members which are in and which not
	DiffGroups [][]string `yaml:"diffGroups"`
}

// LdapConfig Struct
type LdapConfig struct {

	// LDAP server
	Host string `yaml:"host"`

	// LDAP port
	Port int

	// LDAP Search User
	BindUser string `yaml:"bindUser"`

	// LDAP Search User Password
	BindPwd string `yaml:"bindPwd"`

	// LDAP Object Search String
	GroupCNs []string `yaml:"groupCNs"`

	// LDAP entry point for searching
	BaseCN string `yaml:"baseCN"`

	// LDAPS certificates
	Certificates []string `yaml:"certificates"`
}

// NewConfig reads the configuration from the given filePath.
func NewConfig(configFilePath string) (cfg Config, err error) {

	if configFilePath == "" {
		return cfg, errors.New("path to configuration file not provided")
	}

	cfgBytes, err := ioutil.ReadFile(configFilePath)
	if err != nil {
		return cfg, fmt.Errorf("read configuration file: %s", err.Error())
	}
	err = yaml.Unmarshal(cfgBytes, &cfg)
	if err != nil {
		return cfg, fmt.Errorf("parse configuration: %s", err.Error())
	}

	if err := cfg.Slack.validate(); err != nil {
		log.Panic("invalid slack configuration", "err", err)
	}

	if err := cfg.Ldap.validate(); err != nil {
		log.Panic("invalid alertmanager configuration", "err", err)
	}

	return cfg, nil
}

func (l *LdapConfig) validate() error {

	if l.Host == "" {
		return errors.New("incomplete ldap configuration: missing messenger `host`")
	}
	if l.BindUser == "" {
		return errors.New("incomplete ldap configuration: missing messenger `bindUser`")
	}
	if l.BindPwd == "" {
		return errors.New("incomplete ldap configuration: missing messenger `bindPwd`")
	}
	if len(l.GroupCNs) < 1 {
		return errors.New("incomplete ldap configuration: missing messenger `groupCNs`")
	}
	if l.BaseCN == "" {
		return errors.New("incomplete ldap configuration: missing messenger `baseCN`")
	}
	//if strings.HasPrefix(strings.ToLower(s.host), "ldaps://")
	/*if l.Certificates == "" {
		return errors.New("incomplete ldap configuration: missing messenger `securityToken`")
	}*/
	return nil
}

func (s *SlackConfig) validate() error {

	if s.SecurityToken == "" {
		return errors.New("incomplete slack configuration: missing messenger `securityToken`")
	}

	if s.TargetGroup == "" {
		return errors.New("incomplete slack configuration: missing messenger `targetGroup`")
	}

	return nil
}
