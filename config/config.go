package config

import "code.google.com/p/gcfg"

type HTTPConfig struct {
	Host        string
	Port        int
	UseTLS      bool
	TLSCert     string
	TLSKey      string
	TemplateDir string
}

type Config struct {
	Global struct {
		Type string
		Dump string
	}
	Http HTTPConfig
}

func (self *Config) ParseConfig(file string) error {
	if err := gcfg.ReadFileInto(self, file); err != nil {
		return err
	}
	return nil
}
