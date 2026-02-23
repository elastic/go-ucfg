// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package ucfg

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAutoRedactOnNewFrom(t *testing.T) {
	type testConfig struct {
		Username string `config:"username"`
		Password string `config:"password,redact"`
		APIKey   string `config:"api_key,redact"`
		Host     string `config:"host"`
	}

	input := testConfig{
		Username: "admin",
		Password: "secret123",
		APIKey:   "key-abc-123",
		Host:     "localhost",
	}

	// Default behavior: redacted fields are automatically replaced
	cfg, err := NewFrom(input)
	require.NoError(t, err)

	// Unpack into a map to verify values
	result := make(map[string]interface{})
	err = cfg.Unpack(&result)
	require.NoError(t, err)

	assert.Equal(t, "admin", result["username"])
	assert.Equal(t, sREDACT, result["password"])
	assert.Equal(t, sREDACT, result["api_key"])
	assert.Equal(t, "localhost", result["host"])
}

func TestRedactWithShowRedactedOption(t *testing.T) {
	type testConfig struct {
		Username string `config:"username"`
		Password string `config:"password,redact"`
		APIKey   string `config:"api_key,redact"`
		Host     string `config:"host"`
	}

	input := testConfig{
		Username: "admin",
		Password: "secret123",
		APIKey:   "key-abc-123",
		Host:     "localhost",
	}

	// With ShowRedacted option: original values are preserved
	cfg, err := NewFrom(input, ShowRedacted)
	require.NoError(t, err)

	// Unpack into a map to verify values
	result := make(map[string]interface{})
	err = cfg.Unpack(&result)
	require.NoError(t, err)

	assert.Equal(t, "admin", result["username"])
	assert.Equal(t, "secret123", result["password"])
	assert.Equal(t, "key-abc-123", result["api_key"])
	assert.Equal(t, "localhost", result["host"])
}

func TestRedactNested(t *testing.T) {
	type database struct {
		Host     string `config:"host"`
		Password string `config:"password,redact"`
	}

	type testConfig struct {
		AppName  string   `config:"app_name"`
		Database database `config:"database"`
		APIToken string   `config:"api_token,redact"`
	}

	input := testConfig{
		AppName: "myapp",
		Database: database{
			Host:     "db.example.com",
			Password: "dbpass123",
		},
		APIToken: "token-xyz-789",
	}

	// Test default behavior (redacted)
	cfg, err := NewFrom(input)
	require.NoError(t, err)

	var result testConfig
	err = cfg.Unpack(&result)
	require.NoError(t, err)

	assert.Equal(t, "myapp", result.AppName)
	assert.Equal(t, "db.example.com", result.Database.Host)
	assert.Equal(t, sREDACT, result.Database.Password)
	assert.Equal(t, sREDACT, result.APIToken)

	// Test with ShowRedacted option
	cfgUnredacted, err := NewFrom(input, ShowRedacted)
	require.NoError(t, err)

	var resultUnredacted testConfig
	err = cfgUnredacted.Unpack(&resultUnredacted)
	require.NoError(t, err)

	assert.Equal(t, "myapp", resultUnredacted.AppName)
	assert.Equal(t, "db.example.com", resultUnredacted.Database.Host)
	assert.Equal(t, "dbpass123", resultUnredacted.Database.Password)
	assert.Equal(t, "token-xyz-789", resultUnredacted.APIToken)
}

func TestRedactArray(t *testing.T) {
	type credentials struct {
		Username string `config:"username"`
		Password string `config:"password,redact"`
	}

	type testConfig struct {
		Name  string        `config:"name"`
		Creds []credentials `config:"credentials"`
	}

	input := testConfig{
		Name: "test",
		Creds: []credentials{
			{Username: "user1", Password: "pass1"},
			{Username: "user2", Password: "pass2"},
		},
	}

	// Test default behavior (redacted)
	cfg, err := NewFrom(input)
	require.NoError(t, err)

	var result testConfig
	err = cfg.Unpack(&result)
	require.NoError(t, err)

	assert.Equal(t, "test", result.Name)
	require.Len(t, result.Creds, 2)
	assert.Equal(t, "user1", result.Creds[0].Username)
	assert.Equal(t, sREDACT, result.Creds[0].Password)
	assert.Equal(t, "user2", result.Creds[1].Username)
	assert.Equal(t, sREDACT, result.Creds[1].Password)

	// Test with ShowRedacted option
	cfgUnredacted, err := NewFrom(input, ShowRedacted)
	require.NoError(t, err)

	var resultUnredacted testConfig
	err = cfgUnredacted.Unpack(&resultUnredacted)
	require.NoError(t, err)

	assert.Equal(t, "test", resultUnredacted.Name)
	require.Len(t, resultUnredacted.Creds, 2)
	assert.Equal(t, "user1", resultUnredacted.Creds[0].Username)
	assert.Equal(t, "pass1", resultUnredacted.Creds[0].Password)
	assert.Equal(t, "user2", resultUnredacted.Creds[1].Username)
	assert.Equal(t, "pass2", resultUnredacted.Creds[1].Password)
}

func TestRedactNoRedactedFields(t *testing.T) {
	type testConfig struct {
		Name  string `config:"name"`
		Value int    `config:"value"`
	}

	input := testConfig{
		Name:  "test",
		Value: 42,
	}

	cfg, err := NewFrom(input)
	require.NoError(t, err)

	// Unpack to verify nothing changed
	var result testConfig
	err = cfg.Unpack(&result)
	require.NoError(t, err)

	assert.Equal(t, "test", result.Name)
	assert.Equal(t, 42, result.Value)
}

func TestRedactMethodStillWorks(t *testing.T) {
	type testConfig struct {
		Username string `config:"username"`
		Password string `config:"password,redact"`
	}

	input := testConfig{
		Username: "admin",
		Password: "secret123",
	}

	// Create config with ShowRedacted option (preserves original values)
	cfg, err := NewFrom(input, ShowRedacted)
	require.NoError(t, err)

	// Verify original values are preserved
	var original testConfig
	err = cfg.Unpack(&original)
	require.NoError(t, err)
	assert.Equal(t, "admin", original.Username)
	assert.Equal(t, "secret123", original.Password)

	// Call Redact() method to get redacted version
	redacted, err := cfg.Redact()
	require.NoError(t, err)
	require.NotNil(t, redacted)

	// Verify redacted values
	var result testConfig
	err = redacted.Unpack(&result)
	require.NoError(t, err)
	assert.Equal(t, "admin", result.Username)
	assert.Equal(t, sREDACT, result.Password)
}

func TestRedactNilConfig(t *testing.T) {
	var cfg *Config
	redacted, err := cfg.Redact()
	assert.Nil(t, redacted)
	assert.Error(t, err)

	// Check if it's an Error type with Reason
	if ucfgErr, ok := err.(Error); ok {
		assert.Equal(t, ErrNilConfig, ucfgErr.Reason())
	}
}

func TestRedactMixedTypes(t *testing.T) {
	type testConfig struct {
		StringVal string  `config:"string_val,redact"`
		IntVal    int     `config:"int_val,redact"`
		BoolVal   bool    `config:"bool_val,redact"`
		FloatVal  float64 `config:"float_val,redact"`
		NormalVal string  `config:"normal_val"`
	}

	input := testConfig{
		StringVal: "secret",
		IntVal:    12345,
		BoolVal:   true,
		FloatVal:  3.14,
		NormalVal: "public",
	}

	// Test default behavior (all redacted)
	cfg, err := NewFrom(input)
	require.NoError(t, err)

	result := make(map[string]interface{})
	err = cfg.Unpack(&result)
	require.NoError(t, err)

	// All redacted fields should be "[REDACTED]" regardless of original type
	assert.Equal(t, sREDACT, result["string_val"])
	assert.Equal(t, sREDACT, result["int_val"])
	assert.Equal(t, sREDACT, result["bool_val"])
	assert.Equal(t, sREDACT, result["float_val"])
	assert.Equal(t, "public", result["normal_val"])

	// Test with ShowRedacted option
	cfgUnredacted, err := NewFrom(input, ShowRedacted)
	require.NoError(t, err)

	resultUnredacted := make(map[string]interface{})
	err = cfgUnredacted.Unpack(&resultUnredacted)
	require.NoError(t, err)

	// Original values should be preserved
	assert.Equal(t, "secret", resultUnredacted["string_val"])
	assert.Equal(t, uint64(12345), resultUnredacted["int_val"])
	assert.Equal(t, true, resultUnredacted["bool_val"])
	assert.Equal(t, 3.14, resultUnredacted["float_val"])
	assert.Equal(t, "public", resultUnredacted["normal_val"])
}

func TestRedactWithInline(t *testing.T) {
	type inline struct {
		Key    string `config:"key"`
		Secret string `config:"secret,redact"`
	}

	type testConfig struct {
		Name   string `config:"name"`
		Inline inline `config:",inline"`
	}

	input := testConfig{
		Name: "test",
		Inline: inline{
			Key:    "public-key",
			Secret: "private-secret",
		},
	}

	// Test default behavior
	cfg, err := NewFrom(input)
	require.NoError(t, err)

	result := make(map[string]interface{})
	err = cfg.Unpack(&result)
	require.NoError(t, err)

	assert.Equal(t, "test", result["name"])
	assert.Equal(t, "public-key", result["key"])
	assert.Equal(t, sREDACT, result["secret"])

	// Test with ShowRedacted option
	cfgUnredacted, err := NewFrom(input, ShowRedacted)
	require.NoError(t, err)

	resultUnredacted := make(map[string]interface{})
	err = cfgUnredacted.Unpack(&resultUnredacted)
	require.NoError(t, err)

	assert.Equal(t, "test", resultUnredacted["name"])
	assert.Equal(t, "public-key", resultUnredacted["key"])
	assert.Equal(t, "private-secret", resultUnredacted["secret"])
}

func TestRedactIdempotent(t *testing.T) {
	type testConfig struct {
		Public string `config:"public"`
		Secret string `config:"secret,redact"`
	}

	input := testConfig{
		Public: "visible",
		Secret: "hidden",
	}

	// Create config with default behavior (values already redacted during NewFrom)
	cfg, err := NewFrom(input)
	require.NoError(t, err)

	// Call Redact() again - should be idempotent since values are already "[REDACTED]"
	redacted, err := cfg.Redact()
	require.NoError(t, err)

	// Both should have the same result
	var result1, result2 map[string]interface{}
	require.NoError(t, cfg.Unpack(&result1))
	require.NoError(t, redacted.Unpack(&result2))

	assert.Equal(t, "visible", result1["public"])
	assert.Equal(t, sREDACT, result1["secret"])
	assert.Equal(t, "visible", result2["public"])
	assert.Equal(t, sREDACT, result2["secret"])
}

func TestRedactMergeOperation(t *testing.T) {
	type testConfig struct {
		Username string `config:"username"`
		Password string `config:"password,redact"`
	}

	input1 := testConfig{
		Username: "admin",
		Password: "secret123",
	}

	// Create base config
	cfg := New()

	// Merge with default behavior (redacted)
	err := cfg.Merge(input1)
	require.NoError(t, err)

	result := make(map[string]interface{})
	err = cfg.Unpack(&result)
	require.NoError(t, err)

	assert.Equal(t, "admin", result["username"])
	assert.Equal(t, sREDACT, result["password"])

	// Create another config and merge with ShowRedacted option
	cfg2 := New()
	err = cfg2.Merge(input1, ShowRedacted)
	require.NoError(t, err)

	result2 := make(map[string]interface{})
	err = cfg2.Unpack(&result2)
	require.NoError(t, err)

	assert.Equal(t, "admin", result2["username"])
	assert.Equal(t, "secret123", result2["password"])
}

