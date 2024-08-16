package config

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/scch94/Gconfiguration"
	"github.com/scch94/ins_log"
)

var Config smtpConfig

type smtpConfig struct {
	Saludo   string `json:"saludo"`
	LogLevel string `json:"log_Level"`
	LogName  string `json:"log_Name"`
	Port     string `json:"port"`
}

func Upconfig(ctx context.Context) error {
	//traemos el contexto y le setiamos el contexto actual
	// Agregamos el valor "packageName" al contexto
	ctx = ins_log.SetPackageNameInContext(ctx, "config")

	ins_log.Info(ctx, "starting to get the config struct ")
	err := Gconfiguration.GetConfig(&Config, "../config", "smtpServer.json")

	if err != nil {
		ins_log.Fatalf(ctx, "error in Gconfiguration.GetConfig() ", err)
		return err
	}
	return nil
}

// metodo para volver la config es json
func (s smtpConfig) ConfigurationString() string {
	configJSON, err := json.Marshal(s)
	if err != nil {
		return fmt.Sprintf("Error al convertir la configuración a JSON: %v", err)
	}
	return string(configJSON)
}
