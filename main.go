package main

import (
	"context"
	"os"

	"github.com/0xForked/pdf-gen/internal"
	"github.com/spf13/viper"
)

func init() {
	viper.SetConfigFile(".env")
	internal.LoadEnv()
}

func main() {
	// get config instance
	instance := internal.Instance
	// get user input (basic impl)
	var certValItem string
	if len(os.Args) > 1 && os.Args[1] == "course" {
		certValItem = instance.CourseCertVal
	} else if len(os.Args) > 1 && os.Args[1] == "program" {
		certValItem = instance.ProgramCertVal
	} else {
		certValItem = instance.CourseCertVal
	}
	// build app repo & load cert by its key and value
	repo := internal.Repository{BaseURL: instance.BaseURL}
	raw, err := repo.LoadCert(context.Background(),
		instance.CertKey, certValItem)
	if err != nil {
		panic(err)
	}
	// transform it from interface to struct
	cert := internal.PreGenerateCertificate{}
	if err := repo.ToCertStruct(&cert, raw); err != nil {
		panic(err)
	}
	// generate pdf
	if err := cert.GeneratePDF("./gen"); err != nil {
		panic(err)
	}
}
