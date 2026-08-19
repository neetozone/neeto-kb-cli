package output

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
	"unicode/utf8"

	"github.com/alpkeskin/gotoon"
	"golang.org/x/term"
)

type Breadcrumb struct {
	Label   string `json:"label"`
	Command string `json:"command"`
}

type Envelope struct {
	Data        json.RawMessage `json:"data"`
	Breadcrumbs []Breadcrumb    `json:"breadcrumbs,omitempty"`
	Pagination  json.RawMessage `json:"pagination,omitempty"`
}

var (
	ForceJSON bool
	QuietMode bool
	ToonMode  bool
)

// priorityFields controls which columns appear in tables and their order.
var priorityFields = []string{
	"sid", "id", "name", "title", "email", "first_name", "last_name",
	"slug", "state", "status", "organization_role", "category",
	"url", "time_zone", "kind", "type", "disabled", "default",
}

const (
	maxTableColumns = 7
	ellipsis        = "..."
	minContentWidth = 10
	minColWidth     = minContentWidth + len(ellipsis)
	colPadding      = 3
	maxRenderDepth  = 3
	maxPreviewLen   = 100
)

func IsTTY() bool {
	return term.IsTerminal(int(os.Stdout.Fd()))
}

func UseJSON() bool {
	return ForceJSON || QuietMode || !IsTTY()
}

func Print(data json.RawMessage, breadcrumbs []Breadcrumb) {
	if ToonMode {
		printToon(data)
		return
	}

	if QuietMode {
		fmt.Println(string(data))
		return
	}

	if UseJSON() {
		printEnvelope(data, breadcrumbs, nil)
		return
	}

	printPretty(data)
	printBreadcrumbs(breadcrumbs)
}

func PrintWithPagination(data json.RawMessage, pagination json.RawMessage, breadcrumbs []Breadcrumb) {
	if ToonMode {
		printToon(data)
		printToonPagination(pagination)
		return
	}

	if QuietMode {
		fmt.Println(string(data))
		return
	}

	if UseJSON() {
		printEnvelope(data, breadcrumbs, pagination)
		return
	}

	printPretty(data)
	printPaginationSummary(pagination)
	printBreadcrumbs(breadcrumbs)
}

func PrintMessage(msg string) {
	if QuietMode {
		fmt.Println("success")
		return
	}

	if ToonMode {
		fmt.Println(msg)
		return
	}

	if UseJSON() {
		envelope := map[string]string{"message": msg}
		data, _ := json.Marshal(envelope)
		fmt.Println(string(data))
	} else {
		fmt.Println(msg)
	}
}

// PrintQuiet prints a minimal one-line summary of a resource suitable for
// action commands (create/update) where the full response body is not needed.
// In quiet mode it prints only the sid (or id, or name) so the caller gets
// just the identifier. In other modes it falls through to Print.
func PrintQuiet(data json.RawMessage, breadcrumbs []Breadcrumb) {
	if QuietMode {
		if id := extractIdentifier(data); id != "" {
			fmt.Println(id)
			return
		}
	}

	Print(data, breadcrumbs)
}

func extractIdentifier(data json.RawMessage) string {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return ""
	}

	if len(raw) == 1 {
		for _, v := range raw {
			var inner map[string]json.RawMessage
			if json.Unmarshal(v, &inner) == nil {
				raw = inner
			}
		}
	}

	for _, key := range []string{"sid", "id", "name"} {
		if v, ok := raw[key]; ok {
			var s string
			if json.Unmarshal(v, &s) == nil {
				return s
			}
		}
	}

	return ""
}

func printEnvelope(data json.RawMessage, breadcrumbs []Breadcrumb, pagination json.RawMessage) {
	envelope := Envelope{
		Data:        data,
		Breadcrumbs: breadcrumbs,
		Pagination:  pagination,
	}
	out, _ := json.MarshalIndent(envelope, "", "  ")
	fmt.Println(string(out))
}

func printPretty(data json.RawMessage) {
	var arr []map[string]interface{}
	if err := json.Unmarshal(data, &arr); err == nil {
		if len(arr) == 0 {
			fmt.Println("No records found.")
			return
		}
		printTable(arr)
		return
	}

	var obj map[string]interface{}
	if err := json.Unmarshal(data, &obj); err == nil {
		if len(obj) == 1 {
			for _, v := range obj {
				if inner, ok := v.(map[string]interface{}); ok {
					printKeyValue(inner, data)
					return
				}
			}
		}
		printKeyValue(obj, data)
		return
	}

	out, _ := json.MarshalIndent(data, "", "  ")
	fmt.Println(string(out))
}

func printTable(rows []map[string]interface{}) {
	cols := pickColumns(rows)
	if len(cols) == 0 {
		out, _ := json.MarshalIndent(rows, "", "  ")
		fmt.Println(string(out))
		return
	}

	headers := make([]string, len(cols))
	for i, col := range cols {
		headers[i] = formatHeader(col)
	}

	grid := make([][]string, len(rows))
	for i, row := range rows {
		grid[i] = make([]string, len(cols))
		for j, col := range cols {
			grid[i][j] = formatValue(row[col])
		}
	}

	widths := calculateWidths(headers, grid)
	pad := strings.Repeat(" ", colPadding)

	for i, h := range headers {
		if i > 0 {
			fmt.Print(pad)
		}
		cell := truncate(h, widths[i])
		if i < len(headers)-1 {
			cell = padRight(cell, widths[i])
		}
		fmt.Print(cell)
	}
	fmt.Println()

	for i, w := range widths {
		if i > 0 {
			fmt.Print(pad)
		}
		fmt.Print(strings.Repeat("─", w))
	}
	fmt.Println()

	for _, row := range grid {
		for i, val := range row {
			if i > 0 {
				fmt.Print(pad)
			}
			cell := truncate(val, widths[i])
			if i < len(row)-1 {
				cell = padRight(cell, widths[i])
			}
			fmt.Print(cell)
		}
		fmt.Println()
	}
}

func pickColumns(rows []map[string]interface{}) []string {
	scalars := map[string]bool{}
	for k, v := range rows[0] {
		if isScalar(v) {
			scalars[k] = true
		}
	}

	urlFields := map[string]bool{}
	var urlCols []string
	for _, row := range rows {
		for k, v := range row {
			if !scalars[k] || urlFields[k] {
				continue
			}
			if s, ok := v.(string); ok && isURL(s) {
				urlFields[k] = true
				urlCols = append(urlCols, k)
			}
		}
	}
	sort.Strings(urlCols)

	budget := max(1, maxTableColumns-len(urlCols))

	var scalarKeys []string
	for k := range scalars {
		if !urlFields[k] {
			scalarKeys = append(scalarKeys, k)
		}
	}

	cols := orderedKeys(scalarKeys)
	if len(cols) > budget {
		cols = cols[:budget]
	}

	return append(cols, urlCols...)
}

func calculateWidths(headers []string, grid [][]string) []int {
	widths := make([]int, len(headers))
	protected := make([]bool, len(headers))
	for i, h := range headers {
		widths[i] = displayWidth(h)
	}
	for _, row := range grid {
		for i, val := range row {
			if w := displayWidth(val); w > widths[i] {
				widths[i] = w
			}
			if isURL(val) {
				protected[i] = true
			}
		}
	}

	termWidth := getTerminalWidth()
	totalPad := (len(headers) - 1) * colPadding
	available := termWidth - totalPad

	total, flexible := 0, 0
	for i, w := range widths {
		total += w
		if !protected[i] {
			flexible += w
		}
	}

	if total <= available || flexible == 0 {
		return widths
	}

	// URL columns keep their full width; the rest absorb the shortfall, but
	// never shrink past minColWidth nor grow beyond what their content needs.
	budget := max(0, available-(total-flexible))
	for i := range widths {
		if !protected[i] {
			widths[i] = min(widths[i], max(minColWidth, widths[i]*budget/flexible))
		}
	}

	return widths
}

func printKeyValue(obj map[string]interface{}, rawData json.RawMessage) {
	fields := fieldOrder(rawData)
	if len(fields) == 0 {
		fields = orderedKeys(keysOf(obj))
	}
	renderFields(obj, fields, 1)
}

func renderFields(obj map[string]interface{}, fields []string, depth int) {
	indent := strings.Repeat("  ", depth)
	labelWidth := maxLabelWidth(obj, fields)

	for _, k := range fields {
		v, ok := obj[k]
		if !ok {
			continue
		}
		label := formatHeader(k)

		switch val := v.(type) {
		case map[string]interface{}:
			if depth < maxRenderDepth && len(val) > 0 {
				fmt.Printf("%s%s\n", indent, label)
				renderFields(val, orderedKeys(keysOf(val)), depth+1)
			} else {
				fmt.Printf("%s%-*s  %s\n", indent, labelWidth, label, summarizeObject(val))
			}
		case []interface{}:
			fmt.Printf("%s%-*s  %s\n", indent, labelWidth, label, formatArray(val))
		default:
			fmt.Printf("%s%-*s  %s\n", indent, labelWidth, label, previewScalar(v))
		}
	}
}

func keysOf(obj map[string]interface{}) []string {
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	return keys
}

func orderedKeys(keys []string) []string {
	remaining := make(map[string]bool, len(keys))
	for _, k := range keys {
		remaining[k] = true
	}

	var ordered []string
	for _, f := range priorityFields {
		if remaining[f] {
			ordered = append(ordered, f)
			delete(remaining, f)
		}
	}

	var rest []string
	for _, k := range keys {
		if remaining[k] {
			rest = append(rest, k)
		}
	}
	sort.Strings(rest)

	return append(ordered, rest...)
}

func maxLabelWidth(obj map[string]interface{}, fields []string) int {
	width := 0
	for _, k := range fields {
		if _, ok := obj[k]; !ok {
			continue
		}
		if l := len(formatHeader(k)); l > width {
			width = l
		}
	}
	return width
}

func previewScalar(v interface{}) string {
	s := strings.TrimSpace(strings.ReplaceAll(formatValue(v), "\n", " "))
	return truncate(s, maxPreviewLen)
}

func formatArray(arr []interface{}) string {
	if len(arr) == 0 {
		return "(0 items)"
	}

	allScalar := true
	for _, v := range arr {
		if !isScalar(v) {
			allScalar = false
			break
		}
	}

	if allScalar {
		parts := make([]string, len(arr))
		for i, v := range arr {
			parts[i] = formatValue(v)
		}
		if joined := strings.Join(parts, ", "); len(joined) <= maxPreviewLen {
			return joined
		}
	}

	return fmt.Sprintf("(%d items)", len(arr))
}

func summarizeObject(val map[string]interface{}) string {
	compact, _ := json.Marshal(val)
	if len(compact) <= 80 || containsURL(val) {
		return string(compact)
	}
	return fmt.Sprintf("(%d fields)", len(val))
}

func fieldOrder(data json.RawMessage) []string {
	dec := json.NewDecoder(bytes.NewReader(data))
	t, err := dec.Token()
	if err != nil || t != json.Delim('{') {
		return nil
	}

	var outerKeys []string
	var firstValue json.RawMessage

	for dec.More() {
		t, err := dec.Token()
		if err != nil {
			break
		}
		key, ok := t.(string)
		if !ok {
			break
		}
		outerKeys = append(outerKeys, key)

		var val json.RawMessage
		if err := dec.Decode(&val); err != nil {
			break
		}
		if firstValue == nil {
			firstValue = val
		}
	}

	if len(outerKeys) == 1 && firstValue != nil {
		if inner := extractKeys(firstValue); len(inner) > 0 {
			return inner
		}
	}

	return outerKeys
}

func extractKeys(data json.RawMessage) []string {
	dec := json.NewDecoder(bytes.NewReader(data))
	t, err := dec.Token()
	if err != nil || t != json.Delim('{') {
		return nil
	}

	var keys []string
	for dec.More() {
		t, err := dec.Token()
		if err != nil {
			break
		}
		key, ok := t.(string)
		if !ok {
			break
		}
		keys = append(keys, key)

		var val json.RawMessage
		if err := dec.Decode(&val); err != nil {
			break
		}
	}
	return keys
}

func printToon(data json.RawMessage) {
	var v interface{}
	if err := json.Unmarshal(data, &v); err != nil {
		fmt.Println(string(data))
		return
	}

	out, err := gotoon.Encode(v)
	if err != nil {
		fmt.Println(string(data))
		return
	}

	fmt.Print(out)
}

func printToonPagination(pagination json.RawMessage) {
	if pagination == nil {
		return
	}

	var v interface{}
	if err := json.Unmarshal(pagination, &v); err != nil {
		return
	}

	out, err := gotoon.Encode(v)
	if err != nil {
		return
	}

	fmt.Print(out)
}

func isScalar(v interface{}) bool {
	switch v.(type) {
	case nil, string, float64, bool, json.Number:
		return true
	}
	return false
}

func formatHeader(field string) string {
	return strings.ToUpper(strings.ReplaceAll(field, "_", " "))
}

func formatValue(v interface{}) string {
	if v == nil {
		return "-"
	}
	switch val := v.(type) {
	case bool:
		if val {
			return "Yes"
		}
		return "No"
	case float64:
		if val == float64(int64(val)) {
			return fmt.Sprintf("%d", int64(val))
		}
		return fmt.Sprintf("%.2f", val)
	case string:
		return val
	default:
		return fmt.Sprintf("%v", val)
	}
}

func isURL(s string) bool {
	return strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://")
}

func containsURL(v interface{}) bool {
	switch val := v.(type) {
	case string:
		return isURL(val)
	case []interface{}:
		for _, item := range val {
			if containsURL(item) {
				return true
			}
		}
	case map[string]interface{}:
		for _, item := range val {
			if containsURL(item) {
				return true
			}
		}
	}
	return false
}

func displayWidth(s string) int {
	return utf8.RuneCountInString(s)
}

func padRight(s string, width int) string {
	if gap := width - displayWidth(s); gap > 0 {
		return s + strings.Repeat(" ", gap)
	}
	return s
}

func truncate(s string, maxLen int) string {
	if maxLen < 0 {
		maxLen = 0
	}
	if displayWidth(s) <= maxLen || isURL(s) {
		return s
	}
	runes := []rune(s)
	if maxLen <= len(ellipsis) {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-len(ellipsis)]) + ellipsis
}

func getTerminalWidth() int {
	w, _, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil || w <= 0 {
		return 100
	}
	return w
}

func printPaginationSummary(pagination json.RawMessage) {
	if pagination == nil {
		return
	}

	var p struct {
		CurrentPageNumber int `json:"current_page_number"`
		TotalPages        int `json:"total_pages"`
		TotalRecords      int `json:"total_records"`
	}
	if err := json.Unmarshal(pagination, &p); err != nil {
		return
	}

	fmt.Printf("\nPage %d of %d (%d total records)\n", p.CurrentPageNumber, p.TotalPages, p.TotalRecords)
}

func printBreadcrumbs(breadcrumbs []Breadcrumb) {
	if len(breadcrumbs) == 0 {
		return
	}

	fmt.Println()
	for _, b := range breadcrumbs {
		fmt.Printf("  %s: %s\n", b.Label, b.Command)
	}
}
