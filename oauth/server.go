package outh

import (
	cerbosclient "github.com/common/commonmodule/cerbose"
	"github.com/common/commonmodule/config"
	oryclient "github.com/common/commonmodule/ory"
	"github.com/common/commonmodule/ports"
	"github.com/intelops/go-common/logging"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type Adapter struct {
	log          logging.Logger
	db           *gorm.DB
	grpcServer   *grpc.Server
	cfg          *config.Configurtion
	oryClient    oryclient.OryClient
	userapi      ports.UserAPIPort
	cerbosClient *cerbosclient.Client
}
