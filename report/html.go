package report

import (
	"fmt"
	"html"
	"math"
	"os"
	"sort"
	"strings"

	"splito/parser"
)

// color palette — one per competitor (cycles if more than len)
var palette = []string{
	"#e6194b", "#3cb44b", "#4363d8", "#f58231", "#911eb4",
	"#42d4f4", "#f032e6", "#bfef45", "#fabed4", "#469990",
	"#dcbeff", "#9a6324", "#800000", "#aaffc3", "#808000",
	"#ffd8b1", "#000075", "#a9a9a9", "#000000", "#4dabf7",
}

func color(i int) string { return palette[i%len(palette)] }

// classAnchor converts a class name to a safe HTML id/anchor string.
func classAnchor(name string) string {
	r := strings.Map(func(r rune) rune {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' {
			return r
		}
		return '_'
	}, name)
	return "class_" + r
}

// GenerateHTML produces a self-contained UTF-8 HTML file for a single class.
// mode: "splits" = table only, "timelost" = chart only, "all" = both.
func GenerateHTML(cd *parser.ClassData, outPath string, mode string) error {
	var b strings.Builder
	writeHTMLHeader(&b, cd.ClassName+" — "+cd.EventName)
	writeClassSection(&b, cd, mode, false)
	b.WriteString("</body>\n</html>\n")
	return os.WriteFile(outPath, []byte(b.String()), 0644)
}

// GenerateAllClassesHTML produces a single UTF-8 HTML file with all classes,
// a clickable TOC at the top, and one section per class.
func GenerateAllClassesHTML(classes map[string]*parser.ClassData, classOrder []string, eventName string, outPath string, mode string) error {
	var b strings.Builder

	title := "All Classes"
	if eventName != "" {
		title = eventName + " — All Classes"
	}
	writeHTMLHeader(&b, title)

	// ---- TOC ----
	fmt.Fprintf(&b, "<h1>%s</h1>\n", html.EscapeString(title))
	b.WriteString("<nav class=\"toc\">\n<h2 class=\"toc-title\">Categories</h2>\n<ul class=\"toc-list\">\n")
	for _, name := range classOrder {
		cd := classes[name]
		okCount := 0
		for _, c := range cd.Competitors {
			if c.Status == "OK" {
				okCount++
			}
		}
		anchor := classAnchor(name)
		fmt.Fprintf(&b, "<li><a href=\"#%s\">%s</a> <span class=\"toc-meta\">%d controls &bull; %d finishers</span></li>\n",
			anchor, html.EscapeString(name), len(cd.Controls), okCount)
	}
	b.WriteString("</ul>\n</nav>\n")

	// ---- One section per class ----
	for _, name := range classOrder {
		cd := classes[name]
		anchor := classAnchor(name)
		fmt.Fprintf(&b, "<section id=\"%s\" class=\"class-section\">\n", anchor)
		fmt.Fprintf(&b, "<a href=\"#top\" class=\"back-to-top\">&#8593; Back to top</a>\n")
		writeClassSection(&b, cd, mode, true)
		b.WriteString("</section>\n<hr class=\"class-sep\">\n")
	}

	b.WriteString("</body>\n</html>\n")
	return os.WriteFile(outPath, []byte(b.String()), 0644)
}

// writeHTMLHeader writes the <!DOCTYPE>, <head>, and opens <body>.
func writeHTMLHeader(b *strings.Builder, title string) {
	b.WriteString(`<!DOCTYPE html>
<html lang="en" id="top">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>`)
	b.WriteString(html.EscapeString(title))
	b.WriteString(`</title>
<style>
*{box-sizing:border-box;margin:0;padding:0}
body{font-family:'Segoe UI',Arial,sans-serif;font-size:13px;background:#f4f4f4;color:#222;padding:16px}
h1{font-size:1.4rem;margin-bottom:8px}
h2{font-size:1.05rem;margin:20px 0 8px;color:#444}
h2.class-title{font-size:1.2rem;color:#2c3e50;margin-bottom:4px}
h3{font-size:.95rem;margin:16px 0 6px;color:#555}
.meta{font-size:.85rem;color:#666;margin-bottom:12px}
/* ---- TOC ---- */
nav.toc{background:#fff;border:1px solid #ddd;border-radius:4px;padding:12px 16px;display:inline-block;margin-bottom:20px;max-width:100%}
h2.toc-title{font-size:1rem;color:#2c3e50;margin-bottom:8px}
ul.toc-list{list-style:none;column-count:4;column-gap:24px}
ul.toc-list li{margin-bottom:4px;break-inside:avoid}
ul.toc-list a{text-decoration:none;font-weight:600;color:#2980b9}
ul.toc-list a:hover{text-decoration:underline}
.toc-meta{font-size:11px;color:#999;margin-left:4px}
/* ---- class section ---- */
.class-section{padding-top:8px}
hr.class-sep{border:none;border-top:2px solid #ddd;margin:28px 0}
.back-to-top{font-size:11px;color:#aaa;text-decoration:none;float:right;margin-top:4px}
.back-to-top:hover{color:#2980b9}
/* ---- splits table ---- */
.wrap-table{overflow-x:auto;width:100%}
table.splits{border-collapse:collapse;white-space:nowrap;font-size:12px}
table.splits th,table.splits td{border:1px solid #ccc;padding:2px 5px;vertical-align:top}
table.splits thead th{background:#2c3e50;color:#fff;text-align:center;font-weight:600}
table.splits thead th.ctrl{font-size:11px;min-width:56px}
table.splits tbody tr:nth-child(even){background:#f9f9f9}
table.splits tbody tr:hover{background:#eaf4fb}
td.name{font-weight:600;white-space:nowrap}
td.club{color:#555;font-size:11px;white-space:nowrap}
td.pos{text-align:center;font-weight:700;color:#fff;background:#2c3e50;width:28px}
td.pos.dnf{background:#999}
td.sc{text-align:right;padding:1px 4px}
.sc-cumul{font-weight:600;font-size:12px}
.sc-leg{font-size:11px;color:#555}
.rank{font-size:10px;color:#888;margin-left:2px}
.rank.r1{color:#c0392b;font-weight:700}
.rank.r2{color:#e67e22;font-weight:700}
.rank.r3{color:#8e6b00;font-weight:700}
.missing{color:#bbb;font-size:11px}
/* ---- chart ---- */
.chart-container{background:#fff;border:1px solid #ddd;border-radius:4px;padding:12px;overflow-x:auto;margin-top:8px}
svg text{font-family:'Segoe UI',Arial,sans-serif}
@media(max-width:800px){ul.toc-list{column-count:2}}
@media(max-width:480px){ul.toc-list{column-count:1}}
</style>
</head>
<body>
`)
}

// writeClassSection writes the title, meta, splits table, and/or chart for one class.
// If subheading is true, uses h2.class-title instead of h1.
func writeClassSection(b *strings.Builder, cd *parser.ClassData, mode string, subheading bool) {
	numControls := len(cd.Controls)
	numCols := numControls + 1

	var okComps []*parser.Competitor
	for _, c := range cd.Competitors {
		if c.Status == "OK" {
			okComps = append(okComps, c)
		}
	}

	// Per-control rankings
	type rankMap = [][]int
	cumulRank := make(rankMap, numCols)
	legRank := make(rankMap, numCols)

	for ci := 0; ci < numCols; ci++ {
		cumulRank[ci] = make([]int, len(cd.Competitors))
		legRank[ci] = make([]int, len(cd.Competitors))

		type idxTime struct{ idx, t int }
		var cumulTimes []idxTime
		for i, c := range cd.Competitors {
			if c.Status == "OK" && c.Splits[ci] > 0 {
				cumulTimes = append(cumulTimes, idxTime{i, c.Splits[ci]})
			}
		}
		sort.Slice(cumulTimes, func(a, b int) bool { return cumulTimes[a].t < cumulTimes[b].t })
		for rank, it := range cumulTimes {
			cumulRank[ci][it.idx] = rank + 1
		}

		var legTimes []idxTime
		for i, c := range cd.Competitors {
			if c.Status == "OK" {
				leg := legTime(c, ci)
				if leg > 0 {
					legTimes = append(legTimes, idxTime{i, leg})
				}
			}
		}
		sort.Slice(legTimes, func(a, b int) bool { return legTimes[a].t < legTimes[b].t })
		for rank, it := range legTimes {
			legRank[ci][it.idx] = rank + 1
		}
	}

	// Best cumulative per control
	bestCumul := make([]int, numCols)
	for ci := 0; ci < numCols; ci++ {
		bestCumul[ci] = math.MaxInt32
		for _, c := range okComps {
			if c.Splits[ci] > 0 && c.Splits[ci] < bestCumul[ci] {
				bestCumul[ci] = c.Splits[ci]
			}
		}
		if bestCumul[ci] == math.MaxInt32 {
			bestCumul[ci] = 0
		}
	}

	// Title
	if subheading {
		fmt.Fprintf(b, "<h2 class=\"class-title\">%s</h2>\n", html.EscapeString(cd.ClassName))
	} else {
		fmt.Fprintf(b, "<h1>%s &mdash; %s</h1>\n", html.EscapeString(cd.ClassName), html.EscapeString(cd.EventName))
	}

	// Meta
	var meta []string
	if cd.EventName != "" && subheading {
		meta = append(meta, html.EscapeString(cd.EventName))
	}
	if cd.CourseName != "" {
		meta = append(meta, html.EscapeString(cd.CourseName))
	}
	if cd.CourseLen > 0 {
		meta = append(meta, fmt.Sprintf("%.1f&thinsp;km", cd.CourseLen/1000))
	}
	if cd.CourseClimb > 0 {
		meta = append(meta, fmt.Sprintf("%.0f&thinsp;m climb", cd.CourseClimb))
	}
	if len(meta) > 0 {
		fmt.Fprintf(b, "<p class=\"meta\">%s</p>\n", strings.Join(meta, " &bull; "))
	}

	showSplits := mode == "splits" || mode == "all"
	showChart := mode == "timelost" || mode == "all"

	// ======== SPLITS TABLE ========
	if showSplits {
		b.WriteString("<h3>Split Times</h3>\n<div class=\"wrap-table\">\n")
		b.WriteString("<table class=\"splits\">\n<thead><tr>")
		b.WriteString("<th>#</th><th>Name</th><th>Club</th>")
		for ci, ctrl := range cd.Controls {
			fmt.Fprintf(b, "<th class=\"ctrl\">%d<br><small>%s</small></th>", ci+1, html.EscapeString(ctrl))
		}
		b.WriteString("<th class=\"ctrl\">F</th>")
		b.WriteString("</tr></thead>\n<tbody>\n")

		for compIdx, c := range cd.Competitors {
			posStr := ""
			posClass := "pos"
			if c.Status == "OK" {
				posStr = fmt.Sprintf("%d", c.Position)
			} else {
				posStr = c.Status
				posClass = "pos dnf"
			}
			fmt.Fprintf(b, "<tr><td class=\"%s\">%s</td>", posClass, html.EscapeString(posStr))
			fmt.Fprintf(b, "<td class=\"name\">%s</td>", html.EscapeString(c.Name))
			fmt.Fprintf(b, "<td class=\"club\">%s</td>", html.EscapeString(c.Club))

			for ci := 0; ci < numCols; ci++ {
				cumul := c.Splits[ci]
				leg := legTime(c, ci)
				var cumulStr, legStr string

				if cumul <= 0 {
					cumulStr = `<span class="missing">---</span>`
					legStr = `<span class="missing">---</span>`
				} else {
					cr := cumulRank[ci][compIdx]
					crClass := rankClass(cr)
					cumulStr = fmt.Sprintf(`<span class="sc-cumul">%s</span><span class="rank %s">(%d)</span>`,
						parser.FormatTime(cumul), crClass, cr)
					if leg <= 0 {
						legStr = `<span class="missing">---</span>`
					} else {
						lr := legRank[ci][compIdx]
						lrClass := rankClass(lr)
						legStr = fmt.Sprintf(`<span class="sc-leg">%s</span><span class="rank %s">(%d)</span>`,
							parser.FormatTime(leg), lrClass, lr)
					}
				}
				fmt.Fprintf(b, "<td class=\"sc\">%s<br>%s</td>", cumulStr, legStr)
			}
			b.WriteString("</tr>\n")
		}
		b.WriteString("</tbody>\n</table>\n</div>\n")
	}

	// ======== TIME-LOST CHART ========
	if showChart {
		b.WriteString("<h3>Time Lost (behind leader)</h3>\n")
		b.WriteString("<div class=\"chart-container\">\n")
		writeSVGChart(b, cd, okComps, bestCumul, numCols)
		b.WriteString("</div>\n")
	}
}

// legTime returns the leg split time for competitor c at control index ci.
func legTime(c *parser.Competitor, ci int) int {
	if c.Splits[ci] <= 0 {
		return -1
	}
	if ci == 0 {
		return c.Splits[0]
	}
	prev := c.Splits[ci-1]
	if prev <= 0 {
		return -1
	}
	return c.Splits[ci] - prev
}

func rankClass(r int) string {
	switch r {
	case 1:
		return "r1"
	case 2:
		return "r2"
	case 3:
		return "r3"
	default:
		return ""
	}
}

// writeSVGChart draws a WinSplits-style time-behind-leader chart as inline SVG.
func writeSVGChart(b *strings.Builder, cd *parser.ClassData, okComps []*parser.Competitor, bestCumul []int, numCols int) {
	if len(okComps) == 0 {
		b.WriteString("<p>No finishers to chart.</p>")
		return
	}

	const (
		leftMargin   = 160.0
		rightMargin  = 20.0
		topMargin    = 40.0
		bottomMargin = 50.0
		colWidth     = 56.0
		rowHeight    = 26.0
		dotR         = 4.0
	)

	maxBehind := 0
	for _, c := range okComps {
		for ci := 0; ci < numCols; ci++ {
			if c.Splits[ci] > 0 && bestCumul[ci] > 0 {
				behind := c.Splits[ci] - bestCumul[ci]
				if behind > maxBehind {
					maxBehind = behind
				}
			}
		}
	}
	if maxBehind == 0 {
		maxBehind = 60
	}
	yInterval := niceInterval(maxBehind, 8)
	yMax := ((maxBehind / yInterval) + 1) * yInterval

	plotH := float64(len(okComps)) * rowHeight
	if plotH < 200 {
		plotH = 200
	}
	pixPerSec := plotH / float64(yMax)

	svgW := leftMargin + float64(numCols)*colWidth + rightMargin
	svgH := topMargin + plotH + bottomMargin

	fmt.Fprintf(b, `<svg xmlns="http://www.w3.org/2000/svg" width="%.0f" height="%.0f" viewBox="0 0 %.0f %.0f">`,
		svgW, svgH, svgW, svgH)
	b.WriteString("\n")
	fmt.Fprintf(b, `<rect width="%.0f" height="%.0f" fill="#fff"/>`, svgW, svgH)
	b.WriteString("\n")

	// Y grid + labels
	for t := 0; t <= yMax; t += yInterval {
		y := topMargin + float64(t)*pixPerSec
		fmt.Fprintf(b, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#e0e0e0" stroke-width="1"/>`,
			leftMargin, y, leftMargin+float64(numCols)*colWidth, y)
		b.WriteString("\n")
		fmt.Fprintf(b, `<text x="%.1f" y="%.1f" text-anchor="end" font-size="11" fill="#888">+%s</text>`,
			leftMargin-4, y+4, parser.FormatTime(t))
		b.WriteString("\n")
	}

	// X axis: control labels + vertical grid
	for ci := 0; ci < numCols; ci++ {
		x := leftMargin + float64(ci)*colWidth + colWidth/2
		var label string
		if ci < len(cd.Controls) {
			label = fmt.Sprintf("%d/%s", ci+1, cd.Controls[ci])
		} else {
			label = "F"
		}
		fmt.Fprintf(b, `<text x="%.1f" y="%.1f" text-anchor="middle" font-size="11" fill="#555">%s</text>`,
			x, topMargin-8, html.EscapeString(label))
		b.WriteString("\n")
		fmt.Fprintf(b, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#e8e8e8" stroke-width="1"/>`,
			x, topMargin, x, topMargin+plotH)
		b.WriteString("\n")
	}

	// Competitor lines + dots
	for ci2, c := range okComps {
		col := color(ci2)
		var pts []string
		for ci := 0; ci < numCols; ci++ {
			if c.Splits[ci] <= 0 || bestCumul[ci] <= 0 {
				continue
			}
			behind := c.Splits[ci] - bestCumul[ci]
			x := leftMargin + float64(ci)*colWidth + colWidth/2
			y := topMargin + float64(behind)*pixPerSec
			pts = append(pts, fmt.Sprintf("%.1f,%.1f", x, y))
		}
		if len(pts) >= 2 {
			fmt.Fprintf(b, `<polyline points="%s" fill="none" stroke="%s" stroke-width="2" opacity="0.85"/>`,
				strings.Join(pts, " "), col)
			b.WriteString("\n")
		}
		for ci := 0; ci < numCols; ci++ {
			if c.Splits[ci] <= 0 || bestCumul[ci] <= 0 {
				continue
			}
			behind := c.Splits[ci] - bestCumul[ci]
			x := leftMargin + float64(ci)*colWidth + colWidth/2
			y := topMargin + float64(behind)*pixPerSec
			fmt.Fprintf(b, `<circle cx="%.1f" cy="%.1f" r="%.1f" fill="%s"/>`, x, y, dotR, col)
			b.WriteString("\n")
		}
	}

	// Legend at bottom
	legendY := topMargin + plotH + 16
	legendX := leftMargin
	cols4 := (svgW - leftMargin - rightMargin) / 4
	for i, c := range okComps {
		col := color(i)
		x := legendX + float64(i%4)*cols4
		y := legendY + float64(i/4)*16
		fmt.Fprintf(b, `<rect x="%.1f" y="%.1f" width="12" height="4" fill="%s"/>`, x, y+6, col)
		fmt.Fprintf(b, `<text x="%.1f" y="%.1f" font-size="11" fill="#333">%s</text>`,
			x+16, y+11, html.EscapeString(c.ShortName))
		b.WriteString("\n")
	}

	b.WriteString("</svg>\n")
}

func niceInterval(maxVal, targetSteps int) int {
	if maxVal <= 0 {
		return 60
	}
	raw := maxVal / targetSteps
	candidates := []int{10, 15, 20, 30, 60, 90, 120, 180, 300, 600}
	for _, iv := range candidates {
		if iv >= raw {
			return iv
		}
	}
	return ((raw / 60) + 1) * 60
}
