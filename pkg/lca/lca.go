package lca

import (
	"corporate-directory/pkg/util"
	"errors"
	"math"
)

var (
	ErrInvalidTree = errors.New(`graph is not a tree`)
)

// Interface for mocks and ability to swap algorithms easily
type LCASolver interface {
	Setup([][]int) error
	SolveLCA(first, second int) (int, error)
}

// Solver implementation. Implementation includes preprocessing, in which we build orderVisited array during
// a DFS down the tree. Node is added to the array once algorithm reaches it for the first time and each time
// algorithm returns to the node from it's children. Also we build firstVisit array in which we keep first occurence
// of each node id in orderVisitedArray. As well during the DFS we build heights array in which we store height (depth)
// of each node. This way for any two vertices V1 and V2 their LCA will lay somewhere in orderVisited between
// firstVisit[V1] and firstVisit[V2] and will have minimum height. Now we have reduced the problem to RMQ problem.
// RMQ is solved via SQRT-decomposition for the sake of code simplicity. Overall we have O(|V|) preprocessing time and
// O(sqrt(|V|) time complexity for each query.
type OnlineLCASolver struct {
	// Order in which each node is visited during DFS
	orderVisited []int
	// First time we visit each node
	firstVisit []int
	// height of each node in the tree
	heights []int
	// Array of RMQ results for sqrt partitions
	sqrts []util.ArgMinResult
	// SQRT of len(orderVisited), sqrt decomposition block len
	blockLen int
}

// Setup solver with nodes, may be called multiple times on the same structure
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

// Get LCA solution for two arbitrary vertices in the array
func (solver *OnlineLCASolver) SolveLCA(first, second int) (int, error) {
	return solver.solve(first, second), nil
}

// Perform DFS on the tree, populating Solver's structs
func (solver *OnlineLCASolver) prepareDfs(nodes [][]int) error {

	// initialize structures with expected len/cap
	solver.orderVisited = make([]int, 0, 2*len(nodes))
	solver.firstVisit = make([]int, len(nodes))
	solver.heights = make([]int, len(nodes))

	// keep track of how many times we visited each node
	been := make([]int, len(nodes))

	// Use stack approach for DFS to avoid recursion
	dfsStack := make([]int, 1, len(nodes))
	dfsStack[0] = 0

	// Iterate maintaining current height(depth)
	curHeight := 0
	for len(dfsStack) > 0 {

		// Pop from the stack. This implementation uses internal structure of Go's slices, avoiding extra reallocation
		// by reusing the same underlying array
		lastPos := len(dfsStack) - 1
		item := dfsStack[lastPos]
		dfsStack = dfsStack[:lastPos]

		// Main action over here
		solver.orderVisited = append(solver.orderVisited, item)

		// First time we enter some node
		if been[item] == 0 {
			// update height array and record time of first visit
			solver.heights[item] = curHeight
			solver.firstVisit[item] = len(solver.orderVisited) - 1

			// Push children and this node to come back after each child
			for _, child := range nodes[item] {
				dfsStack = append(dfsStack, item, child)
			}
			curHeight++
		}
		// Last time we enter the node, should
		if been[item] == len(nodes[item]) {
			curHeight--
		} else if been[item] > len(nodes[item]) { // Will trigger if graph is a DAG, not a tree or has cycles
			return ErrInvalidTree
		}

		been[item]++
	}
	// If some node has not been visited then the tree is not accessible from the root node
	for _, v := range been {
		if v == 0 {
			return ErrInvalidTree
		}
	}

	return nil
}

// Precalculate RMQ blocks. We divide entire array into blocks of len Sqrt(len(array)) so each query will at most
// cause O(sqrt(|V|) operations
func (solver *OnlineLCASolver) prepareRmq() {
	solver.blockLen = int(math.Sqrt(float64(len(solver.orderVisited))))

	// simply iterate over all vertices and update RMQ for corresponding block
	solver.sqrts = make([]util.ArgMinResult, 1+len(solver.orderVisited)/solver.blockLen)
	for i, vertex := range solver.orderVisited {
		argmin := util.ArgMinResult{vertex, solver.heights[vertex]}
		solver.sqrts[i/solver.blockLen] = util.ArgMin(solver.sqrts[i/solver.blockLen], argmin)
	}
}

// Prepare a response for an online request. Iterating from left to right and take aggregated result from entire block
// in case our request covers an entire block
func (solver *OnlineLCASolver) solve(left, right int) int {
	min := util.ArgMinResult{-1, int(math.MaxInt32)}
	// get first visits of vertices and swap them if necessary
	left = solver.firstVisit[left]
	right = solver.firstVisit[right]
	if right < left {
		left, right = right, left
	}

	for i := left; i <= right; {
		// In case some SQRT block is inside the request we use aggregated value
		if i%solver.blockLen == 0 && i+solver.blockLen-1 <= right {
			min = util.ArgMin(min, solver.sqrts[i/solver.blockLen])
			i += solver.blockLen
		} else { // Update one by one otherwise
			vertex := solver.orderVisited[i]
			argmin := util.ArgMinResult{vertex, solver.heights[vertex]}
			min = util.ArgMin(min, argmin)
			i++
		}
	}
	return min.Pos
}
