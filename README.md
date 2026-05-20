<img height="250" alt="splito" src="https://github.com/user-attachments/assets/c2fe6021-9c55-4142-9f09-ba7a28dc1f4e" />

# splito — Orienteering Results Viewer

A zero-dependency CLI tool that reads [IOF XML 3.0](https://orienteering.sport/iof/it/data-standard-3-0/) result exports (e.g. from MeOS) and generates self-contained UTF-8 HTML reports with:

- **Splits table** — cumulative time + overall rank, leg split + rank on leg; top 3 positions coloured
- **Time-lost chart** — inline SVG WinSplits-style polyline chart (time behind leader at each control)
- **Multi-class report** — all classes in one file with a clickable TOC register at the top

No external libraries. No internet connection needed. Output is a single `.html` file you can open in any browser or share by email.

---

## Repository structure

```
splito/
├── main.go            ← CLI entry point, flag handling
├── go.mod             ← Go module definition (module splito, go 1.21)
├── parser/
│   └── iof.go         ← IOF XML 3.0 parser
└── report/
    └── html.go        ← HTML + SVG report generator
```

**What to commit:**
- `main.go`, `go.mod`, `go.sum`  
  *(rename the directory from `orio` to `splito` before pushing)*
- `parser/iof.go`
- `report/html.go`
- `.gitignore`, `README.md`

**Do not commit:**
- `splito.exe` (compiled binary — build from source)
- `*.xml` result files (personal/club data)
- `*.html` output files

---

## Requirements

- [Go 1.21+](https://go.dev/dl/) — no other dependencies

---

## Build

```bash
git clone <your-repo-url>
cd splito
go build -o splito.exe .   # Windows
go build -o splito .       # Linux / macOS
```

---

## Usage

```
splito [options] <input.xml>

Options:
  -class  string   Class name to process (e.g. "M 19 A")
                   Omit to list all available classes.
  -all             Generate a single HTML with ALL classes (ignores -class)
  -mode   string   Output mode: splits | timelost | all  (default: all)
  -out    string   Output HTML path (default: auto-named next to input file)
  -help            Show help
```

### Examples

```bash
# List all classes in the file
splito results.xml

# Single class, both splits table and time-lost chart
splito -class "M 19 A" results.xml

# Single class, splits table only
splito -class "W 21 E" -mode splits results.xml

# All classes in one file (TOC + per-class sections)
splito -all results.xml

# All classes, time-lost chart only, custom output path
splito -all -mode timelost -out /tmp/race_timelost.html results.xml
```

### Output naming (auto mode)

| Invocation | Output file |
|---|---|
| `-class "M 19 A"` | `<input>_M_19_A_all.html` |
| `-class "M 19 A" -mode splits` | `<input>_M_19_A_splits.html` |
| `-all` | `<input>_all_all.html` |
| `-all -mode timelost` | `<input>_all_timelost.html` |

---

## Input format

IOF XML 3.0 exported from MeOS, OE2010, or any compatible system.

Key parsing rules:
- `<SplitTime status="Additional">` entries are **ignored** (radio controls, spectator splits)
- Split times are cumulative seconds from start
- Competitors with status other than `OK` are shown in the table but excluded from the time-lost chart

---

## Report sections

### Splits table

| Column | Content |
|---|---|
| Pos | Finishing position |
| Name | Competitor name |
| Club | Club/organisation |
| Total | Finish time |
| Control 1…N | Two lines: cumul time + overall rank / leg split + leg rank |

Rank colours: 🥇 1st = red · 🥈 2nd = orange · 🥉 3rd = gold

### Time-lost chart (SVG)

- X axis: controls + finish
- Y axis: time behind leader in seconds (inverted — leader at top)
- One polyline per finisher; hover labels show name + time difference at each point

### All-classes report (`-all`)

- `<nav>` TOC at the top listing every class as a clickable link
- Each class is a `<section id="class_...">` anchor
- "↑ Back to top" link at the end of each section

---

## License

MIT — add a `LICENSE` file if you publish publicly.
