package sessionManager_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"testing"
)

func TestSessionManager(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Extensions/SessionManager Suite")
}
