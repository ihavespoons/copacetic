/*
Copyright © 2024 Ben Gittins
*/

package repository

import (
	"context"
	"copacetic/internal/lang"
	"errors"
	"github.com/go-enry/go-enry/v2"
	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/sourcegraph/conc/pool"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type File struct {
	Path            string
	Language        string
	LanguageGroup   string
	IsDotFile       bool
	IsVendor        bool
	IsConfiguration bool
	IsGenerated     bool
	IsDocumentation bool
	IsTest          bool
}

type Source struct {
	Directory     string
	GitURL        string
	GitRef        string
	Repository    *git.Repository
	Languages     lang.LanguageSet
	Files         []string
	Source        []*File
	Dotfiles      []*File
	Vendored      []*File
	Configuration []*File
	Generated     []*File
	Documentation []*File
	Test          []*File
}

func (g *Source) Clone(stdOut io.Writer) error {
	// Clone repo using the reference name of main if none provided
	clone, err := git.PlainClone(g.Directory, false, &git.CloneOptions{
		URL:           g.GitURL,
		ReferenceName: plumbing.ReferenceName(g.GitRef),
		Progress:      stdOut,
	})

	if errors.Is(err, git.ErrRepositoryAlreadyExists) {
		_ = os.RemoveAll(g.Directory)
	}

	if err != nil {
		return err
	}
	g.Repository = clone
	return nil
}

func (g *Source) Walk() error {
	err := filepath.Walk(g.Directory, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			if !strings.Contains(path, "/.git/") {
				g.Files = append(g.Files, path)
			}

		}
		return nil
	})

	if err != nil {
		return err
	}
	err = g.ProcessLanguages()
	if err != nil {
		return err
	}
	return nil
}

func langWorker(ctx context.Context, path *string) (*File, error) {
	fileContents, err := os.ReadFile(*path)
	if err != nil {
		return nil, err
	}

	if !enry.IsImage(*path) && !enry.IsBinary(fileContents) {
		langFound := enry.GetLanguage(*path, fileContents)

		// Sometimes enry misses files as such a catch all is necessary
		if langFound == "" {
			langFound = "Unknown"
		}

		newFile := File{
			Path:            *path,
			Language:        langFound,
			LanguageGroup:   enry.GetLanguageGroup(langFound),
			IsDotFile:       enry.IsDotFile(*path),
			IsVendor:        enry.IsVendor(*path),
			IsConfiguration: enry.IsConfiguration(*path),
			IsGenerated:     enry.IsGenerated(*path, fileContents),
			IsDocumentation: enry.IsDocumentation(*path),
			IsTest:          enry.IsTest(*path),
		}

		return &newFile, nil
	}

	return nil, nil
}

func (g *Source) ProcessLanguages() error {
	// Create conq pool of workers that will return a slice of pointers to new files using langWorker fn
	p := pool.NewWithResults[*File]().WithContext(context.Background())
	for _, file := range g.Files {
		file := file
		p.Go(func(ctx context.Context) (*File, error) {
			return langWorker(ctx, &file)
		})
	}
	// Wait for execution to finish
	result, err := p.Wait()
	if err != nil {
		return err
	}
	g.Languages = make(map[string]int)
	// Fingerprint languages
	for _, fileFound := range result {
		if fileFound != nil {
			if !g.Languages.Has(fileFound.Language) {
				g.Languages[fileFound.Language] = 1
			} else {
				g.Languages[fileFound.Language] += 1
			}
			switch {
			case fileFound.IsDotFile:
				g.Dotfiles = append(g.Dotfiles, fileFound)
			case fileFound.IsConfiguration:
				g.Configuration = append(g.Configuration, fileFound)
			case fileFound.IsGenerated:
				g.Generated = append(g.Generated, fileFound)
			case fileFound.IsVendor:
				g.Vendored = append(g.Vendored, fileFound)
			case fileFound.IsDocumentation:
				g.Documentation = append(g.Documentation, fileFound)
			case fileFound.IsTest:
				g.Test = append(g.Test, fileFound)
			default:
				g.Source = append(g.Source, fileFound)
			}
		}
	}

	return nil
}
