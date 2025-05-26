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

	return cmd
}

func ItemList() *cobra.Command {
	var cmd = &cobra.Command{
		Use:   "list [flags]",
		Short: "List items",
	}

	var (
		includeItemTypes, excludeItemTypes  string
		searchTerm                          string
		genreIds                            string
		sortBy, sortOrder                   string
		limit, minWidth, maxWidth, parentId int
		hasOverview                         int
	)

	cmd.Flags().StringVar(&sortBy, "sort-by", "", `Options: Album, AlbumArtist, Artist, Budget, 
CommunityRating, CriticRating, DateCreated, DatePlayed, PlayCount, 
PremiereDate, ProductionYear, SortName, Random, Revenue, Runtime`)
	cmd.Flags().StringVar(&sortOrder, "sort-order", "", "Sort Order: Ascending,Descending")
	cmd.Flags().StringVar(&genreIds, "genre-ids", "", "genres id separated by comma")
	cmd.Flags().StringVar(&includeItemTypes, "include-item-types", "", `Comma separated list of item types:
Movie,Series,Video,Game,MusicAlbum,Genres
You may find all types here: https://dev.emby.media/doc/restapi/Item-Types.html`)
	cmd.Flags().StringVar(&excludeItemTypes, "exclude-item-types", "", "exclude item types, same as include-item-types")
	cmd.Flags().StringVar(&searchTerm, "search-term", "", "keywords")
	cmd.Flags().IntVar(&limit, "limit", 50, "results limit")
	cmd.Flags().IntVar(&minWidth, "min-width", 0, "item min width")
	cmd.Flags().IntVar(&maxWidth, "max-width", 0, "item max width")
	cmd.Flags().IntVar(&parentId, "parent-id", 0, "parent id")
	cmd.Flags().IntVar(&hasOverview, "has-overview", -1, "has overview: 1 for yes, 0 for no")

	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		params := url.Values{
			"Limit":  []string{strconv.FormatInt(int64(limit), 10)},
			"Fields": []string{"PremiereDate", "ProductionYear", "Overview", "DateCreated"},
		}
		if hasOverview == 1 {
			params.Add("HasOverview", "true")
		} else if hasOverview == 0 {
			params.Add("HasOverview", "false")
		}
		if includeItemTypes != "" {
			params.Add("Recursive", "true")
			params.Add("IncludeItemTypes", includeItemTypes)
		}
		if excludeItemTypes != "" {
			params.Add("ExcludeItemTypes", excludeItemTypes)
			params.Add("Recursive", "true")
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
		for i, item := range items.Items {
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
