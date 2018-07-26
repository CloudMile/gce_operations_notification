package app

import (
	"context"
	"os"
	"strings"
	"time"

	"google.golang.org/appengine"
	"google.golang.org/appengine/mail"
)

// Mail for send mail
type Mail struct {
	Ctx           context.Context
	OperationType string
	Operations    []*OutputOperation
}

// Send will send mail
func (m *Mail) Send() {
	ctx := m.Ctx
	htmlBody := `<table style="width:100%"><tr><th>Target</th><th>StartAt</th><th>EndAt</th></tr>`
	dateString := "2006-01-02T15:04:05.000-07:00"
	loc, _ := time.LoadLocation(os.Getenv("TIMEZONE"))

	for _, operation := range m.Operations {
		startTime, _ := time.Parse(dateString, operation.StartTime)
		endTime, _ := time.Parse(dateString, operation.EndTime)
		linkString := removeUnusedString(operation.TargetLink)

		htmlBody = htmlBody + "<tr><td>" + linkString + "</td><td>" + startTime.In(loc).String() + "</td><td>" + endTime.In(loc).String() + "</td></tr>"
	}
	htmlBody = htmlBody + "</table>"

	msg := &mail.Message{
		Sender:   "noreply@" + appengine.AppID(ctx) + ".appspotmail.com",
		To:       strings.Split(os.Getenv("TO"), ","),
		Subject:  "[GCE] `" + m.OperationType + "` happened",
		HTMLBody: htmlBody,
	}
	mail.Send(ctx, msg)
}

func removeUnusedString(targetLinkString string) (outString string) {
	unusedStringList := []string{`https://www.googleapis.com/compute/v1/projects/`, `/zones/`, `/instances/`}
	replaceStringList := []string{`<b>`, `</b> (`, `) <b>`}

	for i, unusedString := range unusedStringList {
		targetLinkString = strings.Replace(targetLinkString, unusedString, replaceStringList[i], -1)
	}
	outString = targetLinkString + `</b>`
	return
}
