// Copyright © 2019 Banzai Cloud
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/goph/logur"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewLogger(t *testing.T) {
	logger := NewLogger(Config{Format: "json", Level: "info", NoColor: true})
	require.NotNil(t, logger)

	logger = NewLogger(Config{Format: "logfmt", Level: "debug"})
	require.NotNil(t, logger)

	logger = NewLogger(Config{Format: "json", Level: "not-a-level"})
	require.NotNil(t, logger)
}

func TestWithFields(t *testing.T) {
	base := logur.NewTestLogger()
	logger := WithFields(base, map[string]interface{}{"k": "v"})
	logger.Info("hello")

	require.Equal(t, 1, base.Count())
	assert.Equal(t, "v", base.LastEvent().Fields["k"])
}

func TestWithFieldsForHandlers(t *testing.T) {
	base := logur.NewTestLogger()

	t.Run("without correlation id", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

		logger := WithFieldsForHandlers(c, base, map[string]interface{}{"provider": "google"})
		logger.Info("msg")

		event := base.LastEvent()
		require.NotNil(t, event)
		assert.Equal(t, "google", event.Fields["provider"])
		_, hasCID := event.Fields[correlationIdField]
		assert.False(t, hasCID)
	})

	t.Run("with correlation id and nil fields", func(t *testing.T) {
		c, _ := gin.CreateTestContext(httptest.NewRecorder())
		c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
		c.Set(ContextKey, "cid-xyz")

		logger := WithFieldsForHandlers(c, base, nil)
		logger.Info("msg")

		event := base.LastEvent()
		require.NotNil(t, event)
		assert.Equal(t, "cid-xyz", event.Fields[correlationIdField])
	})
}
