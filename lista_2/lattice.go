package main

type Explorer struct {
	id int
}

type Vertex struct {
	id       int
	x        int
	y        int
	explorer *Explorer
	self     <-chan *Explorer
	north    chan<- *Explorer
	south    chan<- *Explorer
	east     chan<- *Explorer
	west     chan<- *Explorer
}

func CreateLattice(n, m int) [][]Vertex {
	// create all edges first and then create all vertices
	edges := make([]chan *Explorer, n*m)

	for i := 0; i < n*m; i++ {
		edges[i] = make(chan *Explorer)
	}

	vertices := make([][]Vertex, n)
	for y := 0; y < m; y++ {
		vertices[y] = make([]Vertex, n)
		for x := 0; x < n; x++ {
			vertices[y][x] = Vertex{id: y*n + x, x: x, y: y}
		}
	}

	for x := 0; x < n; x++ {
		for y := 0; y < m; y++ {
			if y > 0 {
				vertices[y][x].north = edges[(y-1)*n+x]
			}
			if y < m-1 {
				vertices[y][x].south = edges[(y+1)*n+x]
			}
			if x > 0 {
				vertices[y][x].west = edges[y*n+x-1]
			}
			if x < n-1 {
				vertices[y][x].east = edges[y*n+x+1]
			}
			vertices[y][x].self = edges[y*n+x]
		}
	}

	return vertices
}
