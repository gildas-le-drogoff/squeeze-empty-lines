package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sync"
)

var excludeDirs = map[string]struct{}{
	".git":         {},
	".venv":        {},
	"node_modules": {},
	"target":       {},
	"vendor":       {},
	"venv":         {},
}
var excludeFiles = map[string]struct{}{
	".env":            {},
	".env.example":    {},
	".env.production": {},
	".env.local":      {},
	"go.mod":          {},
	"go.sum":          {},
}
var extensionsAutorisees = map[string]struct{}{
	".bash": {}, ".bat": {}, ".c": {}, ".cc": {}, ".cfg": {}, ".cmd": {},
	".conf": {}, ".cpp": {}, ".cs": {}, ".css": {}, ".cxx": {},
	".fish": {}, ".go": {}, ".h": {}, ".hpp": {}, ".hxx": {},
	".html": {}, ".ini": {}, ".java": {}, ".js": {}, ".json": {},
	".jsonl": {}, ".jsx": {}, ".kt": {}, ".kts": {},
	".md": {}, ".mod": {}, ".sum": {}, ".work": {},
	".php": {}, ".ps1": {}, ".py": {}, ".pyi": {},
	".rb": {}, ".rs": {}, ".scss": {},
	".sh": {}, ".sql": {}, ".swift": {},
	".toml": {}, ".ts": {}, ".tsx": {},
	".txt": {}, ".xml": {},
	".yaml": {}, ".yml": {},
	".zsh": {},
	".vue": {}, ".svelte": {}, ".astro": {},
	".mjs": {}, ".cjs": {},
	".gradle": {}, ".properties": {},
	".dockerfile": {},
}
var extensionsSensiblesIndentation = map[string]struct{}{
	".py":     {},
	".pyi":    {},
	".yaml":   {},
	".yml":    {},
	".nim":    {},
	".coffee": {},
	".pug":    {},
	".jade":   {},
}
var fichiersSansExtensionAutorises = map[string]struct{}{
	"Dockerfile":     {},
	"Makefile":       {},
	".gitignore":     {},
	".gitattributes": {},
	".editorconfig":  {},
}
var regexEspacesInternes = regexp.MustCompile(`[ \t]{2,}`)
var regexNewline = regexp.MustCompile(`\r\n|\r|\n`)
var dryRun bool
var backup bool
var workers int
var collapseInternalSpaces bool
var maxSize int64

const tailleMaxParDefaut int64 = 5 * 1024 * 1024

var includePatterns multiFlag
var excludePatterns multiFlag
var includeRegex []*regexp.Regexp
var excludeRegex []*regexp.Regexp

type multiFlag []string

func (m *multiFlag) String() string { return "" }
func (m *multiFlag) Set(value string) error {
	*m = append(*m, value)
	return nil
}
func main() {
	flag.BoolVar(&dryRun, "dry-run", false, "")
	flag.BoolVar(&backup, "backup", false, "")
	flag.IntVar(&workers, "workers", runtime.NumCPU(), "")
	flag.BoolVar(&collapseInternalSpaces, "collapse-internal-spaces", false, "")
	flag.Int64Var(&maxSize, "max-size", tailleMaxParDefaut, "")
	flag.Var(&includePatterns, "include", "")
	flag.Var(&excludePatterns, "exclude", "")
	flag.Parse()
	includeRegex = compilerRegex(includePatterns)
	excludeRegex = compilerRegex(excludePatterns)
	args := flag.Args()
	if len(args) == 0 {
		args = []string{"."}
	}
	sem := make(chan struct{}, workers)
	var wg sync.WaitGroup
	for _, chemin := range args {
		info, err := os.Stat(chemin)
		if err != nil {
			continue
		}
		if info.IsDir() {
			wg.Add(1)
			go func(root string) {
				defer wg.Done()
				parcourir(root, sem, &wg)
			}(chemin)
		} else {
			wg.Add(1)
			go traiterAvecSemaphore(chemin, sem, &wg)
		}
	}
	wg.Wait()
}
func compilerRegex(list []string) []*regexp.Regexp {
	var out []*regexp.Regexp
	for _, p := range list {
		r, err := regexp.Compile(p)
		if err == nil {
			out = append(out, r)
		}
	}
	return out
}
func matchRegex(list []*regexp.Regexp, path string) bool {
	for _, r := range list {
		if r.MatchString(path) {
			return true
		}
	}
	return false
}
func autoriseParRegex(path string) bool {
	path = filepath.ToSlash(path)
	if len(includeRegex) > 0 {
		if !matchRegex(includeRegex, path) {
			return false
		}
	}
	if matchRegex(excludeRegex, path) {
		return false
	}
	return true
}
func parcourir(root string, sem chan struct{}, wg *sync.WaitGroup) {
	filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			if excluDossier(path) {
				return filepath.SkipDir
			}
			return nil
		}
		if excluFichier(path) {
			return nil
		}
		if !extensionValide(path) {
			return nil
		}
		if !autoriseParRegex(path) {
			return nil
		}
		if fichierTropGros(path) {
			return nil
		}
		wg.Add(1)
		go traiterAvecSemaphore(path, sem, wg)
		return nil
	})
}
func traiterAvecSemaphore(path string, sem chan struct{}, wg *sync.WaitGroup) {
	sem <- struct{}{}
	defer func() {
		<-sem
		wg.Done()
	}()
	traiter(path)
}
func traiter(path string) {
	if fichierTropGros(path) {
		return
	}
	if estBinaire(path) {
		return
	}
	original, mode, ok := lireFichier(path)
	if !ok {
		return
	}
	nouveau := normaliser(path, original)
	if bytes.Equal(original, nouveau) {
		return
	}
	if dryRun {
		fmt.Println("dry:", path)
		return
	}
	if backup {
		os.WriteFile(path+".bak", original, mode)
	}
	os.WriteFile(path, nouveau, mode)
	fmt.Println("fix:", path)
}
func fichierTropGros(path string) bool {
	if maxSize <= 0 {
		return false
	}
	info, err := os.Stat(path)
	if err != nil {
		return true
	}
	return info.Size() > maxSize
}
func lireFichier(path string) ([]byte, os.FileMode, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, 0, false
	}
	if maxSize > 0 && info.Size() > maxSize {
		return nil, 0, false
	}
	f, err := os.Open(path)
	if err != nil {
		return nil, 0, false
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		return nil, 0, false
	}
	return b, info.Mode(), true
}
func normaliser(path string, contenu []byte) []byte {
	contenu = regexNewline.ReplaceAll(contenu, []byte("\n"))
	ext := filepath.Ext(path)
	_, indentationSensible := extensionsSensiblesIndentation[ext]
	collapse := collapseInternalSpaces && !indentationSensible
	reader := bufio.NewReader(bytes.NewReader(contenu))
	var buffer bytes.Buffer
	buffer.Grow(len(contenu))
	for {
		line, err := reader.ReadBytes('\n')
		if len(line) == 0 && err != nil {
			break
		}
		line = bytes.TrimSuffix(line, []byte("\n"))
		if collapse {
			line = compacterSansIndentation(line)
		}
		line = bytes.TrimRight(line, " \t")
		if len(bytes.TrimSpace(line)) != 0 {
			buffer.Write(line)
			buffer.WriteByte('\n')
		}
		if err == io.EOF {
			break
		}
	}
	return buffer.Bytes()
}
func compacterSansIndentation(line []byte) []byte {
	i := 0
	for i < len(line) && (line[i] == ' ' || line[i] == '\t') {
		i++
	}
	indent := line[:i]
	reste := line[i:]
	reste = regexEspacesInternes.ReplaceAll(reste, []byte(" "))
	out := make([]byte, 0, len(line))
	out = append(out, indent...)
	out = append(out, reste...)
	return out
}
func excluDossier(path string) bool {
	base := filepath.Base(path)
	_, ok := excludeDirs[base]
	return ok
}
func excluFichier(path string) bool {
	base := filepath.Base(path)
	_, ok := excludeFiles[base]
	return ok
}
func extensionValide(path string) bool {
	base := filepath.Base(path)
	if _, ok := fichiersSansExtensionAutorises[base]; ok {
		return true
	}
	ext := filepath.Ext(path)
	if ext == "" {
		return false
	}
	_, ok := extensionsAutorisees[ext]
	return ok
}
func estBinaire(path string) bool {
	f, err := os.Open(path)
	if err != nil {
		return true
	}
	defer f.Close()
	buf := make([]byte, 8000)
	n, err := f.Read(buf)
	if err != nil && err != io.EOF {
		return true
	}
	buf = buf[:n]
	if bytes.IndexByte(buf, 0) != -1 {
		return true
	}
	return false
}
