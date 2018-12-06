
# ldap2slack-user-sync

## purpose

Adds User from a specific LDAP Group to an Slack Group

## debug hints

You can use either Enviroment or command line attributes - but this must be set.

| Environment Var | sample |  
|---|---|
| LDAP_REGION |  e.g. development |   
| LDAP_BIND_USER_CN  |  e.g. CN=user_cn_dummy,OU=Users,DC=%s,DC=COMPANY,DC=com (%s will be filled with LDAP_REGION)|   
| LDAP_BIND_PWD |  well password for LDAP_BIND_USER_CN for Read Request |   
| LDAP_HOST | e.g. ldap host, e.g. ldap.%s.company.com  |   
| LDAP_PORT | 636 |   
| LDAP_SEARCH_CN | LDAP search string  e.g CN=USER_CN,OU=Identities,DC=%s,DC=COMPANY,DC=COM |   
| LDAP_BASE_CN | LDAP entry layer e.g. DC=%s,DC=COMPANY,DC=COM |   


If you use Visual Studio Code, you could use this .launch.json
```
{
    "version": "0.2.0",
    "configurations": [
        {
            "name": "Launch",
            "type": "go",
            "request": "launch",
            "mode": "debug",
            "showLog": false,
            "program": "${file}",
            "env": { 
                "LDAP_BIND_USER_CN":"CN=USER_CN,OU=Identities,DC=%s,DC=COMPANY,DC=com", 
                "LDAP_HOST":"ldap.%s.company.com",
                "LDAP_BIND_USER_CN":"CN=user_cn_dummy,OU=Users,DC=%s,DC=COMPANY,DC=com",
                "LDAP_BIND_PWD":"whateverIsSecret" 
                "LDAP_SEARCH_CN":"CN=USER_CN,OU=Identities,DC=%s,DC=COMPANY,DC=COM",
                "LDAP_BASE_CN":"DC=%s,DC=COMPANY,DC=COM",
            }
        }
    ]
}
```