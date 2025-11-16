package main

type DatabaseConfig struct {
	Host     string `mapstructure:"host"`
	Port     uint16 `mapstructure:"port"`
	Database string `mapstructure:"database"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
}
