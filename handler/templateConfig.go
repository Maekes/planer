package handler

import (
	"fmt"
	"html/template"
	"log"
	"strings"
	"time"

	"github.com/Maekes/planer/mongo"
	"github.com/foolin/goview"
	"github.com/foolin/goview/supports/ginview"
	"github.com/jinzhu/now"
	uuid "github.com/satori/go.uuid"
)

var C = goview.Config{
	Root:      "views",
	Extension: ".html",
	Master:    "layouts/master",
	Partials:  []string{},
	Funcs: template.FuncMap{
		"sub": func(a, b int) int {
			return a - b
		},
		"add": func(a, b int) int {
			return a + b
		},
		"copy": func() string {
			return time.Now().Format("2006")
		},
		"convert": func(u uuid.UUID) string {
			return u.String()
		},
		"getDate": func(t time.Time) string {
			return t.Format("02.01.2006")
		},
		"getTime": func(t time.Time) string {
			return t.Format("15:04")
		},
		"getDayTime": func(t time.Time) string {
			return t.Format("02.01.2006 - 15:04:05")
		},
		"getDay": func(t time.Time) string {
			return toGerman(t.Format("Mon"))
		},
		"getDayShort": func(t time.Time) string {
			return toGermanShort(t.Format("Mon"))
		},
		"getMiniNameFromUUID": func(u uuid.UUID) string {
			m, err := miniService.GetMiniByUUID(u)
			if err != nil {
				//TODO
			}
			return m.Vorname + " " + m.Nachname
		},
		"getMiniGruppeUUID": func(u uuid.UUID) string {
			m, err := miniService.GetMiniByUUID(u)
			if err != nil {
				//TODO
			}
			return m.Gruppe
		},
		"initialen": func(v string, n string) string {
			return string(v[0]) + string(n[0])
		},
		"countInPlan": func(m uuid.UUID, p uuid.UUID) int {
			plan, err := planService.GetPlanByUUID(p)
			if err != nil {
				//TODO
			}
			return messeService.CountMiniInMessen(plan.Von, plan.Bis, m)
		},
		"getMessen": func(f time.Time, t time.Time) string {
			var output []string
			messen, err := messeService.GetAllMessenThatAreRelevantFromToDate(f, t)
			if err != nil {
				//TODO
				fmt.Print("error in getMessen")
			}
			for _, messe := range *messen {
				output = append(output, toGermanShort(messe.Datum.Format("Mon")))
				output = append(output, messe.Datum.Format(" 02.01 - 15:04 - "))
				output = append(output, messe.Gottesdienst)
				output = append(output, "\n")
			}
			return strings.Join(output, "")
		},
		"getMessenDividedInWeeks": func(m *[]mongo.MesseModel) *[]Week {
			var w []Week

			me := *m
			if len(me) > 0 {
				start := me[0].Datum
				now.WeekStartDay = time.Saturday
				end := now.New(start).EndOfWeek()
				var whelper Week
				for _, messe := range me {

					if messe.Datum.Before(end) {
						whelper.Messen = append(whelper.Messen, messe)
					} else {
						w = append(w, whelper)

						whelper = Week{}
						start := messe.Datum
						end = now.New(start).EndOfWeek()
						whelper.Messen = append(whelper.Messen, messe)
					}
				}
				w = append(w, whelper)
			} else {
				log.Println("No Messen Found to Divide")
				//TODO
			}

			return &w
		},
	},
	DisableCache: true,
}

func GetTemplateConfig() *ginview.ViewEngine {
	return ginview.New(C)
}

type Week struct {
	Messen []mongo.MesseModel
}

func toGerman(d string) string {
	switch d {
	case "Mon":
		return "Montag"
	case "Tue":
		return "Dienstag"
	case "Wed":
		return "Mittwoch"
	case "Thu":
		return "Donerstag"
	case "Fri":
		return "Freitag"
	case "Sat":
		return "Samstag"
	case "Sun":
		return "Sonntag"
	default:
		return d
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
