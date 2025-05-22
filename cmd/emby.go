package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"net/url"
	"qbit-cli/internal/api"
	"qbit-cli/internal/api/emby"
	"qbit-cli/pkg/utils"
	"strconv"
)

func EmbyCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "emby [command]",
		Short: "Emby management",
	}

	cmd.AddCommand(ItemCmd())

	return cmd
}

func ItemCmd() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "item [command]",
		Short: "Item management",
	}

	cmd.AddCommand(ItemList())
	cmd.AddCommand(ItemInfo())

	return cmd
}

func ItemList() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list [flags]",
		Short: "List items",
	}

	var (
		includeItemTypes                    string
		searchTerm                          string
		genreIds                            string
		sortBy, sortOrder                   string
		limit, minWidth, maxWidth, parentId int
	)

	cmd.Flags().StringVar(&sortBy, "sort-by", "", `Options: Album, AlbumArtist, Artist, Budget, 
CommunityRating, CriticRating, DateCreated, DatePlayed, PlayCount, 
PremiereDate, ProductionYear, SortName, Random, Revenue, Runtime`)
	cmd.Flags().StringVar(&sortOrder, "sort-order", "", "Sort Order: Ascending,Descending")
	cmd.Flags().StringVar(&genreIds, "genre-ids", "", "genres id separated by comma")
	cmd.Flags().StringVar(&includeItemTypes, "include-item-types", "", `Comma separated list of item types:
Movie,Series,Video,Game,MusicAlbum,Genres
You may find all types here: https://dev.emby.media/doc/restapi/Item-Types.html`)
	cmd.Flags().StringVar(&searchTerm, "search-term", "", "keywords")
	cmd.Flags().IntVar(&limit, "limit", 50, "results limit")
	cmd.Flags().IntVar(&minWidth, "min-width", 0, "item min width")
	cmd.Flags().IntVar(&maxWidth, "max-width", 0, "item max width")
	cmd.Flags().IntVar(&parentId, "parent-id", 0, "parent id")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		params := url.Values{
			"Limit":  []string{strconv.FormatInt(int64(limit), 10)},
			"Fields": []string{"PremiereDate", "ProductionYear"},
		}
		if includeItemTypes != "" {
			params.Add("Recursive", "true")
			params.Add("IncludeItemTypes", includeItemTypes)
		}
		if searchTerm != "" {
			params.Add("SearchTerm", searchTerm)
			params.Add("Recursive", "true")
		}
		if sortBy != "" {
			params.Add("SortBy", sortBy)
		}
		if sortOrder != "" {
			params.Add("SortOrder", sortOrder)
		}
		if genreIds != "" {
			params.Add("GenreIds", genreIds)
		}
		if minWidth > 0 {
			params.Add("MinWidth", strconv.Itoa(minWidth))
		}
		if maxWidth > 0 {
			params.Add("MaxWidth", strconv.Itoa(maxWidth))
		}
		if parentId > 0 {
			params.Add("ParentId", strconv.Itoa(parentId))
		}

		items, err := emby.Items(params)
		if err != nil {
			return err
		}
		if items.TotalRecordCount <= 0 {
			return nil
		}

		fmt.Printf("total items: %d\n", items.TotalRecordCount)
		headers := []string{"ID", "Name", "Type", "PremiereDate"}
		var data = make([][]string, len(items.Items))
		for i, item := range items.Items {
			year := ""
			if !item.PremiereDate.IsZero() {
				year = item.PremiereDate.Format("2006-01-02")
			}
			data[i] = []string{item.ID, item.Name, item.Type, year}
		}
		utils.PrintListWithColWidth(headers, &data, map[int]int{1: 50}, false)

		return nil
	}

	return cmd
}

func ItemInfo() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "info <item>",
		Short: "Item info",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("item id is required")
			}
			return nil
		},
	}

	var (
		showSourceList, showStreamList bool
	)
	cmd.Flags().BoolVar(&showSourceList, "source", false, "whether to show media source list")
	cmd.Flags().BoolVar(&showStreamList, "stream", false, "whether to show media stream list")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		item, err := emby.Item(args[0])
		if err != nil {
			return err
		}

		var s []string
		streams := item.MediaVideoStream()
		video := "-"
		if streams != nil && len(*streams) > 0 {
			for _, stream := range *streams {
				s = append(s, stream.DisplayTitle)
			}
			video = fmt.Sprintf("%v", s)
		}

		sourceStr := "-"
		if sources := item.MediaVideoSourceSizeView(); sources != nil && len(sources) > 0 {
			sourceStr = strconv.FormatInt(int64(item.MediaVideoSourceCount()), 10)
		}

		size := "-"
		if s := item.MediaVideoSourceSizeView(); s != nil && len(s) > 0 {
			size = fmt.Sprintf("%v", s)
		}

		header := []string{"ID", "Name", "Type", "Video", "FileCount", "FileSize", "PremiereDate"}
		var data = make([][]string, 1)
		year := ""
		if !item.PremiereDate.IsZero() {
			year = item.PremiereDate.Format("2006-01-02")
		}
		data[0] = []string{item.ID, item.Name, item.Type, video, sourceStr, size, year}
		utils.PrintListWithColWidth(header, &data, map[int]int{1: 50}, false)

		if showSourceList {
			showSources(&item.MediaSources)
		}
		if showStreamList {
			showStreams(&item.MediaSources)
		}

		return nil
	}

	return cmd
}

func showSources(sources *[]api.EmbyMediaSource) {
	if sources == nil {
		return
	}
	fmt.Println("media source list:")
	headers := []string{"ID", "Path", "Size"}
	data := make([][]string, len(*sources))
	for i, source := range *sources {
		data[i] = []string{source.ID, source.Path, utils.FormatFileSizeAuto(source.Size, 1)}
	}
	utils.PrintList(headers, &data)
}

func showStreams(sources *[]api.EmbyMediaSource) {
	if sources == nil {
		return
	}
	fmt.Println("media streams list:")
	headers := []string{"ID", "Index", "Type", "DisplayTitle"}
	data := make([][]string, 0, len(*sources)*2)
	for _, source := range *sources {
		if source.MediaStreams == nil {
			continue
		}
		for _, stream := range source.MediaStreams {
			data = append(data, []string{source.ID, strconv.FormatInt(int64(stream.Index), 10), stream.Type, stream.DisplayTitle})
		}
	}
	utils.PrintList(headers, &data)
}
