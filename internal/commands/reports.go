package commands

import (
	"net/http"
	"net/url"

	"github.com/spf13/cobra"
)

var reportFilters = []string{"duration", "start-date=start_date", "end-date=end_date"}

type report struct {
	use     string
	path    string
	key     string
	short   string
	clear   string
	paged   bool
	urlOnly bool
}

var reports = []report{
	{use: "popular-searches", path: "/reports/popular_searches", key: "popular_searches", short: "Show the most searched keywords", clear: "Delete the popular searches data for the chosen period, or all of it when no period is given", paged: true},
	{use: "failed-searches", path: "/reports/failed_searches", key: "failed_searches", short: "Show searches that found nothing", clear: "Delete the failed searches data for the chosen period, or all of it when no period is given", paged: true},
	{use: "invalid-urls", path: "/reports/invalid_urls", key: "invalid_urls", short: "Show visited URLs that do not exist", clear: "Delete the invalid URLs data for the chosen period or URL, or all of it when no filter is given", paged: true, urlOnly: true},
	{use: "visits", path: "/reports/kb_visits", key: "visits", short: "Show knowledge base visits per day, oldest first"},
}

var reportsCmd = &cobra.Command{
	Use:   "reports",
	Short: "Show search and visit reports",
}

var reportsClearCmd = &cobra.Command{
	Use:   "clear",
	Short: "Delete report data; this cannot be undone",
}

func addReportFlags(cmd *cobra.Command, r report) {
	cmd.Flags().String("duration", "", "Period: all_time (default), last_24_hours, last_7_days, this_week, last_week, last_30_days, this_month, last_month, this_quarter, last_quarter or custom")
	cmd.Flags().String("start-date", "", "Start date (ISO 8601); needed with --duration custom")
	cmd.Flags().String("end-date", "", "End date (ISO 8601); needed with --duration custom")
	if r.urlOnly {
		cmd.Flags().String("url", "", "Only include this URL")
	}
}

func reportFilterSpecs(r report) []string {
	if r.urlOnly {
		return append([]string{"url"}, reportFilters...)
	}
	return reportFilters
}

func init() {
	for _, r := range reports {
		show := &cobra.Command{
			Use:   r.use,
			Short: r.short,
			RunE: func(cmd *cobra.Command, args []string) error {
				params := url.Values{}
				if r.paged {
					params = paginationParams(cmd)
				}
				data, err := fetch(cmd, r.path, queryValues(cmd, params, reportFilterSpecs(r)...))
				if err != nil {
					return err
				}

				printList(data, r.key, nil)
				return nil
			},
		}
		addReportFlags(show, r)
		if r.paged {
			addPaginationFlags(show)
		}
		reportsCmd.AddCommand(show)

		if r.clear == "" {
			continue
		}
		clearCmd := &cobra.Command{
			Use:   r.use,
			Short: r.clear,
			RunE: func(cmd *cobra.Command, args []string) error {
				data, err := send(cmd, http.MethodDelete, r.path, flagValues(cmd, reportFilterSpecs(r)...))
				if err != nil {
					return err
				}

				printActionResult(data, nil)
				return nil
			},
		}
		addReportFlags(clearCmd, r)
		reportsClearCmd.AddCommand(clearCmd)
	}

	reportsCmd.AddCommand(reportsClearCmd)
	register(func(root *cobra.Command) { root.AddCommand(reportsCmd) })
}
