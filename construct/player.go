package construct

import (
	"bufio"
	"context"
	"crypto/sha512"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"strings"
	"sync"
	"time"

	"github.com/Cidan/gomud/color"
	"github.com/Cidan/gomud/config"
	"github.com/rs/zerolog/log"
	uuid "github.com/satori/go.uuid"
)

func hashPassword(pw string) string {
	h := sha512.New()
	io.WriteString(h, pw)
	return hex.EncodeToString(h.Sum(nil))
}

// Player construct
type Player struct {
	connection    net.Conn
	input         chan string //*bufio.Reader
	Data          *playerData
	gameInterp    *Game
	buildInterp   *BuildInterp
	loginInterp   *Login
	currentInterp Interp
	inRoom        *Room
	textBuffer    string
	flagMutex     *sync.RWMutex
	ctx           context.Context
	cancel        context.CancelFunc
}

// This is the main data construct for a human player. Any new flags, attributes
// or other data that needs to carry over between sessions, goes here. Do not
// use this field as storage for temporary variables. Use the Player struct
// above for temporary data that does not need to be saved.
// Additionally, all player fields must be exported in order to be saved.
type playerData struct {
	UUID     string
	Name     string
	Password string
	Room     string
	Flags    map[string]bool
	Prompt   string
	Stats    *playerStats
}

// TODO(lobato): use consts instead of strings.
type playerStats struct {
	Health    int64
	Mana      int64
	Move      int64
	MaxHealth int64
	MaxMana   int64
	MaxMove   int64
}

type roomWalk struct {
	room *Room
	mx   int64
	my   int64
}

// NewPlayer constructs a new player
func NewPlayer() *Player {
	ctx, cancel := context.WithCancel(context.Background())
	p := &Player{
		Data: &playerData{
			UUID:  uuid.NewV4().String(),
			Flags: make(map[string]bool),
			Stats: &playerStats{},
		},
		input:     make(chan string),
		flagMutex: new(sync.RWMutex),
		ctx:       ctx,
		cancel:    cancel,
	}
	p.setDefaults()
	return p
}

// setDefaults sets various defaults for new players.
func (p *Player) setDefaults() {
	p.EnableFlag("prompt")
	p.EnableFlag("color")
	p.DisableFlag("automap")
	p.SetPrompt("<%h{gh{x %m{bm{x %v{yv{x>")
	p.ModifyStat("health", 100, false)
	p.ModifyStat("mana", 100, false)
	p.ModifyStat("move", 100, false)
	p.ModifyStat("max_health", 100, false)
	p.ModifyStat("max_mana", 100, false)
	p.ModifyStat("max_move", 100, false)
}

// playerTick is this specific player's tick timer. This is where you add
// effects such as combat actions, damage dealt, status effects, and other
// things the player should do/have happen to them over time.
func (p *Player) playerTick() {
	// TODO(lobato): setup idle ticker
	secondTicker := time.NewTicker(time.Second)
	for {
		select {
		case <-secondTicker.C:
			break
		case <-p.ctx.Done():
			secondTicker.Stop()
			return
		}
	}
}

// SetConnection sets the player connection object
func (p *Player) SetConnection(c net.Conn) {
	p.connection = c
	s := bufio.NewScanner(c)
	//r := bufio.NewReader(c)
	// Wrap our reader in a channel so that we can select it
	// in the interp loop. When the connection is closed by p.Stop(),
	// this loop will break.
	go func(s *bufio.Scanner) {
		for {
			if !s.Scan() {
				break
			}
			p.input <- s.Text()
		}
	}(s)
}

// Disconnect this player without unloading them from the world.
func (p *Player) Disconnect() {
	p.connection.Close()
}

// Start this player and their interp loop.
func (p *Player) Start() {
	p.buildInterp = NewBuildInterp(p)
	p.gameInterp = NewGameInterp(p)
	p.loginInterp = NewLoginInterp(p)
	p.Login()

	p.Write("Welcome, by what name are you known?")

	for {
		select {
		case <-p.ctx.Done():
			log.Info().Str("player", p.Data.UUID).Msg("Player context canceled, closing connection.")
			if p.connection != nil {
				p.connection.Close()
			}
			return
		case str := <-p.input:
			str = strings.TrimSpace(str)
			err := p.currentInterp.Read(str)
			switch err {
			case ErrCommandNotFound:
				p.Write("Huh?")
			case nil:
				break
			default:
				log.Error().Err(err).
					Str("player", p.Data.UUID).
					Msg("Error interpreting input from player.")
				log.Debug().Msg(str)
			}
		}
	}
}

// Buffer will buffer output text until Flush() is called.
func (p *Player) Buffer(text string, args ...interface{}) {
	p.textBuffer += fmt.Sprintf(text, args...)
}

// Flush will write the player buffer to the player and clear the buffer.
func (p *Player) Flush() {
	if p.Flag("color") {
		p.textBuffer = color.Parse(p.textBuffer)
	} else {
		p.textBuffer = color.Strip(p.textBuffer)
	}

	fmt.Fprintf(p.connection, "%s\r\xff\xf9", p.textBuffer)
	p.WritePrompt()
	p.textBuffer = ""
}

// Write output to a player.
func (p *Player) Write(text string, args ...interface{}) {
	str := fmt.Sprintf(text, args...)
	if p.Flag("color") {
		str = color.Parse(str)
	} else {
		str = color.Strip(str)
	}

	fmt.Fprintf(p.connection, "%s\r\xff\xf9", str)
	p.WritePrompt()
}

// WritePrompt will write the player prompt to the player.
func (p *Player) WritePrompt() {
	str := p.Prompt()
	if p.Flag("color") {
		str = color.Parse(str)
	} else {
		str = color.Strip(str)
	}
	if p.ShowPrompt() {
		fmt.Fprintf(p.connection, "\n\n%s\r\xff\xf9", str)
	}
}

// WriteRaw writes raw text to the player with no transforms.
func (p *Player) WriteRaw(text string, args ...interface{}) {
	fmt.Fprintf(p.connection, text, args...)
}

// Save a player to disk
func (p *Player) Save() error {
	data, err := json.Marshal(p.Data)
	if err != nil {
		return err
	}

	fname := uuid.NewV5(uuid.NamespaceOID, strings.ToLower(p.GetName()))
	err = ioutil.WriteFile(fmt.Sprintf("%s/%s", config.GetString("save_path"), fname), data, 0644)
	if err != nil {
		return err
	}
	return nil
}

// Load a player from source. Returns true if player was loaded.
func (p *Player) Load() (bool, error) {
	fname := uuid.NewV5(uuid.NamespaceOID, strings.ToLower(p.GetName()))
	data, err := ioutil.ReadFile(fmt.Sprintf("%s/%s", config.GetString("save_path"), fname))
	// TODO: Make this more robust, need to know if error is because of file
	// not found, or error reading.
	if err != nil {
		log.Error().Err(err).Str("player", p.Data.Name).Msg("error loading player")
		return false, nil
	}

	err = json.Unmarshal(data, &p.Data)
	if err != nil {
		return false, err
	}

	return true, nil
}

// Stop a player connection and unload the player from the world.
func (p *Player) Stop() {
	// TODO(lobato): Handle error
	p.Save()
	p.FromRoom()
	// Write a new line to ensure some clients don't buffer the last output.
	p.connection.Write([]byte("\n"))
	Atlas.RemovePlayer(p)
	p.cancel()
}

// ToRoom moves a player to a room
// TODO: Eventually, unwind combat, etc.
func (p *Player) ToRoom(target *Room) bool {
	// Player is already in the room, don't do anything.
	if p.inRoom == target {
		return true
	}
	// Remove the player from the current room.
	if p.inRoom != nil {
		p.inRoom.RemovePlayer(p)
	}
	p.inRoom = target
	p.Data.Room = target.Data.UUID
	target.AddPlayer(p)
	return true
}

// FromRoom removes the player from their current room.
func (p *Player) FromRoom() bool {
	if p.inRoom == nil {
		return true
	}
	p.inRoom.RemovePlayer(p)
	p.inRoom = nil
	return true
}

// Command runs a command through the interp for the player.
func (p *Player) Command(cmd string) error {
	return p.currentInterp.Read(cmd)
}

// GetUUID of a player.
func (p *Player) GetUUID() string {
	return p.Data.UUID
}

// GetName of a player.
func (p *Player) GetName() string {
	return p.Data.Name
}

// SetName to a player.
func (p *Player) SetName(name string) {
	p.Data.Name = name
}

// IsPassword takes an unhashed string and returns true if the input matches
// the user password.
func (p *Player) IsPassword(password string) bool {
	return p.Data.Password == hashPassword(password)
}

// SetPassword takes an unhashed string and sets that as the user password.
func (p *Player) SetPassword(password string) {
	p.Data.Password = hashPassword(password)
	return
}

// GetRoom returns the room the player is currently in
func (p *Player) GetRoom() *Room {
	return p.inRoom
}

// SetInterp for a player.
func (p *Player) setInterp(i Interp) {
	p.currentInterp = i
}

// Build switches a player to the Build interp.
func (p *Player) Build() {
	p.setInterp(p.buildInterp)
}

// Game switches a player to the Game interp.
func (p *Player) Game() {
	p.setInterp(p.gameInterp)
}

// Login switches a player to the Login interp.
func (p *Player) Login() {
	p.setInterp(p.loginInterp)
}

// EnableFlag enables a given flag for a player.
func (p *Player) EnableFlag(key string) {
	p.flagMutex.Lock()
	defer p.flagMutex.Unlock()
	p.Data.Flags[key] = true
}

// DisableFlag disables a flag for a player.
func (p *Player) DisableFlag(key string) {
	p.flagMutex.Lock()
	defer p.flagMutex.Unlock()
	p.Data.Flags[key] = false
}

// ToggleFlag will toggle the flag from it's current state, and return the new state.
func (p *Player) ToggleFlag(key string) bool {
	p.flagMutex.Lock()
	defer p.flagMutex.Unlock()
	v, ok := p.Data.Flags[key]

	if !ok || !v {
		p.Data.Flags[key] = true
		return true
	}

	p.Data.Flags[key] = false
	return false
}

// Flag returns the state of a flag for a player.
func (p *Player) Flag(key string) bool {
	p.flagMutex.RLock()
	defer p.flagMutex.RUnlock()
	v, ok := p.Data.Flags[key]
	if !ok {
		return false
	}
	return v
}

// Prompt will return the generated/interpreted prompt for this player.
func (p *Player) Prompt() string {
	str := p.Data.Prompt
	str = strings.ReplaceAll(str, "%h", fmt.Sprintf("%d", p.GetStat("health")))
	str = strings.ReplaceAll(str, "%m", fmt.Sprintf("%d", p.GetStat("mana")))
	str = strings.ReplaceAll(str, "%v", fmt.Sprintf("%d", p.GetStat("move")))
	str = strings.ReplaceAll(str, "%H", fmt.Sprintf("%d", p.GetStat("max_health")))
	str = strings.ReplaceAll(str, "%M", fmt.Sprintf("%d", p.GetStat("max_mana")))
	str = strings.ReplaceAll(str, "%V", fmt.Sprintf("%d", p.GetStat("max_move")))
	return str
}

// SetPrompt will set the prompt for this player.
func (p *Player) SetPrompt(prompt string) {
	p.Data.Prompt = prompt
}

// ShowPrompt returns true if a prompt should be shown.
func (p *Player) ShowPrompt() bool {
	switch {
	case p.IsInGame() && p.Flag("prompt"):
		return true
	default:
		return false
	}
}

// IsInGame returns true if the player is in the game world, i.e. not logging in/creating.
func (p *Player) IsInGame() bool {
	if p.currentInterp == p.gameInterp || p.currentInterp == p.buildInterp {
		return true
	}
	return false
}

// GetStat will return the value of a stat.
func (p *Player) GetStat(key string) int64 {
	switch key {
	case "health":
		return p.Data.Stats.Health
	case "mana":
		return p.Data.Stats.Mana
	case "move":
		return p.Data.Stats.Move
	case "max_health":
		return p.Data.Stats.MaxHealth
	case "max_mana":
		return p.Data.Stats.MaxMana
	case "max_move":
		return p.Data.Stats.MaxMove
	default:
		// Panic and kill the whole game to avoid player corruption.
		log.Panic().Str("stat", key).Msg("invalid stat, panic to stop player corruption")
	}
	return 0
}

// ModifyStat modifies a player's stat to the given number. If relative is set,
// stat will be modified by the given value instead of set to it.
func (p *Player) ModifyStat(key string, value int64, relative bool) {
	switch key {
	case "health":
		p.Data.Stats.Health = setOrModify(p.Data.Stats.Health, value, relative)
	case "mana":
		p.Data.Stats.Mana = setOrModify(p.Data.Stats.Mana, value, relative)
	case "move":
		p.Data.Stats.Move = setOrModify(p.Data.Stats.Move, value, relative)
	case "max_health":
		p.Data.Stats.MaxHealth = setOrModify(p.Data.Stats.MaxHealth, value, relative)
	case "max_mana":
		p.Data.Stats.MaxMana = setOrModify(p.Data.Stats.MaxMana, value, relative)
	case "max_move":
		p.Data.Stats.MaxMove = setOrModify(p.Data.Stats.MaxMove, value, relative)
	}
}

// PlayerDescription returns a short description of the player's state, used in `look`, etc.
func (p *Player) PlayerDescription() string {
	return fmt.Sprintf("%s is here.", p.GetName())
}

func (p *Player) CanExit(dir direction) bool {
	room := p.inRoom
	return room.CanExit(dir)
}

func setOrModify(base int64, value int64, relative bool) int64 {
	if relative {
		return base + value
	}
	return value
}

// Map generates a map from the player's point of view, taking into account
// closed doors, hidden rooms, rooms around the corner, etc. This is a slightly
// more expensive map method that walks exits instead of coordinates, but offers
// a much more accurate view.
func (p *Player) Map(radius int64) string {
	var output string

	// Create the map array that stores map runes.
	str := make([][]string, (radius*2)+1)
	for y := range str {
		str[y] = make([]string, (radius*2)+1)
		for x := range str[y] {
			str[y][x] = " "
		}
	}

	// Keep a record of rooms walked.
	walked := make(map[string]bool)

	// Create a rooms channel that contains the rooms we need to walk and insert
	// the player starting room as the first room, at the center of the map.
	rooms := make(chan roomWalk, radius*20)
	rooms <- roomWalk{p.inRoom, radius, radius}

	// Loop until the channel contains no more entries.
L:
	for {
		select {
		case room := <-rooms:
			// Skip this room if it's out of bounds, which prevents infinite map generation. Note
			// that for the map, 0,0 is the top left -- it should never go below 0, nor should
			// it be larger than the radius + offset.
			if room.mx > ((radius*2)-1) || room.my > ((radius*2)-1) || room.mx <= 0 || room.my <= 0 {
				continue
			}
			// Scan each direction for the current room.
			for _, dir := range exitDirections {
				// Maps are 2D, skip up and down.
				if dir == dirUp || dir == dirDown {
					continue
				}

				if room.room.CanExit(dir) {
					// There is an exit in this direction, get the room reference by that direction.
					nextRoom := room.room.LinkedRoom(dir)

					// If we've already walked that room, skip, otherwise mark.
					if _, ok := walked[nextRoom.Data.UUID]; ok {
						continue
					}
					walked[nextRoom.Data.UUID] = true

					// Get the relative translation for the direction.
					x, y, _ := Atlas.getRelativeDir(dir)

					// Mark the exit room on the map.
					str[room.my-y][room.mx+x] = "#"

					// Add the exit room to the queue to be picked up on the next loop.
					rooms <- roomWalk{nextRoom, room.mx + x, room.my - y}
				}
			}
		default:
			// Queue was empty, close the channel and move on.
			close(rooms)
			break L
		}
	}

	// The player is always at the center.
	str[radius][radius] = "{R*{x"

	// Assemble our output string and return it.
	for row := range str {
		output += "  " + strings.Join(str[row], "")
		if len(str) != row+1 {
			output += "\n"
		}
	}

	return output
}
