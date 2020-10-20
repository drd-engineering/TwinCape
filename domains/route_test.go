package domains_test

import (
	"testing"

	"github.com/drd-engineering/TwinCape/domains"
	"github.com/stretchr/testify/assert"
)

func TestInitiateRoute(t *testing.T) {
	assert.NotPanics(t, func() { domains.InitiateRoutes() }, "initiating route should never be panic")
}
