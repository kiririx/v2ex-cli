package main

import (
	"fmt"
	ui "github.com/gizak/termui/v3"
	"github.com/gizak/termui/v3/widgets"
	"github.com/gocolly/colly/v2"
	"log"
	"strings"
)

func main() {
	if err := ui.Init(); err != nil {
		log.Fatalf("failed to initialize termui: %v", err)
	}
	defer ui.Close()

	header := widgets.NewParagraph()
	header.Text = "Press q to quit, Press h or l to switch tabs"
	header.SetRect(0, 0, 50, 1)
	header.Border = false
	header.TextStyle.Bg = ui.ColorBlue

	p2 := widgets.NewParagraph()
	p2.Text = "Press q to quit\nPress h or l to switch tabs\n"
	p2.Title = "Keys"
	p2.SetRect(5, 5, 40, 15)
	p2.BorderStyle.Fg = ui.ColorYellow

	tabsText := GetTabs()
	tabpane := widgets.NewTabPane(tabsText...)
	tabpane.SetRect(0, 1, 1000, 4)
	tabpane.Border = true

	var currentContent *widgets.List
	renderTab := func() {
		// switch tabpane.ActiveTabIndex {
		// case 0:
		// 	ui.Render(GetTabContent(tabs[tabpane.ActiveTabIndex]))
		// }
		currentContent = GetTabContent(tabs[tabpane.ActiveTabIndex].Name)
		ui.Render(currentContent)
	}

	ui.Render(header, tabpane, p2)

	uiEvents := ui.PollEvents()

	for {
		e := <-uiEvents
		switch e.ID {
		case "q", "<C-c>":
			return
		case "<Left>":
			tabpane.FocusLeft()
			ui.Clear()
			ui.Render(header, tabpane)
			renderTab()
		case "<Right>":
			tabpane.FocusRight()
			ui.Clear()
			ui.Render(header, tabpane)
			renderTab()
		case "<Down>":
			currentContent.ScrollDown()
		}
	}
}

type Tab struct {
	Name string
	Text string
}

var tabs []Tab

func GetTabs() (rows []string) {
	tabs = make([]Tab, 0)
	defer func() {
		rows = make([]string, 0)
		for i, tab := range tabs {
			rows = append(rows, fmt.Sprintf("%d. %s", i+1, tab.Text))
		}
	}()
	c := colly.NewCollector()
	c.OnHTML("#Tabs", func(e *colly.HTMLElement) {
		e.ForEach("a", func(i int, e *colly.HTMLElement) {
			href := e.Attr("href")
			if strings.Contains(href, "=") {
				tabName := href[strings.Index(href, "=")+1:]
				tabs = append(tabs, Tab{Name: tabName, Text: e.Text})
			}
		})
	})
	c.Visit("https://www.v2ex.com/")
	return
}

func GetTabContent(tabName string) (l *widgets.List) {
	topics := make([]Topic, 0)
	defer func() {
		l = widgets.NewList()
		l.Title = "List"
		l.Rows = func() []string {
			rows := make([]string, 0)
			for i, topic := range topics {
				rows = append(rows, fmt.Sprintf("%d.%s", i+1, topic.Text))
			}
			return rows
		}()
		l.TextStyle = ui.NewStyle(ui.ColorGreen)
		l.WrapText = false
		l.SetRect(0, 5, 100, 100)
	}()
	c := colly.NewCollector()

	c.OnHTML("#Main", func(e1 *colly.HTMLElement) {
		e1.ForEach(".cell.item", func(i int, e2 *colly.HTMLElement) {
			topics = append(topics, Topic{
				Text: e2.ChildText(".topic-link"),
				Author: func() string {
					var v string
					e2.ForEachWithBreak("strong", func(i int, e3 *colly.HTMLElement) bool {
						v = e3.ChildText("a")
						return false
					})
					return v
				}(),
				Date:           e2.ChildText(".fade"),
				ReplyCount:     e2.ChildText(".count_livid"),
				FinalReplyUser: e2.ChildText(".last_bump"),
			})
		})
	})

	c.Visit("https://www.v2ex.com/?tab=" + tabName)
	return
}

type Topic struct {
	Text           string
	Author         string
	Date           string
	FinalReplyUser string
	ReplyCount     string
}
