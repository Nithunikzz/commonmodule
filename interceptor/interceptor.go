package interceptor

import (
	"context"
	"fmt"
	"os"
	"strings"

	//"gitlab.com/tariandev_intelops/iam/internal/application/domain"
	"github.com/common/commonmodule/domain"
	"gopkg.in/yaml.v2"

	cerbos "github.com/cerbos/cerbos/client"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// Config contains the configuration for the interceptor.
type Config struct {
	Exclude               []string `yaml:"exclude"`
	Authenticate          []string `yaml:"authenticate"`
	AuthenticateAuthorize []string `yaml:"authenticate-authorize"`
}

func (a Adapter) UnaryInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	modified := strings.Replace(info.FullMethod, "/", "-", -1)
	modified = strings.TrimPrefix(modified, "-")
	if strings.Contains(modified, ".") {
		modified = modified[strings.LastIndex(modified, ".")+1:]
	}
	fmt.Println("full method name: ", modified)
	a.log.Info("func UnaryInterceptor invoked")
	defer a.log.Info("func UnaryInterceptor exited ")
	// Read the config from the YAML file
	config := &domain.Config{}
	err := a.readConfig(config)
	if err != nil {
		a.log.Info("Error occurred while reading config file. Error -", err.Error())
		return nil, err
	}
	// Check if the method is in the Exclude list
	for _, method := range config.Exclude {
		if info.FullMethod == method {
			// If the method is in the Exclude list, skip authentication and authorization
			return handler(ctx, req)
		}
	}
	// Check if the method is in the Authenticate list
	for _, method := range config.Authenticate {
		if info.FullMethod == method {
			// If the method is in the Authenticate list, only check if the session is active
			accessToken, err := a.oryClient.GetTokenFromContext(ctx)
			if err == nil {
				ctx, err = a.oryClient.Authorize(ctx, accessToken)
				if err != nil {
					a.log.Info("Error occurred while authorizing the session id from context. Session Id - "+accessToken+"\nError -", err.Error())
					return nil, err
				}
				return handler(ctx, req)
			} else {
				oauthToken, err := a.oryClient.GetOauthTokenFromContext(ctx)
				if err == nil {
					active, err := a.oryClient.IntrospectToken(ctx, oauthToken)
					if err != nil || !active {
						return nil, status.Error(codes.Unauthenticated, "Invalid oauth token")
					}
					return handler(ctx, req)
				} else {
					a.log.Info("Error occurred while getting oauth token from context. Error -", err.Error())
					return nil, err
				}
			}
		}
	}

	// Check if the method is in the AuthenticateAuthorize list
	// CreateOrg , GetUserDetails should not be put into this category
	// only a user who already there in IAM also has a organisation
	// can be authorized
	for _, method := range config.AuthenticateAuthorize {
		if info.FullMethod == method {
			// If the method is in the AuthenticateAuthorize list,
			// check if the session is active and perform authorization logic
			accessToken, err := a.oryClient.GetTokenFromContext(ctx)
			if err != nil {
				a.log.
					Info("Error occurred while getting session id from context. " +
						fmt.Sprintf("\nError - %s", err.Error()),
					)
				return nil, err
			}
			ctx, err = a.oryClient.Authorize(ctx, accessToken)
			if err != nil {
				a.log.Info(
					"Error occurred while authorizing the session id from context. " +
						fmt.Sprintf("Session Id - %s", accessToken) +
						fmt.Sprintf("\nError - %s", err.Error()),
				)
				return nil, err
			}

			// Get the metadata from the incoming context
			oryid, err := a.oryClient.GetOryIDFromContext(ctx)
			if err != nil {
				a.log.
					Info("Error occurred while getting ory id from context." +
						fmt.Sprintf("\nError - %s", err.Error()),
					)
				return nil, err
			}
			// get userid with oryid
			user, err := a.userapi.GetUserByOryID(oryid)
			if err != nil {
				a.log.Info("Error occurred while getting user by ory id. " +
					fmt.Sprintf("\nError - %s", err.Error()),
				)
				return nil, err
			}
			orgid, err := a.oryClient.GetOrgIdFromContext(ctx)
			if err != nil {
				a.log.Info("Error occurred while getting org id from context. " +
					fmt.Sprintf("\nError - %s", err.Error()),
				)
				return nil, err
			}
			// Get roles associated with user in organization
			var orgUserRoles []domain.OrganisationUserRole
			result := a.db.Where("user_id = ? AND organisation_id = ?", user.ID, orgid).
				Preload("Role").Find(&orgUserRoles)
			if result.Error != nil {
				return nil, fmt.Errorf("error getting organisation user roles: %w",
					result.Error)
			}
			// Get actions associated with roles and remove duplicates
			actionsMap := make(map[string]domain.Action)
			for _, orgUserRole := range orgUserRoles {
				var roleActions []domain.RoleAction
				a.db.Where("role_id = ?", orgUserRole.Role.ID).Preload("Action").
					Find(&roleActions)

				for _, roleAction := range roleActions {
					actionsMap[roleAction.Action.ID] = roleAction.Action
				}
			}
			// Create slice of actions
			actions := make([]string, 0, len(actionsMap))
			for _, action := range actionsMap {
				actions = append(actions, action.Name)
			}
			// here instead of roles we are using actions
			principal := cerbos.NewPrincipal(user.Email, actions...)
			// Remove the leading slash
			input := strings.TrimPrefix(info.FullMethod, "/")

			// Replace all slashes with hyphens
			input = strings.ReplaceAll(input, "/", "-")

			// Replace all dots with hyphens
			input = strings.ReplaceAll(input, ".", "-")
			r := cerbos.NewResource(input, user.Email)

			allowed, err := a.cerbosClient.Client.IsAllowed(context.Background(), principal, r, "*")

			if err != nil {
				a.log.Info("Error occurred while checking is allowed or not. " +
					fmt.Sprintf("\nError - %s", err.Error()),
				)
				return nil, err
			}
			if allowed {
				return handler(ctx, req)
			} else {
				return nil, fmt.Errorf("not allowed")
			}
		}

	}
	return handler(ctx, req)
}

// readConfig reads the config file and decodes it into Config struct
func (a Adapter) readConfig(config *domain.Config) error {
	// Read the file location from an environment variable
	fileLocation := os.Getenv("CONFIG_FILE_LOCATION")
	if fileLocation == "" {
		fileLocation = "/etc/myapp/config.yaml"
	}
	// Open the file
	file, err := os.Open(fileLocation)
	if err != nil {
		return err
	}
	// Close the file when we are done
	defer file.Close()
	// Decode the file into our struct
	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(config)
	if err != nil {
		return err
	}
	return nil
}
