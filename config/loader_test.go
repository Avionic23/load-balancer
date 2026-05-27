package config

import (
	"fmt"
	"testing"
)

func TestLoad(t *testing.T) {
	config := Load("config_test.yaml")
	fmt.Println(config)
}
