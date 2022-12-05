package main

import (
	"context"
	"fmt"
	"net/http"
	"os"

	"github.com/gdamore/tcell/v2"
	"github.com/hasura/go-graphql-client"
	"github.com/rivo/tview"
)

func main() {
	app := tview.NewApplication()

	querySection := tview.NewTextArea()
	querySection.SetPlaceholder("Enter NRQL here...")
	querySection.SetText("SELECT * FROM Transaction", true)
	querySection.SetTitle("NRQL Query")

	resultsSection := tview.NewTextArea()
	resultsSection.SetBorder(true).SetTitle("Results")
	resultsSection.SetText("Results go here...", false)

	metaSection := tview.NewTextView()
	metaSection.SetDynamicColors(true)
	metaSection.SetBorder(true).SetTitle("Meta")

	querySection.SetTitle("Text Area").SetBorder(true)

	updateInfos := func() {
		fromRow, fromColumn, toRow, toColumn := querySection.GetCursor()
		if fromRow == toRow && fromColumn == toColumn {
			metaSection.SetText(fmt.Sprintf("Row: [yellow]%d[white], Column: [yellow]%d ", fromRow, fromColumn))
		} else {
			metaSection.SetText(fmt.Sprintf("[red]From[white] Row: [yellow]%d[white], Column: [yellow]%d[white] - [red]To[white] Row: [yellow]%d[white], To Column: [yellow]%d ", fromRow, fromColumn, toRow, toColumn))
		}
	}

	querySection.SetMovedFunc(updateInfos)
	updateInfos()

	flex := tview.NewFlex().
		AddItem(tview.NewFlex().SetDirection(tview.FlexRow).
			AddItem(querySection, 0, 1, true).
			AddItem(resultsSection, 0, 3, true).
			AddItem(metaSection, 3, 1, false), 0, 2, false).
		AddItem(tview.NewBox().SetBorder(true).SetTitle("Settings"), 20, 1, false)

	app.SetInputCapture(func(event *tcell.EventKey) *tcell.EventKey {
		//just looking around with tab
		if event.Key() == tcell.KeyTab {
			if app.GetFocus() == resultsSection {
				app.SetFocus(querySection)
			} else {
				app.SetFocus(resultsSection)
			}
			return nil
		}

		// option return runs query... would like this to be shift, but it didn't work
		if event.Key() == tcell.KeyEnter && (event.Modifiers()&tcell.ModAlt > 0) {
			result := runQuery()
			resultsSection.SetText(result, true)
			app.SetFocus(resultsSection)
		}
		return event
	})

	if err := app.SetRoot(flex, true).SetFocus(querySection).EnableMouse(true).Run(); err != nil {
		panic(err)
	}
}

func runQuery() string {
	client := graphql.NewClient("https://api.newrelic.com/graphql", nil)

	client = client.WithRequestModifier(setAuthHeader())

	var query struct {
		Actor struct {
			User struct {
				Name string
			}
		}
	}

	err := client.Query(context.Background(), &query, nil)
	if err != nil {
		panic(err)
	}
	return query.Actor.User.Name
}

func setAuthHeader() func(req *http.Request) {
	return func(req *http.Request) {
		req.Header.Add("api-key", os.Getenv("NR_API_KEY"))
	}
}
