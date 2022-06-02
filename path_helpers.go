package otf

import "fmt"

func UploadConfigurationVersionPath(cv *ConfigurationVersion) string {
	return fmt.Sprintf("/configuration-versions/%s/upload", cv.ID())
}
