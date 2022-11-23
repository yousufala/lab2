package main

import (
	"flag"
	"net"
	"net/rpc"
	"uk.ac.bris.cs/gameoflife/stubs"
)

func calculateNextState(world [][]byte, start int, end int) [][]byte {
	newWorld := make([][]byte, end-start)
	for i := range newWorld {
		newWorld[i] = make([]byte, len(world))
	}
	k := 0 // The position where the y would be in a particular slice from the worker since we slice them into start and end
	for y := start; y < end; y++ {
		for x := 0; x < len(world); x++ {
			count := 0
			for j := y - 1; j <= y+1; j++ {
				for i := x - 1; i <= x+1; i++ {
					if j == y && i == x {
						continue
					}
					w, z := i, j

					if z >= len(world) {
						z = 0
					}
					if w >= len(world) {
						w = 0
					}
					if z < 0 {
						z = len(world) - 1
					}
					if w < 0 {
						w = len(world) - 1
					}
					if world[z][w] == 255 {
						count++
					}
				}
			}

			if world[y][x] == 255 {
				if count < 2 {
					newWorld[k][x] = 0
					//c.events <- CellFlipped{turn, util.Cell{X: x, Y: y}}
				} else if count == 2 || count == 3 {
					newWorld[k][x] = 255
				} else {
					newWorld[k][x] = 0
					//c.events <- CellFlipped{turn, util.Cell{X: x, Y: y}}
				}
			} else {
				if count == 3 {
					newWorld[k][x] = 255
					//c.events <- CellFlipped{turn, util.Cell{X: x, Y: y}}
				}
			}
		}
		k++
	}
	return newWorld
}

func worker(world [][]byte, startY int, endY int, out chan<- [][]byte) {
	world1 := calculateNextState(world, startY, endY)
	out <- world1
}

func workerWorks(World [][]byte, threads int) [][]byte {
	WorkerOut := make([]chan [][]byte, threads) // A 2D matrix of channels to put in the slices of the world
	for i := range WorkerOut {
		WorkerOut[i] = make(chan [][]byte)
	}

	sliceHeight := len(World) / 16
	//remaining := len(World) % threads

	for thread := 0; thread < 8; thread++ {
		go worker(World, thread*sliceHeight, (thread+1)*sliceHeight, WorkerOut[thread])
	}

	newWorld := make([][]byte, 0) // A new world slice to append what was taken from the worker out channel
	for i := 0; i < 8; i++ {
		part := <-WorkerOut[i]
		newWorld = append(newWorld, part...)
	}

	return newWorld
}

type Game struct{}

func (s *Game) ProcessGameOfLife(req stubs.Request, res *stubs.Response) (err error) {
	res.World = workerWorks(req.World, req.Thread)
	return
}

func main() {
	pAddr := flag.String("port", "8030", "Port to listen on")
	flag.Parse()
	rpc.Register(&Game{})
	listner, _ := net.Listen("tcp", ":"+*pAddr)
	defer listner.Close()
	rpc.Accept(listner)
}
