package handler

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/360EntSecGroup-Skylar/excelize/v2"
	"github.com/Maekes/planer/mongo"
	"github.com/Maekes/planer/mongo/role"
	"gopkg.in/gomail.v2"

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

var MailPW string
var MailAdress string
var MailResponse string
var MailSMTP string
var MailPort string
var KaplanSecret string
var KaplanHost string

var session *mongo.Session

func InitHandler(url string) {
	session, err := mongo.NewSession(url)
	if err != nil {
		log.Fatalf("Unable to connect to mongo: %s", err)
	}

	MailPW = os.Getenv("MailPW")
	MailAdress = os.Getenv("MailAdress")
	MailResponse = os.Getenv("MailResponse")
	MailSMTP = os.Getenv("MailSMTP")
	MailPort = os.Getenv("MailPort")
	KaplanSecret = os.Getenv("KaplanSecret")
	KaplanHost = os.Getenv("KaplanHost")

	userService = mongo.NewUserService(session.Copy(), dbName, "user")
	messeService = mongo.NewMesseService(session.Copy(), dbName, "messen")
	miniService = mongo.NewMiniService(session.Copy(), dbName, "minis")
	planService = mongo.NewPlanService(session.Copy(), dbName, "plan")

	if !userService.ExistsAdmin() {
		userService.CreateNewUser("admin", os.Getenv("AdminMail"), "admin", role.Admin)
		log.Println("#### Created new admin u:admin/p:admin ####")
	}
}

func LoginHandler(c *gin.Context) {

	c.HTML(http.StatusOK, "login.html", gin.H{"Title": "Login"})

}

func RegisterHandler(c *gin.Context) {
	c.HTML(http.StatusOK, "register.html", gin.H{"Title": "Register"})
}

func RueckmeldungFormHandler(c *gin.Context) {

	publicId, pid := c.GetQuery("pid")
	if pid {
		uuid, err := uuid.FromString(publicId)
		if err != nil {
			Error404Handler(c)
			return
		}

		puuid, err := userService.GetPrivateUUID(uuid)
		if err != nil {
			Error404Handler(c)
			return
		}

		plan, err := planService.GetNewestPlanFromUser(puuid)
		if err != nil {
			Error404Handler(c)
			return
		}

		c.Redirect(http.StatusFound, "/rueckmeldung?id="+plan.UUID.String())
	}

	planId, plid := c.GetQuery("id")
	if plid {
		uuid, err := uuid.FromString(planId)
		if err != nil {
			Error404Handler(c)
			return
		}
		p, err := planService.GetPlanByUUIDPublic(uuid)
		messen, err := messeService.GetAllMessenThatAreRelevantFromToDatePublic(p.Von, p.Bis, p.UserUUID)

		if err != nil {
			//	log.Println(err.Error())
		}

		c.HTML(http.StatusOK, "rueckmeldung-form.html", gin.H{
			"title":         "Messen",
			"messenPayload": messen,
			"planTitle":     p.Titel,
			"planID":        p.UUID,
			"hinweis":       p.RueckmeldungHinweis,
			"from":          p.Von.Format("02.01.2006"),
			"to":            p.Bis.Format("02.01.2006"),
		})
	}

}

func RueckmeldungPostFormHandler(c *gin.Context) {

	var message string
	var error bool

	error = false
	message = "Vielen Dank. Die Rückmeldung wurde verschickt."

	name := c.PostForm("name")
	messen := c.PostFormArray("uuid")
	hinweis := c.PostForm("hinweis")
	uid := c.PostForm("planid")
	planid, err := uuid.FromString(uid)

	if err != nil {
		c.JSON(200, gin.H{
			"error":   true,
			"message": "Es ist ein Fehler aufgetreten",
		})
		return
	}

	planService.NewRueckmeldungPublic(name, messen, hinweis, planid)
	plan, err := planService.GetPlanByUUIDPublic(planid)

	for _, m := range messen {
		uid, err := uuid.FromString(m)
		if err != nil {
			if err != nil {
				c.JSON(200, gin.H{
					"error":   error,
					"message": "Es ist ein Fehler aufgetreten",
				})
				return
			}
		}
		messeService.AddNameToMessePublic(name, uid)
	}

	me, err := messeService.GetAllMessenWithUUIDsPublic(messen)

	type rueckmeldung struct {
		Plantitel    string
		Name         string
		Messen       *[]mongo.MesseModel
		Hinweis      string
		MailResponse string
	}

	r := rueckmeldung{plan.Titel, name, me, hinweis, MailResponse}

	t, err := template.New("rueckmeldung-mail-template.html").Funcs(C.Funcs).ParseFiles("views/rueckmeldung-mail-template.html")
	if err != nil {
		log.Println(err)
		error = true
		if err != nil {
			c.JSON(200, gin.H{
				"error":   error,
				"message": "Es ist ein Fehler aufgetreten",
			})
			return
		}
	}

	var tpl bytes.Buffer
	if err := t.Execute(&tpl, r); err != nil {
		log.Println(err)
		error = true
		if err != nil {
			c.JSON(200, gin.H{
				"error":   error,
				"message": "Es ist ein Fehler aufgetreten",
			})
			return
		}
	}

	result := tpl.String()

	m := gomail.NewMessage()
	m.SetHeader("From", MailAdress)
	m.SetHeader("To", MailResponse)
	m.SetHeader("Subject", name+" | Rückmeldung Plan "+plan.Titel)
	m.SetBody("text/html", result)

	port, _ := strconv.Atoi(MailPort)

	d := gomail.NewDialer(MailSMTP, port, MailAdress, MailPW)
	d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

	// Send the email

	if err := d.DialAndSend(m); err != nil {
		error = true
		message = "Die Nachricht konnte nicht versendet werden. Versuchen Sie es später noch einmal."
	} else {
		error = false
		message = "Vielen Dank. Die Rückmeldung wurde verschickt."
	}

	time.Sleep(2 * time.Second)
	c.JSON(200, gin.H{
		"error":   error,
		"message": message,
	})

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

	if err := userService.CreateNewUser(form.Name, form.Mail, form.Password, role.User); err != nil {
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

	sortingOrder := c.Query("sortBy")

	switch sortingOrder {
	case "group":
		for _, v := range groups {
			m, err := miniService.GetAllMinisFromGroup(v)
			if err != nil {
				//TODO
			}
			minis = append(minis, *m)
		}
	case "name":
		m, err := miniService.GetAllMinis()
		if err != nil {
			//TODO
		}
		minis = append(minis, *m)
	default: //defalut Case is equal to group Case
		sortingOrder = "group"
		for _, v := range groups {
			m, err := miniService.GetAllMinisFromGroup(v)
			if err != nil {
				//TODO
			}
			minis = append(minis, *m)
		}
	}

	c.HTML(http.StatusOK, "zuordnen", gin.H{
		"title":          "Messen",
		"username":       userService.GetUsernameByID(miniService.AktUser),
		"role":           userService.GetRoleByID(miniService.AktUser),
		"UUID":           p.UUID,
		"messenPayload":  messen,
		"minisPayload":   minis,
		"planTitle":      p.Titel,
		"rueckmeldungen": p.Rueckmeldungen,
		"from":           p.Von.Format("02.01.2006"),
		"to":             p.Bis.Format("02.01.2006"),
		"sortBy":         sortingOrder,
	})
}

func MessdienerplanHandler(c *gin.Context) {
	// TODO Handle error
	plan, _ := planService.GetAllPlan()
	maxDate, _ := messeService.GetMaxDate()
	c.HTML(http.StatusOK, "messdienerplan", gin.H{
		"title":       "Messdienerplan",
		"planPayload": plan,
		"maxDate":     maxDate,
		"username":    userService.GetUsernameByID(miniService.AktUser),
		"role":        userService.GetRoleByID(miniService.AktUser),
	})

}
func MessdienerplanCreateHandler(c *gin.Context) {

	l, err := time.LoadLocation("Europe/Berlin")
	daterange := strings.Split(c.PostForm("daterange"), " - ")
	if len(daterange) != 2 || c.PostForm("title") == "" {
		c.Redirect(http.StatusFound, "/messdienerplan")
		return
	}
	from, err := time.ParseInLocation("02.01.2006", daterange[0], l)
	to, err := time.ParseInLocation("02.01.2006", daterange[1], l)

	fromDate := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, l)
	toDate := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 0, l)
	//data, err := messeService.GetAllMessenFromToDate(fromDate, toDate)

	uid := uuid.NewV4()

	err = planService.Create(&mongo.PlanModel{
		UUID:                uid,
		Erstellt:            time.Now(),
		Von:                 fromDate,
		Bis:                 toDate,
		Titel:               c.PostForm("title"),
		RueckmeldungHinweis: c.PostForm("hinweis"),
	})

	if err != nil {
		log.Println(err)
	}

	c.Redirect(http.StatusFound, "/messdienerplan")
}

func MessdienerplanXlsxHandler(c *gin.Context) {
	uid, err := uuid.FromString(c.Param("id"))
	if err != nil {
		c.Redirect(http.StatusFound, "/messdienerplan")
	}
	p, err := planService.GetPlanByUUID(uid)
	if err != nil {
		c.Redirect(http.StatusFound, "/messdienerplan")
	}
	messen, _ := messeService.GetAllMessenThatAreRelevantFromToDate(p.Von, p.Bis)
	messe := *messen

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
		rows = append(rows, []string{toGermanShort(m.Datum.Format("Mon")) + " " + m.Datum.Format("02.01."), m.Datum.Format("15:04"), g, messdiener})
	}

	f := excelize.NewFile()
	// Create a new sheet.
	sheetname := fmt.Sprintf("%v - %v", p.Von.Format("02.01."), p.Bis.Format("02.01."))
	index := f.NewSheet(sheetname)

	for rn, row := range rows {
		for cn, cel := range row {
			celindex, _ := excelize.CoordinatesToCellName(cn+1, rn+1)
			f.SetCellValue(sheetname, celindex, cel)
		}
	}

	// Set active sheet of the workbook.
	f.SetActiveSheet(index)

	c.Header("Content-Disposition", "attachment; filename=Messdienerplan "+p.Titel+".xlsx")
	c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
	//c.Header("Content-Length", r.Header.Get("Content-Length"))

	err = f.Write(c.Writer)
	if err != nil {
		log.Println(err)
	}
}
func MessdienerplanPdfHandler(c *gin.Context) {
	uid, err := uuid.FromString(c.Param("id"))
	if err != nil {
		c.Redirect(http.StatusFound, "/messdienerplan")
	}
	p, err := planService.GetPlanByUUID(uid)
	if err != nil {
		c.Redirect(http.StatusFound, "/messdienerplan")
	}
	messen, _ := messeService.GetAllMessenThatAreRelevantFromToDate(p.Von, p.Bis)
	messe := *messen

	footer, err := userService.GetFooterTextAktUser()
	if err != nil {
		footer = []string{}
	}

	marginCell := 0.2 // margin of top/bottom of cell
	pdf := gofpdf.New("P", "mm", "A4", "")
	utf := pdf.UnicodeTranslatorFromDescriptor("")
	pdf.SetLeftMargin(22)
	pdf.SetTopMargin(8)
	pdf.SetFont("Helvetica", "B", 20)
	pdf.SetHeaderFuncMode(func() {

		//pdf.SetDrawColor(val, val, val)
		//pdf.SetTextColor(val, val, val)
		pdf.Image("logokn.png", 2, 7, 17, 17, false, "", 0, "")
		pdf.SetFont("Helvetica", "", 25)
		pdf.TransformBegin()
		pdf.TransformRotate(90, 12, 255)
		pdf.Text(12, 255, " Messdienerplan   "+utf(p.Titel))
		pdf.TransformEnd()
		pdf.SetFont("Helvetica", "", 10)
		for i, line := range footer {
			pdf.Text(23, float64(286+(i*5)), line)
		}
		pdf.Image("logo.png", 188, 280, 14, 14, false, "", 0, "")

	}, true)
	pdf.AddPage()
	pagew, pageh := pdf.GetPageSize()
	mleft, mright, _, mbottom := pdf.GetMargins()

	cols := []float64{25, 25, 40, pagew - mleft - mright - 25 - 25 - 40}
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

		g := m.Gottesdienst + "\n" + m.Ort + "\n" + m.InfoForPlan
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

	c.Redirect(http.StatusFound, "/minis")
}

func ZuordnenFinishHandler(c *gin.Context) {
	uid := c.Query("uid")
	pid := c.Query("pid")
	status := c.Query("status")

	uuuid, err := uuid.FromString(uid)
	puuid, err := uuid.FromString(pid)
	b, err := strconv.ParseBool(status)
	if b {
		err = planService.AddFinished(uuuid, puuid)

	} else {
		err = planService.RemoveFinished(uuuid, puuid)
	}

	if err != nil {
		log.Println(err)
		c.Status(http.StatusNoContent)
	} else {
		c.Status(http.StatusOK)
	}

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
		"role":     userService.GetRoleByID(miniService.AktUser),
		"payload":  data,
	})
}

func AdminHandler(c *gin.Context) {
	user, _ := userService.GetAllUser()
	c.HTML(http.StatusOK, "adminArea", gin.H{
		"title":    "Administration",
		"user":     user,
		"username": userService.GetUsernameByID(miniService.AktUser),
		"role":     userService.GetRoleByID(miniService.AktUser),
	})
}

func EinstellungenHandler(c *gin.Context) {

	c.HTML(http.StatusOK, "einstellungen", gin.H{
		"title":    "Einstellungen",
		"user":     userService.GetAktUser(),
		"username": userService.GetUsernameByID(miniService.AktUser),
		"role":     userService.GetRoleByID(miniService.AktUser),
	})
}

func MessenHandler(c *gin.Context) {
	var data *[]mongo.MesseModel
	if c.Query("show") == "all" {
		data, _ = messeService.GetAllMessen()
	} else {
		data, _ = messeService.GetAllMessenFromDate(time.Now().AddDate(0, 0, -7)) //Messen aktuell und der letzten 7 Tage
	}

	c.HTML(http.StatusOK, "messen", gin.H{
		"title":    "Messen",
		"payload":  data,
		"username": userService.GetUsernameByID(miniService.AktUser),
		"role":     userService.GetRoleByID(miniService.AktUser),
		"from":     time.Now().Format("2006-01-02"),
		"to":       time.Now().Format("2006-01-02"),
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
	l, err := time.LoadLocation("Europe/Berlin")

	fromDate := time.Date(from.Year(), from.Month(), from.Day(), 0, 0, 0, 0, l)
	toDate := time.Date(to.Year(), to.Month(), to.Day(), 23, 59, 59, 0, l)
	data, err := messeService.GetAllMessenFromToDate(fromDate, toDate)
	if err != nil {
		log.Println(err)
	}

	c.HTML(http.StatusFound, "/messen", gin.H{
		"title":    "Messen",
		"payload":  data,
		"username": userService.GetUsernameByID(miniService.AktUser),
		"role":     userService.GetRoleByID(miniService.AktUser),
		"from":     c.PostForm("from"),
		"to":       c.PostForm("to"),
	})
}

func MessdienerplanDeleteHandler(c *gin.Context) {
	idToSreach, err := uuid.FromString(c.Param("id"))
	if err != nil {
		c.Redirect(http.StatusFound, "/messdienerplan")
	}
	planService.DeletePlanById(idToSreach)
	c.Redirect(http.StatusFound, "/messdienerplan")
}

func MessdienerplanChangeHinweisHandler(c *gin.Context) {
	idToSreach, err := uuid.FromString(c.PostForm("uuid"))
	if err != nil {
		c.Redirect(http.StatusFound, "/messdienerplan")
		return
	}

	planService.UpdateHinweis(idToSreach, c.PostForm("hinweis"))
	if err != nil {
		log.Println(err)
	}
	c.Redirect(http.StatusFound, "/messdienerplan")
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

func UserDeleteHandler(c *gin.Context) {
	idToSreach := c.Param("id")
	userService.DeleteUserById(idToSreach)
	c.Redirect(http.StatusFound, "/adminArea")
}

func UserResetPassword(c *gin.Context) {
	idToSreach := c.Param("id")
	password := "123456"
	userService.AdminChangeUserPasswordById(idToSreach, password)
	c.Redirect(http.StatusFound, "/adminArea")
}

func EinstellungenChangeHandler(c *gin.Context) {
	footerLines := c.PostForm("footerText")
	lines := strings.Split(footerLines, "\n")
	userService.ChangeFooterText(lines)
	c.Redirect(http.StatusFound, "/einstellungen")
}

func PasswordChangeHandler(c *gin.Context) {

	passwordOld := c.PostForm("password_old")
	passwordNew := c.PostForm("password_new")
	passwordNewRepeat := c.PostForm("password_new_repeat")

	if passwordNew != passwordNewRepeat {
		log.Println("Passwords not match")
		c.Redirect(http.StatusFound, "/einstellungen")
	}
	err := userService.ChangeUserPasswordAktUser(passwordOld, passwordNew)
	if err != nil {
		log.Println(err)
	}
	c.Redirect(http.StatusFound, "/einstellungen")
}

func AddMessenHandler(c *gin.Context) {
	l, err := time.LoadLocation("Europe/Berlin")
	d, err := time.ParseInLocation("02.01.2006 15:04", c.PostForm("date"), l)
	g := c.PostForm("gottesdienst")
	t := c.PostForm("tag")
	b := c.PostForm("bemerkung")

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

func AddMessenFromKaplanHandler(c *gin.Context) {
	defer func(c *gin.Context) {
		if rec := recover(); rec != nil {
			c.Redirect(http.StatusFound, "/messen")
			return
		}
	}(c)

	type KaplanData struct {
		Datum        string `json:"Datum"`
		Uhrzeit      string `json:"Uhrzeit"`
		Bis          string `json:"bis"`
		Gottesdienst string `json:"Gottesdienst"`
		Zusatz       string `json:"Zusatz"`
		Ort          string `json:"Ort"`
		ID           string `json:"ID"`
		FaelltAus    bool   `json:"FaelltAus"`
	}

	var messen []KaplanData

	kirchenID := c.PostForm("kirchenID")
	anzahlTage := c.PostForm("anzahlTage")

	url := fmt.Sprintf("https://%s/get.asp?ref=%s&mode=T1&type=json&KG=%s&Days=%s", KaplanHost, KaplanSecret, kirchenID, anzahlTage)

	request, err := http.Get(url)
	if err != nil {
		log.Println(err)
		c.Redirect(http.StatusFound, "/messen")
		return
	}
	json.NewDecoder(request.Body).Decode(&messen)
	fmt.Println(url)
	for _, messe := range messen {
		uid := uuid.NewV4()
		location, _ := time.LoadLocation("Europe/Berlin")
		date, _ := time.ParseInLocation("1/2/2006 15:04", fmt.Sprintf("%s %s", messe.Datum, messe.Uhrzeit), location)

		if messe.Gottesdienst == "Sonntagvorabendmesse" {
			messe.Gottesdienst = "Vorabendmesse"
		}

		m := mongo.MesseModel{
			UUID:            uid,
			KaplanID:        messe.ID,
			Datum:           date,
			Ort:             messe.Ort,
			Gottesdienst:    messe.Gottesdienst,
			LiturgischerTag: "",
			Bemerkung:       messe.Zusatz,
			IsRelevant:      checkIfRelevant(messe.Gottesdienst, messe.Zusatz, date.Format("Mon"), date.Format("15:04"), messe.FaelltAus, messe.Ort),
		}
		messeService.CreateFromKaplan(&m)
	}
	c.Redirect(http.StatusFound, "/messen")

}

func AddMessenFromExcelHandler(c *gin.Context) {

	defer func(c *gin.Context) {
		if rec := recover(); rec != nil {
			c.Redirect(http.StatusFound, "/messen")
			return
		}
	}(c)

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.Redirect(http.StatusFound, "/messen")
		return
	}

	// Detect Fietype (not the best way, but it woks more or less)
	// Better Way with content type, but not shure if it is same for every excel file
	// Content-Type:[application/octet-stream]  => .xls
	// Content-Type:[application/vnd.openxmlformats-officedocument.spreadsheetml.sheet]  => .xlsx

	filename := strings.Split(header.Filename, ".")
	filetype := filename[len(filename)-1]

	l, err := time.LoadLocation("Europe/Berlin")

	switch filetype {
	case "xls":
		if xlFile, err := xls.OpenReader(file, "utf-8"); err == nil {
			if err != nil {
				c.Redirect(http.StatusFound, "/messen")
				return
			}
			for i := 0; i < xlFile.NumSheets(); i++ {
				sheet := xlFile.GetSheet(i)
				for j := 1; j <= int(sheet.MaxRow); j++ {
					row := sheet.Row(j)
					if row.Col(0) == "" {
						break
					}

					//u, err := time.ParseInLocation("15:04", row.Col(2), l)
					t, err := strconv.ParseFloat(row.Col(2), 32)
					if err != nil {
						log.Println("Could not Parse Time: ", err)
						c.Redirect(http.StatusFound, "/messen")
						return
					}

					s, err := time.ParseDuration("1s")
					u := timeFromExcelTime(t, true).Add(s) //Sekunde die Floating Point Fehler ausgleicht
					timeString := strings.ReplaceAll(row.Col(1), "00:00:00", fmt.Sprintf("%02d:%02d:00", u.Hour(), u.Minute()))
					d, err := time.ParseInLocation("2006-01-02T15:04:05Z", timeString, l)

					if err != nil {
						log.Println("Could not Parse Date: ", err)
						c.Redirect(http.StatusFound, "/messen")
						return
					}

					uid := uuid.NewV4()

					m := mongo.MesseModel{
						UUID:            uid,
						Datum:           d,
						Gottesdienst:    row.Col(3),
						LiturgischerTag: row.Col(5),
						Bemerkung:       row.Col(6),
						IsRelevant:      checkIfRelevant(row.Col(3), row.Col(6), d.Format("Mon"), d.Format("15:04"), false, ""),
					}

					messeService.Create(&m)
				}
			}
		}
	case "xlsx":
		if xlFile, err := excelize.OpenReader(file); err == nil {
			sheetName := xlFile.GetSheetName(0)
			rows, err := xlFile.GetRows(sheetName)
			if err != nil {
				c.Redirect(http.StatusFound, "/messen")
				return
			}
			for _, row := range rows {
				if row[0] == "" {
					break
				}

				s, err := time.ParseDuration("1s")
				//u := timeFromExcelTime(t, true).Add(s) //Sekunde die Floating Point Fehler ausgleicht
				d, err := strconv.ParseFloat(row[1], 64)
				t, err := strconv.ParseFloat(row[2], 64)
				date, err := excelize.ExcelDateToTime(d+t, false)
				date, err = time.ParseInLocation("2006-01-02T15:04:05Z", date.Format("2006-01-02T15:04:05Z"), l)
				//timeString := strings.ReplaceAll(row[1], "00:00:00", fmt.Sprintf("%02d:%02d:00", u.Hour(), u.Minute()))
				//d, err := time.ParseInLocation("2006-01-02T15:04:05Z", timeString, l)

				if err != nil {
					log.Println("Could not Parse Date: ", err)
					c.Redirect(http.StatusFound, "/messen")
					return
				}

				uid := uuid.NewV4()

				m := mongo.MesseModel{
					UUID:            uid,
					Datum:           date.Add(s),
					Gottesdienst:    row[3],
					LiturgischerTag: row[5],
					Bemerkung:       row[6],
					IsRelevant:      checkIfRelevant(row[3], row[6], date.Format("Mon"), date.Format("15:04"), false, ""),
				}

				messeService.Create(&m)
			}
		} else {
			c.Redirect(http.StatusFound, "/messen")
			return
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

func Update() {
	u, err := userService.GetAllUser()
	if err != nil {
		//TODO
	}
	for _, user := range *u {
		if uuid.Equal(uuid.Nil, user.PublicID) {
			userService.AddPublicID(user.UUID)
		}
	}
}

func Error404Handler(c *gin.Context) {
	c.HTML(http.StatusNotFound, "404.html", gin.H{})
}

// g = Gottesdienst, b = Bemerkung, t = Tag, u = Uhrzeit, o=ort
func checkIfRelevant(g string, b string, t string, u string, f bool, o string) bool {

	if o == "St. Barbara" {
		return false
	}

	if o == "St. Kamillus" {
		return false
	}

	if o == "St. Sebastian" {
		return false
	}

	if o == "Immaculata (Kapelle)" {
		return false
	}

	// If Fällt aus is true
	if f {
		return false
	}

	if strings.Contains(b, "panisch") {
		return false
	}

	if strings.Contains(g, "Rosenkranz") {
		return false
	}

	if strings.Contains(b, "ortugiesische") {
		return false
	}

	if strings.Contains(b, "talienische") {
		return false
	}

	if strings.Contains(b, "roatische") {
		return false
	}

	if strings.Contains(b, "amilische") {
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

	if strings.Contains(b, "St. Sebastian") {
		return false
	}

	if strings.Contains(b, "mit der Gemeinschaft der Frauen") {
		return false
	}

	if strings.Contains(g, "andacht") {
		return false
	}

	if strings.Contains(g, "Andacht") {
		return false
	}

	if strings.Contains(g, "Orgelstunde") {
		return false
	}

	if strings.Contains(g, "ökum.") {
		return false
	}

	if (t == "Fri") && (u == "09:30") && (g == "Messe" || g == "Festmesse") {
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

func GetFileContentType(out *os.File) (string, error) {

	// Only the first 512 bytes are used to sniff the content type.
	buffer := make([]byte, 512)

	_, err := out.Read(buffer)
	if err != nil {
		return "", err
	}

	// Use the net/http package's handy DectectContentType function. Always returns a valid
	// content-type by returning "application/octet-stream" if no others seemed to match.
	contentType := http.DetectContentType(buffer)

	return contentType, nil
}
