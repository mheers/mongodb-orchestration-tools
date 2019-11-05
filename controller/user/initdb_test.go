package user

import (
	"testing"

	"github.com/percona/mongodb-orchestration-tools/internal/testutils"
	"github.com/stretchr/testify/assert"
)

func TestInitDBUserCreated(t *testing.T) {
	testutils.DoSkipTest(t)
	assert.Greater(t, len(GetInitDatabases()), 0)
}
