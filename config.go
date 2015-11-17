package main

import "github.com/jsternberg/ansible-agent/ansible"

type Config struct {
	SSL  SSLSection
	Ldap ansible.LdapOptions
}

type SSLSection struct {
	Enabled     bool
	Certificate string
	PrivateKey  string `toml:"private_key"`
}

func DefaultConfig() *Config {
	return &Config{}
}
