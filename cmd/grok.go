/*
Copyright © 2024 Ben Gittins
*/

package cmd

import (
	"copacetic/internal/llm/openai"
	"copacetic/internal/repository"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
)

type Grok struct {
	GitRepo    string
	GitRef     string
	Repository repository.Source
}

var TempRepo string

func (g *Grok) Hydrate(cmd *cobra.Command) {
	g.GitRepo, _ = cmd.Flags().GetString("git")
	g.GitRef, _ = cmd.Flags().GetString("ref")
}

// grokCmd represents the grok command
var grokCmd = &cobra.Command{
	Use:   "grok",
	Short: "Scan provided git repository",
	Long:  ``,
	PersistentPostRun: func(cmd *cobra.Command, args []string) {
		// Clear up the tmp directory on exit
		_ = os.RemoveAll(TempRepo)
	},
	Run: func(cmd *cobra.Command, args []string) {
		// Get the configuration
		grokRun := Grok{}
		grokRun.Hydrate(cmd)

		// Create temporary directory for git repository
		tmp, err := os.MkdirTemp("", "")
		cobra.CheckErr(err)

		TempRepo = filepath.Join(tmp, "grok/repo")

		grokRun.Repository = repository.Source{
			Directory:  TempRepo,
			GitURL:     grokRun.GitRepo,
			GitRef:     grokRun.GitRef,
			Repository: nil,
		}

		// Get stdout from CMD to output the git progress
		stdOut := cmd.OutOrStdout()

		err = grokRun.Repository.Clone(stdOut)
		cobra.CheckErr(err)

		// Do initial processing of languages used in code
		err = grokRun.Repository.Walk()
		cobra.CheckErr(err)
		_, err = openai.New(0.1, &grokRun.Repository)
		cobra.CheckErr(err)
	},
}

func init() {
	rootCmd.AddCommand(grokCmd)
	grokCmd.PersistentFlags().StringP("git", "g", "", "Git repository either HTTPS or SSH")
	grokCmd.PersistentFlags().StringP("ref", "r", "main", "Git ref you want to clone defaults to main")
	grokCmd.PersistentFlags().StringP("model", "m", "openai", "The model family to use")
	err := grokCmd.MarkPersistentFlagRequired("git")
	cobra.CheckErr(err)
}
