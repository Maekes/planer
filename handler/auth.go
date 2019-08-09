package handler

import (
	"net/http"
	"time"

	"github.com/Maekes/planer/mongo/role"
	jwtsession "github.com/ScottHuangZL/gin-jwt-session"
	"github.com/gin-gonic/gin"
)

type Login struct {
	Username string `form:"username" json:"username" binding:"required"`
	Password string `form:"password" json:"password" binding:"required"`
}

func ValidateJwtLoginHandler(c *gin.Context) {
	var form Login
	//try get login info
	if err := c.ShouldBind(&form); err != nil {
		jwtsession.SetFlash(c, "Get login info error: "+err.Error())
		//c.Redirect(http.StatusMovedPermanently, "/login")
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"Title": "Login",
			"error": "Get login info error: " + err.Error(),
		})
		return
	}
	//validate login info
	if err := userService.ValidateUser(form.Username, form.Password); err != nil {
		jwtsession.SetFlash(c, "Error : username or password: "+err.Error())
		//c.Redirect(http.StatusMovedPermanently, "/login")
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"Title": "Login",
			"error": "Fehler : " + err.Error(),
		})
		return
	}
	//login info is correct, can generate JWT token and store in clien side now
	tokenDuration, _ := time.ParseDuration("1h")
	tokenString, err := jwtsession.GenerateJWTToken(form.Username, tokenDuration)
	if err != nil {
		jwtsession.SetFlash(c, "Error Generate token string: "+err.Error())
		//c.Redirect(http.StatusMovedPermanently, "/login")
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"Title": "Login",
			"error": "Error Generate token string: " + err.Error(),
		})
		return
	}

	err = jwtsession.SetTokenString(c, tokenString, 60*60) //60 minutes
	if err != nil {
		jwtsession.SetFlash(c, "Error set token string: "+err.Error())
		//c.Redirect(http.StatusMovedPermanently, "/login")
		c.HTML(http.StatusBadRequest, "login.html", gin.H{
			"Title": "Login",
			"error": "Error set token string: " + err.Error(),
		})
		return
	}
	jwtsession.SetFlash(c, "success : successful login")
	jwtsession.SetFlash(c, "username : "+form.Username)

	c.Redirect(http.StatusFound, "/messdienerplan")
	return
}

func ValidationMiddleware(c *gin.Context) {
	//flashes := jwtsession.GetFlashes(c)
	username, err := jwtsession.ValidateJWTToken(c)

	if err == nil && username != "" {
		u, err := userService.GetByUsername(username)
		if err != nil {
			c.Abort()
			c.Redirect(http.StatusTemporaryRedirect, "/login")
		} else {
			planService.ForUser(u.UUID)
			messeService.ForUser(u.UUID)
			miniService.ForUser(u.UUID)
			c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
			c.Header("Pragma", "no-cache") // HTTP 1.0.
			c.Header("Expires", "0")
			c.Next()
		}
	} else {
		c.Abort()
		c.Redirect(http.StatusTemporaryRedirect, "/login")
	}

}

func AdminMiddleware(c *gin.Context) {
	//flashes := jwtsession.GetFlashes(c)
	username, err := jwtsession.ValidateJWTToken(c)

	if err == nil && username != "" {
		u, err := userService.GetByUsername(username)
		if u.Role != role.Admin || err != nil {
			c.Abort()
			c.Redirect(http.StatusTemporaryRedirect, "/messdienerplan")
		} else {
			c.Next()
		}
	} else {
		c.Abort()
		c.Redirect(http.StatusTemporaryRedirect, "/messdienerplan")
	}

}

func LogoutHandler(c *gin.Context) {
	//flashes := jwtsession.GetFlashes(c)
	jwtsession.DeleteAllSession(c)
	c.Redirect(http.StatusMovedPermanently, "/login")

}
