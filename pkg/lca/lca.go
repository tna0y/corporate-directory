package lca

import (
	"corporate-directory/pkg/util"
	"errors"
	"math"
	"sync"
)

var (
	ErrInvalidTree = errors.New(`graph is not a tree`)
)

type LCASolver interface {
	Setup([][]int) error
	SolveLCA(first, second int) (int, error)
}

type OnlineLCASolver struct {
	order         []int
	first         []int
	heights       []int
	sqrts         []util.ArgMinResult
	blockLen      int
	nodeIdToIndex map[int]int
	mutex         sync.Mutex
}

func (solver *OnlineLCASolver) Setup(nodes [][]int) error {
	if len(nodes) == 0 {
		return nil
	}

	err := solver.prepareDfs(nodes)
	if err != nil {
		return err
	}
	solver.prepareRmq()
	return nil
}

func (solver *OnlineLCASolver) SolveLCA(first, second int) (int, error) {
	return solver.solve(first, second), nil
}

func (solver *OnlineLCASolver) prepareDfs(nodes [][]int) error {

	solver.order = make([]int, 0, 2*len(nodes))
	solver.first = make([]int, len(nodes))
	solver.heights = make([]int, len(nodes))

	been := make([]int, len(nodes))

	dfsStack := make([]int, 1, len(nodes))
	dfsStack[0] = 0

	curHeight := 0
	for len(dfsStack) > 0 {
		lastPos := len(dfsStack) - 1
		item := dfsStack[lastPos]
		dfsStack = dfsStack[:lastPos]

		solver.order = append(solver.order, item)

		if been[item] == 0 {
			solver.heights[item] = curHeight
			solver.first[item] = len(solver.order) - 1
			for _, child := range nodes[item] {
				dfsStack = append(dfsStack, item, child)
			}
			curHeight++
		}
		if been[item] == len(nodes[item]) {
			curHeight--
		} else if been[item] > len(nodes[item]) {
			return ErrInvalidTree
		}

		been[item]++
	}
	for _, v := range been {
		if v == 0 {
			return ErrInvalidTree
		}
	}

	return nil
}

func (solver *OnlineLCASolver) prepareRmq() {
	solver.blockLen = int(math.Sqrt(float64(len(solver.order))))

	solver.sqrts = make([]util.ArgMinResult, 1+len(solver.order)/solver.blockLen)
	for i, vertex := range solver.order {
		argmin := util.ArgMinResult{vertex, solver.heights[vertex]}
		solver.sqrts[i/solver.blockLen] = util.ArgMin(solver.sqrts[i/solver.blockLen], argmin)
	}
}

func (solver *OnlineLCASolver) solveRmqDirect(left, right int) util.ArgMinResult {
	min := util.ArgMinResult{-1, int(math.MaxInt32)}
	if right > len(solver.order) {
		right = len(solver.order)
	}
	for _, vertex := range solver.order[left:right] {
		argmin := util.ArgMinResult{vertex, solver.heights[vertex]}
		min = util.ArgMin(min, argmin)
	}
	return min
}

func (solver *OnlineLCASolver) solve(left, right int) int {
	min := util.ArgMinResult{-1, int(math.MaxInt32)}
	left = solver.first[left]
	right = solver.first[right]
	if right < left {
		left, right = right, left
	}
	for i := left; i <= right; {
		if i%solver.blockLen == 0 && i+solver.blockLen-1 <= right {
			min = util.ArgMin(min, solver.sqrts[i/solver.blockLen])
			i += solver.blockLen
		} else {
			vertex := solver.order[i]
			argmin := util.ArgMinResult{vertex, solver.heights[vertex]}
			min = util.ArgMin(min, argmin)
			i++
		}
	}
	return min.Pos
}
