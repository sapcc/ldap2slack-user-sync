package ldap

import (
	"crypto/tls"
	"fmt"

	"github.com/sapcc/ldap2slack-user-sync/pkg/config"
	ldap "gopkg.in/ldap.v2"
	log "github.com/Sirupsen/logrus"
)

// GetLdapUser gets LDAP object by given cn, it's just a wrapper func
func GetLdapUser(cfg *config.LdapConfig) []*ldap.Entry {

	log.Info(fmt.Sprintf("connect %s @ %s:%d", cfg.BindUser, cfg.Host, cfg.Port))

	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	if (len(cfg.Certificates) > 0) {
		c1, _ := tls.X509KeyPair([]byte(cfg.Certificates[0]), []byte(""))
		c2, _ := tls.X509KeyPair([]byte(cfg.Certificates[1]), []byte(""))
		tlsConfig.Certificates = []tls.Certificate{c1, c2}
	}

	l, err := ldap.DialTLS("tcp", fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), tlsConfig)

	if err != nil {
		log.Fatal("LDAP", "> err: ", err)
	}
	defer l.Close()

	// First bind read user
	err = l.Bind(cfg.BindUser, cfg.BindPwd)
	if err != nil {
		log.Fatal("LDAP", ">", err)
	}

	var  searchResult []*ldap.Entry
	for _, searchCN := range cfg.GroupCNs {
		// Search Request Definition
		searchRequest := ldap.NewSearchRequest(
			cfg.BaseCN,
			ldap.ScopeWholeSubtree,
			ldap.NeverDerefAliases,
			0,
			0,
			false,
			fmt.Sprintf("(memberof=%s)", searchCN),
			[]string{"dn", "sn", "givenName", "cn", "displayName"},
			nil,
		)
		sr, err := l.Search(searchRequest)
		if err != nil {
			log.Fatal("LDAP", ">", err)
		}

		if len(sr.Entries) == 0 {
			log.Warn(fmt.Sprintf("Warning: no members in given LDAP Group %s", searchCN))
		}

		searchResult=append(searchResult,sr.Entries...)
	}
	 
	if (log.GetLevel() == log.DebugLevel) {
		for i, s := range searchResult {
			log.Debug(i, s.DN, "\t", s.GetAttributeValue("cn"), s.GetAttributeValue("displayName"))
			//for z, a := range s.Attributes { fmt.Println(z, a) }
		}
	}
	return searchResult
}
