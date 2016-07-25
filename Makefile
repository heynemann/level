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

dependencies deps: gnatsd redis

# get a redis instance up (localhost:8787)
redis: redis-shutdown
	@redis-server ./config/redis.conf; sleep 1
	@redis-cli -p 8787 info > /dev/null
	@echo "redis running at localhost:8787."

# shutdown test redis instance (localhost:7777)
redis-shutdown:
	@-redis-cli -p 8787 shutdown

# start test gnatsd (localhost:7778)
gnatsd: gnatsd-shutdown
	@rm -rf /tmp/level-gnatsd.pid
	@gnatsd -p 8788 --pid /tmp/level-gnatsd.pid &

# shutdown test gnatsd
gnatsd-shutdown:
	@-cat /tmp/level-gnatsd.pid | xargs kill -9

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
	@echo "test-redis running at localhost:7777."

# shutdown test redis instance (localhost:7777)
test-redis-shutdown:
	@-redis-cli -p 7777 shutdown

# start test gnatsd (localhost:7778)
test-gnatsd: test-gnatsd-shutdown
	@rm -rf /tmp/level-gnatsd.pid
	@gnatsd -p 7778 --pid /tmp/level-test-gnatsd.pid &

# shutdown test gnatsd
test-gnatsd-shutdown:
	@-cat /tmp/level-test-gnatsd.pid | xargs kill -9

schema-update: schema-remove
	@easyjson --all messaging/*.go

schema-remove:
	@rm -rf ./messaging/*_easyjson.go
