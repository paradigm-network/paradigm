package main

import (
	"fmt"
	"github.com/paradigm-network/paradigm/common/crypto"
	"github.com/paradigm-network/paradigm/version"
	"gopkg.in/urfave/cli.v1"
	"os"
)

func main() {
	app := cli.NewApp()
	app.Name = "paradigm"
	app.Usage = "Paradigm Network"
	app.HideVersion = true //there is a special command to print the version
	app.Commands = []cli.Command{
		{
			Name:   "keygen",
			Usage:  "Dump new key pair",
			Action: keygen,
		},
		{
			Name:   "run",
			Usage:  "Run paradigm",
			Action: nil,
			Flags:  []cli.Flag{},
		},
		{
			Name:   "version",
			Usage:  "Show version info",
			Action: printVersion,
		},
	}
	app.Run(os.Args)
}

func keygen(c *cli.Context) error {
	pemDump, err := crypto.GeneratePemKey()
	if err != nil {
		fmt.Println("Error generating PemDump")
		os.Exit(2)
	}

	fmt.Println("PublicKey:")
	fmt.Println(pemDump.PublicKey)
	fmt.Println("PrivateKey:")
	fmt.Println(pemDump.PrivateKey)

	return nil
}
func printVersion(c *cli.Context) error {
	fmt.Println(version.Version)
	return nil
}
