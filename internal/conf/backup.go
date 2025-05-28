package conf

import (
	"fmt"

	"github.com/bluenviron/mediamtx/internal/conf/jsonwrapper"
)

type Backup struct {
	Path     string         `json:"path"`
	Replicas BackupReplicas `json:"replicas"`
}

type Backups []Backup

type BackupReplica struct {
	Type            string `json:"type"`
	Bucket          string `json:"bucket"`
	Path            string `json:"path"`
	Region          string `json:"region"`
	Endpoint        string `json:"endpoint"`
	AccessKeyId     string `json:"accessKeyId"`
	SecretAccessKey string `json:"secretAccessKey"`
	ForcePathStyle  bool   `json:"forcePathStyle"`
	StorageClass    string `json:"storageClass"`
}

type BackupReplicas []*BackupReplica

// UnmarshalJSON implements json.Unmarshaler.
func (br *BackupReplica) UnmarshalJSON(b []byte) error {
	type alias BackupReplica
	if err := jsonwrapper.Unmarshal(b, (*alias)(br)); err != nil {
		return err
	}

	switch br.Type {
	case "s3":
	default:
		return fmt.Errorf("invalid backup type '%s'", br.Type)
	}

	return nil
}
