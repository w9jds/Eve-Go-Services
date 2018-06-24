package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var (
	client *http.Client
)

// NameRef ESI Name Reference Response
type NameRef struct {
	Category string `json:"category"`
	ID       int32  `json:"id"`
	Name     string `json:"name"`
}

// KillMail ESI KillMail Response
type KillMail struct {
	Victim        Victim     `json:"victim"`
	Attackers     []Attacker `json:"attackers"`
	KillmailID    int32      `json:"killmail_id"`
	KillmailTime  string     `json:"killmail_time"`
	MoonID        int32      `json:"moon_id"`
	SolarSystemID int32      `json:"solar_system_id"`
	WarID         int32      `json:"war_id"`
}

// Character character base structure
type Character struct {
	ID            int32 `json:"character_id"`
	AllianceID    int32 `json:"alliance_id"`
	CorporationID int32 `json:"corporation_id"`
}

// Attacker ESI structure for killmail attacker
type Attacker struct {
	*Character
	DamageDone     int32   `json:"damage_done"`
	FactionID      int32   `json:"faction_id"`
	FinalBlow      bool    `json:"final_blow"`
	SecurityStatus float64 `json:"security_status"`
	ShipTypeID     int32   `json:"ship_type_id"`
	WeaponTypeID   int32   `json:"weapon_type_id"`
}

// Victim ESI structure for killmail victim
type Victim struct {
	*Character
	DamageTaken int32 `json:"damage_taken"`
	FactionID   int32 `json:"faction_id"`
	ShipTypeID  int32 `json:"ship_type_id"`
}

// KillCache base structure from redis cache killmail endpoint
type KillCache struct {
	Package Bundle `json:"package"`
}

// Bundle structure for cache data
type Bundle struct {
	KillID int32 `json:"killID"`
	ZKB    Zkb   `json:"zkb"`
}

// Zkb structure with zkillboard information
type Zkb struct {
	LocationID  float64 `json:"locationID"`
	FittedValue float64 `json:"fittedValue"`
	TotalValue  float64 `json:"totalValue"`
	Href        string  `json:"href"`
}

func getKillLink() Zkb {
	response, error := http.Get("https://redisq.zkillboard.com/listen.php?queueID=ChingyGoBot")
	if error != nil {
		fmt.Println("Invalid Response from Redisq: ", error)
	}

	body, error := ioutil.ReadAll(response.Body)
	if error != nil {
		fmt.Println("Invalid response: ", error)
	}

	var cache KillCache
	error = json.Unmarshal(body, &cache)
	if error != nil {
		fmt.Println("", error)
	}

	return cache.Package.ZKB
}

func getKillMail(uri string) *KillMail {
	request, error := http.NewRequest("GET", uri, nil)
	request.Header.Add("User-Agent", "Killbot - Chingy Chonga/Jeremy Shore - w9jds@live.com")
	request.Header.Add("Accept", "application/json")

	response, error := client.Do(request)
	if error != nil {
		fmt.Println("Unable to get Killmail: ", error)
	}

	body, error := ioutil.ReadAll(response.Body)
	if error != nil {
		fmt.Println("Invalid response: ", error)
	}

	var killMail KillMail
	error = json.Unmarshal(body, &killMail)
	if error != nil {
		fmt.Println("Unable to parse KillMail: ", error)
	}

	return &killMail
}

func getIds(killMail *KillMail) []int32 {
	var unique map[int32]struct{}

	unique = make(map[int32]struct{})
	unique[killMail.Victim.ID] = struct{}{}
	unique[killMail.SolarSystemID] = struct{}{}
	unique[killMail.Victim.ShipTypeID] = struct{}{}
	unique[killMail.Victim.CorporationID] = struct{}{}

	for _, attacker := range killMail.Attackers {
		if _, ok := unique[attacker.ID]; !ok {
			unique[attacker.ID] = struct{}{}
		}
		if _, ok := unique[attacker.ShipTypeID]; !ok {
			unique[attacker.ShipTypeID] = struct{}{}
		}
		if _, ok := unique[attacker.CorporationID]; !ok {
			unique[attacker.CorporationID] = struct{}{}
		}
		if _, ok := unique[attacker.AllianceID]; !ok {
			unique[attacker.AllianceID] = struct{}{}
		}
	}

	var ids []int32
	for key := range unique {
		ids = append(ids, key)
	}

	return ids
}

func getKillMailNames(killMail *KillMail) map[int32]NameRef {
	ids := getIds(killMail)

	buffer, error := json.Marshal(ids)
	if error != nil {
		fmt.Println("Unable serialize list of ids: ", error)
	}

	request, error := http.NewRequest("POST", "https://esi.evetech.net/v2/universe/names/", bytes.NewBuffer(buffer))
	if error != nil {
		fmt.Println("Error creating new request: ", error)
	}

	request.Header.Add("User-Agent", "Killbot - Chingy Chonga/Jeremy Shore - w9jds@live.com")
	request.Header.Add("Accept", "application/json")

	response, error := client.Do(request)
	if error != nil {
		fmt.Println("Unable to get kill mail names: ", error)
	}

	body, error := ioutil.ReadAll(response.Body)
	if error != nil {
		fmt.Println("Invalid response: ", error)
	}

	var names []NameRef
	error = json.Unmarshal(body, &names)
	if error != nil {
		fmt.Println("Unable to parse names: ", error)
	}

	var references map[int32]NameRef
	references = make(map[int32]NameRef)

	for _, ref := range names {
		references[ref.ID] = ref
	}

	return references
}

func buildAttachment(killMail *KillMail, names map[int32]NameRef) {
	systemName := names[killMail.SolarSystemID].Name

	fmt.Println(systemName)
}

func processKillMail(zkb Zkb) {
	killMail := getKillMail(zkb.Href)
	names := getKillMailNames(killMail)

	buildAttachment(killMail, names)
}

func main() {
	client = &http.Client{}

	for {
		zkb := getKillLink()

		if zkb.Href != "" {
			processKillMail(zkb)
		}
	}
}
