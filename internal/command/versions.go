package command

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"

	"github.com/cyberark/summon/pkg/summon"
	"github.com/cyberark/summon/provider"
)
//Create the AllVersions string to include versions of all the providers
const SUMMON_VERSION = summon.VERSION
var DefaultPath = provider.DefaultPath
var AllVersions = GetVersions(GetProviderFileNames())

//Run `$provider --version`(from DefaultPath),
// 	for each provider in the list of providers
//Then appends the responded versions to string and returns it
func GetVersions(providers []string) string{
	var toReturn = SUMMON_VERSION
	for _, provider := range providers {
		version, err := exec.Command(path.Join(DefaultPath, provider), "--version").Output()
		if err != nil{
			toReturn = toReturn + "\n"+provider+" version "+"ERROR"
			fmt.Println("ERROR: "+provider, err, "running command :"+DefaultPath+"/"+ provider, "--version")	//debug
			continue
		}
		toReturn = toReturn + "\n"+provider+" version "+string(version[0:len(version)-1])
	}
	return toReturn
}

//Create slice of all file names in the default path
func GetProviderFileNames() []string{
	files, err := ioutil.ReadDir(DefaultPath)
    if err != nil {
        fmt.Println(err.Error())
		os.Exit(127)
	}
	names := make([]string, len(files))
	for i, file := range files {
        names[i] = file.Name()
    }

    return names
}
