package main

import (
	"encoding/json"
	. "github.com/runi95/wcmaul-slk-analysis-tool/logger"
	"github.com/runi95/wts-parser/models"
	"github.com/runi95/wts-parser/parser"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var baseUnitMap map[string]*models.SLKUnit
var unitFuncMap map[string]*models.UnitFunc
var builderIdList []string
var allUnitsMap map[string]bool
var logger *Logger
var whitelist map[string]bool

func main() {
	logger = new(Logger)

	if len(os.Args) == 2 {
		inputFolder := os.Args[1]
		allUnitsMap = make(map[string]bool)
		whitelist = make(map[string]bool)
		populateWhitelist()
		loadSLK(inputFolder)
		findBuilders()
		findAllBuildableTowers()
		log.Println("Checking for unbuildable towers...")
		checkForUnbuildableTowers()
		log.Println("Checking for builders with missing classifications...")
		checkForBuildersMissingClassification()
	} else {
		log.Printf("Expected 1 argument (input), but got %d\n", len(os.Args)-1)
	}
}

func populateWhitelist() {
	file, err := ioutil.ReadFile("whitelist.json")
	if err != nil {
		log.Println(err)
		os.Exit(10)
	}

	var list []string
	err = json.Unmarshal([]byte(file), &list)
	if err != nil {
		log.Println(err)
		os.Exit(10)
	}

	for _, id := range list {
		whitelist[id] = true
	}
}

func checkForUnbuildableTowers() {
	for key := range baseUnitMap {
		if !allUnitsMap[key] && !whitelist[key] && baseUnitMap[key].UnitBalance.Isbldg.String == "1" {
			var editorSuffix = ""
			if unitFuncMap[key].Editorsuffix.Valid {
				editorSuffix = " " + unitFuncMap[key].Editorsuffix.String
			}
			logger.Warning(unitFuncMap[key].Name.String + editorSuffix + " (" + key + ")" + " is unbuildable")
		}
	}
}

func findAllUpgradeableTowers(unitId string) {
	unitFunc := unitFuncMap[unitId]
	if unitFunc == nil {
		return
	}

	if unitFunc.Upgrade.Valid {
		unitUpgrades := unitFunc.Upgrade.String
		upgrades := strings.Split(unitUpgrades, ",")
		for _, val := range upgrades {
			if !allUnitsMap[val] {
				allUnitsMap[val] = true
				findAllUpgradeableTowers(val)
			}
		}
	}
}

func findAllBuildableTowers() {
	for _, val := range builderIdList {
		unitFunc := unitFuncMap[val]
		var builds []string
		unitBuilds := unitFunc.Builds.String
		builds = strings.Split(strings.Trim(unitBuilds, "\""), ",")
		for _, val := range builds {
			if !allUnitsMap[val] {
				allUnitsMap[val] = true
				findAllUpgradeableTowers(val)
			}
		}
	}
}

/*

				*/

func findBuilders() {
	for key := range baseUnitMap {
		unitTypes := baseUnitMap[key].UnitBalance.Type.String
		split := strings.Split(strings.Trim(unitTypes, "\""), ",")
		for _, unitType := range split {
			if strings.ToLower(unitType) == "peon" {
				builderIdList = append(builderIdList, key)
			}
		}
	}
}

func loadSLK(inputFolder string) {
	log.Println("Reading UnitAbilities.slk...")

	unitAbilitiesBytes, err := ioutil.ReadFile(filepath.Join(inputFolder, "UnitAbilities.slk"))
	if err != nil {
		log.Println(err)
		os.Exit(10)
	}

	unitAbilitiesMap := parser.SlkToUnitAbilities(unitAbilitiesBytes)

	log.Println("Reading UnitData.slk...")

	unitDataBytes, err := ioutil.ReadFile(filepath.Join(inputFolder, "UnitData.slk"))
	if err != nil {
		log.Println(err)
		os.Exit(10)
	}

	unitDataMap := parser.SlkToUnitData(unitDataBytes)

	log.Println("Reading UnitUI.slk...")

	unitUIBytes, err := ioutil.ReadFile(filepath.Join(inputFolder, "UnitUI.slk"))
	if err != nil {
		log.Println(err)
		os.Exit(10)
	}

	unitUIMap := parser.SLKToUnitUI(unitUIBytes)

	log.Println("Reading UnitWeapons.slk...")

	unitWeaponsBytes, err := ioutil.ReadFile(filepath.Join(inputFolder, "UnitWeapons.slk"))
	if err != nil {
		log.Println(err)
		os.Exit(10)
	}

	unitWeaponsMap := parser.SLKToUnitWeapons(unitWeaponsBytes)

	log.Println("Reading UnitBalance.slk...")

	unitBalanceBytes, err := ioutil.ReadFile(filepath.Join(inputFolder, "UnitBalance.slk"))
	if err != nil {
		log.Println(err)
		os.Exit(10)
	}

	unitBalanceMap := parser.SLKToUnitBalance(unitBalanceBytes)

	log.Println("Reading CampaignUnitFunc.txt...")

	campaignUnitFuncBytes, err := ioutil.ReadFile(filepath.Join(inputFolder, "CampaignUnitFunc.txt"))
	if err != nil {
		log.Println(err)
		os.Exit(10)
	}

	unitFuncMap = parser.TxtToUnitFunc(campaignUnitFuncBytes)

	baseUnitMap = make(map[string]*models.SLKUnit)
	i := 0
	for k := range unitDataMap {
		slkUnit := new(models.SLKUnit)
		slkUnit.UnitAbilities = unitAbilitiesMap[k]
		slkUnit.UnitData = unitDataMap[k]
		slkUnit.UnitUI = unitUIMap[k]
		slkUnit.UnitWeapons = unitWeaponsMap[k]
		slkUnit.UnitBalance = unitBalanceMap[k]

		unitId := strings.Replace(k, "\"", "", -1)
		baseUnitMap[unitId] = slkUnit
		i++
	}
}

func checkForBuildersMissingClassification() {
	for key := range baseUnitMap {
		if strings.Trim(strings.ToLower(baseUnitMap[key].UnitData.Race.String), "\"") == "undead" && baseUnitMap[key].UnitBalance.Isbldg.String == "0" {
			commaSeparatedUnitTypes := strings.Trim(baseUnitMap[key].UnitBalance.Type.String, "\"")
			unitTypes := strings.Split(commaSeparatedUnitTypes, ",")

			isWorker := false
			for _, unitType := range unitTypes {
				if strings.ToLower(unitType) == "peon" {
					isWorker = true
				}
			}

			if !isWorker {
				var editorSuffix = ""
				if unitFuncMap[key].Editorsuffix.Valid {
					editorSuffix = " " + unitFuncMap[key].Editorsuffix.String
				}
				logger.Warning(unitFuncMap[key].Name.String + editorSuffix + " (" + key + ")" + " is a potential builder without the correct classes")
			}
		}
	}
}
