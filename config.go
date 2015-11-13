package main

type Config struct {
	SSL SSLSection
}

type SSLSection struct {
	Enabled     bool
	Certificate string
	PrivateKey  string `toml:"private_key"`
}

func DefaultConfig() *Config {
	return &Config{}
}
