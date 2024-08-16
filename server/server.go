package server

import (
	"context"
	"net/http"

	"github.com/scch94/ins_log"
	"github.com/scch94/micropagosSmtpServer/config"
	"github.com/scch94/micropagosSmtpServer/utils/router"
)

func StartServer(ctx context.Context) error {

	//agregamos el contexto en el que estamos
	ctx = ins_log.SetPackageNameInContext(ctx, "server")
	ins_log.Infof(ctx, "Starting server on Port: %v", config.Config.Port)

	router := router.SetupRouter(ctx)
	serverConfig := &http.Server{
		Addr:    config.Config.Port,
		Handler: router,
	}
	err := serverConfig.ListenAndServe()
	if err != nil {
		ins_log.Errorf(ctx, "cant connect to server: %v", err)
		return err
	}
	return nil
}
