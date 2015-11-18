package main

import (
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"
)

const TestConfig = `
[ssl]
enabled = true
certificate = "/etc/ansible/server.cert"
private_key = "/etc/ansible/server.key"

[ldap]
enabled = true
host = "ldaps://example.com"
port = 636
base_dn = "dc=example,dc=com"
user_filter = "(uid=%s)"
`

func TestDefaultConfig(t *testing.T) {
	assert := assert.New(t)

	config := DefaultConfig()
	assert.False(config.SSL.Enabled)
	assert.False(config.Ldap.Enabled)
}

func TestConfigLoad(t *testing.T) {
	assert := assert.New(t)

	config := DefaultConfig()
	if err := toml.Unmarshal([]byte(TestConfig), config); assert.NoError(err) {
		assert.True(config.SSL.Enabled)
		assert.Equal("/etc/ansible/server.cert", config.SSL.Certificate)
		assert.Equal("/etc/ansible/server.key", config.SSL.PrivateKey)

		assert.True(config.Ldap.Enabled)
		assert.Equal("ldaps://example.com", config.Ldap.Host)
		assert.Equal(uint16(636), config.Ldap.Port)
		assert.Equal("dc=example,dc=com", config.Ldap.BaseDN)
		assert.Equal("(uid=%s)", config.Ldap.UserFilter)
	}
}
