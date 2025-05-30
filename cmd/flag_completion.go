package cmd

import (
	"github.com/spf13/cobra"
	"qbit-cli/internal/api"
	"strings"
)

type FlagsProperty[T any] struct {
	Value    T
	Flag     string
	Register FlagsPropertyRegister
	Options  []string
}

type FlagsPropertyRegister interface {
	complete(toComplete string) []string
}

func (f *FlagsProperty[T]) RegisterCompletion(cmd *cobra.Command) {
	if f.Options != nil && len(f.Options) > 0 {
		_ = cmd.RegisterFlagCompletionFunc(f.Flag, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return f.Options, cobra.ShellCompDirectiveNoFileComp
		})
		return
	}
	if f.Register != nil && f.Flag != "" {
		_ = cmd.RegisterFlagCompletionFunc(f.Flag, func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
			return f.Register.complete(toComplete), cobra.ShellCompDirectiveNoFileComp
		})
		return
	}
}

type TorrentPluginsFlagRegister struct{}

func (f *TorrentPluginsFlagRegister) complete(toComplete string) []string {
	plugins, err := api.SearchPlugins()
	if err != nil {
		return nil
	}

	var result = make([]string, 0, len(*plugins))
	for _, plugin := range *plugins {
		if !plugin.Enabled {
			continue
		}
		if toComplete != "" {
			if strings.Contains(plugin.Name, toComplete) {
				result = append(result, plugin.Name)
			}
		} else {
			result = append(result, plugin.Name)
		}
	}
	return result
}

type TorrentCategoryFlagRegister struct{}

func (f *TorrentCategoryFlagRegister) complete(toComplete string) []string {
	categories, err := api.CategoryList()
	if err != nil {
		return nil
	}
	var result = make([]string, 0, len(*categories))
	for _, category := range *categories {
		if toComplete != "" {
			if strings.Contains(category.Name, toComplete) {
				result = append(result, category.Name)
			}
		} else {
			result = append(result, category.Name)
		}
	}
	return result
}
