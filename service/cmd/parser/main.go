package api

package api

import (
"encoding/json"
"time"
)

//type AggregatesResponse struct {
//	rows []aggregateRow
//}
//
//func (ar *AggregatesResponse) AppendRow(bucket time.Time) {
//	ar.rows = append(ar.rows, aggregateRow{
//		Bucket:     bucket,
//		Action:     0,
//		Origin:     "",
//		BrandID:    "",
//		CategoryID: "",
//	})
//}
//
//type aggregateRow struct {
//	Bucket     time.Time
//	Action     Action
//	Origin     string
//	BrandID    string
//	CategoryID string
//}
//
//func (ar AggregatesResponse) MarshalJSON() ([]byte, error) {
//	type Alias AggregatesResponse
//	internal := &struct {
//		*Alias
//		Rows []aggregateRow `json:"rows"`
//	}{
//		Alias: (*Alias)(&ar),
//		Rows:  make([]aggregateRow, 0),
//	}
//	if ar.rows != nil {
//		internal.Rows = ar.rows
//	}
//
//	return json.Marshal(internal)
//}
//
//func (ar aggregateRow) MarshalJSON() ([]byte, error) {
//	internal := []interface{}{
//		ar.Bucket.Format("2006-01-02T15:04:05"), // FIXME
//		ar.Action,
//	}
//	if len(ar.Origin) > 0 {
//		internal = append(internal, ar.Origin)
//	}
//	if len(ar.BrandID) > 0 {
//		internal = append(internal, ar.BrandID)
//	}
//	if len(ar.CategoryID) > 0 {
//		internal = append(internal, ar.CategoryID)
//	}
//
//	return json.Marshal(internal)
//}

//
//func main() {
//	ar := api.AggregatesResponse{
//		Columns: []string{"1m_bucket", "action"},
//	}
//
//	data, err := json.Marshal(ar)
//	if err != nil {
//		fmt.Println("error")
//		os.Exit(1)
//	}
//
//	fmt.Printf("%s", data)
//}
