package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

var adrDir = "docs/adr"

const indexFile = "README.md"
const templateFile = "template.md"

func toKebabCase(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	s = strings.ReplaceAll(s, "_", "-")
	return s
}

func ensureDir(path string) error {
	return os.MkdirAll(path, os.ModePerm)
}

func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0644)
}

func extractTitleFromFilename(filename string) string {
	name := strings.TrimSuffix(filename, ".md")
	parts := strings.SplitN(name, "-", 2)
	if len(parts) < 2 {
		return filename
	}
	return strings.ReplaceAll(cases.Title(language.English).String(strings.ReplaceAll(parts[1], "-", " ")), "Adr ", "ADR ")
}

func updateIndex() error {
	files, err := os.ReadDir(adrDir)
	if err != nil {
		return err
	}

	var adrs []string
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") || file.Name() == indexFile || file.Name() == templateFile {
			continue
		}
		adrs = append(adrs, file.Name())
	}

	sort.Strings(adrs)

	indexPath := filepath.Join(adrDir, indexFile)
	indexContent := "# 📄 Architecture Decision Records\n\n"

	for _, adr := range adrs {
		title := extractTitleFromFilename(adr)
		indexContent += fmt.Sprintf("- [%s](%s)\n", title, adr)
	}

	return os.WriteFile(indexPath, []byte(indexContent), 0644)
}

func loadTemplateOrDefault() string {
	path := filepath.Join(adrDir, templateFile)
	bytes, err := os.ReadFile(path)
	if err == nil {
		return string(bytes)
	}

	// Default embedded template
	return `# ADR {{number}}: {{title}}

**Status**: {{status}}  
**Date**: {{date}}

---

## 📌 Context

Describe here the problem, need, or motivation for this decision. Include the current scenario, technical or business constraints, and the factors influencing the choice.

## ✅ Decision

Clearly state the decision made. For example:

> We decided to adopt the XYZ framework for developing REST APIs in the ABC project.

## 🤔 Considered Alternatives

- **Alternative A** (chosen): reasons for the choice...
- **Alternative B**: reasons for not choosing...
- **Alternative C**: pros and cons...

## 🎯 Consequences

Explain the impacts of this decision:

- Immediate or long-term benefits
- Possible risks or side effects
- Actions required to implement the decision

## 🔁 Relations

- Replaces ADR: 'adr-XXXX.md' _(if applicable)_
- Replaced by ADR: 'adr-XXXX.md' _(if applicable)_
- Related to: issues, RFCs, previous decisions

---

_This ADR follows the model of [Joel Parker Henderson](https://github.com/joelparkerhenderson/architecture-decision-record)_
`
}

func renderTemplate(template, number, status, title, date string) string {
	replacer := strings.NewReplacer(
		"{{number}}", number,
		"{{status}}", status,
		"{{title}}", title,
		"{{date}}", date,
	)
	return replacer.Replace(template)
}

func adrExists(number string) bool {
	files, err := os.ReadDir(adrDir)
	if err != nil {
		return false
	}

	prefix := fmt.Sprintf("adr-%s-", number)
	for _, file := range files {
		if strings.HasPrefix(file.Name(), prefix) {
			return true
		}
	}
	return false
}

func main() {
	number := flag.String("number", "", "Sequential ADR number (e.g., 001)")
	status := flag.String("status", "", "Decision status (e.g., Accepted, Proposed, Rejected)")
	title := flag.String("title", "", "Descriptive ADR title in quotes")
	flag.Parse()

	if *number == "" || *status == "" {
		fmt.Println("Required flags:")
		fmt.Println("  --number: Sequential ADR number (e.g., 001)")
		fmt.Println("  --status: Decision status (e.g., Accepted, Proposed, Rejected)")
		fmt.Println("\nOptional flag:")
		fmt.Println("  --title: Descriptive ADR title in quotes (required for new ADRs)")
		flag.PrintDefaults()
		return
	}

	isNewAdr := !adrExists(*number)
	if isNewAdr && *title == "" {
		fmt.Println("Error: --title flag is required when creating a new ADR")
		return
	}

	err := ensureDir(adrDir)
	if err != nil {
		fmt.Println("Error creating directory:", err)
		return
	}

	var filename string
	if isNewAdr {
		kebabTitle := toKebabCase(*title)
		filename = fmt.Sprintf("adr-%s-%s.md", *number, kebabTitle)
	} else {
		// For updates, find the existing file
		files, err := os.ReadDir(adrDir)
		if err != nil {
			fmt.Println("Error reading directory:", err)
			return
		}
		prefix := fmt.Sprintf("adr-%s-", *number)
		for _, file := range files {
			if strings.HasPrefix(file.Name(), prefix) {
				filename = file.Name()
				break
			}
		}
	}

	fullPath := filepath.Join(adrDir, filename)
	date := time.Now().Format("2006-01-02")

	var content string
	if isNewAdr {
		template := loadTemplateOrDefault()
		content = renderTemplate(template, *number, *status, *title, date)
	} else {
		// Read existing file
		existingContent, err := os.ReadFile(fullPath)
		if err != nil {
			fmt.Println("Error reading existing ADR:", err)
			return
		}
		content = strings.ReplaceAll(string(existingContent), "**Status**: ", fmt.Sprintf("**Status**: %s\n**Previous Status**: ", *status))
	}

	err = writeFile(fullPath, content)
	if err != nil {
		fmt.Println("Error writing ADR:", err)
		return
	}

	err = updateIndex()
	if err != nil {
		fmt.Println("Error updating index:", err)
		return
	}

	if isNewAdr {
		fmt.Printf("✅ New ADR created successfully: %s\n", fullPath)
	} else {
		fmt.Printf("✅ ADR updated successfully: %s\n", fullPath)
	}
}
