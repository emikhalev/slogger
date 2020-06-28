test:
	go test ./... --race --count=1

test-fast:
	go test ./... --count=1

test-integration:
	docker pull rsyslog/syslog_appliance_alpine
	go test --tags=integration ./... --count=1