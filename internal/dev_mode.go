package internal

import "os"

var DevMode = os.Getenv("OTF_DEV_MODE") != ""
