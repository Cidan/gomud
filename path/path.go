// Package path is a pathing library for 3D arrays.
package path

import (
	"fmt"
	"strings"
)

type Path struct {
	radius int64
	Cells  [][][]Cell
}

type Exit struct {
	Wall   bool
	Closed bool
}

type Cell struct {
	X     int64
	Y     int64
	Z     int64
	Empty bool
	Exits []Exit
}

type CellIterator func(*Cell)

// NewPath creates a new pathing map with a given radius of cells.
func NewPath(radius int64) *Path {
	size := radius * 2

	// Initialize Y
	cells := make([][][]Cell, size)
	for y := range cells {
		// Initialize X
		cells[int64(y)] = make([][]Cell, size)
		for x := range cells[int64(y)] {
			// Initalize Z
			cells[int64(y)][int64(x)] = make([]Cell, size)
			for z := range cells[int64(y)][int64(x)] {
				// Make the cell exits
				exits := make([]Exit, 6)
				for i := range exits {
					exits[i] = Exit{}
				}
				// Create the cell
				cells[int64(y)][int64(x)][int64(z)] = Cell{
					X:     int64(x),
					Y:     int64(y),
					Z:     int64(z),
					Exits: exits,
				}
			}
		}
	}
	return &Path{
		Cells:  cells,
		radius: radius,
	}
}

// AllCells calls the given iterator for each cell.
func (p *Path) AllCells(fn CellIterator) {
	for y := range p.Cells {
		for x := range p.Cells[int64(y)] {
			for z := range p.Cells[int64(y)][int64(x)] {
				fn(&p.Cells[int64(y)][int64(x)][int64(z)])
			}
		}
	}
}

func (p *Path) Cell(x, y, z int64) *Cell {
	if int64(len(p.Cells)) <= x || int64(len(p.Cells[x])) <= y || int64(len(p.Cells[x][y])) <= z {
		return nil
	}
	return &p.Cells[y][x][z]
}

func (p *Path) Map(z int64) string {
	var map_str string
	var str [][]string
	str = make([][]string, (p.radius*4)+2)
	for y := range str {
		str[y] = make([]string, (p.radius*4)*2)
		for x := range str[y] {
			str[y][x] = " "
		}
	}

	var my int64 = 2
	for y := range p.Cells {
		var mx int64 = 2
		for x := range p.Cells[y] {
			cell := p.Cells[y][x][z]
			fmt.Printf("%d %d translate to %d %d\n", y, x, my, mx)
			if !cell.Empty {
				str[my][mx] = "#"
			}
			if cell.Exit("north").Wall {
				str[my-1][mx] = "-"
			}
			if cell.Exit("south").Wall {
				str[my+1][mx] = "-"
			}
			if cell.Exit("west").Wall {
				str[my][mx-1] = "|"
			}
			if cell.Exit("east").Wall {
				str[my][mx+1] = "|"
			}
			mx += 2
		}
		my += 2
	}
	for y := range str {
		map_str += strings.Join(str[y], "") + "\n"
	}
	return map_str
}

// Exit returns an exit pointer for a cell.
func (c *Cell) Exit(dir string) *Exit {
	switch dir {
	case "north":
		return &c.Exits[0]
	case "south":
		return &c.Exits[1]
	case "east":
		return &c.Exits[2]
	case "west":
		return &c.Exits[3]
	case "up":
		return &c.Exits[4]
	case "down":
		return &c.Exits[5]
	default:
		return nil
	}
}
