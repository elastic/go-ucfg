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

func TestDefaultUnpackRedactsFields(t *testing.T) {
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

	// Config stores original values (always)
	cfg, err := NewFrom(input)
	require.NoError(t, err)

	// Default Unpack behavior: redacted fields are replaced with "[REDACTED]"
	result := make(map[string]interface{})
	err = cfg.Unpack(&result)
	require.NoError(t, err)

	assert.Equal(t, "admin", result["username"])
	assert.Equal(t, sREDACT, result["password"])
	assert.Equal(t, sREDACT, result["api_key"])
	assert.Equal(t, "localhost", result["host"])
}

func TestUnpackWithShowRedactedOption(t *testing.T) {
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

	// Config stores original values (always)
	cfg, err := NewFrom(input)
	require.NoError(t, err)

	// Unpack with ShowRedacted option: original values are shown
	result := make(map[string]interface{})
	err = cfg.Unpack(&result, ShowRedacted)
	require.NoError(t, err)

	assert.Equal(t, "admin", result["username"])
	assert.Equal(t, "secret123", result["password"])
	assert.Equal(t, "key-abc-123", result["api_key"])
	assert.Equal(t, "localhost", result["host"])
}

func TestUnpackRedactsNestedStructs(t *testing.T) {
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

	// Test default behavior (redacted during Unpack)
	cfg, err := NewFrom(input)
	require.NoError(t, err)

	var result testConfig
	err = cfg.Unpack(&result)
	require.NoError(t, err)

	assert.Equal(t, "myapp", result.AppName)
	assert.Equal(t, "db.example.com", result.Database.Host)
	assert.Equal(t, sREDACT, result.Database.Password)
	assert.Equal(t, sREDACT, result.APIToken)

	// Test with ShowRedacted option (unredacted during Unpack)
	var resultUnredacted testConfig
	err = cfg.Unpack(&resultUnredacted, ShowRedacted)
	require.NoError(t, err)

	assert.Equal(t, "myapp", resultUnredacted.AppName)
	assert.Equal(t, "db.example.com", resultUnredacted.Database.Host)
	assert.Equal(t, "dbpass123", resultUnredacted.Database.Password)
	assert.Equal(t, "token-xyz-789", resultUnredacted.APIToken)
}

func TestUnpackRedactsArrayElements(t *testing.T) {
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

	// Test default behavior (redacted during Unpack)
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

	// Test with ShowRedacted option (unredacted during Unpack)
	var resultUnredacted testConfig
	err = cfg.Unpack(&resultUnredacted, ShowRedacted)
	require.NoError(t, err)

	assert.Equal(t, "test", resultUnredacted.Name)
	require.Len(t, resultUnredacted.Creds, 2)
	assert.Equal(t, "user1", resultUnredacted.Creds[0].Username)
	assert.Equal(t, "pass1", resultUnredacted.Creds[0].Password)
	assert.Equal(t, "user2", resultUnredacted.Creds[1].Username)
	assert.Equal(t, "pass2", resultUnredacted.Creds[1].Password)
}

func TestUnpackWithNoRedactedFields(t *testing.T) {
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

func TestUnpackRedactsMultipleStrings(t *testing.T) {
	type testConfig struct {
		StringVal1 string `config:"string_val1,redact"`
		StringVal2 string `config:"string_val2,redact"`
		NormalVal  string `config:"normal_val"`
	}

	input := testConfig{
		StringVal1: "secret1",
		StringVal2: "secret2",
		NormalVal:  "public",
	}

	// Test default behavior (strings redacted during Unpack)
	cfg, err := NewFrom(input)
	require.NoError(t, err)

	result := make(map[string]interface{})
	err = cfg.Unpack(&result)
	require.NoError(t, err)

	// Redacted string fields should be "[REDACTED]"
	assert.Equal(t, sREDACT, result["string_val1"])
	assert.Equal(t, sREDACT, result["string_val2"])
	assert.Equal(t, "public", result["normal_val"])

	// Test with ShowRedacted option (unredacted during Unpack)
	resultUnredacted := make(map[string]interface{})
	err = cfg.Unpack(&resultUnredacted, ShowRedacted)
	require.NoError(t, err)

	// Original values should be preserved
	assert.Equal(t, "secret1", resultUnredacted["string_val1"])
	assert.Equal(t, "secret2", resultUnredacted["string_val2"])
	assert.Equal(t, "public", resultUnredacted["normal_val"])
}

func TestUnpackRedactsOnlyStringsByteRune(t *testing.T) {
	type testConfig struct {
		StringVal string  `config:"string_val,redact"`
		IntVal    int     `config:"int_val,redact"`
		BoolVal   bool    `config:"bool_val,redact"`
		FloatVal  float64 `config:"float_val,redact"`
	}

	input := testConfig{
		StringVal: "secret",
		IntVal:    12345,
		BoolVal:   true,
		FloatVal:  3.14,
	}

	// Test default behavior - only string should be redacted
	cfg, err := NewFrom(input)
	require.NoError(t, err)

	result := make(map[string]interface{})
	err = cfg.Unpack(&result)
	require.NoError(t, err)

	// Only string field should be redacted
	assert.Equal(t, sREDACT, result["string_val"])
	// Non-string types should keep their original values (redact tag ignored)
	assert.Equal(t, uint64(12345), result["int_val"])
	assert.Equal(t, true, result["bool_val"])
	assert.Equal(t, 3.14, result["float_val"])
}

func TestUnpackRedactsAllSupportedTypes(t *testing.T) {
	type testConfig struct {
		StringVal string `config:"string_val,redact"`
		BytesVal  []byte `config:"bytes_val,redact"`
		RuneVal   []rune `config:"rune_val,redact"`
		NormalVal string `config:"normal_val"`
	}

	input := testConfig{
		StringVal: "secret-string",
		BytesVal:  []byte("secret-bytes"),
		RuneVal:   []rune("secret-rune"),
		NormalVal: "public",
	}

	// Test default behavior (string, []byte, and []rune redacted during Unpack)
	cfg, err := NewFrom(input)
	require.NoError(t, err)

	var result testConfig
	err = cfg.Unpack(&result)
	require.NoError(t, err)

	// All redactable types should be redacted
	assert.Equal(t, sREDACT, result.StringVal)
	assert.Equal(t, []byte(sREDACT), result.BytesVal)
	assert.Equal(t, []rune(sREDACT), result.RuneVal)
	assert.Equal(t, "public", result.NormalVal)

	// Test with ShowRedacted option (unredacted during Unpack)
	var resultUnredacted testConfig
	err = cfg.Unpack(&resultUnredacted, ShowRedacted)
	require.NoError(t, err)

	// Original values should be preserved
	assert.Equal(t, "secret-string", resultUnredacted.StringVal)
	assert.Equal(t, []byte("secret-bytes"), resultUnredacted.BytesVal)
	assert.Equal(t, []rune("secret-rune"), resultUnredacted.RuneVal)
	assert.Equal(t, "public", resultUnredacted.NormalVal)
}

func TestUnpackRedactsInlineStructs(t *testing.T) {
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

	// Test default behavior (redacted during Unpack)
	cfg, err := NewFrom(input)
	require.NoError(t, err)

	result := make(map[string]interface{})
	err = cfg.Unpack(&result)
	require.NoError(t, err)

	assert.Equal(t, "test", result["name"])
	assert.Equal(t, "public-key", result["key"])
	assert.Equal(t, sREDACT, result["secret"])

	// Test with ShowRedacted option (unredacted during Unpack)
	resultUnredacted := make(map[string]interface{})
	err = cfg.Unpack(&resultUnredacted, ShowRedacted)
	require.NoError(t, err)

	assert.Equal(t, "test", resultUnredacted["name"])
	assert.Equal(t, "public-key", resultUnredacted["key"])
	assert.Equal(t, "private-secret", resultUnredacted["secret"])
}

func TestUnpackRedactsAfterMerge(t *testing.T) {
	type testConfig struct {
		Username string `config:"username"`
		Password string `config:"password,redact"`
	}

	input1 := testConfig{
		Username: "admin",
		Password: "secret123",
	}

	// Create base config and merge (stores original values)
	cfg := New()
	err := cfg.Merge(input1)
	require.NoError(t, err)

	// Default Unpack applies redaction
	result := make(map[string]interface{})
	err = cfg.Unpack(&result)
	require.NoError(t, err)

	assert.Equal(t, "admin", result["username"])
	assert.Equal(t, sREDACT, result["password"])

	// Unpack with ShowRedacted shows original values
	result2 := make(map[string]interface{})
	err = cfg.Unpack(&result2, ShowRedacted)
	require.NoError(t, err)

	assert.Equal(t, "admin", result2["username"])
	assert.Equal(t, "secret123", result2["password"])
}

func TestUnpackRedactsCustomTypes(t *testing.T) {
	// Define custom types based on string, []byte, and []rune
	type CustomByteString []byte
	type CustomString string
	type CustomRuneString []rune

	type CustomStruct struct {
		CustomB CustomByteString `config:"custom_b,redact"`
		CustomS CustomString     `config:"custom_s,redact"`
		CustomR CustomRuneString `config:"custom_r,redact"`
		Normal  string           `config:"normal"`
	}

	input := CustomStruct{
		CustomB: CustomByteString("secret-bytes"),
		CustomS: CustomString("secret-string"),
		CustomR: CustomRuneString("secret-rune"),
		Normal:  "public",
	}

	// Test default behavior (custom types redacted during Unpack)
	cfg, err := NewFrom(input)
	require.NoError(t, err)

	var result CustomStruct
	err = cfg.Unpack(&result)
	require.NoError(t, err)

	// All custom redactable types should be redacted
	assert.Equal(t, CustomString(sREDACT), result.CustomS)
	assert.Equal(t, CustomByteString(sREDACT), result.CustomB)
	assert.Equal(t, CustomRuneString(sREDACT), result.CustomR)
	assert.Equal(t, "public", result.Normal)

	// Test with ShowRedacted option (unredacted during Unpack)
	var resultUnredacted CustomStruct
	err = cfg.Unpack(&resultUnredacted, ShowRedacted)
	require.NoError(t, err)

	// Original values should be preserved for custom types
	assert.Equal(t, CustomString("secret-string"), resultUnredacted.CustomS)
	assert.Equal(t, CustomByteString("secret-bytes"), resultUnredacted.CustomB)
	assert.Equal(t, CustomRuneString("secret-rune"), resultUnredacted.CustomR)
	assert.Equal(t, "public", resultUnredacted.Normal)
}
