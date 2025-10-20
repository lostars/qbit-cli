package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"qbit-cli/internal/api"
	"qbit-cli/pkg/utils"
)

func TorrentCategoryCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "category [command]",
		Short: "Manage torrent category",
	}

	cmd.AddCommand(CategoryList())
	cmd.AddCommand(CategoryDelete())
	cmd.AddCommand(CategoryAdd())
	cmd.AddCommand(CategoryUpdate())

	return cmd
}

func CategoryUpdate() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "update <name> [flags]",
		Short: "Update category",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires a name")
			}
			return nil
		},
	}
	var savePath string
	cmd.Flags().StringVar(&savePath, "save-path", "", "save path")
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := api.CategoryUpdate(args[0], savePath); err != nil {
			return err
		}
		return nil
	}

	return cmd
}

func CategoryDelete() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "delete <name>...",
		Short: "Delete category",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires at least one name")
			}
			return nil
		},
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		if err := api.CategoryDelete(args); err != nil {
			return err
		}
		return nil
	}
	return cmd
}

func CategoryList() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list",
		Short: "List category",
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		categories, err := api.CategoryList()
		if err != nil {
			return err
		}

		header := []string{"Name", "SavePath"}
		var data = make([][]string, 0, len(*categories))
		for _, category := range *categories {
			data = append(data, []string{category.Name, category.SavePath})
		}
		utils.PrintList(header, &data)
		return nil
	}
	return cmd
}

func CategoryAdd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "add <name>... [flags]",
		Short: "Add category",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("name is required")
			}
			return nil
		},
	}

	var savePath string

	cmd.Flags().StringVar(&savePath, "save-path", "", "Save path")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {

		for _, arg := range args {
			if err := api.CategoryAdd(arg, savePath); err != nil {
				fmt.Printf("%s add failed: %s\n", arg, err.Error())
			}
		}

		return nil
	}
	return cmd
}
