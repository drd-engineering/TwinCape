package domains

import (
	"github.com/drd-engineering/TwinCape/domains/authenticator"
	"github.com/drd-engineering/TwinCape/domains/register"
	"github.com/drd-engineering/TwinCape/routes"
)

//InitiateRoutes is method used to create routing for all the domains available
func InitiateRoutes() {
	r := routes.GetInstance()
	apiRoutes := r.Group("/api/v1/sso")
	{
		apiRoutes.Use(routes.DRDApplicationIdentification())

		routeforRegistration := apiRoutes.Group("/register")
		routeforRegistration.POST("/save-user", register.SaveUser)

		routeforAuth := apiRoutes.Group("/auth")
		routeforAuth.POST("/login", authenticator.Login)
		routeforAuth.POST("/refresh-token", authenticator.RefreshToken)
		routeforAuth.Use(routes.AuthorizationBearer())
		{
			routeforAuth.POST("/check-token", authenticator.CheckToken)
			routeforAuth.POST("/get-login-details", authenticator.GetLoginDetails)
		}
	}
	return
}
