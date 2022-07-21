package api

type AggregatesResponse struct {
	Columns []string   `json:"columns"`
	Rows    [][]string `json:"rows"`
}
