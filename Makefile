PACKAGES = $(shell glide novendor)
DIRS = $(shell find ./ -type f -not -path '*/\.*' | grep '.go' | grep -v "^[.]\/vendor" | xargs dirname | sort | uniq | grep -v '^.$$')

setup-hooks:
	@cd .git/hooks && ln -sf ../../hooks/pre-commit.sh pre-commit

setup: setup-hooks
	@go get -u github.com/Masterminds/glide/...
	@glide install

test: test-redis
	@ginkgo --cover $(DIRS)

test-coverage: test
	@echo "mode: count" > test-coverage-all.out
	@bash -c 'for f in $$(find . -name "*.coverprofile"); do tail -n +2 $$f >> test-coverage-all.out; done'
	@go tool cover -html=test-coverage-all.out

# get a test redis instance up (localhost:7777)
test-redis: test-redis-shutdown
	@redis-server ./config/test-redis.conf; sleep 1
	@redis-cli -p 7777 info > /dev/null

# shutdown test redis instance (localhost:7777)
test-redis-shutdown:
	@-redis-cli -p 7777 shutdown
