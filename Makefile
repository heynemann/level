PACKAGES = $(shell glide novendor)
DIRS = $(shell find ./ -type f -not -path '*/\.*' | grep '.go' | grep -v "^[.]\/vendor" | xargs -n1 dirname | sort | uniq | grep -v '^.$$')

setup-hooks:
	@cd .git/hooks && ln -sf ../../hooks/pre-commit.sh pre-commit

setup: setup-hooks
	@go get -u github.com/onsi/ginkgo/ginkgo
	@go get -u github.com/Masterminds/glide/...
	@go get -u github.com/nats-io/gnatsd
	@glide install

build:
	@go build $(PACKAGES)

test: test-redis test-gnatsd-shutdown
	@ginkgo --cover $(DIRS)

test-coverage: test
	@rm -rf _build
	@mkdir -p _build
	@echo "mode: count" > _build/test-coverage-all.out
	@bash -c 'for f in $$(find . -name "*.coverprofile"); do tail -n +2 $$f >> _build/test-coverage-all.out; done'

test-coverage-html: test-coverage
	@go tool cover -html=_build/test-coverage-all.out

# get a test redis instance up (localhost:7777)
test-redis: test-redis-shutdown
	@redis-server ./config/test-redis.conf; sleep 1
	@redis-cli -p 7777 info > /dev/null

# shutdown test redis instance (localhost:7777)
test-redis-shutdown:
	@-redis-cli -p 7777 shutdown

# start test gnatsd (localhost:7778)
test-gnatsd: test-gnatsd-shutdown
	@rm -rf /tmp/level-gnatsd.pid
	@gnatsd -p 7778 --pid /tmp/level-gnatsd.pid &

# shutdown test gnatsd
test-gnatsd-shutdown:
	@-cat /tmp/level-gnatsd.pid | xargs kill -9

schema-update: schema-remove
	@easyjson --all messaging/*.go

schema-remove:
	@rm -rf ./messaging/*_easyjson.go
