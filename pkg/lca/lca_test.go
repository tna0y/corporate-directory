package lca

import "testing"

type solverTestCase struct {
	Left   int
	Right  int
	Answer int
	Error  error
}

func validateSolver(t *testing.T, solver LCASolver, tests []solverTestCase) {
	for _, test := range tests {
		if ans, err := solver.SolveLCA(test.Left, test.Right); ans != test.Answer || err != test.Error {
			errorFmt := "Test failed on test=(%d, %d); expected=%d; result=%d; error=%v; expected_error=%v"
			t.Errorf(errorFmt, test.Left, test.Right, test.Answer, ans, err, test.Error)
		}
	}
}

func TestOnlineLCASolverSingleNode(t *testing.T) {
	nodes := [][]int{
		{},
	}

	tests := []solverTestCase{
		{0, 0, 0, nil},
	}

	solver := &OnlineLCASolver{}
	err := solver.Setup(nodes)
	if err != nil {
		t.Errorf("solver setup failed")
	}
	validateSolver(t, solver, tests)
}

func TestOnlineLCASolverBasic(t *testing.T) {
	nodes := [][]int{
		{1, 2},
		{},
		{},
	}

	tests := []solverTestCase{
		{1, 2, 0, nil},
	}

	solver := &OnlineLCASolver{}
	err := solver.Setup(nodes)
	if err != nil {
		t.Errorf("solver setup failed")
	}
	validateSolver(t, solver, tests)
}

func TestOnlineLCASolverBinaryTree(t *testing.T) {
	nodes := [][]int{
		{1, 2},
		{3, 4},
		{5, 6},
		{},
		{},
		{},
		{},
	}

	tests := []solverTestCase{
		{0, 0, 0, nil},
		{0, 6, 0, nil},
		{1, 6, 0, nil},
		{1, 2, 0, nil},
		{3, 4, 1, nil},
		{6, 5, 2, nil},
		{5, 3, 0, nil},
		{5, 1, 0, nil},
	}

	solver := &OnlineLCASolver{}
	err := solver.Setup(nodes)
	if err != nil {
		t.Errorf("solver setup failed")
	}
	validateSolver(t, solver, tests)
}

func TestOnlineLCASolverEmtpyTree(t *testing.T) {
	nodes := [][]int{}

	solver := &OnlineLCASolver{}
	err := solver.Setup(nodes)
	if err != nil {
		t.Errorf("solver setup failed")
	}
}

func TestOnlineLCASolverDisjointTree(t *testing.T) {
	nodes := [][]int{
		{1},
		{},
		{3},
		{},
	}

	solver := &OnlineLCASolver{}
	err := solver.Setup(nodes)
	if err != ErrInvalidTree {
		t.Errorf("disjoint graph not detected")
	}
}

func TestOnlineLCASolverTreeInvalidRoot(t *testing.T) {
	nodes := [][]int{
		{},
		{},
		{},
		{0, 1, 2},
	}

	solver := &OnlineLCASolver{}
	err := solver.Setup(nodes)
	if err != ErrInvalidTree {
		t.Errorf("disjoint graph not detected")
	}
}
