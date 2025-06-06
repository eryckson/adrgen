package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/manifoldco/promptui"
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

## Context

Describe here the problem, need, or motivation for this decision. Include the current scenario, technical or business constraints, and the factors influencing the choice.

## Decision

Clearly state the decision made. For example:

> We decided to adopt the XYZ framework for developing REST APIs in the ABC project.

## Considered Alternatives

- **Alternative A** (chosen): reasons for the choice...
- **Alternative B**: reasons for not choosing...
- **Alternative C**: pros and cons...

## Consequences

Explain the impacts of this decision:

- Immediate or long-term benefits
- Possible risks or side effects
- Actions required to implement the decision

## Relations

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

func getNextADRNumber() string {
	files, err := os.ReadDir(adrDir)
	if err != nil {
		return "001" // Start with 001 if directory doesn't exist
	}

	maxNum := 0
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".md") || file.Name() == indexFile || file.Name() == templateFile {
			continue
		}

		// Extract number from filename (format: adr-XXX-*.md)
		if strings.HasPrefix(file.Name(), "adr-") {
			numStr := strings.Split(strings.TrimPrefix(file.Name(), "adr-"), "-")[0]
			if num, err := strconv.Atoi(numStr); err == nil {
				if num > maxNum {
					maxNum = num
				}
			}
		}
	}

	return fmt.Sprintf("%03d", maxNum+1)
}

func promptForNumber() (string, error) {
	nextNum := getNextADRNumber()

	validate := func(input string) error {
		if len(input) == 0 {
			return fmt.Errorf("number cannot be empty")
		}
		if len(input) != 3 {
			return fmt.Errorf("number must be 3 digits (e.g., 001)")
		}
		if _, err := strconv.Atoi(input); err != nil {
			return fmt.Errorf("number must be numeric")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:     "ADR Number",
		Validate:  validate,
		Default:   nextNum,
		AllowEdit: true,
	}

	return prompt.Run()
}

func promptForStatus() (string, error) {
	prompt := promptui.Select{
		Label: "Select Status",
		Items: []string{"Accepted", "Proposed", "Rejected", "Superseded", "Deprecated"},
	}

	_, result, err := prompt.Run()
	return result, err
}

func getCurrentTitle(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "# ADR") {
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) == 2 {
				return parts[1]
			}
		}
	}
	return ""
}

func updateTitle(content, newTitle string) string {
	lines := strings.Split(content, "\n")
	for i, line := range lines {
		if strings.HasPrefix(line, "# ADR") {
			parts := strings.SplitN(line, ": ", 2)
			if len(parts) == 2 {
				lines[i] = fmt.Sprintf("%s: %s", parts[0], newTitle)
				break
			}
		}
	}
	return strings.Join(lines, "\n")
}

func promptForTitle(defaultTitle string) (string, error) {
	validate := func(input string) error {
		if len(input) == 0 {
			return fmt.Errorf("title cannot be empty")
		}
		return nil
	}

	prompt := promptui.Prompt{
		Label:     "ADR Title",
		Validate:  validate,
		Default:   defaultTitle,
		AllowEdit: true,
	}

	return prompt.Run()
}

func getCurrentStatus(content string) string {
	lines := strings.Split(content, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "**Status**: ") {
			return strings.TrimSpace(strings.TrimPrefix(line, "**Status**: "))
		}
	}
	return ""
}

func updateStatus(content, newStatus string) string {
	currentStatus := getCurrentStatus(content)
	if currentStatus == newStatus {
		return content // Status hasn't changed, return content as is
	}

	lines := strings.Split(content, "\n")
	newLines := make([]string, 0, len(lines))
	statusFound := false
	dateFound := false

	for _, line := range lines {
		if strings.HasPrefix(line, "**Status**: ") {
			if !statusFound {
				// Add new status line and previous status line only once
				newLines = append(newLines, fmt.Sprintf("**Status**: %s  ", newStatus))
				newLines = append(newLines, fmt.Sprintf("**Previous Status**: %s  ", currentStatus))
				statusFound = true
			}
			continue
		}

		// Skip any existing Previous Status lines
		if strings.HasPrefix(line, "**Previous Status**: ") {
			continue
		}

		// Keep the date line in its original position
		if strings.HasPrefix(line, "**Date**: ") {
			if !dateFound {
				newLines = append(newLines, line)
				dateFound = true
			}
			continue
		}

		// Add all other lines
		newLines = append(newLines, line)
	}

	// If we haven't found and added the status yet, add it after the title
	if !statusFound {
		result := make([]string, 0, len(newLines)+2)
		titleFound := false
		for _, line := range newLines {
			result = append(result, line)
			if strings.HasPrefix(line, "# ADR") {
				titleFound = true
				result = append(result, "")
				result = append(result, fmt.Sprintf("**Status**: %s", newStatus))
				result = append(result, fmt.Sprintf("**Previous Status**: %s", currentStatus))
			}
		}
		if !titleFound {
			// If no title was found, add status at the beginning
			result = append([]string{
				fmt.Sprintf("**Status**: %s", newStatus),
				fmt.Sprintf("**Previous Status**: %s", currentStatus),
				"",
			}, result...)
		}
		newLines = result
	}

	return strings.Join(newLines, "\n")
}

func main() {
	number, err := promptForNumber()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	status, err := promptForStatus()
	if err != nil {
		fmt.Printf("Prompt failed %v\n", err)
		return
	}

	err = ensureDir(adrDir)
	if err != nil {
		fmt.Println("Error creating directory:", err)
		return
	}

	var oldFilename, filename, title string
	isNewAdr := !adrExists(number)

	if isNewAdr {
		title, err = promptForTitle("")
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}
		kebabTitle := toKebabCase(title)
		filename = fmt.Sprintf("adr-%s-%s.md", number, kebabTitle)
	} else {
		// For updates, find the existing file
		files, err := os.ReadDir(adrDir)
		if err != nil {
			fmt.Println("Error reading directory:", err)
			return
		}
		prefix := fmt.Sprintf("adr-%s-", number)
		for _, file := range files {
			if strings.HasPrefix(file.Name(), prefix) {
				oldFilename = file.Name()
				break
			}
		}

		// Read existing content to get current title
		existingContent, err := os.ReadFile(filepath.Join(adrDir, oldFilename))
		if err != nil {
			fmt.Println("Error reading existing ADR:", err)
			return
		}

		currentTitle := getCurrentTitle(string(existingContent))
		title, err = promptForTitle(currentTitle)
		if err != nil {
			fmt.Printf("Prompt failed %v\n", err)
			return
		}

		// Only update filename if title changed
		if title != currentTitle {
			kebabTitle := toKebabCase(title)
			filename = fmt.Sprintf("adr-%s-%s.md", number, kebabTitle)
		} else {
			filename = oldFilename
		}
	}

	fullPath := filepath.Join(adrDir, filename)
	date := time.Now().Format("2006-01-02")

	var content string
	if isNewAdr {
		template := loadTemplateOrDefault()
		content = renderTemplate(template, number, status, title, date)
	} else {
		// Read existing file
		existingContent, err := os.ReadFile(filepath.Join(adrDir, oldFilename))
		if err != nil {
			fmt.Println("Error reading existing ADR:", err)
			return
		}
		content = updateStatus(string(existingContent), status)
		content = updateTitle(content, title)

		// If filename changed, remove old file
		if filename != oldFilename {
			err = os.Remove(filepath.Join(adrDir, oldFilename))
			if err != nil {
				fmt.Printf("Warning: Could not remove old file: %v\n", err)
			}
		}
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
		if filename != oldFilename {
			fmt.Printf("✅ ADR updated successfully (renamed from %s to %s)\n", oldFilename, filename)
		} else {
			fmt.Printf("✅ ADR updated successfully: %s\n", fullPath)
		}
	}
}
