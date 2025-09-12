package cmd

import (
	"errors"
	"fmt"
	"github.com/spf13/cobra"
	"log"
	"net/url"
	"qbit-cli/internal/api"
	"qbit-cli/internal/api/emby"
	"qbit-cli/pkg/utils"
	"strconv"
	"strings"
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
	cmd.AddCommand(ItemRefreshCommand())

	return cmd
}

var refreshMode = []string{"FullRefresh", "Default", "ValidationOnly"}

func ItemRefreshCommand() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "refresh <item>...",
		Short: "Refresh items",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires item id")
			}
			return nil
		},
	}

	var (
		recursive                            bool
		replaceAllMetadata, replaceAllImages bool
		resetBeforeRefresh                   bool
	)

	metadataRefreshMode := FlagsProperty[string]{Flag: "metadata-refresh-mode", Options: refreshMode}
	imageRefreshMode := FlagsProperty[string]{Flag: "image-refresh-mode", Options: refreshMode}

	cmd.Flags().BoolVar(&recursive, "recursive", true, "Indicates if the refresh should occur recursively.")
	cmd.Flags().StringVar(&imageRefreshMode.Value, "image-refresh-mode", "FullRefresh", "image refresh mode: "+strings.Join(refreshMode, ","))
	cmd.Flags().StringVar(&metadataRefreshMode.Value, "metadata-refresh-mode", "FullRefresh", "metadata refresh mode: "+strings.Join(refreshMode, ","))
	cmd.Flags().BoolVar(&replaceAllMetadata, "replace-all-metadata", true, "Determines if metadata should be replaced. Only applicable if mode is FullRefresh")
	cmd.Flags().BoolVar(&replaceAllImages, "replace-all-images", true, "Determines if images should be replaced. Only applicable if mode is FullRefresh")
	cmd.Flags().BoolVar(&resetBeforeRefresh, "reset-before-refresh", false, "Reset all metadata before refresh.")

	metadataRefreshMode.RegisterCompletion(cmd)
	imageRefreshMode.RegisterCompletion(cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		params := url.Values{
			"Recursive":           {strconv.FormatBool(recursive)},
			"MetadataRefreshMode": {metadataRefreshMode.Value},
			"ImageRefreshMode":    {imageRefreshMode.Value},
			"ReplaceAllMetadata":  {strconv.FormatBool(replaceAllMetadata)},
			"ReplaceAllImages":    {strconv.FormatBool(replaceAllImages)},
		}
		for _, arg := range args {
			if resetBeforeRefresh {
				err := emby.ResetItemMetadata(arg)
				if err != nil {
					fmt.Printf("%s reset failed: %s\n", arg, err)
				}
			}
			err := emby.RefreshItem(arg, params)
			if err != nil {
				fmt.Printf("%s refresh failed: %s\n", arg, err)
			}
		}

		return nil
	}

	return cmd
}

var itemTypes = []string{"Audio", "Video", "Folder", "Episode", "Movie", "Trailer", "AdultVideo", "MusicVideo",
	"BoxSet", "MusicAlbum", "MusicArtist", "Season", "Series", "Game", "GameSystem", "Book"}
var sortOrders = []string{"Ascending", "Descending"}
var sortBys = []string{"Album", "AlbumArtist", "Artist", "Budget", "CommunityRating", "CriticRating", "DateCreated",
	"DatePlayed", "PlayCount", "PremiereDate", "ProductionYear", "SortName", "Random", "Revenue", "Runtime"}

func ItemList() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list [flags]",
		Short: "List items",
	}

	var (
		searchTerm                                 string
		genreIds                                   string
		start, limit, minWidth, maxWidth, parentId int
		hasOverview                                int
	)
	includeItemTypes := FlagsProperty[string]{Flag: "include-item-types", Options: itemTypes}
	excludeItemTypes := FlagsProperty[string]{Flag: "exclude-item-types", Options: itemTypes}
	sortOrder := FlagsProperty[string]{Flag: "sort-order", Options: sortOrders}
	sortBy := FlagsProperty[string]{Flag: "sort-by", Options: sortBys}

	cmd.Flags().StringVar(&sortBy.Value, sortBy.Flag, "", `Options: `+strings.Join(sortBys, ","))
	cmd.Flags().StringVar(&sortOrder.Value, sortOrder.Flag, "", "Sort Order: "+strings.Join(sortOrders, "|"))
	cmd.Flags().StringVar(&includeItemTypes.Value, includeItemTypes.Flag, "", `Comma separated list of item types: `+strings.Join(itemTypes, ","))
	cmd.Flags().StringVar(&excludeItemTypes.Value, excludeItemTypes.Flag, "", "Comma separated list of item types: "+strings.Join(itemTypes, ","))
	cmd.Flags().StringVar(&genreIds, "genre-ids", "", "genres id separated by comma")
	cmd.Flags().StringVar(&searchTerm, "search-term", "", "keywords")
	cmd.Flags().IntVar(&limit, "limit", 50, "results limit")
	cmd.Flags().IntVar(&start, "start", 0, "results start index")
	cmd.Flags().IntVar(&minWidth, "min-width", 0, "item min width")
	cmd.Flags().IntVar(&maxWidth, "max-width", 0, "item max width")
	cmd.Flags().IntVar(&parentId, "parent-id", 0, "parent id")
	cmd.Flags().IntVar(&hasOverview, "has-overview", -1, "has overview: 1 for yes, 0 for no")

	// register completion
	includeItemTypes.RegisterCompletion(cmd)
	excludeItemTypes.RegisterCompletion(cmd)
	sortBy.RegisterCompletion(cmd)
	sortOrder.RegisterCompletion(cmd)

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		params := url.Values{
			"Limit":      []string{strconv.FormatInt(int64(limit), 10)},
			"StartIndex": []string{strconv.FormatInt(int64(start), 10)},
			"Fields":     []string{"PremiereDate", "ProductionYear", "Overview", "DateCreated", "People", "ProviderIds"},
		}
		if hasOverview == 1 {
			params.Add("HasOverview", "true")
		} else if hasOverview == 0 {
			params.Add("HasOverview", "false")
		}
		if includeItemTypes.Value != "" {
			params.Add("Recursive", "true")
			params.Add("IncludeItemTypes", includeItemTypes.Value)
		}
		if excludeItemTypes.Value != "" {
			params.Add("ExcludeItemTypes", excludeItemTypes.Value)
			params.Add("Recursive", "true")
		}
		if searchTerm != "" {
			params.Add("SearchTerm", searchTerm)
			params.Add("Recursive", "true")
		}
		if sortBy.Value != "" {
			params.Add("SortBy", sortBy.Value)
		}
		if sortOrder.Value != "" {
			params.Add("SortOrder", sortOrder.Value)
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
		if len(items.Items) <= 0 {
			return nil
		}

		size := items.TotalRecordCount
		if size <= 0 {
			size = len(items.Items)
		}
		fmt.Printf("total items: %d\n", size)
		headers := []string{"ID", "Name", "Type", "IDX", "Created"}
		var data = make([][]string, len(items.Items))
		var noPeopleList []api.EmbyItem
		var noBackdropList []api.EmbyItem
		for i, item := range items.Items {
			if item.People == nil || len(item.People) <= 0 {
				noPeopleList = append(noPeopleList, item)
			}
			if item.BackdropImageTags == nil || len(item.BackdropImageTags) <= 0 {
				noBackdropList = append(noBackdropList, item)
			}
			create := ""
			if !item.CreatedDate.IsZero() {
				create = item.CreatedDate.Format("2006-01-02")
			}
			idx := ""
			if item.IndexNumber > 0 {
				idx = strconv.Itoa(item.IndexNumber)
			}
			data[i] = []string{item.ID, item.Name, item.Type, idx, create}
		}
		utils.PrintListWithColWidth(headers, &data, map[int]int{1: 50}, false)
		for _, item := range noPeopleList {
			log.Printf("no people found: [%s] %s\n", item.ID, item.Name)
		}
		for _, item := range noBackdropList {
			log.Printf("no backdrop found: [%s] %s\n", item.ID, item.Name)
		}

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
		showSourceList, showStreamList, getChildren bool
		childrenLimit                               int
	)
	cmd.Flags().BoolVar(&showSourceList, "source", false, "whether to show media source list")
	cmd.Flags().BoolVar(&showStreamList, "stream", false, "whether to show media stream list")
	cmd.Flags().BoolVar(&getChildren, "children", false, "whether to show children items")
	cmd.Flags().IntVar(&childrenLimit, "children-limit", 100, "limit number of children")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		item, err := emby.Item(args[0])
		if err != nil {
			return err
		}

		header := []string{"ID", "ParentId", "Name", "Type", "PremiereDate", "Path"}
		var data = make([][]string, 1)
		premiereDate := ""
		if !item.PremiereDate.IsZero() {
			premiereDate = item.PremiereDate.Format("2006-01-02")
		}
		data[0] = []string{item.ID, item.ParentId, item.Name, item.Type, premiereDate, item.Path}

		genres := strings.Join(item.Genres, ",")
		childrenType := ""
		if item.IsMovie() {
			header = append(header, "Genres")
			data[0] = append(data[0], genres)
		}
		if item.IsSeries() {
			header = append(header, "Genres", "Seasons")
			data[0] = append(data[0], genres, strconv.Itoa(item.ChildCount))
			childrenType = "Season"
		}
		if item.IsSeason() {
			header = append(header, "Episodes")
			data[0] = append(data[0], strconv.Itoa(item.ChildCount))
			childrenType = "Episode"
		}
		if item.IsMusicAlbum() {
			header = append(header, "Genres", "AlbumArtist")
			data[0] = append(data[0], genres, item.AlbumArtist)
			childrenType = "Audio"
		}
		if item.IsMusicCollection() {
			childrenType = "MusicAlbum"
		}
		if item.IsMovieCollection() {
			childrenType = "Movie"
		}
		if item.IsTVShowCollection() {
			childrenType = "Series"
		}
		if item.IsLiveTV() {
			childrenType = "TvChannel"
		}

		utils.PrintListWithColWidth(header, &data, map[int]int{2: 30}, false)

		if getChildren {
			subCmd := ItemList()
			_ = subCmd.Flags().Set("parent-id", item.ID)
			_ = subCmd.Flags().Set("limit", strconv.Itoa(childrenLimit))
			if childrenType != "" {
				_ = subCmd.Flags().Set("include-item-types", childrenType)
			}
			subCmd.SetArgs([]string{})
			_ = subCmd.Execute()
		}

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
	if len(*sources) == 0 {
		return
	}
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
	if len(*sources) == 0 {
		return
	}
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
