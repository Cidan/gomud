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
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/Cidan/gomud/color"
	"github.com/Cidan/gomud/config"
	"github.com/Cidan/gomud/lock"
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
	connection     net.Conn
	input          chan string //*bufio.Reader
	Data           *playerData
	gameInterp     *Game
	buildInterp    *BuildInterp
	textInterp     *TextInterp
	loginInterp    *Login
	currentInterp  Interp
	inRoom         *Room
	textBuffer     string
	lock           *lock.Lock
	ctx            context.Context
	cancel         context.CancelFunc
	lastActionTime time.Time
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

	uuid := uuid.NewV4().String()
	p := &Player{
		Data: &playerData{
			UUID:  uuid,
			Flags: make(map[string]bool),
			Stats: &playerStats{},
		},
		lastActionTime: time.Now(),
		input:          make(chan string),
		lock:           lock.New(uuid),
		ctx:            ctx,
		cancel:         cancel,
	}
	ictx := lock.Context(p.ctx, p.GetUUID()+"NewPlayer")
	p.setDefaults(ictx)
	return p
}

// setDefaults sets various defaults for new players.
func (p *Player) setDefaults(ctx context.Context) {
	p.EnableFlag(ctx, "prompt")
	p.EnableFlag(ctx, "color")
	p.DisableFlag(ctx, "automap")
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
	minuteTicker := time.NewTicker(time.Minute)
	secondTicker := time.NewTicker(time.Second)
	for {
		select {
		case <-secondTicker.C:
			break
		case <-minuteTicker.C:
			break
		case <-p.ctx.Done():
			secondTicker.Stop()
			minuteTicker.Stop()
			return
		}
	}
}

// SetConnection sets the player connection object
func (p *Player) SetConnection(ctx context.Context, c net.Conn) {
	p.lock.Lock(ctx)
	p.connection = c
	p.lock.Unlock(ctx)
	s := bufio.NewScanner(c)
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

func (p *Player) Context() context.Context {
	return p.ctx
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
	p.textInterp = NewTextInterp(p)
	ctx := lock.Context(p.ctx, p.GetUUID()+"login")
	p.Login(ctx)

	p.Write(ctx, "Welcome, by what name are you known?")

	for {
		select {
		case <-p.ctx.Done():
			log.Info().Str("player", p.Data.UUID).Msg("Player context canceled, closing connection.")
			if p.currentInterp == p.textInterp {
				p.Command(":q")
			}
			if conn := p.connection; conn != nil {
				conn.Close()
			}
			return
		case str := <-p.input:
			// TODO(lobato): interp changes?
			select {
			default:
				break
			}
			str = strings.TrimSpace(str)
			ctx := lock.Context(p.ctx, p.GetUUID()+"interp")
			p.lock.Lock(ctx)
			err := p.currentInterp.Read(ctx, str)
			p.lastActionTime = time.Now()
			p.lock.Unlock(ctx)
			switch err {
			case ErrCommandNotFound:
				p.Write(ctx, "Huh?")
			case nil:
				break
			default:
				log.Error().Err(err).
					Str("player", p.Data.UUID).
					Msg("Error interpreting input from player.")
				log.Debug().Msg(str)
			}
			// Slow down the player a bit.
			time.Sleep(time.Millisecond * 15)
		}
	}
}

// Buffer will buffer output text until Flush() is called.
func (p *Player) Buffer(ctx context.Context, text string, args ...interface{}) {
	p.lock.Lock(ctx)
	defer p.lock.Unlock(ctx)
	p.textBuffer += fmt.Sprintf(text, args...)
}

// Flush will write the player buffer to the player and clear the buffer.
func (p *Player) Flush(ctx context.Context) {
	if p.Flag(ctx, "color") {
		p.textBuffer = color.Parse(p.textBuffer)
	} else {
		p.textBuffer = color.Strip(p.textBuffer)
	}

	p.WriteRaw(ctx, "%s\r\xff\xf9", p.textBuffer)
	p.WritePrompt(ctx)

	p.lock.Lock(ctx)
	p.textBuffer = ""
	p.lock.Unlock(ctx)
}

// Write output to a player.
func (p *Player) Write(ctx context.Context, text string, args ...interface{}) {
	str := fmt.Sprintf(text, args...)
	if p.Flag(ctx, "color") {
		str = color.Parse(str)
	} else {
		str = color.Strip(str)
	}

	p.WriteRaw(ctx, "%s\r\xff\xf9", str)
	p.WritePrompt(ctx)
}

// WritePrompt will write the player prompt to the player.
func (p *Player) WritePrompt(ctx context.Context) {
	str := p.Prompt()
	if p.Flag(ctx, "color") {
		str = color.Parse(str)
	} else {
		str = color.Strip(str)
	}

	p.lock.Lock(ctx)
	if p.currentInterp == p.textInterp {
		defer p.lock.Unlock(ctx)
		p.WriteRaw(ctx, "\n[:w to save, :q to quit]\r\xff\xf9")
		return
	}
	p.lock.Unlock(ctx)

	if p.IsBuilding() {
		p.BuildPrompt(ctx)
		return
	}
	if p.ShowPrompt(ctx) {
		p.WriteRaw(ctx, "\n\n%s\r\xff\xf9", str)
	}
}

// BuildPrompt displays the build prompt
func (p *Player) BuildPrompt(ctx context.Context) {
	room := p.GetRoom(ctx)
	autobuild := "{gtrue{x"
	if !p.Flag(ctx, "autobuild") {
		autobuild = "{rfalse{x"
	}

	str := fmt.Sprintf(
		"\n\nRoom %d,%d,%d autobuild: %s >\r\xff\xf9",
		room.Data.X,
		room.Data.Y,
		room.Data.Z,
		autobuild,
	)

	if p.Flag(ctx, "color") {
		str = color.Parse(str)
	} else {
		str = color.Strip(str)
	}

	p.WriteRaw(ctx, str)
}

// WriteRaw writes raw text to the player with no transforms.
func (p *Player) WriteRaw(ctx context.Context, text string, args ...interface{}) {
	p.lock.Lock(ctx)
	defer p.lock.Unlock(ctx)
	if conn := p.connection; conn != nil {
		fmt.Fprintf(p.connection, text, args...)
	}
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
	var pd playerData
	err = json.Unmarshal(data, &pd)
	if err != nil {
		return false, err
	}

	// Load the data via an atomic pointer swap so we don't have to lock.
	target := (*unsafe.Pointer)(unsafe.Pointer(&p.Data))
	value := unsafe.Pointer(&pd)
	atomic.StorePointer(target, value)
	return true, nil
}

// Stop a player connection and unload the player from the world.
func (p *Player) Stop(ctx context.Context) {
	// TODO(lobato): Handle error
	p.Save()
	p.cancel()
	target := (*unsafe.Pointer)(unsafe.Pointer(&p.inRoom))
	atomic.StorePointer(target, nil)
	/*
		p.Mutex("room").Lock()
		if room := p.inRoom; room != nil {
			room.RemovePlayer(p)
			p.inRoom = nil
		}
		p.Mutex("room").Unlock()
	*/
	// Write a new line to ensure some clients don't buffer the last output.
	p.WriteRaw(ctx, "\n")
	Atlas.RemovePlayer(p)

}

// ToRoom moves a player to a room
// TODO: Eventually, unwind combat, etc.
func (p *Player) ToRoom(ctx context.Context, target *Room) bool {
	p.lock.Lock(ctx)
	defer p.lock.Unlock(ctx)
	room := p.GetRoom(ctx)
	// Player is already in the room, don't do anything.
	if room == target {
		return true
	}

	// Remove the player from the current room.
	if room != nil {
		// TODO(lobato): Room lock?
		room.RemovePlayer(p)
	}

	p.inRoom = target
	p.Data.Room = target.Data.UUID
	target.AddPlayer(p)
	return true
}

// GetRoom returns the room the player is currently in.
func (p *Player) GetRoom(ctx context.Context) *Room {
	p.lock.Lock(ctx)
	defer p.lock.Unlock(ctx)
	return p.inRoom
}

// Command runs a command through the interp for the player.
func (p *Player) Command(cmd string) error {
	// Commands lock the interp via input, so spool this off.
	go func(p *Player, cmd string) {
		p.input <- cmd + "\n"
	}(p, cmd)
	return nil
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

// SetInterp for a player.
func (p *Player) setInterp(ctx context.Context, i Interp) {
	p.lock.Lock(ctx)
	p.currentInterp = i
	p.lock.Unlock(ctx)
}

// Build switches a player to the Build interp.
func (p *Player) Build(ctx context.Context) {
	p.setInterp(ctx, p.buildInterp)
}

// Game switches a player to the Game interp.
func (p *Player) Game(ctx context.Context) {
	p.setInterp(ctx, p.gameInterp)
}

// Login switches a player to the Login interp.
func (p *Player) Login(ctx context.Context) {
	p.setInterp(ctx, p.loginInterp)
}

// EnableFlag enables a given flag for a player.
func (p *Player) EnableFlag(ctx context.Context, key string) {
	p.lock.Lock(ctx)
	defer p.lock.Unlock(ctx)
	p.Data.Flags[key] = true
}

// DisableFlag disables a flag for a player.
func (p *Player) DisableFlag(ctx context.Context, key string) {
	p.lock.Lock(ctx)
	defer p.lock.Unlock(ctx)
	p.Data.Flags[key] = false
}

// ToggleFlag will toggle the flag from it's current state, and return the new state.
func (p *Player) ToggleFlag(ctx context.Context, key string) bool {
	p.lock.Lock(ctx)
	defer p.lock.Unlock(ctx)
	v, ok := p.Data.Flags[key]

	if !ok || !v {
		p.Data.Flags[key] = true
		return true
	}

	p.Data.Flags[key] = false
	return false
}

// Flag returns the state of a flag for a player.
func (p *Player) Flag(ctx context.Context, key string) bool {
	p.lock.Lock(ctx)
	defer p.lock.Unlock(ctx)
	v, ok := p.Data.Flags[key]
	if !ok {
		return false
	}
	return v
}

func (p *Player) GetData() *playerData {
	target := (*unsafe.Pointer)(unsafe.Pointer(&p.Data))
	return (*playerData)(atomic.LoadPointer(target))
}

// Prompt will return the generated/interpreted prompt for this player.
func (p *Player) Prompt() string {
	data := p.GetData()
	str := data.Prompt
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
func (p *Player) ShowPrompt(ctx context.Context) bool {
	switch {
	case p.IsInGame(ctx) && p.Flag(ctx, "prompt"):
		return true
	default:
		return false
	}
}

// IsInGame returns true if the player is in the game world, i.e. not logging in/creating.
func (p *Player) IsInGame(ctx context.Context) bool {
	p.lock.Lock(ctx)
	defer p.lock.Unlock(ctx)
	if p.currentInterp == p.gameInterp || p.currentInterp == p.buildInterp {
		return true
	}
	return false
}

// IsBuilding returns true if the player is in building mode.
func (p *Player) IsBuilding() bool {
	if p.currentInterp == p.buildInterp {
		return true
	}
	return false
}

// GetStat will return the value of a stat.
func (p *Player) GetStat(key string) int64 {
	target := (*unsafe.Pointer)(unsafe.Pointer(&p.Data))
	data := (*playerData)(atomic.LoadPointer(target))
	switch key {
	case "health":
		return data.Stats.Health
	case "mana":
		return data.Stats.Mana
	case "move":
		return data.Stats.Move
	case "max_health":
		return data.Stats.MaxHealth
	case "max_mana":
		return data.Stats.MaxMana
	case "max_move":
		return data.Stats.MaxMove
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

func (p *Player) CanExit(ctx context.Context, dir direction) bool {
	room := p.GetRoom(ctx)
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
func (p *Player) Map(ctx context.Context, radius int64) string {
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

	room := p.GetRoom(ctx)
	if room == nil {
		return ""
	}

	rooms <- roomWalk{room, radius, radius}

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
