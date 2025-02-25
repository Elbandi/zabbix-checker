module certbot-checker

go 1.23.5

require (
	github.com/urfave/cli/v2 v2.27.5
	golang.zabbix.com/agent2 v0.0.0-20250225074525-c34078a4563f
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.5 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/xrash/smetrics v0.0.0-20240521201337-686a1a2994c1 // indirect
	golang.zabbix.com/sdk v1.2.2-0.20250214072554-abd5e97e6797 // indirect
)

replace golang.zabbix.com/agent2 v0.0.0-20250225074525-c34078a4563f => github.com/zabbix/zabbix/src/go v0.0.0-20250225074525-c34078a4563f
