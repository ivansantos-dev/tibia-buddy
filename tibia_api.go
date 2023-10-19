package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Child of CharacterInfo
type Houses struct {
	Name    string `json:"name"`    // The name of the house.
	Town    string `json:"town"`    // The town where the house is located in.
	Paid    string `json:"paid"`    // The date the last paid rent is due.
	HouseID int    `json:"houseid"` // The internal ID of the house.
}

// Child of CharacterInfo
type CharacterGuild struct {
	GuildName string `json:"name,omitempty"` // The name of the guild.
	Rank      string `json:"rank,omitempty"` // The character's rank in the guild.
}

// Child of Character
type CharacterInfo struct {
	Name              string         `json:"name"`                    // The name of the character.
	FormerNames       []string       `json:"former_names,omitempty"`  // List of former names of the character.
	Traded            bool           `json:"traded,omitempty"`        // Whether the character was traded. (last 6 months)
	DeletionDate      string         `json:"deletion_date,omitempty"` // The date when the character will be deleted. (if scheduled for deletion)
	Sex               string         `json:"sex"`                     // The character's sex.
	Title             string         `json:"title"`                   // The character's selected title.
	UnlockedTitles    int            `json:"unlocked_titles"`         // The number of titles the character has unlocked.
	Vocation          string         `json:"vocation"`                // The character's vocation.
	Level             int            `json:"level"`                   // The character's level.
	AchievementPoints int            `json:"achievement_points"`      // The total of achievement points the character has.
	World             string         `json:"world"`                   // The character's current world.
	FormerWorlds      []string       `json:"former_worlds,omitempty"` // List of former worlds the character was in. (last 6 months)
	Residence         string         `json:"residence"`               // The character's current residence.
	MarriedTo         string         `json:"married_to,omitempty"`    // The name of the character's husband/spouse.
	Houses            []Houses       `json:"houses,omitempty"`        // List of houses the character owns currently.
	Guild             CharacterGuild `json:"guild"`                   // The guild that the character is member of.
	LastLogin         string         `json:"last_login,omitempty"`    // The character's last logged in time.
	Position          string         `json:"position,omitempty"`      // The character's special position.
	AccountStatus     string         `json:"account_status"`          // Whether account is Free or Premium.
	Comment           string         `json:"comment,omitempty"`       // The character's comment.
}

// Child of Character
type AccountBadges struct {
	Name        string `json:"name"`        // The name of the badge.
	IconURL     string `json:"icon_url"`    // The URL to the badge's icon.
	Description string `json:"description"` // The description of the badge.
}

// Child of Character
type Achievements struct {
	Name   string `json:"name"`   // The name of the achievement.
	Grade  int    `json:"grade"`  // The grade/stars of the achievement.
	Secret bool   `json:"secret"` // Whether it is a secret achievement or not.
}

// Child of Deaths
type Killers struct {
	Name   string `json:"name"`   // The name of the killer/assist.
	Player bool   `json:"player"` // Whether it is a player or not.
	Traded bool   `json:"traded"` // If the killer/assist was traded after the death.
	Summon string `json:"summon"` // The name of the summoned creature.
}

// Child of Character
type Deaths struct {
	Time    string    `json:"time"`    // The timestamp when the death occurred.
	Level   int       `json:"level"`   // The level when the death occurred.
	Killers []Killers `json:"killers"` // List of killers involved.
	Assists []Killers `json:"assists"` // List of assists involved.
	Reason  string    `json:"reason"`  // The plain text reason of death.
}

// Child of Character
type AccountInformation struct {
	Position     string `json:"position,omitempty"`      // The account's special position.
	Created      string `json:"created,omitempty"`       // The account's date of creation.
	LoyaltyTitle string `json:"loyalty_title,omitempty"` // The account's loyalty title.
}

// Child of Character
type OtherCharacters struct {
	Name     string `json:"name"`               // The name of the character.
	World    string `json:"world"`              // The name of the world.
	Status   string `json:"status"`             // The status of the character being online or offline.
	Deleted  bool   `json:"deleted"`            // Whether the character is scheduled for deletion or not.
	Main     bool   `json:"main"`               // Whether this is the main character or not.
	Traded   bool   `json:"traded"`             // Whether the character has been traded last 6 months or not.
	Position string `json:"position,omitempty"` // // The character's special position.
}

// Child of JSONData
type Character struct {
	CharacterInfo      CharacterInfo      `json:"character"`                     // The character's information.
	AccountBadges      []AccountBadges    `json:"account_badges,omitempty"`      // The account's badges.
	Achievements       []Achievements     `json:"achievements,omitempty"`        // The character's achievements.
	Deaths             []Deaths           `json:"deaths,omitempty"`              // The character's deaths.
	AccountInformation AccountInformation `json:"account_information,omitempty"` // The account information.
	OtherCharacters    []OtherCharacters  `json:"other_characters,omitempty"`    // The account's other characters.
}

// The base includes two levels, Characters and Information
type CharacterResponse struct {
	Character   Character   `json:"characters"`
	Information interface{} `json:"information"`
}

var url = "https://api.tibiadata.com"

func GetCharacter(name string) (*Character, error) {
	resp, err := http.Get(fmt.Sprintf("%s/v3/character/%s", url, name))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	charResponse := &CharacterResponse{}
	json.NewDecoder(resp.Body).Decode(&charResponse)

	if charResponse.Character.CharacterInfo.Name == "" {
		return nil, fmt.Errorf("Character not found")
	}

	return &charResponse.Character, nil
}

// Child of World
type OnlinePlayers struct {
	Name     string `json:"name"`     // The name of the character.
	Level    int    `json:"level"`    // The character's level.
	Vocation string `json:"vocation"` // The character's vocation.
}

// Child of JSONData
type ApiWorld struct {
	Name                string          `json:"name"`                  // The name of the world.
	Status              string          `json:"status"`                // The current status of the world.
	PlayersOnline       int             `json:"players_online"`        // The number of currently online players.
	RecordPlayers       int             `json:"record_players"`        // The world's online players record.
	RecordDate          string          `json:"record_date"`           // The date when the record was achieved.
	CreationDate        string          `json:"creation_date"`         // The year and month it was created.
	Location            string          `json:"location"`              // The physical location of the servers.
	PvpType             string          `json:"pvp_type"`              // The type of PvP.
	PremiumOnly         bool            `json:"premium_only"`          // Whether only premium account players are allowed to play on it.
	TransferType        string          `json:"transfer_type"`         // The type of transfer restrictions it has. regular (if not present) / locked / blocked
	WorldsQuestTitles   []string        `json:"world_quest_titles"`    // List of world quest titles the server has achieved.
	BattleyeProtected   bool            `json:"battleye_protected"`    // The type of BattlEye protection. true if protected / false if "Not protected by BattlEye."
	BattleyeDate        string          `json:"battleye_date"`         // The date when BattlEye was added. "" if since release / else show date?
	GameWorldType       string          `json:"game_world_type"`       // The type of world. regular / experimental / tournament (if Tournament World Type exists)
	TournamentWorldType string          `json:"tournament_world_type"` // The type of tournament world. "" (default?) / regular / restricted
	OnlinePlayers       []OnlinePlayers `json:"online_players"`        // List of players being currently online.
}

type Worlds struct {
	World ApiWorld `json:"world"`
}

// The base includes two levels: World and Information
type WorldResponse struct {
	Worlds      Worlds      `json:"worlds"`
	Information interface{} `json:"information"`
}

func GetWorld(name string) (*ApiWorld, error) {
	resp, err := http.Get(fmt.Sprintf("%s/v3/world/%s", url, name))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	worldResponse := &WorldResponse{}
	json.NewDecoder(resp.Body).Decode(&worldResponse)
	return &worldResponse.Worlds.World, nil
}
