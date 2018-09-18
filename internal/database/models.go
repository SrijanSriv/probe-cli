package database

import (
	"database/sql"
	"os"
	"path/filepath"
	"time"

	"github.com/pkg/errors"
	"upper.io/db.v3/lib/sqlbuilder"
)

// ResultNetwork is used to represent the structure made from the JOIN
// between the results and networks tables.
type ResultNetwork struct {
	Result  `db:",inline"`
	Network `db:",inline"`
}

// MeasurementURLNetwork is used for the JOIN between Measurement and URL
type MeasurementURLNetwork struct {
	Measurement `db:",inline"`
	Network     `db:",inline"`
	Result      `db:",inline"`
	URL         `db:",inline"`
}

// Network represents a network tested by the user
type Network struct {
	ID          int64  `db:"network_id,omitempty"`
	NetworkName string `db:"network_name"`
	NetworkType string `db:"network_type"`
	IP          string `db:"ip"`
	ASN         uint   `db:"asn"`
	CountryCode string `db:"network_country_code"`
}

// URL represents URLs from the testing lists
type URL struct {
	ID           sql.NullInt64  `db:"url_id,omitempty"`
	URL          sql.NullString `db:"url"`
	CategoryCode sql.NullString `db:"category_code"`
	CountryCode  sql.NullString `db:"url_country_code"`
}

// Measurement model
type Measurement struct {
	ID               int64          `db:"measurement_id,omitempty"`
	TestName         string         `db:"test_name"`
	StartTime        time.Time      `db:"measurement_start_time"`
	Runtime          float64        `db:"measurement_runtime"` // Fractional number of seconds
	IsDone           bool           `db:"measurement_is_done"`
	IsUploaded       bool           `db:"measurement_is_uploaded"`
	IsFailed         bool           `db:"measurement_is_failed"`
	FailureMsg       sql.NullString `db:"measurement_failure_msg,omitempty"`
	IsUploadFailed   bool           `db:"measurement_is_upload_failed"`
	UploadFailureMsg sql.NullString `db:"measurement_upload_failure_msg,omitempty"`
	IsRerun          bool           `db:"measurement_is_rerun"`
	ReportID         sql.NullString `db:"report_id,omitempty"`
	URLID            sql.NullInt64  `db:"url_id,omitempty"` // Used to reference URL
	MeasurementID    sql.NullInt64  `db:"collector_measurement_id,omitempty"`
	IsAnomaly        sql.NullBool   `db:"is_anomaly,omitempty"`
	// FIXME we likely want to support JSON. See: https://github.com/upper/db/issues/462
	TestKeys       string `db:"test_keys"`
	ResultID       int64  `db:"result_id"`
	ReportFilePath string `db:"report_file_path"`
}

// Result model
type Result struct {
	ID             int64     `db:"result_id,omitempty"`
	TestGroupName  string    `db:"test_group_name"`
	StartTime      time.Time `db:"result_start_time"`
	NetworkID      int64     `db:"network_id"`     // Used to include a Network
	Runtime        float64   `db:"result_runtime"` // Runtime is expressed in fractional seconds
	IsViewed       bool      `db:"result_is_viewed"`
	IsDone         bool      `db:"result_is_done"`
	DataUsageUp    float64   `db:"result_data_usage_up"`
	DataUsageDown  float64   `db:"result_data_usage_down"`
	MeasurementDir string    `db:"measurement_dir"`
}

// PerformanceTestKeys is the result summary for a performance test
type PerformanceTestKeys struct {
	Upload   float64 `json:"upload"`
	Download float64 `json:"download"`
	Ping     float64 `json:"ping"`
	Bitrate  float64 `json:"median_bitrate"`
}

// Finished marks the result as done and sets the runtime
func (r *Result) Finished(sess sqlbuilder.Database) error {
	if r.IsDone == true || r.Runtime != 0 {
		return errors.New("Result is already finished")
	}
	r.Runtime = time.Now().UTC().Sub(r.StartTime).Seconds()
	r.IsDone = true

	err := sess.Collection("results").Find("result_id", r.ID).Update(r)
	if err != nil {
		return errors.Wrap(err, "updating finished result")
	}
	return nil
}

// Failed writes the error string to the measurement
func (m *Measurement) Failed(sess sqlbuilder.Database, failure string) error {
	m.FailureMsg = sql.NullString{String: failure, Valid: true}
	m.IsFailed = true
	err := sess.Collection("measurements").Find("measurement_id", m.ID).Update(m)
	if err != nil {
		return errors.Wrap(err, "updating measurement")
	}
	return nil
}

// Done marks the measurement as completed
func (m *Measurement) Done(sess sqlbuilder.Database) error {
	runtime := time.Now().UTC().Sub(m.StartTime)
	m.Runtime = runtime.Seconds()
	m.IsDone = true

	err := sess.Collection("measurements").Find("measurement_id", m.ID).Update(m)
	if err != nil {
		return errors.Wrap(err, "updating measurement")
	}
	return nil
}

// UploadFailed writes the error string for the upload failure to the measurement
func (m *Measurement) UploadFailed(sess sqlbuilder.Database, failure string) error {
	m.UploadFailureMsg = sql.NullString{String: failure, Valid: true}
	m.IsUploaded = false

	err := sess.Collection("measurements").Find("measurement_id", m.ID).Update(m)
	if err != nil {
		return errors.Wrap(err, "updating measurement")
	}
	return nil
}

// UploadSucceeded writes the error string for the upload failure to the measurement
func (m *Measurement) UploadSucceeded(sess sqlbuilder.Database) error {
	m.IsUploaded = true

	err := sess.Collection("measurements").Find("measurement_id", m.ID).Update(m)
	if err != nil {
		return errors.Wrap(err, "updating measurement")
	}
	return nil
}

// AddToResult adds a measurement to a result
func (m *Measurement) AddToResult(sess sqlbuilder.Database, result *Result) error {
	var err error

	m.ResultID = result.ID
	finalPath := filepath.Join(result.MeasurementDir,
		filepath.Base(m.ReportFilePath))

	// If the finalPath already exists, it means it has already been moved there.
	// This happens in multi input reports
	if _, err = os.Stat(finalPath); os.IsNotExist(err) {
		err = os.Rename(m.ReportFilePath, finalPath)
		if err != nil {
			return errors.Wrap(err, "moving report file")
		}
	}
	m.ReportFilePath = finalPath

	err = sess.Collection("measurements").Find("measurement_id", m.ID).Update(m)
	if err != nil {
		return errors.Wrap(err, "updating measurement")
	}
	return nil
}
