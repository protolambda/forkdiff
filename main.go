package main

import (
	"bytes"
	"context"
	"crypto/rand"
	"embed"
	"encoding/hex"
	"flag"
	"fmt"
	t2html "github.com/buildkite/terminal-to-html/v3"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/format/diff"
	"github.com/gomarkdown/markdown"
	"github.com/gomarkdown/markdown/html"
	"github.com/gomarkdown/markdown/parser"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
	"sort"
	"text/template"
)

//go:embed page.gohtml
var page embed.FS

func main() {
	repoPathStr := flag.String("repo", ".", "path to local git repository")
	targetRefStr := flag.String("target", "HEAD", "target reference to retrieve diff for")
	baseRefStr := flag.String("base", "master", "base reference to diff against")
	forkPagePathStr := flag.String("fork", "fork.yaml", "fork page definition")
	outStr := flag.String("out", "index.html", "output")
	flag.Parse()

	must := func(err error, msg string, args ...any) {
		if err != nil {
			_, _ = fmt.Fprintf(os.Stderr, msg, args...)
			_, _ = fmt.Fprintf(os.Stderr, "\nerror: %v", err)
			os.Exit(1)
		}
	}
	pageDefinition, err := readPageYaml(*forkPagePathStr)
	must(err, "failed to read page definition %q", *forkPagePathStr)

	repo, err := git.PlainOpen(*repoPathStr)
	must(err, "failed to open git repository %q", *repoPathStr)

	baseRef, err := repo.Reference(plumbing.ReferenceName(*baseRefStr), true)
	must(err, "failed to find base git ref %q", *baseRefStr)

	targetRef, err := repo.Reference(plumbing.ReferenceName(*targetRefStr), true)
	must(err, "failed to find target git ref %q", *targetRefStr)

	baseCommit, err := repo.CommitObject(baseRef.Hash())
	must(err, "failed to open base commit %s", baseRef.Hash())
	baseTree, err := baseCommit.Tree()
	must(err, "failed to open base git tree")

	targetCommit, err := repo.CommitObject(targetRef.Hash())
	must(err, "failed to open target commit %s", targetRef.Hash())
	targetTree, err := targetCommit.Tree()
	must(err, "failed to open target git tree")

	forkPatch, err := baseTree.PatchContext(context.Background(), targetTree)
	must(err, "failed to compute patch between base and target")

	patchByName := make(map[string]diff.FilePatch, len(forkPatch.FilePatches()))
	for _, fp := range forkPatch.FilePatches() {
		from, to := fp.Files()
		if to != nil {
			patchByName[to.Path()] = fp
		} else {
			patchByName[from.Path()] = fp
		}
	}
	remaining := make(map[string]struct{})
	for k := range patchByName {
		remaining[k] = struct{}{}
	}

	templ := template.New("main")
	templ.Funcs(template.FuncMap{
		"renderMarkdown": func(md string) string {
			markdownRenderer := html.NewRenderer(html.RendererOptions{
				Flags:     html.Smartypants | html.SmartypantsFractions | html.SmartypantsDashes | html.SmartypantsLatexDashes,
				Generator: "forkdiff",
			})
			markdownParser := parser.NewWithExtensions(parser.CommonExtensions | parser.OrderedListStart)
			return string(markdown.ToHTML([]byte(md), markdownParser, markdownRenderer))
		},
		"renderPatch": func(path string) (string, error) {
			p, ok := patchByName[path]
			if !ok {
				return "", fmt.Errorf("failed to find file patch %s", path)
			}
			var out bytes.Buffer
			enc := diff.NewUnifiedEncoder(&out, 3)
			enc.SetSrcPrefix(*baseRefStr)
			enc.SetDstPrefix(*targetRefStr)
			enc.SetColor(diff.NewColorConfig())

			err := enc.Encode(FilePatch{filePatch: p})
			if err != nil {
				return "", fmt.Errorf("")
			}
			delete(remaining, path)
			return string(t2html.Render(out.Bytes())), nil
		},
		"remainingPatches": func() (out []string) {
			for k := range remaining {
				out = append(out, k)
			}
			sort.Strings(out)
			return out
		},
		"nestForkDefinition": func(def *ForkDefinition, level int) *NestedForkDefinition {
			return &NestedForkDefinition{Def: def, Level: level}
		},
		"expandGlob": func(globPattern string) (out []string, err error) {
			for name := range patchByName {
				if ok, err := filepath.Match(globPattern, name); err != nil {
					return nil, fmt.Errorf("failed to glob match entry %q against pattern %q", name, globPattern)
				} else if ok {
					out = append(out, name)
				}
			}
			return out, nil
		},
		"randomID": func() (string, error) {
			var out [12]byte
			if _, err := rand.Read(out[:]); err != nil {
				return "", err
			}
			return "id-" + hex.EncodeToString(out[:]), nil
		},
	})
	templ, err = templ.ParseFS(page, "*.gohtml")
	must(err, "failed to parse page template")

	f, err := os.OpenFile(*outStr, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0o755)
	must(err, "failed to open output file")
	defer f.Close()
	must(templ.ExecuteTemplate(f, "main", pageDefinition), "failed to build page")
}

func readPageYaml(path string) (*Page, error) {
	f, err := os.OpenFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, fmt.Errorf("failed to read page YAML file: %w", err)
	}
	defer f.Close()
	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)
	var page Page
	if err := dec.Decode(&page); err != nil {
		return nil, fmt.Errorf("failed to decode page YAML file: %w", err)
	}
	return &page, nil
}

type FilePatch struct {
	filePatch diff.FilePatch
}

var _ diff.Patch = FilePatch{}

func (p FilePatch) FilePatches() []diff.FilePatch {
	return []diff.FilePatch{p.filePatch}
}

func (p FilePatch) Message() string {
	return ""
}

type Page struct {
	Title  string          `yaml:"title"`
	Footer string          `yaml:"footer"`
	Def    *ForkDefinition `yaml:"def"`
}

type ForkDefinition struct {
	Title       string            `yaml:"title,omitempty"`
	Description string            `yaml:"description,omitempty"`
	Globs       []string          `yaml:"globs,omitempty"`
	Sub         []*ForkDefinition `yaml:"sub,omitempty"`
}

type NestedForkDefinition struct {
	Def   *ForkDefinition
	Level int
}

func (nd *NestedForkDefinition) NextLevel() int {
	return nd.Level + 1
}
