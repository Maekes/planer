package main

import (
	"flag"
	"log"

	"github.com/Maekes/planer/handler"
	"github.com/gin-contrib/cors"

	jwtsession "github.com/ScottHuangZL/gin-jwt-session"
	"github.com/gin-gonic/autotls"
	"github.com/gin-gonic/gin"
)

func main() {

	flag.StringVar(&handler.MailPW, "pw", "", "Password for MailServer")
	flag.StringVar(&handler.KaplanSecret, "kaplan", "", "Secret RefID from Kaplan")
	local := flag.Bool("local", false, "Run Server on Local Machine")
	update := flag.Bool("update", false, "Run Update Function")
	noTLS := flag.Bool("notls", false, "Disable TLS for Webserver")
	flag.Parse()

	if *local {
		handler.InitHandler(":27017")
	} else {
		handler.InitHandler("mongo:27017")
		gin.SetMode(gin.ReleaseMode)
	}

	if *update {
		handler.Update()
	}

	r := gin.Default()

	jwtsession.JwtTokenName = "Token"                   //string without blank
	jwtsession.DefaultSessionName = "Session"           //string without blank
	jwtsession.DefaultFlashSessionName = "FlashSession" //string without blank

	jwtsession.NewStore()
	r.Use(jwtsession.ClearMiddleware())

	r.Use(cors.Default())

	//new template engine
	r.HTMLRender = handler.GetTemplateConfig()
	r.Static("/assets", "assets")
	r.Static("/demo", "demo")

	r.GET("/login", handler.LoginHandler)
	r.POST("/login", handler.ValidateJwtLoginHandler)

	r.GET("/register", handler.RegisterHandler)
	r.POST("/register", handler.RegisterPostHandler)

	r.GET("/rueckmeldung", handler.RueckmeldungFormHandler)
	r.POST("/rueckmeldung", handler.RueckmeldungPostFormHandler)

	auth := r.Group("/")
	auth.Use(handler.ValidationMiddleware)
	{
		auth.GET("/", handler.MessdienerplanHandler)
		auth.GET("/minis", handler.MinisHandler)

		auth.POST("/logout", handler.LogoutHandler)

		auth.GET("/messen", handler.MessenHandler)
		auth.GET("/zuordnen/:id", handler.ZuordnenHandler)
		auth.GET("/messdienerplan", handler.MessdienerplanHandler)
		auth.GET("/messdienerplan/pdf/:id", handler.MessdienerplanPdfHandler)
		auth.GET("/messdienerplan/xlsx/:id", handler.MessdienerplanXlsxHandler)
		auth.GET("/messdienerplan/delete/:id", handler.MessdienerplanDeleteHandler)

		auth.POST("/zuordnen/draged", handler.ZuordnenDragedHandler)
		auth.POST("/zuordnen/editInfoForPlan", handler.ZuordnenEditInfoForPlandHandler)
		auth.POST("/zuordnen/delete", handler.ZuordnenDeleteHandler)
		auth.POST("/zuordnen/finish", handler.ZuordnenFinishHandler)

		auth.POST("/messen", handler.AddMessenHandler)
		auth.GET("/messen/delete/:id", handler.MessenDeleteHandler)
		auth.POST("/messen/deleteAll", handler.MessenDeleteAllHandler)
		auth.POST("/messen/updateRelevantState", handler.MessenUpdateStateHandler)
		auth.POST("/messen/importFromExcel", handler.AddMessenFromExcelHandler)
		auth.POST("/messen/importFromKaplan", handler.AddMessenFromKaplanHandler)

		auth.POST("/messdienerplan/changehinweis", handler.MessdienerplanChangeHinweisHandler)
		auth.POST("/messdienerplan/create", handler.MessdienerplanCreateHandler)

		auth.GET("/minis/delete/:id", handler.MinisDeleteHandler)
		auth.POST("/minis", handler.AddMiniHandler)
		auth.POST("/minis/importFromExcel", handler.AddMiniFromExcelHandler)

		auth.GET("/einstellungen", handler.EinstellungenHandler)
		auth.POST("/einstellungen", handler.EinstellungenChangeHandler)
		auth.POST("/einstellungen/changepassword", handler.PasswordChangeHandler)

		admin := auth.Group("adminArea")
		admin.Use(handler.AdminMiddleware)
		{
			admin.GET("/", handler.AdminHandler)
			admin.GET("/user/delete/:id", handler.UserDeleteHandler)
			admin.GET("/user/resetPassword/:id", handler.UserResetPassword)
		}

	}

	r.NoRoute(handler.Error404Handler)

	if *local {
		r.Run("127.0.0.1:8080")
		//r.RunTLS("localhost:8080", "localhost.crt", "localhost.key")
		//log.Fatal(autotls.Run(r, "localhost"))
	} else if *noTLS {
		r.Run("0.0.0.0:80")
	} else {
		log.Fatal(autotls.Run(r, "planer.minis-quirin.de", "www.planer.minis-quirin.de"))
	}

}

func isFlagPassed(name string) bool {
	found := false
	flag.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}
