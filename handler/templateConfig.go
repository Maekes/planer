package handler

import (
	"html/template"
	"time"

	"github.com/foolin/goview"
	"github.com/foolin/goview/supports/ginview"
	uuid "github.com/satori/go.uuid"
)

func GetTemplateConfig() *ginview.ViewEngine {
	return ginview.New(goview.Config{

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
			"getDay": func(t time.Time) string {
				return toGerman(t.Format("Mon"))
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
		},
		DisableCache: true,
	})
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
