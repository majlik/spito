package parser

import (
	"encoding/xml"
	"fmt"
	"os"
	"strings"
)

// IOF XML 3.0 structs

type ResultList struct {
	XMLName     xml.Name      `xml:"ResultList"`
	Event       Event         `xml:"Event"`
	ClassResult []ClassResult `xml:"ClassResult"`
}

type Event struct {
	Name string `xml:"Name"`
}

type ClassResult struct {
	Class        ClassInfo      `xml:"Class"`
	Course       CourseInfo     `xml:"Course"`
	PersonResult []PersonResult `xml:"PersonResult"`
}

type ClassInfo struct {
	ID   string `xml:"Id"`
	Name string `xml:"Name"`
}

type CourseInfo struct {
	Name   string  `xml:"Name"`
	Length float64 `xml:"Length"`
	Climb  float64 `xml:"Climb"`
}

type PersonResult struct {
	Person       Person       `xml:"Person"`
	Organisation Organisation `xml:"Organisation"`
	Result       Result       `xml:"Result"`
}

type Person struct {
	Name PersonName `xml:"Name"`
}

type PersonName struct {
	Family string `xml:"Family"`
	Given  string `xml:"Given"`
}

type Organisation struct {
	Name      string `xml:"Name"`
	ShortName string `xml:"ShortName"`
}

type Result struct {
	StartTime  string      `xml:"StartTime"`
	FinishTime string      `xml:"FinishTime"`
	Time       int         `xml:"Time"`
	TimeBehind int         `xml:"TimeBehind"`
	Position   int         `xml:"Position"`
	Status     string      `xml:"Status"`
	Course     CourseInfo  `xml:"Course"`
	SplitTimes []SplitTime `xml:"SplitTime"`
}

type SplitTime struct {
	ControlCode string `xml:"ControlCode"`
	Time        int    `xml:"Time"`   // cumulative seconds from start; -1 if missing
	Status      string `xml:"status,attr"` // "" = normal, "Additional" = extra control (ignore)
}

// ---- Processed data structures ----

type Competitor struct {
	Name        string // "Lastname Firstname"
	ShortName   string // "L. Firstname" for legend
	Club        string
	Position    int
	Status      string // OK, DNS, DNF, MP, etc.
	TotalTime   int    // seconds
	TimeBehind  int    // seconds behind winner
	Splits      []int  // cumulative time at each control index (len = numControls+1 for finish); -1 = missing
}

type ClassData struct {
	EventName   string
	ClassName   string
	CourseName  string
	CourseLen   float64
	CourseClimb float64
	Controls    []string // ordered control codes (S implicit, then splits, then F implicit)
	Competitors []*Competitor
}

// ParseFile parses the IOF XML file and returns all ClassData indexed by class name.
func ParseFile(path string) (map[string]*ClassData, []string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot read file: %w", err)
	}

	// Strip namespace to simplify parsing
	content := strings.ReplaceAll(string(data), ` xmlns="http://www.orienteering.org/datastandard/3.0"`, "")
	content = strings.ReplaceAll(content, ` xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"`, "")

	var rl ResultList
	if err := xml.Unmarshal([]byte(content), &rl); err != nil {
		return nil, nil, fmt.Errorf("XML parse error: %w", err)
	}

	classes := make(map[string]*ClassData)
	var classOrder []string

	for _, cr := range rl.ClassResult {
		className := cr.Class.Name
		if _, exists := classes[className]; !exists {
			classOrder = append(classOrder, className)
		}

		cd := &ClassData{
			EventName:   rl.Event.Name,
			ClassName:   className,
			CourseName:  cr.Course.Name,
			CourseLen:   cr.Course.Length,
			CourseClimb: cr.Course.Climb,
		}

		// Build control sequence from first OK result with most normal (non-Additional) splits
		var controls []string
		maxSplits := 0
		for _, pr := range cr.PersonResult {
			if pr.Result.Status == "OK" {
				var normal []SplitTime
				for _, st := range pr.Result.SplitTimes {
					if st.Status != "Additional" {
						normal = append(normal, st)
					}
				}
				if len(normal) > maxSplits {
					maxSplits = len(normal)
					controls = make([]string, len(normal))
					for i, st := range normal {
						controls[i] = st.ControlCode
					}
				}
			}
		}
		// If no OK result, use first result with most normal splits
		if len(controls) == 0 {
			for _, pr := range cr.PersonResult {
				var normal []SplitTime
				for _, st := range pr.Result.SplitTimes {
					if st.Status != "Additional" {
						normal = append(normal, st)
					}
				}
				if len(normal) > maxSplits {
					maxSplits = len(normal)
					controls = make([]string, len(normal))
					for i, st := range normal {
						controls[i] = st.ControlCode
					}
				}
			}
		}
		cd.Controls = controls
		numControls := len(controls)

		for _, pr := range cr.PersonResult {
			r := pr.Result
			// Skip DNS
			if r.Status == "DNS" {
				continue
			}

			// Use course from result if class-level course is missing
			if cd.CourseName == "" && r.Course.Name != "" {
				cd.CourseName = r.Course.Name
				cd.CourseLen = r.Course.Length
				cd.CourseClimb = r.Course.Climb
			}

			fullName := strings.TrimSpace(pr.Person.Name.Family + " " + pr.Person.Name.Given)
			// Short name for legend: "Lastname F."
			shortName := pr.Person.Name.Family
			if pr.Person.Name.Given != "" {
				shortName += " " + string([]rune(pr.Person.Name.Given)[0:1]) + "."
			}

			comp := &Competitor{
				Name:       fullName,
				ShortName:  shortName,
				Club:       pr.Organisation.ShortName,
				Position:   r.Position,
				Status:     r.Status,
				TotalTime:  r.Time,
				TimeBehind: r.TimeBehind,
				Splits:     make([]int, numControls+1), // [0..numControls-1]=controls, [numControls]=finish
			}

			// Initialize all splits as -1 (missing)
			for i := range comp.Splits {
				comp.Splits[i] = -1
			}

			// Map splits by position, skipping Additional controls
			pos := 0
			for _, st := range r.SplitTimes {
				if st.Status == "Additional" {
					continue
				}
				if pos < numControls {
					comp.Splits[pos] = st.Time
				}
				pos++
			}
			// Finish time
			comp.Splits[numControls] = r.Time

			cd.Competitors = append(cd.Competitors, comp)
		}

		// Sort: OK first by position, then DNF/MP by total time or partial splits
		sortCompetitors(cd.Competitors)

		if existing, ok := classes[className]; ok {
			// Merge if same class appears in multiple ClassResult blocks
			existing.Competitors = append(existing.Competitors, cd.Competitors...)
		} else {
			classes[className] = cd
		}
	}

	return classes, classOrder, nil
}

func sortCompetitors(comps []*Competitor) {
	// Stable sort: OK by position, then DNF/MP/etc by their position or end
	n := len(comps)
	for i := 0; i < n-1; i++ {
		for j := i + 1; j < n; j++ {
			if rankOrder(comps[i]) > rankOrder(comps[j]) {
				comps[i], comps[j] = comps[j], comps[i]
			} else if rankOrder(comps[i]) == rankOrder(comps[j]) {
				// Same status group: sort by position (for OK) or total time
				if comps[i].Position > comps[j].Position && comps[j].Position > 0 {
					comps[i], comps[j] = comps[j], comps[i]
				}
			}
		}
	}
}

func rankOrder(c *Competitor) int {
	switch c.Status {
	case "OK":
		return 0
	case "DNF", "MP":
		return 1
	default:
		return 2
	}
}

// FormatTime formats seconds as M:SS or MM:SS or H:MM:SS
func FormatTime(secs int) string {
	if secs < 0 {
		return "-----"
	}
	h := secs / 3600
	m := (secs % 3600) / 60
	s := secs % 60
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%d:%02d", m, s)
}

// FormatTimeDiff formats a time difference with +/- sign
func FormatTimeDiff(secs int) string {
	if secs < 0 {
		return fmt.Sprintf("-%s", FormatTime(-secs))
	}
	return fmt.Sprintf("+%s", FormatTime(secs))
}
