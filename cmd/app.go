package cmd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"qbit-cli/internal/api"
	"strconv"
	"strings"
)

func AppCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "app",
		Short: "Manage app",
	}

	cmd.AddCommand(AppInfo())
	cmd.AddCommand(AppPreference())
	cmd.AddCommand(UpdatePreference())

	return cmd
}

func AppInfo() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "info",
		Short: "Show app info",
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		fmt.Println(api.GetQbitServerInfo())
		return nil
	}

	return cmd
}

func AppPreference() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "p",
		Short: "Show app preferences in formated json",
		Long: `Most preferences is key-value format except "scan_dirs":
{
    "/home/user/Downloads/incoming/games": 0,
    "/home/user/Downloads/incoming/movies": 1,
}
So take care when you want to modify this preference.`,
	}

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		d, err := api.QbitAppPreference()
		if err != nil {
			return err
		}

		var prettyJSON bytes.Buffer
		err = json.Indent(&prettyJSON, []byte(d), "", "    ")
		if err != nil {
			return err
		}
		fmt.Println(prettyJSON.String())

		return nil
	}

	return cmd
}

func UpdatePreference() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sp <key>",
		Short: "Update app preferences",
		Long:  `Paths in scan_dirs must exist, otherwise this option will have no effect`,
		Example: `sp "locale" --value="zh_CN"
sp "schedule_from_hour" --value=1
sp "scheduler_enabled" --value=false
sp "scan_dirs" --scan-dirs="/home/user/Downloads/incoming/movies:1"
sp "scan_dirs" --scan-dirs=":" (clear scan_dirs)`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires preference key")
			}
			return nil
		},
	}

	var value SmartValue
	var scanDirs []string
	cmd.Flags().Var(&value, "value", "preference value")
	cmd.Flags().StringSliceVar(&scanDirs, "scan-dirs", []string{}, "scan dirs")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		var v interface{}
		if args[0] == "scan_dirs" {
			if len(scanDirs) < 1 {
				return errors.New("requires at least one scan_dir")
			}
			dirs := make(map[string]int, len(scanDirs))
			for _, dir := range scanDirs {
				array := strings.Split(dir, ":")
				if len(array) != 2 {
					continue
				}
				if number, err := strconv.Atoi(array[1]); err == nil {
					dirs[array[0]] = number
				}
			}
			v = dirs
		} else {
			v = value.Value
		}
		p, _ := json.Marshal(map[string]interface{}{
			args[0]: v,
		})
		fmt.Println(string(p))
		err := api.QbitSetAppPreference(string(p))
		if err != nil {
			return err
		}
		return nil
	}

	return cmd
}
