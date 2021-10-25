package timeseries

type MDataPoint struct {
	Name      string                 `json:"name"` // name of the measurement
	Tags      map[string]string      `json:"tags"`
	Fields    map[string]interface{} `json:"fields"`
	TimeStamp int64                  `json:"ts"` // if 0 , ecollector will set local time
}

type WriteDataPointsRequest struct {
	ProcID     int          `json:"proc_id"`
	Bucket     string       `json:"bucket"` // data storage with retention policy , if not set system will try to auto calculate based on measurement name
	DataPoints []MDataPoint `json:"dp"`
}

type DataPointsFilter struct {
	Tags      map[string]string `json:"tags"`
	Devices   []string          `json:"devices"`
	Locations []string          `json:"locations"`
	DevTypes  []string          `json:"dev_types"`
}

type GetDataPointsRequest struct {
	ProcID            int              `json:"proc_id"`
	FieldName         string           `json:"field_name"`
	DataFunction      string           `json:"data_function"`
	TransformFunction string           `json:"transform_function"`
	MeasurementName   string           `json:"measurement_name"`
	RelativeTime      string           `json:"relative_time"`
	FromTime          string           `json:"from_time"`
	ToTime            string           `json:"to_time"`
	GroupByTime       string           `json:"group_by_time"`
	GroupByTag        string           `json:"group_by_tag"`
	FillType          string           `json:"fill_type"`
	Filters           DataPointsFilter `json:"filters"`
}


// Response represents a list of statement results.
type Response struct {
	Results []Result
	Err     string `json:"error,omitempty"`
}

// Result represents a resultset returned from a single statement.
type Result struct {
	Series   []Row
	Err      string `json:"error,omitempty"`
}

// Row represents a single row returned from the execution of a statement.
type Row struct {
	Name    string            `json:"name,omitempty"`
	Tags    map[string]string `json:"tags,omitempty"`
	Columns []string          `json:"columns,omitempty"`
	Values  [][]interface{}   `json:"values,omitempty"`
	Partial bool              `json:"partial,omitempty"`
}
