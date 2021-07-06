// Package path is a pathing library for 3D arrays.
package path

type Path struct {
	Cells [][][]*Cell
}

type Exit struct {
	Wall   bool
	Closed bool
}

type Cell struct {
	X     int64
	Y     int64
	Z     int64
	Exits []*Exit
}

type CellIterator func(*Cell)

// NewPath creates a new pathing map with a given radius of cells.
func NewPath(radius int64) *Path {
	size := radius * 2

	// Initialize X
	cells := make([][][]*Cell, size)
	for x := range cells {
		// Initialize Y
		cells[int64(x)] = make([][]*Cell, size)
		for y := range cells[int64(x)] {
			// Initalize Z
			cells[int64(x)][int64(y)] = make([]*Cell, size)
			for z := range cells[int64(x)][int64(y)] {
				// Make the cell exits
				exits := make([]*Exit, 6)
				for i := range exits {
					exits[i] = &Exit{}
				}
				// Create the cell
				cells[int64(x)][int64(y)][int64(z)] = &Cell{
					X:     int64(x),
					Y:     int64(y),
					Z:     int64(z),
					Exits: exits,
				}
			}
		}
	}
	return &Path{
		Cells: cells,
	}
}

// AllCells calls the given iterator for each cell.
func (p *Path) AllCells(fn CellIterator) {
	for x := range p.Cells {
		for y := range p.Cells[int64(x)] {
			for z := range p.Cells[int64(y)] {
				fn(p.Cells[int64(x)][int64(y)][int64(z)])
			}
		}
	}
}

// Exit returns an exit pointer for a cell.
func (c *Cell) Exit(dir string) *Exit {
	switch dir {
	case "north":
		return c.Exits[0]
	case "south":
		return c.Exits[1]
	case "east":
		return c.Exits[2]
	case "west":
		return c.Exits[3]
	case "up":
		return c.Exits[4]
	case "down":
		return c.Exits[5]
	default:
		return nil
	}
}
