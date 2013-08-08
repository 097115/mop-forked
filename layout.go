// Copyright (c) 2013 by Michael Dvorkin. All Rights Reserved.
// Use of this source code is governed by a MIT-style
// license that can be found in the LICENSE file.

package mop

import (
	`bytes`
	`fmt`
	`regexp`
	`strings`
	`text/template`
	`time`
)

const TotalColumns = 15

type Column struct {
	width	int
	title	string
}

type Layout struct {
	columns []Column
}

//-----------------------------------------------------------------------------
func (self *Layout) Initialize() *Layout {
	self.columns = []Column{
		{ -7, `Ticker` },
		{ 10, `Last` },
		{ 10, `Change` },
		{ 10, `Change%` },
		{ 10, `Open` },
		{ 10, `Low` },
		{ 10, `High` },
		{ 10, `52w Low` },
		{ 10, `52w High` },
		{ 11, `Volume` },
		{ 11, `AvgVolume` },
		{  9, `P/E` },
		{  9, `Dividend` },
		{  9, `Yield` },
		{ 11, `MktCap` },
	}

	return self
}

//-----------------------------------------------------------------------------
func (self *Layout) Market(market *Market) string {
	if ok, err := market.Ok(); !ok {
		return err
	}

	markup := `{{.Dow.name}}: {{.Dow.change}} ({{.Dow.percent}}) at {{.Dow.latest}}, {{.Sp500.name}}: {{.Sp500.change}} ({{.Sp500.percent}}) at {{.Sp500.latest}}, {{.Nasdaq.name}}: {{.Nasdaq.change}} ({{.Nasdaq.percent}}) at {{.Nasdaq.latest}}
{{.Advances.name}}: {{.Advances.nyse}} ({{.Advances.nysep}}) on NYSE and {{.Advances.nasdaq}} ({{.Advances.nasdaqp}}) on Nasdaq. {{.Declines.name}}: {{.Declines.nyse}} ({{.Declines.nysep}}) on NYSE and {{.Declines.nasdaq}} ({{.Declines.nasdaqp}}) on Nasdaq {{if .IsClosed}}<right>U.S. markets closed</right>{{end}}
New highs: {{.Highs.nyse}} on NYSE and {{.Highs.nasdaq}} on Nasdaq. New lows: {{.Lows.nyse}} on NYSE and {{.Lows.nasdaq}} on Nasdaq.`

	template, err := template.New(`market`).Parse(markup)
	if err != nil {
		panic(err)
	}

	buffer := new(bytes.Buffer)
	highlight(market.Dow, market.Sp500, market.Nasdaq)
	if err := template.Execute(buffer, market); err != nil {
		panic(err)
	}

	return buffer.String()
}

//-----------------------------------------------------------------------------
func (self *Layout) Quotes(quotes *Quotes) string {
	vars := struct {
		Now    string
		Header string
		Stocks []Stock
	}{
		time.Now().Format(`3:04:05pm PST`),
		self.Header(quotes.profile),
		self.prettify(quotes),
	}

	markup := `<right><white>{{.Now}}</></right>



{{.Header}}
{{range.Stocks}}{{if .Advancing}}<green>{{end}}{{.Ticker}}{{.LastTrade}}{{.Change}}{{.ChangePct}}{{.Open}}{{.Low}}{{.High}}{{.Low52}}{{.High52}}{{.Volume}}{{.AvgVolume}}{{.PeRatio}}{{.Dividend}}{{.Yield}}{{.MarketCap}}</>
{{end}}`
	//markup += fmt.Sprintf("[%v]", quotes.profile.Grouped)
	template, err := template.New(`quotes`).Parse(markup)
	if err != nil {
		panic(err)
	}

	buffer := new(bytes.Buffer)
	err = template.Execute(buffer, vars)
	if err != nil {
		panic(err)
	}

	return buffer.String()
}

//-----------------------------------------------------------------------------
func (self *Layout) Header(profile *Profile) string {
	str, selected_column := ``, profile.selected_column

	for i,col := range self.columns {
		arrow := arrow_for(i, profile)
		if i != selected_column {
			str += fmt.Sprintf(`%*s`, col.width, arrow + col.title)
		} else {
			str += fmt.Sprintf(`<r>%*s</r>`, col.width, arrow + col.title)
		}
	}

	return `<u>` + str + `</u>`
}

//-----------------------------------------------------------------------------
func (self *Layout) prettify(quotes *Quotes) []Stock {
	pretty := make([]Stock, len(quotes.stocks))

	for i, q := range quotes.stocks {
		pretty[i].Ticker    = pad(q.Ticker,                   self.columns[0].width)
		pretty[i].LastTrade = pad(currency(q.LastTrade),      self.columns[1].width)
		pretty[i].Change    = pad(currency(q.Change),         self.columns[2].width)
		pretty[i].ChangePct = pad(last(q.ChangePct),          self.columns[3].width)
		pretty[i].Open      = pad(currency(q.Open),           self.columns[4].width)
		pretty[i].Low       = pad(currency(q.Low),            self.columns[5].width)
		pretty[i].High      = pad(currency(q.High),           self.columns[6].width)
		pretty[i].Low52     = pad(currency(q.Low52),          self.columns[7].width)
		pretty[i].High52    = pad(currency(q.High52),         self.columns[8].width)
		pretty[i].Volume    = pad(q.Volume,                   self.columns[9].width)
		pretty[i].AvgVolume = pad(q.AvgVolume,                self.columns[10].width)
		pretty[i].PeRatio   = pad(blank(q.PeRatioX),          self.columns[11].width)
		pretty[i].Dividend  = pad(blank_currency(q.Dividend), self.columns[12].width)
		pretty[i].Yield     = pad(percent(q.Yield),           self.columns[13].width)
		pretty[i].MarketCap = pad(currency(q.MarketCapX),     self.columns[14].width)
		pretty[i].Advancing = q.Advancing
	}

	profile := quotes.profile
	new(Sorter).Initialize(profile).SortByCurrentColumn(pretty)
	//
	// Group stocks by advancing/declining unless sorted by Chanage or Change%
	// in which case the grouping is done already.
	//
	if profile.Grouped && (profile.SortColumn < 2 || profile.SortColumn > 3) {
		pretty = group(pretty)
	}

	return pretty
}

//-----------------------------------------------------------------------------
func highlight(collections ...map[string]string) {
	for _, collection := range collections {
		if collection[`change`][0:1] != `-` {
			collection[`change`] = `<green>` + collection[`change`] + `</>`
		}
	}
}

//-----------------------------------------------------------------------------
func group(stocks []Stock) []Stock {
	grouped := make([]Stock, len(stocks))
	current := 0

	for _,stock := range stocks {
		if strings.Index(stock.Change, "-") == -1 {
			grouped[current] = stock
			current++
		}
	}
	for _,stock := range stocks {
		if strings.Index(stock.Change, "-") != -1 {
			grouped[current] = stock
			current++
		}
	}

	return grouped
}

//-----------------------------------------------------------------------------
func arrow_for(column int, profile *Profile) string {
	if column == profile.SortColumn {
		if profile.Ascending {
			return string('\U00002191')
		}
		return string('\U00002193')
	}
	return ``
}

//-----------------------------------------------------------------------------
func blank(str string) string {
	if len(str) == 3 && str[0:3] == `N/A` {
		return `-`
	}

	return str
}

//-----------------------------------------------------------------------------
func blank_currency(str string) string {
	if str == `0.00` {
		return `-`
	}

	return currency(str)
}

//-----------------------------------------------------------------------------
func last(str string) string {
	if len(str) >= 6 && str[0:6] != `N/A - ` {
		return str
	}

	return str[6:]
}

//-----------------------------------------------------------------------------
func currency(str string) string {
	if str == `N/A` {
		return `-`
	}
	if sign := str[0:1]; sign == `+` || sign == `-` {
		return sign + `$` + str[1:]
	}

	return `$` + str
}

//-----------------------------------------------------------------------------
func percent(str string) string {
	if str == `N/A` {
		return `-`
	}

	return str + `%`
}

//-----------------------------------------------------------------------------
func pad(str string, width int) string {
	re := regexp.MustCompile(`(\.\d+)[MB]?$`)
	match := re.FindStringSubmatch(str)
	if len(match) > 0 {
		switch len(match[1]) {
		case 2:
			str = strings.Replace(str, match[1], match[1] + `0`, 1)
		case 4, 5:
			str = strings.Replace(str, match[1], match[1][0:3], 1)
		}
	}

	return fmt.Sprintf(`%*s`, width, str)
}