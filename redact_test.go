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

func TestRedactBasic(t *testing.T) {
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

	cfg, err := NewFrom(input)
	require.NoError(t, err)

	redacted, err := cfg.Redact()
	require.NoError(t, err)
	require.NotNil(t, redacted)

	// Unpack into a map to verify values
	result := make(map[string]interface{})
	err = redacted.Unpack(&result)
	require.NoError(t, err)

	assert.Equal(t, "admin", result["username"])
	assert.Equal(t, "[REDACTED]", result["password"])
	assert.Equal(t, "[REDACTED]", result["api_key"])
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

	cfg, err := NewFrom(input)
	require.NoError(t, err)

	redacted, err := cfg.Redact()
	require.NoError(t, err)
	require.NotNil(t, redacted)

	// Unpack to verify
	var result testConfig
	err = redacted.Unpack(&result)
	require.NoError(t, err)

	assert.Equal(t, "myapp", result.AppName)
	assert.Equal(t, "db.example.com", result.Database.Host)
	assert.Equal(t, "[REDACTED]", result.Database.Password)
	assert.Equal(t, "[REDACTED]", result.APIToken)
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

	cfg, err := NewFrom(input)
	require.NoError(t, err)

	redacted, err := cfg.Redact()
	require.NoError(t, err)
	require.NotNil(t, redacted)

	// Unpack to verify
	var result testConfig
	err = redacted.Unpack(&result)
	require.NoError(t, err)

	assert.Equal(t, "test", result.Name)
	require.Len(t, result.Creds, 2)
	assert.Equal(t, "user1", result.Creds[0].Username)
	assert.Equal(t, "[REDACTED]", result.Creds[0].Password)
	assert.Equal(t, "user2", result.Creds[1].Username)
	assert.Equal(t, "[REDACTED]", result.Creds[1].Password)
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

	redacted, err := cfg.Redact()
	require.NoError(t, err)
	require.NotNil(t, redacted)

	// Unpack to verify nothing changed
	var result testConfig
	err = redacted.Unpack(&result)
	require.NoError(t, err)

	assert.Equal(t, "test", result.Name)
	assert.Equal(t, 42, result.Value)
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
		StringVal  string  `config:"string_val,redact"`
		IntVal     int     `config:"int_val,redact"`
		BoolVal    bool    `config:"bool_val,redact"`
		FloatVal   float64 `config:"float_val,redact"`
		NormalVal  string  `config:"normal_val"`
	}

	input := testConfig{
		StringVal: "secret",
		IntVal:    12345,
		BoolVal:   true,
		FloatVal:  3.14,
		NormalVal: "public",
	}

	cfg, err := NewFrom(input)
	require.NoError(t, err)

	redacted, err := cfg.Redact()
	require.NoError(t, err)
	require.NotNil(t, redacted)

	// Unpack into a map to verify values
	result := make(map[string]interface{})
	err = redacted.Unpack(&result)
	require.NoError(t, err)

	// All redacted fields should be "[REDACTED]" regardless of original type
	assert.Equal(t, "[REDACTED]", result["string_val"])
	assert.Equal(t, "[REDACTED]", result["int_val"])
	assert.Equal(t, "[REDACTED]", result["bool_val"])
	assert.Equal(t, "[REDACTED]", result["float_val"])
	assert.Equal(t, "public", result["normal_val"])
}
