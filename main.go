package main

import (
	"github.com/Maekes/planer/handler"

	jwtsession "github.com/ScottHuangZL/gin-jwt-session"
	"github.com/gin-gonic/gin"
)

func main() {

	/*
		userService.CreateNewUser("Nutzer", "max@mustermann.de", "123456")

		usr, _ := userService.GetByUsername("Nutzer")
		messeService.ForUser(usr.UUID)
		planService.ForUser(usr.UUID)
		miniService.ForUser(usr.UUID)

		uidp, _ := uuid.NewV4()
		t, _ := time.Parse("02.01.2006", "30.06.2019")

		plan := mongo.PlanModel{
			UUID:     uidp,
			Erstellt: time.Now(),
			Von:      time.Now(),
			Bis:      t,
			Titel:    "April / Juni (Dummy)",
		}

		//Act

		l, err := time.LoadLocation("Europe/Berlin")
		if xlFile, err := xls.Open("plan2.xls", "utf-8"); err == nil {
			for i := 0; i < xlFile.NumSheets(); i++ {
				sheet := xlFile.GetSheet(i)
				for j := 1; j <= int(sheet.MaxRow); j++ { //int(sheet.MaxRow)
					row := sheet.Row(j)
					if row.Col(0) == "" {
						break
					}
					d, err := time.ParseInLocation("2006-01-02T15:04:05Z", row.Col(1), l)
					u, err := time.ParseInLocation("15:04", row.Col(2), l)
					t, err := strconv.ParseFloat(row.Col(2), 32)
					u = timeFromExcelTime(t, true)
					s, err := time.ParseDuration("1s")
					u = u.Add(s) //Sekunde die Floating Point Fehler ausgleicht
					d = d.Add(time.Hour*time.Duration(u.Hour()) +
						time.Minute*time.Duration(u.Minute()) +
						0)

					if err != nil {
						log.Println("Could not Parse Time")
					}

					//ti := format.TimeFromExcelTime(t, true)

					uid, _ := uuid.NewV4()

					m := mongo.MesseModel{
						UUID:            uid,
						Datum:           d,
						Gottesdienst:    row.Col(3),
						LiturgischerTag: row.Col(5),
						Bemerkung:       row.Col(6),
						IsRelevant:      checkIfRelevant(row.Col(3), row.Col(6), d.Format("Mon"), d.Format("15:04")),
						//MinisForPlan:    []uuid.UUID{},
						Rueckmeldungen: []mongo.MiniModel{},
					}
					/*
						for i := 0; i < 10; i++ {
							m.MinisForPlan = append(m.MinisForPlan, uidm)
						}*/
	/*
					messeService.Create(&m)

				}

			}
		}

		planService.Create(&plan)
	*/
	//Add Minis
	/*
		if xlsFile, err := xls.Open("minis.xls", "utf-8"); err == nil {
			sheet := xlsFile.GetSheet(0)
			for j := 1; j < int(sheet.MaxRow); j++ {
				row := sheet.Row(j)
				if row.Col(1) == "" {
					break
				}
				uid, _ := uuid.NewV4()
				log.Println(row.Col(2))

				groups := []string{"gray", "azure", "indigo", "purple", "pink", "red", "orange", "yellow", "lime"}

				m := mongo.MiniModel{
					UUID:     uid,
					Vorname:  row.Col(1),
					Nachname: row.Col(2),
					Gruppe:   groups[rand.Intn(9)],
				}
				miniService.Create(&m)
			}
		}
	*/
	handler.InitHandler()
	//gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	jwtsession.JwtTokenName = "Token"                   //string without blank
	jwtsession.DefaultSessionName = "Session"           //string without blank
	jwtsession.DefaultFlashSessionName = "FlashSession" //string without blank

	jwtsession.NewStore()
	r.Use(jwtsession.ClearMiddleware())

	//new template engine
	r.HTMLRender = handler.GetTemplateConfig()
	r.Static("/assets", "assets")
	r.Static("/demo", "demo")

	r.GET("/login", handler.LoginHandler)
	r.POST("/login", handler.ValidateJwtLoginHandler)

	r.GET("/register", handler.RegisterHandler)
	r.POST("/register", handler.RegisterPostHandler)

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

		auth.POST("/zuordnen/draged", handler.ZuordnenDragedHandler)
		auth.POST("/zuordnen/editInfoForPlan", handler.ZuordnenEditInfoForPlandHandler)
		auth.POST("/zuordnen/delete", handler.ZuordnenDeleteHandler)

		auth.POST("/messen", handler.AddMessenHandler)
		auth.GET("/messen/delete/:id", handler.MessenDeleteHandler)
		auth.POST("/messen/deleteAll", handler.MessenDeleteAllHandler)
		auth.POST("/messen/updateRelevantState", handler.MessenUpdateStateHandler)
		auth.POST("/messen/importFromExcel", handler.AddMessenFromExcelHandler)

		auth.POST("/messdienerplan/create", handler.MessdienerplanCreateHandler)

		auth.GET("/minis/delete/:id", handler.MinisDeleteHandler)
		auth.POST("/minis", handler.AddMiniHandler)
		auth.POST("/minis/importFromExcel", handler.AddMiniFromExcelHandler)

	}

	r.NoRoute(handler.Error404Handler)

	//r.Run("127.0.0.1:8080")
	r.Run("0.0.0.0:80")

}
