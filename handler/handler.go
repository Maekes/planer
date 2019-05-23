package handler

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Maekes/planer/mongo"

	"github.com/gin-gonic/gin"
	"github.com/jung-kurt/gofpdf"
	uuid "github.com/satori/go.uuid"
	"github.com/sergeilem/xls"
	zxcvbn "github.com/trustelem/zxcvbn"
)

const (
	mongoUrl = "localhost:27017"
	dbName   = "test_db"
)

var userService *mongo.UserService
var messeService *mongo.MesseService
var miniService *mongo.MiniService
var planService *mongo.PlanService

var session *mongo.Session

func InitHandler() {
	session, err := mongo.NewSession(":27017")
	if err != nil {
		log.Fatalf("Unable to connect to mongo: %s", err)
	}

	//session.DropDatabase(dbName)

	userService = mongo.NewUserService(session.Copy(), dbName, "user")
	messeService = mongo.NewMesseService(session.Copy(), dbName, "messen")
	miniService = mongo.NewMiniService(session.Copy(), dbName, "minis")
	planService = mongo.NewPlanService(session.Copy(), dbName, "plan")
}

func LoginHandler(c *gin.Context) {

	c.HTML(http.StatusOK, "login.html", gin.H{"Title": "Login"})

}

func RegisterHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", gin.H{"Title": "Register"})
}

type Register struct {
	Name           string `form:"name"  binding:"required"`
	Mail           string `form:"mail" binding:"required,email"`
	Password       string `form:"password" binding:"required"`
	PasswordRepeat string `form:"passwordRepeat" binding:"required"`
}

func RegisterPostHandler(c *gin.Context) {
	var form Register

	if err := c.ShouldBind(&form); err != nil {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"Title": "Register",
			"error": "Bitte alle Felder ausfüllen.",
			"user":  form.Name,
			"mail":  form.Mail,
		})
		return
	}

	score := zxcvbn.PasswordStrength(form.Password, nil)
	if score.Score < 3 {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"Title": "Register",
			"error": "Das gewählte Passwort ist zu unsicher.",
			"user":  form.Name,
			"mail":  form.Mail,
		})
		return
	}

	if form.Password != form.PasswordRepeat {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"Title": "Register",
			"error": "Passwörter sind nicht gleich.",
			"user":  form.Name,
			"mail":  form.Mail,
		})
		return
	}

	if err := userService.CreateNewUser(form.Name, form.Mail, form.Password); err != nil {
		c.HTML(http.StatusBadRequest, "register.html", gin.H{
			"Title": "Register",
			"error": err.Error(),
			"user":  form.Name,
			"mail":  form.Mail,
		})
		return
	}

	c.HTML(http.StatusOK, "login.html", gin.H{
		"Title":   "Login",
		"success": "Dein neuer Account wurde erfolgreich angelegt. Du kannst dich nun einloggen.",
	})
}

func ZuordnenHandler(c *gin.Context) {
	uid, err := uuid.FromString(c.Param("id"))
	if err != nil {
		//TODO
	}
	p, err := planService.GetPlanByUUID(uid)
	if err != nil {
		//TODO
	}

	messen, _ := messeService.GetAllMessenThatAreRelevantFromToDate(p.Von, p.Bis)
	groups := []string{"gray", "azure", "indigo", "purple", "pink", "red", "orange", "yellow", "lime"}
	var minis [][]mongo.MiniModel
	for _, v := range groups {
		m, err := miniService.GetAllMinisFromGroup(v)
		log.Println(m)
		if err != nil {
			//TODO
		}
		minis = append(minis, *m)
	}

	c.HTML(http.StatusOK, "zuordnen", gin.H{
		"title":         "Messen",
		"UUID":          p.UUID,
		"messenPayload": messen,
		"minisPayload":  minis,
		"planTitle":     p.Titel,
		"from":          p.Von.Format("02.01.2006"),
		"to":            p.Bis.Format("02.01.2006"),
	})
}

func MessdienerplanHandler(c *gin.Context) {

	plan, _ := planService.GetAllPlan()
	c.HTML(http.StatusOK, "messdienerplan", gin.H{
		"title":       "Messdienerplan",
		"planPayload": plan,
	})
}
func MessdienerplanCreateHandler(c *gin.Context) {
	log.Println(c.PostForm("from"))
	l, err := time.LoadLocation("Europe/Berlin")
	from, err := time.ParseInLocation("02.01.2006", c.PostForm("from"), l)
	to, err := time.ParseInLocation("02.01.2006", c.PostForm("to"), l)

	fromDate := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, l)
	toDate := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 0, l)
	//data, err := messeService.GetAllMessenFromToDate(fromDate, toDate)

	uid := uuid.NewV4()

	planService.Create(&mongo.PlanModel{
		UUID:     uid,
		Erstellt: time.Now(),
		Von:      fromDate,
		Bis:      toDate,
		Titel:    c.PostForm("title"),
	})

	if err != nil {
		log.Println(err)
	}

	c.Redirect(http.StatusFound, "/zuordnen/"+uid.String())
}

func MessdienerplanPdfHandler(c *gin.Context) {
	uid, err := uuid.FromString(c.Param("id"))
	if err != nil {
		//TODO
	}
	p, err := planService.GetPlanByUUID(uid)
	if err != nil {
		//TODO
	}
	messen, _ := messeService.GetAllMessenThatAreRelevantFromToDate(p.Von, p.Bis)
	messe := *messen

	marginCell := 0.2 // margin of top/bottom of cell
	pdf := gofpdf.New("P", "mm", "A4", "")
	utf := pdf.UnicodeTranslatorFromDescriptor("")
	pdf.SetLeftMargin(22)
	pdf.SetTopMargin(8)
	pdf.SetFont("Helvetica", "B", 20)
	pdf.SetHeaderFuncMode(func() {

		//pdf.SetDrawColor(val, val, val)
		//pdf.SetTextColor(val, val, val)
		pdf.Image("kirche.png", 2, 7, 17, 20, false, "", 0, "")
		pdf.SetFont("Helvetica", "", 25)
		pdf.TransformBegin()
		pdf.TransformRotate(90, 12, 255)
		pdf.Text(12, 255, " Messdienerplan   "+utf(p.Titel))
		pdf.TransformEnd()
		pdf.SetFont("Helvetica", "", 10)
		pdf.Text(23, 286, "Online: www.minis-quirin.de | Mail: leiterrunde@minis-quirin.de | WhatsApp: +49 1590 8120 575")
		pdf.Text(23, 291, "Gruppenstunden: Freitag 17:00 - 18:00 Uhr")
		pdf.Image("logo.png", 188, 280, 14, 14, false, "", 0, "")

	}, true)
	pdf.AddPage()
	pagew, pageh := pdf.GetPageSize()
	mleft, mright, _, mbottom := pdf.GetMargins()

	cols := []float64{25, 25, 50, pagew - mleft - mright - 30 - 30 - 50}
	rows := [][]string{}
	rows = append(rows, []string{"Datum", "Zeit", "Messe", "Messdiener"})

	for _, m := range messe {
		messdiener := ""
		for _, id := range m.MinisForPlan {
			minis, _ := miniService.GetMiniByUUID(id)
			messdiener = messdiener + minis.Vorname + " " + minis.Nachname + ", "
		}

		messdiener = strings.TrimSuffix(messdiener, ", ")
		if messdiener == "" {
			messdiener = "freiwillig"
		}

		g := m.Gottesdienst + "\n" + m.InfoForPlan
		rows = append(rows, []string{toGermanShort(m.Datum.Format("Mon")) + " " + m.Datum.Format("02.01."), m.Datum.Format("15:04"), utf(g), utf(messdiener)})
	}
	if err != nil {
		//TODO
		log.Println(err)
	}

	for rn, row := range rows {
		curx, y := pdf.GetXY()
		x := curx

		height := 0.
		lineHt := 5.5 // pdf.GetFontSize()

		//Calculate hight of Row
		for i, txt := range row {
			lines := pdf.SplitLines([]byte(txt), cols[i])
			h := float64(len(lines))*lineHt + marginCell*2
			if h > height {
				height = h
			}
		}
		width := pagew - mleft - mright
		// add a new page if the height of the row doesn't fit on the page
		if pdf.GetY()+height > pageh-mbottom {
			pdf.Line(x, y, x+width, y)
			pdf.AddPage()
			y = pdf.GetY()
		}
		if rn < 2 {
			pdf.SetLineWidth(0.75)
		} else {
			pdf.SetLineWidth(0.25)
		}
		pdf.Line(x, y, x+width, y)
		for i, txt := range row {
			if i == 0 || rn == 0 {
				pdf.SetFont("Helvetica", "B", 12)

			} else {
				if txt == "freiwillig" {
					pdf.SetFont("Helvetica", "I", 12)
				} else {
					pdf.SetFont("Helvetica", "", 12)
				}

			}
			width = cols[i]
			//	pdf.CellFormat(width, marginCell, "", "", 0, "L", false, 0, "")
			//	pdf.SetXY(x, y+marginCell)
			pdf.MultiCell(width, lineHt, txt, "", "L", false)
			//	pdf.SetXY(x, y+height)
			//	pdf.CellFormat(width, marginCell, "", "", 0, "L", false, 0, "")
			x += width
			pdf.SetXY(x, y)
		}

		pdf.SetXY(curx, y+height+2*marginCell)
		//pdf.Line(x+height+2*marginCell, y, x+width, y)
	}

	c.Header("Content-Disposition", "attachment; filename=Messdienerplan "+p.Titel+".pdf")
	c.Header("Content-Type", "application/pdf")
	//c.Header("Content-Length", r.Header.Get("Content-Length"))
	err = pdf.Output(c.Writer)
	if err != nil {
		log.Println(err)
	}

}

func toGermanShort(d string) string {
	switch d {
	case "Mon":
		return "Mo"
	case "Tue":
		return "Di"
	case "Wed":
		return "Mi"
	case "Thu":
		return "Do"
	case "Fri":
		return "Fr"
	case "Sat":
		return "Sa"
	case "Sun":
		return "So"
	default:
		return d
	}
}

func ZuordnenDragedHandler(c *gin.Context) {
	from := c.Query("from")
	to := c.Query("to")

	ufrom, err := uuid.FromString(from)
	uto, err := uuid.FromString(to)
	m, err := miniService.GetMiniByUUID(ufrom)
	err = messeService.AddMiniToMesse(uto, *m)

	if err != nil {
		log.Println(err)
		c.Status(http.StatusNoContent)
	} else {
		c.Status(http.StatusOK)
	}

	//fmt.Println(messeService.GetAllMessenThatAreRelevant())

}

func ZuordnenDeleteHandler(c *gin.Context) {
	uid := c.Query("uid")
	from := c.Query("from")

	uuuid, err := uuid.FromString(uid)
	uufrom, err := uuid.FromString(from)

	err = messeService.DeleteMiniFromMesse(uuuid, uufrom)

	if err != nil {
		log.Println(err)
		c.Status(http.StatusNoContent)
	} else {
		c.Status(http.StatusOK)
	}

	//fmt.Println(messeService.GetAllMessenThatAreRelevant())
	c.Redirect(http.StatusFound, "/minis")
}

func ZuordnenEditInfoForPlandHandler(c *gin.Context) {
	uid, err := uuid.FromString(c.Query("uid"))
	value := c.Query("value")

	err = messeService.ChangeInfoForPlan(uid, value)
	if err != nil {
		log.Println(err)
		//TODO
	}
}

func MinisHandler(c *gin.Context) {
	data, _ := miniService.GetAllMinis()
	c.HTML(http.StatusOK, "messdienerliste", gin.H{
		"title":    "Messdienerliste",
		"username": userService.GetUsernameByID(miniService.AktUser),
		"payload":  data,
	})
}

func MessenHandler(c *gin.Context) {
	data, _ := messeService.GetAllMessen() //GetAllMessenFromDate(time.Now().AddDate(0, 0, -7))
	c.HTML(http.StatusOK, "messen", gin.H{
		"title":   "Messen",
		"payload": data,
		"from":    time.Now().Format("2006-01-02"),
		"to":      time.Now().Format("2006-01-02"),
	})
}

func MessenDeleteToHandler(c *gin.Context) {
	err := messeService.DeleteAllMessenToDate(time.Now())
	if err != nil {
		//TODO
	}
	c.Redirect(http.StatusFound, "/messen")
}

func MessenfromtoDateHandler(c *gin.Context) {
	from, err := time.Parse("2006-01-02", c.PostForm("from"))
	to, err := time.Parse("2006-01-02", c.PostForm("to"))
	log.Println(c.PostForm("from"))
	l, err := time.LoadLocation("Europe/Berlin")

	fromDate := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, l)
	toDate := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 0, l)
	data, err := messeService.GetAllMessenFromToDate(fromDate, toDate)
	if err != nil {
		log.Println(err)
	}

	c.HTML(http.StatusFound, "/messen", gin.H{
		"title":   "Messen",
		"payload": data,
		"from":    c.PostForm("from"),
		"to":      c.PostForm("to"),
	})
}

func MessenDeleteHandler(c *gin.Context) {
	idToSreach, err := uuid.FromString(c.Param("id"))
	if err != nil {
		c.Redirect(http.StatusFound, "/messen")
	}
	messeService.DeleteMesseByUId(idToSreach)
	c.Redirect(http.StatusFound, "/messen")
}

func MessenDeleteAllHandler(c *gin.Context) {

	err := messeService.DeleteAllMessen()
	log.Println(err)
	c.Redirect(http.StatusFound, "/messen")
}

func MinisDeleteHandler(c *gin.Context) {
	idToSreach := c.Param("id")
	miniService.DeleteMiniById(idToSreach)
	c.Redirect(http.StatusFound, "/minis")
}

func AddMessenHandler(c *gin.Context) {
	l, err := time.LoadLocation("Europe/Berlin")
	d, err := time.ParseInLocation("02.01.2006 15:04", c.PostForm("date"), l)
	g := c.PostForm("gottesdienst")
	t := c.PostForm("tag")
	b := c.PostForm("bemerkung")

	log.Print(d)
	if err != nil || g == "" {
		log.Println(err)
		c.Redirect(http.StatusFound, "/messen")
		return
	}

	if c.PostForm("uuid") == "" {
		uidm := uuid.NewV4()
		messe := mongo.MesseModel{
			UUID:            uidm,
			Datum:           d,
			Gottesdienst:    g,
			LiturgischerTag: t,
			Bemerkung:       b,
			IsRelevant:      true,
		}
		messeService.Create(&messe)
	} else {
		uid, err := uuid.FromString(c.PostForm("uuid"))
		if err != nil {
			c.Redirect(http.StatusFound, "/messe")
			return
		}
		m, err := messeService.GetMesseByUUID(uid)

		if err != nil {
			c.Redirect(http.StatusFound, "/messe")
			return
		}

		m.Datum = d
		m.Gottesdienst = g
		m.LiturgischerTag = t
		m.Bemerkung = b

		err = messeService.UpdateMesse(m)

		if err != nil {
			//TODO
		}
	}
	c.Redirect(http.StatusFound, "/messen")
}

func AddMiniHandler(c *gin.Context) {

	v := c.PostForm("vorname")
	n := c.PostForm("nachname")
	g := c.PostForm("gruppe")
	if g == "" || v == "" || n == "" {
		c.Redirect(http.StatusFound, "/minis")
		return
	}
	if c.PostForm("uuid") == "" {
		uidm := uuid.NewV4()
		log.Println(g)
		mini := mongo.MiniModel{
			UUID:     uidm,
			Vorname:  v,
			Nachname: n,
			Gruppe:   g,
		}
		miniService.Create(&mini)
	} else {
		uid, err := uuid.FromString(c.PostForm("uuid"))
		if err != nil {
			c.Redirect(http.StatusFound, "/minis")
			return
		}
		m, err := miniService.GetMiniByUUID(uid)

		if err != nil {
			c.Redirect(http.StatusFound, "/minis")
			return
		}
		m.Nachname = n
		m.Vorname = v
		m.Gruppe = g
		err = miniService.UpdateMini(m)
		if err != nil {
			//TODO
		}
	}
	c.Redirect(http.StatusFound, "/minis")
}

func AddMiniFromExcelHandler(c *gin.Context) {

	defer func(c *gin.Context) {
		if rec := recover(); rec != nil {
			c.Redirect(http.StatusFound, "/minis")

		}
	}(c)

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.Redirect(http.StatusFound, "/minis")
		return
	}

	if xlsFile, err := xls.OpenReader(file, "utf-8"); err == nil {
		if err != nil {
			c.Redirect(http.StatusFound, "/minis")
			return
		}
		sheet := xlsFile.GetSheet(0)
		for j := 1; j < int(sheet.MaxRow); j++ {

			row := sheet.Row(j)
			if row.Col(0) == "" {
				break
			}

			uid := uuid.NewV4()
			log.Println(row.Col(0))

			groups := []string{"gray", "azure", "indigo", "purple", "pink", "red", "orange", "yellow", "lime"}
			gn, err := strconv.Atoi(row.Col(2))

			if err != nil || gn < 1 || gn > 9 {
				gn = 1
			}

			m := mongo.MiniModel{
				UUID:     uid,
				Vorname:  row.Col(0),
				Nachname: row.Col(1),
				Gruppe:   groups[gn-1],
			}
			miniService.Create(&m)
		}
	}
	c.Redirect(http.StatusFound, "/minis")
}
func AddMessenFromExcelHandler(c *gin.Context) {

	defer func(c *gin.Context) {
		if rec := recover(); rec != nil {
			c.Redirect(http.StatusFound, "/messen")

		}
	}(c)

	file, _, err := c.Request.FormFile("file")
	if err != nil {
		c.Redirect(http.StatusFound, "/messen")
		return
	}

	//
	l, err := time.LoadLocation("Europe/Berlin")
	if xlFile, err := xls.OpenReader(file, "utf-8"); err == nil {
		if err != nil {
			c.Redirect(http.StatusFound, "/messen")
			return
		}
		for i := 0; i < xlFile.NumSheets(); i++ {
			sheet := xlFile.GetSheet(i)
			for j := 1; j <= int(sheet.MaxRow); j++ { //int(sheet.MaxRow)
				row := sheet.Row(j)
				if row.Col(0) == "" {
					break
				}
				fmt.Println(row.Col(1))
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

				uid := uuid.NewV4()

				m := mongo.MesseModel{
					UUID:            uid,
					Datum:           d,
					Gottesdienst:    row.Col(3),
					LiturgischerTag: row.Col(5),
					Bemerkung:       row.Col(6),
					IsRelevant:      checkIfRelevant(row.Col(3), row.Col(6), d.Format("Mon"), d.Format("15:04")),
				}

				messeService.Create(&m)
			}
		}
	}

	c.Redirect(http.StatusFound, "/messen")
}

func MessenUpdateStateHandler(c *gin.Context) {
	uid := c.Query("uid")
	state := c.Query("state")
	s := false
	if state == "true" {
		s = true
	}
	u, err := uuid.FromString(uid)
	err = messeService.UpdateRelevantMesseByUId(u, s)
	if err != nil {
		log.Println(err)
	}

}

func Error404Handler(c *gin.Context) {
	c.HTML(http.StatusNotFound, "404.html", gin.H{})
}

// g = Gottesdienst, b = Bemerkung, t = Tag, u = Uhrzeit
func checkIfRelevant(g string, b string, t string, u string) bool {
	if strings.Contains(b, "panisch") {
		return false
	}

	if strings.Contains(b, "ortugiesische") {
		return false
	}

	if strings.Contains(b, "ENTFÄLLT") {
		return false
	}

	if strings.Contains(b, "auswärts") {
		return false
	}

	if strings.Contains(b, "SSB") {
		return false
	}

	if strings.Contains(g, "andacht") {
		return false
	}

	if strings.Contains(g, "Andacht") {
		return false
	}

	if (t == "Fri") && (u == "09:00") && (g == "Messe" || g == "Festmesse") {
		return false
	}

	switch g {
	case "Schulmesse":
		return false
	case "Schulgottesdienst":
		return false
	case "Andacht":
		return false
	case "Beichtgelegenheit":
		return false
	case "Laudes":
		return false
	case "Vesper":
		return false
	case "Komplet":
		return false
	case "Kreuzweg":
		return false
	}

	return true
}
