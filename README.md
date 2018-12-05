
# ldap2slack-user-sync

## purpose
tbd

## debug hints

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
                "LDAP_BIND_USER":"${env:LDAP_BIND_USER}", 
                "LDAP_BIND_PWD":"${env:LDAP_BIND_PWD}" 
            }
        }
    ]
}
```