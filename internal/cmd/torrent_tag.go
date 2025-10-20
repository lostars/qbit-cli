package cmd

import (
	"errors"
	"github.com/spf13/cobra"
	"qbit-cli/internal/api"
	"qbit-cli/pkg/utils"
)

func TagCmd() *cobra.Command {
	var tagCmd = &cobra.Command{
		Use:   "tag [command]",
		Short: "Tag management",
	}
	tagCmd.AddCommand(TagList())
	tagCmd.AddCommand(DeleteTag())
	tagCmd.AddCommand(AddTag())
	return tagCmd
}

func TagList() *cobra.Command {
	var tagListCmd = &cobra.Command{
		Use:   "list",
		Short: "List tags",
	}

	tagListCmd.RunE = func(cmd *cobra.Command, args []string) error {
		tags, err := api.TagList()
		if err != nil {
			return err
		}

		header := []string{"TAG"}
		data := make([][]string, 0, len(tags))
		for _, tag := range tags {
			data = append(data, []string{tag})
		}
		utils.PrintList(header, &data)

		return nil
	}

	return tagListCmd
}

func DeleteTag() *cobra.Command {
	var deleteTagCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete tags",
	}

	var all bool
	deleteTagCmd.Flags().BoolVar(&all, "all", false, "Delete all tags")

	deleteTagCmd.RunE = func(cmd *cobra.Command, args []string) error {
		if all {
			tags, err := api.TagList()
			if err != nil {
				return err
			}
			args = tags
		} else {
			if len(args) < 1 {
				return errors.New("must provide at least one tag")
			}
		}
		err := api.TagUpdate("deleteTags", args)
		if err != nil {
			return err
		}
		return nil
	}
	return deleteTagCmd
}

func AddTag() *cobra.Command {
	var addTagCmd = &cobra.Command{
		Use:   "add <name>...",
		Short: "Add tags",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("must provide at least one tag")
			}
			return nil
		},
	}

	addTagCmd.RunE = func(cmd *cobra.Command, args []string) error {
		err := api.TagUpdate("createTags", args)
		if err != nil {
			return err
		}
		return nil
	}
	return addTagCmd
}
